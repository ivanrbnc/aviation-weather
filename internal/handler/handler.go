package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"aviation-weather/internal/domain"
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
	r.Get("/airports", h.getAllAirportsWithWeather)
	r.Get("/airport/{faa}", h.getAirportWithWeather)
	r.Post("/airport", h.createAirport)
	r.Put("/airport", h.updateAirport)
	r.Post("/sync", h.syncAllAirports)
	r.Post("/sync/{faa}", h.syncAirportByFAA)
	r.Delete("/airports/{faa}", h.deleteAirportByFAA)

	return r
}

// healthCheck: Simple health endpoint.
func (h *Handler) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "message": "Aviation Weather API is running"})
}

func (h *Handler) createAirport(w http.ResponseWriter, r *http.Request) {
	var airport domain.Airport
	if err := json.NewDecoder(r.Body).Decode(&airport); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "invalid JSON: %s"}`, err), http.StatusBadRequest)
		return
	}

	log.Printf("Creating airport: %+v", airport)

	if err := h.svc.CreateAirport(&airport); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "created", "faa": airport.Faa})
}

func (h *Handler) updateAirport(w http.ResponseWriter, r *http.Request) {
	var airport domain.Airport
	if err := json.NewDecoder(r.Body).Decode(&airport); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "invalid JSON: %s"}`, err), http.StatusBadRequest)
		return
	}

	log.Printf("Updating airport: %+v", airport)

	if err := h.svc.UpdateAirport(&airport); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated", "faa": airport.Faa})
}

// getAirportWithWeather: Fetches and returns enriched airport data.
func (h *Handler) getAirportWithWeather(w http.ResponseWriter, r *http.Request) {
	faa := chi.URLParam(r, "faa")
	if faa == "" {
		http.Error(w, `{"error": "missing faa parameter"}`, http.StatusBadRequest)
		return
	}

	airport, err := h.svc.GetAirportByFAA(faa)
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
	json.NewEncoder(w).Encode(airports)
}

// syncAllAirports: Bulk updates all airports with real API data.
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

// syncAirportByFAA: Syncs a single airport by FAA (fetches APIs, updates DB).
func (h *Handler) syncAirportByFAA(w http.ResponseWriter, r *http.Request) {
	faa := chi.URLParam(r, "faa")
	if faa == "" {
		http.Error(w, `{"error": "missing faa parameter"}`, http.StatusBadRequest)
		return
	}

	airport, err := h.svc.GetAndSaveAirportWithWeather(faa)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err), http.StatusInternalServerError)
		return
	}

	if airport == nil {
		http.Error(w, `{"error": "no data found for sync"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status":  "ok",
		"airport": airport,
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
