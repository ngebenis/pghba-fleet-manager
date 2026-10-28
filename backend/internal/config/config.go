package config

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ServerAddr string
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	DBSSLMode  string

	JWTSecret            string
	JWTAccessExpiry      time.Duration
	JWTRefreshExpiry     time.Duration

	EncryptionKey string // 32-byte hex for AES-256

	// CORSAllowedOrigins is empty by default — the frontend is normally
	// served from the same origin as the API (prod) or proxied through the
	// Vite dev server (dev), so cross-origin requests are never legitimate
	// out of the box. Set CORS_ALLOWED_ORIGINS (comma-separated) only when
	// the frontend is deployed on a different origin from this API.
	CORSAllowedOrigins []string

	SchedulerIntervalMin int
	TempAccessAction     string // "disable" or "delete"
	AutoReloadConf       bool

	// Notification
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPass     string
	SMTPFrom     string
	SMTPTo       string
	SMTPEnabled  bool
	TelegramToken string
	TelegramChatID string
	NotifyEnabled bool

	// Storage export
	BackupStoragePath string // local folder path
	S3Enabled         bool
	S3Endpoint        string
	S3Bucket          string
	S3Region          string
	S3AccessKey       string
	S3SecretKey       string
	S3Prefix          string

	// Auto-backup
	AutoBackupEnabled  bool
	AutoBackupCron     string // cron expression, default "0 2 * * *" (2am daily)
	AutoBackupTarget   string // "local", "s3", "both"
}

func Load() *Config {
	return &Config{
		ServerAddr:           getEnv("SERVER_ADDR", ":8080"),
		DBHost:               getEnv("DB_HOST", "localhost"),
		DBPort:               getEnv("DB_PORT", "5432"),
		DBName:               getEnv("DB_NAME", "pghba_manager"),
		DBUser:               getEnv("DB_USER", "admin"),
		DBPassword:           getEnv("DB_PASSWORD", knownPublicDefaultDBPassword),
		DBSSLMode:            getEnv("DB_SSL_MODE", "disable"),
		JWTSecret:            getEnv("JWT_SECRET", knownPublicDefaultJWTSecret),
		JWTAccessExpiry:      parseDuration(getEnv("JWT_ACCESS_EXPIRY", "8h")),
		JWTRefreshExpiry:     parseDuration(getEnv("JWT_REFRESH_EXPIRY", "168h")),
		EncryptionKey:        getEnv("ENCRYPTION_KEY", knownPublicDefaultEncryptionKey),
		CORSAllowedOrigins:   splitCSV(getEnv("CORS_ALLOWED_ORIGINS", "")),
		SchedulerIntervalMin: parseInt(getEnv("SCHEDULER_INTERVAL_MIN", "1")),
		TempAccessAction:     getEnv("TEMP_ACCESS_ACTION", "disable"),
		AutoReloadConf:       getEnv("AUTO_RELOAD_CONF", "true") == "true",

		SMTPHost:      getEnv("SMTP_HOST", ""),
		SMTPPort:      parseInt(getEnv("SMTP_PORT", "587")),
		SMTPUser:      getEnv("SMTP_USER", ""),
		SMTPPass:      getEnv("SMTP_PASS", ""),
		SMTPFrom:      getEnv("SMTP_FROM", ""),
		SMTPTo:        getEnv("SMTP_TO", ""),
		SMTPEnabled:   getEnv("SMTP_ENABLED", "false") == "true",
		TelegramToken:  getEnv("TELEGRAM_TOKEN", ""),
		TelegramChatID: getEnv("TELEGRAM_CHAT_ID", ""),
		NotifyEnabled: getEnv("NOTIFY_ENABLED", "false") == "true",

		BackupStoragePath: getEnv("BACKUP_STORAGE_PATH", "/var/lib/pghba-manager/backups"),
		S3Enabled:         getEnv("S3_ENABLED", "false") == "true",
		S3Endpoint:        getEnv("S3_ENDPOINT", ""),
		S3Bucket:          getEnv("S3_BUCKET", "pghba-backups"),
		S3Region:          getEnv("S3_REGION", "us-east-1"),
		S3AccessKey:       getEnv("S3_ACCESS_KEY", ""),
		S3SecretKey:       getEnv("S3_SECRET_KEY", ""),
		S3Prefix:          getEnv("S3_PREFIX", "pghba-manager"),

		AutoBackupEnabled: getEnv("AUTO_BACKUP_ENABLED", "false") == "true",
		AutoBackupCron:    getEnv("AUTO_BACKUP_CRON", "0 2 * * *"),
		AutoBackupTarget:  getEnv("AUTO_BACKUP_TARGET", "local"),
	}
}

// knownPublicDefault* are the literal fallback values hardcoded above (and
// mirrored in .env.example / database/migrations). Because this project's
// source is public, these strings are not secrets to anyone — a
// deployment that never overrides them is exactly as forgeable/decryptable
// as if it published its real key. Validate refuses to start on the ones
// with direct, remote-exploitable impact (forging any JWT, decrypting
// every stored SSH key) and only warns on the DB password, which normally
// isn't reachable from outside the deployment's own network.
const (
	knownPublicDefaultJWTSecret     = "change-me-in-production-min-32chars"
	knownPublicDefaultDBPassword    = "secret123"
	knownPublicDefaultEncryptionKey = "0000000000000000000000000000000000000000000000000000000000000000"
)

// Validate refuses insecure configuration before the server ever accepts a
// connection, instead of failing confusingly later (e.g. a nil-pointer
// panic the first time encryption is used) or — worse — silently running
// with a secret value published in this repo's own source/history.
func (c *Config) Validate() error {
	if key, err := hex.DecodeString(c.EncryptionKey); err != nil || len(key) != 32 {
		return fmt.Errorf("ENCRYPTION_KEY must be exactly 64 hex characters (32 bytes) for AES-256 — generate one with `openssl rand -hex 32`")
	}
	if c.EncryptionKey == knownPublicDefaultEncryptionKey {
		return fmt.Errorf("ENCRYPTION_KEY is still the published placeholder from .env.example — it encrypts every stored SSH private key, so it must be a secret only this deployment knows; generate one with `openssl rand -hex 32`")
	}
	if c.JWTSecret == "" || c.JWTSecret == knownPublicDefaultJWTSecret {
		return fmt.Errorf("JWT_SECRET is empty or still the published default from this repo's source — anyone who reads the source can forge valid login tokens for any user; set it to a random value (e.g. `openssl rand -base64 48`)")
	}
	if c.DBPassword == knownPublicDefaultDBPassword {
		log.Printf("WARNING: DB_PASSWORD is still the published default from this repo's source (%q) — change it before exposing this database beyond a trusted private network", knownPublicDefaultDBPassword)
	}
	return nil
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	var out []string
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 8 * time.Hour
	}
	return d
}

func parseInt(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 1
	}
	return n
}
