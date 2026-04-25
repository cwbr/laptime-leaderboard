package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	// HTTP server listen address
	ListenAddr string `json:"listen_addr"`
	// SQLite database file path
	DatabasePath string `json:"database_path"`
}

func DefaultConfig() Config {
	return Config{
		ListenAddr:   ":8080",
		DatabasePath: "leaderboard.db",
	}
}

// Load reads config from a JSON file, falling back to defaults.
func Load(path string) (Config, error) {
	cfg := DefaultConfig()

	// Override with env vars if set
	if v := os.Getenv("LEADERBOARD_LISTEN_ADDR"); v != "" {
		cfg.ListenAddr = v
	}
	if v := os.Getenv("LEADERBOARD_DB_PATH"); v != "" {
		cfg.DatabasePath = v
	}

	if path == "" {
		return cfg, nil
	}

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
