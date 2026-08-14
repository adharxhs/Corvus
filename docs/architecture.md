# Architecture

Corvus is an end-to-end encrypted messaging platform with three independent subsystems: a Tauri 2 client, a Go HTTP backend, and a Rust cryptography crate. The server never sees plaintext or private keys. All encryption happens client-side in Rust.

```
┌──────────────────────────────────────┐
│           Client (Tauri 2)           │
│  React 19 UI  ←→  Rust crypto core   │
│  (no plaintext in TypeScript)        │
└──────────────┬───────────────────────┘
               │ HTTP + WebSocket (TLS via tunnel)
┌──────────────▼───────────────────────┐
│          Go Backend (server/)         │
│  Auth, routing, WS, SQLite           │
│  (no crypto, no plaintext)           │
└──────────────────────────────────────┘
```

## Components

### Client (`client/`)

- **Framework**: Tauri 2 (Rust shell + system WebView)
- **Frontend**: React 19, TypeScript, Vite 7, React Router
- **Rust core**: `src-tauri/src/lib.rs` exposes Tauri commands for profile-picture encrypt/decrypt; the `crypto/` crate is the shared cryptographic library
- **Key directories**:
  - `src/services/` — HTTP client, auth, groups, relationships, profile pictures, prekeys
  - `src/websocket/` — WebSocket service with exponential-backoff reconnection
  - `src/protocol/` — envelope encoding, validation, parser (wire format `{version, type, payload}`)
  - `src/contexts/` — React context providers (auth, chat, websocket, theme, app)
  - `src/pages/` — route-level views (login, register, chats, contacts, groups, settings)
- **Build-time configuration**: `VITE_SERVER_URL` env var sets the API/WS base URL at build time (used in `src/utils/constants.ts`). Falls back to `http://localhost:8080`.

### Backend (`server/`)

- **Language**: Go 1.26
- **Database**: SQLite via `modernc.org/sqlite` (pure-Go, no CGO)
- **Auth**: register/login with Argon2id password hashing, JWT tokens
- **WebSocket**: `github.com/coder/websocket`; per-connection read/write loops, offline message queue, presence broadcast
- **Layers** (enforced import boundaries):
  ```
  api/ → services/ → repository/ → database/
  ```
- **Key packages**:
  - `api/` — HTTP handlers, router, response formatting
  - `auth/` — JWT generation/validation, password hashing, context helpers
  - `websocket/` — Server, Client, Registry (online tracking), Dispatcher (message routing), presence
  - `database/` — Connection, migrations (SQL files embedded at compile time)
  - `services/` — Business logic for users, groups, relationships, prekeys, profile pictures
  - `repository/` — SQL queries for each domain
  - `protocol/` — Wire envelope format, type constants, serialization, validation
  - `middleware/` — CORS, JWT auth, request logging, panic recovery
  - `config/` — Environment-based configuration loading

### Crypto crate (`crypto/`)

- **Language**: Rust, edition 2021
- **Modules**: `x3dh`, `double_ratchet`, `sender_keys`, `identity`, `prekeys`, `session`, `serialization`, `storage`, `random`, `errors`, `util`
- **Dependencies**: `x25519-dalek` 2.0.1, `ed25519-dalek` 2.1.1, `double-ratchet-2` (pinned commit), `aes-gcm` 0.10.3, `hkdf` 0.12, `sha2` 0.10, `serde` + `bincode`
- **Storage abstraction**: `Store` trait with `InMemoryStore` implementation; real persistence provided by the Tauri client at runtime
- **Never depends on**: Tauri, serde_json, or any I/O beyond the `Store` trait

## Wire protocol

Every message exchanged between client and server uses a JSON envelope:

```json
{ "version": 1, "type": "<message_type>", "payload": { ... } }
```

Server→client types: `direct_message`, `group_message`, `sender_key_distribution`, `profile_picture_updated`, `presence_snapshot`, `presence`.

Client→server types: `direct_message`, `group_message`, `sender_key_distribution`, `profile_picture_updated`.

## Encryption model

| Layer | What it encrypts | Where it runs |
|---|---|---|
| X3DH + Double Ratchet | Direct messages, key material | `crypto/` crate (client-side Rust) |
| Sender Keys | Group messages | `crypto/` crate (client-side Rust) |
| Profile key (AES-256-GCM) | Profile pictures | `src-tauri/src/profile_key.rs` |

The server stores only ciphertext. Private keys never leave the client device.

## Database

SQLite with versioned SQL migrations applied at server startup via `schema_migrations` tracking table. Migrations are embedded in the binary via `embed.FS`.

Key tables: `users`, `relationships`, `groups`, `group_members`, `group_invites`, `messages`, `pending_messages`, `sender_key_distributions`, `prekey_bundles`, `profile_pictures`.
