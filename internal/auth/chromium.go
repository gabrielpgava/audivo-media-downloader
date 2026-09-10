package auth

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"audivo-media-downloader/internal/models"
	"audivo-media-downloader/internal/platform"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

type ChromiumBrowser struct {
	Locator BrowserLocator
}

func NewChromiumBrowser() *ChromiumBrowser {
	return &ChromiumBrowser{Locator: SystemBrowserLocator{}}
}

func (b *ChromiumBrowser) Open(parent context.Context, provider models.Service, profileDir string) (BrowserSession, error) {
	if b == nil || b.Locator == nil {
		return nil, ErrBrowserUnavailable
	}
	startURL, err := providerStartURL(provider)
	if err != nil {
		return nil, err
	}
	executable, err := b.Locator.Locate()
	if err != nil {
		return nil, err
	}
	if err := platform.EnsureDirectory(profileDir); err != nil {
		return nil, err
	}
	if err := os.Chmod(profileDir, 0o700); err != nil {
		return nil, err
	}
	allocatorContext, allocatorCancel := chromedp.NewExecAllocator(parent,
		chromedp.ExecPath(executable),
		chromedp.UserDataDir(profileDir),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-sync", true),
		chromedp.Flag("disable-background-networking", true),
	)
	browserContext, browserCancel := chromedp.NewContext(allocatorContext)
	session := &chromiumSession{
		provider:        provider,
		ctx:             browserContext,
		browserCancel:   browserCancel,
		allocatorCancel: allocatorCancel,
	}
	chromedp.ListenTarget(browserContext, func(event interface{}) {
		frame, ok := event.(*page.EventFrameNavigated)
		if !ok || frame.Frame == nil || frame.Frame.ParentID != "" {
			return
		}
		if !allowedProviderURL(provider, frame.Frame.URL) {
			session.blockNavigation(frame.Frame.URL)
			go func() { _ = chromedp.Run(browserContext, page.StopLoading()) }()
		}
	})
	if err := chromedp.Run(browserContext, chromedp.Navigate(startURL)); err != nil {
		_ = session.Close()
		return nil, fmt.Errorf("open provider authentication: %w", err)
	}
	return session, nil
}

type chromiumSession struct {
	provider        models.Service
	ctx             context.Context
	browserCancel   context.CancelFunc
	allocatorCancel context.CancelFunc
	closeOnce       sync.Once
	mu              sync.RWMutex
	blockedURL      string
}

func (s *chromiumSession) Cookies(ctx context.Context) ([]Cookie, error) {
	s.mu.RLock()
	blockedURL := s.blockedURL
	s.mu.RUnlock()
	if blockedURL != "" {
		return nil, fmt.Errorf("provider authentication navigation was blocked: %s", blockedURL)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	cookies, err := network.GetCookies().WithURLs(providerCookieURLs(s.provider)).Do(s.ctx)
	if err != nil {
		return nil, err
	}
	result := make([]Cookie, 0, len(cookies))
	for _, cookie := range cookies {
		if cookie == nil {
			continue
		}
		expires := int64(cookie.Expires)
		if expires < 0 {
			expires = 0
		}
		result = append(result, Cookie{
			Domain:            cookie.Domain,
			IncludeSubdomains: strings.HasPrefix(cookie.Domain, "."),
			Path:              cookie.Path,
			Secure:            cookie.Secure,
			HTTPOnly:          cookie.HTTPOnly,
			Expires:           expires,
			Name:              cookie.Name,
			Value:             cookie.Value,
		})
	}
	return result, nil
}

func providerCookieURLs(provider models.Service) []string {
	switch provider {
	case models.ServiceAppleMusic:
		return []string{"https://music.apple.com/", "https://account.apple.com/", "https://idmsa.apple.com/"}
	case models.ServiceYouTube:
		return []string{"https://www.youtube.com/", "https://accounts.google.com/"}
	default:
		return nil
	}
}

func (s *chromiumSession) blockNavigation(rawURL string) {
	s.mu.Lock()
	if s.blockedURL == "" {
		s.blockedURL = rawURL
	}
	s.mu.Unlock()
}

func (s *chromiumSession) Close() error {
	var err error
	s.closeOnce.Do(func() {
		s.browserCancel()
		s.allocatorCancel()
	})
	return err
}

var _ Browser = (*ChromiumBrowser)(nil)
var _ BrowserSession = (*chromiumSession)(nil)
