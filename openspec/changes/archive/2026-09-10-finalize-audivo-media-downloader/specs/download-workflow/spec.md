## ADDED Requirements

### Requirement: Supported service detection
The backend SHALL parse and validate an HTTP(S) URL before any external engine is invoked. It SHALL recognize YouTube (`youtube.com`, `youtu.be`, and supported subdomains) and Apple Music (`music.apple.com`), and SHALL reject other hosts with a user-facing error.

#### Scenario: YouTube URL is recognized
- **WHEN** the user submits a valid YouTube URL
- **THEN** analysis returns the YouTube service and no engine is invoked with an unvalidated host

#### Scenario: Unsupported URL is rejected
- **WHEN** the user submits an unsupported or malformed URL
- **THEN** analysis returns an actionable unsupported-URL error and no downloader process starts

### Requirement: Metadata analysis
The backend SHALL expose an analysis operation that returns a typed media preview containing service, title when available, thumbnail URL when available, author/channel/artist when available, duration when available, media type, and service-valid format/quality options.

#### Scenario: YouTube metadata is available
- **WHEN** a supported YouTube URL is analyzed successfully
- **THEN** the response contains structured metadata and options derived from yt-dlp without requiring frontend parsing of engine output

#### Scenario: Apple Music requires cookies
- **WHEN** an Apple Music URL is analyzed without configured valid cookies
- **THEN** the response identifies the missing authentication setup without exposing cookie contents or starting a download

### Requirement: Typed single-job lifecycle
The backend SHALL represent download state with explicit states `idle`, `analyzing`, `ready`, `preparing`, `downloading`, `processing`, `completed`, `cancelled`, and `error`. It SHALL allow at most one active download job and SHALL return a stable job ID when a job starts.

#### Scenario: Second download is prevented
- **WHEN** a download job is active and the user starts another download
- **THEN** the second request is rejected with a busy error and the active job remains unchanged

#### Scenario: Successful job reaches completion
- **WHEN** the selected engine and post-processing finish successfully
- **THEN** the job emits a completed event containing the output file paths and output directory

### Requirement: Structured progress and technical logs
The backend SHALL translate engine output into typed progress events containing job ID, state, stage, percent, speed, ETA, filename, and user-facing message. Raw stdout/stderr SHALL remain separate technical logs and SHALL never be required for normal frontend rendering.

#### Scenario: Progress template is parsed
- **WHEN** yt-dlp emits a recognized progress record
- **THEN** the backend emits a typed progress event with numeric percent and ETA where available

#### Scenario: Malformed engine output is received
- **WHEN** an engine emits an unrecognized line
- **THEN** the line is retained as sanitized technical log data and the job continues unless the process exits unsuccessfully

### Requirement: Explicit cancellation and shutdown cleanup
The backend SHALL expose `CancelDownload(jobID)` and SHALL cancel the job context, terminate the controlled process tree, clean temporary job files where safe, emit a cancelled state, and permit a new job immediately afterward. Application shutdown SHALL perform the same cleanup for an active job.

#### Scenario: User cancels an active job
- **WHEN** the requested job ID is active and the user cancels it
- **THEN** yt-dlp, gamdl, ffmpeg, and their child processes are terminated, the frontend receives `cancelled`, and no stale cancellation listener is registered

#### Scenario: Unknown job is cancelled
- **WHEN** the user calls `CancelDownload` with an unknown or already terminal job ID
- **THEN** the backend returns a deterministic not-active result without affecting another job

### Requirement: Safe process and output invocation
All engine commands SHALL use `exec.CommandContext` or an equivalent direct process API with each user-controlled value passed as a separate argument. The backend SHALL use absolute validated output directories and SHALL not invoke a shell for downloader execution.

#### Scenario: URL contains shell metacharacters
- **WHEN** a URL or output value contains shell metacharacters
- **THEN** the value is passed as data to the process and cannot create an additional command or argument

#### Scenario: Output directory is invalid
- **WHEN** the requested output directory does not exist, is not writable, or is not an absolute path
- **THEN** the request is rejected before the engine starts with a path error
