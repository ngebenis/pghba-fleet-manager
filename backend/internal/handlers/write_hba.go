package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pghba-manager/backend/internal/middleware"
	"github.com/pghba-manager/backend/internal/models"
	"github.com/pghba-manager/backend/internal/services"
	"gorm.io/gorm"
)

type WriteHBAHandler struct {
	db     *gorm.DB
	writer *services.HBAWriter
	crypto *services.CryptoService
}

func NewWriteHBAHandler(db *gorm.DB, audit *services.AuditService, crypto *services.CryptoService) *WriteHBAHandler {
	return &WriteHBAHandler{
		db:     db,
		writer: services.NewHBAWriter(db, audit, crypto),
		crypto: crypto,
	}
}

// WriteAndReload — tulis rules ke pg_hba.conf via SSH lalu reload PostgreSQL
func (h *WriteHBAHandler) WriteAndReload(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	claims := middleware.GetClaims(c)

	if err := h.writer.WriteAndReload(id, &claims.UserID, claims.Username, c.ClientIP()); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":  "write or reload failed",
			"detail": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "pg_hba.conf updated and PostgreSQL reloaded",
		"instance_id": id,
	})
}

// PreviewDiff — compare pg_hba.conf di server vs rules dari DB
func (h *WriteHBAHandler) PreviewDiff(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))

	// Load instance with host
	var inst models.PGInstance
	if err := h.db.Preload("Host").First(&inst, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
		return
	}

	if inst.Host.SSHKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no SSH key configured"})
		return
	}

	// Decrypt SSH key
	privateKey, err := h.crypto.Decrypt(inst.Host.SSHKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decrypt SSH key"})
		return
	}

	// Connect SSH and read current file
	sshClient, err := services.ConnectToHost(h.db, inst.Host.ID, inst.Host.IPAddress, inst.Host.SSHPort, inst.Host.SSHUser, privateKey, inst.Host.SSHHostKeyFingerprint)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "SSH connection failed: " + err.Error()})
		return
	}
	defer sshClient.Close()
	// Preview baca pg_hba.conf milik postgres — wajib user non-postgres;
	// sesi postgres + RemoveIPC=yes = IPC/shm disapu saat logout.
	if perr := sshClient.RequireNonPostgresUser(); perr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": perr.Error()})
		return
	}

	// cat via sudo agar jalan dari user non-postgres.
	currentContent, err := sshClient.RunAsPostgres("cat " + services.ShellQuotePath(inst.HBAPath))
	if err != nil {
		currentContent = "(failed to read file: " + err.Error() + ")"
	}

	// Generate new content from DB rules
	var rules []models.HBARule
	h.db.Where("instance_id = ?", id).
		Order("rule_order ASC, created_at ASC").
		Find(&rules)

	newContent := renderHBAFromRules(rules)

	// Compute line-by-line diff
	diff := computeDiff(currentContent, newContent)

	// Count stats
	added, removed, unchanged := 0, 0, 0
	for _, d := range diff {
		switch d.Type {
		case "added":
			added++
		case "removed":
			removed++
		default:
			unchanged++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"instance_id":   id,
		"instance_name": inst.InstanceName,
		"port":          inst.Port,
		"hba_path":      inst.HBAPath,
		"current":       currentContent,
		"new":           newContent,
		"diff":          diff,
		"stats": gin.H{
			"added":     added,
			"removed":   removed,
			"unchanged": unchanged,
			"changed":   added > 0 || removed > 0,
		},
	})
}

type DiffLine struct {
	Type    string `json:"type"` // "added" | "removed" | "unchanged"
	LineNum int    `json:"line_num"`
	Content string `json:"content"`
}

func computeDiff(oldContent, newContent string) []DiffLine {
	oldLines := strings.Split(strings.TrimRight(oldContent, "\n"), "\n")
	newLines := strings.Split(strings.TrimRight(newContent, "\n"), "\n")

	// Build LCS-based diff
	type lcsKey struct{ i, j int }
	memo := map[lcsKey]int{}

	var lcs func(i, j int) int
	lcs = func(i, j int) int {
		if i == 0 || j == 0 {
			return 0
		}
		k := lcsKey{i, j}
		if v, ok := memo[k]; ok {
			return v
		}
		if oldLines[i-1] == newLines[j-1] {
			memo[k] = lcs(i-1, j-1) + 1
		} else {
			a, b := lcs(i-1, j), lcs(i, j-1)
			if a > b {
				memo[k] = a
			} else {
				memo[k] = b
			}
		}
		return memo[k]
	}

	// Limit for performance on large files
	maxLines := 200
	if len(oldLines) > maxLines || len(newLines) > maxLines {
		return simpleDiff(oldLines, newLines)
	}

	// Backtrack to get diff
	var result []DiffLine
	i, j := len(oldLines), len(newLines)
	lineNum := 0

	var backtrack func(i, j int)
	backtrack = func(i, j int) {
		if i == 0 && j == 0 {
			return
		}
		if i > 0 && j > 0 && oldLines[i-1] == newLines[j-1] {
			backtrack(i-1, j-1)
			lineNum++
			result = append(result, DiffLine{Type: "unchanged", LineNum: lineNum, Content: oldLines[i-1]})
		} else if j > 0 && (i == 0 || lcs(i, j-1) >= lcs(i-1, j)) {
			backtrack(i, j-1)
			lineNum++
			result = append(result, DiffLine{Type: "added", LineNum: lineNum, Content: newLines[j-1]})
		} else {
			backtrack(i-1, j)
			lineNum++
			result = append(result, DiffLine{Type: "removed", LineNum: lineNum, Content: oldLines[i-1]})
		}
	}

	backtrack(i, j)
	return result
}

// simpleDiff for large files — line by line comparison
func simpleDiff(oldLines, newLines []string) []DiffLine {
	oldSet := map[string]bool{}
	newSet := map[string]bool{}
	for _, l := range oldLines {
		oldSet[l] = true
	}
	for _, l := range newLines {
		newSet[l] = true
	}

	var result []DiffLine
	lineNum := 0

	for _, l := range oldLines {
		lineNum++
		if !newSet[l] {
			result = append(result, DiffLine{Type: "removed", LineNum: lineNum, Content: l})
		} else {
			result = append(result, DiffLine{Type: "unchanged", LineNum: lineNum, Content: l})
		}
	}
	for _, l := range newLines {
		if !oldSet[l] {
			lineNum++
			result = append(result, DiffLine{Type: "added", LineNum: lineNum, Content: l})
		}
	}
	return result
}

func renderHBAFromRules(rules []models.HBARule) string {
	var sb strings.Builder
	sb.WriteString("# pg_hba.conf — managed by PGHBA Manager\n")
	sb.WriteString("# TYPE  DATABASE        USER            ADDRESS                 METHOD\n\n")

	for _, r := range rules {
		// Skip disabled temporary rules
		if !r.Enabled && r.IsTemporary {
			continue
		}
		// Skip comment for temporary rules
		if r.Comment != "" && !r.IsTemporary {
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
	return sb.String()
}

// BulkWriteAndReload — apply pg_hba.conf to multiple instances
func (h *WriteHBAHandler) BulkWriteAndReload(c *gin.Context) {
	type BulkReq struct {
		HostID      string   `json:"host_id"`      // filter by host
		Environment string   `json:"environment"`  // filter by environment
		InstanceIDs []string `json:"instance_ids"` // specific instances
	}
	var req BulkReq
	c.ShouldBindJSON(&req)

	claims := middleware.GetClaims(c)

	// Build query
	var instances []models.PGInstance
	q := h.db.Preload("Host")
	if len(req.InstanceIDs) > 0 {
		q = q.Where("id IN ?", req.InstanceIDs)
	} else if req.HostID != "" {
		q = q.Where("host_id = ?", req.HostID)
	} else if req.Environment != "" {
		q = q.Joins("JOIN db_hosts ON db_hosts.id = pg_instances.host_id").
			Where("db_hosts.environment = ?", req.Environment)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provide host_id, environment, or instance_ids"})
		return
	}
	q.Find(&instances)

	if len(instances) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "no instances found", "results": []interface{}{}})
		return
	}

	type Result struct {
		InstanceName string `json:"instance"`
		Port         int    `json:"port"`
		HostName     string `json:"host"`
		Status       string `json:"status"`
		Error        string `json:"error,omitempty"`
	}

	results := []Result{}
	success := 0
	failed := 0

	for _, inst := range instances {
		r := Result{
			InstanceName: inst.InstanceName,
			Port:         inst.Port,
			HostName:     inst.Host.Name,
		}

		if err := h.writer.WriteAndReload(inst.ID, &claims.UserID, claims.Username, c.ClientIP()); err != nil {
			r.Status = "failed"
			r.Error = err.Error()
			failed++
		} else {
			r.Status = "success"
			success++
		}
		results = append(results, r)
	}

	// Audit log via direct DB insert
	h.db.Exec(`INSERT INTO audit_logs (id, user_id, username, event, target, client_ip, detail, created_at)
		VALUES (gen_random_uuid(), ?, ?, 'ReloadConf', 'bulk apply', ?, ?, NOW())`,
		claims.UserID, claims.Username, c.ClientIP(),
		fmt.Sprintf(`{"total":%d,"success":%d,"failed":%d}`, len(instances), success, failed))

	c.JSON(http.StatusOK, gin.H{
		"message": "bulk apply complete",
		"total":   len(instances),
		"success": success,
		"failed":  failed,
		"results": results,
	})
}
