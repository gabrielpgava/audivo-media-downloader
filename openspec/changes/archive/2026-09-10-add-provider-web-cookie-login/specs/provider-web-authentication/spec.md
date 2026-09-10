## ADDED Requirements

### Requirement: Provider-specific authentication controls
As Configurações SHALL expose separate authentication controls for Apple Music and YouTube, and each control SHALL display only the connection state and an actionable message.

#### Scenario: User starts Apple Music authentication
- **WHEN** the user clicks `Conectar` for Apple Music
- **THEN** Audivo starts an Apple Music authentication session and reports that provider as `opening` or `waiting_login` without changing the YouTube state

#### Scenario: User starts YouTube authentication
- **WHEN** the user clicks `Conectar` for YouTube
- **THEN** Audivo starts a YouTube authentication session and reports that provider as `opening` or `waiting_login` without changing the Apple Music state

### Requirement: Visible provider login session
The system SHALL open a visible, app-controlled web session using an app-owned persistent profile for the selected provider, and SHALL navigate only to allowlisted provider and authentication domains.

#### Scenario: Authentication browser is available
- **WHEN** the selected provider has a supported browser runtime available
- **THEN** Audivo opens the provider's official web page in the provider's own authentication profile and exposes a waiting state to Configurações

#### Scenario: Navigation target is not allowlisted
- **WHEN** a navigation request would leave the configured provider/authentication allowlist
- **THEN** Audivo blocks the navigation and reports an authentication error without capturing cookies

### Requirement: Reuse an authenticated session
The system SHALL detect a valid authenticated session in the provider-specific app-owned profile and capture it without requiring the user to enter credentials again.

#### Scenario: Session is already authenticated
- **WHEN** the user starts authentication and the provider profile contains a valid session
- **THEN** Audivo captures and validates the provider cookies, closes or marks the web session complete, and reports `connected`

### Requirement: Interactive login completion
The system SHALL keep the visible provider page available for human login, MFA, consent, or CAPTCHA and SHALL complete capture after a valid session becomes available.

#### Scenario: User is not authenticated
- **WHEN** the provider profile has no valid session at session start
- **THEN** Audivo leaves the provider login page open, reports `waiting_login`, and tells the user to finish login in that window

#### Scenario: Login creates a valid session
- **WHEN** the user completes login and the provider session exposes valid cookies
- **THEN** Audivo serializes and validates the cookies, reports `connected`, and makes the new session available to the corresponding engine

#### Scenario: Provider does not expose a reliable automatic signal
- **WHEN** the user has completed login but automatic detection cannot confirm the session
- **THEN** Configurações offers `Verificar sessão`, and a successful verification completes the same capture flow without displaying cookie values

### Requirement: Authentication lifecycle and cancellation
The system SHALL allow at most one active provider authentication session, SHALL support explicit cancellation, and SHALL preserve the last valid session until a replacement is validated.

#### Scenario: User cancels authentication
- **WHEN** the user clicks `Cancelar` while a provider authentication session is opening or waiting
- **THEN** Audivo closes the session, removes temporary artifacts, reports `cancelled`, and leaves any previously connected cookies unchanged

#### Scenario: Second authentication is requested
- **WHEN** a provider authentication session is already active and the user tries to start another one
- **THEN** Audivo disables or rejects the second start with a recoverable busy message

### Requirement: Browser-unavailable fallback
The system SHALL report when automatic web authentication is unavailable and SHALL provide the manual cookie import action for that provider.

#### Scenario: No supported browser is detected
- **WHEN** the user starts authentication and no supported browser runtime can be located
- **THEN** Audivo reports `unavailable` with an actionable explanation and leaves `Importar arquivo` available

#### Scenario: Automatic capture fails
- **WHEN** the web session completes but the captured cookie set is missing or invalid
- **THEN** Audivo reports a provider-specific recoverable error and does not replace the last valid cookie file

### Requirement: Redacted authentication status
Authentication methods, events, UI state, and errors SHALL never contain raw cookie values, cookie headers, cookie file contents, or full command arguments.

#### Scenario: Frontend receives an authentication event
- **WHEN** the backend emits a provider authentication update
- **THEN** the payload contains only provider, state, safe message, opaque session identifier, and non-sensitive timestamp data

#### Scenario: User expands technical details
- **WHEN** the user opens technical details after authentication or download
- **THEN** no cookie value or unredacted cookie path appears in the rendered log
