package models

// Service identifies an upstream media service supported by Audivo.
type Service string

const (
	ServiceYouTube    Service = "youtube"
	ServiceAppleMusic Service = "apple_music"
)

type MediaKind string

const (
	MediaKindSingle     MediaKind = "single"
	MediaKindCollection MediaKind = "collection"
)

// AuthState is the safe, user-facing lifecycle of a provider authentication
// session. Cookie values and paths never cross this contract.
type AuthState string

const (
	AuthStateIdle         AuthState = "idle"
	AuthStateOpening      AuthState = "opening"
	AuthStateWaitingLogin AuthState = "waiting_login"
	AuthStateConnected    AuthState = "connected"
	AuthStateCancelled    AuthState = "cancelled"
	AuthStateUnavailable  AuthState = "unavailable"
	AuthStateError        AuthState = "error"
)

const (
	AuthErrorBusy               = "auth_busy"
	AuthErrorBrowserUnavailable = "auth_browser_unavailable"
	AuthErrorBrowserFailed      = "auth_browser_failed"
	AuthErrorCaptureInvalid     = "auth_capture_invalid"
	AuthErrorDisconnectFailed   = "auth_disconnect_failed"
	AuthErrorProviderInvalid    = "auth_provider_invalid"
	AuthErrorTimeout            = "auth_timeout"
)

// MediaType identifies the kind of output selected by the user.
type MediaType string

const (
	MediaTypeAudio MediaType = "audio"
	MediaTypeVideo MediaType = "video"
)

// Format is intentionally small and user-facing. Engine-specific format
// expressions never cross this boundary.
type Format string

const (
	FormatMP3 Format = "mp3"
	FormatM4A Format = "m4a"
	FormatMP4 Format = "mp4"
	FormatM4V Format = "m4v"
)

// Quality is a normalized user-facing quality choice.
type Quality string

const (
	QualityBest Quality = "best"
	Quality360  Quality = "360"
	Quality480  Quality = "480"
	Quality540  Quality = "540"
	Quality720  Quality = "720"
	Quality1080 Quality = "1080"
	Quality1440 Quality = "1440"
	Quality2160 Quality = "2160"
)

// JobState is shared by backend events and the frontend reducer.
type JobState string

const (
	StateIdle                JobState = "idle"
	StateAnalyzing           JobState = "analyzing"
	StateReady               JobState = "ready"
	StateQueued              JobState = "queued"
	StatePreparing           JobState = "preparing"
	StateDownloading         JobState = "downloading"
	StateRetrying            JobState = "retrying"
	StateProcessing          JobState = "processing"
	StateCollection          JobState = "collection-downloading"
	StateCompleted           JobState = "completed"
	StatePartial             JobState = "collection-partial"
	StateCollectionCompleted JobState = "collection-completed"
	StateCancelled           JobState = "cancelled"
	StateError               JobState = "error"
)

// Stage describes the current human-readable step of an operation.
type Stage string

const (
	StagePreparing        Stage = "preparing"
	StageAnalyzing        Stage = "analyzing"
	StageDownloading      Stage = "downloading"
	StageDownloadingAudio Stage = "downloading_audio"
	StageMerging          Stage = "merging"
	StageConverting       Stage = "converting"
	StageMetadata         Stage = "metadata"
	StageFinalizing       Stage = "finalizing"
	StageCompleted        Stage = "completed"
)

type MediaOption struct {
	MediaType MediaType `json:"mediaType"`
	Format    Format    `json:"format"`
	Quality   Quality   `json:"quality,omitempty"`
	Label     string    `json:"label"`
}

type CollectionItem struct {
	ID              string `json:"id,omitempty"`
	URL             string `json:"url"`
	Title           string `json:"title"`
	Artist          string `json:"artist,omitempty"`
	Album           string `json:"album,omitempty"`
	Index           int    `json:"index"`
	DurationSeconds int    `json:"durationSeconds,omitempty"`
}

type MediaPreview struct {
	URL             string           `json:"url"`
	Service         Service          `json:"service"`
	Kind            MediaKind        `json:"kind"`
	Title           string           `json:"title"`
	ItemCount       int              `json:"itemCount,omitempty"`
	Items           []CollectionItem `json:"items,omitempty"`
	ThumbnailURL    string           `json:"thumbnailUrl,omitempty"`
	Author          string           `json:"author,omitempty"`
	DurationSeconds int              `json:"durationSeconds,omitempty"`
	MediaType       MediaType        `json:"mediaType"`
	Options         []MediaOption    `json:"options"`
}

type MediaInfo = MediaPreview

type DownloadRequest struct {
	URL             string           `json:"url"`
	Title           string           `json:"title,omitempty"`
	MediaType       MediaType        `json:"mediaType"`
	Format          Format           `json:"format"`
	Quality         Quality          `json:"quality,omitempty"`
	OutputDir       string           `json:"outputDir"`
	CollectionTitle string           `json:"collectionTitle,omitempty"`
	CollectionItems []CollectionItem `json:"collectionItems,omitempty"`
}

type DownloadJob struct {
	ID       string   `json:"id"`
	State    JobState `json:"state"`
	Filename string   `json:"filename,omitempty"`
	Message  string   `json:"message,omitempty"`
}

type DownloadResult struct {
	Files          []string `json:"files"`
	Directory      string   `json:"directory"`
	TotalItems     int      `json:"totalItems,omitempty"`
	CompletedItems int      `json:"completedItems,omitempty"`
	FailedItems    int      `json:"failedItems,omitempty"`
	Partial        bool     `json:"partial,omitempty"`
}

type UserError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
	Details   string `json:"details,omitempty"`
}

type DownloadEvent struct {
	JobID          string          `json:"jobId"`
	State          JobState        `json:"state"`
	Stage          Stage           `json:"stage,omitempty"`
	Percent        float64         `json:"percent"`
	OverallPercent float64         `json:"overallPercent,omitempty"`
	ItemPercent    float64         `json:"itemPercent,omitempty"`
	CurrentIndex   int             `json:"currentIndex,omitempty"`
	TotalItems     int             `json:"totalItems,omitempty"`
	ItemTitle      string          `json:"itemTitle,omitempty"`
	Speed          string          `json:"speed,omitempty"`
	ETASeconds     int             `json:"etaSeconds,omitempty"`
	RetryAttempt   int             `json:"retryAttempt,omitempty"`
	RetryLimit     int             `json:"retryLimit,omitempty"`
	Filename       string          `json:"filename,omitempty"`
	Message        string          `json:"message,omitempty"`
	Error          *UserError      `json:"error,omitempty"`
	Result         *DownloadResult `json:"result,omitempty"`
}

type DownloadProgress = DownloadEvent

type TechnicalLogLine struct {
	JobID   string `json:"jobId"`
	Stream  string `json:"stream"`
	Message string `json:"message"`
}

type AnalyzeResponse struct {
	Preview   *MediaPreview `json:"preview,omitempty"`
	Duplicate *HistoryItem  `json:"duplicate,omitempty"`
	Error     *UserError    `json:"error,omitempty"`
}

type StartDownloadResponse struct {
	Job   *DownloadJob `json:"job,omitempty"`
	Error *UserError   `json:"error,omitempty"`
}

type ActionResponse struct {
	OK    bool       `json:"ok"`
	Value string     `json:"value,omitempty"`
	Error *UserError `json:"error,omitempty"`
}

type Settings struct {
	DownloadDirectory     string  `json:"downloadDirectory"`
	Theme                 string  `json:"theme"`
	AudioFormat           Format  `json:"audioFormat"`
	VideoQuality          Quality `json:"videoQuality,omitempty"`
	QuickMode             bool    `json:"quickMode,omitempty"`
	AppleMusicCookiesPath string  `json:"appleMusicCookiesPath,omitempty"`
	YouTubeCookiesPath    string  `json:"youtubeCookiesPath,omitempty"`
	LastEngineCheckUnix   int64   `json:"lastEngineCheckUnix,omitempty"`
}

// SettingsView is the public Wails settings contract. Secret-bearing cookie
// paths remain in Settings, which is used only by the local config/service
// layers, while the frontend receives connection state instead.
type SettingsView struct {
	DownloadDirectory   string  `json:"downloadDirectory"`
	Theme               string  `json:"theme"`
	AudioFormat         Format  `json:"audioFormat"`
	VideoQuality        Quality `json:"videoQuality,omitempty"`
	QuickMode           bool    `json:"quickMode,omitempty"`
	AppleMusicConnected bool    `json:"appleMusicConnected"`
	YouTubeConnected    bool    `json:"youtubeConnected"`
	LastEngineCheckUnix int64   `json:"lastEngineCheckUnix,omitempty"`
	Warning             string  `json:"warning,omitempty"`
}

// ProviderAuthStatus is deliberately free of cookie paths, headers, and
// values. It is safe to emit to the React frontend.
type ProviderAuthStatus struct {
	Provider         Service    `json:"provider"`
	State            AuthState  `json:"state"`
	Connected        bool       `json:"connected"`
	Message          string     `json:"message,omitempty"`
	SessionID        string     `json:"sessionId,omitempty"`
	LastCapturedUnix int64      `json:"lastCapturedUnix,omitempty"`
	Error            *UserError `json:"error,omitempty"`
}

type EngineStatus struct {
	Name            string `json:"name"`
	Version         string `json:"version,omitempty"`
	RequiredVersion string `json:"requiredVersion,omitempty"`
	Path            string `json:"path,omitempty"`
	Available       bool   `json:"available"`
	Valid           bool   `json:"valid"`
	Message         string `json:"message,omitempty"`
}

type EngineResponse struct {
	Engines []EngineStatus `json:"engines"`
	Error   *UserError     `json:"error,omitempty"`
}

type ProviderHealth struct {
	Provider      Service `json:"provider"`
	Available     bool    `json:"available"`
	Authenticated bool    `json:"authenticated"`
	Message       string  `json:"message,omitempty"`
}

type HistoryItem struct {
	ID             string    `json:"id"`
	SourceURL      string    `json:"sourceUrl"`
	Service        Service   `json:"service"`
	Title          string    `json:"title"`
	Type           MediaKind `json:"type"`
	FilePath       string    `json:"filePath,omitempty"`
	OutputDir      string    `json:"outputDir"`
	CompletedAt    string    `json:"completedAt"`
	Status         string    `json:"status"`
	ItemCount      int       `json:"itemCount,omitempty"`
	CompletedItems int       `json:"completedItems,omitempty"`
	Missing        bool      `json:"missing,omitempty"`
}

type HistoryResponse struct {
	Items []HistoryItem `json:"items"`
}

type DiagnosticResponse struct {
	Report string     `json:"report"`
	Error  *UserError `json:"error,omitempty"`
}

type UpdateResponse struct {
	CurrentVersion string     `json:"currentVersion"`
	LatestVersion  string     `json:"latestVersion,omitempty"`
	ReleaseURL     string     `json:"releaseUrl,omitempty"`
	Available      bool       `json:"available"`
	Error          *UserError `json:"error,omitempty"`
}
