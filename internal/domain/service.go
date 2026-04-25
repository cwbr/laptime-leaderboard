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

// GetLeaderboard returns the leaderboard for a given query.
func (s *Service) GetLeaderboard(ctx context.Context, query LeaderboardQuery) ([]LeaderboardEntry, error) {
	return s.laps.GetLeaderboard(ctx, query)
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
