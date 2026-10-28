package handlers

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pghba-manager/backend/internal/config"
	"github.com/pghba-manager/backend/internal/models"
	"gorm.io/gorm"
)

type SettingsHandler struct {
	db        *gorm.DB
	cfg       *config.Config
	startTime time.Time
}

func NewSettingsHandler(db *gorm.DB, cfg *config.Config) *SettingsHandler {
	return &SettingsHandler{db: db, cfg: cfg, startTime: time.Now()}
}

func (h *SettingsHandler) GetSystemInfo(c *gin.Context) {
	// Count stats
	var totalHosts, totalInstances, totalRules, totalUsers, totalBackups, totalAuditLogs int64
	h.db.Model(&models.DBHost{}).Count(&totalHosts)
	h.db.Model(&models.PGInstance{}).Count(&totalInstances)
	h.db.Model(&models.HBARule{}).Count(&totalRules)
	h.db.Model(&models.User{}).Count(&totalUsers)
	h.db.Model(&models.HBABackup{}).Count(&totalBackups)
	h.db.Model(&models.AuditLog{}).Count(&totalAuditLogs)

	// DB size
	var dbSize string
	h.db.Raw("SELECT pg_size_pretty(pg_database_size(current_database()))").Scan(&dbSize)

	uptime := time.Since(h.startTime)
	hours := int(uptime.Hours())
	minutes := int(uptime.Minutes()) % 60

	c.JSON(http.StatusOK, gin.H{
		"version":          "1.0.0",
		"go_version":       runtime.Version(),
		"go_os":            runtime.GOOS,
		"go_arch":          runtime.GOARCH,
		"uptime_seconds":   int(uptime.Seconds()),
		"uptime_label":     formatUptime(hours, minutes),
		"scheduler_interval":   h.cfg.SchedulerIntervalMin,
		"temp_access_action":   h.cfg.TempAccessAction,
		"auto_backup_enabled":  h.cfg.AutoBackupEnabled,
		"auto_backup_cron":     h.cfg.AutoBackupCron,
		"auto_backup_target":   h.cfg.AutoBackupTarget,
		"stats": gin.H{
			"hosts":      totalHosts,
			"instances":  totalInstances,
			"rules":      totalRules,
			"users":      totalUsers,
			"backups":    totalBackups,
			"audit_logs": totalAuditLogs,
			"db_size":    dbSize,
		},
	})
}

func (h *SettingsHandler) PurgeAuditLog(c *gin.Context) {
	type req struct {
		OlderThanDays int `json:"older_than_days"`
	}
	var r req
	c.ShouldBindJSON(&r)
	if r.OlderThanDays == 0 {
		r.OlderThanDays = 90
	}

	cutoff := time.Now().AddDate(0, 0, -r.OlderThanDays)
	result := h.db.Where("created_at < ?", cutoff).Delete(&models.AuditLog{})

	c.JSON(http.StatusOK, gin.H{
		"message":       "audit log purged",
		"deleted":       result.RowsAffected,
		"older_than":    cutoff.Format("2006-01-02"),
		"days":          r.OlderThanDays,
	})
}

func formatUptime(h, m int) string {
	if h == 0 {
		return "< 1 hour"
	}
	days := h / 24
	hrs  := h % 24
	if days > 0 {
		return itoa(days) + "d " + itoa(hrs) + "h " + itoa(m) + "m"
	}
	return itoa(hrs) + "h " + itoa(m) + "m"
}


// GetEOLStatus — return EOL status for all PG versions in use
func (h *SettingsHandler) GetEOLStatus(c *gin.Context) {
	pgEOLMap := map[string]string{
		"9.5": "2021-02-11", "9.6": "2021-11-11",
		"10":  "2022-11-10", "11":  "2023-11-09",
		"12":  "2024-11-14", "13":  "2025-11-13",
		"14":  "2026-11-12", "15":  "2027-11-11",
		"16":  "2028-11-09", "17":  "2029-11-08",
	}

	type VersionInfo struct {
		Version     string `json:"version"`
		EOLDate     string `json:"eol_date"`
		IsEOL       bool   `json:"is_eol"`
		DaysUntilEOL int   `json:"days_until_eol"`
		Instances   int64  `json:"instances"`
	}

	// Get all PG versions in use
	type Row struct {
		PgVersion string
		Count     int64
	}
	var rows []Row
	h.db.Raw(`SELECT pg_version, COUNT(*) as count FROM pg_instances GROUP BY pg_version`).Scan(&rows)

	var versions []VersionInfo
	for _, row := range rows {
		eolDate := pgEOLMap[row.PgVersion]
		isEOL   := false
		days    := 0
		if eolDate != "" {
			t, _ := time.Parse("2006-01-02", eolDate)
			diff  := time.Until(t)
			days   = int(diff.Hours() / 24)
			isEOL  = diff < 0
		}
		versions = append(versions, VersionInfo{
			Version:      row.PgVersion,
			EOLDate:      eolDate,
			IsEOL:        isEOL,
			DaysUntilEOL: days,
			Instances:    row.Count,
		})
	}

	// Count EOL instances
	eolCount := 0
	for _, v := range versions {
		if v.IsEOL {
			eolCount += int(v.Instances)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"versions":      versions,
		"eol_count":     eolCount,
		"has_eol":       eolCount > 0,
	})
}
