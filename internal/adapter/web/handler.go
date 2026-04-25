package web

import (
	"embed"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/chris/laptime-leaderboard/internal/domain"
)

//go:embed templates/*.html
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

// Handler serves the leaderboard website.
type Handler struct {
	service *domain.Service
	logger  *slog.Logger
	tmpl    *template.Template
}

func NewHandler(service *domain.Service, logger *slog.Logger) *Handler {
	funcMap := template.FuncMap{
		"lapTime": domain.LapTimeFormatted,
	}
	tmpl := template.Must(template.New("").Funcs(funcMap).ParseFS(templateFS, "templates/*.html"))

	return &Handler{service: service, logger: logger, tmpl: tmpl}
}

// RegisterRoutes mounts the web UI routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /", h.handleIndex)
	mux.HandleFunc("GET /leaderboard", h.handleLeaderboard)
	mux.Handle("GET /static/", http.FileServerFS(staticFS))
}

func (h *Handler) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	games, err := h.service.GetGames(r.Context())
	if err != nil {
		h.logger.Error("failed to get games", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Games": games,
	}

	h.render(w, "index.html", data)
}

func (h *Handler) handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	gameIDStr := r.URL.Query().Get("game_id")
	trackIDStr := r.URL.Query().Get("track_id")
	carIDStr := r.URL.Query().Get("car_id")

	gameID, _ := strconv.ParseInt(gameIDStr, 10, 64)
	trackID, _ := strconv.ParseInt(trackIDStr, 10, 64)
	carID, _ := strconv.ParseInt(carIDStr, 10, 64)

	// Get available tracks and cars for filtering
	var tracks []domain.Track
	var cars []domain.Car
	if gameID > 0 {
		tracks, _ = h.service.GetTracks(r.Context(), gameID)
		cars, _ = h.service.GetCars(r.Context(), gameID)
	}

	var entries []domain.LeaderboardEntry
	if trackID > 0 {
		var err error
		entries, err = h.service.GetLeaderboard(r.Context(), domain.LeaderboardQuery{
			GameID:  gameID,
			TrackID: trackID,
			CarID:   carID,
			Limit:   100,
		})
		if err != nil {
			h.logger.Error("leaderboard query failed", "error", err)
		}
	}

	games, _ := h.service.GetGames(r.Context())

	data := map[string]any{
		"Games":        games,
		"Tracks":       tracks,
		"Cars":         cars,
		"Entries":      entries,
		"SelectedGame":  gameID,
		"SelectedTrack": trackID,
		"SelectedCar":   carID,
	}

	// If HTMX request, only render the partial
	if r.Header.Get("HX-Request") == "true" {
		h.render(w, "leaderboard_partial.html", data)
		return
	}

	h.render(w, "leaderboard.html", data)
}

func (h *Handler) render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, name, data); err != nil {
		h.logger.Error("template render failed", "template", name, "error", err)
		http.Error(w, "render error", http.StatusInternalServerError)
	}
}
