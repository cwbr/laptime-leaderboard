package sqlite

import (
	"context"

	"github.com/chris/laptime-leaderboard/internal/domain"
)

// CarRepo implements port.CarRepository for SQLite.
type CarRepo struct {
	db *DB
}

func NewCarRepo(db *DB) *CarRepo {
	return &CarRepo{db: db}
}

func (r *CarRepo) FindOrCreate(ctx context.Context, gameID int64, internalID, name, class string) (*domain.Car, error) {
	var c domain.Car
	err := r.db.conn.QueryRowContext(ctx,
		"SELECT id, game_id, internal_id, name, class FROM cars WHERE game_id = ? AND internal_id = ?",
		gameID, internalID,
	).Scan(&c.ID, &c.GameID, &c.InternalID, &c.Name, &c.Class)
	if err == nil {
		return &c, nil
	}

	displayName := name
	if displayName == "" {
		displayName = internalID
	}

	result, err := r.db.conn.ExecContext(ctx,
		"INSERT INTO cars (game_id, internal_id, name, class) VALUES (?, ?, ?, ?)",
		gameID, internalID, displayName, class,
	)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return &domain.Car{ID: id, GameID: gameID, InternalID: internalID, Name: displayName, Class: class}, nil
}

func (r *CarRepo) GetByID(ctx context.Context, id int64) (*domain.Car, error) {
	var c domain.Car
	err := r.db.conn.QueryRowContext(ctx,
		"SELECT id, game_id, internal_id, name, class FROM cars WHERE id = ?", id,
	).Scan(&c.ID, &c.GameID, &c.InternalID, &c.Name, &c.Class)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CarRepo) ListByGame(ctx context.Context, gameID int64) ([]domain.Car, error) {
	rows, err := r.db.conn.QueryContext(ctx,
		"SELECT id, game_id, internal_id, name, class FROM cars WHERE game_id = ? ORDER BY name", gameID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cars []domain.Car
	for rows.Next() {
		var c domain.Car
		if err := rows.Scan(&c.ID, &c.GameID, &c.InternalID, &c.Name, &c.Class); err != nil {
			return nil, err
		}
		cars = append(cars, c)
	}
	return cars, rows.Err()
}
