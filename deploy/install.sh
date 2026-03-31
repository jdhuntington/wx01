#!/bin/bash
# Run this ON the Incus container (Debian/Ubuntu) to install wx01.
# Prerequisites: Docker must already be installed.
#
# Usage: ./install.sh

set -euo pipefail

echo "=== wx01 installer ==="

# Check for required files in the same directory as this script
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

for f in wx01-linux-amd64 wx01.service wx01-db wx01-backup; do
  if [ ! -f "${SCRIPT_DIR}/${f}" ]; then
    echo "Missing: ${SCRIPT_DIR}/${f}" >&2
    exit 1
  fi
done

# Install binary
echo "Installing wx01 binary..."
cp "${SCRIPT_DIR}/wx01-linux-amd64" /usr/local/bin/wx01
chmod +x /usr/local/bin/wx01

# Install helper scripts
echo "Installing helper scripts..."
cp "${SCRIPT_DIR}/wx01-db" /usr/local/bin/wx01-db
chmod +x /usr/local/bin/wx01-db
cp "${SCRIPT_DIR}/wx01-backup" /usr/local/bin/wx01-backup
chmod +x /usr/local/bin/wx01-backup

# Create data directory
mkdir -p /var/lib/wx01/pgdata
mkdir -p /var/lib/wx01/backups

# Install systemd unit
echo "Installing systemd service..."
cp "${SCRIPT_DIR}/wx01.service" /etc/systemd/system/wx01.service
systemctl daemon-reload
systemctl enable wx01

# Pull the TimescaleDB image
echo "Pulling TimescaleDB image..."
docker pull timescale/timescaledb:latest-pg17

# Start the service
echo "Starting wx01..."
systemctl start wx01

echo ""
echo "=== wx01 installed ==="
echo "  Dashboard:  http://$(hostname -I | awk '{print $1}'):3100"
echo "  Status:     systemctl status wx01"
echo "  Logs:       journalctl -u wx01 -f"
echo "  Backup:     wx01-backup"
echo ""
