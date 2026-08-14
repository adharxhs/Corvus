#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
CLIENT_DIR="$ROOT_DIR/client"

export WEBKIT_DISABLE_DMABUF_RENDERER="${WEBKIT_DISABLE_DMABUF_RENDERER:-1}"

cd "$CLIENT_DIR"

if [ ! -d node_modules ]; then
    echo "==> Installing client dependencies..."
    npm install
fi

echo "==> Starting Corvus client dev server"
echo "==> Vite: http://localhost:1420"
exec npm run dev
