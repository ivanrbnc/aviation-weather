// docker-compose exec app go run test_repo.go

package main

import (
	"database/sql"
	"fmt"
	"log"

	"aviation-weather/config"
	"aviation-weather/internal/repository"

	_ "github.com/lib/pq"
)

func main() {
	cfg := config.Load()
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	repo := repository.NewRepository(db)
	airport, err := repo.GetByFAA("ATL")
	if err != nil {
		log.Fatal(err)
	}
	if airport == nil {
		fmt.Println("No airport found for ATL")
		return
	}
	fmt.Printf("Found: %+v\n", airport)
}
