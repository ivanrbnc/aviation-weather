package repository

import (
	"errors"
	"testing"

	"aviation-weather/internal/domain"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

var sampleAirport = domain.Airport{
	SiteNumber:    "12345",
	FacilityName:  "Test Airport",
	Faa:           "TST",
	Icao:          "KTST",
	StateCode:     "CA",
	StateFull:     "California",
	County:        "Test County",
	City:          "Test City",
	OwnershipType: "Public",
	UseType:       "Public Use",
	Manager:       "Test Manager",
	ManagerPhone:  "123-456-7890",
	Latitude:      "34.0522",
	Longitude:     "-118.2437",
	AirportStatus: "Open",
	Weather:       "Clear",
}

const anErrorMsg = "assert.AnError general error for testing"

func TestCreateAirport(t *testing.T) {
	tests := []struct {
		name        string
		setupDB     func(sqlmock.Sqlmock)
		expected    []domain.Airport
		expectedErr string
	}{
		{
			name: "success",
			setupDB: func(mock sqlmock.Sqlmock) {
				query := `INSERT INTO airport \(
					site_number, facility_name, faa, icao, state_code, state_full, county,
					city, ownership_type, use_type, manager, manager_phone,
					latitude, longitude, airport_status, weather
				\)
				VALUES \(\$1, \$2, \$3, \$4, \$5, \$6, \$7, \$8, \$9, \$10, \$11, \$12, \$13, \$14, \$15, \$16\)
				ON CONFLICT \(faa\) DO NOTHING`
				mock.ExpectExec(query).
					WithArgs(
						sampleAirport.SiteNumber, sampleAirport.FacilityName, sampleAirport.Faa, sampleAirport.Icao,
						sampleAirport.StateCode, sampleAirport.StateFull, sampleAirport.County, sampleAirport.City,
						sampleAirport.OwnershipType, sampleAirport.UseType, sampleAirport.Manager, sampleAirport.ManagerPhone,
						sampleAirport.Latitude, sampleAirport.Longitude, sampleAirport.AirportStatus, sampleAirport.Weather,
					).
					WillReturnResult(sqlmock.NewResult(1, 1)) // 1 row affected
			},
			expectedErr: "",
		},
		{
			name: "db exec error",
			setupDB: func(mock sqlmock.Sqlmock) {
				query := `INSERT INTO airport` // Partial match
				mock.ExpectExec(query).
					WillReturnError(errors.New(anErrorMsg))
			},
			expectedErr: "failed to create airport: " + anErrorMsg,
		},
		{
			name: "no rows affected",
			setupDB: func(mock sqlmock.Sqlmock) {
				query := `INSERT INTO airport` // Partial match
				mock.ExpectExec(query).
					WillReturnResult(sqlmock.NewResult(1, 0)) // 0 rows affected
			},
			expectedErr: "no airport found to create for TST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Fake DB connection and a mock controller
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			r := NewRepository(db)
			tt.setupDB(mock) // Mock query

			err = r.CreateAirport(&sampleAirport)
			if tt.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expectedErr)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUpdateAirport(t *testing.T) {
	tests := []struct {
		name        string
		setupDB     func(sqlmock.Sqlmock)
		expectedErr string
	}{
		{
			name: "success",
			setupDB: func(mock sqlmock.Sqlmock) {
				query := `UPDATE airport
					SET site_number = \$2, facility_name = \$3, icao = \$4, state_code = \$5, state_full = \$6,
					    county = \$7, city = \$8, ownership_type = \$9, use_type = \$10, manager = \$11,
					    manager_phone = \$12, latitude = \$13, longitude = \$14,
					    airport_status = \$15, weather = \$16
					WHERE faa = \$1`
				mock.ExpectExec(query).
					WithArgs(
						sampleAirport.Faa, sampleAirport.SiteNumber, sampleAirport.FacilityName, sampleAirport.Icao,
						sampleAirport.StateCode, sampleAirport.StateFull, sampleAirport.County, sampleAirport.City,
						sampleAirport.OwnershipType, sampleAirport.UseType, sampleAirport.Manager, sampleAirport.ManagerPhone,
						sampleAirport.Latitude, sampleAirport.Longitude, sampleAirport.AirportStatus, sampleAirport.Weather,
					).
					WillReturnResult(sqlmock.NewResult(1, 1)) // 1 row affected
			},
			expectedErr: "",
		},
		{
			name: "db exec error",
			setupDB: func(mock sqlmock.Sqlmock) {
				query := `UPDATE airport` // Partial match
				mock.ExpectExec(query).
					WillReturnError(errors.New(anErrorMsg))
			},
			expectedErr: "failed to update airport TST: " + anErrorMsg,
		},
		{
			name: "no rows affected",
			setupDB: func(mock sqlmock.Sqlmock) {
				query := `UPDATE airport` // Partial match
				mock.ExpectExec(query).
					WillReturnResult(sqlmock.NewResult(1, 0)) // 0 rows affected
			},
			expectedErr: "no airport found to update for TST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			r := NewRepository(db)
			tt.setupDB(mock)

			err = r.UpdateAirport(&sampleAirport)
			if tt.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expectedErr)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDeleteByFAA(t *testing.T) {
	tests := []struct {
		name        string
		faa         string
		setupDB     func(sqlmock.Sqlmock)
		expectedErr string
	}{
		{
			name: "success",
			faa:  "TST",
			setupDB: func(mock sqlmock.Sqlmock) {
				query := `DELETE FROM airport WHERE faa = \$1`
				mock.ExpectExec(query).
					WithArgs("TST").
					WillReturnResult(sqlmock.NewResult(1, 1)) // 1 row affected
			},
			expectedErr: "",
		},
		{
			name: "db exec error",
			faa:  "ERR",
			setupDB: func(mock sqlmock.Sqlmock) {
				query := `DELETE FROM airport` // Partial match
				mock.ExpectExec(query).
					WillReturnError(errors.New(anErrorMsg))
			},
			expectedErr: "failed to delete airport ERR: " + anErrorMsg,
		},
		{
			name: "no rows affected",
			faa:  "NF",
			setupDB: func(mock sqlmock.Sqlmock) {
				query := `DELETE FROM airport` // Partial match
				mock.ExpectExec(query).
					WillReturnResult(sqlmock.NewResult(1, 0)) // 0 rows affected
			},
			expectedErr: "no airport found for NF",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			r := NewRepository(db)
			tt.setupDB(mock)

			err = r.DeleteByFAA(tt.faa)
			if tt.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expectedErr)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetAllAirports(t *testing.T) {
	const anErrorMsg = "assert.AnError general error for testing"

	fullCols := []string{
		"site_number", "facility_name", "faa", "icao", "state_code", "state_full", "county",
		"city", "ownership_type", "use_type", "manager", "manager_phone",
		"latitude", "longitude", "airport_status", "weather",
	}
	mismatchCols := fullCols[:15] // Fewer columns to cause scan mismatch (15<16)

	tests := []struct {
		name        string
		setupDB     func(sqlmock.Sqlmock)
		expected    []domain.Airport
		expectedErr string
	}{
		{
			name: "success",
			setupDB: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(fullCols).AddRow(
					sampleAirport.SiteNumber, sampleAirport.FacilityName, sampleAirport.Faa, sampleAirport.Icao,
					sampleAirport.StateCode, sampleAirport.StateFull, sampleAirport.County,
					sampleAirport.City, sampleAirport.OwnershipType, sampleAirport.UseType, sampleAirport.Manager, sampleAirport.ManagerPhone,
					sampleAirport.Latitude, sampleAirport.Longitude, sampleAirport.AirportStatus, sampleAirport.Weather,
				)
				query := `SELECT site_number, facility_name, faa, icao, state_code, state_full, county,
				       city, ownership_type, use_type, manager, manager_phone,
				       latitude, longitude, airport_status, weather
				FROM airport
				ORDER BY faa`
				mock.ExpectQuery(query).
					WillReturnRows(rows)
			},
			expected:    []domain.Airport{sampleAirport},
			expectedErr: "",
		},
		{
			name: "db query error",
			setupDB: func(mock sqlmock.Sqlmock) {
				query := `SELECT site_number, facility_name, faa, icao, state_code, state_full, county,
				       city, ownership_type, use_type, manager, manager_phone,
				       latitude, longitude, airport_status, weather
				FROM airport
				ORDER BY faa`
				mock.ExpectQuery(query).
					WillReturnError(errors.New(anErrorMsg))
			},
			expected:    nil,
			expectedErr: "failed to query all airports: " + anErrorMsg,
		},
		{
			name: "no rows",
			setupDB: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(fullCols)
				query := `SELECT site_number, facility_name, faa, icao, state_code, state_full, county,
				       city, ownership_type, use_type, manager, manager_phone,
				       latitude, longitude, airport_status, weather
				FROM airport
				ORDER BY faa`
				mock.ExpectQuery(query).
					WillReturnRows(rows)
			},
			expected:    nil,
			expectedErr: "",
		},
		{
			name: "scan error",
			setupDB: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(mismatchCols).AddRow(
					sampleAirport.FacilityName, sampleAirport.Faa, sampleAirport.Icao,
					sampleAirport.StateCode, sampleAirport.StateFull, sampleAirport.County,
					sampleAirport.City, sampleAirport.OwnershipType, sampleAirport.UseType, sampleAirport.Manager, sampleAirport.ManagerPhone,
					sampleAirport.Latitude, sampleAirport.Longitude, sampleAirport.AirportStatus, sampleAirport.Weather,
				)
				query := `SELECT site_number, facility_name, faa, icao, state_code, state_full, county,
				       city, ownership_type, use_type, manager, manager_phone,
				       latitude, longitude, airport_status, weather
				FROM airport
				ORDER BY faa`
				mock.ExpectQuery(query).
					WillReturnRows(rows)
			},
			expected:    nil,
			expectedErr: "failed to scan airport row: sql: expected 15 destination arguments in Scan, not 16",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			r := NewRepository(db)
			tt.setupDB(mock)

			airports, err := r.GetAllAirports()
			assert.Equal(t, tt.expected, airports)
			if tt.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expectedErr)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetAirportByFAA(t *testing.T) {
	const anErrorMsg = "assert.AnError general error for testing"

	fullCols := []string{
		"site_number", "facility_name", "faa", "icao", "state_code", "state_full", "county",
		"city", "ownership_type", "use_type", "manager", "manager_phone",
		"latitude", "longitude", "airport_status", "weather",
	}
	mismatchCols := fullCols[:15]

	tests := []struct {
		name        string
		faa         string
		setupDB     func(sqlmock.Sqlmock)
		expected    *domain.Airport
		expectedErr string
	}{
		{
			name: "success",
			faa:  "TST",
			setupDB: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(fullCols).AddRow(
					sampleAirport.SiteNumber, sampleAirport.FacilityName, sampleAirport.Faa, sampleAirport.Icao,
					sampleAirport.StateCode, sampleAirport.StateFull, sampleAirport.County,
					sampleAirport.City, sampleAirport.OwnershipType, sampleAirport.UseType, sampleAirport.Manager, sampleAirport.ManagerPhone,
					sampleAirport.Latitude, sampleAirport.Longitude, sampleAirport.AirportStatus, sampleAirport.Weather,
				)
				query := `SELECT site_number, facility_name, faa, icao, state_code, state_full, county,
                       city, ownership_type, use_type, manager, manager_phone,
                       latitude, longitude, airport_status, weather
                FROM airport
                WHERE faa = \$1`
				mock.ExpectQuery(query).
					WithArgs("TST").
					WillReturnRows(rows)
			},
			expected:    &sampleAirport,
			expectedErr: "",
		},
		{
			name: "db query error",
			faa:  "ERR",
			setupDB: func(mock sqlmock.Sqlmock) {
				query := `SELECT site_number, facility_name, faa, icao, state_code, state_full, county,
                       city, ownership_type, use_type, manager, manager_phone,
                       latitude, longitude, airport_status, weather
                FROM airport
                WHERE faa = \$1`
				mock.ExpectQuery(query).
					WithArgs("ERR").
					WillReturnError(errors.New(anErrorMsg))
			},
			expected:    nil,
			expectedErr: "failed to query airport: " + anErrorMsg,
		},
		{
			name: "not found",
			faa:  "NF",
			setupDB: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(fullCols)
				query := `SELECT site_number, facility_name, faa, icao, state_code, state_full, county,
                       city, ownership_type, use_type, manager, manager_phone,
                       latitude, longitude, airport_status, weather
                FROM airport
                WHERE faa = \$1`
				mock.ExpectQuery(query).
					WithArgs("NF").
					WillReturnRows(rows)
			},
			expected:    nil,
			expectedErr: "",
		},
		{
			name: "scan error",
			faa:  "SCAN",
			setupDB: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(mismatchCols).AddRow(
					sampleAirport.SiteNumber, sampleAirport.FacilityName, sampleAirport.Faa, sampleAirport.Icao,
					sampleAirport.StateCode, sampleAirport.StateFull, sampleAirport.County,
					sampleAirport.City, sampleAirport.OwnershipType, sampleAirport.UseType, sampleAirport.Manager, sampleAirport.ManagerPhone,
					sampleAirport.Latitude, sampleAirport.Longitude, sampleAirport.AirportStatus,
				)
				query := `SELECT site_number, facility_name, faa, icao, state_code, state_full, county,
                       city, ownership_type, use_type, manager, manager_phone,
                       latitude, longitude, airport_status, weather
                FROM airport
                WHERE faa = \$1`
				mock.ExpectQuery(query).
					WithArgs("SCAN").
					WillReturnRows(rows)
			},
			expected:    nil,
			expectedErr: "failed to scan airport row: sql: expected 15 destination arguments in Scan, not 16",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			r := NewRepository(db)
			tt.setupDB(mock)

			airport, err := r.GetAirportByFAA(tt.faa)
			assert.Equal(t, tt.expected, airport)
			if tt.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expectedErr)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
