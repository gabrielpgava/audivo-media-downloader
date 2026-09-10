package download

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"audivo-media-downloader/internal/models"
)

// ProviderRegistry keeps detection in one pipeline while leaving provider
// adapters free to own their engine-specific download logic.
type ProviderRegistry struct{}

func NewProviderRegistry() ProviderRegistry { return ProviderRegistry{} }

func (ProviderRegistry) Detect(rawURL string) (models.Service, *models.UserError) {
	return DetectService(rawURL)
}

// DetectService validates the URL shape and returns the supported upstream
// service. Host matching uses label boundaries so lookalike hosts such as
// youtube.com.attacker.example cannot pass validation.
func DetectService(rawURL string) (models.Service, *models.UserError) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed.Scheme == "" || parsed.Hostname() == "" {
		return "", unsupportedURLError()
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", unsupportedURLError()
	}
	if parsed.User != nil {
		return "", unsupportedURLError()
	}

	host := strings.ToLower(strings.TrimSuffix(parsed.Hostname(), "."))
	switch {
	case host == "youtu.be", host == "youtube.com", strings.HasSuffix(host, ".youtube.com"):
		return models.ServiceYouTube, nil
	case host == "music.apple.com", strings.HasSuffix(host, ".music.apple.com"):
		return models.ServiceAppleMusic, nil
	default:
		return "", unsupportedURLError()
	}
}

// NormalizeInput is deliberately conservative: it trims user paste noise and
// normalizes only the host casing, leaving provider-specific query parameters
// intact because they can change the selected media.
func NormalizeInput(rawURL string) string {
	rawURL = strings.TrimSpace(strings.Trim(rawURL, "<>"))
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Hostname() == "" {
		return rawURL
	}
	parsed.Host = strings.ToLower(parsed.Host)
	return parsed.String()
}

func unsupportedURLError() *models.UserError {
	return &models.UserError{
		Code:      "unsupported_url",
		Message:   "Cole um link do YouTube ou Apple Music.",
		Retryable: false,
	}
}

func ValidateOutputDirectory(path string) *models.UserError {
	if path == "" {
		return &models.UserError{Code: "output_directory_missing", Message: "Escolha uma pasta para salvar o arquivo.", Retryable: false}
	}
	if !filepath.IsAbs(path) {
		return &models.UserError{Code: "output_directory_not_absolute", Message: "A pasta de destino precisa ser um caminho absoluto.", Retryable: false}
	}
	return nil
}

func FormatDuration(seconds int) string {
	if seconds <= 0 {
		return ""
	}
	minutes := seconds / 60
	remaining := seconds % 60
	if minutes >= 60 {
		return fmt.Sprintf("%dh %02dm", minutes/60, minutes%60)
	}
	return fmt.Sprintf("%dm %02ds", minutes, remaining)
}
