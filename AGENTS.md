# AGENTS.md

End-to-end encrypted messaging platform with three components: client (Tauri 2 + React), Go backend, and Rust crypto core.

## Layout and current state

- `client/` — Tauri 2 + React 19 + TypeScript + Vite 7. Fully implemented: auth, chat UI, groups, contacts, settings/themes, profile pictures, protocol layer, WebSocket service with reconnection.
- `server/` — Go 1.26 module named `server`. Full implementation: auth (register/login/JWT), prekey bundle storage, WebSocket server with message dispatch, group management, relationships (chat requests), profile pictures, presence, and database migrations. All via SQLite (modernc.org/sqlite).
- `crypto/` — Rust crate implementing X3DH, Double Ratchet, Sender Keys, identity/prekey lifecycle, serialization, and an `InMemoryStore` trait boundary. Fully implemented; all 12 tests pass.
- `shared/` — planned protocol docs/schemas only; must stay documentation-only.
- CI, linter: not yet set up in repo.

## Commands

- Client dev server: `cd client && npm run dev` — Vite runs on **fixed port 1420** (strictPort; change `tauri.conf.json` `devUrl` if it moves).
- Full Tauri app: `cd client && npm run tauri dev`
- Client typecheck/build: `npm run build` (= `tsc && vite build`). No test or lint scripts exist.
- Backend: `cd server && go build ./...` / `go test ./...`
- Crypto: `cd crypto && cargo check` / `cargo test`

## API surface (HTTP)

All authenticated routes require `Authorization: Bearer <JWT>`.

| Method | Path | Purpose |
|---|---|---|
| POST | `/register` | Create account |
| POST | `/login` | Get JWT |
| GET | `/users/by-username/{username}` | Exact-username lookup (returns `{"id": "..."}` or 404) |
| GET | `/users/{id}` | ID→username resolution (contact list) |
| POST | `/chat-request` | Send chat request (`{"recipient_id": "..."}`) |
| GET | `/chat-requests` | List incoming pending requests |
| POST | `/chat-request/{requester_id}/accept` | Accept a request |
| POST | `/chat-request/{requester_id}/reject` | Reject a request (silent) |
| POST | `/groups` | Create group (auto-adds creator as member) |
| GET | `/groups/invites` | List pending group invites for authed user |
| GET | `/groups/{group_id}/members` | List group members |
| POST | `/groups/{group_id}/invite` | Invite user (requires accepted personal relationship + group membership) |
| POST | `/groups/{group_id}/invite/accept` | Accept group invite → become member |
| DELETE | `/groups/{group_id}/member` | Leave group unilaterally |
| POST | `/groups/{group_id}/profile-picture` | Upload group profile picture (must be member; newer version) |
| GET | `/groups/{group_id}/profile-picture` | Get group profile picture (must be member) |
| POST | `/prekey` | Upsert own prekey bundle |
| GET | `/prekey/{id}` | Fetch a user's prekey bundle (ungated) |
| GET | `/profile-picture/{id}` | Get encrypted profile picture (requires accepted relationship) |
| POST | `/profile-picture` | Upload encrypted profile picture (must be newer version) |
| GET | `/ws` | WebSocket connection |

Server-originated WS control messages (not in HTTP surface):

| Type | Payload | When |
|---|---|---|
| `presence_snapshot` | `{"online": [user_id, ...]}` | Immediately after a client connects; lists online accepted contacts |
| `presence` | `{"user_id": "...", "status": "online"\|"offline"}` | Broadcast to online accepted contacts on every connect/disconnect |
| `group_profile_picture_updated` | `{"group_id": "...", "version": N}` | Broadcast to online group members when a group picture is uploaded |

## Android packaging

- Release APK build: `cd client && VITE_SERVER_URL=https://<server-host> npm run tauri android build -- --apk --split-per-abi --ci`
- Signed APKs land in `client/src-tauri/gen/android/app/build/outputs/apk/{arm64,arm,x86,x86_64}/release/`. Universal build (all ABIs): omit `--split-per-abi` → 36 MB; per-ABI ~12 MB.
- Signing keystore: `client/src-tauri/keystore/corvus-release.keystore` (gitignored; properties in `keystore.properties`). Must keep the same keystore for upgrades over an installed app.
- Server URL baked in at build time via `VITE_SERVER_URL`. The client derives both HTTP and WS URLs from this single value (`https://` → `wss://` automatically). Falls back to `http://localhost:8080` if unset.
- Release builds disable `usesCleartextTraffic` by default — HTTPS endpoints required.

## Server deployment (Termux + Tailscale Funnel)

- Go server runs on Android via Termux: `pkg install golang tailscale`.
- Expose publicly: `tailscaled --tun=userspace-networking &` → `tailscale up` → `tailscale funnel 8080` → stable `https://<name>.ts.net` URL, no VPS needed.
- Keep `JWT_SECRET` persistent across restarts (save to a file or use `termux-wake-lock`).
- SQLite is fine for 100 concurrent users.

## Machine quirks

- Tauri dev needs `WEBKIT_DISABLE_DMABUF_RENDERER=1` on this Linux box or the window renders blank. It's set in `.envrc` (gitignored) via direnv — re-apply manually if you run `tauri dev` outside direnv.

## Hard constraints

- All cryptography lives in Rust (`crypto/` crate + `src-tauri`). Never implement crypto in TypeScript; the Go backend never encrypts/decrypts, generates keys, or touches plaintext.
- Backend uses SQLite via modernc.org/sqlite (declared in `server/go.mod`). Do not introduce PostgreSQL or any other database.
- Every protocol message is `{version, type, payload}`.
- Import boundaries are enforced: backend `api -> services -> repository -> database`; React `pages -> components -> services -> websocket`; React only talks to Rust via Tauri commands.
- Naming: client package/product/Rust lib is `corvuss` (double-s); backend dir and Go module are `server`.

## Design decisions

- **Chat request/accept**: `relationships` table with `(requester_id, recipient_id, status)` unique constraint. State machine: `pending → accepted | rejected`. Accept is bidirectional and permanent. Re-request after rejection is cooldown-gated (24h default, `CHAT_REQUEST_COOLDOWN` env, configurable).
- **Direct message enforcement**: dispatcher rejects sends to users without an accepted relationship; returns protocol error `relationship_required`.
- **Contact list**: entirely client-side (no server-side social graph). Client derives from adds + accepted requests + message senders.
- **Group invites**: require accepted personal-chat relationship between inviter and invitee. Invitee must accept before becoming a member. Leave is unilateral.
- **Profile pictures**: encrypted client-side under per-account profile key. Server stores only ciphertext + nonce. `GET /profile-picture/{id}` gated behind accepted relationship. Version strictly increases per upload; no key rotation per edit. `profile_picture_updated { version }` control message broadcast to accepted contacts via dispatcher.
- **Prekey bundle fetch** (`GET /prekey/{id}`) stays ungated regardless of chat-request status; X3DH session establishment is decoupled from the accept step.
- **Rate-limiting on username lookups**: deferred, standard mitigation to revisit later.
- **Blocking**: deferred including profile-key rotation on block.
