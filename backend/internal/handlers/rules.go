package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pghba-manager/backend/internal/middleware"
	"github.com/pghba-manager/backend/internal/models"
	"github.com/pghba-manager/backend/internal/services"
	"gorm.io/gorm"
)

type RuleHandler struct {
	db     *gorm.DB
	audit  *services.AuditService
	hba    *services.HBAService
	writer *services.HBAWriter
}

func NewRuleHandler(db *gorm.DB, audit *services.AuditService, hba *services.HBAService, crypto *services.CryptoService) *RuleHandler {
	return &RuleHandler{db: db, audit: audit, hba: hba, writer: services.NewHBAWriter(db, audit, crypto)}
}

func (h *RuleHandler) List(c *gin.Context) {
	var rules []models.HBARule
	q := h.db.Preload("Instance").Preload("Instance.Host")
	if iid := c.Query("instance_id"); iid != "" {
		q = q.Where("instance_id = ?", iid)
	}
	q.Order("rule_order ASC, created_at ASC").Find(&rules)
	c.JSON(http.StatusOK, rules)
}

type createRuleReq struct {
	InstanceID  string            `json:"instance_id"  binding:"required"`
	RuleType    models.RuleType   `json:"rule_type"`
	Database    string            `json:"database"`
	DBUser      string            `json:"db_user"`
	Address     string            `json:"address"`
	AuthMethod  models.AuthMethod `json:"auth_method"`
	Comment     string            `json:"comment"`
	RuleOrder   int               `json:"rule_order"`
}

func (h *RuleHandler) Create(c *gin.Context) {
	var req createRuleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	claims := middleware.GetClaims(c)
	uid := claims.UserID
	iid, _ := uuid.Parse(req.InstanceID)

	rule := &models.HBARule{
		InstanceID:  iid,
		RuleType:    req.RuleType,
		Database:    req.Database,
		DBUser:      req.DBUser,
		Address:     req.Address,
		AuthMethod:  req.AuthMethod,
		Comment:     req.Comment,
		RuleOrder:   req.RuleOrder,
		Enabled:     true,
		CreatedByID: &uid,
	}
	h.db.Create(rule)

	h.db.Preload("Instance").Preload("Instance.Host").First(rule, "id = ?", rule.ID)

	h.audit.Log(&uid, claims.Username, models.EventAddRule,
		rule.Address, &iid,
		rule.Instance.InstanceName+":"+fmt.Sprint(rule.Instance.Port), c.ClientIP(),
		map[string]interface{}{"rule_type": rule.RuleType, "method": rule.AuthMethod})

	c.JSON(http.StatusCreated, rule)
}

func (h *RuleHandler) Update(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var rule models.HBARule
	if err := h.db.Preload("Instance").First(&rule, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "rule not found"})
		return
	}
	var req createRuleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.hba.CreateBackup(rule.InstanceID, "before_edit", "auto-backup before edit rule", nil)
	h.db.Model(&rule).Updates(map[string]interface{}{
		"rule_type":   req.RuleType,
		"database":    req.Database,
		"db_user":     req.DBUser,
		"address":     req.Address,
		"auth_method": req.AuthMethod,
		"comment":     req.Comment,
		"rule_order":  req.RuleOrder,
	})
	claims := middleware.GetClaims(c)
	h.audit.Log(&claims.UserID, claims.Username, models.EventEditRule, rule.Address, &rule.InstanceID, "", c.ClientIP(), nil)
	c.JSON(http.StatusOK, rule)
}

func (h *RuleHandler) Delete(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var rule models.HBARule
	if err := h.db.First(&rule, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "rule not found"})
		return
	}
	h.hba.CreateBackup(rule.InstanceID, "before_delete", "auto-backup before delete rule", nil)
	h.db.Delete(&rule)
	claims := middleware.GetClaims(c)
	h.audit.Log(&claims.UserID, claims.Username, models.EventDeleteRule, rule.Address, &rule.InstanceID, "", c.ClientIP(), nil)
	c.JSON(http.StatusOK, gin.H{"message": "rule deleted"})
}

func (h *RuleHandler) Enable(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var rule models.HBARule
	if err := h.db.First(&rule, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "rule not found"})
		return
	}
	h.db.Model(&rule).Update("enabled", true)
	claims := middleware.GetClaims(c)
	h.audit.Log(&claims.UserID, claims.Username, models.EventEnableRule, rule.Address, &rule.InstanceID, "", c.ClientIP(), nil)
	c.JSON(http.StatusOK, gin.H{"enabled": true})
}

func (h *RuleHandler) Disable(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var rule models.HBARule
	if err := h.db.First(&rule, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "rule not found"})
		return
	}
	h.db.Model(&rule).Update("enabled", false)
	claims := middleware.GetClaims(c)
	h.audit.Log(&claims.UserID, claims.Username, models.EventDisableRule, rule.Address, &rule.InstanceID, "", c.ClientIP(), nil)
	c.JSON(http.StatusOK, gin.H{"enabled": false})
}

// SetTemporary — set existing rule as temporary with expiry time
type setTempReq struct {
	DurationValue int                     `json:"duration_value" binding:"required,min=1"`
	DurationUnit  models.TempDurationUnit `json:"duration_unit"  binding:"required"`
	Reason        string                  `json:"reason"`
}

func (h *RuleHandler) SetTemporary(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var rule models.HBARule
	if err := h.db.Preload("Instance").First(&rule, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "rule not found"})
		return
	}

	var req setTempReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims := middleware.GetClaims(c)
	uid := claims.UserID

	// Calculate expiry
	expiresAt := calcRuleExpiry(req.DurationValue, req.DurationUnit)

	// Mark rule as temporary
	comment := fmt.Sprintf("Temporary access — expires %s", expiresAt.UTC().Format(time.RFC3339))
	h.db.Model(&rule).Updates(map[string]interface{}{
		"is_temporary": true,
		"comment":      comment,
		"enabled":      true,
	})

	// Create temp_access record to track expiry
	ta := &models.TempAccess{
		RuleID:        rule.ID,
		InstanceID:    rule.InstanceID,
		GrantedByID:   uid,
		DurationValue: req.DurationValue,
		DurationUnit:  req.DurationUnit,
		ExpiresAt:     expiresAt,
		Reason:        req.Reason,
	}
	h.db.Create(ta)

	h.audit.Log(&uid, claims.Username, models.EventGrantTempAccess,
		rule.Address, &rule.InstanceID,
		fmt.Sprintf("%s:%d", rule.Instance.InstanceName, rule.Instance.Port),
		c.ClientIP(),
		map[string]interface{}{
			"duration":   req.DurationValue,
			"unit":       req.DurationUnit,
			"expires_at": expiresAt,
			"reason":     req.Reason,
		})

	c.JSON(http.StatusOK, gin.H{
		"message":    "rule set as temporary",
		"rule_id":    rule.ID,
		"expires_at": expiresAt,
		"reason":     req.Reason,
	})
}

func calcRuleExpiry(value int, unit models.TempDurationUnit) time.Time {
	now := time.Now()
	switch unit {
	case models.DurMinutes: return now.Add(time.Duration(value) * time.Minute)
	case models.DurHours:   return now.Add(time.Duration(value) * time.Hour)
	case models.DurDays:    return now.AddDate(0, 0, value)
	case models.DurWeeks:   return now.AddDate(0, 0, value*7)
	case models.DurMonths:  return now.AddDate(0, value, 0)
	default:                return now.Add(time.Duration(value) * time.Hour)
	}
}


// Reorder — update rule_order for multiple rules at once
func (h *RuleHandler) Reorder(c *gin.Context) {
	type reorderItem struct {
		ID    string `json:"id"`
		Order int    `json:"order"`
	}
	type reorderReq struct {
		Rules []reorderItem `json:"rules" binding:"required"`
	}
	var req reorderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	claims := middleware.GetClaims(c)

	for _, item := range req.Rules {
		h.db.Model(&models.HBARule{}).Where("id = ?", item.ID).
			Update("rule_order", item.Order)
	}

	// Get instance_id from first rule for write-back
	if len(req.Rules) > 0 {
		var rule models.HBARule
		if err := h.db.First(&rule, "id = ?", req.Rules[0].ID).Error; err == nil {
			uid := claims.UserID
			go func() {
				h.writer.WriteAndReload(rule.InstanceID, &uid, claims.Username, c.ClientIP())
			}()
		}
	}

	h.audit.Log(&claims.UserID, claims.Username, models.EventEditRule,
		"reorder rules", nil, "", c.ClientIP(),
		map[string]interface{}{"count": len(req.Rules)})

	c.JSON(http.StatusOK, gin.H{"message": "rules reordered", "count": len(req.Rules)})
}

// CloneRules — copy rules dari instance A ke instance B
func (h *RuleHandler) CloneRules(c *gin.Context) {
	type cloneReq struct {
		SourceInstanceID string `json:"source_instance_id" binding:"required"`
		TargetInstanceID string `json:"target_instance_id" binding:"required"`
		OverwriteExisting bool   `json:"overwrite_existing"`
		WriteBack         bool   `json:"write_back"`
	}
	var req cloneReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims := middleware.GetClaims(c)
	uid := claims.UserID

	srcID, _ := uuid.Parse(req.SourceInstanceID)
	dstID, _ := uuid.Parse(req.TargetInstanceID)

	if srcID == dstID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "source and target must be different"})
		return
	}

	// Verify both instances exist
	var src, dst models.PGInstance
	if err := h.db.Preload("Host").First(&src, "id = ?", srcID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "source instance not found"})
		return
	}
	if err := h.db.Preload("Host").First(&dst, "id = ?", dstID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "target instance not found"})
		return
	}

	// Load source rules (non-temporary only)
	var srcRules []models.HBARule
	h.db.Where("instance_id = ? AND is_temporary = false", srcID).
		Order("rule_order ASC").Find(&srcRules)

	if len(srcRules) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "no rules to clone",
			"cloned":  0, "skipped": 0,
		})
		return
	}

	// Backup target before clone
	h.hba.CreateBackup(dstID, "before_clone",
		fmt.Sprintf("auto-backup before clone from %s:%d", src.InstanceName, src.Port), &uid)

	// If overwrite: delete existing non-temporary rules in target
	if req.OverwriteExisting {
		h.db.Where("instance_id = ? AND is_temporary = false", dstID).Delete(&models.HBARule{})
	}

	cloned := 0
	skipped := 0

	for _, rule := range srcRules {
		// Check for duplicate
		if !req.OverwriteExisting {
			var existing models.HBARule
			err := h.db.Where(
				"instance_id = ? AND rule_type = ? AND database = ? AND db_user = ? AND address = ?",
				dstID, rule.RuleType, rule.Database, rule.DBUser, rule.Address,
			).First(&existing).Error
			if err == nil {
				skipped++
				continue
			}
		}

		newRule := models.HBARule{
			InstanceID:  dstID,
			RuleType:    rule.RuleType,
			Database:    rule.Database,
			DBUser:      rule.DBUser,
			Address:     rule.Address,
			AuthMethod:  rule.AuthMethod,
			Comment:     rule.Comment,
			RuleOrder:   rule.RuleOrder,
			Enabled:     rule.Enabled,
			IsTemporary: false,
			CreatedByID: &uid,
		}

		if err := h.db.Create(&newRule).Error; err == nil {
			cloned++
		}
	}

	// Write-back if requested
	if req.WriteBack && cloned > 0 {
		go func() {
			h.writer.WriteAndReload(dstID, &uid, claims.Username, c.ClientIP())
		}()
	}

	h.audit.Log(&uid, claims.Username, models.EventAddRule,
		fmt.Sprintf("clone from %s:%d to %s:%d", src.InstanceName, src.Port, dst.InstanceName, dst.Port),
		&dstID, fmt.Sprintf("%s:%d", dst.InstanceName, dst.Port), c.ClientIP(),
		map[string]interface{}{
			"source":   fmt.Sprintf("%s:%d", src.InstanceName, src.Port),
			"target":   fmt.Sprintf("%s:%d", dst.InstanceName, dst.Port),
			"cloned":   cloned,
			"skipped":  skipped,
			"overwrite": req.OverwriteExisting,
		})

	c.JSON(http.StatusOK, gin.H{
		"message":    "clone complete",
		"source":     fmt.Sprintf("%s:%d @ %s", src.InstanceName, src.Port, src.Host.Name),
		"target":     fmt.Sprintf("%s:%d @ %s", dst.InstanceName, dst.Port, dst.Host.Name),
		"cloned":     cloned,
		"skipped":    skipped,
		"write_back": req.WriteBack && cloned > 0,
	})
}
