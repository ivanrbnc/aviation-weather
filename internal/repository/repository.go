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

	// All fields except faa are nullable: Use Null types for scan
	var siteNumber, facilityName, icao, stateCode, stateFull, county, city, ownershipType, useType, manager, managerPhone, airportStatus, weather sql.NullString
	var latitude, longitude sql.NullFloat64

	err := row.Scan(
		&siteNumber,
		&facilityName,
		&airport.Faa,
		&icao,
		&stateCode,
		&stateFull,
		&county,
		&city,
		&ownershipType,
		&useType,
		&manager,
		&managerPhone,
		&latitude,
		&longitude,
		&airportStatus,
		&weather,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get airport by FAA %s: %w", faa, err)
	}

	return &airport, nil
}

// Create new airport (ignores if faa exists).
func (r *Repository) Create(airport *domain.Airport) error {
	query := `
		INSERT INTO airport (
			site_number, facility_name, faa, icao, state_code, state_full, county, city,
			ownership_type, use_type, manager, manager_phone, latitude, longitude,
			airport_status, weather
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		ON CONFLICT (faa) DO NOTHING`

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

// GetAllAirports fetches all airports from the DB.
func (r *Repository) GetAllAirports() ([]domain.Airport, error) {
	query := `
		SELECT site_number, facility_name, faa, icao, state_code, state_full, county, city,
		       ownership_type, use_type, manager, manager_phone, latitude, longitude,
		       airport_status, weather
		FROM airport ORDER BY faa`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all airports: %w", err)
	}
	defer rows.Close()

	var airports []domain.Airport
	for rows.Next() {
		var airport domain.Airport

		// All fields except faa are nullable: Use Null types for scan
		var siteNumber, facilityName, icao, stateCode, stateFull, county, city, ownershipType, useType, manager, managerPhone, airportStatus, weather sql.NullString
		var latitude, longitude sql.NullFloat64

		err := rows.Scan(
			&siteNumber,
			&facilityName,
			&airport.Faa,
			&icao,
			&stateCode,
			&stateFull,
			&county,
			&city,
			&ownershipType,
			&useType,
			&manager,
			&managerPhone,
			&latitude,
			&longitude,
			&airportStatus,
			&weather,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan airport row: %w", err)
		}

		airports = append(airports, airport)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return airports, nil
}

// UpdateAirport updates an existing airport by FAA code (plain UPDATE).
func (r *Repository) UpdateAirport(airport *domain.Airport) error {
	query := `
		UPDATE airport SET
			site_number = $2,
			facility_name = $3,
			icao = $4,
			state_code = $5,
			state_full = $6,
			county = $7,
			city = $8,
			ownership_type = $9,
			use_type = $10,
			manager = $11,
			manager_phone = $12,
			latitude = $13,
			longitude = $14,
			airport_status = $15,
			weather = $16
		WHERE faa = $1`

	result, err := r.db.Exec(query,
		airport.Faa,
		airport.SiteNumber,
		airport.FacilityName,
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
		return fmt.Errorf("failed to update airport %s: %w", airport.Faa, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected for %s: %w", airport.Faa, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no airport found to update for %s", airport.Faa)
	}

	return nil
}

// DeleteByFAA deletes an airport by its FAA identifier.
func (r *Repository) DeleteByFAA(faa string) error {
	query := `DELETE FROM airport WHERE faa = $1`

	result, err := r.db.Exec(query, faa)
	if err != nil {
		return fmt.Errorf("failed to delete airport %s: %w", faa, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected for %s: %w", faa, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no airport found for %s", faa)
	}

	return nil
}
