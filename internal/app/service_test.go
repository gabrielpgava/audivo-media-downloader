package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	providerauth "audivo-media-downloader/internal/auth"
	"audivo-media-downloader/internal/models"
	"audivo-media-downloader/internal/platform"
)

func TestSanitizeLogRedactsCookiePathAndOption(t *testing.T) {
	message := "running --cookies-path /private/apple.cookies.txt --cookies /private/youtube.cookies.txt"
	got := sanitizeLog(message, "/private/apple.cookies.txt", "/private/youtube.cookies.txt")
	if got == message || strings.Contains(got, "/private/cookies.txt") {
		t.Fatalf("cookie path was not redacted: %q", got)
	}
	if strings.Contains(got, "/private/apple.cookies.txt") || strings.Contains(got, "/private/youtube.cookies.txt") {
		t.Fatalf("cookie path was not redacted: %q", got)
	}
	if got == "" || !strings.Contains(got, "[redacted-cookie-option]") {
		t.Fatalf("cookie option was not redacted: %q", got)
	}
}

func TestNormalizeCookieSettingsMigratesLegacyAppleAndClearsInvalidFields(t *testing.T) {
	root := t.TempDir()
	store := providerauth.NewCookieStore(platform.Paths{CookieDir: filepath.Join(root, "cookies"), AuthProfileDir: filepath.Join(root, "profiles")})
	legacy := filepath.Join(root, "legacy-apple.txt")
	if err := os.WriteFile(legacy, []byte("# Netscape HTTP Cookie File\n.music.apple.com\tTRUE\t/\tTRUE\t0\tsid\tsecret\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	settings, warning, changed := normalizeCookieSettings(store, models.Settings{AppleMusicCookiesPath: legacy, YouTubeCookiesPath: filepath.Join(root, "missing.txt")})
	if !changed || warning == "" {
		t.Fatalf("expected migration warning/change, got changed=%t warning=%q", changed, warning)
	}
	if settings.AppleMusicCookiesPath != store.CookiePath(models.ServiceAppleMusic) {
		t.Fatalf("Apple Music path was not migrated: %+v", settings)
	}
	if settings.YouTubeCookiesPath != "" {
		t.Fatalf("invalid YouTube path was not cleared: %+v", settings)
	}
}

func TestValidateCookiesRequiresNetscapeShape(t *testing.T) {
	root := t.TempDir()
	valid := filepath.Join(root, "cookies.txt")
	if err := os.WriteFile(valid, []byte("# Netscape HTTP Cookie File\n.music.apple.com\tTRUE\t/\tTRUE\t0\tsid\tsecret\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := validateCookies(valid); err != nil {
		t.Fatal(err)
	}
	invalid := filepath.Join(root, "invalid.txt")
	if err := os.WriteFile(invalid, []byte("not a cookie export\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := validateCookies(invalid); err == nil {
		t.Fatal("expected invalid cookie format")
	}
}
