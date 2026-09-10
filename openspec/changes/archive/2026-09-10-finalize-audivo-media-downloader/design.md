## Context

Audivo is a small Wails v2 application with a Go backend, React/TypeScript frontend, and direct calls to external download tools. The current code has provider-specific bindings, a tracked Python-script yt-dlp artifact, a literal `~/Downloads` output template, an Apple Music method that calls yt-dlp, a leaked cancellation listener, no job model, and no metadata/progress contract. A clean checkout also requires the frontend build before Go embedding can compile.

The design must keep Wails v2, Go, React, TypeScript, Vite, and MUI; remain local-only; avoid a server/database/account; and keep the main workflow understandable in a few seconds. It must also treat yt-dlp, gamdl, FFmpeg, Deno/EJS, Python, cookies, and child processes as operational boundaries rather than frontend concerns.

## Goals / Non-Goals

**Goals:**

- Provide one typed, cancellable, one-job download pipeline for YouTube and Apple Music.
- Make engine discovery and installation independent of the working directory and global user packages.
- Deliver metadata preview, visible progress, friendly errors, settings, native path actions, and persistent preferences.
- Make the core domain testable with deterministic fake processes and fixture output.
- Restore clean development/build/CI/release hygiene without committing platform binaries.

**Non-Goals:**

- No download history, media library, player, queue, account system, cloud sync, telemetry, or generic browser.
- No additional providers beyond YouTube and Apple Music.
- No custom implementation of Apple Music private APIs, decryption, muxing, or media extraction.
- No shell-based downloader execution, automatic silent application update, or server-side cookie handling.

## Decisions

### 1. Small package-oriented Go backend

Use `internal/app`, `internal/download`, `internal/engines`, `internal/config`, and `internal/platform` packages. Keep interfaces narrow and concrete: a download service, engine manager, process supervisor, config store, and platform helpers are sufficient.

The bound `App` owns the Wails context and composes these services. Public Wails methods use concrete JSON-friendly request/response types; domain packages do not depend on Wails runtime APIs except for the application event publisher adapter.

### 2. Unified typed API and event contract

Expose `AnalyzeURL`, `StartDownload`, `CancelDownload`, settings methods, engine status/update methods, native selection methods, and open-file/open-folder methods. `Service` is derived by the backend and is not trusted from the frontend.

Use one `DownloadEvent` on `download:updated` for state/progress/terminal data and a separate `download:log` event for sanitized technical lines. The React app consumes these events with a reducer and stores the unsubscribe function returned by `EventsOn`.

### 3. One active job with process-tree ownership

The job manager stores one active job with a context cancel function and a wait group. Every runner uses `exec.CommandContext` and direct arguments. POSIX starts use a process group; Windows uses a Job Object or the platform equivalent so cancellation reaches FFmpeg/gamdl/yt-dlp descendants. Shutdown cancels and waits for the active job before returning.

This is preferred over a frontend cancellation event because the backend owns the lifecycle and can enforce cleanup even when the UI is reloaded or closed.

### 4. Engine resolution and provisioning

Resolve engines in this order: application-managed cache, release-provided resource directory, then a validated system executable where allowed. The managed root is derived from `os.UserConfigDir`/`os.UserCacheDir` and includes OS/architecture directories; no relative `assets/` path is used.

Use a versioned manifest with source URL, platform, expected version, SHA-256, executable name, and license reference. Official standalone yt-dlp assets are preferred. Deno and EJS support are validated as part of the YouTube engine contract. FFmpeg/ffprobe are resolved together. Gamdl runs from an application-local Python environment with a supported Python version and pinned package set; provisioning is performed through an application-managed bootstrap tool/environment and never writes to global site-packages.

Automatic engine checks use a persisted timestamp and run no more than once per 24 hours. Settings exposes update now and the current health of each engine.

### 5. Provider-specific adapters behind one workflow

The YouTube adapter uses yt-dlp metadata-only JSON for analysis and a delimiter-prefixed progress template for downloads. The Apple Music adapter uses gamdl's documented Python API through a small helper for metadata and gamdl's CLI for downloads. The adapter maps simple UI options to fixed allowlisted arguments, including FFmpeg conversion when the requested output format differs from the native engine output.

Cookies are passed only as a validated file path to gamdl, never as content or a log value. Apple Music analysis/download returns an actionable configuration error when cookies are absent or unreadable.

### 6. Local configuration and native actions

Store settings as a small JSON file under the platform user config directory. Write to a temporary file, flush, rename atomically, and apply restrictive permissions where supported. Resolve the default Downloads directory using platform APIs/user-directory conventions, then fall back to a verified `Downloads` child of the user home.

Use Wails v2 native dialogs for output directory and cookies. Use platform-native direct executables for opening files/folders; do not use `sh -c`, `cmd /c`, or PowerShell for downloader or path execution.

### 7. Single-screen React state machine

Replace the provider navbar and page components with a small `App` reducer plus `UrlInput`, `MediaPreview`, `DownloadOptions`, `DownloadProgress`, `CompletedDownload`, and `SettingsDialog` components. MUI remains the design system. Theme selection is system/light/dark, with one Audivo accent and no provider-colored application chrome.

The reducer owns the explicit UI states and prevents stale events from a previous job from changing a new job. Technical logs are hidden behind a disclosure. Retry reuses the last validated request but requires a fresh job ID.

### 8. Validation and release

Backend tests use a cross-platform fake engine helper that emits stdout/stderr, progress, success, failure, slow execution, and child-process behavior. Frontend tests cover the reducer and user-visible option/state mapping. CI installs frontend dependencies, builds frontend assets, runs Go test/vet, regenerates/checks bindings, and performs Wails smoke builds on supported runners. Release workflows build only runner-tested targets and attach checksums/notices.

## Risks / Trade-offs

- **[Upstream engine behavior changes]** → Pin tested versions in the manifest, parse only documented/owned output contracts, expose update-now, and keep raw diagnostics available after sanitization.
- **[Gamdl/Python provisioning is platform-sensitive]** → Keep the environment app-local, validate Python/package versions before use, isolate bootstrap failures as engine status, and never fall back to global writes.
- **[Child process termination differs across OSes]** → Encapsulate process-tree handling behind platform files and test cancellation with a fake parent/child process on each CI OS.
- **[Real provider acceptance requires network/subscription/cookies]** → Keep CI deterministic with fakes and record manual YouTube/Apple Music smoke evidence separately; do not mark provider E2E complete from unit tests alone.
- **[Wails generated bindings can drift]** → Make binding generation part of the documented workflow and CI validation, and keep bound structs JSON-friendly.
- **[Large scope increases integration risk]** → Implement in dependency order: contracts/paths, runner, engines, bindings, UI, tests/packaging; keep one active job and no speculative queue/history abstractions.

## Migration Plan

1. Create the new internal packages and bound types while keeping the old methods only until the frontend migration compiles.
2. Replace frontend calls with the unified API and regenerate bindings.
3. Remove the legacy provider files/methods and tracked `assets/yt-dlp` after focused tests pass.
4. Remove the absolute `go.mod` replacement and update dependency locks.
5. Add engine manifests, local configuration migration/defaulting, CI, release metadata, README, and third-party notices.
6. Validate with fake-engine tests, build checks, and manual provider smoke tests. If the change must be rolled back, restore the previous frontend bindings and provider files; no user database migration is required.

## Open Questions

- Exact upstream asset URLs/checksum filenames may change and must be resolved when the engine manifest is implemented; the manifest format and validation behavior are fixed by this design.
- Native code signing/notarization credentials are not available in this repository; release artifacts can be built and checksummed, while signing remains a documented deployment follow-up.
