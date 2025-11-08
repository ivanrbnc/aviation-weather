package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"aviation-weather/internal/domain"
	"aviation-weather/internal/service"
	"aviation-weather/internal/utils"

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
	r.Get("/airports", h.getAllAirports)
	r.Get("/airport/{faa}", h.getAirport)
	r.Post("/airport", h.createAirport)
	r.Put("/airport", h.updateAirport)
	r.Post("/sync", h.syncAllAirports)
	r.Post("/sync/{faa}", h.syncAirportByFAA)
	r.Delete("/airports/{faa}", h.deleteAirportByFAA)

	return r
}

// healthCheck: Simple health endpoint.
func (h *Handler) healthCheck(w http.ResponseWriter, r *http.Request) {
	utils.EncodeResponseToUser(w, "OK", "Aviation Weather API is Running", "No data passed")
}

func (h *Handler) createAirport(w http.ResponseWriter, r *http.Request) {
	var airport domain.Airport
	if err := json.NewDecoder(r.Body).Decode(&airport); err != nil {
		utils.EncodeResponseToUser(w, "Bad Request", "Invalid JSON", "No data passed")
		return
	}

	if err := h.svc.CreateAirport(&airport); err != nil {
		fmt.Print(err)
		utils.EncodeResponseToUser(w, "Error", "Service Error", "No data passed")
		return
	}

	utils.EncodeResponseToUser(w, "OK", "Airport is Created", airport)
}

func (h *Handler) updateAirport(w http.ResponseWriter, r *http.Request) {
	var airport domain.Airport
	if err := json.NewDecoder(r.Body).Decode(&airport); err != nil {
		utils.EncodeResponseToUser(w, "Bad Request", "Invalid JSON", "No data passed")
		return
	}

	if err := h.svc.UpdateAirport(&airport); err != nil {
		fmt.Print(err)
		utils.EncodeResponseToUser(w, "Error", "Service Error", "No data passed")
		return
	}

	utils.EncodeResponseToUser(w, "OK", "Airport is Updated", airport)
}

func (h *Handler) deleteAirportByFAA(w http.ResponseWriter, r *http.Request) {
	faa := chi.URLParam(r, "faa")
	if faa == "" {
		utils.EncodeResponseToUser(w, "Bad Request", "Missing FAA Parameter", "No data passed")
		return
	}

	err := h.svc.DeleteAirportByFAA(faa)
	if err != nil {
		fmt.Print(err)
		utils.EncodeResponseToUser(w, "Error", "Airport Not Found", "No data passed")
		return
	}

	utils.EncodeResponseToUser(w, "OK", "Airport is Deleted", faa)
}

func (h *Handler) getAirport(w http.ResponseWriter, r *http.Request) {
	faa := chi.URLParam(r, "faa")
	if faa == "" {
		utils.EncodeResponseToUser(w, "Bad Request", "Missing FAA Parameter", "No data passed")
		return
	}

	airport, err := h.svc.GetAirportByFAA(faa)
	if err != nil {
		fmt.Print(err)
		utils.EncodeResponseToUser(w, "Error", "Service Error", "No data passed")
		return
	}

	if airport == nil {
		utils.EncodeResponseToUser(w, "Error", "Airport Not Found", "No data passed")
		return
	}

	utils.EncodeResponseToUser(w, "OK", "Airport is Fetched", airport)
}

func (h *Handler) getAllAirports(w http.ResponseWriter, r *http.Request) {
	airports, err := h.svc.GetAllAirports()
	if err != nil {
		fmt.Print(err)
		utils.EncodeResponseToUser(w, "Error", "Service Error", "No data passed")
		return
	}

	utils.EncodeResponseToUser(w, "OK", "Airports are Fetched", airports)
}

// syncAirportByFAA: Syncs a single airport by FAA (fetches APIs, updates DB).
func (h *Handler) syncAirportByFAA(w http.ResponseWriter, r *http.Request) {
	faa := chi.URLParam(r, "faa")
	if faa == "" {
		utils.EncodeResponseToUser(w, "Bad Request", "Missing FAA Parameter", "No data passed")
		return
	}

	airport, err := h.svc.SyncAirportByFAA(faa)
	if err != nil {
		fmt.Print(err)
		utils.EncodeResponseToUser(w, "Error", "Service Error", "No data passed")
		return
	}

	if airport == nil {
		utils.EncodeResponseToUser(w, "Error", "Airport Not Found", "No data passed")
		return
	}

	utils.EncodeResponseToUser(w, "OK", "Airport is Synced", airport)
}

// syncAllAirports: Bulk updates all airports with real API data.
func (h *Handler) syncAllAirports(w http.ResponseWriter, r *http.Request) {
	updated, err := h.svc.SyncAllAirports()
	if err != nil {
		fmt.Print(err)
		utils.EncodeResponseToUser(w, "Error", "Service Error", "No data passed")
		return
	}

	utils.EncodeResponseToUser(w, "OK", fmt.Sprintf("%d Airports are Synced", updated), "None")
}
