package ingesthttp

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/chris/laptime-leaderboard/internal/domain"
)

// Handler is the HTTP adapter for lap ingestion.
// It receives POST requests from game server plugins and feeds them
// into the core service via the LapIngester port.
type Handler struct {
	ingester domain.LapIngester
	logger   *slog.Logger
}

func NewHandler(ingester domain.LapIngester, logger *slog.Logger) *Handler {
	return &Handler{ingester: ingester, logger: logger}
}

// RegisterRoutes mounts the ingest API routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/laps", h.handleIngestLap)
}

func (h *Handler) handleIngestLap(w http.ResponseWriter, r *http.Request) {
	// Extract API key from Authorization header
	apiKey := extractBearerToken(r)
	if apiKey == "" {
		http.Error(w, `{"error":"missing Authorization header"}`, http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req domain.IngestLapRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.LapTimeMs == 0 {
		http.Error(w, `{"error":"lap_time_ms is required"}`, http.StatusBadRequest)
		return
	}
	if req.PlayerID == "" || req.PlayerPlatform == "" {
		http.Error(w, `{"error":"player_platform and player_id are required"}`, http.StatusBadRequest)
		return
	}
	if req.Track == "" {
		http.Error(w, `{"error":"track is required"}`, http.StatusBadRequest)
		return
	}
	if req.Car == "" {
		http.Error(w, `{"error":"car is required"}`, http.StatusBadRequest)
		return
	}
	if req.GameSlug == "" {
		http.Error(w, `{"error":"game_slug is required"}`, http.StatusBadRequest)
		return
	}

	req.ServerAPIKey = apiKey

	if err := h.ingester.IngestLap(r.Context(), req); err != nil {
		h.logger.Error("lap ingestion failed", "error", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusUnprocessableEntity)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return ""
}
