# Local Settings and Privacy

## Purpose

Audivo keeps settings, media paths, cookies, and engine details local while providing safe persistence and native file actions.

## Requirements

### Requirement: Local configuration persistence
The application SHALL persist a small JSON configuration containing the absolute download directory, theme, default audio format, Apple Music cookie path, and engine-check metadata. Writes SHALL be atomic and configuration secrets SHALL not be written to logs.

#### Scenario: First launch chooses defaults
- **WHEN** no configuration file exists
- **THEN** the application uses the platform Downloads directory when available, system theme, and safe default media preferences

#### Scenario: Configuration is corrupt
- **WHEN** the configuration file cannot be decoded or fails validation
- **THEN** the application keeps a safe in-memory default, reports a recoverable settings warning, and does not crash

### Requirement: Native directory and cookie selection
The application SHALL provide native Wails dialogs for selecting an output directory and a Netscape-format Apple Music cookie file. The selected paths SHALL be normalized to absolute paths and validated before use.

#### Scenario: Output directory is selected
- **WHEN** the user chooses a writable directory through the native dialog
- **THEN** the path is saved and displayed with a friendly label such as `Downloads`

#### Scenario: Cookie file is selected
- **WHEN** the user selects a readable cookie file
- **THEN** only its path and configured status are stored; cookie content is never displayed, copied into logs, or sent to an Audivo service

### Requirement: Local-only privacy boundary
Audivo SHALL not require an account, remote backend, database, mandatory telemetry, or external upload service. URLs, media, configuration, cookies, and engine logs SHALL remain local except for direct requests made by the selected upstream downloader engine.

#### Scenario: Technical details are expanded
- **WHEN** the user opens technical details for a job
- **THEN** logs are sanitized and contain no cookie contents, cookie command values, or secret material

### Requirement: Native open actions and clipboard
The application SHALL provide clipboard paste for URL input and native open-file/open-folder actions for completed downloads without routing paths through a shell.

#### Scenario: Clipboard contains a supported URL
- **WHEN** the user clicks Paste
- **THEN** the clipboard text fills the URL input and remains subject to normal validation before analysis

#### Scenario: Completed file is opened
- **WHEN** the user clicks Open file or Open folder
- **THEN** the platform-native opener receives the validated absolute path and the frontend receives a friendly error if the path is unavailable
