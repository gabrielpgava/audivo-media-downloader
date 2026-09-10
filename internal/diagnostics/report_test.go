package diagnostics

import (
	"os"
	"strings"
	"testing"

	"audivo-media-downloader/internal/models"
)

func TestBuildRedactsSecretsAndHomePaths(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	report := Build(Input{
		Version:   "0.1.0",
		Providers: map[models.Service]string{models.ServiceYouTube: "connected"},
		Engines:   []models.EngineStatus{{Name: "yt-dlp", Valid: true, Path: home + "/.cache/audivo/yt-dlp"}},
		Error:     "token=abc123 --cookies-path " + home + "/cookies.txt",
		Paths:     []string{home + "/cookies.txt"},
	})
	if strings.Contains(report, "abc123") || strings.Contains(report, "cookies.txt") {
		t.Fatalf("report leaked sensitive data: %s", report)
	}
	if !strings.Contains(report, "~/") {
		t.Fatalf("report did not normalize home path: %s", report)
	}
}
