// service_test.go
package service

import (
	"fmt"
	"testing"

	"aviation-weather/config"
	"aviation-weather/internal/domain"
	mocks "aviation-weather/internal/mock" // No conflict with testify

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func TestCreateAirport(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*mocks.RepositoryMock)
		err       error
	}{
		{
			name: "success",
			setupMock: func(m *mocks.RepositoryMock) {
				m.On("CreateAirport", &sampleAirport).Return(nil)
			},
			err: nil,
		},
		{
			name: "repo error",
			setupMock: func(m *mocks.RepositoryMock) {
				m.On("CreateAirport", &sampleAirport).Return(assert.AnError)
			},
			err: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.RepositoryMock{} // Use the repo mock to fake the return
			tt.setupMock(mockRepo)
			s := NewService(mockRepo, &config.Config{})

			err := s.CreateAirport(&sampleAirport)
			assert.Equal(t, tt.err, err)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUpdateAirport(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*mocks.RepositoryMock)
		err       error
	}{
		{
			name: "success",
			setupMock: func(m *mocks.RepositoryMock) {
				m.On("UpdateAirport", &sampleAirport).Return(nil)
			},
			err: nil,
		},
		{
			name: "repo error",
			setupMock: func(m *mocks.RepositoryMock) {
				m.On("UpdateAirport", &sampleAirport).Return(assert.AnError)
			},
			err: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.RepositoryMock{}
			tt.setupMock(mockRepo)
			s := NewService(mockRepo, &config.Config{})

			err := s.UpdateAirport(&sampleAirport)
			assert.Equal(t, tt.err, err)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestDeleteAirportByFAA(t *testing.T) {
	tests := []struct {
		name      string
		faa       string
		setupMock func(*mocks.RepositoryMock)
		err       error
	}{
		{
			name: "success",
			faa:  "TST",
			setupMock: func(m *mocks.RepositoryMock) {
				m.On("DeleteByFAA", "TST").Return(nil)
			},
			err: nil,
		},
		{
			name: "repo error",
			faa:  "ERR",
			setupMock: func(m *mocks.RepositoryMock) {
				m.On("DeleteByFAA", "ERR").Return(assert.AnError)
			},
			err: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.RepositoryMock{}
			tt.setupMock(mockRepo)
			s := NewService(mockRepo, &config.Config{})

			err := s.DeleteAirportByFAA(tt.faa)
			assert.Equal(t, tt.err, err)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetAirportByFAA(t *testing.T) {
	tests := []struct {
		name      string
		faa       string
		setupMock func(*mocks.RepositoryMock)
		expected  *domain.Airport
		err       error
	}{
		{
			name: "success",
			faa:  "TST",
			setupMock: func(m *mocks.RepositoryMock) {
				m.On("GetAirportByFAA", "TST").Return(&sampleAirport, nil)
			},
			expected: &sampleAirport,
			err:      nil,
		},
		{
			name: "repo error",
			faa:  "ERR",
			setupMock: func(m *mocks.RepositoryMock) {
				m.On("GetAirportByFAA", "ERR").Return((*domain.Airport)(nil), assert.AnError)
			},
			expected: nil,
			err:      fmt.Errorf("failed to get airport for ERR: %w", assert.AnError),
		},
		{
			name: "not found",
			faa:  "NF",
			setupMock: func(m *mocks.RepositoryMock) {
				m.On("GetAirportByFAA", "NF").Return((*domain.Airport)(nil), nil)
			},
			expected: nil,
			err:      fmt.Errorf("no airport found for NF"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.RepositoryMock{}
			tt.setupMock(mockRepo)
			s := NewService(mockRepo, &config.Config{})

			airport, err := s.GetAirportByFAA(tt.faa)
			assert.Equal(t, tt.expected, airport)
			if tt.err != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetAllAirports(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*mocks.RepositoryMock)
		expected  []domain.Airport
		err       error
	}{
		{
			name: "success",
			setupMock: func(m *mocks.RepositoryMock) {
				m.On("GetAllAirports").Return([]domain.Airport{sampleAirport}, nil)
			},
			expected: []domain.Airport{sampleAirport},
			err:      nil,
		},
		{
			name: "repo error",
			setupMock: func(m *mocks.RepositoryMock) {
				m.On("GetAllAirports").Return([]domain.Airport{}, assert.AnError)
			},
			expected: nil,
			err:      fmt.Errorf("failed to get airports: %w", assert.AnError),
		},
		{
			name: "no airports",
			setupMock: func(m *mocks.RepositoryMock) {
				m.On("GetAllAirports").Return([]domain.Airport{}, nil)
			},
			expected: []domain.Airport{},
			err:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.RepositoryMock{}
			tt.setupMock(mockRepo)
			s := NewService(mockRepo, &config.Config{})

			airports, err := s.GetAllAirports()
			assert.Equal(t, tt.expected, airports)
			if tt.err != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSyncAirportByFAA(t *testing.T) {
	tests := []struct {
		name      string
		faa       string
		setupMock func(*mocks.RepositoryMock)
		expected  *domain.Airport
		err       error
	}{
		{
			name: "repo update error",
			faa:  "TST",
			setupMock: func(m *mocks.RepositoryMock) {
				m.On("UpdateAirport", mock.Anything).Return(assert.AnError)
			},
			expected: nil,
			err:      fmt.Errorf("failed to update airport TST: %w", assert.AnError),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.RepositoryMock{}
			tt.setupMock(mockRepo)
			s := NewService(mockRepo, &config.Config{}).(*Service) // cast to concrete type so internal helper can be used

			// mock external API calls
			s.FetchAirportFromAviationAPI = func(faa string) (*domain.Airport, error) {
				return &domain.Airport{Faa: faa, City: "Jakarta"}, nil
			}
			s.FetchWeatherFromWeatherAPI = func(city string) (string, error) {
				return "Sunny", nil
			}

			airport, err := s.SyncAirportByFAA(tt.faa)
			assert.Equal(t, tt.expected, airport)
			if tt.err != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.err.Error())
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSyncAllAirports(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*mocks.RepositoryMock)
		expected  int
		err       error
	}{
		{
			name: "no airports",
			setupMock: func(m *mocks.RepositoryMock) {
				m.On("GetAllAirports").Return([]domain.Airport{}, nil)
			},
			expected: 0,
			err:      fmt.Errorf("no airports to sync"),
		},
		{
			name: "repo get error",
			setupMock: func(m *mocks.RepositoryMock) {
				m.On("GetAllAirports").Return([]domain.Airport{}, assert.AnError)
			},
			expected: 0,
			err:      fmt.Errorf("failed to get airports: %w", assert.AnError),
		},
		{
			name: "successful sync with mocked APIs",
			setupMock: func(m *mocks.RepositoryMock) {
				m.On("GetAllAirports").Return([]domain.Airport{
					{Faa: "TST", FacilityName: "Test Airport", City: "Jakarta"},
				}, nil)
				m.On("UpdateAirport", mock.Anything).Return(nil)
			},
			expected: 1,
			err:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.RepositoryMock{}
			tt.setupMock(mockRepo)

			s := NewService(mockRepo, &config.Config{}).(*Service) // cast to concrete type so internal helper can be used

			// mock batch API call (updated to return []domain.Airport)
			s.FetchAirportsFromAviationAPI = func(faaList []string) ([]domain.Airport, error) {
				airports := []domain.Airport{}
				for _, faa := range faaList {
					airports = append(airports, domain.Airport{
						Faa:          faa,
						City:         "Jakarta",
						FacilityName: "Mock Airport",
					})
				}
				return airports, nil
			}

			// mock weather API call
			s.FetchWeatherFromWeatherAPI = func(city string) (string, error) {
				return "Clear skies", nil
			}

			updated, err := s.SyncAllAirports()
			assert.Equal(t, tt.expected, updated)

			if tt.err != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
