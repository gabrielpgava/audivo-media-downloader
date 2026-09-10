package main

import (
	"context"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	backend "audivo-media-downloader/internal/app"
	"audivo-media-downloader/internal/models"
	"audivo-media-downloader/internal/notifications"
	"audivo-media-downloader/internal/platform"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Aliases keep the generated Wails bindings readable while the implementation
// remains organized under internal packages.
type DownloadRequest = models.DownloadRequest
type MediaPreview = models.MediaPreview
type DownloadJob = models.DownloadJob
type DownloadEvent = models.DownloadEvent
type DownloadResult = models.DownloadResult
type EngineStatus = models.EngineStatus
type Settings = models.SettingsView
type ProviderAuthStatus = models.ProviderAuthStatus

type App struct {
	ctx                context.Context
	service            *backend.Service
	windowFocused      atomic.Bool
	notificationsReady atomic.Bool
}

func NewApp() *App {
	app := &App{}
	app.windowFocused.Store(true)
	app.service = backend.NewService(
		func(event models.DownloadEvent) {
			if app.ctx != nil {
				runtime.EventsEmit(app.ctx, backend.DownloadUpdatedEvent, event)
				app.notify(event)
			}
		},
		func(line models.TechnicalLogLine) {
			if app.ctx != nil {
				runtime.EventsEmit(app.ctx, backend.DownloadLogEvent, line)
			}
		},
		func(status models.ProviderAuthStatus) {
			if app.ctx != nil {
				runtime.EventsEmit(app.ctx, backend.ProviderAuthUpdatedEvent, status)
			}
		},
	)
	return app
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	if err := a.service.Startup(ctx); err != nil {
		return
	}
	a.initializeNotifications()
}

func (a *App) shutdown(ctx context.Context) {
	a.notificationsReady.Store(false)
	shutdownContext, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	_ = a.service.Shutdown(shutdownContext)
	if a.ctx != nil {
		runtime.CleanupNotifications(a.ctx)
	}
}

func (a *App) SetWindowFocused(focused bool) {
	a.windowFocused.Store(focused)
}

func (a *App) initializeNotifications() {
	if a.ctx == nil || runtime.InitializeNotifications(a.ctx) != nil || !runtime.IsNotificationAvailable(a.ctx) {
		return
	}
	authorized, err := runtime.CheckNotificationAuthorization(a.ctx)
	if err == nil && authorized {
		a.notificationsReady.Store(true)
	}
}

func (a *App) notify(event models.DownloadEvent) {
	if a.ctx == nil || runtime.WindowIsMinimised(a.ctx) || !notifications.ShouldSend(event, a.windowFocused.Load(), a.notificationsReady.Load()) {
		return
	}
	content := notifications.ContentFor(event)
	_ = runtime.SendNotification(a.ctx, runtime.NotificationOptions{
		ID:    "audivo-" + event.JobID,
		Title: content.Title,
		Body:  content.Body,
	})
}

func (a *App) AnalyzeURL(rawURL string) models.AnalyzeResponse {
	return a.service.AnalyzeURL(rawURL)
}

func (a *App) StartDownload(request models.DownloadRequest) models.StartDownloadResponse {
	return a.service.StartDownload(request)
}

func (a *App) CancelDownload(jobID string) models.ActionResponse {
	return a.service.CancelDownload(jobID)
}

func (a *App) RecentDownloads(limit int) models.HistoryResponse {
	return a.service.RecentDownloads(limit)
}

func (a *App) ClearHistory() models.ActionResponse {
	return a.service.ClearHistory()
}

func (a *App) Diagnostic() models.DiagnosticResponse {
	return a.service.Diagnostic()
}

func (a *App) CopyDiagnostic() models.ActionResponse {
	if a.ctx == nil {
		return unavailableResponse()
	}
	report := a.service.Diagnostic()
	if report.Error != nil {
		return models.ActionResponse{Error: report.Error}
	}
	if err := runtime.ClipboardSetText(a.ctx, report.Report); err != nil {
		return errorResponse("clipboard_unavailable", "Não foi possível copiar o diagnóstico.", err.Error())
	}
	return models.ActionResponse{OK: true}
}

func (a *App) OpenGitHubIssues() models.ActionResponse {
	if a.ctx == nil {
		return unavailableResponse()
	}
	runtime.BrowserOpenURL(a.ctx, "https://github.com/gabrielpgava/audivo-media-downloader/issues/new")
	return models.ActionResponse{OK: true}
}

func (a *App) CheckForUpdate() models.UpdateResponse {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	checkContext, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	return a.service.CheckForUpdate(checkContext)
}

func (a *App) OpenRelease(releaseURL string) models.ActionResponse {
	if a.ctx == nil {
		return unavailableResponse()
	}
	if !isOfficialReleaseURL(releaseURL) {
		return errorResponse("release_url_invalid", "O endereço da atualização oficial é inválido.", "release URL is outside the Audivo GitHub releases path")
	}
	runtime.BrowserOpenURL(a.ctx, releaseURL)
	return models.ActionResponse{OK: true}
}

func isOfficialReleaseURL(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	return err == nil && parsed.Scheme == "https" && strings.EqualFold(parsed.Host, "github.com") && parsed.User == nil && strings.HasPrefix(parsed.Path, "/gabrielpgava/audivo-media-downloader/releases/")
}

func (a *App) GetSettings() models.SettingsView {
	return a.service.SettingsView()
}

func (a *App) SaveSettings(settings models.SettingsView) models.ActionResponse {
	return a.service.SaveSettingsView(settings)
}

func (a *App) SelectDownloadDirectory() models.ActionResponse {
	if a.ctx == nil {
		return unavailableResponse()
	}
	path, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{Title: "Escolha a pasta de downloads", DefaultDirectory: a.service.GetSettings().DownloadDirectory})
	if err != nil {
		return errorResponse("dialog_error", "Não foi possível abrir o seletor de pasta.", err.Error())
	}
	if path == "" {
		return models.ActionResponse{}
	}
	return a.service.SetDownloadDirectory(path)
}

func (a *App) SelectAppleMusicCookies() models.ActionResponse {
	return a.SelectProviderCookies(models.ServiceAppleMusic)
}

func (a *App) SelectProviderCookies(provider models.Service) models.ActionResponse {
	if a.ctx == nil {
		return unavailableResponse()
	}
	title := "Selecione o arquivo de cookies"
	if provider == models.ServiceAppleMusic {
		title = "Selecione o arquivo de cookies do Apple Music"
	} else if provider == models.ServiceYouTube {
		title = "Selecione o arquivo de cookies do YouTube"
	} else {
		return errorResponse("auth_provider_invalid", "Este provedor não possui autenticação configurável.", "unsupported provider")
	}
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{Title: title, DefaultDirectory: a.service.GetSettings().DownloadDirectory})
	if err != nil {
		return errorResponse("dialog_error", "Não foi possível abrir o seletor de arquivo.", err.Error())
	}
	if path == "" {
		return models.ActionResponse{}
	}
	return a.service.SetProviderCookies(provider, path)
}

func (a *App) StartProviderAuth(provider models.Service) models.ProviderAuthStatus {
	return a.service.StartProviderAuth(provider)
}

func (a *App) VerifyProviderAuth(provider models.Service) models.ProviderAuthStatus {
	return a.service.VerifyProviderAuth(provider)
}

func (a *App) GetProviderAuthStatus(provider models.Service) models.ProviderAuthStatus {
	return a.service.GetProviderAuthStatus(provider)
}

func (a *App) CancelProviderAuth(provider models.Service) models.ProviderAuthStatus {
	return a.service.CancelProviderAuth(provider)
}

func (a *App) DisconnectProvider(provider models.Service) models.ProviderAuthStatus {
	return a.service.DisconnectProvider(provider)
}

func (a *App) GetEngineStatus() models.EngineResponse {
	return a.service.EngineStatus()
}

func (a *App) UpdateEngines() models.EngineResponse {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	updateContext, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	return a.service.UpdateEngines(updateContext)
}

func (a *App) OpenDownloadedFile(path string) models.ActionResponse {
	if err := platform.OpenPath(path, false); err != nil {
		return errorResponse("open_file_failed", "Não foi possível abrir o arquivo.", err.Error())
	}
	return models.ActionResponse{OK: true}
}

func (a *App) OpenDownloadFolder(path string) models.ActionResponse {
	if err := platform.OpenPath(path, false); err != nil {
		return errorResponse("open_folder_failed", "Não foi possível abrir a pasta.", err.Error())
	}
	return models.ActionResponse{OK: true}
}

func unavailableResponse() models.ActionResponse {
	return errorResponse("app_not_ready", "O Audivo ainda está iniciando.", "Wails context is unavailable")
}

func errorResponse(code, message, details string) models.ActionResponse {
	return models.ActionResponse{Error: &models.UserError{Code: code, Message: message, Details: details, Retryable: true}}
}
