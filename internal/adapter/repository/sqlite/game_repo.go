package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/chris/laptime-leaderboard/internal/domain"
)

// GameRepo implements port.GameRepository for SQLite.
type GameRepo struct {
	db *DB
}

func NewGameRepo(db *DB) *GameRepo {
	return &GameRepo{db: db}
}

func (r *GameRepo) GetGameBySlug(ctx context.Context, slug string) (*domain.Game, error) {
	var g domain.Game
	err := r.db.conn.QueryRowContext(ctx,
		"SELECT id, slug, name FROM games WHERE slug = ?", slug,
	).Scan(&g.ID, &g.Slug, &g.Name)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("game not found: " + slug)
	}
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *GameRepo) CreateGame(ctx context.Context, game *domain.Game) (int64, error) {
	result, err := r.db.conn.ExecContext(ctx,
		"INSERT INTO games (slug, name) VALUES (?, ?)", game.Slug, game.Name,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *GameRepo) ListGames(ctx context.Context) ([]domain.Game, error) {
	rows, err := r.db.conn.QueryContext(ctx, "SELECT id, slug, name FROM games ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []domain.Game
	for rows.Next() {
		var g domain.Game
		if err := rows.Scan(&g.ID, &g.Slug, &g.Name); err != nil {
			return nil, err
		}
		games = append(games, g)
	}
	return games, rows.Err()
}
