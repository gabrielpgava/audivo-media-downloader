export type Service = 'youtube' | 'apple_music'
export type MediaKind = 'single' | 'collection'
export type ProviderAuthState = 'idle' | 'opening' | 'waiting_login' | 'connected' | 'cancelled' | 'unavailable' | 'error'
export type MediaType = 'audio' | 'video'
export type Format = 'mp3' | 'm4a' | 'mp4' | 'm4v'
export type Quality = 'best' | '360' | '480' | '540' | '720' | '1080' | '1440' | '2160'
export type JobState = 'idle' | 'analyzing' | 'ready' | 'queued' | 'preparing' | 'downloading' | 'retrying' | 'processing' | 'collection-downloading' | 'completed' | 'collection-partial' | 'collection-completed' | 'cancelled' | 'error'

export interface MediaOption {
  mediaType: MediaType
  format: Format
  quality?: Quality
  label: string
}

export interface CollectionItem {
  id?: string
  url: string
  title: string
  artist?: string
  album?: string
  index: number
  durationSeconds?: number
}

export interface MediaPreview {
  url: string
  service: Service
  kind?: MediaKind
  title: string
  itemCount?: number
  items?: CollectionItem[]
  thumbnailUrl?: string
  author?: string
  durationSeconds?: number
  mediaType: MediaType
  options: MediaOption[]
}

export interface DownloadRequest {
  url: string
  title?: string
  mediaType: MediaType
  format: Format
  quality?: Quality
  outputDir: string
  collectionTitle?: string
  collectionItems?: CollectionItem[]
}

export interface DownloadJob {
  id: string
  state: JobState
  filename?: string
  message?: string
}

export interface UserError {
  code: string
  message: string
  retryable: boolean
  details?: string
}

export interface DownloadResult {
  files: string[]
  directory: string
  totalItems?: number
  completedItems?: number
  failedItems?: number
  partial?: boolean
}

export interface DownloadEvent {
  jobId: string
  state: JobState
  stage?: string
  percent: number
  overallPercent?: number
  itemPercent?: number
  currentIndex?: number
  totalItems?: number
  itemTitle?: string
  retryAttempt?: number
  retryLimit?: number
  speed?: string
  etaSeconds?: number
  filename?: string
  message?: string
  error?: UserError
  result?: DownloadResult
}

export interface TechnicalLogLine {
  jobId: string
  stream: string
  message: string
}

export interface AnalyzeResponse {
  preview?: MediaPreview
  duplicate?: HistoryItem
  error?: UserError
}

export interface StartDownloadResponse {
  job?: DownloadJob
  error?: UserError
}

export interface ActionResponse {
  ok: boolean
  value?: string
  error?: UserError
}

export interface Settings {
  downloadDirectory: string
  theme: 'system' | 'light' | 'dark'
  audioFormat: 'mp3' | 'm4a'
  videoQuality: Quality
  quickMode: boolean
  appleMusicConnected: boolean
  youtubeConnected: boolean
  lastEngineCheckUnix?: number
  warning?: string
}

export interface ProviderAuthStatus {
  provider: Service
  state: ProviderAuthState
  connected: boolean
  message?: string
  sessionId?: string
  lastCapturedUnix?: number
  error?: UserError
}

export interface EngineStatus {
  name: string
  version?: string
  requiredVersion?: string
  path?: string
  available: boolean
  valid: boolean
  message?: string
}

export interface EngineResponse {
  engines: EngineStatus[]
  error?: UserError
}

export interface DiagnosticResponse {
  report: string
  error?: UserError
}

export interface UpdateResponse {
  currentVersion: string
  latestVersion?: string
  releaseUrl?: string
  available: boolean
  error?: UserError
}

export interface HistoryItem {
  id: string
  sourceUrl: string
  service: Service
  title: string
  type: MediaKind
  filePath?: string
  outputDir: string
  completedAt: string
  status: string
  itemCount?: number
  completedItems?: number
  missing?: boolean
}

export interface HistoryResponse {
  items: HistoryItem[]
}
