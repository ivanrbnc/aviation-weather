package main

import (
	"aviation-weather/config"
	"aviation-weather/internal/repository"
	"aviation-weather/internal/service"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"github.com/robfig/cron/v3"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Connect to PostgreSQL
	db, err := sql.Open(
		"postgres",
		fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
			cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
		),
	)
	if err != nil {
		log.Fatalf("failed to open DB: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping DB: %v", err)
	}
	log.Println("Connected to PostgreSQL")

	// Initialize app layers
	repo := repository.NewRepository(db)
	svc := service.NewService(repo, cfg)

	// Initialize cron scheduler
	cronScheduler := cron.New()

	// Schedule SyncAllAirports to run every 12 hours
	_, err = cronScheduler.AddFunc("0 0,12 * * *", func() {
		log.Println("Starting SyncAllAirports...")
		updated, err := svc.SyncAllAirports()
		if err != nil {
			log.Printf("Error in SyncAllAirports: %v", err)
			return
		}
		log.Printf("SyncAllAirports completed, updated %d airports", updated)
	})
	if err != nil {
		log.Fatalf("Failed to schedule SyncAllAirports: %v", err)
	}

	// Start the cron scheduler
	cronScheduler.Start()
	log.Println("Scheduler started, running SyncAllAirports every 12 hours")

	// Keep the application running
	select {}
}
