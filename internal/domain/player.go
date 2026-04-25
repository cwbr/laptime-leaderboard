package domain

// Player represents a person across games.
type Player struct {
	ID         int64  `json:"id"`
	Platform   string `json:"platform"`    // e.g. "steam"
	PlatformID string `json:"platform_id"` // e.g. "76561198012345678"
	Name       string `json:"name"`
	Country    string `json:"country"`
}
