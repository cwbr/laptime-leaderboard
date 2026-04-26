package domain

import (
	"context"
	"errors"
	"log/slog"
	"time"
)

// Service is the core application service that orchestrates lap ingestion
// and leaderboard queries. It implements port.LapIngester.
type Service struct {
	laps    LapRepository
	servers ServerRepository
	games   GameRepository
	tracks  TrackRepository
	cars    CarRepository
	players PlayerRepository
	logger  *slog.Logger
}

func NewService(
	laps LapRepository,
	servers ServerRepository,
	games GameRepository,
	tracks TrackRepository,
	cars CarRepository,
	players PlayerRepository,
	logger *slog.Logger,
) *Service {
	return &Service{
		laps:    laps,
		servers: servers,
		games:   games,
		tracks:  tracks,
		cars:    cars,
		players: players,
		logger:  logger,
	}
}

// IngestLap validates and stores a lap from any data source.
func (s *Service) IngestLap(ctx context.Context, req IngestLapRequest) error {
	// 1. Authenticate server
	server, err := s.servers.GetServerByAPIKey(ctx, req.ServerAPIKey)
	if err != nil {
		return errors.New("invalid server API key")
	}

	// 2. Resolve game
	game, err := s.games.GetGameBySlug(ctx, req.GameSlug)
	if err != nil {
		return errors.New("unknown game: " + req.GameSlug)
	}

	// Verify server belongs to this game
	if server.GameID != game.ID {
		return errors.New("server is not registered for this game")
	}

	// 3. Find or create track, car, player
	track, err := s.tracks.FindOrCreate(ctx, game.ID, req.Track, req.TrackConfig, req.TrackName)
	if err != nil {
		return err
	}

	car, err := s.cars.FindOrCreate(ctx, game.ID, req.Car, req.CarName, req.CarClass)
	if err != nil {
		return err
	}

	player, err := s.players.FindOrCreate(ctx, req.PlayerPlatform, req.PlayerID, req.PlayerName, req.PlayerCountry)
	if err != nil {
		return err
	}

	// 4. Store the lap
	lap := &Lap{
		ServerID:    server.ID,
		PlayerID:    player.ID,
		TrackID:     track.ID,
		CarID:       car.ID,
		LapTimeMs:   req.LapTimeMs,
		SectorsMs:   req.SectorsMs,
		Cuts:        req.Cuts,
		Valid:       req.Valid,
		Grip:        req.Grip,
		SessionType: req.SessionType,
		CreatedAt:   time.Now().UTC(),
	}

	id, err := s.laps.StoreLap(ctx, lap)
	if err != nil {
		return err
	}

	s.logger.Info("lap ingested",
		"lap_id", id,
		"player", req.PlayerName,
		"track", req.Track,
		"car", req.Car,
		"time_ms", req.LapTimeMs,
		"valid", req.Valid,
	)

	return nil
}

// GetLeaderboard returns the leaderboard for a given query, with delta times.
func (s *Service) GetLeaderboard(ctx context.Context, query LeaderboardQuery) ([]LeaderboardEntry, error) {
	entries, err := s.laps.GetLeaderboard(ctx, query)
	if err != nil {
		return nil, err
	}
	// Compute delta to leader
	if len(entries) > 0 {
		leaderTime := entries[0].LapTimeMs
		for i := range entries {
			entries[i].DeltaMs = int32(entries[i].LapTimeMs) - int32(leaderTime)
		}
	}
	return entries, nil
}

// GetGames returns all registered games.
func (s *Service) GetGames(ctx context.Context) ([]Game, error) {
	return s.games.ListGames(ctx)
}

// GetTracks returns all tracks for a game.
func (s *Service) GetTracks(ctx context.Context, gameID int64) ([]Track, error) {
	return s.tracks.ListByGame(ctx, gameID)
}

// GetCars returns all cars for a game.
func (s *Service) GetCars(ctx context.Context, gameID int64) ([]Car, error) {
	return s.cars.ListByGame(ctx, gameID)
}

// GetPlayer returns a player by ID.
func (s *Service) GetPlayer(ctx context.Context, id int64) (*Player, error) {
	return s.players.GetByID(ctx, id)
}

// GetCar returns a car by ID.
func (s *Service) GetCar(ctx context.Context, id int64) (*Car, error) {
	return s.cars.GetByID(ctx, id)
}

// GetLapDetail returns a fully resolved lap.
func (s *Service) GetLapDetail(ctx context.Context, lapID int64) (*LapDetail, error) {
	return s.laps.GetLapDetail(ctx, lapID)
}

// GetPlayerBestLaps returns the player's best lap per track.
func (s *Service) GetPlayerBestLaps(ctx context.Context, playerID int64, limit, offset int) ([]PlayerBestLap, int, error) {
	return s.laps.GetPlayerBestLaps(ctx, playerID, limit, offset)
}

// GetPlayerRecentLaps returns the player's most recent laps.
func (s *Service) GetPlayerRecentLaps(ctx context.Context, playerID int64, limit, offset int) ([]PlayerRecentLap, int, error) {
	return s.laps.GetPlayerRecentLaps(ctx, playerID, limit, offset)
}

// GetCarTrackBests returns the fastest lap per track for a car.
func (s *Service) GetCarTrackBests(ctx context.Context, carID int64, limit, offset int) ([]CarTrackBest, int, error) {
	return s.laps.GetCarTrackBests(ctx, carID, limit, offset)
}
