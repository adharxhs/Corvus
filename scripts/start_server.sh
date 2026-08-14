#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
SERVER_DIR="$ROOT_DIR/server"
ENV_FILE="${ENV_FILE:-$ROOT_DIR/.env}"
BIN="$SERVER_DIR/corvus-server"

if [ ! -f "$ENV_FILE" ]; then
    echo "Error: no environment file at $ENV_FILE. Run ./scripts/setup.sh first." >&2
    exit 1
fi

set -a
source <(grep -v '^#' "$ENV_FILE" | grep -v '^\s*$')
set +a

if [ -z "${JWT_SECRET:-}" ]; then
    echo "Error: JWT_SECRET is not set in $ENV_FILE. Run ./scripts/setup.sh first." >&2
    exit 1
fi

if [ ! -f "$BIN" ]; then
    echo "==> Server binary not found, building..."
    if ! command -v go >/dev/null 2>&1; then
        echo "Error: 'go' command not found. Please install the Go toolchain." >&2
        exit 1
    fi
    export CGO_ENABLED=0
    (cd "$SERVER_DIR" && go build -ldflags="-s -w" -o "$BIN" ./cmd/server)
fi

echo "==> Starting Corvus server on :${HTTP_PORT:-8080}"
echo "==> Database: $DATABASE_PATH"
echo "==> CORS_ORIGIN: ${CORS_ORIGIN:-*}"
cd "$SERVER_DIR"
exec "$BIN"
