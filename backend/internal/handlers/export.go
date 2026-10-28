package handlers

import (
	"archive/zip"
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pghba-manager/backend/internal/config"
	"github.com/pghba-manager/backend/internal/middleware"
	"github.com/pghba-manager/backend/internal/models"
	"github.com/pghba-manager/backend/internal/services"
	"gorm.io/gorm"
)

// safePathSegment neutralizes a single path segment before it's joined
// into a filesystem path or storage key. instance_name is either typed
// directly by a user creating an instance, or copied verbatim from a
// PostgreSQL cluster name discovered on a managed host (Discover) — either
// way it must never be trusted to stay inside one path segment: a name
// like "../../etc" would otherwise let backup export write outside
// BackupStoragePath via filepath.Join.
func safePathSegment(s string) string {
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")
	s = strings.ReplaceAll(s, "..", "_")
	s = strings.TrimSpace(s)
	if s == "" || s == "." {
		return "unknown"
	}
	return s
}

type ExportHandler struct {
	db    *gorm.DB
	cfg   *config.Config
	audit *services.AuditService
}

func NewExportHandler(db *gorm.DB, cfg *config.Config, audit *services.AuditService) *ExportHandler {
	return &ExportHandler{db: db, cfg: cfg, audit: audit}
}

// ExportConfig — export all config as JSON/YAML for DR
func (h *ExportHandler) ExportConfig(c *gin.Context) {
	format := c.DefaultQuery("format", "json")
	claims := middleware.GetClaims(c)

	type ExportData struct {
		ExportedAt time.Time           `json:"exported_at"`
		ExportedBy string              `json:"exported_by"`
		Version    string              `json:"version"`
		Hosts      []models.DBHost     `json:"hosts"`
		Instances  []models.PGInstance `json:"instances"`
		Rules      []models.HBARule    `json:"rules"`
		Users      []models.User       `json:"users"`
		Summary    map[string]int      `json:"summary"`
	}

	var hosts []models.DBHost
	var instances []models.PGInstance
	var rules []models.HBARule
	var users []models.User

	h.db.Find(&hosts)
	h.db.Preload("Host").Find(&instances)
	h.db.Preload("Instance").Preload("Instance.Host").
		Where("is_temporary = false").Find(&rules)
	h.db.Find(&users)

	// Mask sensitive data
	for i := range hosts {
		hosts[i].SSHKey = ""
	}
	for i := range users {
		users[i].Password = ""
	}

	data := ExportData{
		ExportedAt: time.Now(),
		ExportedBy: claims.Username,
		Version:    "1.0.0",
		Hosts:      hosts,
		Instances:  instances,
		Rules:      rules,
		Users:      users,
		Summary: map[string]int{
			"hosts": len(hosts), "instances": len(instances),
			"rules": len(rules), "users": len(users),
		},
	}

	h.audit.Log(&claims.UserID, claims.Username, "ExportConfig",
		fmt.Sprintf("export %s", format), nil, "", c.ClientIP(),
		map[string]interface{}{"format": format, "hosts": len(hosts), "rules": len(rules)})

	ts := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("pghba-config-%s.%s", ts, format)

	if format == "yaml" {
		out, _ := json.MarshalIndent(data, "", "  ")
		yaml := "# PGHBA Manager Configuration Export\n# Generated: " +
			time.Now().Format("2006-01-02 15:04:05") + "\n# Hosts: " +
			fmt.Sprintf("%d", len(hosts)) + " | Instances: " + fmt.Sprintf("%d", len(instances)) +
			" | Rules: " + fmt.Sprintf("%d", len(rules)) + "\n\n" + string(out)
		c.Header("Content-Disposition", "attachment; filename="+filename)
		c.Header("Content-Type", "application/x-yaml")
		c.String(http.StatusOK, yaml)
		return
	}

	out, _ := json.MarshalIndent(data, "", "  ")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/json")
	c.Data(http.StatusOK, "application/json", out)
}

// ExportBackups — export all backups as zip
func (h *ExportHandler) ExportBackups(c *gin.Context) {
	claims := middleware.GetClaims(c)
	instanceID := c.Query("instance_id")

	var backups []models.HBABackup
	q := h.db.Preload("Instance")
	if instanceID != "" {
		q = q.Where("instance_id = ?", instanceID)
	}
	q.Order("created_at DESC").Limit(200).Find(&backups)

	if len(backups) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "no backups found"})
		return
	}

	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	for _, bk := range backups {
		fname := safePathSegment(bk.Filename)
		if bk.Filename == "" {
			fname = fmt.Sprintf("backup-%s.bak", bk.ID.String()[:8])
		}
		// Group by instance name in zip — sanitized so a crafted
		// instance_name can't zip-slip out of the extraction directory
		// for whoever opens this archive.
		if bk.Instance.InstanceName != "" {
			fname = safePathSegment(bk.Instance.InstanceName) + "/" + fname
		}
		w, err := zw.Create(fname)
		if err != nil {
			continue
		}
		w.Write([]byte(bk.Content))
	}
	zw.Close()

	ts := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("pghba-backups-%s.zip", ts)

	h.audit.Log(&claims.UserID, claims.Username, "ExportBackups",
		filename, nil, "", c.ClientIP(),
		map[string]interface{}{"count": len(backups)})

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/zip")
	c.Data(http.StatusOK, "application/zip", buf.Bytes())
}

// SaveBackupToStorage — save to local folder or S3
func (h *ExportHandler) SaveBackupToStorage(c *gin.Context) {
	claims := middleware.GetClaims(c)

	type req struct {
		BackupID string `json:"backup_id"`
		Target   string `json:"target"` // "local" | "s3"
	}
	var r req
	c.ShouldBindJSON(&r)
	if r.Target == "" {
		r.Target = "local"
	}

	var backups []models.HBABackup
	q := h.db.Preload("Instance")
	if r.BackupID != "" {
		q = q.Where("id = ?", r.BackupID)
	} else {
		q = q.Order("created_at DESC").Limit(200)
	}
	q.Find(&backups)

	if len(backups) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "no backups found"})
		return
	}

	results := []map[string]interface{}{}

	if r.Target == "s3" {
		if !h.cfg.S3Enabled || h.cfg.S3Endpoint == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "S3 not configured. Set S3_ENABLED, S3_ENDPOINT, S3_BUCKET, S3_ACCESS_KEY, S3_SECRET_KEY in .env",
			})
			return
		}
		for _, bk := range backups {
			instName := safePathSegment(bk.Instance.InstanceName)
			key := fmt.Sprintf("%s/backups/%s/%s", h.cfg.S3Prefix, instName, safePathSegment(bk.Filename))
			err := s3PutObject(h.cfg, key, []byte(bk.Content))
			status, errMsg := "success", ""
			if err != nil {
				status = "failed"
				errMsg = err.Error()
			}
			results = append(results, map[string]interface{}{
				"filename": bk.Filename, "instance": instName,
				"status": status, "key": key, "error": errMsg,
			})
		}
	} else {
		// Local
		storagePath := h.cfg.BackupStoragePath
		if err := os.MkdirAll(storagePath, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot create storage dir: " + err.Error()})
			return
		}
		for _, bk := range backups {
			instName := safePathSegment(bk.Instance.InstanceName)
			dir := filepath.Join(storagePath, instName)
			os.MkdirAll(dir, 0755)
			fpath := filepath.Join(dir, safePathSegment(bk.Filename))
			err := os.WriteFile(fpath, []byte(bk.Content), 0644)
			status, errMsg := "success", ""
			if err != nil {
				status = "failed"
				errMsg = err.Error()
			}
			results = append(results, map[string]interface{}{
				"filename": bk.Filename, "instance": instName,
				"status": status, "path": fpath, "error": errMsg,
			})
		}
	}

	saved := 0
	for _, res := range results {
		if res["status"] == "success" {
			saved++
		}
	}

	h.audit.Log(&claims.UserID, claims.Username, "ExportBackups",
		fmt.Sprintf("save to %s", r.Target), nil, "", c.ClientIP(),
		map[string]interface{}{"target": r.Target, "saved": saved, "total": len(backups)})

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("saved %d/%d backups to %s", saved, len(backups), r.Target),
		"target":  r.Target,
		"saved":   saved,
		"total":   len(backups),
		"results": results,
	})
}

// GetStorageStatus — check storage config
func (h *ExportHandler) GetStorageStatus(c *gin.Context) {
	localPath := h.cfg.BackupStoragePath
	localExists := false
	localFiles := 0
	if info, err := os.Stat(localPath); err == nil && info.IsDir() {
		localExists = true
		filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() && strings.HasSuffix(path, ".bak") {
				localFiles++
			}
			return nil
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"local": gin.H{
			"path": localPath, "exists": localExists,
			"files": localFiles, "enabled": true,
		},
		"s3": gin.H{
			"enabled":    h.cfg.S3Enabled,
			"endpoint":   h.cfg.S3Endpoint,
			"bucket":     h.cfg.S3Bucket,
			"prefix":     h.cfg.S3Prefix,
			"configured": h.cfg.S3AccessKey != "",
		},
	})
}

// ── AWS Signature V4 for MinIO ──────────────────────────────────────────────

func s3PutObject(cfg *config.Config, key string, data []byte) error {
	endpoint := strings.TrimRight(cfg.S3Endpoint, "/")
	bucket   := cfg.S3Bucket
	region   := cfg.S3Region
	if region == "" {
		region = "us-east-1"
	}

	// Build URL
	rawURL := fmt.Sprintf("%s/%s/%s", endpoint, bucket, key)
	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	dateShort := now.Format("20060102")
	dateLong  := now.Format("20060102T150405Z")

	// Payload hash
	payloadHash := sha256Hex(data)

	// Headers to sign
	host := u.Host
	headers := map[string]string{
		"host":                 host,
		"x-amz-content-sha256": payloadHash,
		"x-amz-date":           dateLong,
	}

	// Canonical headers + signed headers
	signedHeaders := "host;x-amz-content-sha256;x-amz-date"
	canonicalHeaders := fmt.Sprintf("host:%s\nx-amz-content-sha256:%s\nx-amz-date:%s\n",
		host, payloadHash, dateLong)

	// Canonical request
	canonicalRequest := strings.Join([]string{
		"PUT",
		u.EscapedPath(),
		"",
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	}, "\n")

	// String to sign
	credentialScope := fmt.Sprintf("%s/%s/s3/aws4_request", dateShort, region)
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		dateLong,
		credentialScope,
		sha256Hex([]byte(canonicalRequest)),
	}, "\n")

	// Signing key
	signingKey := hmacSHA256(
		hmacSHA256(
			hmacSHA256(
				hmacSHA256(
					[]byte("AWS4"+cfg.S3SecretKey),
					[]byte(dateShort),
				),
				[]byte(region),
			),
			[]byte("s3"),
		),
		[]byte("aws4_request"),
	)
	signature := hex.EncodeToString(hmacSHA256(signingKey, []byte(stringToSign)))

	// Authorization header
	authorization := fmt.Sprintf(
		"AWS4-HMAC-SHA256 Credential=%s/%s,SignedHeaders=%s,Signature=%s",
		cfg.S3AccessKey, credentialScope, signedHeaders, signature,
	)

	// Build request
	req, err := http.NewRequest("PUT", rawURL, bytes.NewReader(data))
	if err != nil {
		return err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("Authorization", authorization)
	req.Header.Set("Content-Type", "text/plain")
	req.ContentLength = int64(len(data))

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("S3 request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		return fmt.Errorf("S3 upload failed HTTP %d: %s", resp.StatusCode, buf.String())
	}
	return nil
}

func sha256Hex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}
