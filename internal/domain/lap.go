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
}

// LeaderboardEntry is a projected view: best lap per player/track/car.
type LeaderboardEntry struct {
	Rank        int       `json:"rank"`
	PlayerName  string    `json:"player_name"`
	Country     string    `json:"country"`
	CarName     string    `json:"car_name"`
	CarInternal string    `json:"car_internal"`
	LapTimeMs   uint32    `json:"lap_time_ms"`
	SectorsMs   []uint32  `json:"sectors_ms,omitempty"`
	Grip        float32   `json:"grip,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
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
