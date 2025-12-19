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

	syncQueue    chan syncJob
	syncAllQueue chan syncAllJob
}

type ServiceInterface interface {
	CreateAirport(a *domain.Airport) error
	UpdateAirport(a *domain.Airport) error
	DeleteAirportByFAA(faa string) error
	GetAirportByFAA(faa string) (*domain.Airport, error)
	GetAllAirports() ([]domain.Airport, error)
	SyncAirportByFAA(faa string) (*domain.Airport, error)
	SyncAllAirports() (int, error)

	SyncAirportQueued(faa string) (*domain.Airport, error)
	SyncAllAirportsQueued() (int, error)
}

func NewService(repo repository.RepositoryInterface, cfg *config.Config) ServiceInterface {
	s := &Service{
		repo: repo,
		cfg:  cfg,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		syncQueue:    make(chan syncJob, 100),
		syncAllQueue: make(chan syncAllJob, 100),
	}
	s.FetchAirportFromAviationAPI = s.fetchAirportFromAviationAPI
	s.FetchAirportsFromAviationAPI = s.fetchAirportsFromAviationAPI
	s.FetchWeatherFromWeatherAPI = s.fetchWeatherFromWeatherAPI

	go s.runSyncWorker()
	go s.runSyncAllWorker()

	return s
}

type syncJob struct {
	faa      string
	resultCh chan *domain.Airport
	errCh    chan error
}

func (s *Service) runSyncWorker() {
	for job := range s.syncQueue {
		airport, err := s.SyncAirportByFAA(job.faa)
		if err != nil {
			job.errCh <- err
		} else {
			job.resultCh <- airport
		}
	}
}

func (s *Service) SyncAirportQueued(faa string) (*domain.Airport, error) {
	job := syncJob{
		faa:      faa,
		resultCh: make(chan *domain.Airport, 1),
		errCh:    make(chan error, 1),
	}
	s.syncQueue <- job
	select {
	case airport := <-job.resultCh:
		return airport, nil
	case err := <-job.errCh:
		return nil, err
	}
}

type syncAllJob struct {
	resultCh chan int
	errCh    chan error
}

func (s *Service) runSyncAllWorker() {
	for job := range s.syncAllQueue {
		updated, err := s.SyncAllAirports()
		if err != nil {
			job.errCh <- err
		} else {
			job.resultCh <- updated
		}
	}
}

func (s *Service) SyncAllAirportsQueued() (int, error) {
	job := syncAllJob{
		resultCh: make(chan int, 1),
		errCh:    make(chan error, 1),
	}
	s.syncAllQueue <- job
	select {
	case updated := <-job.resultCh:
		return updated, nil
	case err := <-job.errCh:
		return 0, err
	}
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
	// First check DB
	airport, err := s.repo.GetAirportByFAA(faa)
	if err != nil {
		return nil, fmt.Errorf("failed to get airport for %s: %w", faa, err)
	}
	if airport == nil {
		return nil, fmt.Errorf("no airport found for %s", faa)
	}

	// Determine if static fields are missing
	needsAirportFetch := airport.SiteNumber == "" ||
		airport.FacilityName == "" ||
		airport.Icao == "" ||
		airport.StateCode == "" ||
		airport.StateFull == "" ||
		airport.County == "" ||
		airport.City == "" ||
		airport.OwnershipType == "" ||
		airport.UseType == "" ||
		airport.Manager == "" ||
		airport.ManagerPhone == "" ||
		airport.Latitude == "" ||
		airport.Longitude == "" ||
		airport.AirportStatus == ""

	if needsAirportFetch {
		// Fetch airport details from Aviation API
		airportData, err := s.FetchAirportFromAviationAPI(faa)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch airport for %s: %w", faa, err)
		}
		if airportData == nil {
			return nil, fmt.Errorf("no airport found for %s", faa)
		}
		airport = airportData
	}

	// Always refresh weather
	weatherText, err := s.FetchWeatherFromWeatherAPI(airport.City)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weather for %s: %w", airport.City, err)
	}
	airport.Weather = weatherText

	// Save back to DB
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
	resultCh := make(chan result, numChunks)

	processChunk := func(chunk []domain.Airport) {
		updated, errors := 0, 0

		// Split into two groups: incomplete (need Aviation API) vs complete (only weather)
		var incompleteFAA []string
		var completeAirports []domain.Airport

		for _, a := range chunk {
			needsAirportFetch := a.SiteNumber == "" ||
				a.FacilityName == "" ||
				a.Icao == "" ||
				a.StateCode == "" ||
				a.StateFull == "" ||
				a.County == "" ||
				a.City == "" ||
				a.OwnershipType == "" ||
				a.UseType == "" ||
				a.Manager == "" ||
				a.ManagerPhone == "" ||
				a.Latitude == "" ||
				a.Longitude == "" ||
				a.AirportStatus == ""

			if needsAirportFetch {
				incompleteFAA = append(incompleteFAA, a.Faa)
			} else {
				completeAirports = append(completeAirports, a)
			}
		}

		// Batch fetch for incomplete airports
		var fetchedAirports []domain.Airport
		var batchErr error
		if len(incompleteFAA) > 0 {
			for attempt := 0; attempt < 2; attempt++ {
				fetchedAirports, batchErr = s.FetchAirportsFromAviationAPI(incompleteFAA)
				if batchErr == nil {
					break
				}
				if attempt == 0 {
					log.Printf("WARN: Batch fetch failed, retrying...")
					time.Sleep(1 * time.Second)
				}
			}
			if batchErr != nil {
				log.Printf("ERROR: Batch fetch failed, falling back to individual fetches: %v", batchErr)
				for _, faa := range incompleteFAA {
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
			}
		}

		// Merge fetched airports with complete ones
		allAirports := append(fetchedAirports, completeAirports...)

		// Refresh weather for all
		for i := range allAirports {
			weatherText, err := s.FetchWeatherFromWeatherAPI(allAirports[i].City)
			if err != nil {
				errors++
				log.Printf("ERROR: Failed to fetch weather for %s: %v", allAirports[i].City, err)
				continue
			}
			allAirports[i].Weather = weatherText

			if err := s.repo.UpdateAirport(&allAirports[i]); err != nil {
				errors++
				log.Printf("ERROR: Failed to update %s: %v", allAirports[i].Faa, err)
				continue
			}

			updated++
			log.Printf("INFO: Synced %s (%s) in %s: %s", allAirports[i].Faa, allAirports[i].FacilityName, allAirports[i].City, allAirports[i].Weather)
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
	for i := 0; i < numChunks; i++ {
		res := <-resultCh
		totalUpdated += res.updated
		totalErrors += res.errors
	}

	if totalErrors > 0 && totalUpdated == 0 {
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
