package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"aviation-weather/config"
	"aviation-weather/internal/handler"
	"aviation-weather/internal/repository"
	"aviation-weather/internal/service"

	_ "github.com/lib/pq"
)

// Build DSN helper (shared).
func buildDSN(cfg *config.Config) string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
}

func main() {
	// Load config
	cfg := config.Load()

	// Connect to DB
	db, err := sql.Open("postgres", buildDSN(cfg))
	if err != nil {
		log.Fatalf("failed to open DB: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping DB: %v", err)
	}
	log.Println("Connected to PostgreSQL")

	// Initialize layers
	repo := repository.NewRepository(db)
	svc := service.NewService(repo, cfg)
	h := handler.NewHandler(svc)

	// Start server with Chi router
	port := ":" + cfg.AppPort
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(port, h.Router()))
}
