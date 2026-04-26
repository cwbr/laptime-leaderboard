package sqlite

import (
	"context"

	"github.com/chris/laptime-leaderboard/internal/domain"
)

// PlayerRepo implements port.PlayerRepository for SQLite.
type PlayerRepo struct {
	db *DB
}

func NewPlayerRepo(db *DB) *PlayerRepo {
	return &PlayerRepo{db: db}
}

func (r *PlayerRepo) FindOrCreate(ctx context.Context, platform, platformID, name, country string) (*domain.Player, error) {
	var p domain.Player
	err := r.db.conn.QueryRowContext(ctx,
		"SELECT id, platform, platform_id, name, country FROM players WHERE platform = ? AND platform_id = ?",
		platform, platformID,
	).Scan(&p.ID, &p.Platform, &p.PlatformID, &p.Name, &p.Country)
	if err == nil {
		// Update name/country if changed
		if p.Name != name || p.Country != country {
			_, _ = r.db.conn.ExecContext(ctx,
				"UPDATE players SET name = ?, country = ? WHERE id = ?",
				name, country, p.ID,
			)
			p.Name = name
			p.Country = country
		}
		return &p, nil
	}

	result, err := r.db.conn.ExecContext(ctx,
		"INSERT INTO players (platform, platform_id, name, country) VALUES (?, ?, ?, ?)",
		platform, platformID, name, country,
	)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return &domain.Player{ID: id, Platform: platform, PlatformID: platformID, Name: name, Country: country}, nil
}

func (r *PlayerRepo) GetByID(ctx context.Context, id int64) (*domain.Player, error) {
	var p domain.Player
	err := r.db.conn.QueryRowContext(ctx,
		"SELECT id, platform, platform_id, name, country FROM players WHERE id = ?", id,
	).Scan(&p.ID, &p.Platform, &p.PlatformID, &p.Name, &p.Country)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PlayerRepo) SearchByName(ctx context.Context, name string) ([]domain.Player, error) {
	rows, err := r.db.conn.QueryContext(ctx,
		"SELECT id, platform, platform_id, name, country FROM players WHERE name LIKE ? ORDER BY name LIMIT 25",
		"%"+name+"%",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []domain.Player
	for rows.Next() {
		var p domain.Player
		if err := rows.Scan(&p.ID, &p.Platform, &p.PlatformID, &p.Name, &p.Country); err != nil {
			return nil, err
		}
		players = append(players, p)
	}
	return players, rows.Err()
}
