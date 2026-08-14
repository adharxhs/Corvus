#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
CLIENT_DIR="$ROOT_DIR/client"

export WEBKIT_DISABLE_DMABUF_RENDERER="${WEBKIT_DISABLE_DMABUF_RENDERER:-1}"

if command -v adb >/dev/null 2>&1 && adb devices 2>/dev/null | grep -q 'device$'; then
    echo "==> Setting up ADB reverse port forwarding (phone:8080 -> desktop:8080)"
    adb reverse tcp:8080 tcp:8080 2>/dev/null || true
    echo "==> Setting up ADB reverse port forwarding (phone:1420 -> desktop:1420)"
    adb reverse tcp:1420 tcp:1420 2>/dev/null || true
    echo "==> Setting up ADB reverse port forwarding (phone:1421 -> desktop:1421)"
    adb reverse tcp:1421 tcp:1421 2>/dev/null || true
else
    echo "==> No ADB device found; skipping port forwarding"
fi

cd "$CLIENT_DIR"

if [ ! -d node_modules ]; then
    echo "==> Installing client dependencies..."
    npm install
fi

echo "==> Starting Corvus Tauri app"
TAURI_DEV_HOST="${TAURI_DEV_HOST:-localhost}"
exec npm run tauri android dev
