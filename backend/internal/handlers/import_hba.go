package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pghba-manager/backend/internal/middleware"
	"github.com/pghba-manager/backend/internal/models"
	"github.com/pghba-manager/backend/internal/services"
	"gorm.io/gorm"
)

// ImportHBA — baca pg_hba.conf dari server via SSH, parse dan import ke DB
func (h *InstanceHandler) ImportHBA(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var inst models.PGInstance
	if err := h.db.Preload("Host").First(&inst, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
		return
	}

	if inst.Host.SSHKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "SSH key belum dikonfigurasi. Edit host dan tambahkan SSH private key.",
		})
		return
	}

	result, err := importHBAForInstance(h.db, h.crypto, h.audit, &inst, c.ClientIP())
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	claims := middleware.GetClaims(c)
	h.audit.Log(
		&claims.UserID, claims.Username,
		models.EventAddRule,
		fmt.Sprintf("import %s", inst.HBAPath),
		&inst.ID,
		fmt.Sprintf("%s:%d", inst.InstanceName, inst.Port),
		c.ClientIP(),
		map[string]interface{}{
			"action":   "import_hba",
			"imported": result["imported"],
			"skipped":  result["skipped"],
			"path":     inst.HBAPath,
		},
	)

	c.JSON(http.StatusOK, gin.H{
		"message":  "import complete",
		"instance": inst.InstanceName,
		"port":     inst.Port,
		"path":     inst.HBAPath,
		"imported": result["imported"],
		"skipped":  result["skipped"],
		"rules":    result["rules"],
	})
}

// BulkImportHBA — import pg_hba.conf dari semua instances sekaligus
func (h *InstanceHandler) BulkImportHBA(c *gin.Context) {
	type BulkReq struct {
		HostID string `json:"host_id"` // optional — filter by host
	}
	var req BulkReq
	c.ShouldBindJSON(&req)

	// Load instances
	var instances []models.PGInstance
	q := h.db.Preload("Host")
	if req.HostID != "" {
		q = q.Where("host_id = ?", req.HostID)
	}
	q.Find(&instances)

	claims := middleware.GetClaims(c)

	type InstanceResult struct {
		InstanceName string `json:"instance"`
		Port         int    `json:"port"`
		HostName     string `json:"host"`
		HBAPath      string `json:"hba_path"`
		Imported     int    `json:"imported"`
		Skipped      int    `json:"skipped"`
		Error        string `json:"error,omitempty"`
		Status       string `json:"status"`
	}

	results := []InstanceResult{}
	totalImported := 0
	totalSkipped := 0
	totalErrors := 0

	for _, inst := range instances {
		r := InstanceResult{
			InstanceName: inst.InstanceName,
			Port:         inst.Port,
			HostName:     inst.Host.Name,
			HBAPath:      inst.HBAPath,
		}

		if inst.Host.SSHKey == "" {
			r.Error = "no SSH key"
			r.Status = "skipped"
			totalErrors++
			results = append(results, r)
			continue
		}

		result, err := importHBAForInstance(h.db, h.crypto, h.audit, &inst, c.ClientIP())
		if err != nil {
			r.Error = err.Error()
			r.Status = "error"
			totalErrors++
		} else {
			r.Imported = result["imported"].(int)
			r.Skipped = result["skipped"].(int)
			r.Status = "success"
			totalImported += r.Imported
			totalSkipped += r.Skipped
		}

		results = append(results, r)

		h.audit.Log(
			&claims.UserID, claims.Username,
			models.EventAddRule,
			fmt.Sprintf("bulk-import %s", inst.HBAPath),
			&inst.ID,
			fmt.Sprintf("%s:%d", inst.InstanceName, inst.Port),
			c.ClientIP(),
			map[string]interface{}{"imported": r.Imported, "skipped": r.Skipped},
		)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "bulk import complete",
		"total_imported":  totalImported,
		"total_skipped":   totalSkipped,
		"total_errors":    totalErrors,
		"total_instances": len(instances),
		"results":         results,
	})
}

// importHBAForInstance — shared logic untuk import satu instance
func importHBAForInstance(
	db *gorm.DB,
	crypto *services.CryptoService,
	audit *services.AuditService,
	inst *models.PGInstance,
	clientIP string,
) (map[string]interface{}, error) {
	privateKey, err := crypto.Decrypt(inst.Host.SSHKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt SSH key")
	}

	sshClient, err := services.ConnectToHost(db, inst.Host.ID, inst.Host.IPAddress, inst.Host.SSHPort, inst.Host.SSHUser, privateKey, inst.Host.SSHHostKeyFingerprint)
	if err != nil {
		return nil, fmt.Errorf("SSH connection failed: %w", err)
	}
	defer sshClient.Close()
	// Baca pg_hba.conf milik postgres — wajib user non-postgres;
	// sesi postgres + RemoveIPC=yes = IPC/shm disapu saat logout.
	if perr := sshClient.RequireNonPostgresUser(); perr != nil {
		return nil, perr
	}

	// cat via sudo agar jalan dari user non-postgres.
	content, err := sshClient.RunAsPostgres("cat " + services.ShellQuotePath(inst.HBAPath))
	if err != nil {
		return nil, fmt.Errorf("failed to read pg_hba.conf: %w", err)
	}

	if content == "" {
		return map[string]interface{}{
			"imported": 0, "skipped": 0, "rules": []interface{}{},
		}, nil
	}

	rules := services.ParseHBAContent(content, inst.ID)
	imported := 0
	skipped := 0
	var ruleResults []gin.H

	for _, rule := range rules {
		var existing models.HBARule
		err := db.Where(
			"instance_id = ? AND rule_type = ? AND database = ? AND db_user = ? AND address = ?",
			inst.ID, rule.RuleType, rule.Database, rule.DBUser, rule.Address,
		).First(&existing).Error

		if err == gorm.ErrRecordNotFound {
			newRule := rule
			if createErr := db.Create(&newRule).Error; createErr == nil {
				imported++
				ruleResults = append(ruleResults, gin.H{
					"action":  "imported",
					"type":    rule.RuleType,
					"db":      rule.Database,
					"user":    rule.DBUser,
					"address": rule.Address,
					"method":  rule.AuthMethod,
				})
			}
		} else if err == nil {
			skipped++
		}
	}

	return map[string]interface{}{
		"imported": imported,
		"skipped":  skipped,
		"rules":    ruleResults,
	}, nil
}
