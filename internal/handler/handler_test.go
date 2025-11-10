package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"aviation-weather/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Fake service that won't call any API or functionalities
type MockService struct {
	mock.Mock
}

func (m *MockService) CreateAirport(a *domain.Airport) error {
	args := m.Called(a)
	return args.Error(0)
}

func (m *MockService) UpdateAirport(a *domain.Airport) error {
	args := m.Called(a)
	return args.Error(0)
}

func (m *MockService) DeleteAirportByFAA(faa string) error {
	args := m.Called(faa)
	return args.Error(0)
}

func (m *MockService) GetAirportByFAA(faa string) (*domain.Airport, error) {
	args := m.Called(faa)
	return args.Get(0).(*domain.Airport), args.Error(1)
}

func (m *MockService) GetAllAirports() ([]domain.Airport, error) {
	args := m.Called()
	return args.Get(0).([]domain.Airport), args.Error(1)
}

func (m *MockService) SyncAirportByFAA(faa string) (*domain.Airport, error) {
	args := m.Called(faa)
	return args.Get(0).(*domain.Airport), args.Error(1)
}

func (m *MockService) SyncAllAirports() (int, error) {
	args := m.Called()
	return args.Int(0), args.Error(1)
}

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
	h := NewHandler(&MockService{})
	r := h.Router()

	req := httptest.NewRequest("GET", "/health", nil) // Fake request
	rec := httptest.NewRecorder()                     // Fake response writer, no connection to web server

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "HTTP status code should be 200")
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"), "Header should be JSON")
	assert.JSONEq(t, `{"status":"OK","message":"Aviation Weather API is Running","data":null}`, rec.Body.String(), "JSON body should match")
}

func TestGetAllAirports(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockService)
		expectedCode   int
		expectedJSON   string
		expectedStatus string
		expectedMsg    string
	}{
		{
			name: "success",
			setupMock: func(m *MockService) {
				m.On("GetAllAirports").Return([]domain.Airport{sampleAirport}, nil)
			},
			expectedCode:   http.StatusOK,
			expectedJSON:   `{"status":"OK","message":"Airports are Fetched","data":[{"site_number":"12345","facility_name":"Test Airport","faa_ident":"TST","icao_ident":"KTST","state":"CA","state_full":"California","county":"Test County","city":"Test City","ownership":"Public","use":"Public Use","manager":"Test Manager","manager_phone":"123-456-7890","latitude":"34.0522","longitude":"-118.2437","status":"Open","weather":"Clear"}]}`, // Note: JSONEq for fuzzy match
			expectedStatus: "OK",
			expectedMsg:    "Airports are Fetched",
		},
		{
			name: "service error",
			setupMock: func(m *MockService) {
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
			mockSvc := &MockService{}
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
		setupMock    func(*MockService)
		expectedCode int
		expectedJSON string
	}{
		{
			name: "success",
			faa:  "TST",
			setupMock: func(m *MockService) {
				m.On("GetAirportByFAA", "TST").Return(&sampleAirport, nil)
			},
			expectedCode: http.StatusOK,
			expectedJSON: `{"status":"OK","message":"Airport is Fetched","data":{"site_number":"12345","facility_name":"Test Airport","faa_ident":"TST","icao_ident":"KTST","state":"CA","state_full":"California","county":"Test County","city":"Test City","ownership":"Public","use":"Public Use","manager":"Test Manager","manager_phone":"123-456-7890","latitude":"34.0522","longitude":"-118.2437","status":"Open","weather":"Clear"}}`,
		},
		{
			name: "missing faa",
			faa:  "",
			setupMock: func(m *MockService) {
				// No call expected
			},
			expectedCode: http.StatusBadRequest,
			expectedJSON: `{"status":"Bad Request","message":"Missing FAA Parameter","data":null}`,
		},
		{
			name: "not found",
			faa:  "NF",
			setupMock: func(m *MockService) {
				m.On("GetAirportByFAA", "NF").Return((*domain.Airport)(nil), nil)
			},
			expectedCode: http.StatusNotFound,
			expectedJSON: `{"status":"Error","message":"Airport Not Found","data":null}`,
		},
		{
			name: "service error",
			faa:  "ERR",
			setupMock: func(m *MockService) {
				m.On("GetAirportByFAA", "ERR").Return((*domain.Airport)(nil), assert.AnError)
			},
			expectedCode: http.StatusInternalServerError,
			expectedJSON: `{"status":"Error","message":"Service Error","data":null}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &MockService{}
			tt.setupMock(mockSvc)
			h := NewHandler(mockSvc)
			r := h.Router()

			urlPath := "/airport/" + tt.faa
			req := httptest.NewRequest("GET", urlPath, nil)
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
		setupMock    func(*MockService)
		expectedCode int
		expectedJSON string
	}{
		{
			name: "success",
			body: []byte(sampleAirportJSON),
			setupMock: func(m *MockService) {
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
			setupMock: func(m *MockService) {
				// No call expected
			},
			expectedCode: http.StatusBadRequest,
			expectedJSON: `{"status":"Bad Request","message":"Invalid JSON","data":null}`,
		},
		{
			name: "service error",
			body: []byte(sampleAirportJSON),
			setupMock: func(m *MockService) {
				m.On("CreateAirport", mock.MatchedBy(func(a *domain.Airport) bool {
					return a.Faa == "TST"
				})).Return(assert.AnError)
			},
			expectedCode: http.StatusInternalServerError,
			expectedJSON: `{"status":"Error","message":"Airport Not Found","data":null}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &MockService{}
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
		setupMock    func(*MockService)
		expectedCode int
		expectedJSON string
	}{
		{
			name: "success",
			body: []byte(sampleAirportJSON),
			setupMock: func(m *MockService) {
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
			setupMock: func(m *MockService) {
				// No call expected
			},
			expectedCode: http.StatusBadRequest,
			expectedJSON: `{"status":"Bad Request","message":"Invalid JSON","data":null}`,
		},
		{
			name: "service error",
			body: []byte(sampleAirportJSON),
			setupMock: func(m *MockService) {
				m.On("UpdateAirport", mock.MatchedBy(func(a *domain.Airport) bool {
					return a.Faa == "TST"
				})).Return(assert.AnError)
			},
			expectedCode: http.StatusInternalServerError,
			expectedJSON: `{"status":"Error","message":"Airport Not Found","data":null}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &MockService{}
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
		setupMock    func(*MockService)
		expectedCode int
		expectedJSON string
	}{
		{
			name: "success",
			faa:  "TST",
			setupMock: func(m *MockService) {
				m.On("DeleteAirportByFAA", "TST").Return(nil)
			},
			expectedCode: http.StatusOK,
			expectedJSON: `{"status":"OK","message":"Airport is Deleted","data":"TST"}`,
		},
		{
			name: "missing faa",
			faa:  "",
			setupMock: func(m *MockService) {
				// No call expected
			},
			expectedCode: http.StatusBadRequest,
			expectedJSON: `{"status":"Bad Request","message":"Missing FAA Parameter","data":null}`,
		},
		{
			name: "service error",
			faa:  "ERR",
			setupMock: func(m *MockService) {
				m.On("DeleteAirportByFAA", "ERR").Return(assert.AnError)
			},
			expectedCode: http.StatusNotFound,
			expectedJSON: `{"status":"Error","message":"Airport Not Found","data":null}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &MockService{}
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
		setupMock    func(*MockService)
		expectedCode int
		expectedJSON string
	}{
		{
			name: "success",
			faa:  "TST",
			setupMock: func(m *MockService) {
				m.On("SyncAirportByFAA", "TST").Return(&sampleAirport, nil)
			},
			expectedCode: http.StatusOK,
			expectedJSON: `{"status":"OK","message":"Airport is Synced","data":{"site_number":"12345","facility_name":"Test Airport","faa_ident":"TST","icao_ident":"KTST","state":"CA","state_full":"California","county":"Test County","city":"Test City","ownership":"Public","use":"Public Use","manager":"Test Manager","manager_phone":"123-456-7890","latitude":"34.0522","longitude":"-118.2437","status":"Open","weather":"Clear"}}`,
		},
		{
			name: "missing faa",
			faa:  "",
			setupMock: func(m *MockService) {
				// No call expected
			},
			expectedCode: http.StatusBadRequest,
			expectedJSON: `{"status":"Bad Request","message":"Missing FAA Parameter","data":null}`,
		},
		{
			name: "not found",
			faa:  "NF",
			setupMock: func(m *MockService) {
				m.On("SyncAirportByFAA", "NF").Return((*domain.Airport)(nil), nil)
			},
			expectedCode: http.StatusNotFound,
			expectedJSON: `{"status":"Error","message":"Airport Not Found","data":null}`,
		},
		{
			name: "service error",
			faa:  "ERR",
			setupMock: func(m *MockService) {
				m.On("SyncAirportByFAA", "ERR").Return((*domain.Airport)(nil), assert.AnError)
			},
			expectedCode: http.StatusInternalServerError,
			expectedJSON: `{"status":"Error","message":"Service Error","data":null}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &MockService{}
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
		setupMock    func(*MockService)
		expectedCode int
		expectedJSON string
	}{
		{
			name: "success",
			setupMock: func(m *MockService) {
				m.On("SyncAllAirports").Return(1, nil)
			},
			expectedCode: http.StatusOK,
			expectedJSON: `{"status":"OK","message":"1 Airports are Synced","data":null}`,
		},
		{
			name: "service error",
			setupMock: func(m *MockService) {
				m.On("SyncAllAirports").Return(0, assert.AnError)
			},
			expectedCode: http.StatusInternalServerError,
			expectedJSON: `{"status":"Error","message":"Service Error","data":null}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &MockService{}
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
