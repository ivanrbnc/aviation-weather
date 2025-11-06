package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"aviation-weather/config"
	"aviation-weather/internal/domain"
	"aviation-weather/internal/repository"
)

type Service struct {
	repo       *repository.Repository
	cfg        *config.Config
	httpClient *http.Client
}

func NewService(repo *repository.Repository, cfg *config.Config) *Service {
	return &Service{
		repo:       repo,
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// GetAirportWithWeather fetches airport details and persists/updates in DB
func (s *Service) GetAirportWithWeather(faa string) (*domain.Airport, error) {
	// Step 1: Fetch from Aviation API
	airport, err := s.fetchAirportFromAviationAPI(faa)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch airport from Aviation API: %w", err)
	}
	if airport == nil {
		return nil, fmt.Errorf("no airport found for FAA code %s", faa)
	}

	// Step 2: Enrich with weather
	weatherText, err := s.fetchWeather(airport.City)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weather for city %s: %w", airport.City, err)
	}
	airport.Weather = weatherText

	// Step 3: Persist via repo (creates if new, skips/ignores if exists—but weather isn't updated here; add UpdateWeather if needed)
	if err := s.repo.Create(airport); err != nil {
		return nil, fmt.Errorf("failed to persist airport: %w", err)
	}

	return airport, nil
}

// fetchAirportFromAviationAPI: Internal helper for Aviation API call (based on your demo).
func (s *Service) fetchAirportFromAviationAPI(faa string) (*domain.Airport, error) {
	apiURL := fmt.Sprintf("https://api.aviationapi.com/v1/airports?apt=%s", url.QueryEscape(faa))

	resp, err := s.httpClient.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("aviation API returned status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Temporary struct for unmarshal (API returns lat/long as strings)
	type apiAirport struct {
		SiteNumber    string `json:"site_number"`
		FacilityName  string `json:"facility_name"`
		Faa           string `json:"faa"`
		Icao          string `json:"icao"`
		StateCode     string `json:"state_code"`
		StateFull     string `json:"state_full"`
		County        string `json:"county"`
		City          string `json:"city"`
		OwnershipType string `json:"ownership_type"`
		UseType       string `json:"use_type"`
		Manager       string `json:"manager"`
		ManagerPhone  string `json:"manager_phone"`
		Latitude      string `json:"latitude"`  // String for unmarshal
		Longitude     string `json:"longitude"` // String for unmarshal
		AirportStatus string `json:"airport_status"`
		// Weather not in API
	}

	var airports map[string][]apiAirport
	if err := json.Unmarshal(body, &airports); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Aviation API response: %w", err)
	}

	airportList, ok := airports[faa]
	if !ok || len(airportList) == 0 {
		return nil, nil // Not found
	}

	// Take first since FAA is unique; parse lat/long to float64
	apiApt := airportList[0]

	return &domain.Airport{
		SiteNumber:    apiApt.SiteNumber,
		FacilityName:  apiApt.FacilityName,
		Faa:           apiApt.Faa,
		Icao:          apiApt.Icao,
		StateCode:     apiApt.StateCode,
		StateFull:     apiApt.StateFull,
		County:        apiApt.County,
		City:          apiApt.City,
		OwnershipType: apiApt.OwnershipType,
		UseType:       apiApt.UseType,
		Manager:       apiApt.Manager,
		ManagerPhone:  apiApt.ManagerPhone,
		Latitude:      apiApt.Latitude,
		Longitude:     apiApt.Longitude,
		AirportStatus: apiApt.AirportStatus,
		Weather:       "", // Filled later
	}, nil
}

// fetchWeather: Internal helper for Weather API
func (s *Service) fetchWeather(city string) (string, error) {
	if s.cfg.WeatherAPIKey == "" {
		return "Weather API key not configured", fmt.Errorf("missing WEATHER_API_KEY")
	}

	apiURL := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s",
		url.QueryEscape(s.cfg.WeatherAPIKey), url.QueryEscape(city))

	resp, err := s.httpClient.Get(apiURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("weather API returned status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var weather domain.WeatherResponse
	if err := json.Unmarshal(body, &weather); err != nil {
		return "", fmt.Errorf("failed to unmarshal Weather API response: %w", err)
	}

	return weather.Current.Condition.Text, nil
}

func (s *Service) SyncAllAirports() (int, error) {
	airports, err := s.repo.GetAllAirports()
	if err != nil {
		return 0, fmt.Errorf("failed to get all airports: %w", err)
	}
	if len(airports) == 0 {
		return 0, fmt.Errorf("no airports to sync")
	}

	updated := 0
	var errors []string
	for _, airport := range airports {
		// Fetch real aviation data (overwrites dummies)
		realAirport, err := s.fetchAirportFromAviationAPI(airport.Faa)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: aviation fetch failed - %v", airport.Faa, err))
			continue
		}
		if realAirport == nil {
			errors = append(errors, fmt.Sprintf("%s: no aviation data found", airport.Faa))
			continue
		}

		// Use real data (keep faa as key)
		realAirport.Weather = airport.Weather // Temp; will overwrite below

		// Enrich with weather
		weatherText, err := s.fetchWeather(realAirport.City)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: weather fetch failed - %v", realAirport.Faa, err))
			realAirport.Weather = "" // Or skip update; here we set empty
		} else {
			realAirport.Weather = weatherText
		}

		// Update in DB
		if err := s.repo.UpdateAirport(realAirport); err != nil {
			errors = append(errors, fmt.Sprintf("%s: update failed - %v", realAirport.Faa, err))
			continue
		}
		updated++
		log.Printf("Synced %s: %s, %s", realAirport.Faa, realAirport.FacilityName, realAirport.Weather)

		// Rate limiting delay (adjust as needed)
		time.Sleep(200 * time.Millisecond)
	}

	if len(errors) > 0 {
		return updated, fmt.Errorf("synced %d/%d; errors: %s", updated, len(airports), strings.Join(errors, "; "))
	}
	return updated, nil
}

// GetAllAirportsWithWeather fetches all airports from DB and enriches each with current weather.
// Returns a slice of enriched airports.
func (s *Service) GetAllAirportsWithWeather() ([]domain.Airport, error) {
	airports, err := s.repo.GetAllAirports()
	if err != nil {
		return nil, fmt.Errorf("failed to get all airports: %w", err)
	}
	if len(airports) == 0 {
		return []domain.Airport{}, nil // Empty list, no error
	}

	enriched := make([]domain.Airport, len(airports))
	for i, airport := range airports {
		// Skip weather fetch if city is empty (avoids 400 Bad Request)
		var weatherText string
		if airport.City != "" {
			weatherText, err = s.fetchWeather(airport.City)
			if err != nil {
				log.Printf("warning: failed to fetch weather for %s (%s): %v", airport.Faa, airport.City, err)
				weatherText = "Weather unavailable"
			}
		} else {
			log.Printf("skipping weather for %s: empty city", airport.Faa)
			weatherText = "City not available"
		}

		enrichedAirport := airport
		enrichedAirport.Weather = weatherText
		enriched[i] = enrichedAirport
	}

	return enriched, nil
}

// DeleteAirportByFAA deletes an airport by FAA code.
func (s *Service) DeleteAirportByFAA(faa string) error {
	return s.repo.DeleteByFAA(faa)
}
