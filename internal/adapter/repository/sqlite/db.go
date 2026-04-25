package sqlite

import (
	"database/sql"
	"embed"
	"fmt"
	"log/slog"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// DB wraps *sql.DB and provides repository implementations.
type DB struct {
	conn   *sql.DB
	logger *slog.Logger
}

// New opens a SQLite database and runs migrations.
func New(dsn string, logger *slog.Logger) (*DB, error) {
	conn, err := sql.Open("sqlite3", dsn+"?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	db := &DB{conn: conn, logger: logger}
	if err := db.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return db, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) migrate() error {
	// Create migrations tracking table
	if _, err := db.conn.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		return err
	}

	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return err
	}

	for i, entry := range entries {
		version := i + 1
		var applied int
		err := db.conn.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", version).Scan(&applied)
		if err != nil {
			return err
		}
		if applied > 0 {
			continue
		}

		content, err := migrationsFS.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return err
		}

		if _, err := db.conn.Exec(string(content)); err != nil {
			return fmt.Errorf("migration %s: %w", entry.Name(), err)
		}

		if _, err := db.conn.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version); err != nil {
			return err
		}

		db.logger.Info("applied migration", "file", entry.Name(), "version", version)
	}

	return nil
}
