package domain

import "context"

// LapIngester is the "driving" port — the interface that all data sources
// (HTTP adapter, log parser, file watcher, etc.) feed into.
type LapIngester interface {
	IngestLap(ctx context.Context, req IngestLapRequest) error
}

// IngestLapRequest is the normalized input from any data source.
// This is game-agnostic: the adapter translates game-specific data into this.
type IngestLapRequest struct {
	// Server identification (resolved by API key or config)
	ServerAPIKey string `json:"-"`
	ServerID     int64  `json:"-"` // set after auth

	// Game context
	GameSlug string `json:"game_slug"` // e.g. "assetto-corsa"

	// Player
	PlayerPlatform string `json:"player_platform"` // e.g. "steam"
	PlayerID       string `json:"player_id"`        // platform-specific ID
	PlayerName     string `json:"player_name"`
	PlayerCountry  string `json:"player_country"`

	// Track
	Track       string `json:"track"`        // internal ID
	TrackConfig string `json:"track_config"` // layout variant
	TrackName   string `json:"track_name"`   // display name (optional)

	// Car
	Car      string `json:"car"`       // internal ID
	CarName  string `json:"car_name"`  // display name (optional)
	CarClass string `json:"car_class"` // e.g. "GT3" (optional)

	// Lap data
	LapTimeMs   uint32   `json:"lap_time_ms"`
	SectorsMs   []uint32 `json:"sectors_ms,omitempty"`
	Cuts        int      `json:"cuts"`
	Valid       bool     `json:"valid"`
	Grip        float32  `json:"grip,omitempty"`
	SessionType string   `json:"session_type,omitempty"`

	// CSP-only fields (nil for non-CSP clients)
	ABSLevel         *int     `json:"abs_level"`
	TCLevel          *int     `json:"tc_level"`
	StabilityControl *float64 `json:"stability_control"`
	AutoShifting     *bool    `json:"auto_shifting"`
	InputMethod      *int     `json:"input_method"`
	TyreCompound     *int     `json:"tyre_compound"`

	// Extensible metadata (per-game extras)
	Metadata map[string]any `json:"metadata,omitempty"`
}
