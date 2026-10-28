package services

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pghba-manager/backend/internal/models"
	"gorm.io/gorm"
)

type HBAService struct {
	db     *gorm.DB
	audit  *AuditService
	crypto *CryptoService
}

func NewHBAService(db *gorm.DB, audit *AuditService) *HBAService {
	return &HBAService{db: db, audit: audit}
}

func NewHBAServiceWithCrypto(db *gorm.DB, audit *AuditService, crypto *CryptoService) *HBAService {
	return &HBAService{db: db, audit: audit, crypto: crypto}
}

// RenderHBAContent generates the full pg_hba.conf content from DB rules
func (s *HBAService) RenderHBAContent(instanceID uuid.UUID) (string, error) {
	var rules []models.HBARule
	if err := s.db.
		Where("instance_id = ?", instanceID).
		Order("rule_order ASC, created_at ASC").
		Find(&rules).Error; err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString("# pg_hba.conf — managed by PGHBA Manager\n")
	sb.WriteString(fmt.Sprintf("# Generated: %s\n\n", time.Now().UTC().Format(time.RFC3339)))
	sb.WriteString("# TYPE  DATABASE        USER            ADDRESS                 METHOD\n\n")

	for _, r := range rules {
		if r.Comment != "" {
			sb.WriteString("# " + r.Comment + "\n")
		}
		prefix := ""
		if !r.Enabled {
			prefix = "# "
		}
		if r.RuleType == models.TypeLocal {
			sb.WriteString(fmt.Sprintf("%slocal\t%s\t%s\t\t%s\n", prefix, r.Database, r.DBUser, r.AuthMethod))
		} else {
			sb.WriteString(fmt.Sprintf("%s%s\t%s\t%s\t%s\t%s\n", prefix, r.RuleType, r.Database, r.DBUser, r.Address, r.AuthMethod))
		}
	}
	return sb.String(), nil
}

// CreateBackup saves a timestamped backup of the current content
func (s *HBAService) CreateBackup(instanceID uuid.UUID, triggerEvent string, note string, createdByID *uuid.UUID) error {
	content, err := s.RenderHBAContent(instanceID)
	if err != nil {
		return err
	}

	ts := time.Now().UTC().Format("20060102_150405")
	filename := fmt.Sprintf("pg_hba.conf.%s.bak", ts)

	backup := &models.HBABackup{
		InstanceID:   instanceID,
		Filename:     filename,
		FilePath:     "/var/backups/pghba/" + filename,
		FileSize:     int64(len(content)),
		Content:      content,
		TriggerEvent: triggerEvent,
		Note:         note,
		CreatedByID:  createdByID,
	}
	return s.db.Create(backup).Error
}

// DisableExpiredTempRules finds expired temp rules, disables them, and writes back to pg_hba.conf
func (s *HBAService) DisableExpiredTempRules(action string) (int, error) {
	var expired []models.TempAccess
	if err := s.db.
		Where("expires_at <= NOW() AND revoked = false").
		Find(&expired).Error; err != nil {
		return 0, err
	}

	count := 0
	// Track which instances need write-back
	affectedInstances := map[uuid.UUID]bool{}

	for _, ta := range expired {
		tx := s.db.Begin()
		tx.Model(&ta).Updates(map[string]interface{}{
			"revoked":    true,
			"revoked_at": time.Now(),
		})

		// Only hard-delete a rule on expiry if it was created brand new by the
		// Grant Temp Access flow. Pre-existing rules later marked temporary via
		// "Set Temporary" must only ever be disabled, never deleted — they
		// default to rule_order 0 just like grant-created rules, so that field
		// can't be used to tell the two apart (see created_for_temp_access).
		var createdForTempAccess bool
		tx.Raw("SELECT created_for_temp_access FROM hba_rules WHERE id = ?", ta.RuleID).Scan(&createdForTempAccess)
		if action == "delete" && createdForTempAccess {
			tx.Delete(&models.HBARule{}, "id = ?", ta.RuleID)
		} else {
			// Existing rule that was set temporary — just disable it
			tx.Model(&models.HBARule{}).Where("id = ?", ta.RuleID).Update("enabled", false)
		}

		if err := tx.Commit().Error; err != nil {
			continue
		}

		affectedInstances[ta.InstanceID] = true

		s.audit.Log(nil, "system", models.EventExpireTempAccess,
			ta.RuleID.String(), &ta.InstanceID, "", "scheduler",
			map[string]interface{}{"action": action})
		count++
	}

	// Write-back pg_hba.conf for each affected instance
	for instanceID := range affectedInstances {
		if err := s.writeBackToServer(instanceID); err != nil {
			log.Printf("[scheduler] write-back failed for instance %s: %v", instanceID, err)
		} else {
			log.Printf("[scheduler] pg_hba.conf updated for instance %s", instanceID)
		}
	}

	return count, nil
}

// writeBackToServer writes rules to pg_hba.conf via SSH and reloads PostgreSQL
func (s *HBAService) writeBackToServer(instanceID uuid.UUID) error {
	// Serialize writes per instance — see instance_lock.go for why.
	unlock := lockInstanceWrite(instanceID)
	defer unlock()

	// Load instance with host
	var inst models.PGInstance
	if err := s.db.Preload("Host").First(&inst, "id = ?", instanceID).Error; err != nil {
		return fmt.Errorf("instance not found: %w", err)
	}

	if inst.Host.SSHKey == "" {
		return fmt.Errorf("no SSH key for host %s", inst.Host.Name)
	}

	// Decrypt SSH key
	privateKey, err := s.crypto.Decrypt(inst.Host.SSHKey)
	if err != nil {
		return fmt.Errorf("decrypt SSH key failed: %w", err)
	}

	// Load all rules
	var rules []models.HBARule
	if err := s.db.Where("instance_id = ?", instanceID).
		Order("rule_order ASC, created_at ASC").
		Find(&rules).Error; err != nil {
		return err
	}

	// Render content
	content := renderHBAContent(rules)

	// SSH connect
	sshClient, err := ConnectToHost(s.db, inst.Host.ID, inst.Host.IPAddress, inst.Host.SSHPort, inst.Host.SSHUser, privateKey, inst.Host.SSHHostKeyFingerprint)
	if err != nil {
		return fmt.Errorf("SSH connect failed: %w", err)
	}
	defer sshClient.Close()

	// Refuse to write if the server currently has active rules this app
	// doesn't know about — this path runs unattended from the scheduler
	// (temp-access expiry), with no human reviewing a diff beforehand, so
	// silently regenerating the file from an incomplete database would
	// wipe out whatever the DB never learned about. A failed read (as
	// opposed to the file genuinely not existing yet) must also abort —
	// see readCurrentHBAFile.
	current, err := readCurrentHBAFile(sshClient, inst.HBAPath)
	if err != nil {
		return fmt.Errorf("failed to verify current pg_hba.conf before writing: %w", err)
	}
	if missing := FindUntrackedLiveRules(current, instanceID, rules); len(missing) > 0 {
		return fmt.Errorf(
			"refusing to write pg_hba.conf: server has %d active rule(s) not tracked in the database — use \"Import HBA\" on this instance to sync first: %s",
			len(missing), strings.Join(missing, "; "),
		)
	}

	// Backup dulu (postgres-owned file — via sudo)
	ts := time.Now().Format("20060102_150405")
	_, _ = sshClient.RunAsPostgres(fmt.Sprintf("cp %s %s.%s.bak", shellQuote(inst.HBAPath), shellQuote(inst.HBAPath), ts))

	// Write file — temp file + atomic rename, see writeHBAFileAtomic
	if err := writeHBAFileAtomic(sshClient, inst.HBAPath, content); err != nil {
		return fmt.Errorf("write pg_hba.conf failed: %w", err)
	}

	// Reload PostgreSQL
	return s.reloadPostgres(sshClient, &inst)
}

// reloadPostgres reloads PostgreSQL config
func (s *HBAService) reloadPostgres(ssh *SSHClient, inst *models.PGInstance) error {
	var reloadCmd, pgHost string
	s.db.Raw("SELECT reload_command, pg_host FROM pg_instances WHERE id = ?", inst.ID).
		Row().Scan(&reloadCmd, &pgHost)

	if reloadCmd == "pg_reload_conf" || reloadCmd == "" {
		if pgHost == "" {
			pgHost = "127.0.0.1"
		}
		cmd := fmt.Sprintf("psql -h %s -p %d -c 'SELECT pg_reload_conf();' 2>&1", pgHost, inst.Port)
		_, err := ssh.RunAsPostgres(cmd)
		return err
	}
	// Custom reload command: runs as postgres (stored command must not
	// assume an interactive shell; include sudo itself if it needs
	// privilege beyond the postgres user).
	_, err := ssh.RunAsPostgres(reloadCmd)
	return err
}
