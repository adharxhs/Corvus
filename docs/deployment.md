# Deployment

## Server on Android (Termux + Tailscale Funnel)

The primary deployment target is an Android phone running Termux with Tailscale for public HTTPS exposure. This gives a free stable URL with no VPS.

**Full setup walkthrough (account setup, `tailscaled` flags, Funnel, keeping it alive on Android, client wiring, troubleshooting): see [`docs/tailscale.md`](./tailscale.md).** The quick version:

### Prerequisites

- Android phone (daily-use device is fine)
- [Termux](https://f-droid.org/en/packages/com.termux/) (from F-Droid, not Play Store)
- [Tailscale](https://tailscale.com/) account (free tier supports 100 devices)

### Setup

```bash
# In Termux
pkg update && pkg install golang tailscale

# Get the server code onto the phone
# Option A: git clone from your remote
# Option B: adb push from a connected PC
#   adb push server/ /sdcard/server
#   cp -r /sdcard/server ~/server

cd server
go build -o corvus-server ./cmd/server
```

### Running

```bash
# Keep CPU awake while server is running
termux-wake-lock

# Start Tailscale in userspace networking mode
tailscaled --tun=userspace-networking &
tailscale up

# Expose port 8080 publicly over HTTPS, persistently
tailscale funnel --bg 8080
# Prints something like: https://myphone.tailnet-name.ts.net

# Run the server
JWT_SECRET="$(cat ~/.jwt_secret)" ./corvus-server
```

The `tailscale funnel` command gives a stable `https://<name>.ts.net` URL. Friends connect to this URL — they do **not** need Tailscale installed. For why `--tun=userspace-networking` and `--bg` are both required (not optional) on Android, see `docs/tailscale.md`.

### Persistent JWT_SECRET

Save the JWT secret to a file so it survives Termux restarts:

```bash
openssl rand -hex 32 > ~/.jwt_secret
# Then always start with:
JWT_SECRET="$(cat ~/.jwt_secret)" ./corvus-server
```

### Battery considerations

The Go server is lightweight, but running on a phone will drain battery. Use `termux-wake-lock` while hosting and stop the server when not needed.

## Server on Linux (development or VPS)

```bash
cd server
go build -o corvus-server ./cmd/server
JWT_SECRET=<secret> ./corvus-server
# Listens on :8080 by default
```

### Configuration

All configuration is via environment variables:

| Variable                | Required | Default       | Description                                             |
| ----------------------- | -------- | ------------- | ------------------------------------------------------- |
| `JWT_SECRET`            | Yes      | —             | Secret key for signing JWTs                             |
| `HTTP_PORT`             | No       | `8080`        | Listen port                                             |
| `DATABASE_PATH`         | No       | `corvus.db`   | SQLite database file path                               |
| `CORS_ORIGIN`           | No       | `*`           | Allowed CORS origins (comma-separated)                  |
| `JWT_EXPIRATION`        | No       | `24h`         | JWT token lifetime                                      |
| `CHAT_REQUEST_COOLDOWN` | No       | `24h`         | Cooldown before re-sending a rejected chat request      |
| `LOG_LEVEL`             | No       | `info`        | Structured log level (`debug`, `info`, `warn`, `error`) |
| `ENVIRONMENT`           | No       | `development` | Deployment context label                                |

## Exposing to the internet

### Option A: Tailscale Funnel (recommended)

Free, gives HTTPS automatically, works behind CGNAT. See the Termux section above.

### Option B: Reverse proxy (nginx, Caddy)

Put a reverse proxy in front of `localhost:8080`. Use Caddy for automatic TLS:

```
chat.yourdomain.com {
    reverse_proxy localhost:8080
}
```

### Option C: Cloudflare Tunnel

```bash
cloudflared tunnel create corvus
cloudflared tunnel route dns corvus chat.yourdomain.com
# Run: cloudflared tunnel run corvus
```

The server always binds to `:<HTTP_PORT>` on all interfaces. For local-only access behind a reverse proxy, use a firewall to block direct connections.

## Health check

```bash
curl http://localhost:8080/
# {"status":"ok"}
```
