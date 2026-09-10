## Why

Audivo currently has two disconnected frontend pages and direct process calls that do not provide a reliable end-to-end download flow. The current Apple Music path invokes the wrong engine, paths depend on the working directory, cancellation leaks listeners/processes, and a clean checkout does not build; this change finishes the product around one simple, local desktop workflow.

## What Changes

- Replace the page-specific download methods with a typed Go download service and one active cancellable job.
- Add allowlisted YouTube and Apple Music URL detection, metadata analysis, preview data, format/quality options, structured progress, friendly errors, and completion actions.
- **BREAKING** replace `DownloadYoutube` and `DownloadAppleMusic` with unified Wails bindings centered on `AnalyzeURL`, `StartDownload`, and `CancelDownload`.
- Use official `yt-dlp` for YouTube and `gamdl` for Apple Music; isolate engines in application-managed directories and validate versions/checksums before execution.
- Provision gamdl in an application-local Python environment without modifying global Python, and keep Apple Music cookie contents out of logs and telemetry.
- Resolve platform download/config/cache paths through system APIs, persist small settings atomically, and provide native directory/file selection and open actions.
- Replace the YouTube/Apple Music navbar with one task-first MUI screen covering idle, analysis, ready, download, processing, completion, cancellation, error, settings, theme, and technical-detail states.
- Add backend unit tests, fake process tests, frontend state tests, CI/build checks, release scaffolding, updated documentation, and third-party notices.
- Remove the tracked `assets/yt-dlp` artifact and the machine-specific Go module replacement; keep generated build output, cookies, logs, caches, and dependencies untracked.

## Capabilities

### New Capabilities

- `download-workflow`: Validated service detection, metadata analysis, typed download requests, one active job, structured progress, cancellation, completion, and friendly failures.
- `engine-management`: Application-local yt-dlp, Deno/EJS, FFmpeg, ffprobe, and gamdl provisioning, discovery, validation, update checks, and platform selection.
- `local-settings-and-privacy`: Atomic local settings, default/download directory selection, theme and audio preferences, Apple Music cookie path handling, redacted diagnostics, and local-only execution.
- `unified-desktop-ui`: Single-screen React/MUI workflow, preview/options/progress/completion states, native dialogs, clipboard paste, accessibility, responsive sizing, and error recovery.
- `quality-and-distribution`: Unit/fake-engine coverage, frontend/backend checks, CI, release build matrix, README, third-party notices, and repository hygiene.

### Modified Capabilities

None. The repository has no existing capability specs; all contracts introduced by this change are new.

## Impact

- Go backend: `app.go`, `main.go`, `amDownloader.go`, and `ytDownloader.go` become a small package-oriented application and process boundary.
- Wails bindings: generated TypeScript methods and event payloads change from string-based page-specific calls to typed unified operations.
- React frontend: `App.tsx`, the two downloader pages, and `style.css` are replaced or consolidated into a single workflow and settings surface.
- Dependencies and packaging: Wails v2 and compatible frontend dependencies are refreshed; engine binaries are no longer committed; CI and release workflows are added.
- Runtime storage: user config, engine cache, temporary job files, and optional cookie paths are local to the user and never require an account, server, database, telemetry, or remote API owned by Audivo.
