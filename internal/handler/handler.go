package handler

import (
	"encoding/json"
	"fmt"
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
	r.Get("/airports", h.getAllAirportsWithWeather)
	r.Post("/sync", h.syncAllAirports)
	r.Delete("/airports/{faa}", h.deleteAirportByFAA)

	return r
}

// healthCheck: Simple health endpoint.
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
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err), http.StatusInternalServerError)
		return
	}
	if airport == nil {
		http.Error(w, `{"error": "airport not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(airport)
}

// getAllAirportsWithWeather: Returns all airports enriched with current weather.
func (h *Handler) getAllAirportsWithWeather(w http.ResponseWriter, r *http.Request) {
	airports, err := h.svc.GetAllAirportsWithWeather()
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(airports) // JSON array of airports
}

// syncAllAirports: Bulk updates all airports with real API data
func (h *Handler) syncAllAirports(w http.ResponseWriter, r *http.Request) {
	updated, err := h.svc.SyncAllAirports()
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"updated": updated,
	})
}

// deleteAirportByFAA: Deletes airport by FAA.
func (h *Handler) deleteAirportByFAA(w http.ResponseWriter, r *http.Request) {
	faa := chi.URLParam(r, "faa")
	if faa == "" {
		http.Error(w, `{"error": "missing faa parameter"}`, http.StatusBadRequest)
		return
	}

	err := h.svc.DeleteAirportByFAA(faa)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted", "faa": faa})
}
