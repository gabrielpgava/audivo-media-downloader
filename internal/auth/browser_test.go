package auth

import (
	"context"
	"strings"
	"testing"

	"audivo-media-downloader/internal/models"
)

func TestAllowedProviderURLKeepsAuthenticationInsideProviderDomains(t *testing.T) {
	tests := []struct {
		provider models.Service
		url      string
		allowed  bool
	}{
		{models.ServiceAppleMusic, "https://music.apple.com/", true},
		{models.ServiceAppleMusic, "https://account.apple.com/sign-in", true},
		{models.ServiceAppleMusic, "https://example.com/", false},
		{models.ServiceYouTube, "https://accounts.google.com/", true},
		{models.ServiceYouTube, "https://www.youtube.com/", true},
		{models.ServiceYouTube, "https://google.com.evil.example/", false},
		{models.ServiceYouTube, "http://www.youtube.com/", false},
	}
	for _, test := range tests {
		if got := allowedProviderURL(test.provider, test.url); got != test.allowed {
			t.Errorf("allowedProviderURL(%s, %q) = %t, want %t", test.provider, test.url, got, test.allowed)
		}
	}
}

func TestBlockedNavigationPreventsCookieCapture(t *testing.T) {
	session := &chromiumSession{provider: models.ServiceYouTube, blockedURL: "https://evil.example/"}
	if _, err := session.Cookies(context.Background()); err == nil || !strings.Contains(err.Error(), "blocked") {
		t.Fatalf("expected blocked navigation error, got %v", err)
	}
}
