package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"aviation-weather/config"

	_ "github.com/lib/pq"
)

func main() {
	// Parse flags
	up := flag.Bool("up", false, "Run migration up (create)")     // docker-compose exec app go run cmd/migration/main.go --up
	down := flag.Bool("down", false, "Run migration down (drop)") // docker-compose exec app go run cmd/migration/main.go --down
	flag.Parse()

	// VERIFY: docker-compose exec postgres psql -U postgres -d aviation_weather -c "\d airport"

	// Default: --up
	if !*up && !*down {
		*up = true
		log.Println("No flags provided; defaulting to --up")
	}

	// Can't do both!
	if *up && *down {
		log.Fatal("Error: Cannot specify both --up and --down")
	}

	// Load config
	cfg := config.Load()

	// Build Data Source Name (DSN)
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	// Connect to DB
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Error connecting to DB: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Error pinging DB: %v", err)
	}
	log.Println("Connected to PostgreSQL")

	// Determine and read SQL file
	var filename string
	if *up {
		filename = "migrations/create_airport.up.sql"
	} else {
		filename = "migrations/drop_airport.down.sql"
	}

	// Read SQL to memory
	sqlBytes, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Error reading migration file %s: %v", filename, err)
	}

	// Execute migration
	_, err = db.Exec(string(sqlBytes))
	if err != nil {
		log.Fatalf("Error executing migration %s: %v", filename, err)
	}

	if *up {
		log.Println("Migration up completed: Airport table created")
	} else {
		log.Println("Migration down completed: Airport table dropped")
	}
}
