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

const pageSize = 50

// Handler serves the leaderboard website.
type Handler struct {
	service *domain.Service
	logger  *slog.Logger
	tmpls   map[string]*template.Template
}

func NewHandler(service *domain.Service, logger *slog.Logger) *Handler {
	funcMap := template.FuncMap{
		"lapTime":   domain.LapTimeFormatted,
		"deltaTime": domain.DeltaFormatted,
		"plus1":     func(i int) int { return i + 1 },
		"mul":       func(a float32, b float64) float64 { return float64(a) * b },
	}

	// Parse each page template separately so block definitions don't collide.
	pages := map[string][]string{
		"leaderboard.html":         {"templates/base.html", "templates/leaderboard.html", "templates/leaderboard_partial.html"},
		"leaderboard_partial.html": {"templates/leaderboard_partial.html"},
		"leaderboard_rows.html":    {"templates/leaderboard_rows.html"},
		"driver.html":              {"templates/base.html", "templates/driver.html", "templates/pagination.html"},
		"car.html":                 {"templates/base.html", "templates/car.html", "templates/pagination.html"},
		"lap.html":                 {"templates/base.html", "templates/lap.html"},
	}
	tmpls := make(map[string]*template.Template, len(pages))
	for name, files := range pages {
		tmpls[name] = template.Must(template.New(name).Funcs(funcMap).ParseFS(templateFS, files...))
	}

	return &Handler{service: service, logger: logger, tmpls: tmpls}
}

// RegisterRoutes mounts the web UI routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /leaderboard/more", h.handleLeaderboardMore)
	mux.HandleFunc("GET /driver/{id}", h.handleDriver)
	mux.HandleFunc("GET /car/{id}", h.handleCar)
	mux.HandleFunc("GET /lap/{id}", h.handleLap)
	mux.HandleFunc("GET /", h.handleLeaderboard)
	mux.Handle("GET /static/", http.FileServerFS(staticFS))
}

func (h *Handler) handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "/leaderboard" {
		http.NotFound(w, r)
		return
	}

	gameIDStr := r.URL.Query().Get("game_id")
	trackIDStr := r.URL.Query().Get("track_id")
	carIDStr := r.URL.Query().Get("car_id")

	gameID, _ := strconv.ParseInt(gameIDStr, 10, 64)
	trackID, _ := strconv.ParseInt(trackIDStr, 10, 64)
	carID, _ := strconv.ParseInt(carIDStr, 10, 64)

	games, _ := h.service.GetGames(r.Context())

	// Auto-select first game if none specified
	if gameID == 0 && len(games) > 0 {
		gameID = games[0].ID
	}

	// Get available tracks and cars for filtering
	var tracks []domain.Track
	var cars []domain.Car
	if gameID > 0 {
		tracks, _ = h.service.GetTracks(r.Context(), gameID)
		cars, _ = h.service.GetCars(r.Context(), gameID)
	}

	// Auto-select first track if none specified
	if trackID == 0 && len(tracks) > 0 {
		trackID = tracks[0].ID
	}

	var entries []domain.LeaderboardEntry
	var hasMore bool
	if trackID > 0 {
		var err error
		entries, err = h.service.GetLeaderboard(r.Context(), domain.LeaderboardQuery{
			GameID:  gameID,
			TrackID: trackID,
			CarID:   carID,
			Limit:   pageSize + 1,
		})
		if err != nil {
			h.logger.Error("leaderboard query failed", "error", err)
		}
		if len(entries) > pageSize {
			entries = entries[:pageSize]
			hasMore = true
		}
	}

	data := map[string]any{
		"Games":        games,
		"Tracks":       tracks,
		"Cars":         cars,
		"Entries":      entries,
		"SelectedGame":  gameID,
		"SelectedTrack": trackID,
		"SelectedCar":   carID,
		"HasMore":       hasMore,
		"NextOffset":    pageSize,
	}

	// If HTMX request, only render the partial
	if r.Header.Get("HX-Request") == "true" {
		h.render(w, "leaderboard_partial.html", data)
		return
	}

	h.render(w, "leaderboard.html", data)
}

func (h *Handler) handleLeaderboardMore(w http.ResponseWriter, r *http.Request) {
	trackIDStr := r.URL.Query().Get("track_id")
	carIDStr := r.URL.Query().Get("car_id")
	gameIDStr := r.URL.Query().Get("game_id")
	offsetStr := r.URL.Query().Get("offset")

	trackID, _ := strconv.ParseInt(trackIDStr, 10, 64)
	carID, _ := strconv.ParseInt(carIDStr, 10, 64)
	gameID, _ := strconv.ParseInt(gameIDStr, 10, 64)
	offset, _ := strconv.Atoi(offsetStr)
	loadAll := r.URL.Query().Get("all") == "1"

	if trackID == 0 || offset == 0 {
		w.WriteHeader(http.StatusOK)
		return
	}

	limit := pageSize + 1
	if loadAll {
		limit = 10000
	}

	entries, err := h.service.GetLeaderboard(r.Context(), domain.LeaderboardQuery{
		GameID:  gameID,
		TrackID: trackID,
		CarID:   carID,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		h.logger.Error("leaderboard more query failed", "error", err)
		w.WriteHeader(http.StatusOK)
		return
	}

	var hasMore bool
	if !loadAll && len(entries) > pageSize {
		entries = entries[:pageSize]
		hasMore = true
	}

	data := map[string]any{
		"Entries":       entries,
		"HasMore":       hasMore,
		"NextOffset":    offset + pageSize,
		"SelectedGame":  gameID,
		"SelectedTrack": trackID,
		"SelectedCar":   carID,
	}

	h.render(w, "leaderboard_rows.html", data)
}

func (h *Handler) render(w http.ResponseWriter, name string, data any) {
	t, ok := h.tmpls[name]
	if !ok {
		h.logger.Error("template not found", "template", name)
		http.Error(w, "template not found", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.ExecuteTemplate(w, name, data); err != nil {
		h.logger.Error("template render failed", "template", name, "error", err)
		http.Error(w, "render error", http.StatusInternalServerError)
	}
}

func (h *Handler) handleDriver(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	player, err := h.service.GetPlayer(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	bests, bestsTotal, _ := h.service.GetPlayerBestLaps(r.Context(), id, pageSize, 0)
	recent, recentTotal, _ := h.service.GetPlayerRecentLaps(r.Context(), id, pageSize, 0)

	bestsPage := domain.NewPageInfo(1, pageSize, bestsTotal)
	recentPage := domain.NewPageInfo(1, pageSize, recentTotal)

	// Support HTMX partial page requests
	bestsPageStr := r.URL.Query().Get("bests_page")
	recentPageStr := r.URL.Query().Get("recent_page")

	if p, _ := strconv.Atoi(bestsPageStr); p > 1 {
		offset := (p - 1) * pageSize
		bests, bestsTotal, _ = h.service.GetPlayerBestLaps(r.Context(), id, pageSize, offset)
		bestsPage = domain.NewPageInfo(p, pageSize, bestsTotal)
	}
	if p, _ := strconv.Atoi(recentPageStr); p > 1 {
		offset := (p - 1) * pageSize
		recent, recentTotal, _ = h.service.GetPlayerRecentLaps(r.Context(), id, pageSize, offset)
		recentPage = domain.NewPageInfo(p, pageSize, recentTotal)
	}

	data := map[string]any{
		"Player":     player,
		"Bests":      bests,
		"Recent":     recent,
		"BestsPage":  bestsPage,
		"RecentPage": recentPage,
	}

	h.render(w, "driver.html", data)
}

func (h *Handler) handleCar(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	car, err := h.service.GetCar(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	pageStr := r.URL.Query().Get("page")
	page := 1
	if p, _ := strconv.Atoi(pageStr); p > 1 {
		page = p
	}
	offset := (page - 1) * pageSize

	bests, total, _ := h.service.GetCarTrackBests(r.Context(), id, pageSize, offset)
	pageInfo := domain.NewPageInfo(page, pageSize, total)

	data := map[string]any{
		"Car":      car,
		"Bests":    bests,
		"PageInfo": pageInfo,
	}

	h.render(w, "car.html", data)
}

func (h *Handler) handleLap(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	lap, err := h.service.GetLapDetail(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	data := map[string]any{
		"Lap": lap,
	}

	h.render(w, "lap.html", data)
}
