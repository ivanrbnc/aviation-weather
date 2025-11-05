package handler

import (
	"encoding/json"
	"net/http"

	"aviation-weather/internal/service"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc *service.Service
}

func NewHandler(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Router() *chi.Mux {
	r := chi.NewRouter()

	// Routes
	r.Get("/health", h.healthCheck)
	r.Get("/airports/{faa}", h.getAirportWithWeather)

	return r
}

// healthCheck: Simple health endpoint (from server stub).
func (h *Handler) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "message": "Aviation Weather API is running"})
}

// getAirportWithWeather: Fetches and returns enriched airport data.
func (h *Handler) getAirportWithWeather(w http.ResponseWriter, r *http.Request) {
	faa := chi.URLParam(r, "faa")
	if faa == "" {
		http.Error(w, `{"error": "missing faa parameter"}`, http.StatusBadRequest)
		return
	}

	airport, err := h.svc.GetAirportWithWeather(faa)
	if err != nil {
		if airport == nil {
			http.Error(w, `{"error": "airport not found"}`, http.StatusNotFound)
		} else {
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(airport)
}
