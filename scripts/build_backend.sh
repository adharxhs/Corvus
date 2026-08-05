#!/usr/bin/env bash
set -e

# Determine directory paths
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
SERVER_DIR="$ROOT_DIR/server"

echo "==> Building Corvus Backend..."

if ! command -v go >/dev/null 2>&1; then
    echo "Error: 'go' command not found. Please install the Go toolchain." >&2
    exit 1
fi

export CGO_ENABLED=0

cd "$SERVER_DIR"
go build -ldflags="-s -w" -o "$SERVER_DIR/corvus-server" ./cmd/server

echo "==> Build successful: $SERVER_DIR/corvus-server"
