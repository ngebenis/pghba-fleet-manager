package server

import (
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/pghba-manager/backend/internal/config"
	"github.com/pghba-manager/backend/internal/handlers"
	"github.com/pghba-manager/backend/internal/middleware"
	"github.com/pghba-manager/backend/internal/models"
	"github.com/pghba-manager/backend/internal/services"
	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB, cfg *config.Config) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), middleware.SecurityHeaders())

	// No configured origins (the default) means the frontend is same-origin
	// with this API — same-origin requests never hit CORS at all, so there
	// is nothing to allow. AllowOrigins:["*"] + AllowCredentials:true (the
	// old config) is an anti-pattern: it tells browsers any site may make
	// credentialed requests here, which only matters for defense-in-depth
	// today (the JWT lives in localStorage, not a cookie) but still removes
	// a layer for no reason — explicit origins only, never a wildcard.
	if len(cfg.CORSAllowedOrigins) > 0 {
		r.Use(cors.New(cors.Config{
			AllowOrigins:     cfg.CORSAllowedOrigins,
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
		}))
	}

	// Serve Vue frontend
	r.Static("/assets", "./static/assets")
	r.StaticFile("/favicon.ico", "./static/favicon.ico")
	r.StaticFile("/favicon.svg", "./static/favicon.svg")
	r.NoRoute(func(c *gin.Context) {
		if len(c.Request.URL.Path) > 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{"error": "endpoint not found"})
			return
		}
		c.File("./static/index.html")
	})

	// Health check
	r.GET("/api/auth/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Services
	audit     := services.NewAuditService(db)
	hba       := services.NewHBAService(db, audit)
	crypto, err := services.NewCryptoService(cfg.EncryptionKey)
	if err != nil {
		// Should be unreachable — cfg.Validate() in main.go already rejects
		// an invalid ENCRYPTION_KEY before NewRouter is ever called. Kept as
		// a hard failure rather than silently running with a nil crypto
		// service, which would panic on first use instead of at startup.
		log.Fatalf("invalid ENCRYPTION_KEY: %v", err)
	}

	// Handlers
	authH  := handlers.NewAuthHandler(db, cfg, audit)
	hostH  := handlers.NewHostHandler(db, audit, crypto)
	instH  := handlers.NewInstanceHandler(db, audit, crypto)
	writeH := handlers.NewWriteHBAHandler(db, audit, crypto)
	ruleH  := handlers.NewRuleHandler(db, audit, hba, crypto)
	tempH  := handlers.NewTempAccessHandler(db, audit, crypto)
	backH  := handlers.NewBackupHandler(db, audit, hba, crypto)
	backSchedH := handlers.NewBackupScheduleHandler(db, audit)
	userH  := handlers.NewUserHandler(db, audit)
	auditH := handlers.NewAuditHandler(db)
	dashH  := handlers.NewDashboardHandler(db)
	notifH  := handlers.NewNotificationHandler(db)
	settH     := handlers.NewSettingsHandler(db, cfg)
	conflictH := handlers.NewConflictHandler(db)
	exportH   := handlers.NewExportHandler(db, cfg, audit)
	healthH  := handlers.NewHealthCheckHandler(db, cfg, crypto, services.NewNotifier(cfg), audit)

	api := r.Group("/api")

	// ── Public ───────────────────────────────────────────────────────────────
	api.POST("/auth/login",   middleware.LoginRateLimit(), authH.Login)
	api.POST("/auth/refresh", authH.Refresh)

	// ── Protected ─────────────────────────────────────────────────────────────
	p := api.Group("")
	p.Use(middleware.AuthRequired(cfg, db), middleware.RequirePasswordChanged())
	{
		p.POST("/auth/logout",          authH.Logout)
		p.GET("/auth/me",               authH.Me)
		p.POST("/auth/change-password", authH.ChangePassword)

		p.GET("/dashboard/stats", dashH.Stats)
		p.GET("/notifications",   notifH.List)

		// ── Hosts ─────────────────────────────────────────────────────────
		hg := p.Group("/hosts")
		hg.Use(middleware.RequireRole(models.RoleSuperAdmin, models.RoleDBA))
		{
			hg.GET("",        hostH.List)
			hg.GET("/:id",    hostH.Get)
			hg.POST("/validate", hostH.Validate)
			hg.POST("",       hostH.Create)
			hg.PUT("/:id",    hostH.Update)
			hg.DELETE("/:id",          hostH.Delete)
			hg.POST("/:id/deploy-key", hostH.DeploySSHKey)
			hg.POST("/:id/reset-host-key", hostH.ResetHostKeyFingerprint)
		}

		// ── Instances ─────────────────────────────────────────────────────
		ig := p.Group("/instances")
		ig.Use(middleware.RequireRole(models.RoleSuperAdmin, models.RoleDBA))
		{
			ig.POST("",                  instH.Create)
			ig.POST("/discover",         instH.Discover)
			ig.POST("/scan-all",         instH.ScanAll)
			ig.POST("/bulk-import",      instH.BulkImportHBA)
			ig.POST("/:id/scan",         instH.Scan)
			ig.POST("/:id/import-hba",   instH.ImportHBA)
			ig.POST("/bulk-apply",        writeH.BulkWriteAndReload)
		}
		// List/Get are also usable by AccessManager and TempAccessManager:
		// both need the instance list to populate the Rules/Temp Access
		// selectors, but neither gets the rest of instance management.
		// preview-diff/write-hba stay narrower — AccessManager pushes rule
		// changes live via the explicit Write & Reload button, but
		// TempAccessManager's grant/revoke flow writes back internally on
		// the server side and never calls these directly. Registered
		// outside the ig group so they don't inherit its SuperAdmin/DBA-only
		// middleware.
		instReadRG := middleware.RequireRole(models.RoleSuperAdmin, models.RoleDBA, models.RoleAccessManager, models.RoleTempAccessManager)
		instWriteRG := middleware.RequireRole(models.RoleSuperAdmin, models.RoleDBA, models.RoleAccessManager)
		p.GET("/instances", instReadRG, instH.List)
		p.GET("/instances/:id", instReadRG, instH.Get)
		p.GET("/instances/:id/preview-diff", instWriteRG, writeH.PreviewDiff)
		p.POST("/instances/:id/write-hba", instWriteRG, writeH.WriteAndReload)

		// ── Rules ─────────────────────────────────────────────────────────
		rg := p.Group("/rules")
		{
			// TempAccessManager is excluded here (unlike every other route
			// in this file with no RequireRole) since it should not see
			// Access Rules at all — every other existing role keeps the
			// unrestricted read access it already had.
			rg.GET("", middleware.RequireRole(models.RoleSuperAdmin, models.RoleDBA, models.RoleOperator, models.RoleViewer, models.RoleAccessManager), ruleH.List)
			rg.POST("",       middleware.RequireRole(models.RoleSuperAdmin, models.RoleDBA, models.RoleAccessManager), ruleH.Create)
			rg.PUT("/:id",    middleware.RequireRole(models.RoleSuperAdmin, models.RoleDBA, models.RoleAccessManager), ruleH.Update)
			rg.DELETE("/:id", middleware.RequireRole(models.RoleSuperAdmin, models.RoleDBA, models.RoleAccessManager), ruleH.Delete)
			rg.POST("/:id/enable",        middleware.RequireRole(models.RoleSuperAdmin, models.RoleDBA, models.RoleOperator, models.RoleAccessManager), ruleH.Enable)
			rg.POST("/:id/disable",       middleware.RequireRole(models.RoleSuperAdmin, models.RoleDBA, models.RoleOperator, models.RoleAccessManager), ruleH.Disable)
			rg.POST("/:id/set-temporary", middleware.RequireRole(models.RoleSuperAdmin, models.RoleDBA, models.RoleOperator, models.RoleAccessManager), ruleH.SetTemporary)
			rg.POST("/reorder",           middleware.RequireRole(models.RoleSuperAdmin, models.RoleDBA, models.RoleAccessManager), ruleH.Reorder)
			rg.POST("/clone",             middleware.RequireRole(models.RoleSuperAdmin, models.RoleDBA, models.RoleAccessManager), ruleH.CloneRules)
		}

		// ── Temporary Access ──────────────────────────────────────────────
		tg := p.Group("/temporary-access")
		tg.Use(middleware.RequireRole(models.RoleSuperAdmin, models.RoleDBA, models.RoleOperator, models.RoleAccessManager, models.RoleTempAccessManager))
		{
			tg.GET("",             tempH.List)
			tg.POST("",            tempH.Grant)
			tg.POST("/:id/revoke", tempH.Revoke)
		}

		// ── Backups ───────────────────────────────────────────────────────
		bg := p.Group("/backups")
		bg.Use(middleware.RequireRole(models.RoleSuperAdmin, models.RoleDBA))
		{
			bg.GET("",              backH.List)
			bg.POST("",             backH.Create)
			bg.GET("/:id",          backH.Preview)
			bg.POST("/:id/restore", backH.Restore)
			bg.DELETE("/:id",       backH.Delete)
			bg.GET("/:id/download", backH.Download)
		}

		// ── Backup Schedules ──────────────────────────────────────────────
		bsg := p.Group("/backup-schedules")
		bsg.Use(middleware.RequireRole(models.RoleSuperAdmin, models.RoleDBA))
		{
			bsg.GET("",       backSchedH.List)
			bsg.POST("",      backSchedH.Create)
			bsg.PUT("/:id",   backSchedH.Update)
			bsg.DELETE("/:id", backSchedH.Delete)
		}

		// ── Users ─────────────────────────────────────────────────────────
		ug := p.Group("/users")
		ug.Use(middleware.RequireRole(models.RoleSuperAdmin))
		{
			ug.GET("",                     userH.List)
			ug.POST("",                    userH.Create)
			ug.PUT("/:id",                 userH.Update)
			ug.POST("/:id/reset-password", userH.ResetPassword)
			ug.DELETE("/:id",              userH.Delete)
		}

		// ── Audit ─────────────────────────────────────────────────────────
		p.GET("/settings/system",       settH.GetSystemInfo)

		// ── Export ────────────────────────────────────────────────────────────
		p.GET("/export/config",          middleware.RequireRole(models.RoleSuperAdmin, models.RoleDBA), exportH.ExportConfig)
		p.GET("/export/backups",         middleware.RequireRole(models.RoleSuperAdmin, models.RoleDBA), exportH.ExportBackups)
		p.POST("/export/save-storage",   middleware.RequireRole(models.RoleSuperAdmin), exportH.SaveBackupToStorage)
		p.GET("/export/storage-status",  middleware.RequireRole(models.RoleSuperAdmin, models.RoleDBA), exportH.GetStorageStatus)

		// ── Health Check ──────────────────────────────────────────────────────
		p.GET("/health/check",           middleware.RequireRole(models.RoleSuperAdmin, models.RoleDBA), healthH.Check)
		p.GET("/health/check/:id",       middleware.RequireRole(models.RoleSuperAdmin, models.RoleDBA), healthH.CheckHost)
		p.GET("/health/history",         middleware.RequireRole(models.RoleSuperAdmin, models.RoleDBA), healthH.GetHistory)
		p.GET("/rules/conflicts",        conflictH.Detect)
		p.POST("/settings/purge-audit",  middleware.RequireRole(models.RoleSuperAdmin), settH.PurgeAuditLog)
		p.GET("/settings/eol-status",    settH.GetEOLStatus)
		p.GET("/audit",           middleware.RequireRole(models.RoleSuperAdmin, models.RoleDBA), auditH.List)
		p.GET("/audit/diff/:id",  middleware.RequireRole(models.RoleSuperAdmin, models.RoleDBA), auditH.GetDiff)
	}

	return r
}
