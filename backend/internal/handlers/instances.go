package handlers

import (
	"fmt"
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

type InstanceHandler struct {
	db     *gorm.DB
	audit  *services.AuditService
	crypto *services.CryptoService
}

func NewInstanceHandler(db *gorm.DB, audit *services.AuditService, crypto *services.CryptoService) *InstanceHandler {
	return &InstanceHandler{db: db, audit: audit, crypto: crypto}
}

func (h *InstanceHandler) List(c *gin.Context) {
	var instances []models.PGInstance
	q := h.db.Preload("Host")
	if hid := c.Query("host_id"); hid != "" {
		q = q.Where("host_id = ?", hid)
	}
	if s := c.Query("status"); s != "" {
		q = q.Where("status = ?", s)
	}
	q.Order("host_id, port").Find(&instances)
	c.JSON(http.StatusOK, instances)
}

func (h *InstanceHandler) Get(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var inst models.PGInstance
	if err := h.db.Preload("Host").First(&inst, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
		return
	}
	c.JSON(http.StatusOK, inst)
}

type createInstanceReq struct {
	HostID        string `json:"host_id"        binding:"required"`
	InstanceName  string `json:"instance_name"  binding:"required"`
	PGVersion     string `json:"pg_version"     binding:"required"`
	Port          int    `json:"port"           binding:"required"`
	DataDirectory string `json:"data_directory" binding:"required"`
	HBAPath       string `json:"hba_path"       binding:"required"`
}

func (h *InstanceHandler) Create(c *gin.Context) {
	var req createInstanceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	hid, _ := uuid.Parse(req.HostID)
	inst := &models.PGInstance{
		HostID:        hid,
		InstanceName:  req.InstanceName,
		PGVersion:     req.PGVersion,
		Port:          req.Port,
		DataDirectory: req.DataDirectory,
		HBAPath:       req.HBAPath,
		Status:        models.StatusUnknown,
	}
	if err := h.db.Create(inst).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "instance already exists on this host/port"})
		return
	}
	claims := middleware.GetClaims(c)
	label := fmt.Sprintf("%s:%d", inst.InstanceName, inst.Port)
	h.audit.Log(&claims.UserID, claims.Username, models.EventAddInstance,
		inst.InstanceName, &inst.ID, label, c.ClientIP(), nil)
	c.JSON(http.StatusCreated, inst)
}

// Discover — SSH ke host, jalankan semua lsclusters commands, upsert instances
func (h *InstanceHandler) Discover(c *gin.Context) {
	type req struct {
		HostID string `json:"host_id" binding:"required"`
	}
	var r req
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hid, _ := uuid.Parse(r.HostID)
	var host models.DBHost
	if err := h.db.First(&host, "id = ?", hid).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "host not found"})
		return
	}

	if host.SSHKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "SSH key belum dikonfigurasi untuk host ini. Edit host dan tambahkan SSH private key.",
		})
		return
	}

	// Decrypt SSH key
	privateKey, err := h.crypto.Decrypt(host.SSHKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decrypt SSH key"})
		return
	}

	// Buka koneksi SSH (discovery membaca config postgres — wajib user
	// non-postgres; sesi postgres + RemoveIPC=yes = IPC/shm disapu)
	sshClient, err := services.ConnectToHost(h.db, host.ID, host.IPAddress, host.SSHPort, host.SSHUser, privateKey, host.SSHHostKeyFingerprint)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":  "SSH connection failed",
			"detail": err.Error(),
			"hint":   fmt.Sprintf("Pastikan public key sudah ada di %s@%s:~/.ssh/authorized_keys", host.SSHUser, host.IPAddress),
		})
		return
	}
	defer sshClient.Close()
	if perr := sshClient.RequireNonPostgresUser(); perr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  perr.Error(),
			"detail": fmt.Sprintf("Host %s masih memakai ssh_user=postgres. Ganti ke user non-postgres (mis. dco) dengan sudo NOPASSWD untuk postgres.", host.Name),
		})
		return
	}

	// Discover semua clusters (pg_lsclusters + pg16-lsclusters + fallback)
	clusters, err := sshClient.DiscoverAllClusters()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":  "failed to discover clusters",
			"detail": err.Error(),
		})
		return
	}

	if len(clusters) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message":  "tidak ada cluster ditemukan di host ini",
			"host":     host.Name,
			"upserted": 0,
		})
		return
	}

	claims := middleware.GetClaims(c)
	now := time.Now()
	upserted := 0
	var results []gin.H

	for _, cl := range clusters {
		// Tentukan status
		instStatus := models.StatusOffline
		if strings.Contains(strings.ToLower(cl.Status), "online") {
			instStatus = models.StatusOnline
		}

		var existing models.PGInstance
		err := h.db.Where("host_id = ? AND port = ?", hid, cl.Port).First(&existing).Error

		if err == gorm.ErrRecordNotFound {
			// Insert baru
			newInst := &models.PGInstance{
				HostID:        hid,
				InstanceName:  cl.ClusterName,
				PGVersion:     cl.Version,
				Port:          cl.Port,
				DataDirectory: cl.DataDir,
				HBAPath:       cl.HBAPath,
				Status:        instStatus,
				LastScan:      &now,
			}
			if createErr := h.db.Create(newInst).Error; createErr == nil {
				upserted++
				results = append(results, gin.H{
					"action":   "created",
					"instance": cl.ClusterName,
					"version":  cl.Version,
					"port":     cl.Port,
					"status":   instStatus,
					"hba_path": cl.HBAPath,
				})
				h.audit.Log(&claims.UserID, claims.Username, models.EventAddInstance,
					cl.ClusterName, &newInst.ID,
					fmt.Sprintf("%s:%d", cl.ClusterName, cl.Port),
					c.ClientIP(), nil)
			}
		} else if err == nil {
			// Update existing
			h.db.Model(&existing).Updates(map[string]interface{}{
				"instance_name":  cl.ClusterName,
				"pg_version":     cl.Version,
				"data_directory": cl.DataDir,
				"hba_path":       cl.HBAPath,
				"status":         instStatus,
				"last_scan":      now,
			})
			upserted++
			results = append(results, gin.H{
				"action":   "updated",
				"instance": cl.ClusterName,
				"version":  cl.Version,
				"port":     cl.Port,
				"status":   instStatus,
				"hba_path": cl.HBAPath,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "discovery complete",
		"host":     host.Name,
		"upserted": upserted,
		"clusters": results,
	})
}

// Scan — cek status satu instance via TCP + SSH fallback
func (h *InstanceHandler) Scan(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var inst models.PGInstance
	if err := h.db.Preload("Host").First(&inst, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
		return
	}

	now := time.Now()
	newStatus := models.StatusOffline
	method := "tcp"

	if services.CheckTCPPort(inst.Host.IPAddress, inst.Port, 5*time.Second) {
		newStatus = models.StatusOnline
	} else if inst.Host.SSHKey != "" {
		privateKey, err := h.crypto.Decrypt(inst.Host.SSHKey)
		if err == nil {
			sshClient, err := services.ConnectToHost(
				h.db, inst.Host.ID, inst.Host.IPAddress, inst.Host.SSHPort, inst.Host.SSHUser, privateKey, inst.Host.SSHHostKeyFingerprint,
			)
			if err == nil {
				defer sshClient.Close()
				method = "ssh"
				// pg_lsclusters baca config postgres — via sudo.
				out, err := sshClient.RunAsPostgres(
					fmt.Sprintf("pg_lsclusters --no-header 2>/dev/null | awk '$3==%d{print $4}'", inst.Port),
				)
				if err == nil && strings.Contains(out, "online") {
					newStatus = models.StatusOnline
				}
			}
		}
	}

	h.db.Model(&inst).Updates(map[string]interface{}{
		"status":    newStatus,
		"last_scan": now,
	})

	c.JSON(http.StatusOK, gin.H{
		"instance":   inst.InstanceName,
		"port":       inst.Port,
		"host":       inst.Host.IPAddress,
		"status":     newStatus,
		"method":     method,
		"scanned_at": now.Format(time.RFC3339),
	})
}

// ScanAll — scan semua instances untuk satu host
func (h *InstanceHandler) ScanAll(c *gin.Context) {
	type req struct {
		HostID string `json:"host_id" binding:"required"`
	}
	var r req
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hid, _ := uuid.Parse(r.HostID)
	var host models.DBHost
	h.db.First(&host, "id = ?", hid)

	var instances []models.PGInstance
	h.db.Where("host_id = ?", hid).Find(&instances)

	now := time.Now()
	results := make([]gin.H, 0, len(instances))

	// Buka satu koneksi SSH untuk semua instances
	var sshClient *services.SSHClient
	if host.SSHKey != "" {
		privateKey, err := h.crypto.Decrypt(host.SSHKey)
		if err == nil {
			sshClient, _ = services.ConnectToHost(
				h.db, host.ID, host.IPAddress, host.SSHPort, host.SSHUser, privateKey, host.SSHHostKeyFingerprint,
			)
		}
	}
	if sshClient != nil {
		defer sshClient.Close()
		// ScanAll fallback SSH hanya untuk user non-postgres — sesi
		// postgres + RemoveIPC=yes = IPC/shm disapu saat logout.
		if perr := sshClient.RequireNonPostgresUser(); perr != nil {
			log.Printf("[scanall] SSH fallback dinonaktifkan untuk host %s: %s", host.Name, perr.Error())
			sshClient.Close()
			sshClient = nil
		}
	}

	for _, inst := range instances {
		status := models.StatusOffline

		if services.CheckTCPPort(host.IPAddress, inst.Port, 3*time.Second) {
			status = models.StatusOnline
		} else if sshClient != nil {
			// pg_lsclusters baca config postgres — via sudo.
			out, err := sshClient.RunAsPostgres(
				fmt.Sprintf("{ pg_lsclusters --no-header 2>/dev/null; pg16-lsclusters 2>/dev/null; } | grep -w %d | grep -i online", inst.Port),
			)
			if err == nil && strings.TrimSpace(out) != "" {
				status = models.StatusOnline
			}
		}

		h.db.Model(&inst).Updates(map[string]interface{}{
			"status":    status,
			"last_scan": now,
		})

		results = append(results, gin.H{
			"instance": inst.InstanceName,
			"port":     inst.Port,
			"status":   status,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"host":       host.Name,
		"scanned":    len(results),
		"results":    results,
		"scanned_at": now.Format(time.RFC3339),
	})
}

// ImportHBA — baca pg_hba.conf dari server via SSH, parse dan import ke DB
