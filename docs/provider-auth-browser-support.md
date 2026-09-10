# Provider authentication browser support

Audivo uses a visible Chromium-family browser process with an application-owned profile for provider authentication. It does not read the user's existing Chrome, Edge, Safari, Firefox, or Brave profile.

| Environment observed | Browser | Version observed | Minimum/version policy | Result |
| --- | --- | --- | --- | --- |
| macOS arm64 development host (checked 2026-09-10) | Google Chrome | 152.0.7977.83 | No hard gate yet; current Chromium CDP runtime is required | Detected by the locator |
| macOS arm64 development host | Microsoft Edge | 147.0.3912.98 | No hard gate yet; current Chromium CDP runtime is required | Detected by the locator |
| Windows/Linux CI runners | Not available in this session | Pending | Must be recorded during runner acceptance before release | Covered by locator tests and manual release acceptance |

The browser is launched with `headless=false`, `--user-data-dir=<Audivo auth profile>`, `--no-first-run`, `--no-default-browser-check`, and `--disable-sync`. The profile is private to Audivo and separated by provider.

If no supported browser is available, the UI reports that automatic authentication cannot start and keeps manual Netscape cookie import available. Provider login, MFA, and CAPTCHA remain human interactions; Audivo does not automate credentials or bypass provider controls.
