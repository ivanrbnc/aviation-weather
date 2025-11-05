package service

import (
	"encoding/json"
	"fmt"
	"io"
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
	// Note: Aviation API endpoint; assumes public or key in query if needed (add ?key=... if required)
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

	var airports map[string][]domain.Airport
	if err := json.Unmarshal(body, &airports); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Aviation API response: %w", err)
	}

	airportList, ok := airports[faa]
	if !ok || len(airportList) == 0 {
		return nil, nil // Not found
	}

	// Take first since FAA is unique
	return &airportList[0], nil
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
		return "", fmt.Errorf("weather API returned status: %s", resp.Status) // Fixed: lowercase "weather"
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
