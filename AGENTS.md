# AGENTS.md

Early-stage end-to-end encrypted messaging platform. Three components planned, only the client is scaffolded.

## Read the plan first

- `_md_files/` (gitignored, local-only) holds the authoritative implementation plan: `00_Master_Plan.md` (stages, dependency order, hard constraints) through `14_Testing.md`.
- Implementation order is mandatory and sequential: Environment -> Backend -> DB -> Auth -> WebSocket -> Client -> Cloudflare -> X3DH -> Double Ratchet -> Sender Keys -> Packaging. No stage may start before its dependencies.
- Plan docs lag the repo: they say `corvus-server/`, `docs/`, and a `crypto/` crate exist. Reality: `server/` is the Go module, plan docs live in `_md_files/`, and `crypto/` + `shared/` are empty stubs. Do not create duplicate dirs to match the docs.

## Layout and current state

- `client/` — Tauri 2 + React 19 + TypeScript + Vite 7. Only implemented component (still default `greet` scaffold).
- `server/` — Go 1.26 module named `server` (not `corvus-server`). Dependencies declared in `go.mod` only; no Go source yet. `go build ./...` currently matches no packages.
- `crypto/` — initialized Rust crate (X3DH, Double Ratchet, Sender Keys). All specified dependencies (`x25519-dalek`, `ed25519-dalek`, `double-ratchet-2`, `aes-gcm`, `rand_core`) are installed and resolved.
- `shared/` — planned protocol docs/schemas only; must stay documentation-only.
- No CI, no tests, no linter. Not yet set up in repo.

## Commands

- Client dev server: `cd client && npm run dev` — Vite runs on **fixed port 1420** (strictPort; change `tauri.conf.json` `devUrl` if it moves).
- Full Tauri app: `cd client && npm run tauri dev`
- Client typecheck/build: `npm run build` (= `tsc && vite build`). No test or lint scripts exist.
- Backend: `cd server && go build ./...` / `go test ./...`
- Crypto deps to add per `_md_files/02_Environment_Setup.md`: x25519-dalek, ed25519-dalek, double-ratchet-2, aes-gcm, rand_core.

## Machine quirks

- Tauri dev needs `WEBKIT_DISABLE_DMABUF_RENDERER=1` on this Linux box or the window renders blank. It's set in `.envrc` (gitignored) via direnv — re-apply manually if you run `tauri dev` outside direnv.

## Hard constraints (from `_md_files/00_Master_Plan.md`)

- All cryptography lives in Rust (`crypto/` crate + `src-tauri`). Never implement crypto in TypeScript; the Go backend never encrypts/decrypts, generates keys, or touches plaintext.
- Backend uses SQLite via modernc.org/sqlite (declared in `server/go.mod`). Later plan docs mention PostgreSQL — that contradicts the master plan; keep SQLite.
- Every protocol message is `{version, type, payload}`.
- Import boundaries are enforced: backend `api -> services -> repository -> database`; React `pages -> components -> services -> websocket`; React only talks to Rust via Tauri commands.
- Naming: client package/product/Rust lib is `corvuss` (double-s); backend dir and Go module are `server`.
