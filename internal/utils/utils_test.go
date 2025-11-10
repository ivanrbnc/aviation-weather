package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeResponseToUser(t *testing.T) {
	// Test cases
	tests := []struct {
		name         string
		status       string
		message      string
		data         any
		expectedCode int
		expectedJSON string
	}{
		// Test 1: Normal success (default code 200, with data)
		{
			name:         "success with data",
			status:       "OK",
			message:      "Test message",
			data:         "some data",
			expectedCode: http.StatusOK, // 200
			expectedJSON: `{"status":"OK","message":"Test message","data":"some data"}`,
		},
		// Test 2: Success with no data (default code 200, with nil)
		{
			name:         "success no data",
			status:       "OK",
			message:      "No data here",
			data:         nil,
			expectedCode: http.StatusOK, // 200
			expectedJSON: `{"status":"OK","message":"No data here","data":null}`,
		},
		// Test 3: Error with bad request
		{
			name:         "error bad request",
			status:       "Bad Request",
			message:      "Invalid input",
			data:         nil,
			expectedCode: http.StatusBadRequest, // 400
			expectedJSON: `{"status":"Bad Request","message":"Invalid input","data":null}`,
		},
		// Test 4: Internal server error
		{
			name:         "error server",
			status:       "Error",
			message:      "Something broke",
			data:         "error details",
			expectedCode: http.StatusInternalServerError, // 500
			expectedJSON: `{"status":"Error","message":"Something broke","data":"error details"}`,
		},
	}

	for _, tt := range tests {

		// Sub test
		t.Run(tt.name, func(t *testing.T) {
			// Fake response writer, no connection to web server
			rec := httptest.NewRecorder()

			if tt.expectedCode == http.StatusOK {
				EncodeResponseToUser(rec, tt.status, tt.message, tt.data)
			} else {
				EncodeResponseToUser(rec, tt.status, tt.message, tt.data, tt.expectedCode)
			}

			// Checking the status code, header, and body
			assert.Equal(t, tt.expectedCode, rec.Code, "HTTP status code should match")
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"), "Header should be JSON")
			assert.JSONEq(t, tt.expectedJSON, rec.Body.String(), "JSON body should match")
		})
	}
}
