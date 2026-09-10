package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"audivo-media-downloader/internal/models"
)

const (
	defaultAuthPollInterval = time.Second
	defaultAuthTimeout      = 10 * time.Minute
)

type StatusPublisher func(models.ProviderAuthStatus)
type CookieCaptured func(models.Service, string) error

type AuthManager struct {
	mu           sync.Mutex
	store        *CookieStore
	browser      Browser
	publish      StatusPublisher
	onCaptured   CookieCaptured
	rootContext  context.Context
	statuses     map[models.Service]models.ProviderAuthStatus
	active       *activeSession
	pollInterval time.Duration
	authTimeout  time.Duration
}

type activeSession struct {
	provider models.Service
	id       string
	ctx      context.Context
	cancel   context.CancelFunc
	browser  BrowserSession
	done     chan struct{}
}

func NewAuthManager(store *CookieStore, browser Browser, publish StatusPublisher, onCaptured CookieCaptured) *AuthManager {
	manager := &AuthManager{
		store:        store,
		browser:      browser,
		publish:      publish,
		onCaptured:   onCaptured,
		statuses:     make(map[models.Service]models.ProviderAuthStatus),
		pollInterval: defaultAuthPollInterval,
		authTimeout:  defaultAuthTimeout,
	}
	for _, provider := range supportedProviders() {
		manager.statuses[provider] = models.ProviderAuthStatus{Provider: provider, State: models.AuthStateIdle}
	}
	return manager
}

func (m *AuthManager) Startup(ctx context.Context, settings models.Settings) {
	m.mu.Lock()
	m.rootContext = ctx
	m.mu.Unlock()
	m.SyncSettings(settings)
}

func (m *AuthManager) SyncSettings(settings models.Settings) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, provider := range supportedProviders() {
		path := cookiePathFor(provider, settings)
		status := m.statuses[provider]
		status.Provider = provider
		status.Connected = m.store != nil && m.store.IsConnected(provider, path)
		if status.Connected {
			status.State = models.AuthStateConnected
			status.Message = "Sessão conectada."
			status.Error = nil
		} else if status.State == models.AuthStateConnected {
			status.State = models.AuthStateIdle
			status.Message = "Não conectado."
			status.Error = nil
		}
		m.statuses[provider] = status
	}
}

func (m *AuthManager) Status(provider models.Service) models.ProviderAuthStatus {
	m.mu.Lock()
	defer m.mu.Unlock()
	if status, ok := m.statuses[provider]; ok {
		return status
	}
	return errorStatus(provider, "auth_provider_invalid", "Este provedor não possui autenticação configurável.", false)
}

func (m *AuthManager) Start(provider models.Service) models.ProviderAuthStatus {
	if !isSupportedProvider(provider) {
		return errorStatus(provider, "auth_provider_invalid", "Este provedor não possui autenticação configurável.", false)
	}
	m.mu.Lock()
	if m.active != nil {
		status := m.statuses[provider]
		status.Error = &models.UserError{Code: "auth_busy", Message: "Outra sessão de autenticação já está aberta.", Retryable: true}
		status.Message = status.Error.Message
		m.statuses[provider] = status
		m.mu.Unlock()
		m.emit(status)
		return status
	}
	if m.browser == nil || m.store == nil {
		status := m.statuses[provider]
		status.State = models.AuthStateUnavailable
		status.Message = "A autenticação web não está disponível nesta execução."
		status.Error = &models.UserError{Code: "auth_browser_unavailable", Message: status.Message, Retryable: true}
		m.statuses[provider] = status
		m.mu.Unlock()
		m.emit(status)
		return status
	}
	root := m.rootContext
	if root == nil {
		root = context.Background()
	}
	ctx, cancel := context.WithTimeout(root, m.authTimeout)
	active := &activeSession{provider: provider, id: newSessionID(), ctx: ctx, cancel: cancel, done: make(chan struct{})}
	m.active = active
	status := m.statuses[provider]
	status.Provider = provider
	status.State = models.AuthStateOpening
	status.Connected = m.statuses[provider].Connected
	status.SessionID = active.id
	status.Message = "Abrindo a sessão de autenticação…"
	status.Error = nil
	m.statuses[provider] = status
	m.mu.Unlock()
	m.emit(status)
	go m.run(active)
	return status
}

func (m *AuthManager) Verify(provider models.Service) models.ProviderAuthStatus {
	m.mu.Lock()
	active := m.active
	if active == nil || active.provider != provider || active.browser == nil {
		status := m.statuses[provider]
		connected := status.Connected
		m.mu.Unlock()
		if connected {
			return status
		}
		return m.Start(provider)
	}
	m.mu.Unlock()
	if err := m.capture(active); err != nil {
		status := m.Status(provider)
		status.State = models.AuthStateWaitingLogin
		status.Message = "A sessão ainda não foi validada. Conclua o login e tente novamente."
		status.Error = &models.UserError{Code: "auth_capture_invalid", Message: "A sessão ainda não foi validada. Conclua o login e tente novamente.", Retryable: true}
		m.mu.Lock()
		if m.active == active {
			m.statuses[provider] = status
		}
		m.mu.Unlock()
		m.emit(status)
		return status
	}
	if active.browser != nil {
		_ = active.browser.Close()
	}
	m.finish(active, models.AuthStateConnected, "Sessão conectada.", nil)
	return m.Status(provider)
}

func (m *AuthManager) Cancel(provider models.Service) models.ProviderAuthStatus {
	m.mu.Lock()
	active := m.active
	status := m.statuses[provider]
	if active == nil || active.provider != provider {
		m.mu.Unlock()
		return status
	}
	active.cancel()
	browser := active.browser
	m.mu.Unlock()
	if browser != nil {
		_ = browser.Close()
	}
	awaitSessionDone(active)
	return m.Status(provider)
}

func (m *AuthManager) Disconnect(provider models.Service) models.ProviderAuthStatus {
	if !isSupportedProvider(provider) {
		return errorStatus(provider, "auth_provider_invalid", "Este provedor não possui autenticação configurável.", false)
	}
	_ = m.Cancel(provider)
	if m.store != nil {
		if err := m.store.Remove(provider); err != nil {
			return errorStatus(provider, "auth_disconnect_failed", "Não foi possível remover a sessão local.", true)
		}
		if err := m.store.RemoveProfile(provider); err != nil {
			return errorStatus(provider, "auth_disconnect_failed", "Não foi possível remover a sessão local.", true)
		}
	}
	m.mu.Lock()
	status := m.statuses[provider]
	status.State = models.AuthStateIdle
	status.Connected = false
	status.SessionID = ""
	status.LastCapturedUnix = 0
	status.Message = "Não conectado."
	status.Error = nil
	m.statuses[provider] = status
	m.mu.Unlock()
	m.emit(status)
	return status
}

func (m *AuthManager) Shutdown() {
	m.mu.Lock()
	active := m.active
	m.mu.Unlock()
	if active != nil {
		_ = m.Cancel(active.provider)
	}
}

func (m *AuthManager) run(active *activeSession) {
	session, err := m.browser.Open(active.ctx, active.provider, m.store.ProfilePath(active.provider))
	if err != nil {
		if errors.Is(active.ctx.Err(), context.Canceled) {
			m.finish(active, models.AuthStateCancelled, "Autenticação cancelada.", nil)
			return
		}
		state := models.AuthStateError
		code := "auth_browser_failed"
		message := "Não foi possível abrir a sessão de autenticação."
		if errors.Is(err, ErrBrowserUnavailable) {
			state = models.AuthStateUnavailable
			code = "auth_browser_unavailable"
			message = "Nenhum navegador compatível foi encontrado. Importe um arquivo de cookies manualmente."
		}
		m.finish(active, state, message, &models.UserError{Code: code, Message: message, Retryable: true})
		return
	}
	m.mu.Lock()
	if m.active != active {
		m.mu.Unlock()
		_ = session.Close()
		return
	}
	active.browser = session
	status := m.statuses[active.provider]
	status.State = models.AuthStateWaitingLogin
	status.Message = "Faça login na janela do provedor e aguarde a captura da sessão."
	m.statuses[active.provider] = status
	m.mu.Unlock()
	m.emit(status)

	ticker := time.NewTicker(m.pollInterval)
	defer ticker.Stop()
	for {
		if err := m.capture(active); err == nil {
			_ = session.Close()
			m.finish(active, models.AuthStateConnected, "Sessão conectada.", nil)
			return
		}
		select {
		case <-active.ctx.Done():
			_ = session.Close()
			if errors.Is(active.ctx.Err(), context.DeadlineExceeded) {
				m.finish(active, models.AuthStateError, "O tempo para concluir o login terminou.", &models.UserError{Code: "auth_timeout", Message: "O tempo para concluir o login terminou.", Retryable: true})
			} else {
				m.finish(active, models.AuthStateCancelled, "Autenticação cancelada.", nil)
			}
			return
		case <-ticker.C:
		}
	}
}

func (m *AuthManager) capture(active *activeSession) error {
	m.mu.Lock()
	session := active.browser
	m.mu.Unlock()
	if session == nil {
		return errors.New("authentication browser is not ready")
	}
	cookies, err := session.Cookies(active.ctx)
	if err != nil {
		return err
	}
	if err := ValidateCookies(active.provider, cookies); err != nil {
		return err
	}
	path, err := m.store.Save(active.provider, cookies)
	if err != nil {
		return err
	}
	if m.onCaptured != nil {
		if err := m.onCaptured(active.provider, path); err != nil {
			return err
		}
	}
	return nil
}

func (m *AuthManager) finish(active *activeSession, state models.AuthState, message string, userErr *models.UserError) {
	active.cancel()
	m.mu.Lock()
	if m.active != active {
		m.mu.Unlock()
		return
	}
	status := m.statuses[active.provider]
	status.State = state
	status.Message = message
	status.Error = userErr
	status.SessionID = ""
	if state == models.AuthStateConnected {
		status.Connected = true
		status.LastCapturedUnix = time.Now().Unix()
	}
	m.statuses[active.provider] = status
	m.active = nil
	close(active.done)
	m.mu.Unlock()
	m.emit(status)
}

func awaitSessionDone(active *activeSession) {
	if active == nil || active.done == nil {
		return
	}
	select {
	case <-active.done:
	case <-time.After(2 * time.Second):
	}
}

func (m *AuthManager) emit(status models.ProviderAuthStatus) {
	if m.publish != nil {
		m.publish(status)
	}
}

func errorStatus(provider models.Service, code, message string, retryable bool) models.ProviderAuthStatus {
	return models.ProviderAuthStatus{Provider: provider, State: models.AuthStateError, Message: message, Error: &models.UserError{Code: code, Message: message, Retryable: retryable}}
}

func supportedProviders() []models.Service {
	return []models.Service{models.ServiceAppleMusic, models.ServiceYouTube}
}

func isSupportedProvider(provider models.Service) bool {
	return provider == models.ServiceAppleMusic || provider == models.ServiceYouTube
}

func cookiePathFor(provider models.Service, settings models.Settings) string {
	if provider == models.ServiceAppleMusic {
		return settings.AppleMusicCookiesPath
	}
	return settings.YouTubeCookiesPath
}

func newSessionID() string {
	var data [12]byte
	if _, err := rand.Read(data[:]); err == nil {
		return hex.EncodeToString(data[:])
	}
	return fmt.Sprintf("session-%d", time.Now().UnixNano())
}
