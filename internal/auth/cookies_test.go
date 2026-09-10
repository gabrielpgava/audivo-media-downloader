package auth

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"audivo-media-downloader/internal/models"
	"audivo-media-downloader/internal/platform"
)

func testPaths(root string) platform.Paths {
	return platform.Paths{
		ConfigDir:      filepath.Join(root, "config"),
		CookieDir:      filepath.Join(root, "config", "cookies"),
		AuthProfileDir: filepath.Join(root, "config", "auth-profiles"),
	}
}

func validCookies(provider models.Service) []Cookie {
	domain := ".youtube.com"
	if provider == models.ServiceAppleMusic {
		domain = ".music.apple.com"
	}
	return []Cookie{{Domain: domain, IncludeSubdomains: true, Path: "/", Secure: true, HTTPOnly: true, Name: "session", Value: "synthetic"}}
}

func TestCookieStoreSavesPrivateNetscapeFileAtomically(t *testing.T) {
	root := t.TempDir()
	store := NewCookieStore(testPaths(root))
	path, err := store.Save(models.ServiceYouTube, validCookies(models.ServiceYouTube))
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o077 != 0 {
		t.Fatalf("cookie permissions = %o, want private", info.Mode().Perm())
	}
	dirInfo, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	if dirInfo.Mode().Perm()&0o077 != 0 {
		t.Fatalf("cookie directory permissions = %o, want private", dirInfo.Mode().Perm())
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(data), netscapeHeader) || !strings.Contains(string(data), "#HttpOnly_.youtube.com") {
		t.Fatalf("unexpected Netscape file: %q", string(data))
	}
	if err := store.Validate(models.ServiceYouTube, path); err != nil {
		t.Fatal(err)
	}
}

func TestCookieStoreRejectsWrongProviderAndMalformedRecords(t *testing.T) {
	root := t.TempDir()
	store := NewCookieStore(testPaths(root))
	wrong := filepath.Join(root, "wrong.txt")
	if err := os.WriteFile(wrong, []byte(netscapeHeader+".youtube.com\tTRUE\t/\tTRUE\t0\tsession\tsynthetic\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Import(models.ServiceAppleMusic, wrong); err == nil {
		t.Fatal("expected wrong provider to be rejected")
	}
	bad := filepath.Join(root, "bad.txt")
	if err := os.WriteFile(bad, []byte(netscapeHeader+"not-a-cookie\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := store.Validate(models.ServiceYouTube, bad); err == nil {
		t.Fatal("expected malformed cookie file to be rejected")
	}
}

func TestCookieStoreImportsIntoManagedPathAndDisconnects(t *testing.T) {
	root := t.TempDir()
	store := NewCookieStore(testPaths(root))
	source := filepath.Join(root, "source.txt")
	if err := os.WriteFile(source, serializeNetscape(validCookies(models.ServiceAppleMusic)), 0o600); err != nil {
		t.Fatal(err)
	}
	managed, err := store.Import(models.ServiceAppleMusic, source)
	if err != nil {
		t.Fatal(err)
	}
	if managed != store.CookiePath(models.ServiceAppleMusic) {
		t.Fatalf("managed path = %q, want %q", managed, store.CookiePath(models.ServiceAppleMusic))
	}
	if err := store.Remove(models.ServiceAppleMusic); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(managed); !os.IsNotExist(err) {
		t.Fatalf("managed cookie still exists, stat err = %v", err)
	}
	profile := store.ProfilePath(models.ServiceAppleMusic)
	if err := os.MkdirAll(profile, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(profile, "profile-marker"), []byte("synthetic"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := store.RemoveProfile(models.ServiceAppleMusic); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(profile); !os.IsNotExist(err) {
		t.Fatalf("managed auth profile still exists, stat err = %v", err)
	}
}

func TestCookieStoreInvalidReplacementPreservesPreviousFile(t *testing.T) {
	root := t.TempDir()
	store := NewCookieStore(testPaths(root))
	path, err := store.Save(models.ServiceYouTube, validCookies(models.ServiceYouTube))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Save(models.ServiceYouTube, []Cookie{{Domain: ".apple.com", Path: "/", Name: "wrong", Value: "secret"}}); err == nil {
		t.Fatal("expected wrong-provider replacement to fail")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "session\tsynthetic") {
		t.Fatalf("previous cookie file was replaced after failed capture: %q", string(data))
	}
}
