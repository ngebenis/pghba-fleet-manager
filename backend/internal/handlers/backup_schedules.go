package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pghba-manager/backend/internal/middleware"
	"github.com/pghba-manager/backend/internal/models"
	"github.com/pghba-manager/backend/internal/services"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

type BackupScheduleHandler struct {
	db    *gorm.DB
	audit *services.AuditService
}

func NewBackupScheduleHandler(db *gorm.DB, audit *services.AuditService) *BackupScheduleHandler {
	return &BackupScheduleHandler{db: db, audit: audit}
}

func (h *BackupScheduleHandler) List(c *gin.Context) {
	var schedules []models.BackupSchedule
	h.db.Order("created_at DESC").Find(&schedules)
	c.JSON(http.StatusOK, schedules)
}

type backupScheduleReq struct {
	Name           string `json:"name"            binding:"required"`
	CronExpression string `json:"cron_expression" binding:"required"`
	Target         string `json:"target"          binding:"required"`
	ScopeType      string `json:"scope_type"      binding:"required"`
	ScopeID        string `json:"scope_id"`
	Enabled        *bool  `json:"enabled"`
}

func (req backupScheduleReq) validate() error {
	if _, err := cron.ParseStandard(req.CronExpression); err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}
	switch req.Target {
	case "local", "s3", "both":
	default:
		return fmt.Errorf("target must be one of: local, s3, both")
	}
	switch req.ScopeType {
	case "all":
	case "host", "instance":
		if req.ScopeID == "" {
			return fmt.Errorf("scope_id is required when scope_type is %q", req.ScopeType)
		}
		if _, err := uuid.Parse(req.ScopeID); err != nil {
			return fmt.Errorf("invalid scope_id")
		}
	default:
		return fmt.Errorf("scope_type must be one of: all, host, instance")
	}
	return nil
}

func (h *BackupScheduleHandler) Create(c *gin.Context) {
	var req backupScheduleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := req.validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	claims := middleware.GetClaims(c)
	uid := claims.UserID

	sched := &models.BackupSchedule{
		Name:           req.Name,
		CronExpression: req.CronExpression,
		Target:         req.Target,
		ScopeType:      req.ScopeType,
		Enabled:        true,
		CreatedByID:    &uid,
	}
	if req.ScopeType != "all" {
		sid, _ := uuid.Parse(req.ScopeID)
		sched.ScopeID = &sid
	}
	if req.Enabled != nil {
		sched.Enabled = *req.Enabled
	}

	if err := h.db.Create(sched).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create schedule: " + err.Error()})
		return
	}

	h.audit.Log(&uid, claims.Username, models.EventAddBackupSchedule, sched.Name, nil, "", c.ClientIP(),
		map[string]interface{}{"cron": sched.CronExpression, "target": sched.Target, "scope_type": sched.ScopeType})

	c.JSON(http.StatusCreated, sched)
}

func (h *BackupScheduleHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var sched models.BackupSchedule
	if err := h.db.First(&sched, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "schedule not found"})
		return
	}

	var req backupScheduleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := req.validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{
		"name":            req.Name,
		"cron_expression": req.CronExpression,
		"target":          req.Target,
		"scope_type":      req.ScopeType,
		"scope_id":        nil,
	}
	if req.ScopeType != "all" {
		sid, _ := uuid.Parse(req.ScopeID)
		updates["scope_id"] = sid
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	h.db.Model(&sched).Updates(updates)

	claims := middleware.GetClaims(c)
	h.audit.Log(&claims.UserID, claims.Username, models.EventEditBackupSchedule, sched.Name, nil, "", c.ClientIP(),
		map[string]interface{}{"cron": req.CronExpression, "target": req.Target, "scope_type": req.ScopeType})

	c.JSON(http.StatusOK, gin.H{"message": "schedule updated"})
}

func (h *BackupScheduleHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var sched models.BackupSchedule
	if err := h.db.First(&sched, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "schedule not found"})
		return
	}
	h.db.Delete(&sched)

	claims := middleware.GetClaims(c)
	h.audit.Log(&claims.UserID, claims.Username, models.EventDeleteBackupSchedule, sched.Name, nil, "", c.ClientIP(), nil)

	c.JSON(http.StatusOK, gin.H{"message": "schedule deleted"})
}
