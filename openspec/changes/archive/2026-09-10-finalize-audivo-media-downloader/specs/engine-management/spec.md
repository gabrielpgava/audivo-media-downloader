## ADDED Requirements

### Requirement: Application-managed engine locations
The application SHALL resolve engine paths independently of the process working directory and SHALL store managed engines in an application-owned cache/data directory derived from platform APIs. It MAY use a valid system engine as a fallback only after validating its executable and version.

#### Scenario: Application starts from another working directory
- **WHEN** Audivo is launched from a directory that does not contain the repository
- **THEN** engine discovery still resolves the configured or managed engine paths correctly

#### Scenario: Engine is missing
- **WHEN** a required engine is absent
- **THEN** the engine manager reports a preparing/missing status and provides a recovery path without instructing the user to run a terminal command

### Requirement: Platform and checksum validation
The engine manager SHALL select artifacts for Windows x64, macOS arm64, macOS amd64, and Linux amd64. Downloaded artifacts SHALL be checked against a pinned manifest and SHA-256 checksum when the upstream source provides one, and SHALL be rejected when the platform or checksum does not match.

#### Scenario: Matching artifact is downloaded
- **WHEN** a manifest entry matches the current OS/architecture and its checksum validates
- **THEN** the artifact is installed with executable permissions where supported and is eligible for use

#### Scenario: Checksum does not match
- **WHEN** a downloaded artifact fails checksum validation
- **THEN** it is discarded and the engine remains unavailable with a clear setup error

### Requirement: YouTube engine contract
YouTube downloads SHALL use the managed official yt-dlp executable, structured metadata mode for analysis, and a machine-parseable progress template for downloads. The engine manager SHALL provision or validate the JavaScript runtime and EJS support required by the selected yt-dlp distribution.

#### Scenario: YouTube analysis runs
- **WHEN** a valid YouTube URL is analyzed
- **THEN** yt-dlp is invoked in metadata-only mode and its JSON is parsed into `MediaPreview` without downloading media

#### Scenario: YouTube download uses simple options
- **WHEN** the user chooses audio or a supported video quality
- **THEN** the argument builder produces only allowlisted yt-dlp flags, uses FFmpeg for required post-processing, and never exposes raw format syntax to the frontend

### Requirement: Apple Music engine contract
Apple Music downloads SHALL use gamdl as the download engine, with output path, temporary path, codec/quality mapping, FFmpeg path, and cookie path supplied by the backend. The application SHALL use an isolated Python environment for gamdl and SHALL not modify global Python packages.

#### Scenario: Gamdl environment is prepared
- **WHEN** gamdl is missing from the application-managed environment
- **THEN** the engine manager prepares the local Python environment and reports progress without requiring administrator privileges or a global `pip install`

#### Scenario: Apple Music download starts
- **WHEN** valid cookies and a supported Apple Music URL are present
- **THEN** gamdl receives direct arguments for the URL and output paths, runs non-interactively, and emits normalized job progress/log events

### Requirement: Engine health and update controls
The application SHALL expose installed version, required version, availability, and validation status for each engine. Automatic update checks SHALL be throttled to at most once per 24 hours, and settings SHALL provide an explicit update-now operation.

#### Scenario: Periodic check is due
- **WHEN** the last successful engine check is older than 24 hours
- **THEN** the application performs a background health/update check and stores its timestamp

#### Scenario: User requests an immediate update
- **WHEN** the user selects update now in settings
- **THEN** the manager checks and updates engines immediately without starting a media download
