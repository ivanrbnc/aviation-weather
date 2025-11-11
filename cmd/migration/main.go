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
	up := flag.Bool("up", false, "Run migration up (create)")                                  // docker-compose exec app go run cmd/migration/main.go --up
	down := flag.Bool("down", false, "Run migration down (drop)")                              // docker-compose exec app go run cmd/migration/main.go --down
	fill := flag.Bool("fill", false, "Fill table with top US airports via SQL (implies --up)") // docker-compose exec app go run cmd/migration/main.go --fill
	flag.Parse()

	// VERIFY TABLE: docker-compose exec postgres psql -U postgres -d aviation_weather -c "\d airport"

	// Default flag behavior
	switch {
	case *fill && *down:
		log.Fatal("error: cannot use --fill with --down")
	case *up && *down:
		log.Fatal("error: cannot specify both --up and --down")
	case !*up && !*down && !*fill:
		*up = true
		log.Println("No flags provided; defaulting to --up")
	}

	if *fill {
		*up = true
		log.Println("--fill requested: Will run --up then seed data")
	}

	// Load config and connect
	cfg := config.Load()
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("db connection error: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("db ping error: %v", err)
	}
	log.Println("Connected to PostgreSQL")

	// Run migration
	runMigration := func(filename, action string) {
		sqlBytes, err := os.ReadFile(filename)
		if err != nil {
			log.Fatalf("error reading %s: %v", filename, err)
		}
		if _, err := db.Exec(string(sqlBytes)); err != nil {
			log.Fatalf("error executing %s: %v", filename, err)
		}
		log.Printf("%s completed: %s", action, filename)
	}

	switch {
	case *down:
		runMigration("migrations/drop_airport.sql", "Migration down")
		return // Early exit after down—no fill possible
	case *up:
		runMigration("migrations/create_airport.sql", "Migration up")
		if *fill {
			runMigration("migrations/fill_airport.sql", "Fill (seed data)")
		}
	}
}
