package sqlite

import (
	"context"
	"encoding/json"
	"fmt"

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
	// Find the best (minimum) valid lap per player, then join back to get the full row.
	outerWhere := "l.valid = 1 AND l.track_id = ?"
	innerWhere := "l2.valid = 1 AND l2.track_id = ?"
	args := []any{query.TrackID}

	if query.CarID > 0 {
		outerWhere += " AND l.car_id = ?"
		innerWhere += " AND l2.car_id = ?"
		args = append(args, query.CarID)
	}

	if query.GameID > 0 {
		outerWhere += " AND t.game_id = ?"
		innerWhere += " AND t2.game_id = ?"
		args = append(args, query.GameID)
	}

	limit := query.Limit
	if limit <= 0 {
		limit = 50
	}

	q := `
		SELECT l.id, p.id, p.name, p.country, c.id, c.name, c.internal_id,
		       l.lap_time_ms, l.sectors_json, l.grip, l.created_at
		FROM laps l
		JOIN players p ON p.id = l.player_id
		JOIN cars c ON c.id = l.car_id
		JOIN tracks t ON t.id = l.track_id
		INNER JOIN (
			SELECT l2.player_id, MIN(l2.lap_time_ms) AS best_time
			FROM laps l2
			JOIN tracks t2 ON t2.id = l2.track_id
			WHERE ` + innerWhere + `
			GROUP BY l2.player_id
		) best ON best.player_id = l.player_id AND best.best_time = l.lap_time_ms
		WHERE ` + outerWhere + `
		GROUP BY l.player_id
		ORDER BY l.lap_time_ms ASC
		LIMIT ? OFFSET ?
	`
	// args are used twice (once in subquery, once in outer WHERE)
	allArgs := make([]any, 0, len(args)*2+2)
	allArgs = append(allArgs, args...) // inner WHERE
	allArgs = append(allArgs, args...) // outer WHERE
	allArgs = append(allArgs, limit, query.Offset)

	rows, err := r.db.conn.QueryContext(ctx, q, allArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []domain.LeaderboardEntry
	rank := query.Offset
	for rows.Next() {
		rank++
		var e domain.LeaderboardEntry
		var sectorsJSON *string
		err := rows.Scan(&e.LapID, &e.PlayerID, &e.PlayerName, &e.Country, &e.CarID, &e.CarName, &e.CarInternal, &e.LapTimeMs, &sectorsJSON, &e.Grip, &e.CreatedAt)
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

func (r *LapRepo) GetLapDetail(ctx context.Context, lapID int64) (*domain.LapDetail, error) {
	var d domain.LapDetail
	var sectorsJSON *string
	err := r.db.conn.QueryRowContext(ctx, `
		SELECT l.id, l.server_id, l.player_id, l.track_id, l.car_id,
		       l.lap_time_ms, l.sectors_json, l.cuts, l.valid, l.grip, l.session_type, l.created_at,
		       p.name, p.country, t.name, t.config, c.name, c.class, s.name, g.name
		FROM laps l
		JOIN players p ON p.id = l.player_id
		JOIN tracks t ON t.id = l.track_id
		JOIN cars c ON c.id = l.car_id
		JOIN servers s ON s.id = l.server_id
		JOIN games g ON g.id = t.game_id
		WHERE l.id = ?
	`, lapID).Scan(
		&d.ID, &d.ServerID, &d.PlayerID, &d.TrackID, &d.CarID,
		&d.LapTimeMs, &sectorsJSON, &d.Cuts, &d.Valid, &d.Grip, &d.SessionType, &d.CreatedAt,
		&d.PlayerName, &d.PlayerCountry, &d.TrackName, &d.TrackConfig, &d.CarName, &d.CarClass,
		&d.ServerName, &d.GameName,
	)
	if err != nil {
		return nil, err
	}
	if sectorsJSON != nil {
		_ = json.Unmarshal([]byte(*sectorsJSON), &d.SectorsMs)
	}
	return &d, nil
}

func (r *LapRepo) GetPlayerBestLaps(ctx context.Context, playerID int64, limit, offset int) ([]domain.PlayerBestLap, int, error) {
	if limit <= 0 {
		limit = 50
	}

	// Count total
	var total int
	err := r.db.conn.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT l.track_id)
		FROM laps l
		INNER JOIN (
			SELECT track_id, MIN(lap_time_ms) AS best_time
			FROM laps WHERE player_id = ? AND valid = 1
			GROUP BY track_id
		) best ON best.track_id = l.track_id AND best.best_time = l.lap_time_ms
		WHERE l.player_id = ? AND l.valid = 1
	`, playerID, playerID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.conn.QueryContext(ctx, `
		SELECT t.id, t.name, c.name, l.lap_time_ms, l.sectors_json, l.grip, l.created_at
		FROM laps l
		JOIN tracks t ON t.id = l.track_id
		JOIN cars c ON c.id = l.car_id
		INNER JOIN (
			SELECT track_id, MIN(lap_time_ms) AS best_time
			FROM laps
			WHERE player_id = ? AND valid = 1
			GROUP BY track_id
		) best ON best.track_id = l.track_id AND best.best_time = l.lap_time_ms
		WHERE l.player_id = ? AND l.valid = 1
		GROUP BY l.track_id
		ORDER BY t.name
		LIMIT ? OFFSET ?
	`, playerID, playerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []domain.PlayerBestLap
	for rows.Next() {
		var b domain.PlayerBestLap
		var sectorsJSON *string
		if err := rows.Scan(&b.TrackID, &b.TrackName, &b.CarName, &b.LapTimeMs, &sectorsJSON, &b.Grip, &b.CreatedAt); err != nil {
			return nil, 0, err
		}
		if sectorsJSON != nil {
			_ = json.Unmarshal([]byte(*sectorsJSON), &b.SectorsMs)
		}
		results = append(results, b)
	}
	return results, total, rows.Err()
}

func (r *LapRepo) GetPlayerRecentLaps(ctx context.Context, playerID int64, limit, offset int) ([]domain.PlayerRecentLap, int, error) {
	if limit <= 0 {
		limit = 50
	}

	var total int
	err := r.db.conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM laps WHERE player_id = ?`, playerID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.conn.QueryContext(ctx, `
		SELECT l.id, t.name, c.name, l.lap_time_ms, l.sectors_json, l.valid, l.grip, l.created_at
		FROM laps l
		JOIN tracks t ON t.id = l.track_id
		JOIN cars c ON c.id = l.car_id
		WHERE l.player_id = ?
		ORDER BY l.created_at DESC
		LIMIT ? OFFSET ?
	`, playerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []domain.PlayerRecentLap
	for rows.Next() {
		var r domain.PlayerRecentLap
		var sectorsJSON *string
		if err := rows.Scan(&r.LapID, &r.TrackName, &r.CarName, &r.LapTimeMs, &sectorsJSON, &r.Valid, &r.Grip, &r.CreatedAt); err != nil {
			return nil, 0, err
		}
		if sectorsJSON != nil {
			_ = json.Unmarshal([]byte(*sectorsJSON), &r.SectorsMs)
		}
		results = append(results, r)
	}
	return results, total, rows.Err()
}

func (r *LapRepo) GetCarTrackBests(ctx context.Context, carID int64, limit, offset int) ([]domain.CarTrackBest, int, error) {
	if limit <= 0 {
		limit = 50
	}

	var total int
	err := r.db.conn.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT l.track_id)
		FROM laps l
		INNER JOIN (
			SELECT track_id, MIN(lap_time_ms) AS best_time
			FROM laps WHERE car_id = ? AND valid = 1
			GROUP BY track_id
		) best ON best.track_id = l.track_id AND best.best_time = l.lap_time_ms
		WHERE l.car_id = ? AND l.valid = 1
	`, carID, carID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.conn.QueryContext(ctx, `
		SELECT t.id, t.name, p.name, l.lap_time_ms, l.sectors_json, l.created_at
		FROM laps l
		JOIN tracks t ON t.id = l.track_id
		JOIN players p ON p.id = l.player_id
		INNER JOIN (
			SELECT track_id, MIN(lap_time_ms) AS best_time
			FROM laps
			WHERE car_id = ? AND valid = 1
			GROUP BY track_id
		) best ON best.track_id = l.track_id AND best.best_time = l.lap_time_ms
		WHERE l.car_id = ? AND l.valid = 1
		GROUP BY l.track_id
		ORDER BY t.name
		LIMIT ? OFFSET ?
	`, carID, carID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []domain.CarTrackBest
	for rows.Next() {
		var b domain.CarTrackBest
		var sectorsJSON *string
		if err := rows.Scan(&b.TrackID, &b.TrackName, &b.PlayerName, &b.LapTimeMs, &sectorsJSON, &b.CreatedAt); err != nil {
			return nil, 0, err
		}
		if sectorsJSON != nil {
			_ = json.Unmarshal([]byte(*sectorsJSON), &b.SectorsMs)
		}
		results = append(results, b)
	}
	return results, total, rows.Err()
}

func (r *LapRepo) DeleteLap(ctx context.Context, lapID int64) error {
	result, err := r.db.conn.ExecContext(ctx, "DELETE FROM laps WHERE id = ?", lapID)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("lap %d not found", lapID)
	}
	return nil
}

func (r *LapRepo) DeletePlayerLaps(ctx context.Context, playerID int64) (int64, error) {
	result, err := r.db.conn.ExecContext(ctx, "DELETE FROM laps WHERE player_id = ?", playerID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
