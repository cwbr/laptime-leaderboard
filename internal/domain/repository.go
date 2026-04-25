package domain

import "context"

// LapRepository persists and queries lap data.
// This is the primary "driven" port — implemented by SQLite, PostgreSQL, etc.
type LapRepository interface {
	StoreLap(ctx context.Context, lap *Lap) (int64, error)
	GetLeaderboard(ctx context.Context, query LeaderboardQuery) ([]LeaderboardEntry, error)
	GetPlayerLaps(ctx context.Context, playerID, trackID int64, limit int) ([]Lap, error)
}

// LeaderboardQuery defines filters for leaderboard retrieval.
type LeaderboardQuery struct {
	GameID  int64
	TrackID int64
	CarID   int64 // 0 = all cars
	Limit   int
}

// ServerRepository manages server registrations.
type ServerRepository interface {
	GetServerByAPIKey(ctx context.Context, apiKey string) (*Server, error)
	CreateServer(ctx context.Context, server *Server) (int64, error)
	ListServers(ctx context.Context) ([]Server, error)
}

// GameRepository manages game definitions.
type GameRepository interface {
	GetGameBySlug(ctx context.Context, slug string) (*Game, error)
	CreateGame(ctx context.Context, game *Game) (int64, error)
	ListGames(ctx context.Context) ([]Game, error)
}

// TrackRepository manages track definitions.
type TrackRepository interface {
	FindOrCreate(ctx context.Context, gameID int64, internalID, config, name string) (*Track, error)
	GetByID(ctx context.Context, id int64) (*Track, error)
	ListByGame(ctx context.Context, gameID int64) ([]Track, error)
}

// CarRepository manages car definitions.
type CarRepository interface {
	FindOrCreate(ctx context.Context, gameID int64, internalID, name, class string) (*Car, error)
	GetByID(ctx context.Context, id int64) (*Car, error)
	ListByGame(ctx context.Context, gameID int64) ([]Car, error)
}

// PlayerRepository manages player records.
type PlayerRepository interface {
	FindOrCreate(ctx context.Context, platform, platformID, name, country string) (*Player, error)
	GetByID(ctx context.Context, id int64) (*Player, error)
}
