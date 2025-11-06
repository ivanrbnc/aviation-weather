package utils

import (
	"encoding/json"
	"net/http"
)

func EncodeResponseToUser(w http.ResponseWriter, status string, message string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": status, "message": message})
}
