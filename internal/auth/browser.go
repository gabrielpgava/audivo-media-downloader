package auth

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"audivo-media-downloader/internal/models"
)

var ErrBrowserUnavailable = errors.New("no supported Chromium browser was found")

// Browser opens a visible, app-owned provider session. The interface keeps
// provider orchestration independent from the concrete CDP implementation.
type Browser interface {
	Open(context.Context, models.Service, string) (BrowserSession, error)
}

// AuthBrowser is the domain name used by the service contract. Browser is
// kept as the shorter implementation-facing name so test doubles remain easy
// to read.
type AuthBrowser = Browser

type BrowserSession interface {
	Cookies(context.Context) ([]Cookie, error)
	Close() error
}

type AuthBrowserSession = BrowserSession

type BrowserLocator interface {
	Locate() (string, error)
}

type SystemBrowserLocator struct{}

func (SystemBrowserLocator) Locate() (string, error) {
	for _, candidate := range browserCandidates() {
		if candidate == "" {
			continue
		}
		if filepath.IsAbs(candidate) {
			if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
				return candidate, nil
			}
			continue
		}
		if path, err := exec.LookPath(candidate); err == nil {
			return path, nil
		}
	}
	return "", ErrBrowserUnavailable
}

func browserCandidates() []string {
	common := []string{"google-chrome", "google-chrome-stable", "microsoft-edge", "brave-browser", "chromium", "chromium-browser"}
	switch runtime.GOOS {
	case "darwin":
		return append([]string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
			"/Applications/Brave Browser.app/Contents/MacOS/Brave Browser",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
		}, common...)
	case "windows":
		candidates := append([]string{}, common...)
		for _, root := range []string{os.Getenv("LOCALAPPDATA"), os.Getenv("PROGRAMFILES"), os.Getenv("PROGRAMFILES(X86)")} {
			if root == "" {
				continue
			}
			candidates = append(candidates,
				filepath.Join(root, "Google", "Chrome", "Application", "chrome.exe"),
				filepath.Join(root, "Microsoft", "Edge", "Application", "msedge.exe"),
				filepath.Join(root, "BraveSoftware", "Brave-Browser", "Application", "brave.exe"),
			)
		}
		return candidates
	default:
		return common
	}
}

func providerStartURL(provider models.Service) (string, error) {
	switch provider {
	case models.ServiceAppleMusic:
		return "https://music.apple.com/", nil
	case models.ServiceYouTube:
		return "https://www.youtube.com/", nil
	default:
		return "", fmt.Errorf("unsupported authentication provider %q", provider)
	}
}

func allowedProviderURL(provider models.Service, rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.Port() != "" {
		return false
	}
	host := strings.ToLower(strings.TrimSuffix(parsed.Hostname(), "."))
	switch provider {
	case models.ServiceAppleMusic:
		return host == "apple.com" || strings.HasSuffix(host, ".apple.com")
	case models.ServiceYouTube:
		return host == "youtube.com" || strings.HasSuffix(host, ".youtube.com") || host == "google.com" || strings.HasSuffix(host, ".google.com")
	default:
		return false
	}
}
