package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pghba-manager/backend/internal/models"
	"gorm.io/gorm"
)

type AuditHandler struct {
	db *gorm.DB
}

func NewAuditHandler(db *gorm.DB) *AuditHandler {
	return &AuditHandler{db: db}
}

func (h *AuditHandler) List(c *gin.Context) {
	var logs []models.AuditLog
	q := h.db.Model(&models.AuditLog{})

	if event := c.Query("event"); event != "" {
		q = q.Where("event = ?", event)
	}
	if username := c.Query("username"); username != "" {
		q = q.Where("username ILIKE ?", "%"+username+"%")
	}
	if instanceID := c.Query("instance_id"); instanceID != "" {
		q = q.Where("instance_id = ?", instanceID)
	}
	if from := c.Query("from"); from != "" {
		if t, err := time.Parse("2006-01-02", from); err == nil {
			q = q.Where("created_at >= ?", t)
		}
	}
	if to := c.Query("to"); to != "" {
		if t, err := time.Parse("2006-01-02", to); err == nil {
			q = q.Where("created_at < ?", t.AddDate(0, 0, 1))
		}
	}

	var total int64
	q.Count(&total)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset := (page - 1) * limit

	q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs)

	// Attach diff info for ReloadConf events
	for i := range logs {
		if logs[i].Event == models.EventReloadConf {
			var diff models.HBADiff
			if err := h.db.Where("audit_log_id = ?", logs[i].ID).First(&diff).Error; err == nil {
				var detail map[string]interface{}
				json.Unmarshal([]byte(logs[i].Detail), &detail)
				if detail == nil {
					detail = map[string]interface{}{}
				}
				detail["has_diff"] = true
				detail["diff_id"]  = diff.ID
				b, _ := json.Marshal(detail)
				logs[i].Detail = string(b)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  logs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *AuditHandler) GetDiff(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var diff models.HBADiff
	if err := h.db.Where("audit_log_id = ? OR id = ?", id, id).First(&diff).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "diff not found"})
		return
	}

	var diffLines []map[string]interface{}
	json.Unmarshal([]byte(diff.DiffLines), &diffLines)

	added := 0
	removed := 0
	for _, l := range diffLines {
		if t, ok := l["type"].(string); ok {
			if t == "added" { added++ }
			if t == "removed" { removed++ }
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"id":             diff.ID,
		"audit_log_id":   diff.AuditLogID,
		"instance_id":    diff.InstanceID,
		"before_content": diff.BeforeContent,
		"after_content":  diff.AfterContent,
		"diff_lines":     diffLines,
		"added":          added,
		"removed":        removed,
		"created_at":     diff.CreatedAt,
	})
}
