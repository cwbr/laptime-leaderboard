package domain

// Game represents a supported game (e.g. "Assetto Corsa", "iRacing").
type Game struct {
	ID   int64  `json:"id"`
	Slug string `json:"slug"` // e.g. "assetto-corsa"
	Name string `json:"name"` // e.g. "Assetto Corsa"
}
