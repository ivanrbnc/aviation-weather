package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
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
		repo: repo,
		cfg:  cfg,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *Service) CreateAirport(a *domain.Airport) error {
	return s.repo.Create(a)
}

func (s *Service) UpdateAirport(a *domain.Airport) error {
	return s.repo.UpdateAirport(a)
}

func (s *Service) GetAirportByFAA(faa string) (*domain.Airport, error) {
	return s.repo.GetAirportByFAA(faa)
}

// GetAndSaveAirportWithWeather fetches airport details and persists/updates in DB
func (s *Service) GetAndSaveAirportWithWeather(faa string) (*domain.Airport, error) {
	// Step 1: Fetch from Aviation API
	airport, err := s.fetchAirportFromAviationAPI(faa)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch airport from Aviation API: %w", err)
	}
	if airport == nil {
		return nil, fmt.Errorf("no airport found for FAA code %s", faa)
	}

	// Step 2: Enrich with weather
	weatherText, err := s.fetchWeatherFromWeatherAPI(airport.City)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weather for city %s: %w", airport.City, err)
	}
	airport.Weather = weatherText

	// Step 3: Persist via repo
	if err := s.repo.UpdateAirport(airport); err != nil {
		return nil, fmt.Errorf("failed to persist airport: %w", err)
	}

	return airport, nil
}

// fetchAirportFromAviationAPI: Internal helper for Aviation API call
func (s *Service) fetchAirportFromAviationAPI(faa string) (*domain.Airport, error) {
	apiURL := fmt.Sprintf("https://api.aviationapi.com/v1/airports?apt=%s", url.QueryEscape(faa))
	resp, err := http.Get(apiURL)
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

	var airports map[string][]domain.Airport
	if err := json.Unmarshal(body, &airports); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Aviation API response: %w", err)
	}

	var airport domain.Airport
	if len(airports[faa]) > 0 {
		airport = airports[faa][0]
	}

	return &airport, nil
}

// fetchWeatherFromWeatherAPI: Internal helper for Weather API
func (s *Service) fetchWeatherFromWeatherAPI(city string) (string, error) {
	if s.cfg.WeatherAPIKey == "" {
		return "Weather API key not configured", fmt.Errorf("missing WEATHER_API_KEY")
	}

	apiURL := fmt.Sprintf(
		"https://api.weatherapi.com/v1/current.json?key=%s&q=%s",
		url.QueryEscape(s.cfg.WeatherAPIKey),
		url.QueryEscape(city),
	)

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
		airport, err := s.GetAndSaveAirportWithWeather(airport.Faa)
		if err != nil {
			log.Println(err)
			continue
		}
		if airport == nil {
			log.Println("Error: Airport not found")
			continue
		}
		updated++
		log.Printf("Synced %s: %s, %s", airport.Faa, airport.FacilityName, airport.Weather)

		time.Sleep(200 * time.Millisecond)
	}

	if len(errors) > 0 {
		return updated, fmt.Errorf("synced %d/%d", updated, len(airports))
	}

	return updated, nil
}

// GetAllAirportsWithWeather fetches all airports from DB and enriches each with current weather
func (s *Service) GetAllAirportsWithWeather() ([]domain.Airport, error) {
	airports, err := s.repo.GetAllAirports()
	if err != nil {
		return nil, fmt.Errorf("failed to get all airports: %w", err)
	}

	if len(airports) == 0 {
		return []domain.Airport{}, nil
	}

	return airports, nil
}

// DeleteAirportByFAA deletes an airport by FAA code
func (s *Service) DeleteAirportByFAA(faa string) error {
	return s.repo.DeleteByFAA(faa)
}
