package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pghba-manager/backend/internal/models"
	"gorm.io/gorm"
)

type HBAWriter struct {
	db     *gorm.DB
	audit  *AuditService
	crypto *CryptoService
}

func NewHBAWriter(db *gorm.DB, audit *AuditService, crypto *CryptoService) *HBAWriter {
	return &HBAWriter{db: db, audit: audit, crypto: crypto}
}

// WriteAndReload — render rules dari DB, tulis ke pg_hba.conf via SSH, reload PostgreSQL
func (w *HBAWriter) WriteAndReload(instanceID uuid.UUID, userID *uuid.UUID, username string, clientIP string) error {
	// Serialize writes per instance — see instance_lock.go for why.
	unlock := lockInstanceWrite(instanceID)
	defer unlock()

	// Load instance dengan host
	var inst models.PGInstance
	if err := w.db.Preload("Host").First(&inst, "id = ?", instanceID).Error; err != nil {
		return fmt.Errorf("instance not found: %w", err)
	}

	if inst.Host.SSHKey == "" {
		return fmt.Errorf("SSH key not configured for host %s", inst.Host.Name)
	}

	// Decrypt SSH key
	privateKey, err := w.crypto.Decrypt(inst.Host.SSHKey)
	if err != nil {
		return fmt.Errorf("failed to decrypt SSH key: %w", err)
	}

	// Load semua rules untuk instance ini
	var rules []models.HBARule
	if err := w.db.Where("instance_id = ?", instanceID).
		Order("rule_order ASC, created_at ASC").
		Find(&rules).Error; err != nil {
		return fmt.Errorf("failed to load rules: %w", err)
	}

	// Read current content from server for diff
	var beforeContent string

	// Render new content
	content := renderHBAContent(rules)

	// Buka SSH
	sshClient, err := ConnectToHost(w.db, inst.Host.ID, inst.Host.IPAddress, inst.Host.SSHPort, inst.Host.SSHUser, privateKey, inst.Host.SSHHostKeyFingerprint)
	if err != nil {
		return fmt.Errorf("SSH connection failed: %w", err)
	}
	defer sshClient.Close()

	// Read current content before overwrite. A failed read (as opposed to
	// the file genuinely not existing yet) must abort the write — see
	// readCurrentHBAFile for why.
	beforeContent, err = readCurrentHBAFile(sshClient, inst.HBAPath)
	if err != nil {
		return fmt.Errorf("failed to verify current pg_hba.conf before writing: %w", err)
	}

	// Refuse to write if the server currently has active rules this app
	// doesn't know about — every write regenerates the whole file from the
	// database, so proceeding would silently delete them. This can happen
	// after a rule is added by hand on the server, or was never imported.
	if missing := FindUntrackedLiveRules(beforeContent, instanceID, rules); len(missing) > 0 {
		return fmt.Errorf(
			"refusing to write pg_hba.conf: server has %d active rule(s) not tracked in the database — use \"Import HBA\" on this instance to sync first: %s",
			len(missing), strings.Join(missing, "; "),
		)
	}

	// Backup dulu di server (postgres-owned file — via sudo)
	ts := time.Now().Format("20060102_150405")
	backupPath := inst.HBAPath + "." + ts + ".bak"
	_, _ = sshClient.RunAsPostgres(fmt.Sprintf("cp %s %s", shellQuote(inst.HBAPath), shellQuote(backupPath)))

	// Tulis ke file sementara lewat stdin (tidak ada masalah quoting/ukuran),
	// lalu rename atomik ke tujuan — pg_hba.conf tidak pernah terlihat
	// dalam keadaan kosong/terpotong walau proses ini terinterupsi.
	if err := writeHBAFileAtomic(sshClient, inst.HBAPath, content); err != nil {
		return fmt.Errorf("failed to write pg_hba.conf: %w", err)
	}

	// Reload PostgreSQL
	reloadErr := w.reloadPostgres(sshClient, &inst)

	// Audit log
	w.audit.Log(userID, username, models.EventReloadConf,
		inst.HBAPath, &instanceID,
		fmt.Sprintf("%s:%d", inst.InstanceName, inst.Port),
		clientIP,
		map[string]interface{}{
			"rules_written": len(rules),
			"reload_error":  reloadErr,
		},
	)

	// Save diff to DB
	go func() {
		diffLines := computeSimpleDiff(beforeContent, content)
		diffJSON, _ := json.Marshal(diffLines)
		var lastLog models.AuditLog
		w.db.Where("instance_id = ? AND event = 'ReloadConf'", instanceID).
			Order("created_at DESC").First(&lastLog)
		w.db.Create(&models.HBADiff{
			AuditLogID:    lastLog.ID,
			InstanceID:    &instanceID,
			BeforeContent: beforeContent,
			AfterContent:  content,
			DiffLines:     string(diffJSON),
		})
	}()

	if reloadErr != nil {
		return fmt.Errorf("file written but reload failed: %w", reloadErr)
	}
	return nil
}

// reloadPostgres reload konfigurasi PostgreSQL sesuai instance type
func (w *HBAWriter) reloadPostgres(ssh *SSHClient, inst *models.PGInstance) error {
	// Cek apakah ada reload_command custom di DB
	var reloadCmd, pgHost, socketPath string
	w.db.Raw("SELECT reload_command, pg_host, socket_path FROM pg_instances WHERE id = ?", inst.ID).
		Row().Scan(&reloadCmd, &pgHost, &socketPath)

	// Custom reload command takes priority (runs as-is; the stored
	// command itself must include sudo if it needs privilege when the
	// SSH user is not postgres)
	if reloadCmd != "" && reloadCmd != "pg_reload_conf" {
		out, err := ssh.RunAsPostgres(reloadCmd)
		if err != nil {
			return fmt.Errorf("reload command failed: %w, output: %s", err, out)
		}
		return nil
	}

	// Try pg_reload_conf via psql — multiple strategies
	// Strategy 1: Unix socket (no password needed for postgres user)
	socketDirs := []string{
		"/var/run/postgresql",
		"/tmp",
		fmt.Sprintf("/var/run/postgresql/.s.PGSQL.%d", inst.Port),
	}
	if socketPath != "" {
		socketDirs = append([]string{socketPath}, socketDirs...)
	}

	for _, dir := range socketDirs {
		cmd := fmt.Sprintf("psql -h %s -p %d -U postgres -c 'SELECT pg_reload_conf();' 2>&1", dir, inst.Port)
		out, err := ssh.RunAsPostgres(cmd)
		if err == nil && (out == "" || !contains(out, "error") && !contains(out, "failed")) {
			return nil
		}
	}

	// Strategy 2: pg_ctl reload using data directory
	dataDir := inst.DataDirectory
	if dataDir != "" {
		pgctlPaths := []string{
			fmt.Sprintf("pg_ctl reload -D %s", shellQuote(dataDir)),
			fmt.Sprintf("/usr/lib/postgresql/%s/bin/pg_ctl reload -D %s", inst.PGVersion, shellQuote(dataDir)),
			fmt.Sprintf("/usr/pgsql-%s/bin/pg_ctl reload -D %s", inst.PGVersion, shellQuote(dataDir)),
		}
		for _, cmd := range pgctlPaths {
			out, err := ssh.RunAsPostgres(cmd)
			if err == nil {
				_ = out
				return nil
			}
		}
	}

	// Strategy 3: send SIGHUP to postmaster (postgres-owned socket read —
	// needs postgres context when connected as non-postgres user)
	cmd := fmt.Sprintf(`kill -HUP $(head -1 $(psql -h /var/run/postgresql -p %d -U postgres -t -c 'SHOW data_directory;' 2>/dev/null | tr -d ' ')/postmaster.pid 2>/dev/null) 2>&1 || true`, inst.Port)
	ssh.RunAsPostgres(cmd)

	// File is written — reload best effort, not fatal
	return nil
}

// readCurrentHBAFile reads hbaPath from the server, distinguishing "the
// file genuinely doesn't exist yet" (safe — nothing to protect, treated as
// empty) from any other read failure (permission denied, wrong path, a
// dropped SSH command, etc. — must NOT be treated as empty).
//
// The previous version used `cat %s 2>/dev/null || echo ”`, which masks
// every failure as a successful empty read. That silently defeated
// FindUntrackedLiveRules: a permission error or misconfigured hba_path
// looked identical to "no file yet", so the untracked-rules guard saw
// nothing to compare against and let the write through — regenerating the
// file from however few rules the database had (zero, for an instance
// that was never imported) and locking out every real connection,
// including local ones, with no warning at all.
func readCurrentHBAFile(ssh *SSHClient, hbaPath string) (string, error) {
	out, err := ssh.RunAsPostgres(fmt.Sprintf("cat %s", shellQuote(hbaPath)))
	if err == nil {
		return out, nil
	}
	if strings.Contains(out, "No such file or directory") {
		return "", nil
	}
	return "", fmt.Errorf("could not read %s from server: %w (output: %s)", hbaPath, err, strings.TrimSpace(out))
}

// writeHBAFileAtomic writes content to hbaPath via a temp file + atomic
// rename, streaming content over stdin (no shell quoting/escaping and no
// heredoc terminator-collision risk). Runs the whole pipeline as the
// postgres OS user when the SSH session user is not postgres, so
// chmod/chown --reference and mv on postgres-owned files succeed.
// chmod/chown best-effort preserve the original file's owner/permissions
// since `mv` on rename replaces them.
func writeHBAFileAtomic(ssh *SSHClient, hbaPath string, content string) error {
	tmpPath := hbaPath + ".tmp"
	// The chmod/chown/mv trio must only run if `cat` fully succeeded — grouped
	// in `{ }` so the leading `&&` gates entry into the whole block. Without
	// this grouping, `;` doesn't inherit `cat`'s exit status, so `mv` would
	// run unconditionally even when `cat` failed or was interrupted midway
	// (e.g. the app process getting killed during a restart while streaming
	// content over SSH), overwriting the real file with an empty/truncated
	// temp file instead of leaving it untouched.
	cmd := fmt.Sprintf(
		"cat > %s && { chmod --reference=%s %s 2>/dev/null; chown --reference=%s %s 2>/dev/null; mv -f %s %s; }",
		shellQuote(tmpPath), shellQuote(hbaPath), shellQuote(tmpPath), shellQuote(hbaPath), shellQuote(tmpPath), shellQuote(tmpPath), shellQuote(hbaPath),
	)
	_, err := ssh.RunAsPostgresWithInput(cmd, content)
	return err
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// renderHBAContent render rules menjadi pg_hba.conf content
func renderHBAContent(rules []models.HBARule) string {
	var sb strings.Builder
	sb.WriteString("# pg_hba.conf — managed by PGHBA Manager\n")
	sb.WriteString(fmt.Sprintf("# Last updated: %s\n\n", time.Now().UTC().Format(time.RFC3339)))
	sb.WriteString("# TYPE  DATABASE        USER            ADDRESS                 METHOD\n\n")

	for _, r := range rules {
		// Comment
		if r.Comment != "" {
			sb.WriteString("# " + r.Comment + "\n")
		}

		prefix := ""
		if !r.Enabled {
			prefix = "# "
		}

		if r.RuleType == models.TypeLocal {
			sb.WriteString(fmt.Sprintf("%slocal\t%s\t%s\t\t\t\t%s\n",
				prefix, r.Database, r.DBUser, r.AuthMethod))
		} else {
			sb.WriteString(fmt.Sprintf("%s%s\t%s\t%s\t%s\t\t%s\n",
				prefix, r.RuleType, r.Database, r.DBUser, r.Address, r.AuthMethod))
		}
	}
	return sb.String()
}

// computeSimpleDiff generates a line-by-line diff
func computeSimpleDiff(before, after string) []map[string]interface{} {
	oldLines := strings.Split(before, "\n")
	newLines := strings.Split(after, "\n")
	var result []map[string]interface{}

	oldSet := map[string]bool{}
	newSet := map[string]bool{}
	for _, l := range oldLines {
		if strings.TrimSpace(l) != "" {
			oldSet[l] = true
		}
	}
	for _, l := range newLines {
		if strings.TrimSpace(l) != "" {
			newSet[l] = true
		}
	}

	for _, l := range oldLines {
		if strings.TrimSpace(l) == "" {
			continue
		}
		if !newSet[l] {
			result = append(result, map[string]interface{}{"type": "removed", "content": l})
		}
	}
	for _, l := range newLines {
		if strings.TrimSpace(l) == "" {
			continue
		}
		if !oldSet[l] {
			result = append(result, map[string]interface{}{"type": "added", "content": l})
		}
	}
	return result
}
