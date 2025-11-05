package repository

import (
	"database/sql"
	"fmt"

	"aviation-weather/internal/domain"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetByFAA retrieves an airport by its FAA identifier (e.g., "ATL").
func (r *Repository) GetByFAA(faa string) (*domain.Airport, error) {
	query := `
		SELECT site_number, facility_name, faa, icao, state_code, state_full, county, city,
		       ownership_type, use_type, manager, manager_phone, latitude, longitude,
		       airport_status, weather
		FROM airport WHERE faa = $1`

	row := r.db.QueryRow(query, faa)

	var airport domain.Airport
	err := row.Scan(
		&airport.SiteNumber,
		&airport.FacilityName,
		&airport.Faa,
		&airport.Icao,
		&airport.StateCode,
		&airport.StateFull,
		&airport.County,
		&airport.City,
		&airport.OwnershipType,
		&airport.UseType,
		&airport.Manager,
		&airport.ManagerPhone,
		&airport.Latitude,
		&airport.Longitude,
		&airport.AirportStatus,
		&airport.Weather,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found: return nil, no error
		}
		return nil, fmt.Errorf("failed to get airport by FAA %s: %w", faa, err)
	}

	return &airport, nil
}

// Create new airport
// test create manually: docker-compose exec postgres psql -U postgres -d aviation_weather -c "INSERT INTO airport (site_number, facility_name, faa, icao, state_code, state_full, county, city, ownership_type, use_type, manager, manager_phone, latitude, longitude, airport_status, weather) VALUES ('ATL_SITE', 'Hartsfield-Jackson Atlanta Intl', 'ATL', 'KATL', 'GA', 'Georgia', 'Fulton', 'Atlanta', 'Public', 'Public Use', 'Mgr Name', '404-123-4567', 33.6404, -84.4267, 'open', 'Cloudy') ON CONFLICT (site_number) DO NOTHING;"
func (r *Repository) Create(airport *domain.Airport) error {
	query := `
		INSERT INTO airport (
			site_number, facility_name, faa, icao, state_code, state_full, county, city,
			ownership_type, use_type, manager, manager_phone, latitude, longitude,
			airport_status, weather
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		ON CONFLICT (site_number) DO NOTHING`

	_, err := r.db.Exec(query,
		airport.SiteNumber,
		airport.FacilityName,
		airport.Faa,
		airport.Icao,
		airport.StateCode,
		airport.StateFull,
		airport.County,
		airport.City,
		airport.OwnershipType,
		airport.UseType,
		airport.Manager,
		airport.ManagerPhone,
		airport.Latitude,
		airport.Longitude,
		airport.AirportStatus,
		airport.Weather,
	)
	if err != nil {
		return fmt.Errorf("failed to create airport: %w", err)
	}

	return nil
}
