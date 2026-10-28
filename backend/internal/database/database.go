package database

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"sort"

	"github.com/pghba-manager/backend/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// 001_init.sql bootstraps the baseline schema and is mounted straight into
// Postgres' /docker-entrypoint-initdb.d, which only runs on a brand-new
// (empty) data directory — so by the time this app connects, it has always
// already been applied. Migrate() records it as applied without
// re-executing it (its CREATE TABLE/CREATE TYPE statements aren't
// idempotent, unlike every migration after it).
const baselineMigration = "001_init.sql"

func Connect(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=%s TimeZone=UTC",
		cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBUser, cfg.DBPassword, cfg.DBSSLMode,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	log.Println("database connected")
	return db, nil
}

// Migrate applies every embedded SQL migration that hasn't run against this
// database yet, tracked in a schema_migrations table. Every migration after
// 001_init.sql is written to be idempotent (IF NOT EXISTS / ON CONFLICT
// guards), so it's safe to apply on every startup — this is what makes a
// deploy self-healing instead of requiring someone to remember to run a new
// .sql file by hand against production.
func Migrate(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get underlying sql.DB: %w", err)
	}

	if _, err := sqlDB.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version     TEXT PRIMARY KEY,
			applied_at  TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read embedded migrations: %w", err)
	}
	versions := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			versions = append(versions, e.Name())
		}
	}
	sort.Strings(versions)

	for _, version := range versions {
		applied, err := isApplied(sqlDB, version)
		if err != nil {
			return fmt.Errorf("check migration %s: %w", version, err)
		}
		if applied {
			continue
		}

		if version == baselineMigration {
			// Already applied via Postgres' docker-entrypoint-initdb.d — just record it.
			if err := recordApplied(sqlDB, version); err != nil {
				return fmt.Errorf("record baseline migration %s: %w", version, err)
			}
			log.Printf("migrations: %s (baseline, applied via initdb)", version)
			continue
		}

		content, err := migrationsFS.ReadFile("migrations/" + version)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", version, err)
		}
		if _, err := sqlDB.Exec(string(content)); err != nil {
			return fmt.Errorf("apply migration %s: %w", version, err)
		}
		if err := recordApplied(sqlDB, version); err != nil {
			return fmt.Errorf("record migration %s: %w", version, err)
		}
		log.Printf("migrations: applied %s", version)
	}

	return nil
}

func isApplied(sqlDB *sql.DB, version string) (bool, error) {
	var exists bool
	err := sqlDB.QueryRow(`SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)`, version).Scan(&exists)
	return exists, err
}

func recordApplied(sqlDB *sql.DB, version string) error {
	_, err := sqlDB.Exec(`INSERT INTO schema_migrations (version) VALUES ($1) ON CONFLICT DO NOTHING`, version)
	return err
}
