package download

import (
	"testing"

	"audivo-media-downloader/internal/models"
)

func TestDetectService(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want models.Service
	}{
		{name: "youtube", url: "https://www.youtube.com/watch?v=abc", want: models.ServiceYouTube},
		{name: "short youtube", url: "https://youtu.be/abc", want: models.ServiceYouTube},
		{name: "music youtube", url: "https://music.youtube.com/watch?v=abc", want: models.ServiceYouTube},
		{name: "apple music", url: "https://music.apple.com/br/album/example/1", want: models.ServiceAppleMusic},
		{name: "apple subdomain", url: "https://www.music.apple.com/us/album/example/1", want: models.ServiceAppleMusic},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, userErr := DetectService(test.url)
			if userErr != nil {
				t.Fatalf("DetectService() error = %+v", userErr)
			}
			if got != test.want {
				t.Fatalf("DetectService() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestProviderRegistryUsesExplicitDetectionResult(t *testing.T) {
	service, userErr := NewProviderRegistry().Detect("https://youtu.be/example")
	if userErr != nil || service != models.ServiceYouTube {
		t.Fatalf("registry detection = %q, %+v", service, userErr)
	}
	if service, userErr := NewProviderRegistry().Detect("https://example.com"); service != "" || userErr == nil || userErr.Code != "unsupported_url" {
		t.Fatalf("registry unsupported result = %q, %+v", service, userErr)
	}
}

func TestDetectServiceRejectsUnsupportedAndLookalikeHosts(t *testing.T) {
	for _, rawURL := range []string{
		"",
		"ftp://youtube.com/video",
		"https://youtube.com.attacker.example/video",
		"https://example.com/?url=https://youtube.com",
		"https://user:password@youtube.com/video",
	} {
		if service, userErr := DetectService(rawURL); service != "" || userErr == nil || userErr.Code != "unsupported_url" {
			t.Errorf("DetectService(%q) = service %q, error %+v; want unsupported_url", rawURL, service, userErr)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	if got := FormatDuration(3723); got != "1h 02m" {
		t.Fatalf("FormatDuration() = %q", got)
	}
	if got := FormatDuration(125); got != "2m 05s" {
		t.Fatalf("FormatDuration() = %q", got)
	}
}
