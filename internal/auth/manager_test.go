package auth

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"audivo-media-downloader/internal/models"
	"audivo-media-downloader/internal/platform"
)

type fakeBrowser struct {
	mu      sync.Mutex
	openErr error
	session *fakeSession
}

func (b *fakeBrowser) Open(context.Context, models.Service, string) (BrowserSession, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.openErr != nil {
		return nil, b.openErr
	}
	return b.session, nil
}

type fakeSession struct {
	mu       sync.Mutex
	cookies  []Cookie
	sequence [][]Cookie
	closed   bool
}

func (s *fakeSession) Cookies(context.Context) ([]Cookie, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil, errors.New("closed")
	}
	if len(s.sequence) > 0 {
		cookies := s.sequence[0]
		s.sequence = s.sequence[1:]
		return cookies, nil
	}
	return s.cookies, nil
}

func (s *fakeSession) Close() error {
	s.mu.Lock()
	s.closed = true
	s.mu.Unlock()
	return nil
}

func managerTestStore(t *testing.T) *CookieStore {
	t.Helper()
	root := t.TempDir()
	return NewCookieStore(platform.Paths{
		CookieDir:      filepath.Join(root, "cookies"),
		AuthProfileDir: filepath.Join(root, "profiles"),
	})
}

func waitForStatus(t *testing.T, manager *AuthManager, provider models.Service, state models.AuthState) models.ProviderAuthStatus {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		status := manager.Status(provider)
		if status.State == state {
			return status
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s, got %+v", state, manager.Status(provider))
	return models.ProviderAuthStatus{}
}

func TestAuthManagerCapturesExistingSessionAndPublishesSafeStatus(t *testing.T) {
	store := managerTestStore(t)
	fake := &fakeBrowser{session: &fakeSession{cookies: validCookies(models.ServiceYouTube)}}
	manager := NewAuthManager(store, fake, nil, nil)
	manager.pollInterval = time.Millisecond
	manager.Startup(context.Background(), models.Settings{})
	starting := manager.Start(models.ServiceYouTube)
	if starting.State != models.AuthStateOpening {
		t.Fatalf("start state = %s", starting.State)
	}
	connected := waitForStatus(t, manager, models.ServiceYouTube, models.AuthStateConnected)
	if !connected.Connected || connected.SessionID != "" {
		t.Fatalf("unsafe or incomplete status: %+v", connected)
	}
	if err := store.Validate(models.ServiceYouTube, store.CookiePath(models.ServiceYouTube)); err != nil {
		t.Fatal(err)
	}
}

func TestAuthManagerWaitsForLoginAndVerifyCompletesSession(t *testing.T) {
	store := managerTestStore(t)
	fake := &fakeBrowser{session: &fakeSession{sequence: [][]Cookie{{}, validCookies(models.ServiceAppleMusic)}}}
	manager := NewAuthManager(store, fake, nil, nil)
	manager.pollInterval = time.Hour
	manager.Startup(context.Background(), models.Settings{})
	manager.Start(models.ServiceAppleMusic)
	waitForStatus(t, manager, models.ServiceAppleMusic, models.AuthStateWaitingLogin)
	verified := manager.Verify(models.ServiceAppleMusic)
	if verified.State != models.AuthStateConnected {
		t.Fatalf("verify state = %+v", verified)
	}
}

func TestAuthManagerCancellationPreservesPreviousSession(t *testing.T) {
	store := managerTestStore(t)
	path, err := store.Save(models.ServiceYouTube, validCookies(models.ServiceYouTube))
	if err != nil {
		t.Fatal(err)
	}
	fake := &fakeBrowser{session: &fakeSession{cookies: nil}}
	manager := NewAuthManager(store, fake, nil, nil)
	manager.pollInterval = time.Millisecond
	manager.Startup(context.Background(), models.Settings{YouTubeCookiesPath: path})
	manager.Start(models.ServiceYouTube)
	waitForStatus(t, manager, models.ServiceYouTube, models.AuthStateWaitingLogin)
	manager.Cancel(models.ServiceYouTube)
	cancelled := waitForStatus(t, manager, models.ServiceYouTube, models.AuthStateCancelled)
	if !cancelled.Connected {
		t.Fatal("cancellation should preserve connected state")
	}
	if err := store.Validate(models.ServiceYouTube, path); err != nil {
		t.Fatal(err)
	}
}

func TestAuthManagerReportsUnavailableBrowserAndBusySession(t *testing.T) {
	store := managerTestStore(t)
	fake := &fakeBrowser{openErr: ErrBrowserUnavailable}
	manager := NewAuthManager(store, fake, nil, nil)
	manager.Startup(context.Background(), models.Settings{})
	manager.Start(models.ServiceAppleMusic)
	unavailable := waitForStatus(t, manager, models.ServiceAppleMusic, models.AuthStateUnavailable)
	if unavailable.Error == nil || unavailable.Error.Code != "auth_browser_unavailable" {
		t.Fatalf("unexpected unavailable status: %+v", unavailable)
	}

	blockingSession := &fakeSession{}
	manager = NewAuthManager(store, &fakeBrowser{session: blockingSession}, nil, nil)
	manager.pollInterval = time.Hour
	manager.Startup(context.Background(), models.Settings{})
	manager.Start(models.ServiceYouTube)
	waitForStatus(t, manager, models.ServiceYouTube, models.AuthStateWaitingLogin)
	busy := manager.Start(models.ServiceAppleMusic)
	if busy.Error == nil || busy.Error.Code != "auth_busy" {
		t.Fatalf("expected busy status, got %+v", busy)
	}
	manager.Cancel(models.ServiceYouTube)
}

func TestAuthManagerVerifyReportsInvalidCaptureAndAllowsRetry(t *testing.T) {
	store := managerTestStore(t)
	fake := &fakeBrowser{session: &fakeSession{sequence: [][]Cookie{{{Domain: ".example.com", Path: "/", Name: "bad", Value: "nope"}}, {{Domain: ".example.com", Path: "/", Name: "bad", Value: "nope"}}, validCookies(models.ServiceYouTube)}}}
	manager := NewAuthManager(store, fake, nil, nil)
	manager.pollInterval = time.Hour
	manager.Startup(context.Background(), models.Settings{})
	manager.Start(models.ServiceYouTube)
	waitForStatus(t, manager, models.ServiceYouTube, models.AuthStateWaitingLogin)
	failed := manager.Verify(models.ServiceYouTube)
	if failed.Error == nil || failed.Error.Code != "auth_capture_invalid" {
		t.Fatalf("expected invalid capture status, got %+v", failed)
	}
	verified := manager.Verify(models.ServiceYouTube)
	if verified.State != models.AuthStateConnected || !verified.Connected {
		t.Fatalf("expected retry to connect, got %+v", verified)
	}
}
