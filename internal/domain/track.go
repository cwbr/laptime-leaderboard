package domain

// Track represents a track/circuit within a game.
type Track struct {
	ID         int64  `json:"id"`
	GameID     int64  `json:"game_id"`
	InternalID string `json:"internal_id"` // e.g. "ks_vallelunga"
	Config     string `json:"config"`      // e.g. "extended_circuit"
	Name       string `json:"name"`        // display name
}
