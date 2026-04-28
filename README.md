<div align="center">

<img src="web/public/logo.png" alt="wx01" height="96" />

# wx01

**A small dashboard for the [WeatherFlow Tempest](https://weatherflow.com/tempest-weather-system/).**
One Go binary, one Postgres container, real-time updates over the local network.

<img src="docs/screenshot.png" alt="wx01 dashboard" width="720" />

</div>

---

## What this is

A lightweight, self-hosted web dashboard for your weather. Charts for
temperature, humidity, wind, pressure, rain, solar, UV, and lightning, with
a per-strike scatter plot when storms get close.

## Architecture

```mermaid
flowchart LR
    Tempest([Tempest station])
    Go[Go binary<br/>collector + web server]
    DB[(TimescaleDB)]
    SPA([React SPA])

    Tempest -->|UDP :50222| Go
    Go -->|INSERT + pg_notify| DB
    DB -.->|LISTEN| Go
    Go -->|SSE| SPA
```

The collector and the web server live in one binary today, but they only
share a database connection string. Splitting them into separate processes
later wouldn't take any code changes.

For the file map and API endpoints, see [CLAUDE.md](./CLAUDE.md).

## Easiest way to run it

A Linux box (amd64) with Docker and systemd. Grab the prebuilt binary, run
TimescaleDB in a container next to it, point one at the other.

A prebuilt linux/amd64 binary is published on every push to `main`. Direct
link:

[`wx01-linux-amd64` (latest)](https://github.com/jdhuntington/wx01/releases/latest/download/wx01-linux-amd64)

```bash
# 1. Grab the binary
sudo curl -L -o /usr/local/bin/wx01 \
  https://github.com/jdhuntington/wx01/releases/latest/download/wx01-linux-amd64
sudo chmod +x /usr/local/bin/wx01

# 2. Run TimescaleDB in Docker (data persists in the named volume)
docker run -d \
  --name wx01-db \
  --restart unless-stopped \
  -p 5432:5432 \
  -v wx01-pgdata:/var/lib/postgresql/data \
  -e POSTGRES_USER=wx01 \
  -e POSTGRES_PASSWORD=wx01 \
  -e POSTGRES_DB=wx01 \
  timescale/timescaledb:latest-pg17

# 3. Drop in a systemd unit
sudo tee /etc/systemd/system/wx01.service >/dev/null <<'EOF'
[Unit]
Description=wx01 weather dashboard
After=network.target docker.service
Wants=docker.service

[Service]
Type=simple
ExecStart=/usr/local/bin/wx01
Restart=on-failure
RestartSec=5
Environment=WX01_DATABASE_URL=postgres://wx01:wx01@127.0.0.1:5432/wx01?sslmode=disable

[Install]
WantedBy=multi-user.target
EOF

# 4. Start it
sudo systemctl daemon-reload
sudo systemctl enable --now wx01
```

Open `http://<host>:3100`. If your Tempest is on the same LAN segment (UDP
broadcasts don't cross NAT), data starts arriving within ~30 seconds.

If you want this on port 80, edit the unit to add `WX01_HTTP_PORT=80` and
`AmbientCapabilities=CAP_NET_BIND_SERVICE`.

To upgrade later, re-run the `curl` from step 1 and `sudo systemctl restart
wx01`. The Postgres container keeps running on its own.

## Running it from source

You need Go 1.22+, Node 20+, and Docker.

```bash
# Terminal 1: TimescaleDB
docker run --rm -p 5432:5432 \
  -e POSTGRES_USER=wx01 -e POSTGRES_PASSWORD=wx01 -e POSTGRES_DB=wx01 \
  timescale/timescaledb:latest-pg17

# Terminal 2: the Go server (UDP :50222 + HTTP :3100)
make dev

# Terminal 3: the Vite dev server (proxies /api to :3100, with hot reload)
cd web && npm run dev
```

Open <http://localhost:5173>. If a Tempest hub is on your network, data
should start showing up within about half a minute. If you don't have one,
the dashboard will load but stay empty. There's no synthetic data generator
yet.

## Building a binary

```bash
make build           # current platform (macOS in my case)
make build-linux     # cross-compile for linux/amd64
```

The Go binary embeds the built React app, so the resulting file is the whole
thing.

## Configuration

| Env var             | Default                                                    |
| ------------------- | ---------------------------------------------------------- |
| `WX01_DATABASE_URL` | `postgres://wx01:wx01@localhost:5432/wx01?sslmode=disable` |
| `WX01_UDP_PORT`     | `50222`                                                    |
| `WX01_HTTP_PORT`    | `3100`                                                     |

## Deployment

I run wx01 inside an Incus container on a Linux box. The scripts in
[`deploy/`](./deploy/) automate that specific setup, and they're tailored to
my environment rather than meant as a general installer. If you want the
full walkthrough, or just want to see one way of wiring everything together,
[`deploy/README.md`](./deploy/README.md) has it.

Other setups are on you, but the binary is configured entirely through the
env vars above, so it should fit anywhere you can run a Go binary and reach
a Postgres database.

## Tech

- Go (stdlib HTTP, [pgx](https://github.com/jackc/pgx))
- React 19 with [lightweight-charts](https://github.com/tradingview/lightweight-charts) v5
- TimescaleDB on Postgres 17
