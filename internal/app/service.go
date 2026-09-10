package app

import (
	"bufio"
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	providerauth "audivo-media-downloader/internal/auth"
	"audivo-media-downloader/internal/config"
	"audivo-media-downloader/internal/diagnostics"
	"audivo-media-downloader/internal/download"
	"audivo-media-downloader/internal/engines"
	"audivo-media-downloader/internal/history"
	"audivo-media-downloader/internal/logstore"
	"audivo-media-downloader/internal/models"
	"audivo-media-downloader/internal/platform"
	"audivo-media-downloader/internal/process"
	"audivo-media-downloader/internal/updates"
	"audivo-media-downloader/internal/version"
)

const (
	DownloadUpdatedEvent     = "download:updated"
	DownloadLogEvent         = "download:log"
	ProviderAuthUpdatedEvent = "provider-auth:updated"
)

type EventPublisher func(models.DownloadEvent)
type LogPublisher func(models.TechnicalLogLine)
type AuthStatusPublisher func(models.ProviderAuthStatus)

type Service struct {
	mu              sync.RWMutex
	ctx             context.Context
	paths           platform.Paths
	settings        models.Settings
	store           *config.Store
	history         *history.Store
	logStore        *logstore.Store
	updateChecker   *updates.Checker
	engines         *engines.Manager
	jobs            *download.JobManager
	registry        download.ProviderRegistry
	runner          *process.Runner
	cookies         *providerauth.CookieStore
	auth            *providerauth.AuthManager
	onEvent         EventPublisher
	onLog           LogPublisher
	onAuth          AuthStatusPublisher
	settingsWarning string
	readyErr        error
}

func NewService(onEvent EventPublisher, onLog LogPublisher, authPublishers ...AuthStatusPublisher) *Service {
	var onAuth AuthStatusPublisher
	if len(authPublishers) > 0 {
		onAuth = authPublishers[0]
	}
	return &Service{onEvent: onEvent, onLog: onLog, onAuth: onAuth, runner: process.NewRunner(), registry: download.NewProviderRegistry()}
}

func (s *Service) Startup(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ctx = ctx
	paths, err := platform.ResolvePaths()
	if err != nil {
		s.readyErr = err
		return err
	}
	if err := platform.EnsureDirectory(paths.TempDir); err != nil {
		s.readyErr = err
		return err
	}
	s.paths = paths
	s.store = config.NewStore(paths)
	s.history = history.NewStore(paths)
	s.logStore = logstore.New(filepath.Join(paths.ConfigDir, "audivo.log"), 1<<20, 3)
	s.updateChecker = updates.NewChecker(filepath.Join(paths.ConfigDir, "update.json"))
	settings, loadErr := s.store.Load()
	if loadErr != nil {
		settings = s.store.Defaults()
	}
	s.settingsWarning = ""
	if loadErr != nil {
		s.settingsWarning = "Uma configuração salva foi ignorada e voltou aos valores seguros."
	}
	s.cookies = providerauth.NewCookieStore(paths)
	var migrated bool
	settings, migrationWarning, changed := normalizeCookieSettings(s.cookies, settings)
	if migrationWarning != "" {
		s.settingsWarning = migrationWarning
	}
	if changed {
		if err := s.store.Save(settings); err != nil {
			s.settingsWarning = "Uma sessão salva não pôde ser normalizada; importe-a novamente em Configurações."
		} else {
			migrated = true
		}
	}
	s.settings = settings
	s.auth = providerauth.NewAuthManager(s.cookies, providerauth.NewChromiumBrowser(), s.publishAuth, s.persistCapturedCookie)
	s.auth.Startup(ctx, settings)
	if migrated {
		s.settings = settings
	}
	manifest := engines.Manifest{}
	manifestPath := filepath.Join(paths.ConfigDir, "engines.json")
	if loaded, err := engines.LoadManifest(manifestPath); err == nil {
		manifest = loaded
	}
	s.engines = engines.NewManager(paths, manifest)
	s.jobs = download.NewJobManager(s.publishEvent, s.publishLog)
	s.readyErr = nil
	if s.engines.CheckDue(engineCheckTime(settings.LastEngineCheckUnix), time.Now()) {
		go s.runScheduledEngineCheck(ctx)
	}
	go s.runScheduledUpdateCheck(ctx)
	return nil
}

func (s *Service) Shutdown(ctx context.Context) error {
	s.mu.RLock()
	jobs := s.jobs
	authManager := s.auth
	s.mu.RUnlock()
	if authManager != nil {
		authManager.Shutdown()
	}
	if jobs == nil {
		return nil
	}
	return jobs.Shutdown(ctx)
}

func (s *Service) AnalyzeURL(rawURL string) models.AnalyzeResponse {
	if err := s.ready(); err != nil {
		return models.AnalyzeResponse{Error: download.AsUserError(err)}
	}
	rawURL = download.NormalizeInput(rawURL)
	service, userErr := s.registry.Detect(rawURL)
	if userErr != nil {
		return models.AnalyzeResponse{Error: userErr}
	}
	ctx := s.currentContext()
	switch service {
	case models.ServiceYouTube:
		path, err := s.engines.Ensure(ctx, engines.EngineYTDLP)
		if err != nil {
			return models.AnalyzeResponse{Error: engineError(err)}
		}
		deno, err := s.engines.Ensure(ctx, engines.EngineDeno)
		if err != nil {
			return models.AnalyzeResponse{Error: engineError(err)}
		}
		cookiePath := s.optionalProviderCookiePath(models.ServiceYouTube)
		var stdout []string
		err = s.runner.Run(ctx, path, engines.BuildYTDLPMetadataArgsWithCookies(rawURL, deno, cookiePath), "", func(line process.Line) {
			if line.Stream == "stdout" {
				stdout = append(stdout, line.Message)
			} else {
				s.publishLog(models.TechnicalLogLine{Stream: line.Stream, Message: sanitizeLog(line.Message, cookiePath)})
			}
		})
		if err != nil {
			return models.AnalyzeResponse{Error: engineError(err)}
		}
		preview, err := engines.ParseYTDLPMetadata([]byte(strings.Join(stdout, "\n")), rawURL)
		if err != nil {
			return models.AnalyzeResponse{Error: engineError(err)}
		}
		return s.analysisResponse(&preview)
	case models.ServiceAppleMusic:
		cookies, err := s.requiredProviderCookiePath(models.ServiceAppleMusic)
		if err != nil {
			return models.AnalyzeResponse{Error: providerCookieError(models.ServiceAppleMusic)}
		}
		if _, err := s.engines.Ensure(ctx, engines.EngineGamdl); err != nil {
			return models.AnalyzeResponse{Error: engineError(err)}
		}
		environment, err := engines.ResolveGamdlEnvironment(s.paths)
		if err != nil {
			return models.AnalyzeResponse{Error: engineError(err)}
		}
		helper, err := engines.WriteGamdlMetadataHelper(s.paths)
		if err != nil {
			return models.AnalyzeResponse{Error: engineError(err)}
		}
		args, err := engines.BuildGamdlMetadataArgs(helper, rawURL, cookies)
		if err != nil {
			return models.AnalyzeResponse{Error: engineError(err)}
		}
		var stdout []string
		err = s.runner.Run(ctx, environment.Python, args, s.paths.TempDir, func(line process.Line) {
			if line.Stream == "stdout" {
				stdout = append(stdout, line.Message)
			} else {
				s.publishLog(models.TechnicalLogLine{Stream: line.Stream, Message: sanitizeLog(line.Message, cookies)})
			}
		})
		if err != nil {
			return models.AnalyzeResponse{Error: engineError(err)}
		}
		preview, err := engines.ParseGamdlMetadata([]byte(strings.Join(stdout, "\n")), rawURL)
		if err != nil {
			return models.AnalyzeResponse{Error: engineError(err)}
		}
		return s.analysisResponse(&preview)
	default:
		return models.AnalyzeResponse{Error: &models.UserError{Code: "unsupported_url", Message: "Cole um link do YouTube ou Apple Music.", Retryable: false}}
	}
}

func (s *Service) analysisResponse(preview *models.MediaPreview) models.AnalyzeResponse {
	response := models.AnalyzeResponse{Preview: preview}
	if preview != nil && s.history != nil {
		if duplicate, err := s.history.Find(preview.URL); err == nil {
			response.Duplicate = duplicate
		}
	}
	return response
}

func (s *Service) StartDownload(request models.DownloadRequest) models.StartDownloadResponse {
	if err := s.ready(); err != nil {
		return models.StartDownloadResponse{Error: download.AsUserError(err)}
	}
	request.URL = download.NormalizeInput(request.URL)
	if request.OutputDir == "" {
		request.OutputDir = s.GetSettings().DownloadDirectory
	}
	if err := prepareOutputDirectory(request.OutputDir); err != nil {
		return models.StartDownloadResponse{Error: &models.UserError{Code: "output_directory_invalid", Message: "Escolha uma pasta de destino válida e gravável.", Details: err.Error(), Retryable: false}}
	}
	service, userErr := s.registry.Detect(request.URL)
	if userErr != nil {
		return models.StartDownloadResponse{Error: userErr}
	}
	request.OutputDir = filepath.Clean(request.OutputDir)
	job, startErr := s.jobs.Start(s.currentContext(), func(ctx context.Context, jobID string, emit func(models.DownloadEvent), log func(models.TechnicalLogLine)) (models.DownloadResult, error) {
		tempDir, tempErr := download.NewTempWorkspace(s.paths.TempDir, jobID)
		if tempErr != nil {
			return models.DownloadResult{}, download.NewDetailedError("temp_unavailable", "Não foi possível preparar o espaço temporário.", tempErr.Error(), true)
		}
		before := snapshotMediaFiles(request.OutputDir)
		emit(models.DownloadEvent{JobID: jobID, State: models.StatePreparing, Stage: models.StagePreparing, Message: "Preparando o download…"})
		var result models.DownloadResult
		var err error
		if request.CollectionTitle != "" || len(request.CollectionItems) > 0 {
			result, err = s.downloadCollection(ctx, jobID, request, tempDir, emit, log)
		} else {
			result, err = s.downloadOne(ctx, jobID, request, before, tempDir, emit, log)
		}
		if err == nil || result.Partial {
			s.recordHistory(request, service, result)
		}
		if cleanupErr := download.FinalizeTempWorkspace(tempDir, tempWorkspaceOutcome(result, err)); cleanupErr != nil {
			log(models.TechnicalLogLine{JobID: jobID, Stream: "audivo", Message: "limpeza do temporário: " + cleanupErr.Error()})
		}
		return result, err
	})
	if startErr != nil {
		return models.StartDownloadResponse{Error: startErr}
	}
	return models.StartDownloadResponse{Job: job}
}

func (s *Service) downloadYouTube(ctx context.Context, jobID string, request models.DownloadRequest, before map[string]struct{}, emit func(models.DownloadEvent), log func(models.TechnicalLogLine)) (models.DownloadResult, error) {
	deps, err := s.engines.ResolveYouTubeDependencies(ctx)
	if err != nil {
		return models.DownloadResult{}, engineFailure(err)
	}
	cookiePath := s.optionalProviderCookiePath(models.ServiceYouTube)
	args, err := engines.BuildYTDLPDownloadArgsWithCookies(request, request.OutputDir, deps.FFmpeg, deps.Deno, cookiePath)
	if err != nil {
		return models.DownloadResult{}, err
	}
	emit(models.DownloadEvent{JobID: jobID, State: models.StateDownloading, Stage: models.StageDownloading, Message: "Baixando…"})
	err = s.runner.Run(ctx, deps.YTDLP, args, request.OutputDir, func(line process.Line) {
		if event, ok := engines.ParseYTDLPProgress(line.Message); ok {
			event.JobID = jobID
			emit(event)
			return
		}
		log(models.TechnicalLogLine{JobID: jobID, Stream: line.Stream, Message: sanitizeLog(line.Message, cookiePath)})
	})
	if err != nil {
		return models.DownloadResult{}, engineFailure(err)
	}
	return models.DownloadResult{Files: newMediaFiles(request.OutputDir, before), Directory: request.OutputDir}, nil
}

func (s *Service) downloadAppleMusic(ctx context.Context, jobID string, request models.DownloadRequest, before map[string]struct{}, tempDir string, emit func(models.DownloadEvent), log func(models.TechnicalLogLine)) (models.DownloadResult, error) {
	cookiePath, cookieErr := s.requiredProviderCookiePath(models.ServiceAppleMusic)
	if cookieErr != nil {
		return models.DownloadResult{}, download.NewError("cookies_missing", "Configure a sessão do Apple Music em Configurações.", false)
	}
	if _, err := s.engines.Ensure(ctx, engines.EngineGamdl); err != nil {
		return models.DownloadResult{}, engineFailure(err)
	}
	environment, err := engines.ResolveGamdlEnvironment(s.paths)
	if err != nil {
		return models.DownloadResult{}, engineFailure(err)
	}
	ffmpeg, ffmpegErr := s.engines.Resolve(engines.EngineFFmpeg)
	if request.MediaType == models.MediaTypeVideo || (request.MediaType == models.MediaTypeAudio && request.Format == models.FormatMP3) {
		if ffmpegErr != nil {
			return models.DownloadResult{}, engineFailure(ffmpegErr)
		}
	}
	if tempDir == "" {
		tempDir = s.paths.TempDir
	}
	args, err := engines.BuildGamdlArgs(request, request.OutputDir, tempDir, cookiePath, ffmpeg)
	if err != nil {
		return models.DownloadResult{}, err
	}
	emit(models.DownloadEvent{JobID: jobID, State: models.StateDownloading, Stage: models.StageDownloading, Message: "Baixando do Apple Music…"})
	err = s.runner.Run(ctx, environment.Gamdl, args, request.OutputDir, func(line process.Line) {
		message := strings.TrimSpace(line.Message)
		safeMessage := sanitizeLog(message, cookiePath)
		if strings.Contains(strings.ToLower(safeMessage), "download") {
			emit(models.DownloadEvent{JobID: jobID, State: models.StateDownloading, Stage: models.StageDownloading, Message: safeMessage})
		}
		log(models.TechnicalLogLine{JobID: jobID, Stream: line.Stream, Message: safeMessage})
	})
	if err != nil {
		return models.DownloadResult{}, engineFailure(err)
	}
	files := newMediaFiles(request.OutputDir, before)
	if request.MediaType == models.MediaTypeAudio && request.Format == models.FormatMP3 {
		if ffmpeg == "" {
			return models.DownloadResult{}, download.NewError("ffmpeg_missing", "O FFmpeg é necessário para converter este áudio para MP3.", true)
		}
		emit(models.DownloadEvent{JobID: jobID, State: models.StateProcessing, Stage: models.StageConverting, Message: "Convertendo para MP3…"})
		converted := make([]string, 0, len(files))
		for _, input := range files {
			if filepath.Ext(input) != ".m4a" && filepath.Ext(input) != ".aac" {
				continue
			}
			output := strings.TrimSuffix(input, filepath.Ext(input)) + ".mp3"
			conversionArgs, buildErr := engines.BuildMP3ConversionArgs(input, output, ffmpeg)
			if buildErr != nil {
				return models.DownloadResult{}, buildErr
			}
			if runErr := s.runner.Run(ctx, ffmpeg, conversionArgs, request.OutputDir, func(line process.Line) {
				log(models.TechnicalLogLine{JobID: jobID, Stream: line.Stream, Message: sanitizeLog(line.Message, cookiePath)})
			}); runErr != nil {
				return models.DownloadResult{}, engineFailure(runErr)
			}
			_ = os.Remove(input)
			converted = append(converted, output)
		}
		if len(converted) > 0 {
			files = converted
		}
	}
	return models.DownloadResult{Files: files, Directory: request.OutputDir}, nil
}

func (s *Service) CancelDownload(jobID string) models.ActionResponse {
	if err := s.ready(); err != nil {
		return models.ActionResponse{Error: download.AsUserError(err)}
	}
	if userErr := s.jobs.Cancel(jobID); userErr != nil {
		return models.ActionResponse{Error: userErr}
	}
	return models.ActionResponse{OK: true}
}

func (s *Service) GetSettings() models.Settings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.settings
}

func (s *Service) SettingsView() models.SettingsView {
	settings := s.GetSettings()
	view := models.SettingsView{
		DownloadDirectory:   settings.DownloadDirectory,
		Theme:               settings.Theme,
		AudioFormat:         settings.AudioFormat,
		VideoQuality:        settings.VideoQuality,
		QuickMode:           settings.QuickMode,
		LastEngineCheckUnix: settings.LastEngineCheckUnix,
	}
	s.mu.RLock()
	view.Warning = s.settingsWarning
	authManager := s.auth
	cookieStore := s.cookies
	s.mu.RUnlock()
	if authManager != nil {
		view.AppleMusicConnected = authManager.Status(models.ServiceAppleMusic).Connected
		view.YouTubeConnected = authManager.Status(models.ServiceYouTube).Connected
	} else if cookieStore != nil {
		view.AppleMusicConnected = cookieStore.IsConnected(models.ServiceAppleMusic, settings.AppleMusicCookiesPath)
		view.YouTubeConnected = cookieStore.IsConnected(models.ServiceYouTube, settings.YouTubeCookiesPath)
	}
	return view
}

func (s *Service) SaveSettings(settings models.Settings) models.ActionResponse {
	if err := s.ready(); err != nil {
		return models.ActionResponse{Error: download.AsUserError(err)}
	}
	if err := config.Validate(settings); err != nil {
		return models.ActionResponse{Error: &models.UserError{Code: "settings_invalid", Message: "Revise as configurações informadas.", Details: err.Error(), Retryable: false}}
	}
	if err := s.store.Save(settings); err != nil {
		return models.ActionResponse{Error: download.AsUserError(err)}
	}
	s.mu.Lock()
	s.settings = settings
	s.mu.Unlock()
	if authManager := s.authManager(); authManager != nil {
		authManager.SyncSettings(settings)
	}
	return models.ActionResponse{OK: true}
}

func (s *Service) SaveSettingsView(view models.SettingsView) models.ActionResponse {
	settings := s.GetSettings()
	settings.DownloadDirectory = view.DownloadDirectory
	settings.Theme = view.Theme
	settings.AudioFormat = view.AudioFormat
	settings.VideoQuality = view.VideoQuality
	settings.QuickMode = view.QuickMode
	return s.SaveSettings(settings)
}

func (s *Service) SetDownloadDirectory(path string) models.ActionResponse {
	if err := prepareOutputDirectory(path); err != nil {
		return models.ActionResponse{Error: &models.UserError{Code: "output_directory_invalid", Message: "Escolha uma pasta de destino válida e gravável.", Details: err.Error(), Retryable: false}}
	}
	settings := s.GetSettings()
	settings.DownloadDirectory = filepath.Clean(path)
	return s.SaveSettings(settings)
}

func (s *Service) SetAppleMusicCookiesPath(path string) models.ActionResponse {
	return s.SetProviderCookies(models.ServiceAppleMusic, path)
}

func (s *Service) SetProviderCookies(provider models.Service, path string) models.ActionResponse {
	if err := s.ready(); err != nil {
		return models.ActionResponse{Error: download.AsUserError(err)}
	}
	if !isSupportedAuthProvider(provider) {
		return models.ActionResponse{Error: &models.UserError{Code: "auth_provider_invalid", Message: "Este provedor não possui autenticação configurável.", Retryable: false}}
	}
	if s.cookies == nil {
		return models.ActionResponse{Error: &models.UserError{Code: "auth_unavailable", Message: "O armazenamento de sessões não está disponível.", Retryable: true}}
	}
	managedPath, err := s.cookies.Import(provider, path)
	if err != nil {
		return models.ActionResponse{Error: &models.UserError{Code: providerCookieImportCode(provider), Message: providerCookieImportMessage(provider), Retryable: false}}
	}
	settings := s.GetSettings()
	setProviderCookiePath(&settings, provider, managedPath)
	return s.SaveSettings(settings)
}

func (s *Service) DisconnectProvider(provider models.Service) models.ProviderAuthStatus {
	if err := s.ready(); err != nil {
		return providerAuthError(provider, "auth_unavailable", "O armazenamento de sessões não está disponível.", true)
	}
	if s.auth == nil {
		return providerAuthError(provider, "auth_unavailable", "A autenticação web não está disponível nesta execução.", true)
	}
	status := s.auth.Disconnect(provider)
	if status.Error != nil {
		return status
	}
	settings := s.GetSettings()
	setProviderCookiePath(&settings, provider, "")
	if response := s.SaveSettings(settings); response.Error != nil {
		status.State = models.AuthStateError
		status.Message = "A sessão foi removida, mas as configurações não puderam ser atualizadas."
		status.Error = response.Error
	}
	return status
}

func (s *Service) StartProviderAuth(provider models.Service) models.ProviderAuthStatus {
	if err := s.ready(); err != nil {
		return providerAuthError(provider, "auth_unavailable", "A autenticação ainda não está pronta.", true)
	}
	if s.auth == nil {
		return providerAuthError(provider, "auth_unavailable", "A autenticação web não está disponível nesta execução.", true)
	}
	return s.auth.Start(provider)
}

func (s *Service) VerifyProviderAuth(provider models.Service) models.ProviderAuthStatus {
	if err := s.ready(); err != nil {
		return providerAuthError(provider, "auth_unavailable", "A autenticação ainda não está pronta.", true)
	}
	if s.auth == nil {
		return providerAuthError(provider, "auth_unavailable", "A autenticação web não está disponível nesta execução.", true)
	}
	return s.auth.Verify(provider)
}

func (s *Service) GetProviderAuthStatus(provider models.Service) models.ProviderAuthStatus {
	if s.auth == nil {
		return providerAuthError(provider, "auth_unavailable", "A autenticação web não está disponível nesta execução.", true)
	}
	return s.auth.Status(provider)
}

func (s *Service) CancelProviderAuth(provider models.Service) models.ProviderAuthStatus {
	if s.auth == nil {
		return providerAuthError(provider, "auth_unavailable", "A autenticação web não está disponível nesta execução.", true)
	}
	return s.auth.Cancel(provider)
}

func (s *Service) authManager() *providerauth.AuthManager {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.auth
}

func (s *Service) persistCapturedCookie(provider models.Service, path string) error {
	settings := s.GetSettings()
	setProviderCookiePath(&settings, provider, path)
	response := s.SaveSettings(settings)
	if response.Error != nil {
		return errors.New(response.Error.Message)
	}
	return nil
}

func normalizeCookieSettings(store *providerauth.CookieStore, settings models.Settings) (models.Settings, string, bool) {
	if store == nil {
		return settings, "", false
	}
	warning := ""
	changed := false
	for _, provider := range []models.Service{models.ServiceAppleMusic, models.ServiceYouTube} {
		path := providerCookiePath(settings, provider)
		if path == "" {
			continue
		}
		managed := store.CookiePath(provider)
		if filepath.Clean(path) == filepath.Clean(managed) {
			if err := store.Validate(provider, managed); err != nil {
				setProviderCookiePath(&settings, provider, "")
				warning = "Uma sessão salva não pôde ser validada e foi desconectada localmente."
				changed = true
			}
			continue
		}
		if imported, err := store.Import(provider, path); err == nil {
			setProviderCookiePath(&settings, provider, imported)
			changed = true
			continue
		}
		setProviderCookiePath(&settings, provider, "")
		warning = "Uma sessão salva não pôde ser importada e foi desconectada localmente."
		changed = true
	}
	return settings, warning, changed
}

func providerCookiePath(settings models.Settings, provider models.Service) string {
	if provider == models.ServiceAppleMusic {
		return settings.AppleMusicCookiesPath
	}
	if provider == models.ServiceYouTube {
		return settings.YouTubeCookiesPath
	}
	return ""
}

func setProviderCookiePath(settings *models.Settings, provider models.Service, path string) {
	if settings == nil {
		return
	}
	if provider == models.ServiceAppleMusic {
		settings.AppleMusicCookiesPath = path
	} else if provider == models.ServiceYouTube {
		settings.YouTubeCookiesPath = path
	}
}

func (s *Service) optionalProviderCookiePath(provider models.Service) string {
	settings := s.GetSettings()
	path := providerCookiePath(settings, provider)
	if path == "" || s.cookies == nil || s.cookies.Validate(provider, path) != nil {
		return ""
	}
	return path
}

func (s *Service) requiredProviderCookiePath(provider models.Service) (string, error) {
	settings := s.GetSettings()
	path := providerCookiePath(settings, provider)
	if path == "" || s.cookies == nil {
		return "", errors.New("provider cookies are not configured")
	}
	if err := s.cookies.Validate(provider, path); err != nil {
		return "", err
	}
	return path, nil
}

func providerCookieError(provider models.Service) *models.UserError {
	if provider == models.ServiceAppleMusic {
		return &models.UserError{Code: "cookies_missing", Message: "Conecte ou importe a sessão do Apple Music em Configurações.", Retryable: false}
	}
	return &models.UserError{Code: providerCookieImportCode(provider), Message: "A sessão do YouTube não está disponível. Conecte ou importe cookies válidos em Configurações.", Retryable: true}
}

func providerCookieImportCode(provider models.Service) string {
	if provider == models.ServiceAppleMusic {
		return "apple_music_cookies_invalid"
	}
	return "youtube_cookies_invalid"
}

func providerCookieImportMessage(provider models.Service) string {
	if provider == models.ServiceAppleMusic {
		return "Selecione um arquivo Netscape válido com cookies do Apple Music."
	}
	return "Selecione um arquivo Netscape válido com cookies do YouTube."
}

func providerAuthError(provider models.Service, code, message string, retryable bool) models.ProviderAuthStatus {
	return models.ProviderAuthStatus{
		Provider: provider,
		State:    models.AuthStateError,
		Message:  message,
		Error:    &models.UserError{Code: code, Message: message, Retryable: retryable},
	}
}

func isSupportedAuthProvider(provider models.Service) bool {
	return provider == models.ServiceAppleMusic || provider == models.ServiceYouTube
}

func (s *Service) EngineStatus() models.EngineResponse {
	if err := s.ready(); err != nil {
		return models.EngineResponse{Error: download.AsUserError(err)}
	}
	return models.EngineResponse{Engines: s.engines.Status()}
}

func (s *Service) Diagnostic() models.DiagnosticResponse {
	if err := s.ready(); err != nil {
		return models.DiagnosticResponse{Error: download.AsUserError(err)}
	}
	settings := s.GetSettings()
	view := s.SettingsView()
	providers := map[models.Service]string{
		models.ServiceYouTube:    boolStatus(view.YouTubeConnected),
		models.ServiceAppleMusic: boolStatus(view.AppleMusicConnected),
	}
	report := diagnostics.Build(diagnostics.Input{
		Version:   version.Current,
		Providers: providers,
		Engines:   s.engines.Status(),
		Operation: "idle",
		Paths:     []string{settings.YouTubeCookiesPath, settings.AppleMusicCookiesPath, s.paths.ConfigDir, s.paths.AuthProfileDir},
	})
	return models.DiagnosticResponse{Report: report}
}

func (s *Service) CheckForUpdate(ctx context.Context) models.UpdateResponse {
	if err := s.ready(); err != nil {
		return models.UpdateResponse{CurrentVersion: version.Current, Error: download.AsUserError(err)}
	}
	s.mu.RLock()
	checker := s.updateChecker
	s.mu.RUnlock()
	return checker.Check(ctx, version.Current)
}

func (s *Service) runScheduledUpdateCheck(ctx context.Context) {
	response := s.CheckForUpdate(ctx)
	if response.Available {
		s.publishLog(models.TechnicalLogLine{Stream: "updates", Message: "Atualização disponível: " + response.LatestVersion})
	}
}

func (s *Service) UpdateEngines(ctx context.Context) models.EngineResponse {
	if err := s.ready(); err != nil {
		return models.EngineResponse{Error: download.AsUserError(err)}
	}
	if s.jobs.IsActive() {
		return models.EngineResponse{Engines: s.engines.Status(), Error: &models.UserError{Code: "engine_update_busy", Message: "Aguarde o download atual terminar antes de atualizar os engines.", Retryable: false}}
	}
	if err := s.engines.Update(ctx); err != nil {
		return models.EngineResponse{Engines: s.engines.Status(), Error: engineError(err)}
	}
	settings := s.GetSettings()
	settings.LastEngineCheckUnix = time.Now().Unix()
	_ = s.store.Save(settings)
	s.mu.Lock()
	s.settings = settings
	s.mu.Unlock()
	return models.EngineResponse{Engines: s.engines.Status()}
}

func (s *Service) runScheduledEngineCheck(ctx context.Context) {
	if s.jobs.IsActive() {
		s.publishLog(models.TechnicalLogLine{Stream: "engines", Message: "Verificação automática adiada enquanto há um download ativo."})
		return
	}
	if err := s.engines.Update(ctx); err != nil {
		s.publishLog(models.TechnicalLogLine{Stream: "engines", Message: "Verificação automática dos engines: " + err.Error()})
		return
	}
	settings := s.GetSettings()
	settings.LastEngineCheckUnix = time.Now().Unix()
	if err := s.store.Save(settings); err != nil {
		s.publishLog(models.TechnicalLogLine{Stream: "engines", Message: "Não foi possível salvar a data da verificação dos engines."})
		return
	}
	s.mu.Lock()
	s.settings = settings
	s.mu.Unlock()
}

func (s *Service) ready() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.readyErr != nil {
		return s.readyErr
	}
	if s.jobs == nil || s.engines == nil {
		return errors.New("application services are not initialized")
	}
	return nil
}

func (s *Service) currentContext() context.Context {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.ctx == nil {
		return context.Background()
	}
	return s.ctx
}

func (s *Service) publishEvent(event models.DownloadEvent) {
	if s.onEvent != nil {
		s.onEvent(event)
	}
}

func (s *Service) publishLog(line models.TechnicalLogLine) {
	if s.logStore != nil {
		_ = s.logStore.Append(line.Stream, line.Message)
	}
	if s.onLog != nil {
		s.onLog(line)
	}
}

func boolStatus(connected bool) string {
	if connected {
		return "connected"
	}
	return "disconnected"
}

func (s *Service) publishAuth(status models.ProviderAuthStatus) {
	if s.onAuth != nil {
		s.onAuth(status)
	}
}

func engineError(err error) *models.UserError {
	if err == nil {
		return nil
	}
	if userErr := download.AsUserError(err); userErr.Code != "operation_failed" {
		return userErr
	}
	lower := strings.ToLower(err.Error())
	if strings.Contains(lower, "checksum") {
		return &models.UserError{Code: "engine_invalid", Message: "Um engine instalado falhou na validação. Abra Configurações e atualize os engines.", Details: err.Error(), Retryable: true}
	}
	if strings.Contains(lower, "not installed") || strings.Contains(lower, "unavailable") || strings.Contains(lower, "python") {
		return &models.UserError{Code: "engine_missing", Message: "Um engine necessário não está disponível. Abra Configurações para prepará-lo.", Details: err.Error(), Retryable: true}
	}
	return &models.UserError{Code: "engine_failed", Message: "O engine não conseguiu concluir esta operação.", Details: err.Error(), Retryable: true}
}

func engineFailure(err error) error {
	if err == nil {
		return nil
	}
	userErr := download.AsUserError(err)
	if userErr.Code != "operation_failed" {
		return &download.UserError{Value: userErr}
	}
	lower := strings.ToLower(err.Error())
	if strings.Contains(lower, "checksum") {
		return download.NewDetailedError("engine_invalid", "Um engine instalado falhou na validação.", err.Error(), true)
	}
	if strings.Contains(lower, "not installed") || strings.Contains(lower, "unavailable") || strings.Contains(lower, "python") {
		return download.NewDetailedError("engine_missing", "Um engine necessário não está disponível.", err.Error(), true)
	}
	return download.NewDetailedError("engine_failed", "O engine não conseguiu concluir esta operação.", err.Error(), true)
}

func validateCookies(path string) error {
	if path == "" || !filepath.IsAbs(path) {
		return errors.New("cookies path must be absolute")
	}
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return errors.New("cookies path is a directory")
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	validRecord := false
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		if len(strings.Split(line, "\t")) < 7 {
			return errors.New("cookies file is not in Netscape format")
		}
		validRecord = true
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if !validRecord {
		return errors.New("cookies file is empty or not in Netscape format")
	}
	return nil
}

func engineCheckTime(unix int64) time.Time {
	if unix <= 0 {
		return time.Time{}
	}
	return time.Unix(unix, 0)
}

func prepareOutputDirectory(path string) error {
	if path == "" || !filepath.IsAbs(path) {
		return errors.New("output directory must be absolute")
	}
	if err := platform.EnsureDirectory(path); err != nil {
		return err
	}
	if !platform.IsWritableDirectory(path) {
		return errors.New("output directory is not writable")
	}
	return nil
}

func snapshotMediaFiles(root string) map[string]struct{} {
	files := map[string]struct{}{}
	_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil
		}
		if isMediaFile(path) {
			files[path] = struct{}{}
		}
		return nil
	})
	return files
}

func newMediaFiles(root string, before map[string]struct{}) []string {
	files := snapshotMediaFiles(root)
	result := make([]string, 0)
	for path := range files {
		if _, existed := before[path]; !existed {
			result = append(result, path)
		}
	}
	sort.Strings(result)
	return result
}

func isMediaFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".mp3", ".m4a", ".aac", ".mp4", ".m4v", ".webm", ".mkv", ".opus":
		return true
	default:
		return false
	}
}

func sanitizeLog(message string, secretPaths ...string) string {
	for _, secretPath := range secretPaths {
		if secretPath != "" {
			message = strings.ReplaceAll(message, secretPath, "[redacted-cookie-path]")
		}
	}
	message = strings.ReplaceAll(message, "--cookies-path", "[redacted-cookie-option]")
	return strings.ReplaceAll(message, "--cookies", "[redacted-cookie-option]")
}

func tempWorkspaceOutcome(result models.DownloadResult, err error) download.TempOutcome {
	switch {
	case errors.Is(err, context.Canceled):
		return download.TempCancelled
	case result.Partial:
		return download.TempPartial
	case err != nil && download.AsUserError(err).Retryable:
		return download.TempRecoverableFailure
	case err != nil:
		return download.TempPermanentFailure
	default:
		return download.TempSucceeded
	}
}
