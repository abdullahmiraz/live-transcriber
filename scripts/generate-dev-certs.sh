#!/usr/bin/env bash
# Generate self-signed TLS certs for LAN/mobile dev (camera/mic require HTTPS on non-localhost).
# Usage: bash scripts/generate-dev-certs.sh [LAN_IP]
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
CERT_DIR="$ROOT/infra/nginx/certs"
LAN_IP="${1:-${LAN_IP:-}}"

if [[ -z "$LAN_IP" ]]; then
  if command -v ipconfig >/dev/null 2>&1; then
    # Prefer a typical home Wi‑Fi adapter (192.168.x.x or 10.x.x.x), not Docker/WSL virtual NICs.
    LAN_IP=$(ipconfig 2>/dev/null | tr -d '\r' | grep -a "IPv4" | sed -E 's/.*: *//' | grep -E '^(192\.168\.|10\.)' | head -1 || true)
    if [[ -z "$LAN_IP" ]]; then
      LAN_IP=$(ipconfig 2>/dev/null | tr -d '\r' | grep -a "IPv4" | head -1 | sed -E 's/.*: *//')
    fi
  elif command -v hostname >/dev/null 2>&1; then
    LAN_IP=$(hostname -I 2>/dev/null | awk '{print $1}')
  fi
fi

mkdir -p "$CERT_DIR"
cd "$CERT_DIR"

SAN="DNS:localhost,DNS:*.local,IP:127.0.0.1"
if [[ -n "$LAN_IP" && "$LAN_IP" != "127.0.0.1" ]]; then
  SAN="${SAN},IP:${LAN_IP}"
  echo "Including LAN IP in certificate: $LAN_IP"
fi

# MSYS_NO_PATHCONV avoids Git Bash rewriting /CN=... and absolute paths for OpenSSL on Windows.
MSYS_NO_PATHCONV=1 openssl req -x509 -nodes -days 825 -newkey rsa:2048 \
  -keyout dev.key \
  -out dev.crt \
  -subj "/CN=live-meet-dev/O=LiveMeet/C=US" \
  -addext "subjectAltName=${SAN}"

echo ""
echo "Wrote $CERT_DIR/dev.crt and dev.key"
echo "Phone/tablet (camera/mic): https://${LAN_IP:-YOUR_LAN_IP}/"
echo "This PC:                   http://localhost/  or  https://localhost/ (accept cert warning)"
