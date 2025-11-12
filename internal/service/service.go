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
	repo       repository.RepositoryInterface
	cfg        *config.Config
	httpClient *http.Client

	// Internal helper so that it can be overriden
	FetchAirportFromAviationAPI  func(faa string) (*domain.Airport, error)
	FetchAirportsFromAviationAPI func(faa []string) ([]domain.Airport, error)
	FetchWeatherFromWeatherAPI   func(city string) (string, error)
}

type ServiceInterface interface {
	CreateAirport(a *domain.Airport) error
	UpdateAirport(a *domain.Airport) error
	DeleteAirportByFAA(faa string) error
	GetAirportByFAA(faa string) (*domain.Airport, error)
	GetAllAirports() ([]domain.Airport, error)
	SyncAirportByFAA(faa string) (*domain.Airport, error)
	SyncAllAirports() (int, error)
}

func NewService(repo repository.RepositoryInterface, cfg *config.Config) ServiceInterface {
	s := &Service{
		repo: repo,
		cfg:  cfg,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
	s.FetchAirportFromAviationAPI = s.fetchAirportFromAviationAPI
	s.FetchAirportsFromAviationAPI = s.fetchAirportsFromAviationAPI
	s.FetchWeatherFromWeatherAPI = s.fetchWeatherFromWeatherAPI
	return s
}

func (s *Service) CreateAirport(a *domain.Airport) error {
	return s.repo.CreateAirport(a)
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
		return []domain.Airport{}, nil
	}

	return airports, nil
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

	type result struct {
		updated int
		errors  int
	}

	chunkSize := 20
	numChunks := (len(airports) + chunkSize - 1) / chunkSize
	resultCh := make(chan result, numChunks) // Safer than list because this variable will be used at the same time

	processChunk := func(chunk []domain.Airport) {
		updated, errors := 0, 0

		// Build FAA list and try batch fetch with retry
		faaList := make([]string, 0, len(chunk))
		for _, a := range chunk {
			faaList = append(faaList, a.Faa)
		}

		var fetchedAirports []domain.Airport
		var batchErr error

		// Max attempt per batch is 2, felt like giving them a chance haha
		for attempt := range 2 {
			fetchedAirports, batchErr = s.FetchAirportsFromAviationAPI(faaList)
			if batchErr == nil {
				break
			}
			if attempt == 0 {
				log.Printf("WARN: Batch fetch failed, retrying...")
				time.Sleep(1 * time.Second)
			}
		}

		// Fallback to individual fetches if batch fails
		if batchErr != nil {
			log.Printf("ERROR: Batch fetch failed, using individual fetches: %v", batchErr)
			for _, faa := range faaList {
				airport, err := s.SyncAirportByFAA(faa)
				if err != nil {
					errors++
					log.Printf("ERROR: Failed to sync %s: %v", faa, err)
				} else {
					updated++
					log.Printf("INFO: Synced %s (%s) in %s: %s", airport.Faa, airport.FacilityName, airport.City, airport.Weather)
				}
				time.Sleep(200 * time.Millisecond)
			}
			resultCh <- result{updated, errors}
			return
		}

		// Process the weather & update db
		for i := range fetchedAirports {
			weatherText, err := s.FetchWeatherFromWeatherAPI(fetchedAirports[i].City)
			if err != nil {
				errors++
				log.Printf("ERROR: Failed to fetch weather for %s: %v", fetchedAirports[i].City, err)
				continue
			}
			fetchedAirports[i].Weather = weatherText

			if err := s.repo.UpdateAirport(&fetchedAirports[i]); err != nil {
				errors++
				log.Printf("ERROR: Failed to update %s: %v", fetchedAirports[i].Faa, err)
				continue
			}

			updated++
			log.Printf("INFO: Synced %s (%s) in %s: %s", fetchedAirports[i].Faa, fetchedAirports[i].FacilityName, fetchedAirports[i].City, fetchedAirports[i].Weather)
			time.Sleep(200 * time.Millisecond)
		}

		resultCh <- result{updated, errors}
	}

	// Launch goroutines for each chunk
	for i := 0; i < len(airports); i += chunkSize {
		end := min(i+chunkSize, len(airports))
		go processChunk(airports[i:end])
	}

	// Collect results
	totalUpdated, totalErrors := 0, 0
	for range numChunks {
		res := <-resultCh
		totalUpdated += res.updated
		totalErrors += res.errors
	}

	if totalErrors > 0 && totalUpdated > 0 {
		log.Printf("INFO: Partial sync (%d/%d)", totalUpdated, len(airports))
		return totalUpdated, nil
	}

	if totalErrors > 0 && totalUpdated <= 0 {
		return 0, fmt.Errorf("failed to sync all airports")
	}

	return totalUpdated, nil
}

// Internal helper
func (s *Service) fetchAirportFromAviationAPI(faa string) (*domain.Airport, error) {
	apiURL := fmt.Sprintf("https://api.aviationapi.com/v1/airports?apt=%s", url.QueryEscape(faa))
	resp, err := s.httpClient.Get(apiURL)
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

// Internal Helper
func (s *Service) fetchAirportsFromAviationAPI(faaList []string) ([]domain.Airport, error) {
	if len(faaList) == 0 {
		return nil, fmt.Errorf("empty FAA list")
	}

	aptParam := strings.Join(faaList, ",")
	apiURL := fmt.Sprintf("https://api.aviationapi.com/v1/airports?apt=%s", url.QueryEscape(aptParam))

	resp, err := s.httpClient.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("batch request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("batch API returned %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read batch response: %w", err)
	}

	var resultMap map[string][]domain.Airport
	if err := json.Unmarshal(body, &resultMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal batch: %w", err)
	}

	// Flatten the map into a single array
	airports := []domain.Airport{}
	for _, airportList := range resultMap {
		if len(airportList) > 0 {
			airports = append(airports, airportList[0]) // Take first airport from each list
		}
	}

	return airports, nil
}

// Internal helper
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
