package sqlite

import (
	"context"
	"encoding/json"

	"github.com/chris/laptime-leaderboard/internal/domain"
)

// LapRepo implements port.LapRepository for SQLite.
type LapRepo struct {
	db *DB
}

func NewLapRepo(db *DB) *LapRepo {
	return &LapRepo{db: db}
}

func (r *LapRepo) StoreLap(ctx context.Context, lap *domain.Lap) (int64, error) {
	var sectorsJSON []byte
	if len(lap.SectorsMs) > 0 {
		var err error
		sectorsJSON, err = json.Marshal(lap.SectorsMs)
		if err != nil {
			return 0, err
		}
	}

	result, err := r.db.conn.ExecContext(ctx, `
		INSERT INTO laps (server_id, player_id, track_id, car_id, lap_time_ms, sectors_json, cuts, valid, grip, session_type, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, lap.ServerID, lap.PlayerID, lap.TrackID, lap.CarID,
		lap.LapTimeMs, sectorsJSON, lap.Cuts, lap.Valid, lap.Grip, lap.SessionType, lap.CreatedAt)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *LapRepo) GetLeaderboard(ctx context.Context, query domain.LeaderboardQuery) ([]domain.LeaderboardEntry, error) {
	q := `
		SELECT
			p.name,
			p.country,
			c.name,
			c.internal_id,
			l.lap_time_ms,
			l.sectors_json,
			l.grip,
			l.created_at
		FROM laps l
		JOIN players p ON p.id = l.player_id
		JOIN cars c ON c.id = l.car_id
		JOIN tracks t ON t.id = l.track_id
		WHERE l.valid = 1
		  AND t.id = ?
	`
	args := []any{query.TrackID}

	if query.CarID > 0 {
		q += " AND l.car_id = ?"
		args = append(args, query.CarID)
	}

	if query.GameID > 0 {
		q += " AND t.game_id = ?"
		args = append(args, query.GameID)
	}

	// Best lap per player: use a subquery to pick the fastest valid lap per player
	q = `
		SELECT sub.name, sub.country, sub.car_name, sub.car_internal, sub.lap_time_ms, sub.sectors_json, sub.grip, sub.created_at
		FROM (` + q + ` ORDER BY l.lap_time_ms ASC) sub
		GROUP BY sub.name
		ORDER BY sub.lap_time_ms ASC
	`

	limit := query.Limit
	if limit <= 0 {
		limit = 100
	}
	q += " LIMIT ?"
	args = append(args, limit)

	rows, err := r.db.conn.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []domain.LeaderboardEntry
	rank := 0
	for rows.Next() {
		rank++
		var e domain.LeaderboardEntry
		var sectorsJSON *string
		err := rows.Scan(&e.PlayerName, &e.Country, &e.CarName, &e.CarInternal, &e.LapTimeMs, &sectorsJSON, &e.Grip, &e.CreatedAt)
		if err != nil {
			return nil, err
		}
		e.Rank = rank
		if sectorsJSON != nil {
			_ = json.Unmarshal([]byte(*sectorsJSON), &e.SectorsMs)
		}
		entries = append(entries, e)
	}

	return entries, rows.Err()
}

func (r *LapRepo) GetPlayerLaps(ctx context.Context, playerID, trackID int64, limit int) ([]domain.Lap, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.conn.QueryContext(ctx, `
		SELECT id, server_id, player_id, track_id, car_id, lap_time_ms, sectors_json, cuts, valid, grip, session_type, created_at
		FROM laps
		WHERE player_id = ? AND track_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`, playerID, trackID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var laps []domain.Lap
	for rows.Next() {
		var l domain.Lap
		var sectorsJSON *string
		err := rows.Scan(&l.ID, &l.ServerID, &l.PlayerID, &l.TrackID, &l.CarID,
			&l.LapTimeMs, &sectorsJSON, &l.Cuts, &l.Valid, &l.Grip, &l.SessionType, &l.CreatedAt)
		if err != nil {
			return nil, err
		}
		if sectorsJSON != nil {
			_ = json.Unmarshal([]byte(*sectorsJSON), &l.SectorsMs)
		}
		laps = append(laps, l)
	}
	return laps, rows.Err()
}
