package sqlite

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"

	"github.com/chris/laptime-leaderboard/internal/domain"
)

// ServerRepo implements port.ServerRepository for SQLite.
type ServerRepo struct {
	db *DB
}

func NewServerRepo(db *DB) *ServerRepo {
	return &ServerRepo{db: db}
}

func hashAPIKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}

func (r *ServerRepo) GetServerByAPIKey(ctx context.Context, apiKey string) (*domain.Server, error) {
	hash := hashAPIKey(apiKey)
	var s domain.Server
	err := r.db.conn.QueryRowContext(ctx,
		"SELECT id, game_id, name FROM servers WHERE api_key_hash = ?", hash,
	).Scan(&s.ID, &s.GameID, &s.Name)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("server not found")
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *ServerRepo) CreateServer(ctx context.Context, server *domain.Server) (int64, error) {
	hash := hashAPIKey(server.APIKey)
	result, err := r.db.conn.ExecContext(ctx,
		"INSERT INTO servers (game_id, name, api_key_hash) VALUES (?, ?, ?)",
		server.GameID, server.Name, hash,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *ServerRepo) ListServers(ctx context.Context) ([]domain.Server, error) {
	rows, err := r.db.conn.QueryContext(ctx, "SELECT id, game_id, name FROM servers")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []domain.Server
	for rows.Next() {
		var s domain.Server
		if err := rows.Scan(&s.ID, &s.GameID, &s.Name); err != nil {
			return nil, err
		}
		servers = append(servers, s)
	}
	return servers, rows.Err()
}

func (r *ServerRepo) ListByGame(ctx context.Context, gameID int64) ([]domain.Server, error) {
	rows, err := r.db.conn.QueryContext(ctx, "SELECT id, game_id, name FROM servers WHERE game_id = ?", gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []domain.Server
	for rows.Next() {
		var s domain.Server
		if err := rows.Scan(&s.ID, &s.GameID, &s.Name); err != nil {
			return nil, err
		}
		servers = append(servers, s)
	}
	return servers, rows.Err()
}
