package domain

import (
	"fmt"
	"time"
)

// Lap is the core domain entity — a single recorded lap.
type Lap struct {
	ID           int64     `json:"id"`
	ServerID     int64     `json:"server_id"`
	PlayerID     int64     `json:"player_id"`
	TrackID      int64     `json:"track_id"`
	CarID        int64     `json:"car_id"`
	LapTimeMs    uint32    `json:"lap_time_ms"`
	SectorsMs    []uint32  `json:"sectors_ms,omitempty"`
	Cuts         int       `json:"cuts"`
	Valid        bool      `json:"valid"`
	Grip         float32   `json:"grip,omitempty"`
	SessionType  string    `json:"session_type,omitempty"`
	CreatedAt    time.Time `json:"created_at"`

	// CSP-only fields (nil for non-CSP clients)
	ABSLevel         *int     `json:"abs_level"`
	TCLevel          *int     `json:"tc_level"`
	StabilityControl *float64 `json:"stability_control"`
	AutoShifting     *bool    `json:"auto_shifting"`
	InputMethod      *int     `json:"input_method"`
	TyreCompound     *int     `json:"tyre_compound"`
}

// LeaderboardEntry is a projected view: best lap per player/track/car.
type LeaderboardEntry struct {
	Rank        int       `json:"rank"`
	LapID       int64     `json:"lap_id"`
	PlayerID    int64     `json:"player_id"`
	PlayerName  string    `json:"player_name"`
	Country     string    `json:"country"`
	CarID       int64     `json:"car_id"`
	CarName     string    `json:"car_name"`
	CarInternal string    `json:"car_internal"`
	LapTimeMs   uint32    `json:"lap_time_ms"`
	DeltaMs     int32     `json:"delta_ms"`
	SectorsMs   []uint32  `json:"sectors_ms,omitempty"`
	Grip        float32   `json:"grip,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// LapDetail is a fully resolved lap with all related entity names.
type LapDetail struct {
	Lap
	PlayerName  string `json:"player_name"`
	PlayerCountry string `json:"player_country"`
	TrackName   string `json:"track_name"`
	TrackConfig string `json:"track_config"`
	CarName     string `json:"car_name"`
	CarClass    string `json:"car_class"`
	ServerName  string `json:"server_name"`
	GameName    string `json:"game_name"`
}

// PlayerBestLap is a player's best lap on a specific track.
type PlayerBestLap struct {
	TrackID     int64     `json:"track_id"`
	TrackName   string    `json:"track_name"`
	CarName     string    `json:"car_name"`
	LapTimeMs   uint32    `json:"lap_time_ms"`
	SectorsMs   []uint32  `json:"sectors_ms,omitempty"`
	Grip        float32   `json:"grip,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// PlayerRecentLap is a recent lap with track/car info.
type PlayerRecentLap struct {
	LapID       int64     `json:"lap_id"`
	TrackName   string    `json:"track_name"`
	CarName     string    `json:"car_name"`
	LapTimeMs   uint32    `json:"lap_time_ms"`
	SectorsMs   []uint32  `json:"sectors_ms,omitempty"`
	Valid       bool      `json:"valid"`
	Grip        float32   `json:"grip,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// CarTrackBest is the fastest lap for a car on a specific track.
type CarTrackBest struct {
	TrackID     int64     `json:"track_id"`
	TrackName   string    `json:"track_name"`
	PlayerName  string    `json:"player_name"`
	LapTimeMs   uint32    `json:"lap_time_ms"`
	SectorsMs   []uint32  `json:"sectors_ms,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// PageInfo carries pagination metadata for templates.
type PageInfo struct {
	CurrentPage int
	TotalPages  int
	TotalItems  int
	PageSize    int
}

func NewPageInfo(page, pageSize, totalItems int) PageInfo {
	totalPages := (totalItems + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}
	return PageInfo{
		CurrentPage: page,
		TotalPages:  totalPages,
		TotalItems:  totalItems,
		PageSize:    pageSize,
	}
}

func (p PageInfo) HasPrev() bool { return p.CurrentPage > 1 }
func (p PageInfo) HasNext() bool { return p.CurrentPage < p.TotalPages }
func (p PageInfo) PrevPage() int { return p.CurrentPage - 1 }
func (p PageInfo) NextPage() int { return p.CurrentPage + 1 }

// Pages returns a slice of page numbers for template iteration.
func (p PageInfo) Pages() []int {
	pages := make([]int, p.TotalPages)
	for i := range pages {
		pages[i] = i + 1
	}
	return pages
}

// DeltaFormatted returns a delta string like "+0.432" or "" for the leader.
func DeltaFormatted(deltaMs int32) string {
	if deltaMs == 0 {
		return ""
	}
	sec := float64(deltaMs) / 1000.0
	return fmt.Sprintf("+%.3f", sec)
}

// LapTimeFormatted returns a human-readable lap time string like "1:38.432".
func LapTimeFormatted(ms uint32) string {
	totalSeconds := ms / 1000
	minutes := totalSeconds / 60
	seconds := totalSeconds % 60
	millis := ms % 1000

	if minutes > 0 {
		return fmt.Sprintf("%d:%02d.%03d", minutes, seconds, millis)
	}
	return fmt.Sprintf("%d.%03d", seconds, millis)
}
