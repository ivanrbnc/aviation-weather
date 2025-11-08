package utils

import (
	"encoding/json"
	"net/http"

	"aviation-weather/internal/domain"
)

func EncodeResponseToUser(w http.ResponseWriter, status string, message string, data any, code ...int) {
	// Default = 200
	httpCode := http.StatusOK
	if len(code) > 0 {
		httpCode = code[0]
	}

	w.WriteHeader(httpCode)

	w.Header().Set("Content-Type", "application/json")
	resp := domain.ApiResponse{
		Status:  status,
		Message: message,
		Data:    data,
	}
	json.NewEncoder(w).Encode(resp)
}
