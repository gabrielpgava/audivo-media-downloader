## ADDED Requirements

### Requirement: Separate provider cookie storage
Audivo SHALL maintain independent cookie paths and connection state for Apple Music and YouTube, and SHALL never use one provider's cookie file for the other provider.

#### Scenario: Apple Music cookies are configured
- **WHEN** the Apple Music session is connected or imported
- **THEN** only the Apple Music cookie path and status are updated

#### Scenario: YouTube cookies are configured
- **WHEN** the YouTube session is connected or imported
- **THEN** only the YouTube cookie path and status are updated

#### Scenario: Provider cookie is missing
- **WHEN** a provider operation requests its cookie path
- **THEN** the backend reads only the field belonging to that provider and never falls back to the other provider's field

### Requirement: Private atomic cookie files
Captured cookies SHALL be serialized to Netscape format under an application-private cookie directory, written with restrictive permissions, and promoted atomically only after validation.

#### Scenario: Captured cookies are valid
- **WHEN** a provider authentication session returns a valid cookie set
- **THEN** Audivo writes a private Netscape file, stores its path, and makes the file available to the provider engine

#### Scenario: Capture is interrupted
- **WHEN** the browser closes or cancellation occurs before cookie validation completes
- **THEN** Audivo removes the temporary file and preserves the previous managed cookie file

### Requirement: Provider-aware cookie validation
The system SHALL validate that imported or captured files are readable Netscape cookie files with at least one record belonging to the selected provider's domain allowlist.

#### Scenario: Manual import matches the provider
- **WHEN** the user selects a readable Netscape file containing records for the selected provider
- **THEN** Audivo copies or promotes it into private managed storage and marks that provider as connected

#### Scenario: Manual import has the wrong provider
- **WHEN** the selected file is valid Netscape format but contains no record for the selected provider
- **THEN** Audivo rejects it with a provider-specific validation error and does not change either provider's state

#### Scenario: Manual import is malformed
- **WHEN** the selected file is unreadable, a directory, empty, or not Netscape format
- **THEN** Audivo rejects it with an actionable error and does not persist the path

### Requirement: Correct engine routing
The backend SHALL pass YouTube cookies only to `yt-dlp` and Apple Music cookies only to `gamdl` whenever the corresponding cookie path is configured.

#### Scenario: YouTube analysis uses a connected session
- **WHEN** a YouTube URL is analyzed and a valid YouTube cookie path exists
- **THEN** `yt-dlp` receives that path through its cookie argument and no Apple Music cookie path is read

#### Scenario: YouTube download has no connected session
- **WHEN** a public YouTube URL is analyzed or downloaded without YouTube cookies
- **THEN** the public flow remains available and no Apple Music cookie argument is sent to `yt-dlp`

#### Scenario: Apple Music analysis uses a connected session
- **WHEN** an Apple Music URL is analyzed with a valid Apple Music cookie path
- **THEN** `gamdl` receives only that path through `--cookies-path`

#### Scenario: Apple Music session is missing
- **WHEN** an Apple Music URL is analyzed or downloaded without valid Apple Music cookies
- **THEN** Audivo stops before the engine operation and reports that Apple Music authentication must be configured

### Requirement: Backward-compatible configuration migration
The configuration loader SHALL preserve a valid existing Apple Music cookie path, add an optional YouTube cookie path, and recover safely from invalid or incomplete new fields.

#### Scenario: Existing configuration has Apple Music cookies
- **WHEN** Audivo loads a configuration created before this change with a valid `AppleMusicCookiesPath`
- **THEN** Apple Music remains usable and the new YouTube cookie state defaults to disconnected

#### Scenario: Existing configuration is invalid
- **WHEN** a configuration contains a relative, unreadable, or invalid provider cookie path
- **THEN** Audivo uses safe defaults for the invalid field, reports a recoverable settings issue, and does not expose the cookie contents

### Requirement: Secret-safe logs and responses
Cookie values, file contents, cookie headers, and unredacted cookie paths SHALL be excluded from logs, technical details, Wails responses, and error details.

#### Scenario: Engine emits a cookie argument
- **WHEN** an engine writes a command or error containing a cookie option/path
- **THEN** Audivo redacts the option and secret path before publishing the technical log or user-facing error

#### Scenario: Settings are read
- **WHEN** the frontend requests provider settings/status
- **THEN** the response contains only configured state and safe metadata, not cookie file content or values

### Requirement: Local disconnect and replacement
Configurações SHALL allow the user to disconnect a provider locally and replace its cookie file through a later login or import.

#### Scenario: User disconnects a provider
- **WHEN** the user clicks `Desconectar` for a connected provider
- **THEN** Audivo removes the managed cookie file, clears only that provider's path, and reports disconnected without attempting to sign out of the provider website

#### Scenario: User replaces a provider session
- **WHEN** a new valid capture or import is completed for a connected provider
- **THEN** Audivo atomically replaces only that provider's managed file and keeps the other provider unchanged
