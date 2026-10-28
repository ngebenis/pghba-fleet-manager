package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pghba-manager/backend/internal/middleware"
	"github.com/pghba-manager/backend/internal/models"
	"github.com/pghba-manager/backend/internal/services"
	"gorm.io/gorm"
)

type BackupHandler struct {
	db     *gorm.DB
	audit  *services.AuditService
	hba    *services.HBAService
	crypto *services.CryptoService
}

func NewBackupHandler(db *gorm.DB, audit *services.AuditService, hba *services.HBAService, crypto *services.CryptoService) *BackupHandler {
	return &BackupHandler{db: db, audit: audit, hba: hba, crypto: crypto}
}

func (h *BackupHandler) List(c *gin.Context) {
	var backups []models.HBABackup
	q := h.db.Preload("Instance").Preload("Instance.Host")
	if iid := c.Query("instance_id"); iid != "" {
		q = q.Where("instance_id = ?", iid)
	}
	if env := c.Query("environment"); env != "" {
		q = q.Joins("JOIN pg_instances ON pg_instances.id = hba_backups.instance_id").
			Joins("JOIN db_hosts ON db_hosts.id = pg_instances.host_id").
			Where("db_hosts.environment = ?", env)
	}
	var total int64
	q.Model(&models.HBABackup{}).Count(&total)
	limit := 50
	q.Order("hba_backups.created_at DESC").Limit(limit).Find(&backups)
	c.JSON(http.StatusOK, gin.H{"data": backups, "total": total})
}

func (h *BackupHandler) Create(c *gin.Context) {
	type req struct {
		InstanceID string `json:"instance_id" binding:"required"`
		Note       string `json:"note"`
	}
	var r req
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	claims := middleware.GetClaims(c)
	uid := claims.UserID
	iid, _ := uuid.Parse(r.InstanceID)

	note := r.Note
	if note == "" {
		note = fmt.Sprintf("Manual backup by %s", claims.Username)
	}

	if err := h.hba.CreateBackup(iid, "manual", note, &uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get the created backup
	var backup models.HBABackup
	h.db.Preload("Instance").Where("instance_id = ? AND trigger_event = 'manual'", iid).
		Order("created_at DESC").First(&backup)

	h.audit.Log(&uid, claims.Username, models.EventCreateBackup,
		backup.Filename, &iid,
		fmt.Sprintf("%s:%d", backup.Instance.InstanceName, backup.Instance.Port),
		c.ClientIP(), map[string]interface{}{"note": note})

	c.JSON(http.StatusCreated, gin.H{
		"message":  "backup created",
		"id":       backup.ID,
		"filename": backup.Filename,
		"size":     backup.FileSize,
	})
}

func (h *BackupHandler) Restore(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	claims := middleware.GetClaims(c)
	uid := claims.UserID

	var backup models.HBABackup
	if err := h.db.Preload("Instance").Preload("Instance.Host").
		First(&backup, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "backup not found"})
		return
	}

	if backup.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "backup content is empty"})
		return
	}

	// Auto-backup current state before restore
	h.hba.CreateBackup(backup.InstanceID, "before_restore",
		fmt.Sprintf("auto-backup before restore to %s", backup.Filename), &uid)

	// Delete all current non-temporary rules for this instance
	h.db.Where("instance_id = ? AND is_temporary = false", backup.InstanceID).
		Delete(&models.HBARule{})

	// Parse backup content and re-insert rules
	rules := parseHBAContent(backup.Content, backup.InstanceID)
	imported := 0
	for i := range rules {
		rules[i].CreatedByID = &uid
		if err := h.db.Create(&rules[i]).Error; err == nil {
			imported++
		}
	}

	// Write back to server
	writer := services.NewHBAWriter(h.db, h.audit, h.crypto)
	writeErr := ""
	if err := writer.WriteAndReload(backup.InstanceID, &uid, claims.Username, c.ClientIP()); err != nil {
		writeErr = err.Error()
	}

	h.audit.Log(&uid, claims.Username, models.EventRestoreBackup,
		backup.Filename, &backup.InstanceID,
		fmt.Sprintf("%s:%d", backup.Instance.InstanceName, backup.Instance.Port),
		c.ClientIP(),
		map[string]interface{}{
			"backup_id": id,
			"filename":  backup.Filename,
			"imported":  imported,
		})

	resp := gin.H{
		"message":  "restore complete",
		"filename": backup.Filename,
		"imported": imported,
	}
	if writeErr != "" {
		resp["warning"] = "rules restored but write-back failed: " + writeErr
	}
	c.JSON(http.StatusOK, resp)
}

func (h *BackupHandler) Download(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var backup models.HBABackup
	if err := h.db.First(&backup, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.Header("Content-Disposition", "attachment; filename="+backup.Filename)
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Header("Content-Length", fmt.Sprintf("%d", len(backup.Content)))
	c.String(http.StatusOK, backup.Content)
}

func (h *BackupHandler) Delete(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	claims := middleware.GetClaims(c)
	var backup models.HBABackup
	if err := h.db.First(&backup, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	h.db.Delete(&backup)
	h.audit.Log(&claims.UserID, claims.Username, models.EventCreateBackup,
		"deleted: "+backup.Filename, &backup.InstanceID, "", c.ClientIP(), nil)
	c.JSON(http.StatusOK, gin.H{"message": "backup deleted"})
}

// parseHBAContent — parse pg_hba.conf content into HBARule slice
func parseHBAContent(content string, instanceID uuid.UUID) []models.HBARule {
	var rules []models.HBARule
	order := 0
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		ruleType := models.RuleType(fields[0])
		if ruleType != models.TypeLocal && ruleType != models.TypeHost &&
			ruleType != models.TypeHostSSL && ruleType != models.TypeHostNoSSL {
			continue
		}

		rule := models.HBARule{
			InstanceID: instanceID,
			RuleType:   ruleType,
			Database:   fields[1],
			DBUser:     fields[2],
			Enabled:    true,
			RuleOrder:  order,
		}
		order++

		if ruleType == models.TypeLocal {
			if len(fields) >= 4 {
				rule.AuthMethod = models.AuthMethod(fields[3])
			}
		} else {
			rule.Address = fields[3]
			if len(fields) >= 5 {
				rule.AuthMethod = models.AuthMethod(fields[4])
			}
		}

		// Set default
		if rule.AuthMethod == "" {
			rule.AuthMethod = models.MethodMD5
		}
		rules = append(rules, rule)
	}
	return rules
}

// Preview backup content
func (h *BackupHandler) Preview(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var backup models.HBABackup
	if err := h.db.Preload("Instance").First(&backup, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	rules := parseHBAContent(backup.Content, backup.InstanceID)
	expiresIn := ""
	if time.Since(backup.CreatedAt) < 24*time.Hour {
		expiresIn = "recent"
	}
	c.JSON(http.StatusOK, gin.H{
		"id":          backup.ID,
		"filename":    backup.Filename,
		"created_at":  backup.CreatedAt,
		"note":        backup.Note,
		"size":        backup.FileSize,
		"content":     backup.Content,
		"rule_count":  len(rules),
		"expires_in":  expiresIn,
		"instance":    backup.Instance,
	})
}
