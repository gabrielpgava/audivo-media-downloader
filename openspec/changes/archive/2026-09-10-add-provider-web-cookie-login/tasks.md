## 1. Browser capability and contracts

- [x] 1.1 Validate the visible Chromium/CDP authentication approach on the supported macOS, Windows, and Linux runners, recording detected browser names, minimum versions, profile arguments, and known limitations.
- [x] 1.2 Add provider/authentication enums, public status models, safe error codes, and the `provider-auth:updated` event contract in `internal/models`.
- [x] 1.3 Define the `AuthBrowser` and `AuthManager` interfaces with fake implementations suitable for session reuse, login completion, cancellation, timeout, and browser-unavailable tests.

## 2. Provider cookie storage and configuration

- [x] 2.1 Extend `models.Settings` and the TypeScript settings contract with an optional YouTube cookie path while preserving the existing Apple Music field and JSON compatibility.
- [x] 2.2 Add platform-derived private cookie directories and per-provider profile paths with restrictive permissions.
- [x] 2.3 Implement `CookieStore` for Netscape serialization, provider-domain allowlists, readability/format validation, atomic replacement, managed copies, and local disconnect cleanup.
- [x] 2.4 Migrate valid legacy Apple Music cookie paths safely and keep invalid/relative paths out of the active configuration without exposing their contents.
- [x] 2.5 Add Go tests for serialization, permissions, atomic failure behavior, provider mismatch, malformed imports, legacy migration, and disconnect/replacement.

## 3. Visible authentication session

- [x] 3.1 Implement the supported Chromium/CDP browser locator and visible app-owned persistent profile launch for Apple Music and YouTube.
- [x] 3.2 Add provider-specific start URLs, authentication-domain allowlists, session observation, cookie extraction, and the automatic `connected` transition for an existing or newly completed session.
- [x] 3.3 Add the explicit `Verificar sessão` path for cases where automatic provider detection is inconclusive.
- [x] 3.4 Enforce one active authentication session, explicit cancellation, cleanup of temporary artifacts, timeout handling, and preservation of the last valid cookie file.
- [x] 3.5 Add fake-browser and integration-level Go tests for already-authenticated sessions, interactive login completion, blocked navigation, cancellation, unavailable browser, invalid capture, and replacement.

## 4. Backend services, engine routing, and Wails bindings

- [x] 4.1 Add service methods to start, verify, query, cancel, import, and disconnect provider authentication, and publish only redacted status events.
- [x] 4.2 Add Wails `App` bindings for the authentication methods, provider cookie selection, status lookup, and cancellation; regenerate the TypeScript bindings.
- [x] 4.3 Correct YouTube analysis/download to read only `YouTubeCookiesPath`, and keep Apple Music analysis/download restricted to `AppleMusicCookiesPath`.
- [x] 4.4 Extend `yt-dlp` metadata/download argument builders with an optional validated `--cookies` path while preserving public YouTube behavior without cookies.
- [x] 4.5 Keep `gamdl` cookie arguments provider-specific and require valid Apple Music cookies before Apple Music analysis/download.
- [x] 4.6 Update log sanitization and user-error mapping to redact both cookie options/paths and return provider-specific, actionable authentication errors.
- [x] 4.7 Add backend tests proving argument separation, optional YouTube cookies, required Apple Music cookies, redacted logs, status transitions, and concurrent-session guards.

## 5. React/MUI settings experience

- [x] 5.1 Add typed API wrappers, provider-auth event subscriptions, and frontend state for `idle`, `opening`, `waiting_login`, `connected`, `cancelled`, `unavailable`, and `error`.
- [x] 5.2 Replace the single Apple Music cookie row with a `Contas e sessões` section containing independent Apple Music and YouTube rows, status text, connect/verify/cancel actions, and clear provider identity.
- [x] 5.3 Add manual `Importar arquivo` and local `Desconectar` actions per provider, retaining the existing native file dialog behavior and surfacing validation errors inline.
- [x] 5.4 Keep unrelated settings usable while authentication is active, prevent a second auth session, and make loading/error/disabled/focus states accessible and responsive.
- [x] 5.5 Add frontend tests for provider isolation, event transitions, existing-session success, interactive-login feedback, cancellation, fallback import, disconnect, and redacted status rendering.

## 6. Documentation, acceptance, and release gates

- [x] 6.1 Update README and privacy documentation with the app-owned authentication profile, local cookie storage, manual fallback, disconnect behavior, supported browser limitations, and no-backend boundary.
- [x] 6.2 Ensure CI uses only synthetic cookie fixtures and does not launch provider authentication or persist real credentials.
- [x] 6.3 Run deterministic frontend/backend tests, typecheck/build, `go vet`, binding generation checks, and strict OpenSpec validation.
- [x] 6.4 Execute and record manual acceptance with disposable accounts for both providers: existing session, new login, MFA/CAPTCHA handoff, verify-session fallback, invalid import, wrong-provider import, cancel, disconnect, expired session, and browser unavailable.
- [x] 6.5 Execute provider download smoke cases proving YouTube cookies are used only by `yt-dlp`, Apple Music cookies only by `gamdl`, public YouTube still works without login, and no cookie material appears in logs or artifacts.
