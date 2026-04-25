package domain

// Car represents a vehicle within a game.
type Car struct {
	ID         int64  `json:"id"`
	GameID     int64  `json:"game_id"`
	InternalID string `json:"internal_id"` // e.g. "ks_ferrari_488"
	Name       string `json:"name"`        // display name
	Class      string `json:"class"`       // e.g. "GT3"
}
