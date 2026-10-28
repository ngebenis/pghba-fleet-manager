package database

import (
	"fmt"
	"log"

	"github.com/pghba-manager/backend/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

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

func Migrate(db *gorm.DB) error {
	// Schema dihandle oleh SQL migration di database/migrations/001_init.sql
	// yang di-mount ke /docker-entrypoint-initdb.d/ saat postgres pertama kali start.
	// AutoMigrate dinonaktifkan untuk menghindari konflik constraint naming dengan GORM.
	log.Println("migrations: using SQL migration files, skipping AutoMigrate")
	return nil
}
