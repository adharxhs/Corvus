#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
SERVER_DIR="$ROOT_DIR/server"
ENV_FILE="${ENV_FILE:-$ROOT_DIR/.env}"

if [ -f "$ENV_FILE" ]; then
    set -a
    source <(grep -v '^#' "$ENV_FILE" | grep -v '^\s*$')
    set +a
fi

export HTTP_PORT="${HTTP_PORT:-8080}"
export DATABASE_PATH="${DATABASE_PATH:-$SERVER_DIR/corvus.db}"
export JWT_EXPIRATION="${JWT_EXPIRATION:-24h}"
export LOG_LEVEL="${LOG_LEVEL:-debug}"
export ENVIRONMENT="${ENVIRONMENT:-development}"
export CORS_ORIGIN="${CORS_ORIGIN:-*}"

if [ -z "${JWT_SECRET:-}" ]; then
    export JWT_SECRET="$(head -c 32 /dev/urandom | base64 | tr -dc 'a-zA-Z0-9' | head -c 32)"
    printf '\nJWT_SECRET=%s\n' "$JWT_SECRET" >> "$ENV_FILE"
fi

echo "==> Starting Corvus dev server on :$HTTP_PORT"
echo "==> Database: $DATABASE_PATH"
echo "==> CORS_ORIGIN: $CORS_ORIGIN"
cd "$SERVER_DIR"
exec go run ./cmd/server
