#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
SERVER_DIR="$ROOT_DIR/server"
ENV_FILE="${ENV_FILE:-$ROOT_DIR/.env}"

if [ -f "$ENV_FILE" ]; then
    echo "==> Loading environment from $ENV_FILE"
    set -a
    # Load env file while ignoring comment lines
    source <(grep -v '^#' "$ENV_FILE" | grep -v '^\s*$')
    set +a
fi

if [ ! -f "$SERVER_DIR/corvus-server" ]; then
    echo "==> Binary not found. Running build script..."
    "$SCRIPT_DIR/build_backend.sh"
fi

if [ -z "$JWT_SECRET" ]; then
    echo "Error: JWT_SECRET environment variable is not set." >&2
    exit 1
fi

export CORS_ORIGIN="${CORS_ORIGIN:-*}"

echo "==> Starting Corvus Server..."
echo "==> CORS_ORIGIN: $CORS_ORIGIN"
exec "$SERVER_DIR/corvus-server"
