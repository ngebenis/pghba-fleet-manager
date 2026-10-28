package handlers

import (
	"log"
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

type TempAccessHandler struct {
	db     *gorm.DB
	audit  *services.AuditService
	writer *services.HBAWriter
}

func NewTempAccessHandler(db *gorm.DB, audit *services.AuditService, crypto *services.CryptoService) *TempAccessHandler {
	return &TempAccessHandler{
		db:     db,
		audit:  audit,
		writer: services.NewHBAWriter(db, audit, crypto),
	}
}

func (h *TempAccessHandler) List(c *gin.Context) {
	query := `
		SELECT
			ta.id, ta.rule_id, ta.instance_id,
			ta.granted_by_id as granted_by,
			ta.duration_value, ta.duration_unit,
			ta.expires_at, ta.reason, ta.revoked,
			ta.revoked_by_id::text as revoked_by,
			ta.revoked_at, ta.created_at,
			COALESCE(r.address::text, '')     as ip_address,
			COALESCE(r.database, 'all')       as database,
			COALESCE(r.db_user, 'all')        as db_user,
			COALESCE(r.auth_method::text, 'md5') as auth_method,
			COALESCE(r.enabled, false)        as enabled,
			COALESCE(u.username, '')          as granted_by_username,
			COALESCE(u.full_name, '')         as granted_by_full_name
		FROM temp_access ta
		LEFT JOIN hba_rules r ON r.id = ta.rule_id
		LEFT JOIN users u     ON u.id = ta.granted_by_id
	`
	args := []interface{}{}
	if iid := c.Query("instance_id"); iid != "" {
		query += " WHERE ta.instance_id = ?"
		args = append(args, iid)
	}
	query += " ORDER BY ta.created_at DESC"

	type Row struct {
		ID                  string     `json:"id"`
		RuleID              string     `json:"rule_id"`
		InstanceID          string     `json:"instance_id"`
		GrantedBy           string     `json:"granted_by"`
		DurationValue       int        `json:"duration_value"`
		DurationUnit        string     `json:"duration_unit"`
		ExpiresAt           time.Time  `json:"expires_at"`
		Reason              string     `json:"reason"`
		Revoked             bool       `json:"revoked"`
		RevokedBy           *string    `json:"revoked_by"`
		RevokedAt           *time.Time `json:"revoked_at"`
		CreatedAt           time.Time  `json:"created_at"`
		IPAddress           string     `json:"ip_address"`
		Database            string     `json:"database"`
		DBUser              string     `json:"db_user"`
		AuthMethod          string     `json:"auth_method"`
		Enabled             bool       `json:"enabled"`
		GrantedByUsername   string     `json:"granted_by_username"`
		GrantedByFullName   string     `json:"granted_by_full_name"`
	}

	var rows []Row
	if err := h.db.Raw(query, args...).Scan(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if rows == nil {
		rows = []Row{}
	}
	c.JSON(http.StatusOK, rows)
}

type grantTempReq struct {
	InstanceID    string                  `json:"instance_id"    binding:"required"`
	IPAddress     string                  `json:"ip_address"     binding:"required"`
	Database      string                  `json:"database"`
	DBUser        string                  `json:"db_user"`
	AuthMethod    models.AuthMethod       `json:"auth_method"`
	DurationValue int                     `json:"duration_value" binding:"required,min=1"`
	DurationUnit  models.TempDurationUnit `json:"duration_unit"  binding:"required"`
	Reason        string                  `json:"reason"`
}

func (h *TempAccessHandler) Grant(c *gin.Context) {
	var req grantTempReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	claims := middleware.GetClaims(c)
	iid, _ := uuid.Parse(req.InstanceID)

	db := req.Database
	if db == "" { db = "all" }
	usr := req.DBUser
	if usr == "" { usr = "all" }
	method := req.AuthMethod
	if method == "" { method = models.MethodMD5 }

	address := req.IPAddress
	if !strings.Contains(address, "/") {
		address = address + "/32"
	}

	expiresAt := calcExpiry(req.DurationValue, req.DurationUnit)
	uid := claims.UserID

	rule := &models.HBARule{
		InstanceID:           iid,
		RuleType:             models.TypeHost,
		Database:             db,
		DBUser:               usr,
		Address:              address,
		AuthMethod:           method,
		Comment:              "Temporary access — expires " + expiresAt.UTC().Format(time.RFC3339),
		Enabled:              true,
		IsTemporary:          true,
		CreatedForTempAccess: true,
		CreatedByID:          &uid,
	}
	if err := h.db.Create(rule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create rule: " + err.Error()})
		return
	}

	ta := &models.TempAccess{
		RuleID:        rule.ID,
		InstanceID:    iid,
		GrantedByID:   uid,
		DurationValue: req.DurationValue,
		DurationUnit:  req.DurationUnit,
		ExpiresAt:     expiresAt,
		Reason:        req.Reason,
	}
	if err := h.db.Create(ta).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create temp access: " + err.Error()})
		return
	}

	// Write-back pg_hba.conf ke server via SSH + reload PostgreSQL
	go func() {
		if err := h.writer.WriteAndReload(iid, &uid, claims.Username, c.ClientIP()); err != nil {
			log.Printf("[temp_access] write-back failed for instance %s: %v", iid, err)
		} else {
			log.Printf("[temp_access] pg_hba.conf updated after grant for instance %s", iid)
		}
	}()

	h.audit.Log(&uid, claims.Username, models.EventGrantTempAccess,
		address, &iid, "", c.ClientIP(),
		map[string]interface{}{"duration": req.DurationValue, "unit": req.DurationUnit, "expires_at": expiresAt})

	c.JSON(http.StatusCreated, gin.H{
		"id":          ta.ID,
		"rule_id":     rule.ID,
		"instance_id": iid,
		"ip_address":  address,
		"database":    db,
		"db_user":     usr,
		"auth_method": method,
		"expires_at":  expiresAt,
		"reason":      req.Reason,
		"revoked":     false,
	})
}

func (h *TempAccessHandler) Revoke(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var ta models.TempAccess
	if err := h.db.First(&ta, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	claims := middleware.GetClaims(c)
	now := time.Now()
	h.db.Model(&ta).Updates(map[string]interface{}{
		"revoked":       true,
		"revoked_by_id": claims.UserID,
		"revoked_at":    now,
	})
	// Disable rule
	h.db.Model(&models.HBARule{}).Where("id = ?", ta.RuleID).Update("enabled", false)

	// Write-back pg_hba.conf setelah revoke
	uid := claims.UserID
	go func() {
		if err := h.writer.WriteAndReload(ta.InstanceID, &uid, claims.Username, c.ClientIP()); err != nil {
			log.Printf("[temp_access] write-back failed on revoke for instance %s: %v", ta.InstanceID, err)
		}
	}()

	h.audit.Log(&claims.UserID, claims.Username, models.EventRevokeTempAccess,
		ta.RuleID.String(), &ta.InstanceID, "", c.ClientIP(), nil)
	c.JSON(http.StatusOK, gin.H{"message": "temp access revoked"})
}

func calcExpiry(value int, unit models.TempDurationUnit) time.Time {
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
