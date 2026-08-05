#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
SERVER_DIR="$ROOT_DIR/server"
ENV_FILE="$ROOT_DIR/.env"

echo "=========================================="
echo "      Corvus Backend Deployment Setup     "
echo "=========================================="

# 1. Environment Detection
IS_TERMUX=false
if [ -d "/data/data/com.termux" ] || [ -n "$TERMUX_VERSION" ]; then
    IS_TERMUX=true
    echo "Detected Environment: Android (Termux)"
else
    echo "Detected Environment: Generic Linux ($(uname -m))"
fi

# 2. Check for Go
if ! command -v go >/dev/null 2>&1; then
    echo "Go toolchain not found."
    if [ "$IS_TERMUX" = true ]; then
        echo "Installing Go via Termux package manager..."
        pkg update && pkg install -y golang
    else
        echo "Please install Go (1.22+) using your package manager:"
        echo "  Fedora:  sudo dnf install golang"
        echo "  Debian/Ubuntu: sudo apt install golang-go"
        echo "  Arch:    sudo pacman -S go"
        exit 1
    fi
fi

# 3. Create .env file if missing
if [ ! -f "$ENV_FILE" ]; then
    echo "==> Creating default production .env configuration..."
    JWT_SECRET_GEN=$(head -c 32 /dev/urandom | base64 | tr -dc 'a-zA-Z0-9' | head -c 32)
    cat <<EOF > "$ENV_FILE"
HTTP_PORT=8080
DATABASE_PATH=$SERVER_DIR/corvus.db
JWT_SECRET=$JWT_SECRET_GEN
JWT_EXPIRATION=24h
LOG_LEVEL=info
ENVIRONMENT=production
EOF
    echo "Generated $ENV_FILE with random JWT secret."
else
    echo "==> Existing configuration found at $ENV_FILE"
fi

# 4. Build binary
echo "==> Building server binary (CGO-free)..."
"$SCRIPT_DIR/build_backend.sh"

echo "=========================================="
echo "Deployment Ready!"
echo "=========================================="
echo "To start the backend, run:"
echo "  $SCRIPT_DIR/run_backend.sh"
echo ""
echo "Or run in background:"
echo "  nohup $SCRIPT_DIR/run_backend.sh > corvus.log 2>&1 &"
echo "=========================================="
