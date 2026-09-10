# Unified Desktop UI

## Purpose

Audivo presents one accessible, task-first desktop workflow for analyzing media, choosing options, monitoring progress, and recovering from errors.

## Requirements

### Requirement: Single task-first screen
The frontend SHALL present one primary downloader screen with a prominent URL input, Paste action, Audivo identity, and Settings action. It SHALL not require the user to choose YouTube or Apple Music from a navigation bar.

#### Scenario: Idle state is displayed
- **WHEN** the application opens without a URL
- **THEN** the user sees the URL input and primary action without a console, provider tabs, or unnecessary configuration cards

#### Scenario: Provider is detected
- **WHEN** a valid URL is analyzed
- **THEN** the screen identifies YouTube or Apple Music in the preview without changing the primary navigation structure

### Requirement: Preview and simple options
The frontend SHALL render available thumbnail, title, author, platform, duration, media type, format, quality, and destination data from the typed backend response. It SHALL show only service-valid user-friendly options.

#### Scenario: Preview is ready
- **WHEN** analysis succeeds
- **THEN** the user can choose a simple audio/video format and quality and start the download from the same screen

#### Scenario: Quality is unavailable
- **WHEN** the analyzed media does not offer a requested quality
- **THEN** that option is not shown or is disabled with an understandable label

### Requirement: Visual progress and completion
The frontend SHALL render explicit analyzing, preparing, downloading, processing, completed, cancelled, and error states from typed events. It SHALL keep the technical console secondary behind a details disclosure.

#### Scenario: Download is active
- **WHEN** a progress event is received
- **THEN** the screen shows stage, progress, speed, ETA where available, filename, and a Cancel action without exposing raw engine output as the primary content

#### Scenario: Download completes
- **WHEN** a completed event includes output paths
- **THEN** the screen shows a clean confirmation with Open file and Open folder actions

### Requirement: Friendly recoverable errors
The frontend SHALL translate known error codes into concise actionable messages and SHALL provide retry or settings actions where appropriate. It SHALL preserve sanitized technical details behind an explicit disclosure.

#### Scenario: Engine is unavailable
- **WHEN** the backend reports a missing or invalid engine
- **THEN** the UI explains that Audivo is preparing or needs an engine update and offers the relevant settings/retry action

#### Scenario: URL is unsupported
- **WHEN** analysis returns an unsupported URL error
- **THEN** the UI identifies the supported services and keeps the user on the input state without starting a job

### Requirement: Theme, accessibility, and window behavior
The UI SHALL support system, light, and dark themes using MUI theme configuration, visible keyboard focus, accessible labels, clear disabled/loading states, and a minimum window size that avoids clipping in a smaller desktop window.

#### Scenario: System theme changes
- **WHEN** the user selects system theme or the operating system theme changes
- **THEN** the UI updates its palette without losing the current form or job state

#### Scenario: Keyboard-only interaction
- **WHEN** a user navigates with the keyboard
- **THEN** URL input, Paste, options, Settings, Download, Cancel, retry, disclosure, and completion actions are reachable with visible focus
