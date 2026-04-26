package domain

import "context"

// LapRepository persists and queries lap data.
// This is the primary "driven" port — implemented by SQLite, PostgreSQL, etc.
type LapRepository interface {
	StoreLap(ctx context.Context, lap *Lap) (int64, error)
	GetLeaderboard(ctx context.Context, query LeaderboardQuery) ([]LeaderboardEntry, error)
	GetPlayerLaps(ctx context.Context, playerID, trackID int64, limit int) ([]Lap, error)
	GetLapDetail(ctx context.Context, lapID int64) (*LapDetail, error)
	GetPlayerBestLaps(ctx context.Context, playerID int64, limit, offset int) ([]PlayerBestLap, int, error)
	GetPlayerRecentLaps(ctx context.Context, playerID int64, limit, offset int) ([]PlayerRecentLap, int, error)
	GetCarTrackBests(ctx context.Context, carID int64, limit, offset int) ([]CarTrackBest, int, error)
	DeleteLap(ctx context.Context, lapID int64) error
	DeletePlayerLaps(ctx context.Context, playerID int64) (int64, error)
}

// LeaderboardQuery defines filters for leaderboard retrieval.
type LeaderboardQuery struct {
	GameID   int64
	TrackID  int64
	CarID    int64 // 0 = all cars
	ServerID int64 // 0 = all servers
	Limit    int
	Offset   int
}

// ServerRepository manages server registrations.
type ServerRepository interface {
	GetServerByAPIKey(ctx context.Context, apiKey string) (*Server, error)
	CreateServer(ctx context.Context, server *Server) (int64, error)
	ListServers(ctx context.Context) ([]Server, error)
	ListByGame(ctx context.Context, gameID int64) ([]Server, error)
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
	UpsertDisplayName(ctx context.Context, gameID int64, internalID, config, name string) error
}

// CarRepository manages car definitions.
type CarRepository interface {
	FindOrCreate(ctx context.Context, gameID int64, internalID, name, class string) (*Car, error)
	GetByID(ctx context.Context, id int64) (*Car, error)
	ListByGame(ctx context.Context, gameID int64) ([]Car, error)
	UpsertDisplayName(ctx context.Context, gameID int64, internalID, name string) error
}

// PlayerRepository manages player records.
type PlayerRepository interface {
	FindOrCreate(ctx context.Context, platform, platformID, name, country string) (*Player, error)
	GetByID(ctx context.Context, id int64) (*Player, error)
	SearchByName(ctx context.Context, name string) ([]Player, error)
}
