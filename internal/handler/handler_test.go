package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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

var sampleAirportJSON = `{"site_number":"12345","facility_name":"Test Airport","faa_ident":"TST","icao_ident":"KTST","state":"CA","state_full":"California","county":"Test County","city":"Test City","ownership":"Public","use":"Public Use","manager":"Test Manager","manager_phone":"123-456-7890","latitude":"34.0522","longitude":"-118.2437","status":"Open","weather":"Clear"}`

func TestHealthCheck(t *testing.T) {
	h := NewHandler(&mocks.ServiceMock{})
	r := h.Router()

	req := httptest.NewRequest("GET", "/health", nil) // Fake request
	rec := httptest.NewRecorder()                     // Fake response writer, no connection to web server

	r.ServeHTTP(rec, req) // Simulation HTTP Request in memory. Run the handler as if a real client made this HTTP request, and store the result in rec

	assert.Equal(t, http.StatusOK, rec.Code, "HTTP status code should be 200")
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"), "Header should be JSON")
	assert.JSONEq(t, `{"status":"OK","message":"Aviation Weather API is Running","data":null}`, rec.Body.String(), "JSON body should match")
}

func TestGetAllAirports(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*mocks.ServiceMock)
		expectedCode   int
		expectedJSON   string
		expectedStatus string
		expectedMsg    string
	}{
		// Get all success
		{
			name: "success",
			setupMock: func(m *mocks.ServiceMock) {
				m.On("GetAllAirports").Return([]domain.Airport{sampleAirport}, nil)
			},
			expectedCode:   http.StatusOK,
			expectedJSON:   `{"status":"OK","message":"Airports are Fetched","data":[{"site_number":"12345","facility_name":"Test Airport","faa_ident":"TST","icao_ident":"KTST","state":"CA","state_full":"California","county":"Test County","city":"Test City","ownership":"Public","use":"Public Use","manager":"Test Manager","manager_phone":"123-456-7890","latitude":"34.0522","longitude":"-118.2437","status":"Open","weather":"Clear"}]}`, // Note: JSONEq for fuzzy match
			expectedStatus: "OK",
			expectedMsg:    "Airports are Fetched",
		},
		// Service error
		{
			name: "service error",
			setupMock: func(m *mocks.ServiceMock) {
				m.On("GetAllAirports").Return([]domain.Airport{}, assert.AnError)
			},
			expectedCode:   http.StatusInternalServerError,
			expectedJSON:   `{"status":"Error","message":"Service Error","data":null}`,
			expectedStatus: "Error",
			expectedMsg:    "Service Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mocks.ServiceMock{} // Use the service mock to fake the return
			tt.setupMock(mockSvc)
			h := NewHandler(mockSvc)
			r := h.Router()

			req := httptest.NewRequest("GET", "/airports", nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code, "HTTP status code should match")
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"), "Header should be JSON")
			assert.JSONEq(t, tt.expectedJSON, rec.Body.String(), "JSON body should match")
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestGetAirport(t *testing.T) {
	tests := []struct {
		name         string
		faa          string
		setupMock    func(*mocks.ServiceMock)
		expectedCode int
		expectedJSON string
	}{
		{
			name: "success",
			faa:  "TST",
			setupMock: func(m *mocks.ServiceMock) {
				m.On("GetAirportByFAA", "TST").Return(&sampleAirport, nil)
			},
			expectedCode: http.StatusOK,
			expectedJSON: `{"status":"OK","message":"Airport is Fetched","data":{"site_number":"12345","facility_name":"Test Airport","faa_ident":"TST","icao_ident":"KTST","state":"CA","state_full":"California","county":"Test County","city":"Test City","ownership":"Public","use":"Public Use","manager":"Test Manager","manager_phone":"123-456-7890","latitude":"34.0522","longitude":"-118.2437","status":"Open","weather":"Clear"}}`,
		},
		{
			name: "missing faa",
			faa:  "",
			setupMock: func(m *mocks.ServiceMock) {
				// No call expected
			},
			expectedCode: http.StatusBadRequest,
			expectedJSON: `{"status":"Bad Request","message":"Missing FAA Parameter","data":null}`,
		},
		{
			name: "not found",
			faa:  "NF",
			setupMock: func(m *mocks.ServiceMock) {
				m.On("GetAirportByFAA", "NF").Return((*domain.Airport)(nil), assert.AnError)
			},
			expectedCode: http.StatusNotFound,
			expectedJSON: `{"status":"Error","message":"Airport Not Found","data":null}`,
		},
		{
			name: "service error",
			faa:  "ERR",
			setupMock: func(m *mocks.ServiceMock) {
				m.On("GetAirportByFAA", "ERR").Return((*domain.Airport)(nil), assert.AnError)
			},
			expectedCode: http.StatusNotFound,
			expectedJSON: `{"status":"Error","message":"Airport Not Found","data":null}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mocks.ServiceMock{}
			tt.setupMock(mockSvc)
			h := NewHandler(mockSvc)
			r := h.Router()

			urlPath := "/airport/" + tt.faa
			req := httptest.NewRequest(http.MethodGet, urlPath, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code, "HTTP status code should match")
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"), "Header should be JSON")
			assert.JSONEq(t, tt.expectedJSON, rec.Body.String(), "JSON body should match")
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestCreateAirport(t *testing.T) {
	tests := []struct {
		name         string
		body         []byte
		setupMock    func(*mocks.ServiceMock)
		expectedCode int
		expectedJSON string
	}{
		{
			name: "success",
			body: []byte(sampleAirportJSON),
			setupMock: func(m *mocks.ServiceMock) {
				m.On("CreateAirport", mock.MatchedBy(func(a *domain.Airport) bool {
					return a.Faa == "TST"
				})).Return(nil)
			},
			expectedCode: http.StatusOK,
			expectedJSON: `{"status":"OK","message":"Airport is Created","data":{"site_number":"12345","facility_name":"Test Airport","faa_ident":"TST","icao_ident":"KTST","state":"CA","state_full":"California","county":"Test County","city":"Test City","ownership":"Public","use":"Public Use","manager":"Test Manager","manager_phone":"123-456-7890","latitude":"34.0522","longitude":"-118.2437","status":"Open","weather":"Clear"}}`,
		},
		{
			name: "invalid json",
			body: []byte(`{invalid}`),
			setupMock: func(m *mocks.ServiceMock) {
				// No call expected
			},
			expectedCode: http.StatusBadRequest,
			expectedJSON: `{"status":"Bad Request","message":"Invalid JSON","data":null}`,
		},
		// JSON has empty faa
		{
			name: "empty faa",
			body: func() []byte {
				airport := sampleAirport // Copy sampleAirport
				airport.Faa = ""         // Set Faa to empty
				data, err := json.Marshal(airport)
				if err != nil {
					t.Fatalf("Failed to marshal JSON: %v", err)
				}
				return data
			}(),
			setupMock: func(m *mocks.ServiceMock) {
				// No call expected
			},
			expectedCode: http.StatusBadRequest,
			expectedJSON: `{"status":"Bad Request","message":"Missing FAA Value","data":null}`,
		},
		{
			name: "service error",
			body: []byte(sampleAirportJSON),
			setupMock: func(m *mocks.ServiceMock) {
				m.On("CreateAirport", mock.MatchedBy(func(a *domain.Airport) bool {
					return a.Faa == "TST"
				})).Return(assert.AnError)
			},
			expectedCode: http.StatusInternalServerError,
			expectedJSON: `{"status":"Error","message":"Duplicate Airport","data":null}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mocks.ServiceMock{}
			tt.setupMock(mockSvc)
			h := NewHandler(mockSvc)
			r := h.Router()

			req := httptest.NewRequest("POST", "/airport", bytes.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code, "HTTP status code should match")
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"), "Header should be JSON")
			assert.JSONEq(t, tt.expectedJSON, rec.Body.String(), "JSON body should match")
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestUpdateAirport(t *testing.T) {
	tests := []struct {
		name         string
		body         []byte
		setupMock    func(*mocks.ServiceMock)
		expectedCode int
		expectedJSON string
	}{
		{
			name: "success",
			body: []byte(sampleAirportJSON),
			setupMock: func(m *mocks.ServiceMock) {
				m.On("UpdateAirport", mock.MatchedBy(func(a *domain.Airport) bool {
					return a.Faa == "TST"
				})).Return(nil)
			},
			expectedCode: http.StatusOK,
			expectedJSON: `{"status":"OK","message":"Airport is Updated","data":{"site_number":"12345","facility_name":"Test Airport","faa_ident":"TST","icao_ident":"KTST","state":"CA","state_full":"California","county":"Test County","city":"Test City","ownership":"Public","use":"Public Use","manager":"Test Manager","manager_phone":"123-456-7890","latitude":"34.0522","longitude":"-118.2437","status":"Open","weather":"Clear"}}`,
		},
		{
			name: "invalid json",
			body: []byte(`{invalid}`),
			setupMock: func(m *mocks.ServiceMock) {
				// No call expected
			},
			expectedCode: http.StatusBadRequest,
			expectedJSON: `{"status":"Bad Request","message":"Invalid JSON","data":null}`,
		},
		{
			name: "service error",
			body: []byte(sampleAirportJSON),
			setupMock: func(m *mocks.ServiceMock) {
				m.On("UpdateAirport", mock.MatchedBy(func(a *domain.Airport) bool {
					return a.Faa == "TST"
				})).Return(assert.AnError)
			},
			expectedCode: http.StatusNotFound,
			expectedJSON: `{"status":"Error","message":"Airport Not Found","data":null}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mocks.ServiceMock{}
			tt.setupMock(mockSvc)
			h := NewHandler(mockSvc)
			r := h.Router()

			req := httptest.NewRequest("PUT", "/airport", bytes.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code, "HTTP status code should match")
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"), "Header should be JSON")
			assert.JSONEq(t, tt.expectedJSON, rec.Body.String(), "JSON body should match")
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestDeleteAirportByFAA(t *testing.T) {
	tests := []struct {
		name         string
		faa          string
		setupMock    func(*mocks.ServiceMock)
		expectedCode int
		expectedJSON string
	}{
		{
			name: "success",
			faa:  "TST",
			setupMock: func(m *mocks.ServiceMock) {
				m.On("DeleteAirportByFAA", "TST").Return(nil)
			},
			expectedCode: http.StatusOK,
			expectedJSON: `{"status":"OK","message":"Airport is Deleted","data":"TST"}`,
		},
		{
			name: "missing faa",
			faa:  "",
			setupMock: func(m *mocks.ServiceMock) {
				// No call expected
			},
			expectedCode: http.StatusBadRequest,
			expectedJSON: `{"status":"Bad Request","message":"Missing FAA Parameter","data":null}`,
		},
		{
			name: "service error",
			faa:  "ERR",
			setupMock: func(m *mocks.ServiceMock) {
				m.On("DeleteAirportByFAA", "ERR").Return(assert.AnError)
			},
			expectedCode: http.StatusNotFound,
			expectedJSON: `{"status":"Error","message":"Airport Not Found","data":null}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mocks.ServiceMock{}
			tt.setupMock(mockSvc)
			h := NewHandler(mockSvc)
			r := h.Router()

			urlPath := "/airports/" + tt.faa
			req := httptest.NewRequest("DELETE", urlPath, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code, "HTTP status code should match")
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"), "Header should be JSON")
			assert.JSONEq(t, tt.expectedJSON, rec.Body.String(), "JSON body should match")
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestSyncAirportByFAA(t *testing.T) {
	tests := []struct {
		name         string
		faa          string
		setupMock    func(*mocks.ServiceMock)
		expectedCode int
		expectedJSON string
	}{
		{
			name: "success",
			faa:  "TST",
			setupMock: func(m *mocks.ServiceMock) {
				m.On("SyncAirportByFAA", "TST").Return(&sampleAirport, nil)
			},
			expectedCode: http.StatusOK,
			expectedJSON: `{"status":"OK","message":"Airport is Synced","data":{"site_number":"12345","facility_name":"Test Airport","faa_ident":"TST","icao_ident":"KTST","state":"CA","state_full":"California","county":"Test County","city":"Test City","ownership":"Public","use":"Public Use","manager":"Test Manager","manager_phone":"123-456-7890","latitude":"34.0522","longitude":"-118.2437","status":"Open","weather":"Clear"}}`,
		},
		{
			name: "missing faa",
			faa:  "",
			setupMock: func(m *mocks.ServiceMock) {
				// No call expected
			},
			expectedCode: http.StatusBadRequest,
			expectedJSON: `{"status":"Bad Request","message":"Missing FAA Parameter","data":null}`,
		},
		{
			name: "not found",
			faa:  "NF",
			setupMock: func(m *mocks.ServiceMock) {
				m.On("SyncAirportByFAA", "NF").Return((*domain.Airport)(nil), assert.AnError)
			},
			expectedCode: http.StatusNotFound,
			expectedJSON: `{"status":"Error","message":"Airport Not Found","data":null}`,
		},
		{
			name: "service error",
			faa:  "ERR",
			setupMock: func(m *mocks.ServiceMock) {
				m.On("SyncAirportByFAA", "ERR").Return((*domain.Airport)(nil), assert.AnError)
			},
			expectedCode: http.StatusNotFound,
			expectedJSON: `{"status":"Error","message":"Airport Not Found","data":null}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mocks.ServiceMock{}
			tt.setupMock(mockSvc)
			h := NewHandler(mockSvc)
			r := h.Router()

			urlPath := "/sync/" + tt.faa
			req := httptest.NewRequest("POST", urlPath, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code, "HTTP status code should match")
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"), "Header should be JSON")
			assert.JSONEq(t, tt.expectedJSON, rec.Body.String(), "JSON body should match")
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestSyncAllAirports(t *testing.T) {
	tests := []struct {
		name         string
		setupMock    func(*mocks.ServiceMock)
		expectedCode int
		expectedJSON string
	}{
		{
			name: "success",
			setupMock: func(m *mocks.ServiceMock) {
				m.On("SyncAllAirports").Return(1, nil)
			},
			expectedCode: http.StatusOK,
			expectedJSON: `{"status":"OK","message":"1 Airports are Synced","data":null}`,
		},
		{
			name: "no airports updated",
			setupMock: func(m *mocks.ServiceMock) {
				m.On("SyncAllAirports").Return(0, nil)
			},
			expectedCode: http.StatusOK,
			expectedJSON: `{"status":"OK","message":"0 Airports are Synced","data":null}`,
		},
		{
			name: "no airports to sync with error",
			setupMock: func(m *mocks.ServiceMock) {
				m.On("SyncAllAirports").Return(0, assert.AnError)
			},
			expectedCode: http.StatusOK,
			expectedJSON: `{"status":"Error","message":"No Airport to Sync","data":null}`,
		},
		{
			name: "service error with updates",
			setupMock: func(m *mocks.ServiceMock) {
				m.On("SyncAllAirports").Return(1, assert.AnError)
			},
			expectedCode: http.StatusInternalServerError,
			expectedJSON: `{"status":"Error","message":"Service Error","data":null}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mocks.ServiceMock{}
			tt.setupMock(mockSvc)
			h := NewHandler(mockSvc)
			r := h.Router()

			req := httptest.NewRequest("POST", "/sync", nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code, "HTTP status code should match")
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"), "Header should be JSON")
			assert.JSONEq(t, tt.expectedJSON, rec.Body.String(), "JSON body should match")
			mockSvc.AssertExpectations(t)
		})
	}
}
