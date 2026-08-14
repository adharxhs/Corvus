#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
SERVER_DIR="$ROOT_DIR/server"
CLIENT_DIR="$ROOT_DIR/client"
ENV_FILE="$ROOT_DIR/.env"
CLIENT_ENV_FILE="$CLIENT_DIR/.env.local"

echo "=============================================="
echo "  Corvus Server Setup + Release APK Build    "
echo "=============================================="

if ! command -v go >/dev/null 2>&1; then
    echo "Error: 'go' command not found. Please install the Go toolchain." >&2
    exit 1
fi

# 1. Environment file (existing data is preserved)
if [ -f "$ENV_FILE" ]; then
    echo "==> Using existing environment at $ENV_FILE"
    if ! grep -q '^JWT_SECRET=' "$ENV_FILE"; then
        echo "Error: JWT_SECRET is missing from $ENV_FILE." >&2
        exit 1
    fi
else
    echo "==> Creating .env with a fresh JWT secret..."
    JWT_SECRET_GEN="$(head -c 32 /dev/urandom | base64 | tr -dc 'a-zA-Z0-9' | head -c 32)"
    cat <<EOF > "$ENV_FILE"
HTTP_PORT=8080
DATABASE_PATH=$SERVER_DIR/corvus.db
JWT_SECRET=$JWT_SECRET_GEN
JWT_EXPIRATION=24h
CHAT_REQUEST_COOLDOWN=24h
LOG_LEVEL=info
ENVIRONMENT=production
CORS_ORIGIN=*
EOF
    echo "==> Wrote $ENV_FILE"
fi

# 2. Tailscale server address -> client/.env.local
echo
echo "Enter the Tailscale IP or hostname of this server."
echo "Examples: 100.101.102.103  |  my-device.tailnet.ts.net  |  https://my-device.tailnet.ts.net"
read -rp "Server address: " SERVER_ADDR
SERVER_ADDR="$(printf '%s' "${SERVER_ADDR:-}" | sed 's/[[:space:]]*$//')"
if [ -z "$SERVER_ADDR" ]; then
    echo "Error: no server address provided." >&2
    exit 1
fi
case "$SERVER_ADDR" in
    http://*|https://*) VITE_SERVER_URL="${SERVER_ADDR%/}" ;;
    *) VITE_SERVER_URL="https://${SERVER_ADDR%/}" ;;
esac
printf 'VITE_SERVER_URL=%s\n' "$VITE_SERVER_URL" > "$CLIENT_ENV_FILE"
echo "==> Wrote $CLIENT_ENV_FILE -> VITE_SERVER_URL=$VITE_SERVER_URL"

# 3. Build the server binary
echo "==> Building server binary..."
export CGO_ENABLED=0
(cd "$SERVER_DIR" && go build -ldflags="-s -w" -o "$SERVER_DIR/corvus-server" ./cmd/server)

# 4. Client dependencies
if [ ! -d "$CLIENT_DIR/node_modules" ]; then
    echo "==> Installing client dependencies..."
    (cd "$CLIENT_DIR" && npm install)
fi

# 5. Production release APK (never debug) -> single Corvus.apk
if [ -z "${ANDROID_HOME:-}" ]; then
    echo "Warning: ANDROID_HOME is not set; the Android build may fail without it." >&2
fi
echo "==> Building production release APK (universal, all ABIs)..."
(cd "$CLIENT_DIR" && npm run tauri android build -- --apk --ci)
APK="$CLIENT_DIR/src-tauri/gen/android/app/build/outputs/apk/universal/release/app-universal-release.apk"
if [ ! -f "$APK" ]; then
    echo "Error: expected APK not found at $APK" >&2
    exit 1
fi
cp "$APK" "$CLIENT_DIR/Corvus.apk"
echo "==> APK: $CLIENT_DIR/Corvus.apk"

# 6. Hand off to the start script (the one that actually runs the server)
echo "==> Setup complete. Starting the server..."
exec "$SCRIPT_DIR/start_server.sh"
