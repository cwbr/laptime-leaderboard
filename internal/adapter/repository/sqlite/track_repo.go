package sqlite

import (
	"context"

	"github.com/chris/laptime-leaderboard/internal/domain"
)

// TrackRepo implements port.TrackRepository for SQLite.
type TrackRepo struct {
	db *DB
}

func NewTrackRepo(db *DB) *TrackRepo {
	return &TrackRepo{db: db}
}

func (r *TrackRepo) FindOrCreate(ctx context.Context, gameID int64, internalID, config, name string) (*domain.Track, error) {
	var t domain.Track
	err := r.db.conn.QueryRowContext(ctx,
		"SELECT id, game_id, internal_id, config, name FROM tracks WHERE game_id = ? AND internal_id = ? AND config = ?",
		gameID, internalID, config,
	).Scan(&t.ID, &t.GameID, &t.InternalID, &t.Config, &t.Name)
	if err == nil {
		return &t, nil
	}

	// Auto-create
	displayName := name
	if displayName == "" {
		displayName = internalID
		if config != "" {
			displayName += " (" + config + ")"
		}
	}

	result, err := r.db.conn.ExecContext(ctx,
		"INSERT INTO tracks (game_id, internal_id, config, name) VALUES (?, ?, ?, ?)",
		gameID, internalID, config, displayName,
	)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return &domain.Track{ID: id, GameID: gameID, InternalID: internalID, Config: config, Name: displayName}, nil
}

func (r *TrackRepo) GetByID(ctx context.Context, id int64) (*domain.Track, error) {
	var t domain.Track
	err := r.db.conn.QueryRowContext(ctx,
		"SELECT id, game_id, internal_id, config, name FROM tracks WHERE id = ?", id,
	).Scan(&t.ID, &t.GameID, &t.InternalID, &t.Config, &t.Name)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TrackRepo) ListByGame(ctx context.Context, gameID int64) ([]domain.Track, error) {
	rows, err := r.db.conn.QueryContext(ctx, `
		SELECT DISTINCT t.id, t.game_id, t.internal_id, t.config, t.name
		FROM tracks t
		JOIN laps l ON l.track_id = t.id AND l.valid = 1
		WHERE t.game_id = ?
		ORDER BY t.name`, gameID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracks []domain.Track
	for rows.Next() {
		var t domain.Track
		if err := rows.Scan(&t.ID, &t.GameID, &t.InternalID, &t.Config, &t.Name); err != nil {
			return nil, err
		}
		tracks = append(tracks, t)
	}
	return tracks, rows.Err()
}

func (r *TrackRepo) UpsertDisplayName(ctx context.Context, gameID int64, internalID, config, name string) error {
	_, err := r.db.conn.ExecContext(ctx, `
		INSERT INTO tracks (game_id, internal_id, config, name) VALUES (?, ?, ?, ?)
		ON CONFLICT(game_id, internal_id, config) DO UPDATE SET name = excluded.name
	`, gameID, internalID, config, name)
	return err
}
