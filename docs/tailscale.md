# Tailscale Funnel Setup

Corvus exposes its Go server to the internet over HTTPS using [Tailscale Funnel](https://tailscale.com/kb/1223/tailscale-funnel) instead of a VPS, reverse proxy, or Cloudflare Tunnel. Funnel gives a stable `https://<name>.<tailnet>.ts.net` URL with automatic Let's Encrypt certificates, and works from behind Android's CGNAT with no port forwarding.

This doc covers two setups:

- **Linux (Fedora laptop)** — for local dev/testing before deploying to the phone. Uses systemd, kernel-mode networking, no special flags.
- **Termux (Android phone)** — the actual production deployment target. Needs `--tun=userspace-networking` because Android blocks the netlink access Tailscale normally uses.

Do the account setup (Part 1) once. Then follow either the Linux or Termux section depending on which machine is running the server.

---

## Part 1 — Account setup (one-time, browser only)

This part is identical regardless of which machine will run `tailscaled`.

1. **Create a Tailscale account** at [login.tailscale.com/start](https://login.tailscale.com/start) (GitHub/Google/Microsoft SSO). This creates your tailnet — a private namespace like `tailxxxxx.ts.net`, or a custom name you can set later.

2. **Enable MagicDNS.** Admin console → [DNS](https://login.tailscale.com/admin/dns) tab → toggle **MagicDNS** on. Required before HTTPS certs can be issued.

3. **Enable HTTPS Certificates.** Same DNS page → **HTTPS Certificates** → **Enable HTTPS**. You'll be asked to acknowledge that your tailnet DNS name and device names are published in the public Certificate Transparency log — this is standard for any publicly-trusted TLS cert (same as a VPS + Let's Encrypt), not a Corvus-specific exposure. Confirm.

Nothing is installed on any device yet — this is pure account configuration.

---

## Part 2A — Linux (Fedora), for local dev/testing

### Install

```bash
curl -fsSL https://tailscale.com/install.sh | sh
tailscale version   # sanity check
```

### Start the daemon

Fedora gets normal systemd integration and kernel-mode networking — no manual backgrounding, no userspace flag needed:

```bash
sudo systemctl enable --now tailscaled
sudo systemctl status tailscaled   # should show "active (running)"
```

### Authenticate

```bash
sudo tailscale up
```

Opens a URL (or launches your browser directly). Log in with the account from Part 1.

```bash
tailscale status
# 100.xx.xx.xx   fedora    you@   linux   -
```

### Run the server

```bash
cd server
JWT_SECRET=test-secret-123 go run ./cmd/server
# listens on :8080
```

### Turn on Funnel

```bash
tailscale funnel --bg 8080
```

`--bg` is required for Funnel to persist after the terminal that started it closes — without it, Funnel stops the moment you Ctrl+C or close the shell. With `--bg`, it also survives a `tailscale down` / `tailscale up` cycle or a reboot.

First run triggers a one-time web consent prompt confirming you want port 8080 public — approve it.

```bash
tailscale funnel status
# https://fedora.<tailnet>.ts.net (Funnel on)
# |-- / proxy http://127.0.0.1:8080
```

### Verify externally

From a **different network** (phone on mobile data, not Wi-Fi/Tailscale):

```bash
curl -i https://fedora.<tailnet>.ts.net/register
```

Any HTTP response (even 400/405) confirms MagicDNS, the HTTPS cert, Funnel, and the Go server are all working end to end. A timeout means Funnel isn't live; a TLS error means Part 1 step 3 wasn't actually saved.

### Cleanup

```bash
tailscale funnel 8080 off
sudo systemctl disable --now tailscaled
```

---

## Part 2B — Termux (Android phone), production deployment

### Install

```bash
pkg update && pkg upgrade -y
pkg install tailscale -y
tailscale version   # sanity check
```

### Start the daemon

```bash
mkdir -p $PREFIX/var/lib/tailscale
tailscaled --tun=userspace-networking \
  --state=$PREFIX/var/lib/tailscale/tailscaled.state \
  --socket=$PREFIX/var/run/tailscale/tailscaled.sock &
```

**Why `--tun=userspace-networking` is mandatory here, not optional:** Android 11+ (API 30+) blocks unprivileged apps from binding netlink sockets, which is what `tailscaled`'s normal kernel-mode tunnel needs to create a real `/dev/net/tun` interface. Without root, that fails outright with `permission denied`. Userspace mode routes packets through a SOCKS/HTTP proxy layer inside the `tailscaled` process itself instead, so it needs no root and no special Android permission.

No `sudo` needed — Termux doesn't have `sudo` by default, and userspace mode is exactly what avoids needing it.

Confirm it's running:

```bash
tailscale status
# Tailscale is stopped.
# Log in at: https://login.tailscale.com/a/xxxxxxxxxxxx
```

### Authenticate

```bash
tailscale up
```

Open the printed URL in any browser (phone or another device) logged into the account from Part 1.

```bash
tailscale status
# 100.xx.xx.xx   your-device-name   you@   android   -
```

### Run the server

```bash
cd ~/corvus/server
JWT_SECRET="$(cat ~/.jwt_secret)" ./corvus-server
```

See `docs/deployment.md` for building the binary and persisting `JWT_SECRET`.

### Turn on Funnel

```bash
tailscale funnel --bg 8080
```

Same one-time consent prompt as the Linux flow. `--bg` matters even more here — without it, Funnel dies the moment Termux gets backgrounded, which happens constantly on Android.

```bash
tailscale funnel status
# https://your-device-name.<tailnet>.ts.net (Funnel on)
# |-- / proxy http://127.0.0.1:8080
```

This URL is what you'll pass as `VITE_SERVER_URL` when building the client (see [Client wiring](#client-wiring) below).

### Verify externally

From a network the phone isn't on:

```bash
curl -i https://your-device-name.<tailnet>.ts.net/register
```

### Keeping it alive

Android will kill backgrounded processes aggressively unless told not to:

1. **Run inside tmux**, so a closed Termux session doesn't kill the daemon/server:

   ```bash
   pkg install tmux -y
   tmux new -s corvus
   # start tailscaled, then the server, inside this session
   # detach: Ctrl+b, then d — reattach later with: tmux attach -t corvus
   ```

2. **Disable battery optimization for Termux.** Android Settings → Apps → Termux → Battery → **Unrestricted**. Without this, the low-memory killer reaps the process regardless of Tailscale.

3. **Termux wake lock** (requires the Termux:API companion app):

   ```bash
   termux-wake-lock
   ```

   Keeps the CPU from deep-sleeping, which is needed to keep WebSocket connections alive, not just HTTP.

4. **Lock Termux in the recent-apps view.** Open the recents screen → long-press the Termux card → tap the lock icon. On MIUI/HyperOS and similar OEM skins, this matters more than the battery setting alone.

If `tailscaled` itself gets killed and restarted, re-run `tailscale funnel --bg 8080` — the server binary restarting on its own does **not** require this, since Funnel just proxies to whatever is listening on `127.0.0.1:8080`.

---

## Client wiring

The Tauri client already derives both the HTTP API base and the WebSocket base from a single env var — `VITE_SERVER_URL` — in `client/src/utils/constants.ts`:

```ts
function resolveBaseUrls(): { api: string; ws: string } {
  const serverUrl = import.meta.env.VITE_SERVER_URL;
  if (serverUrl) {
    const base = serverUrl.replace(/\/+$/, "");
    return { api: base, ws: base.replace(/^http/, "ws") };
  }
  // falls back to http://localhost:8080 / ws://localhost:8080 if unset
}
```

`https://` automatically becomes `wss://` — no separate WS config needed.

### Tauri desktop dev

```fish
cd client
set -x VITE_SERVER_URL https://<your-funnel-hostname>.ts.net
npm run tauri dev
```

Or persist it in `client/.env.local` (gitignored by Vite's default `.gitignore`, verify before committing anything):

```bash
echo "VITE_SERVER_URL=https://<your-funnel-hostname>.ts.net" > client/.env.local
```

Verify in the Tauri window's devtools (Network tab) that requests go to the `.ts.net` host, not `localhost`.

### Android APK build

Baked in at build time, since a shipped binary has no runtime env:

```fish
cd client
set -x VITE_SERVER_URL https://<your-funnel-hostname>.ts.net
npm run tauri android build -- --apk --split-per-abi --ci
```

See `docs/android.md` for the full build/signing process.

### CORS note

`CORS_ORIGIN` defaults to `*` (allow-all) server-side, so this doesn't block anything during dev/testing. Only relevant if `CORS_ORIGIN` is later locked down for a production deployment, in which case it needs to match whichever `.ts.net` origin the client is served from.

---

## Troubleshooting

| Symptom                                                            | Cause                                                           | Fix                                                        |
| ------------------------------------------------------------------ | --------------------------------------------------------------- | ---------------------------------------------------------- |
| `tailscaled` exits immediately with `permission denied` on netlink | Missing `--tun=userspace-networking` (Termux only)              | Add the flag                                               |
| `tailscale funnel status` shows nothing / `Funnel off`             | `--bg` omitted, or the terminal that ran it was closed          | Re-run `tailscale funnel --bg 8080`                        |
| `curl` to the `.ts.net` URL times out from outside                 | MagicDNS or HTTPS Certificates not enabled in the admin console | Recheck Part 1, steps 2–3                                  |
| Server unreachable a few minutes after locking the phone           | Android battery optimization / low-memory killer                | See "Keeping it alive" above                               |
| `tailscale up` hangs with no browser prompt                        | No default browser, or URL wasn't opened manually               | Copy the printed URL into any browser by hand              |
| Client still hits `localhost:8080`                                 | `VITE_SERVER_URL` not set before the build/dev command          | Confirm with `echo $VITE_SERVER_URL` (fish) before running |
