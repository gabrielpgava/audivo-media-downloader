## 1. Baseline and dependency cleanup

- [x] 1.1 Remove the machine-specific Wails `replace` directive and update the Go/frontend dependency manifests to stable compatible versions.
- [x] 1.2 Add reproducible frontend `typecheck`, test, and build scripts and confirm Wails asset generation order.
- [x] 1.3 Replace the minimal ignore rules with entries for generated frontend output, dependencies, engines, cookies, config, downloads, logs, caches, and temporary files.
- [x] 1.4 Remove the tracked `assets/yt-dlp` artifact after managed engine resolution is in place.

## 2. Domain contracts, paths, and configuration

- [x] 2.1 Create typed service, media, format, quality, job state, progress, result, engine status, settings, and user-error models.
- [x] 2.2 Implement strict URL parsing and YouTube/Apple Music host detection with unit coverage for valid, invalid, and lookalike hosts.
- [x] 2.3 Implement platform-aware default Downloads, config, cache, engine, temporary, and cookie path helpers using absolute paths.
- [x] 2.4 Implement atomic JSON configuration load/save, validation, defaults, migration-free recovery, and restrictive permissions where supported.
- [x] 2.5 Implement native output-directory/cookie selection and validated native open-file/open-folder platform helpers.

## 3. Process execution and download lifecycle

- [x] 3.1 Implement a direct-argument process runner that captures stdout, stderr, exit status, context cancellation, and sanitized technical logs.
- [x] 3.2 Implement platform process-tree ownership for POSIX process groups and Windows child-process cleanup.
- [x] 3.3 Implement the one-active-job manager with stable job IDs, context cancellation, terminal states, restart-after-cancel, and shutdown cleanup.
- [x] 3.4 Implement structured event publication for `download:updated` and `download:log` without backend cancellation listeners.
- [x] 3.5 Add fake-engine process tests for success, progress, stderr, failure, slow execution, cancellation, and child-process termination.

## 4. Engine manager and YouTube provider

- [x] 4.1 Implement engine manifest models, managed cache resolution, executable validation, platform/architecture selection, and checksum verification.
- [x] 4.2 Implement engine health/status reporting, 24-hour throttled checks, and explicit update-now behavior.
- [x] 4.3 Implement yt-dlp metadata analysis with structured JSON parsing into `MediaPreview` and service-valid option generation.
- [x] 4.4 Implement yt-dlp download argument building for MP3/M4A audio and simple video quality choices with structured progress templates.
- [x] 4.5 Implement Deno/EJS/FFmpeg/ffprobe dependency discovery and user-facing missing/invalid-engine errors.
- [x] 4.6 Add unit tests for manifest selection, checksums, yt-dlp arguments, metadata fixtures, and progress parser edge cases.

## 5. Apple Music provider and gamdl isolation

- [x] 5.1 Implement application-local Python environment/bootstrap detection for gamdl without global package writes or administrator privileges.
- [x] 5.2 Implement the embedded/documented gamdl metadata helper and map its output into the shared media preview contract.
- [x] 5.3 Implement gamdl argument building for songs, albums, playlists, music videos, cookies, output paths, quality, codec, and FFmpeg settings without interactive prompts.
- [x] 5.4 Implement service-aware Apple Music format mapping and safe FFmpeg conversion when the requested output format differs from gamdl output.
- [x] 5.5 Add tests for gamdl arguments, cookie-path redaction, missing environment/cookies, metadata fixtures, and cancellation.

## 6. Wails application bindings and lifecycle

- [x] 6.1 Compose the application services in `App`, expose unified `AnalyzeURL`, `StartDownload`, `CancelDownload`, settings, engine, selection, and open-action bindings.
- [x] 6.2 Register Wails startup/shutdown callbacks and ensure active process cleanup completes before application exit.
- [x] 6.3 Regenerate Wails TypeScript bindings and remove obsolete provider-specific bindings.
- [x] 6.4 Configure the Wails window title, initial size, minimum size, assets, and production build settings for the new workflow.

## 7. Unified React/MUI experience

- [x] 7.1 Replace provider navigation and page-specific forms with a single reducer-driven app state machine covering all required states.
- [x] 7.2 Implement URL input, Paste action, validation/analyzing feedback, and typed backend event subscription with correct unsubscribe cleanup.
- [x] 7.3 Implement media preview, service-aware format/quality controls, destination selector, and Download/Cancel actions.
- [x] 7.4 Implement visual progress, processing stages, sanitized technical-details disclosure, completion actions, retry, and friendly error views.
- [x] 7.5 Implement settings dialog for theme, output directory, Apple Music cookies, engine statuses, and update-now.
- [x] 7.6 Configure MUI system/light/dark themes, Audivo accent styling, accessible labels/focus/disabled states, responsive layout, and dead CSS/import cleanup.
- [x] 7.7 Add frontend reducer and component tests for state transitions, stale events, options, errors, clipboard, and settings behavior.

## 8. Quality, CI, release, and documentation

- [x] 8.1 Add CI workflow for frontend install/typecheck/build, Go test, Go vet, binding validation, and Wails smoke builds.
- [x] 8.2 Add release workflow/matrix for Windows x64, macOS arm64/amd64, and Linux amd64 only where native smoke builds pass.
- [x] 8.3 Add `THIRD_PARTY_NOTICES.md` covering Wails, yt-dlp, gamdl, FFmpeg, Deno/EJS, Python/bootstrap tooling, and other distributed artifacts.
- [x] 8.4 Rewrite README with product flow, supported services, development, build, engine setup, Apple Music cookies, architecture, testing, release, licensing, and limitations.
- [x] 8.5 Run focused backend/frontend tests, full build checks, and OpenSpec validation; fix regressions before manual smoke testing.

## 9. Manual acceptance and handoff

- [x] 9.1 Execute and record YouTube video/audio, merge, cancellation, invalid URL, missing engine, and restart-after-cancel smoke cases.
- [x] 9.2 Execute and record Apple Music song/album, cookies, cancellation, output selection, and engine preparation smoke cases.
- [x] 9.3 Execute and record persistence, theme, clipboard, native open actions, smaller-window layout, and accessibility keyboard checks.
- [x] 9.4 Mark only verified tasks complete, document real remaining limitations, and leave the change ready for archive only when all required acceptance evidence exists.
