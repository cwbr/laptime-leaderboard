-- Initial schema
CREATE TABLE IF NOT EXISTS games (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS servers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    game_id INTEGER NOT NULL REFERENCES games(id),
    name TEXT NOT NULL,
    api_key_hash TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS tracks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    game_id INTEGER NOT NULL REFERENCES games(id),
    internal_id TEXT NOT NULL,
    config TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL DEFAULT '',
    UNIQUE(game_id, internal_id, config)
);

CREATE TABLE IF NOT EXISTS cars (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    game_id INTEGER NOT NULL REFERENCES games(id),
    internal_id TEXT NOT NULL,
    name TEXT NOT NULL DEFAULT '',
    class TEXT NOT NULL DEFAULT '',
    UNIQUE(game_id, internal_id)
);

CREATE TABLE IF NOT EXISTS players (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    platform TEXT NOT NULL,
    platform_id TEXT NOT NULL,
    name TEXT NOT NULL DEFAULT '',
    country TEXT NOT NULL DEFAULT '',
    UNIQUE(platform, platform_id)
);

CREATE TABLE IF NOT EXISTS laps (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    server_id INTEGER NOT NULL REFERENCES servers(id),
    player_id INTEGER NOT NULL REFERENCES players(id),
    track_id INTEGER NOT NULL REFERENCES tracks(id),
    car_id INTEGER NOT NULL REFERENCES cars(id),
    lap_time_ms INTEGER NOT NULL,
    sectors_json TEXT, -- JSON array of sector times in ms
    cuts INTEGER NOT NULL DEFAULT 0,
    valid INTEGER NOT NULL DEFAULT 1,
    grip REAL,
    session_type TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for leaderboard queries
CREATE INDEX IF NOT EXISTS idx_laps_track_valid ON laps(track_id, valid, lap_time_ms);
CREATE INDEX IF NOT EXISTS idx_laps_player ON laps(player_id);
CREATE INDEX IF NOT EXISTS idx_laps_car ON laps(car_id);
CREATE INDEX IF NOT EXISTS idx_laps_created ON laps(created_at);

-- Seed default game
INSERT OR IGNORE INTO games (slug, name) VALUES ('assetto-corsa', 'Assetto Corsa');
