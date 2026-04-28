# wx01

Lightweight dashboard for a WeatherFlow Tempest. Single Go binary with the React
frontend embedded, backed by TimescaleDB. The Tempest UDP-broadcasts to the
local network on port 50222; this daemon listens, parses, stores, and serves a
real-time dashboard over HTTP.

For architecture, file map, API endpoints, and the real-time data flow, see
[CLAUDE.md](./CLAUDE.md).

---

## Hey, here's how we think we deployed this bad boy

I run wx01 inside an Incus container on a separate Linux box on the same LAN as
the Tempest hub. Here's the actual shape of the deploy, end to end. **None of
this is opinion — it's what `make deploy` and `deploy/install.sh` actually do.**

### The big picture

```
                      ┌─────────────────────── Incus container ──────────────────────┐
                      │                                                              │
   Tempest hub        │   systemd unit "wx01"                                        │
   (broadcasts        │   ├── ExecStartPre: /usr/local/bin/wx01-db start             │
   on LAN) ──UDP:50222┼──▶│   │            (launches Docker container "wx01-db",     │
                      │   │             TimescaleDB on :5433, --network=host)        │
                      │   └── ExecStart:   /usr/local/bin/wx01                       │
                      │                    (Go binary, listens UDP:50222 + HTTP:80,  │
                      │                     embeds the React SPA, talks to PG :5433) │
                      │                                                              │
                      └──────────────────────────────────────────────────────────────┘
                                              ▲
                                              │  HTTP :80
                                              │
                                          your browser
```

Two things run inside the container:

1. **A Docker container** named `wx01-db` running TimescaleDB (Postgres 17).
   - Uses `--network=host` because Docker's default bridge needs sysctls that
     unprivileged Incus containers can't set.
   - Listens on port **5433** (not the default 5432) to keep out of the way of
     anything else on the host's network namespace.
   - Data lives at `/var/lib/wx01/pgdata` on the container's filesystem.
2. **The wx01 Go binary** at `/usr/local/bin/wx01`, run by systemd.
   - Listens on UDP **50222** (Tempest broadcasts) and HTTP **80**.
   - Binds :80 via `AmbientCapabilities=CAP_NET_BIND_SERVICE` — no root needed.
   - Connects to TimescaleDB at `127.0.0.1:5433` (host networking, remember).

### What's actually on the deploy host

After `./deploy/install.sh user@host` finishes, the remote looks like this:

| Path                                    | What it is                                          |
|-----------------------------------------|-----------------------------------------------------|
| `/usr/local/bin/wx01`                   | The Go binary (with embedded SPA)                   |
| `/usr/local/bin/wx01-db`                | start/stop/status helper for the Docker container   |
| `/usr/local/bin/wx01-backup`            | `pg_dump` → gzipped SQL, retains last 30            |
| `/etc/systemd/system/wx01.service`      | The systemd unit                                    |
| `/var/lib/wx01/pgdata/`                 | TimescaleDB data directory (mounted into Docker)    |
| `/var/lib/wx01/backups/`                | Where `wx01-backup` writes dumps                    |
| Docker container `wx01-db`              | TimescaleDB, `--restart unless-stopped`             |

### Prereqs on the deploy host

The container/VM that wx01 runs in needs:

- **Linux amd64** (the binary is cross-compiled for that target).
- **Docker** installed and working. Inside Incus this means the container
  profile needs `security.nesting=true` and (for unprivileged containers)
  `security.privileged=true` *or* the right kernel modules + cgroup access.
  If `docker run hello-world` works, you're fine.
- **systemd** (any modern distro).
- **Root SSH key auth** — the install script SSHes in and writes to
  `/usr/local/bin` and `/etc/systemd/system`.
- **UDP :50222 reachable on the same L2 segment as the Tempest hub.** Tempest
  broadcasts; if the container is on a bridged interface that sees the hub's
  broadcasts, you're good. NAT'd networks won't work — broadcasts don't
  cross NAT.
- **HTTP :80 reachable** from wherever you want to look at the dashboard. If
  you'd rather use a different port, edit `WX01_HTTP_PORT` in
  `deploy/wx01.service` and drop the `CAP_NET_BIND_SERVICE` line.

### Building & deploying

From the dev machine (Mac):

```bash
make deploy TARGET=root@wx01-host
```

That:

1. Builds the React frontend (`cd web && npm run build`).
2. Copies `web/dist` → `cmd/wx01/dist` so `//go:embed` can pick it up.
3. Cross-compiles the Go binary for `linux/amd64`.
4. Runs `deploy/install.sh root@wx01-host`, which:
   - `scp`s the binary + helper scripts + service file to `/tmp/wx01-deploy/`.
   - Stops the existing service (if any).
   - Installs the binary and helpers to `/usr/local/bin`.
   - Creates `/var/lib/wx01/{pgdata,backups}`.
   - Installs and enables the systemd unit.
   - `docker pull`s the TimescaleDB image (skipped if already present).
   - Starts the service.

The first start takes longer because TimescaleDB initializes the data
directory. Subsequent restarts are quick.

### Day-to-day operations

On the deploy host:

```bash
systemctl status wx01           # is it running?
journalctl -u wx01 -f           # tail the logs
systemctl restart wx01          # restart everything (db container stays up)

wx01-db status                  # is the postgres container alive?
wx01-db stop                    # stop just the db
wx01-db start                   # start it again

wx01-backup                     # write a fresh dump to /var/lib/wx01/backups
ls -lh /var/lib/wx01/backups    # see what we've got (30 retained)
```

To upgrade: just run `make deploy TARGET=...` again. The install script stops
the service, replaces the binary, and restarts. Postgres keeps running on its
own; the data on disk is untouched.

### Rolling back / disaster recovery

The data directory `/var/lib/wx01/pgdata` is the source of truth — if you
nuke the container and recreate it pointing at the same directory, you get
your data back. To restore from a backup:

```bash
gunzip -c /var/lib/wx01/backups/wx01-YYYYMMDD-HHMMSS.sql.gz \
  | docker exec -i wx01-db psql -U wx01 -d wx01
```

### Things to double-check if it's not working

- **No data showing up?** `journalctl -u wx01 -f` and look for "received UDP
  packet" lines. If silence, the broadcast isn't reaching the container —
  check the network mode of the Incus container (bridged, not NAT'd).
- **Can't connect to the dashboard?** `curl http://localhost` from inside the
  container first to rule out the network. Then check Incus port forwarding /
  bridging from outside.
- **`wx01-db` won't start inside Incus?** Almost always Docker permissions.
  Try `docker run --rm hello-world` from inside the container — if that
  fails, fix Docker before worrying about wx01.

### Local dev (no deploy required)

```bash
# Terminal 1: a local TimescaleDB
docker run --rm -p 5432:5432 \
  -e POSTGRES_USER=wx01 -e POSTGRES_PASSWORD=wx01 -e POSTGRES_DB=wx01 \
  timescale/timescaledb:latest-pg17

# Terminal 2: the Go server
make dev

# Terminal 3: the Vite dev server (proxies /api → :3100)
cd web && npm run dev
```
