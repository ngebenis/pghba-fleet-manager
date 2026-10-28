package scheduler

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pghba-manager/backend/internal/config"
	"github.com/pghba-manager/backend/internal/models"
	"github.com/pghba-manager/backend/internal/services"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

type Scheduler struct {
	cron       *cron.Cron
	hbaService *services.HBAService
	notifier   *services.Notifier
	cfg        *config.Config
	db         *gorm.DB
	crypto     *services.CryptoService
}

func New(db *gorm.DB, cfg *config.Config) *Scheduler {
	audit     := services.NewAuditService(db)
	crypto, _ := services.NewCryptoService(cfg.EncryptionKey)
	hba       := services.NewHBAServiceWithCrypto(db, audit, crypto)
	notifier  := services.NewNotifier(cfg)
	return &Scheduler{
		cron:       cron.New(),
		hbaService: hba,
		notifier:   notifier,
		cfg:        cfg,
		db:         db,
		crypto:     crypto,
	}
}

func (s *Scheduler) Start() {
	spec := fmt.Sprintf("@every %dm", s.cfg.SchedulerIntervalMin)

	// Main job: expire temp rules + notify
	s.cron.AddFunc(spec, func() {
		s.checkAndExpireTempRules()
	})

	// Warning job: notify 15 min before expiry (runs every minute)
	s.cron.AddFunc("@every 1m", func() {
		s.sendExpiryWarnings()
	})

	// Health check job — every 5 minutes
	s.cron.AddFunc("@every 5m", func() {
		s.runHealthCheck()
	})

	// Auto-backup job (legacy single global policy via .env)
	if s.cfg.AutoBackupEnabled {
		s.cron.AddFunc(s.cfg.AutoBackupCron, func() {
			s.runAutoBackup()
		})
		log.Printf("[scheduler] auto-backup enabled — cron: %s, target: %s",
			s.cfg.AutoBackupCron, s.cfg.AutoBackupTarget)
	}

	// Backup schedules — user-managed schedules from Settings (backup_schedules
	// table). Checked every minute; each schedule's own cron_expression decides
	// whether it's due, so edits/adds/deletes via the API take effect on the
	// next tick with no need to re-register anything here.
	s.cron.AddFunc("@every 1m", func() {
		s.runDueBackupSchedules()
	})

	s.cron.Start()
	log.Printf("[scheduler] started — checking every %d minute(s)", s.cfg.SchedulerIntervalMin)
}

func (s *Scheduler) checkAndExpireTempRules() {
	count, err := s.hbaService.DisableExpiredTempRules(s.cfg.TempAccessAction)
	if err != nil {
		log.Printf("[scheduler] error checking expired temp rules: %v", err)
		return
	}
	if count > 0 {
		log.Printf("[scheduler] expired %d temporary rule(s)", count)
		// Notify expired
		expired := s.getRecentlyExpired()
		if len(expired) > 0 {
			s.notifier.SendTempAccessExpired(expired)
		}
	}
}

func (s *Scheduler) sendExpiryWarnings() {
	now := time.Now()
	warn15 := now.Add(15 * time.Minute)
	warn5  := now.Add(5 * time.Minute)

	type Row struct {
		ID            string
		IPAddress     string
		InstanceName  string
		Port          int
		HostName      string
		ExpiresAt     time.Time
		Username      string
		Reason        string
	}

	var rows []Row
	s.db.Raw(`
		SELECT ta.id,
			COALESCE(r.address::text,'') as ip_address,
			COALESCE(i.instance_name,'') as instance_name,
			COALESCE(i.port,0)           as port,
			COALESCE(h.name,'')          as host_name,
			ta.expires_at,
			COALESCE(u.username,'system') as username,
			COALESCE(ta.reason,'')        as reason
		FROM temp_access ta
		LEFT JOIN hba_rules r    ON r.id = ta.rule_id
		LEFT JOIN pg_instances i ON i.id = ta.instance_id
		LEFT JOIN db_hosts h     ON h.id = i.host_id
		LEFT JOIN users u        ON u.id = ta.granted_by_id
		WHERE ta.revoked = false
		  AND ta.expires_at > NOW()
		  AND ta.expires_at <= ?
	`, warn15).Scan(&rows)

	if len(rows) == 0 {
		return
	}

	var alerts []services.TempAccessAlert
	for _, r := range rows {
		diff := r.ExpiresAt.Sub(now)
		minutesLeft := int(diff.Minutes())

		// Only notify at specific thresholds to avoid spam
		if minutesLeft != 15 && minutesLeft != 10 && minutesLeft != 5 && minutesLeft != 1 {
			_ = warn5
			continue
		}

		alerts = append(alerts, services.TempAccessAlert{
			IPAddress:     r.IPAddress,
			InstanceLabel: fmt.Sprintf("%s:%d", r.InstanceName, r.Port),
			HostName:      r.HostName,
			GrantedBy:     r.Username,
			ExpiresAt:     r.ExpiresAt,
			MinutesLeft:   minutesLeft,
			Reason:        r.Reason,
		})
	}

	if len(alerts) > 0 {
		log.Printf("[scheduler] sending expiry warning for %d temp access", len(alerts))
		s.notifier.SendTempAccessWarning(alerts)
	}
}

func (s *Scheduler) getRecentlyExpired() []services.TempAccessAlert {
	type Row struct {
		IPAddress    string
		InstanceName string
		Port         int
		HostName     string
		Username     string
		ExpiresAt    time.Time
	}
	var rows []Row
	s.db.Raw(`
		SELECT COALESCE(r.address::text,'') as ip_address,
			COALESCE(i.instance_name,'') as instance_name,
			COALESCE(i.port,0) as port,
			COALESCE(h.name,'') as host_name,
			COALESCE(u.username,'system') as username,
			ta.expires_at
		FROM temp_access ta
		LEFT JOIN hba_rules r    ON r.id = ta.rule_id
		LEFT JOIN pg_instances i ON i.id = ta.instance_id
		LEFT JOIN db_hosts h     ON h.id = i.host_id
		LEFT JOIN users u        ON u.id = ta.granted_by_id
		WHERE ta.revoked = true
		  AND ta.revoked_at >= NOW() - INTERVAL '2 minutes'
	`).Scan(&rows)

	var alerts []services.TempAccessAlert
	for _, r := range rows {
		alerts = append(alerts, services.TempAccessAlert{
			IPAddress:     r.IPAddress,
			InstanceLabel: fmt.Sprintf("%s:%d", r.InstanceName, r.Port),
			HostName:      r.HostName,
			GrantedBy:     r.Username,
			ExpiresAt:     r.ExpiresAt,
		})
	}
	return alerts
}

func (s *Scheduler) runAutoBackup() {
	log.Printf("[scheduler] running auto-backup (target: %s)...", s.cfg.AutoBackupTarget)

	// Get all instances
	var instances []models.PGInstance
	s.db.Preload("Host").Find(&instances)

	runStart := time.Now()
	created := 0
	for _, inst := range instances {
		// Create backup in DB
		if err := s.hbaService.CreateBackup(inst.ID, "auto", "Scheduled auto-backup", nil); err != nil {
			log.Printf("[scheduler] backup failed for %s:%d: %v", inst.InstanceName, inst.Port, err)
			continue
		}
		created++
	}

	log.Printf("[scheduler] auto-backup: created %d/%d backups", created, len(instances))

	// Save to storage target
	if created > 0 {
		var backups []models.HBABackup
		s.db.Preload("Instance").
			Where("trigger_event = 'auto' AND created_at >= ?", runStart).
			Find(&backups)

		saved := s.saveBackupsToTarget(s.cfg.AutoBackupTarget, backups)
		log.Printf("[scheduler] auto-backup complete — %d created, %d saved to %s",
			created, saved, s.cfg.AutoBackupTarget)
	}
}

// saveBackupsToTarget uploads/writes already-created HBABackup rows to the
// configured storage target(s). Returns how many writes succeeded — shared
// between the legacy .env auto-backup and the DB-managed backup schedules.
func (s *Scheduler) saveBackupsToTarget(target string, backups []models.HBABackup) int {
	saved := 0
	if target == "s3" || target == "both" {
		for _, bk := range backups {
			instName := bk.Instance.InstanceName
			if instName == "" { instName = "unknown" }
			key := fmt.Sprintf("%s/auto-backups/%s/%s", s.cfg.S3Prefix, instName, bk.Filename)
			if err := s3Upload(s.cfg, key, []byte(bk.Content)); err != nil {
				log.Printf("[scheduler] S3 upload failed for %s: %v", bk.Filename, err)
			} else {
				saved++
			}
		}
	}
	if target == "local" || target == "both" {
		for _, bk := range backups {
			instName := bk.Instance.InstanceName
			if instName == "" { instName = "unknown" }
			dir := s.cfg.BackupStoragePath + "/" + instName
			os.MkdirAll(dir, 0755)
			fpath := dir + "/" + bk.Filename
			if err := os.WriteFile(fpath, []byte(bk.Content), 0644); err != nil {
				log.Printf("[scheduler] local save failed: %v", err)
			} else {
				saved++
			}
		}
	}
	return saved
}

// runDueBackupSchedules checks every user-managed backup schedule and runs
// any whose cron expression says they're due, based on their last run time
// (or creation time, if they've never run).
func (s *Scheduler) runDueBackupSchedules() {
	var schedules []models.BackupSchedule
	if err := s.db.Where("enabled = true").Find(&schedules).Error; err != nil {
		log.Printf("[scheduler] error loading backup schedules: %v", err)
		return
	}

	now := time.Now()
	for _, sched := range schedules {
		parsed, err := cron.ParseStandard(sched.CronExpression)
		if err != nil {
			log.Printf("[scheduler] backup schedule %q has invalid cron %q: %v", sched.Name, sched.CronExpression, err)
			continue
		}
		reference := sched.CreatedAt
		if sched.LastRunAt != nil {
			reference = *sched.LastRunAt
		}
		if parsed.Next(reference).After(now) {
			continue // not due yet
		}
		s.runBackupSchedule(sched)
	}
}

// runBackupSchedule runs a single due schedule: backs up every instance in
// its scope, saves the backups to its storage target, and records the
// outcome back onto the schedule row.
func (s *Scheduler) runBackupSchedule(sched models.BackupSchedule) {
	log.Printf("[scheduler] running backup schedule %q (scope: %s)", sched.Name, sched.ScopeType)

	var instances []models.PGInstance
	q := s.db.Preload("Host")
	switch sched.ScopeType {
	case "host":
		if sched.ScopeID != nil {
			q = q.Where("host_id = ?", *sched.ScopeID)
		}
	case "instance":
		if sched.ScopeID != nil {
			q = q.Where("id = ?", *sched.ScopeID)
		}
	}
	q.Find(&instances)

	runStart := time.Now()
	instanceIDs := make([]uuid.UUID, 0, len(instances))
	created := 0
	for _, inst := range instances {
		instanceIDs = append(instanceIDs, inst.ID)
		if err := s.hbaService.CreateBackup(inst.ID, "auto", "Scheduled backup: "+sched.Name, nil); err != nil {
			log.Printf("[scheduler] backup schedule %q: failed for %s:%d: %v", sched.Name, inst.InstanceName, inst.Port, err)
			continue
		}
		created++
	}

	status := "success"
	message := fmt.Sprintf("%d/%d backup(s) created", created, len(instances))
	switch {
	case len(instances) == 0:
		status = "failed"
		message = "no matching instance found for schedule scope"
	case created < len(instances):
		status = "partial"
	}

	saved := 0
	if created > 0 {
		var backups []models.HBABackup
		s.db.Preload("Instance").
			Where("instance_id IN ? AND trigger_event = 'auto' AND created_at >= ?", instanceIDs, runStart).
			Find(&backups)
		saved = s.saveBackupsToTarget(sched.Target, backups)
		message = fmt.Sprintf("%s, %d saved to %s", message, saved, sched.Target)
	}

	s.db.Model(&models.BackupSchedule{}).Where("id = ?", sched.ID).Updates(map[string]interface{}{
		"last_run_at":      time.Now(),
		"last_run_status":  status,
		"last_run_message": message,
	})

	log.Printf("[scheduler] backup schedule %q complete — %s", sched.Name, message)
}

func (s *Scheduler) runHealthCheck() {
	var hosts []models.PGInstance
	// Check for instances that haven't been scanned in >10 minutes
	s.db.Where("status = 'Online' AND last_scan < NOW() - INTERVAL '10 minutes'").
		Find(&hosts)

	if len(hosts) > 0 {
		log.Printf("[healthcheck] %d instances may be stale, triggering scan", len(hosts))
	}
	// Full host check is done via API endpoint
}

func (s *Scheduler) Stop() {
	s.cron.Stop()
}

// Check if models package is imported
var _ = models.TempAccess{}


// s3Upload — AWS Signature V4 PUT for MinIO/S3
func s3Upload(cfg *config.Config, key string, data []byte) error {
	if cfg.S3Endpoint == "" || cfg.S3Bucket == "" { return fmt.Errorf("S3 not configured") }
	endpoint := strings.TrimRight(cfg.S3Endpoint, "/")
	region   := cfg.S3Region
	if region == "" { region = "us-east-1" }
	rawURL := fmt.Sprintf("%s/%s/%s", endpoint, cfg.S3Bucket, key)
	u, err := url.Parse(rawURL)
	if err != nil { return err }
	now := time.Now().UTC()
	dateShort := now.Format("20060102")
	dateLong  := now.Format("20060102T150405Z")
	h256 := sha256.Sum256(data)
	payloadHash := hex.EncodeToString(h256[:])
	host := u.Host
	canonicalHeaders := fmt.Sprintf("host:%s\\nx-amz-content-sha256:%s\\nx-amz-date:%s\\n", host, payloadHash, dateLong)
	signedHeaders    := "host;x-amz-content-sha256;x-amz-date"
	crBytes := sha256.Sum256([]byte(strings.Join([]string{"PUT", u.EscapedPath(), "", canonicalHeaders, signedHeaders, payloadHash}, "\n")))
	credScope   := fmt.Sprintf("%s/%s/s3/aws4_request", dateShort, region)
	stringToSign := strings.Join([]string{"AWS4-HMAC-SHA256", dateLong, credScope, hex.EncodeToString(crBytes[:])}, "\n")
	sign := func(k, d []byte) []byte { m := hmac.New(sha256.New, k); m.Write(d); return m.Sum(nil) }
	sigKey := sign(sign(sign(sign([]byte("AWS4"+cfg.S3SecretKey), []byte(dateShort)), []byte(region)), []byte("s3")), []byte("aws4_request"))
	sig := hex.EncodeToString(sign(sigKey, []byte(stringToSign)))
	auth := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s,SignedHeaders=%s,Signature=%s", cfg.S3AccessKey, credScope, signedHeaders, sig)
	req, err := http.NewRequest("PUT", rawURL, bytes.NewReader(data))
	if err != nil { return err }
	req.Header.Set("host", host)
	req.Header.Set("x-amz-content-sha256", payloadHash)
	req.Header.Set("x-amz-date", dateLong)
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "text/plain")
	req.ContentLength = int64(len(data))
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil { return fmt.Errorf("S3 request failed: %w", err) }
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		buf := new(bytes.Buffer); buf.ReadFrom(resp.Body)
		return fmt.Errorf("S3 HTTP %d: %s", resp.StatusCode, buf.String())
	}
	return nil
}

// pgEOL — PostgreSQL End of Life tracking
var pgEOLMap = map[string]string{
	"9.5": "2021-02-11", "9.6": "2021-11-11",
	"10":  "2022-11-10", "11":  "2023-11-09",
	"12":  "2024-11-14", "13":  "2025-11-13",
	"14":  "2026-11-12", "15":  "2027-11-11",
	"16":  "2028-11-09", "17":  "2029-11-08",
}

func IsVersionEOL(pgVersion string) bool {
	eol, ok := pgEOLMap[pgVersion]
	if !ok { return false }
	t, _ := time.Parse("2006-01-02", eol)
	return time.Now().After(t)
}

func VersionEOLDate(pgVersion string) string {
	if d, ok := pgEOLMap[pgVersion]; ok { return d }
	return ""
}
