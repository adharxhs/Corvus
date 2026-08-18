# Corvus

An end-to-end encrypted messaging platform with a Go backend, Tauri 2 client, and a Rust crypto core.

## Status

All core subsystems are implemented, tested, and integrated. The project is in the packaging/deployment phase.

- **Backend** (Go): fully implemented — auth, groups (create/get/rename with names, invites, membership), chat requests, relationships, profile pictures, prekey bundles, presence, WebSocket message dispatch, member-join broadcasts.
- **Client** (Tauri 2 + React 19): fully implemented — auth, chat UI with persistent header and profile pics on every message, groups with names and rename UI, contacts, settings/themes, profile picture picker, protocol layer, WebSocket service with reconnection, participants list with avatars/presence in group settings. **E2EE wired**: all direct and group messages are encrypted via the Rust crypto core before sending and decrypted on receipt.
- **Crypto** (Rust crate): fully implemented — X3DH, Double Ratchet, Sender Keys, identity/prekey lifecycle, serialization, storage trait. **Linked to Tauri client** via `crypto_manager.rs` + `commands.rs` (15 IPC commands) and `services/crypto.ts` (TypeScript invoke wrappers).
- **Android APK**: signed release build configured, single `Corvus.apk` output. Includes the full Rust crypto stack (X3DH, Double Ratchet, Sender Keys) compiled into the native binary.
- **Server deployment**: runs on Termux via Tailscale Funnel (free HTTPS, no VPS required).

## Architecture

```
┌──────────────────────────────────────┐
│           Client (Tauri 2)           │
│  React 19 UI  ←→  Rust crypto core   │
│  (no plaintext in TS)                │
└──────────────┬───────────────────────┘
               │ HTTP + WebSocket
┌──────────────▼───────────────────────┐
│          Go Backend (server/)         │
│  Auth, routing, WS, SQLite           │
│  (no crypto, no plaintext)           │
└──────────────────────────────────────┘
```

## Tech Stack

- **Client**: Tauri 2, React 19, TypeScript, Vite 7
- **Backend**: Go 1.26, net/http, modernc.org/sqlite
- **Crypto**: Rust (x25519-dalek, ed25519-dalek, double-ratchet-2, aes-gcm)
- **Deployment**: Termux + Tailscale Funnel

## Repository Structure

```
corvus/
├── client/                 # Tauri 2 + React
│   ├── src/                # React components, services, protocol, websocket
│   │   └── services/crypto.ts  # TypeScript invoke wrappers for Rust crypto
│   └── src-tauri/          # Rust core
│       └── src/
│           ├── lib.rs              # Tauri command registration + state
│           ├── crypto_manager.rs   # Crypto state wrapper (identity, sessions, sender keys)
│           └── commands.rs         # 15 Tauri IPC commands for E2EE
├── server/                 # Go backend module
│   ├── api/                # HTTP handlers + router
│   ├── auth/               # JWT, password hashing
│   ├── websocket/          # WS server, registry, dispatcher, presence
│   ├── database/           # SQLite, migrations
│   ├── services/           # Business logic
│   ├── repository/         # DB access
│   ├── models/             # Data structures
│   ├── protocol/           # Wire protocol
│   ├── middleware/         # CORS, auth, recovery
│   ├── config/             # Env-based configuration
│   ├── logging/            # Structured logging
│   └── cmd/server/main.go  # Entry point
├── crypto/                 # Rust crypto library
│   └── src/                # x3dh, double_ratchet, sender_keys, identity, prekeys
└── shared/                 # Protocol docs/schemas (documentation only)
```

## Quick Start

### Client (dev)

```bash
cd client
npm install
npm run dev          # Vite on port 1420
```

### Full Tauri app

```bash
cd client
npm run tauri dev    # WEBKIT_DISABLE_DMABUF_RENDERER=1 needed on Linux
```

### Backend

```bash
cd server
go build ./cmd/server
JWT_SECRET=<secret> go run ./cmd/server    # listens on :8080
```

### Server + APK setup

One-time setup: creates `.env` if missing, asks for the Tailscale IP/hostname, writes `client/.env.local`, builds the production release APK as `client/Corvus.apk`, then starts the server.

```bash
./scripts/setup.sh
```

Start the server with existing data (loads `.env`):

```bash
./scripts/start_server.sh
```

### Crypto tests

```bash
cd crypto
cargo test
```

### Android APK

Built automatically by `./scripts/setup.sh` (asks for the Tailscale address, then builds a signed production release):

```bash
cd client
VITE_SERVER_URL=https://<your-name>.<tailnet>.ts.net \
  npm run tauri android build -- --apk --ci
```

Signed APK: `client/Corvus.apk` (universal, all ABIs)

### Server on Android (Termux)

```bash
pkg install golang tailscale
tailscaled --tun=userspace-networking &
tailscale up
tailscale funnel 8080    # gives stable HTTPS URL
JWT_SECRET=<saved-secret> ./corvus-server
```

## Environment Variables

| Variable | Required | Default | Purpose |
|---|---|---|---|
| `JWT_SECRET` | Yes | — | Signing key for JWTs |
| `HTTP_PORT` | No | `8080` | Server listen port |
| `DATABASE_PATH` | No | `corvus.db` | SQLite file path |
| `CORS_ORIGIN` | No | `*` | Allowed origins |
| `CHAT_REQUEST_COOLDOWN` | No | `24h` | Cooldown after re-request |
| `LOG_LEVEL` | No | `info` | Logging verbosity |
| `ENVIRONMENT` | No | `development` | Deployment context |
| `VITE_SERVER_URL` | No | `http://localhost:8080` | Server URL baked into APK at build time |

## License

Apache License 2.0. See [LICENSE](LICENSE).
