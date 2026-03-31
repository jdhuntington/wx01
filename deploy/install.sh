#!/bin/bash
# Deploy wx01 to a remote server over SSH.
# Run from your dev machine (Mac).
#
# Usage: ./install.sh user@host
#    or: ./install.sh user host
#
# Prerequisites on the remote:
#   - Docker installed
#   - SSH key auth configured (as root)

set -euo pipefail

if [ $# -eq 1 ]; then
  REMOTE="$1"
elif [ $# -eq 2 ]; then
  REMOTE="${1}@${2}"
else
  echo "Usage: $0 user@host" >&2
  echo "       $0 user host" >&2
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# Verify all files are present locally
for f in wx01-linux-amd64 wx01.service wx01-db wx01-backup; do
  if [ ! -f "${SCRIPT_DIR}/${f}" ]; then
    echo "Missing: ${SCRIPT_DIR}/${f}" >&2
    echo "Run 'make deploy' first." >&2
    exit 1
  fi
done

echo "=== Deploying wx01 to ${REMOTE} ==="

# Copy files to a temp directory on the remote
echo "Uploading files..."
ssh "$REMOTE" "mkdir -p /tmp/wx01-deploy"
scp -q \
  "${SCRIPT_DIR}/wx01-linux-amd64" \
  "${SCRIPT_DIR}/wx01.service" \
  "${SCRIPT_DIR}/wx01-db" \
  "${SCRIPT_DIR}/wx01-backup" \
  "${REMOTE}:/tmp/wx01-deploy/"

# Run the install on the remote
echo "Installing..."
ssh "$REMOTE" 'bash -s' <<'REMOTE_SCRIPT'
set -euo pipefail
SRC=/tmp/wx01-deploy

# Stop existing service if running
systemctl stop wx01 2>/dev/null || true

# Install binary
cp "${SRC}/wx01-linux-amd64" /usr/local/bin/wx01
chmod +x /usr/local/bin/wx01

# Install helper scripts
cp "${SRC}/wx01-db" /usr/local/bin/wx01-db
chmod +x /usr/local/bin/wx01-db
cp "${SRC}/wx01-backup" /usr/local/bin/wx01-backup
chmod +x /usr/local/bin/wx01-backup

# Data directories
mkdir -p /var/lib/wx01/pgdata
mkdir -p /var/lib/wx01/backups

# Systemd
cp "${SRC}/wx01.service" /etc/systemd/system/wx01.service
systemctl daemon-reload
systemctl enable wx01

# Pull TimescaleDB image (skip if already present)
if ! docker image inspect timescale/timescaledb:latest-pg17 &>/dev/null; then
  echo "Pulling TimescaleDB image..."
  docker pull timescale/timescaledb:latest-pg17
fi

# Start
systemctl start wx01

# Clean up
rm -rf /tmp/wx01-deploy

echo ""
echo "=== wx01 deployed ==="
echo "  Dashboard:  http://$(hostname -I | awk '{print $1}')"
echo "  Status:     systemctl status wx01"
echo "  Logs:       journalctl -u wx01 -f"
echo "  Backup:     wx01-backup"
REMOTE_SCRIPT

echo ""
echo "Done. Dashboard at http://${REMOTE#*@}"
