package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/pghba-manager/backend/internal/config"
	"github.com/pghba-manager/backend/internal/database"
	"github.com/pghba-manager/backend/internal/scheduler"
	"github.com/pghba-manager/backend/internal/server"
)

func main() {
	// Load .env (optional — falls back to real env vars)
	_ = godotenv.Load()

	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("insecure configuration: %v", err)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := database.Migrate(db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	// Start background scheduler (temp-access expiry checks)
	sched := scheduler.New(db, cfg)
	sched.Start()
	defer sched.Stop()

	r := server.NewRouter(db, cfg)

	log.Printf("PGHBA Manager starting on %s", cfg.ServerAddr)
	if err := r.Run(cfg.ServerAddr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
