package repository

import (
	"database/sql"
	"fmt"

	"aviation-weather/internal/domain"
)

type Repository struct {
	db *sql.DB
}

type RepositoryInterface interface {
	CreateAirport(airport *domain.Airport) error
	UpdateAirport(airport *domain.Airport) error
	DeleteByFAA(faa string) error
	GetAllAirports() ([]domain.Airport, error)
	GetAirportByFAA(faaFilter string) (*domain.Airport, error)
}

func NewRepository(db *sql.DB) RepositoryInterface {
	return &Repository{db: db}
}

// Create inserts a new airport record if it does not already exist.
func (r *Repository) CreateAirport(airport *domain.Airport) error {
	query := `
		INSERT INTO airport (
			site_number, facility_name, faa, icao, state_code, state_full, county,
			city, ownership_type, use_type, manager, manager_phone,
			latitude, longitude, airport_status, weather
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		ON CONFLICT (faa) DO NOTHING
	`

	result, err := r.db.Exec(
		query,
		airport.SiteNumber, airport.FacilityName, airport.Faa, airport.Icao,
		airport.StateCode, airport.StateFull, airport.County, airport.City,
		airport.OwnershipType, airport.UseType, airport.Manager, airport.ManagerPhone,
		airport.Latitude, airport.Longitude, airport.AirportStatus, airport.Weather,
	)
	if err != nil {
		return fmt.Errorf("failed to create airport: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected for %s: %w", airport.Faa, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no airport found to create for %s", airport.Faa)
	}

	return nil
}

// UpdateAirport updates an existing airport by FAA code.
func (r *Repository) UpdateAirport(airport *domain.Airport) error {
	query := `
		UPDATE airport
		SET site_number = $2, facility_name = $3, icao = $4, state_code = $5, state_full = $6,
		    county = $7, city = $8, ownership_type = $9, use_type = $10, manager = $11,
		    manager_phone = $12, latitude = $13, longitude = $14,
		    airport_status = $15, weather = $16
		WHERE faa = $1
	`

	result, err := r.db.Exec(
		query,
		airport.Faa, airport.SiteNumber, airport.FacilityName, airport.Icao,
		airport.StateCode, airport.StateFull, airport.County, airport.City,
		airport.OwnershipType, airport.UseType, airport.Manager, airport.ManagerPhone,
		airport.Latitude, airport.Longitude, airport.AirportStatus, airport.Weather,
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

// GetAllAirports fetches all airports from the DB.
func (r *Repository) GetAllAirports() ([]domain.Airport, error) {
	query := `
		SELECT site_number, facility_name, faa, icao, state_code, state_full, county,
		       city, ownership_type, use_type, manager, manager_phone,
		       latitude, longitude, airport_status, weather
		FROM airport
		ORDER BY faa
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all airports: %w", err)
	}
	defer rows.Close()

	var airports []domain.Airport
	for rows.Next() {
		var a domain.Airport
		var siteNumber, facilityName, faa, icao, stateCode, stateFull,
			county, city, ownershipType, useType, manager, managerPhone,
			latitude, longitude, airportStatus, weather sql.NullString

		if err := rows.Scan(
			&siteNumber, &facilityName, &faa, &icao, &stateCode, &stateFull,
			&county, &city, &ownershipType, &useType, &manager, &managerPhone,
			&latitude, &longitude, &airportStatus, &weather,
		); err != nil {
			return nil, fmt.Errorf("failed to scan airport row: %w", err)
		}

		a.SiteNumber = siteNumber.String
		a.FacilityName = facilityName.String
		a.Faa = faa.String
		a.Icao = icao.String
		a.StateCode = stateCode.String
		a.StateFull = stateFull.String
		a.County = county.String
		a.City = city.String
		a.OwnershipType = ownershipType.String
		a.UseType = useType.String
		a.Manager = manager.String
		a.ManagerPhone = managerPhone.String
		a.Latitude = latitude.String
		a.Longitude = longitude.String
		a.AirportStatus = airportStatus.String
		a.Weather = weather.String

		airports = append(airports, a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return airports, nil
}

// GetAirportByFAA fetches an airport by FAA code.
func (r *Repository) GetAirportByFAA(faaFilter string) (*domain.Airport, error) {
	query := `
        SELECT site_number, facility_name, faa, icao, state_code, state_full, county,
               city, ownership_type, use_type, manager, manager_phone,
               latitude, longitude, airport_status, weather
        FROM airport
        WHERE faa = $1
    `

	rows, err := r.db.Query(query, faaFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to query airport: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		// No rows found, return nil, nil to indicate no airport exists
		return nil, nil
	}

	var a domain.Airport
	var siteNumber, facilityName, faa, icao, stateCode, stateFull,
		county, city, ownershipType, useType, manager, managerPhone,
		latitude, longitude, airportStatus, weather sql.NullString

	if err := rows.Scan(
		&siteNumber, &facilityName, &faa, &icao, &stateCode, &stateFull,
		&county, &city, &ownershipType, &useType, &manager, &managerPhone,
		&latitude, &longitude, &airportStatus, &weather,
	); err != nil {
		return nil, fmt.Errorf("failed to scan airport row: %w", err)
	}

	a.SiteNumber = siteNumber.String
	a.FacilityName = facilityName.String
	a.Faa = faa.String
	a.Icao = icao.String
	a.StateCode = stateCode.String
	a.StateFull = stateFull.String
	a.County = county.String
	a.City = city.String
	a.OwnershipType = ownershipType.String
	a.UseType = useType.String
	a.Manager = manager.String
	a.ManagerPhone = managerPhone.String
	a.Latitude = latitude.String
	a.Longitude = longitude.String
	a.AirportStatus = airportStatus.String
	a.Weather = weather.String

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return &a, nil
}
