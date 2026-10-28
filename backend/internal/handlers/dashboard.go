package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pghba-manager/backend/internal/models"
	"gorm.io/gorm"
)

type DashboardHandler struct {
	db *gorm.DB
}

func NewDashboardHandler(db *gorm.DB) *DashboardHandler {
	return &DashboardHandler{db: db}
}

func (h *DashboardHandler) Stats(c *gin.Context) {
	env := c.Query("environment") // optional filter

	var totalHosts, totalInstances, totalRules, tempRules, onlineInstances int64

	hostQ := h.db.Model(&models.DBHost{}).Where("active = true")
	instQ := h.db.Model(&models.PGInstance{})
	if env != "" {
		hostQ = hostQ.Where("environment = ?", env)
		instQ = instQ.Joins("JOIN db_hosts ON db_hosts.id = pg_instances.host_id").
			Where("db_hosts.environment = ?", env)
	}

	hostQ.Count(&totalHosts)
	instQ.Count(&totalInstances)

	onlineQ := h.db.Model(&models.PGInstance{}).Where("status = 'Online'")
	if env != "" {
		onlineQ = onlineQ.Joins("JOIN db_hosts ON db_hosts.id = pg_instances.host_id").
			Where("db_hosts.environment = ?", env)
	}
	onlineQ.Count(&onlineInstances)

	ruleQ := h.db.Model(&models.HBARule{})
	if env != "" {
		ruleQ = ruleQ.Joins("JOIN pg_instances ON pg_instances.id = hba_rules.instance_id").
			Joins("JOIN db_hosts ON db_hosts.id = pg_instances.host_id").
			Where("db_hosts.environment = ?", env)
	}
	ruleQ.Count(&totalRules)

	h.db.Model(&models.TempAccess{}).
		Where("revoked = false AND expires_at > ?", time.Now()).
		Count(&tempRules)

	// Per-environment stats
	type envStat struct {
		Environment string `json:"environment"`
		Hosts       int64  `json:"hosts"`
		Instances   int64  `json:"instances"`
		Online      int64  `json:"online"`
		Rules       int64  `json:"rules"`
	}
	var envStats []envStat
	h.db.Raw(`
		SELECT 
			h.environment,
			COUNT(DISTINCT h.id) as hosts,
			COUNT(DISTINCT i.id) as instances,
			COUNT(DISTINCT CASE WHEN i.status = 'Online' THEN i.id END) as online,
			COUNT(DISTINCT r.id) as rules
		FROM db_hosts h
		LEFT JOIN pg_instances i ON i.host_id = h.id
		LEFT JOIN hba_rules r ON r.instance_id = i.id
		WHERE h.active = true
		GROUP BY h.environment
		ORDER BY h.environment
	`).Scan(&envStats)

	// Version distribution
	type versionCount struct {
		PGVersion string `json:"pg_version"`
		Count     int64  `json:"count"`
	}
	var versionDist []versionCount
	vq := h.db.Model(&models.PGInstance{}).
		Select("pg_version, COUNT(*) as count").
		Group("pg_version").
		Order("pg_version DESC")
	if env != "" {
		vq = vq.Joins("JOIN db_hosts ON db_hosts.id = pg_instances.host_id").
			Where("db_hosts.environment = ?", env)
	}
	vq.Scan(&versionDist)

	// Recent audit logs
	var recentLogs []models.AuditLog
	h.db.Order("created_at DESC").Limit(10).Find(&recentLogs)

	c.JSON(http.StatusOK, gin.H{
		"total_hosts":       totalHosts,
		"total_instances":   totalInstances,
		"online_instances":  onlineInstances,
		"offline_instances": totalInstances - onlineInstances,
		"total_rules":       totalRules,
		"temp_rules_active": tempRules,
		"version_dist":      versionDist,
		"env_stats":         envStats,
		"recent_activity":   recentLogs,
		"environment":       env,
	})
}
