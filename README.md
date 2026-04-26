# Laptime Leaderboard

A self-hosted lap time leaderboard for sim racing servers. Supports multiple games, servers, tracks, and cars. Built to work with Assetto Corsa via the [LeaderboardPlugin](https://github.com/cwbr/AssettoServer/tree/feature/leaderboard-plugin) for AssettoServer, but the REST API is game-agnostic.

## Features

- **Multi-game / multi-server** — one leaderboard instance aggregates data from many game servers
- **Best lap per player** — leaderboard shows each driver's personal best per track
- **Infinite scroll** — main leaderboard loads rows on scroll via HTMX
- **Detail pages** — driver profiles (personal bests + recent laps), car pages (track records), individual lap detail
- **Cascading filters** — filter by game, track, and car with HTMX partial swaps
- **Client-side sorting** — click column headers; auto-fetches remaining rows before sorting
- **Car/track name mappings** — import display names from JSON files
- **Admin CLI** — manage servers, games, mappings, and moderate cheaters
- **Dark theme** — minimal, responsive design

## Tech Stack

- **Go** (single binary, no runtime dependencies beyond libc)
- **SQLite** (WAL mode, via `mattn/go-sqlite3`)
- **HTMX 2.0** (CDN-loaded, used for filter swaps and infinite scroll)
- **HTML templates** (Go `html/template`, embedded via `//go:embed`)

## Quick Start

### Prerequisites

- Go 1.22+ with CGO enabled (`CGO_ENABLED=1`)
- A C compiler (gcc/clang) for SQLite

### Build & Run

```bash
go build ./cmd/leaderboard/
./leaderboard serve
```

Server starts on `:8080` by default. Visit http://localhost:8080.

### Docker

```bash
docker compose up -d
```

### Configuration

Environment variables (or a JSON config file via `LEADERBOARD_CONFIG`):

| Variable | Default | Description |
|---|---|---|
| `LEADERBOARD_LISTEN_ADDR` | `:8080` | HTTP listen address |
| `LEADERBOARD_DB_PATH` | `leaderboard.db` | SQLite database path |

## Setup

### 1. Create a game

```bash
./leaderboard admin create-game --slug assetto-corsa --name "Assetto Corsa"
```

### 2. Import car/track display names (optional)

```bash
./leaderboard admin import-mappings --game assetto-corsa --file mappings/assetto-corsa.json
```

### 3. Register a server

```bash
./leaderboard admin create-server --name "My Server" --game assetto-corsa
```

This outputs an API key. Configure the game server plugin with this key.

## API

### Ingest a lap

```
POST /api/v1/laps
Authorization: Bearer <api-key>
Content-Type: application/json
```

```json
{
  "game_slug": "assetto-corsa",
  "player_platform": "steam",
  "player_id": "76561198012345678",
  "player_name": "DriverName",
  "player_country": "DE",
  "track": "spa",
  "track_config": "",
  "car": "ks_ferrari_488_gt3",
  "car_name": "Ferrari 488 GT3",
  "lap_time_ms": 138432,
  "sectors_ms": [44100, 50200, 44132],
  "cuts": 0,
  "valid": true,
  "grip": 0.99,
  "session_type": "practice"
}
```

Required fields: `game_slug`, `player_platform`, `player_id`, `track`, `car`, `lap_time_ms`.

### Response

- `201 Created` — `{"status":"ok"}`
- `400/422` — `{"error":"..."}`

## Admin CLI

```
./leaderboard admin <command> [flags]
```

| Command | Description |
|---|---|
| `create-server --name "..." --game slug` | Register a server, returns API key |
| `list-servers` | List all servers |
| `create-game --slug "..." --name "..."` | Create a game |
| `list-games` | List all games |
| `import-mappings --game slug --file path` | Import car/track display names from JSON |
| `find-player --name "..."` | Search players by name (partial match) |
| `delete-lap --id 123` | Delete a specific lap |
| `delete-player-laps --player-id 5` | Delete all laps for a player |

## Project Structure

```
cmd/leaderboard/          Entry point (serve / admin subcommands)
internal/
  domain/                 Entities, repository interfaces, service
  adapter/
    ingest/http/          REST API handler (lap ingestion)
    repository/sqlite/    SQLite implementations + migrations
    web/                  Web UI handler + embedded templates
  admin/                  CLI admin commands
  config/                 Configuration loading
mappings/                 Car/track display name JSON files
```
