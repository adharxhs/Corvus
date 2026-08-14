# Android Build & Packaging

Corvus builds as a native Android APK via Tauri 2. The APK includes a WebView UI (React) and a Rust native library (profile-key encryption). The server URL is baked in at build time.

## Prerequisites

- Rust toolchain with Android targets:
  ```bash
  rustup target add aarch64-linux-android armv7-linux-androideabi i686-linux-android x86_64-linux-android
  ```
- Android SDK + NDK (via Android Studio or standalone). `ANDROID_HOME` must be set.
- Java 17+ (the JBR bundled with Android Studio works)

## Build commands

### Universal APK (all architectures, ~36 MB)

```bash
cd client
VITE_SERVER_URL=https://your-server-host \
  npm run tauri android build -- --apk --ci
```

Output: `src-tauri/gen/android/app/build/outputs/apk/universal/release/app-universal-release.apk`

### Split per-ABI (~12 MB each, recommended for distribution)

```bash
cd client
VITE_SERVER_URL=https://your-server-host \
  npm run tauri android build -- --apk --split-per-abi --ci
```

Output per architecture:
- `src-tauri/gen/android/app/build/outputs/apk/arm64/release/app-arm64-release.apk` (most phones)
- `src-tauri/gen/android/app/build/outputs/apk/arm/release/app-arm-release.apk`
- `src-tauri/gen/android/app/build/outputs/apk/x86_64/release/app-x86_64-release.apk`
- `src-tauri/gen/android/app/build/outputs/apk/x86/release/app-x86-release.apk`

Share the `arm64` APK with friends — covers the vast majority of Android phones.

## Server URL configuration

The `VITE_SERVER_URL` env var sets the API and WebSocket base URL at Vite build time. The client auto-derives both:

- `API_BASE_URL` = the value of `VITE_SERVER_URL`
- `WS_BASE_URL` = `VITE_SERVER_URL` with `http` → `ws` / `https` → `wss`

If unset, both default to `http://localhost:8080` / `ws://localhost:8080` (useful for local dev only).

**Note**: Release APKs have `usesCleartextTraffic=false`. The server must be reachable over HTTPS.

## Signing

Release builds are signed with a project keystore:

- **Keystore**: `client/src-tauri/keystore/corvus-release.keystore`
- **Properties**: `client/src-tauri/keystore.properties` (gitignored)
- **Certificate**: `CN=Corvus, OU=Corvus, O=Corvus`

**Important**: The same keystore must be used for all updates to the same `applicationId` (`com.adharxhs.corvus`). Android rejects APKs signed with a different key. Back up the keystore and properties file outside the repository.

If `keystore.properties` is absent, the build produces an unsigned APK (not installable).

## Installing on a device

```bash
# Connect device via USB, enable USB debugging
adb devices

# Uninstall old build first if switching between debug/release keys
adb uninstall com.adharxhs.corvus

# Install the release APK
adb install client/src-tauri/gen/android/app/build/outputs/apk/arm64/release/app-arm64-release.apk
```

Or transfer the APK file to the phone and open it in a file manager.

## Debug vs release

| | Debug | Release |
|---|---|---|
| `usesCleartextTraffic` | `true` (HTTP works) | `false` (HTTPS required) |
| APK size | ~280 MB (unstripped debug symbols) | ~12 MB per ABI |
| Keystore | auto-generated debug key | project release keystore |
| Minification | disabled | enabled |

The large debug APK is normal — it contains full Rust debug symbols for all four CPU architectures. The release APK strips these, giving the ~12 MB size.

## Machine quirks

On some Linux systems, the Tauri dev window renders blank. Set `WEBKIT_DISABLE_DMABUF_RENDERER=1` before running `tauri dev`.
