# Quality and Distribution

## Purpose

Audivo maintains repeatable automated validation, documented releases, and a clean repository suitable for supported desktop distributions.

## Requirements

### Requirement: Backend unit and process tests
The repository SHALL test service detection, platform paths, atomic configuration, yt-dlp and gamdl argument builders, progress parsing, job state transitions, cancellation, shutdown, stderr handling, and engine checksum validation without requiring real media downloads in CI.

#### Scenario: Fake engine succeeds
- **WHEN** a fake engine emits progress and exits zero
- **THEN** the runner produces the expected progress and completed job events

#### Scenario: Fake engine fails or is cancelled
- **WHEN** a fake engine emits stderr and exits non-zero or is cancelled while running
- **THEN** the runner produces the expected error/cancelled state and leaves no controlled child process running

### Requirement: Frontend and build validation
The repository SHALL provide reproducible commands for frontend installation, typecheck, frontend build, Go tests, Go vet, and Wails development/production builds. CI SHALL run the checks in an order that produces frontend assets before Go embedding/build steps.

#### Scenario: Clean CI checkout
- **WHEN** CI checks out the repository and installs declared dependencies
- **THEN** frontend typecheck/build, Go test/vet, and the supported Wails smoke build complete without relying on local absolute paths

#### Scenario: Generated bindings are refreshed
- **WHEN** public Go bindings change
- **THEN** generated TypeScript bindings are regenerated and typechecked as part of the validation workflow

### Requirement: Release and license documentation
The repository SHALL document development, build, supported services, engine setup, Apple Music cookie setup, architecture, contribution, and license information. Distributed third-party engines and runtimes SHALL be listed with source and license information in `THIRD_PARTY_NOTICES.md`.

#### Scenario: Release target is buildable
- **WHEN** a release matrix runs for a supported OS/architecture
- **THEN** it produces a labeled Audivo artifact only after the platform smoke build succeeds

#### Scenario: Third-party artifact is updated
- **WHEN** an engine or runtime version changes
- **THEN** its manifest/checksum and third-party notice are updated together

### Requirement: Repository hygiene
The repository SHALL not track generated frontend output, dependency directories, downloaded engines, cookies, personal configuration, downloads, logs, or caches. Machine-specific absolute replacements SHALL not remain in module configuration.

#### Scenario: Fresh status after documented build
- **WHEN** a developer runs the documented install/build/test workflow
- **THEN** generated and personal artifacts remain ignored and no machine-specific path is required for success
