package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pghba-manager/backend/internal/config"
	"github.com/pghba-manager/backend/internal/models"
	"github.com/pghba-manager/backend/internal/services"
	"gorm.io/gorm"
)

type HealthCheckHandler struct {
	db       *gorm.DB
	cfg      *config.Config
	crypto   *services.CryptoService
	notifier *services.Notifier
	audit    *services.AuditService
}

func NewHealthCheckHandler(db *gorm.DB, cfg *config.Config, crypto *services.CryptoService, notifier *services.Notifier, audit *services.AuditService) *HealthCheckHandler {
	return &HealthCheckHandler{db: db, cfg: cfg, crypto: crypto, notifier: notifier, audit: audit}
}

type HostHealth struct {
	HostID      string    `json:"host_id"`
	HostName    string    `json:"host_name"`
	IPAddress   string    `json:"ip_address"`
	Environment string    `json:"environment"`
	Status      string    `json:"status"` // "online" | "offline" | "degraded"
	Instances   int       `json:"instances"`
	Online      int       `json:"online"`
	Offline     int       `json:"offline"`
	ResponseMs  int64     `json:"response_ms"`
	LastCheck   time.Time `json:"last_check"`
	Error       string    `json:"error,omitempty"`
	Changed     bool      `json:"changed"` // true if status changed since last check
}

// Check — run health check on all hosts and return results
func (h *HealthCheckHandler) Check(c *gin.Context) {
	var hosts []models.DBHost
	h.db.Find(&hosts)

	results := []HostHealth{}
	offlineCount := 0
	onlineCount := 0

	for _, host := range hosts {
		hh := h.checkHost(host)
		results = append(results, hh)
		if hh.Status == "online" {
			onlineCount++
		} else {
			offlineCount++
		}
	}

	// Update instance statuses in DB
	go h.updateInstanceStatuses()

	c.JSON(http.StatusOK, gin.H{
		"hosts":      results,
		"total":      len(hosts),
		"online":     onlineCount,
		"offline":    offlineCount,
		"checked_at": time.Now(),
	})
}

// CheckHost — check single host
func (h *HealthCheckHandler) CheckHost(c *gin.Context) {
	hostID := c.Param("id")
	var host models.DBHost
	if err := h.db.First(&host, "id = ?", hostID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "host not found"})
		return
	}
	result := h.checkHost(host)
	c.JSON(http.StatusOK, result)
}

func (h *HealthCheckHandler) checkHost(host models.DBHost) HostHealth {
	hh := HostHealth{
		HostID:      host.ID.String(),
		HostName:    host.Name,
		IPAddress:   host.IPAddress,
		Environment: string(host.Environment),
		LastCheck:   time.Now(),
	}

	// Get instance counts
	var total, online int64
	h.db.Model(&models.PGInstance{}).Where("host_id = ?", host.ID).Count(&total)
	h.db.Model(&models.PGInstance{}).Where("host_id = ? AND status = 'Online'", host.ID).Count(&online)
	hh.Instances = int(total)
	hh.Online = int(online)
	hh.Offline = int(total) - int(online)

	if host.SSHKey == "" {
		hh.Status = "unknown"
		hh.Error = "SSH key not configured"
		return hh
	}

	// Test SSH connection
	privateKey, err := h.crypto.Decrypt(host.SSHKey)
	if err != nil {
		hh.Status = "unknown"
		hh.Error = "failed to decrypt SSH key"
		return hh
	}

	start := time.Now()
	sshPort := host.SSHPort
	if sshPort == 0 {
		sshPort = 22
	}

	client, err := services.ConnectToHost(h.db, host.ID, host.IPAddress, sshPort, host.SSHUser, privateKey, host.SSHHostKeyFingerprint)
	hh.ResponseMs = time.Since(start).Milliseconds()

	if err == nil {
		// Jangan pernah health check sebagai OS user postgres:
		// tiap buka-tutup sesi postgres + RemoveIPC=yes = IPC/shm
		// disapu saat sesi terakhir logout (insiden 2026-10-23).
		if perr := client.RequireNonPostgresUser(); perr != nil {
			client.Close()
			hh.Status = "unknown"
			hh.Error = perr.Error()
			return hh
		}
	}

	if err != nil {
		hh.Status = "offline"
		hh.Error = err.Error()

		// Check if status changed (was online before)
		prevOnline := online > 0
		if prevOnline {
			hh.Changed = true
			// Mark all instances as offline
			h.db.Model(&models.PGInstance{}).
				Where("host_id = ?", host.ID).
				Update("status", "Offline")
			// Send alert
			go h.sendOfflineAlert(host, err.Error())
			log.Printf("[healthcheck] HOST OFFLINE: %s (%s) — %s", host.Name, host.IPAddress, err.Error())
		}
		return hh
	}
	defer client.Close()

	// Run quick command to verify
	out, cmdErr := client.RunCommand("uptime 2>/dev/null || echo 'ok'")
	if cmdErr != nil {
		hh.Status = "degraded"
		hh.Error = "SSH connected but command failed"
		return hh
	}

	hh.Status = "online"
	_ = out

	// If was offline before, mark recovery
	if online < total && total > 0 {
		hh.Changed = true
		// Trigger scan to update instance statuses
		go h.scanHostInstances(host, privateKey)
	}

	return hh
}

func (h *HealthCheckHandler) sendOfflineAlert(host models.DBHost, errMsg string) {
	if !h.cfg.NotifyEnabled {
		return
	}

	alerts := []services.TempAccessAlert{{
		HostName:      host.Name,
		InstanceLabel: fmt.Sprintf("%s (%s)", host.Name, host.IPAddress),
		Reason:        fmt.Sprintf("HOST OFFLINE: %s", errMsg),
	}}
	_ = alerts

	// Send via Telegram
	if h.cfg.TelegramToken != "" && h.cfg.TelegramChatID != "" {
		msg := fmt.Sprintf(
			"🔴 <b>PGHBA Manager — Host Offline</b>\n\nHost: <code>%s</code>\nIP: <code>%s</code>\nEnv: %s\nError: %s\nTime: %s",
			host.Name, host.IPAddress, host.Environment, errMsg,
			time.Now().Format("2006-01-02 15:04:05"),
		)
		h.notifier.SendRaw(msg)
	}
}

func (h *HealthCheckHandler) scanHostInstances(host models.DBHost, privateKey string) {
	if host.SSHUser == "postgres" {
		log.Printf("[healthcheck] SKIP recovery scan %s: ssh_user=postgres berisiko IPC cleanup — ganti ke user non-postgres", host.Name)
		return
	}
	client, err := services.ConnectToHost(h.db, host.ID, host.IPAddress, host.SSHPort, host.SSHUser, privateKey, host.SSHHostKeyFingerprint)
	if err != nil {
		return
	}
	defer client.Close()

	// pg_lsclusters baca config postgres — via sudo agar jalan dari user
	// non-postgres.
	out, _ := client.RunAsPostgres("pg_lsclusters --no-header 2>/dev/null || echo ''")
	if out != "" {
		// Mark all as online
		h.db.Model(&models.PGInstance{}).Where("host_id = ?", host.ID).Update("status", "Online")
		log.Printf("[healthcheck] HOST RECOVERED: %s — instances back online", host.Name)
	}
}

func (h *HealthCheckHandler) updateInstanceStatuses() {
	// Quick check — update last_scan timestamp for all online instances
	h.db.Exec("UPDATE pg_instances SET last_scan = NOW() WHERE status = 'Online'")
}

// GetHistory — get health check history from audit log
func (h *HealthCheckHandler) GetHistory(c *gin.Context) {
	var logs []models.AuditLog
	h.db.Where("event = 'HealthCheck' OR (target LIKE '%offline%' AND event = 'ReloadConf')").
		Order("created_at DESC").Limit(50).Find(&logs)
	c.JSON(http.StatusOK, gin.H{"history": logs})
}
