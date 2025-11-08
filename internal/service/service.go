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

func (s *Service) DeleteAirportByFAA(faa string) error {
	return s.repo.DeleteByFAA(faa)
}

func (s *Service) GetAirportByFAA(faa string) (*domain.Airport, error) {
	airport, err := s.repo.GetAirportByFAA(faa)
	if err != nil {
		return nil, fmt.Errorf("failed to get airport for %s: %w", faa, err)
	}

	if airport == nil {
		return nil, fmt.Errorf("no airport found for %s", faa)
	}

	return airport, nil
}

func (s *Service) GetAllAirports() ([]domain.Airport, error) {
	airports, err := s.repo.GetAllAirports()
	if err != nil {
		return nil, fmt.Errorf("failed to get airports: %w", err)
	}

	if len(airports) == 0 {
		return nil, fmt.Errorf("no airports found")
	}

	return airports, nil
}

func (s *Service) FetchAirportFromAviationAPI(faa string) (*domain.Airport, error) {
	apiURL := fmt.Sprintf("https://api.aviationapi.com/v1/airports?apt=%s", url.QueryEscape(faa))
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed for %s: %w", faa, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned %s for %s", resp.Status, faa)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response for %s: %w", faa, err)
	}

	var airports map[string][]domain.Airport
	if err := json.Unmarshal(body, &airports); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response for %s: %w", faa, err)
	}

	var airport domain.Airport
	if len(airports[faa]) > 0 {
		airport = airports[faa][0]
	}

	return &airport, nil
}

func (s *Service) FetchWeatherFromWeatherAPI(city string) (string, error) {
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
		return "", fmt.Errorf("HTTP request failed for %s: %w", city, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned %s for %s", resp.Status, city)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response for %s: %w", city, err)
	}

	var weather domain.WeatherResponse
	if err := json.Unmarshal(body, &weather); err != nil {
		return "", fmt.Errorf("failed to unmarshal response for %s: %w", city, err)
	}

	return weather.Current.Condition.Text, nil
}

func (s *Service) SyncAirportByFAA(faa string) (*domain.Airport, error) {
	// Fetch from Aviation API
	airport, err := s.FetchAirportFromAviationAPI(faa)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch airport for %s: %w", faa, err)
	}

	if airport == nil {
		return nil, fmt.Errorf("no airport found for %s", faa)
	}

	// Fetch from Weather API
	weatherText, err := s.FetchWeatherFromWeatherAPI(airport.City)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weather for %s: %w", airport.City, err)
	}
	airport.Weather = weatherText

	// Save to DB
	if err := s.repo.UpdateAirport(airport); err != nil {
		return nil, fmt.Errorf("failed to update airport %s: %w", faa, err)
	}

	return airport, nil
}

func (s *Service) SyncAllAirports() (int, error) {
	airports, err := s.repo.GetAllAirports()
	if err != nil {
		return 0, fmt.Errorf("failed to get airports: %w", err)
	}

	if len(airports) == 0 {
		return 0, fmt.Errorf("no airports to sync")
	}

	updated := 0
	errorFound := 0

	for _, airport := range airports {

		// Loop the individual sync
		fetchedAirport, err := s.SyncAirportByFAA(airport.Faa)
		if err != nil {
			errorFound++
			log.Printf("ERROR: Failed to sync %s (%s): %v", airport.Faa, airport.FacilityName, err)
			continue
		}

		if fetchedAirport == nil {
			errorFound++
			log.Printf("ERROR: Airport %s not found", airport.Faa)
			continue
		}

		updated++
		log.Printf("INFO: Synced %s (%s) in %s: %s", fetchedAirport.Faa, fetchedAirport.FacilityName, fetchedAirport.City, fetchedAirport.Weather)

		time.Sleep(200 * time.Millisecond)
	}

	if errorFound > 0 && updated > 0 {
		log.Printf("INFO: Partial sync (%d/%d)", updated, len(airports))
		return updated, nil
	}

	if errorFound > 0 && updated <= 0 {
		return 0, fmt.Errorf("failed to sync all airports")
	}

	return updated, nil
}
