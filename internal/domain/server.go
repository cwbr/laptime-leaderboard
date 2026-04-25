package domain

// Server represents a game server that reports lap times.
type Server struct {
	ID     int64  `json:"id"`
	GameID int64  `json:"game_id"`
	Name   string `json:"name"`
	APIKey string `json:"-"` // never exposed in JSON
}
