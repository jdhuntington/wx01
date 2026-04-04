# wx01

Lightweight weather station dashboard for the WeatherFlow Tempest. Single Go binary with embedded React frontend, backed by TimescaleDB.

## Architecture

```
Tempest station --UDP--> Go binary --INSERT + pg_notify--> TimescaleDB (Docker)
                                                              |
                                                    LISTEN ---+
                                                              |
                         Go binary --SSE--> React SPA     <---+
```

- **UDP listener** (`internal/ingest/`) — receives Tempest packets on port 50222, parses 6 message types (obs_st, rapid_wind, evt_precip, evt_strike, device_status, hub_status), inserts into DB
- **HTTP API** (`internal/api/`) — REST endpoints serve time-bucketed aggregates and current conditions; SSE endpoint for real-time push; serves the embedded SPA
- **Notify hub** (`internal/notify/`) — dedicated pgx connection LISTENs on `wx01_data` channel, broadcasts to SSE clients
- **Database** (`internal/db/`) — pgx connection pool, embedded SQL migrations, TimescaleDB hypertables with automatic compression; `pg_notify` after each insert for real-time push
- **Frontend** (`web/src/`) — React 19 + lightweight-charts, SSE-driven (no polling), dark theme

## Key files

| Area | Files |
|------|-------|
| Entrypoint | `cmd/wx01/main.go` — config, lifecycle, embeds `dist/` |
| UDP parsing | `internal/ingest/packet.go` — 6 message type parsers |
| DB schema | `internal/db/migrations/001_create_tables.sql` — hypertables |
| DB views | `internal/db/migrations/002_create_views.sql` — aggregation views |
| Notify hub | `internal/notify/hub.go` — PG LISTEN + SSE broadcast |
| API routes | `internal/api/server.go` — 9 endpoints (8 data + SSE), bucketing logic |
| React app | `web/src/App.tsx` — layout, state, 7 chart instances |
| Charts | `web/src/components/TimeSeriesChart.tsx` — lightweight-charts wrapper |
| Conditions | `web/src/components/CurrentConditions.tsx` — 7-card metric grid with sparklines |
| Sparklines | `web/src/components/Sparkline.tsx` — SVG polyline sparkline |
| Unit conversion | `web/src/units.ts` — metric/imperial with localStorage |
| Styling | `web/src/style.css` — plain CSS, dark theme, responsive breakpoints |

## Build & dev

```bash
make build          # frontend + Go binary (macOS)
make build-linux    # cross-compile for Linux amd64
make dev            # Go server only (use `cd web && npm run dev` for Vite)
make deploy TARGET=user@host   # build + scp + install via systemd
```

Frontend dev: `cd web && npm run dev` proxies `/api` to `localhost:3100` (see `web/vite.config.ts`).

The Go binary embeds `cmd/wx01/dist/` at compile time via `//go:embed`. The Makefile copies `web/dist/` there before building.

## Environment variables

| Var | Default |
|-----|---------|
| `WX01_DATABASE_URL` | `postgres://wx01:wx01@localhost:5432/wx01?sslmode=disable` |
| `WX01_UDP_PORT` | `50222` |
| `WX01_HTTP_PORT` | `3100` |

## Deployment

Deploys as a systemd service (`deploy/wx01.service`). Database runs as a Docker container managed by `deploy/wx01-db` (TimescaleDB on port 5433, `--network=host` for Incus compatibility). Backup script at `deploy/wx01-backup` retains last 30 dumps.

The install script (`deploy/install.sh`) handles: stop service, copy binary + helpers, pull Docker image if needed, start service.

## API endpoints

All time-series endpoints accept `?range=` (6h, 24h, 168h, 720h). Bucketing: 5min for <=6h, 15min for 24h, 1h for 7d, 6h for 30d.

- `GET /api/current` — latest observation + rain stats
- `GET /api/temperature` — avg/min/max temp per bucket
- `GET /api/humidity` — avg/min/max humidity per bucket
- `GET /api/wind` — avg wind, max gust, min lull per bucket
- `GET /api/pressure` — avg station pressure per bucket
- `GET /api/rain` — summed rain per bucket (histogram)
- `GET /api/solar` — avg/max solar radiation per bucket
- `GET /api/uv` — avg/max UV index per bucket
- `GET /api/events` — SSE stream; emits named events (obs_st, rapid_wind, etc.) on new data

## Frontend notes

- Charts use [lightweight-charts](https://github.com/nicepkg/lightweight-charts) v5 (financial charting lib)
- Chart height is 220px normally, 100px in compact mode (<= 400px viewport)
- Unit preference persisted in localStorage (`wx01_units`)
- Real-time updates via SSE (`/api/events`) — no polling; a global `EventSource` bumps a version counter that triggers refetches
- Compact mode detected via `window.matchMedia` and CSS media queries
- Sparklines in condition cards use the same time-series data as the main charts

## Real-time data flow

The collector and web server are decoupled via PostgreSQL `LISTEN/NOTIFY`:

1. Store inserts data and calls `pg_notify('wx01_data', 'obs_st')` (or other message type)
2. The notify hub holds a dedicated pgx connection with `LISTEN wx01_data`
3. On notification, hub broadcasts to all connected SSE clients
4. Frontend `EventSource` receives the event and triggers a refetch of all data

This design means the collector and web server only share a database connection string — they can be split into separate binaries later with no code changes.
