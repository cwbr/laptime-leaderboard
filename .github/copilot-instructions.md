# Copilot Instructions — Laptime Leaderboard

## Project Overview

This is a self-hosted lap time leaderboard for sim racing. It ingests lap data from game server plugins via a REST API and presents a web UI with leaderboards, driver profiles, car pages, and lap detail views.

**Stack**: Go + SQLite + HTMX. Single binary, no JS framework, no ORM.

## Architecture

Hexagonal / ports-and-adapters pattern:

- **`internal/domain/`** — Core business logic. Contains entities (`lap.go`), repository interfaces (`repository.go`), the application service (`service.go`), and the ingest request type (`ingester.go`). This package has zero external dependencies.
- **`internal/adapter/ingest/http/`** — Driving adapter. REST API handler that accepts `POST /api/v1/laps` with Bearer auth.
- **`internal/adapter/repository/sqlite/`** — Driven adapter. SQLite implementations of all repository interfaces. Migrations are in `migrations/` subfolder and auto-applied on startup via `db.go`.
- **`internal/adapter/web/`** — Driving adapter. Serves the HTML UI. Templates are embedded via `//go:embed` from `templates/`. Each page gets its own `template.Template` instance to avoid `define` block name collisions.
- **`internal/admin/`** — CLI admin commands (create-server, manage games, moderate players).
- **`internal/config/`** — Config loading (JSON file or env vars).
- **`cmd/leaderboard/`** — Entry point. Routes to `serve` or `admin` subcommand.

## Key Design Decisions

- **One template.Template per page** — `leaderboard.html`, `driver.html`, `car.html`, `lap.html` each define `title` and `content` blocks. They are parsed into separate `template.Template` instances to avoid block name collisions. The template map is `map[string]*template.Template` in `handler.go`.
- **HTMX for partial updates** — Track/car filter changes use `hx-get` to swap `#leaderboard-area`. Game changes use full navigation (`onchange` with `window.location`).
- **Infinite scroll on leaderboard** — A sentinel `<tr class="scroll-sentinel">` with `hx-trigger="revealed"` fetches `/leaderboard/more?offset=N`. When a sort header is clicked with unloaded rows, JS fetches all remaining via `&all=1` before sorting.
- **Pagination on detail pages** — Driver page has two independent page params (`bests_page`, `recent_page`). Car page uses a single `page` param. Standard `PageInfo` struct with `Pages()`, `HasPrev()`, `HasNext()` methods.
- **Delta times** — Computed in the service layer, not SQL. Leader gets `DeltaMs=0`, others get delta from leader's time.
- **Player identity** — Keyed by `(platform, platform_id)`. Same name with different platform IDs creates separate player records.
- **Auth** — API key is SHA256-hashed before storage. Header: `Authorization: Bearer <key>`.

## Database

SQLite with WAL mode, foreign keys enabled, busy timeout 5000ms.

Tables: `games`, `servers`, `players`, `tracks`, `cars`, `laps`.

- `tracks` have `(game_id, internal_id, config)` as the unique key (some tracks have multiple configs like Nürburgring GP/Nordschleife).
- `cars` have `(game_id, internal_id)` as the unique key.
- `laps` reference server, player, track, car via foreign keys.

## Build & Run

```bash
CGO_ENABLED=1 go build ./cmd/leaderboard/
./leaderboard serve          # starts HTTP server on :8080
./leaderboard admin <cmd>    # admin CLI
```

The binary requires CGO for `mattn/go-sqlite3`.

## Template System

Templates live in `internal/adapter/web/templates/`. Key files:

- `base.html` — Layout shell, CSS, and all JS (sort, clickable rows, date formatting, infinite scroll)
- `leaderboard.html` — Full page with filter dropdowns
- `leaderboard_partial.html` — Table-only partial (returned for HTMX swaps and initial render)
- `leaderboard_rows.html` — Rows-only fragment for infinite scroll appends
- `pagination.html` — Shared pagination nav component
- `driver.html`, `car.html`, `lap.html` — Detail pages

Template functions: `lapTime` (ms→"1:38.432"), `deltaTime` (ms→"+0.432"), `plus1`, `mul`.

## Coding Conventions

- No ORM — raw SQL in repository files
- Errors bubble up; logging happens at the handler/service level
- Repository methods return domain types, never SQL-specific types
- `context.Context` is threaded through all layers
- Admin commands print to stdout/stderr and call `os.Exit` on error
- Page size is 50 rows (const `pageSize` in `handler.go`)

## Related Project

The Assetto Corsa server plugin that sends laps to this API lives at:
`/Users/chris/Documents/Code/AssettoServer/LeaderboardPlugin/`

It's a C# BackgroundService that posts lap data via `PostAsJsonAsync` using a `Channel<LapPayload>`.
