import type { DownloadEvent, DownloadResult, HistoryItem, JobState, MediaPreview, ProviderAuthStatus, Service, TechnicalLogLine, UserError } from './types'

export type AppPhase = Extract<JobState, 'idle' | 'analyzing' | 'ready' | 'queued' | 'preparing' | 'downloading' | 'retrying' | 'processing' | 'collection-downloading' | 'completed' | 'collection-partial' | 'collection-completed' | 'cancelled' | 'error'>

export interface AppState {
  phase: AppPhase
  url: string
  preview?: MediaPreview
  duplicate?: HistoryItem
  preferenceNotice?: string
  mediaType: 'audio' | 'video'
  format: 'mp3' | 'm4a' | 'mp4' | 'm4v'
  quality: string
  jobId?: string
  progress: DownloadEvent
  result?: DownloadResult
  error?: UserError
  logs: string[]
}

export const initialState: AppState = {
  phase: 'idle',
  url: '',
  mediaType: 'audio',
  format: 'mp3',
  quality: 'best',
  progress: { jobId: '', state: 'idle', percent: 0 },
  logs: [],
}

export type Action =
  | { type: 'urlChanged'; url: string }
  | { type: 'analysisStarted' }
  | { type: 'analysisSucceeded'; preview: MediaPreview; duplicate?: HistoryItem; preferredAudioFormat?: 'mp3' | 'm4a'; preferredVideoQuality?: string }
  | { type: 'analysisFailed'; error: UserError }
  | { type: 'optionChanged'; mediaType?: AppState['mediaType']; format?: AppState['format']; quality?: string }
  | { type: 'downloadStarted'; jobId: string }
  | { type: 'downloadEvent'; event: DownloadEvent }
  | { type: 'technicalLog'; line: TechnicalLogLine }
  | { type: 'reset' }

export function reducer(state: AppState, action: Action): AppState {
  switch (action.type) {
    case 'urlChanged':
      return {
        ...state,
        url: action.url,
        phase: state.phase === 'idle' || state.phase === 'error' ? 'idle' : state.phase,
        jobId: state.phase === 'idle' || state.phase === 'error' ? undefined : state.jobId,
        error: undefined,
      }
    case 'analysisStarted':
      return { ...state, phase: 'analyzing', jobId: undefined, result: undefined, duplicate: undefined, preferenceNotice: undefined, error: undefined, logs: [] }
    case 'analysisSucceeded': {
      const firstOption = action.preview.options.find((option) => option.mediaType === action.preview.mediaType)
      const preferredOption = action.preview.mediaType === 'audio' && action.preferredAudioFormat
        ? action.preview.options.find((option) => option.mediaType === 'audio' && option.format === action.preferredAudioFormat)
        : action.preview.mediaType === 'video' && action.preferredVideoQuality
          ? action.preview.options.find((option) => option.mediaType === 'video' && option.quality === action.preferredVideoQuality)
            ?? action.preview.options.find((option) => option.mediaType === 'video' && option.quality === 'best')
          : undefined
      const selectedOption = preferredOption ?? firstOption
      const preferenceNotice = action.preview.mediaType === 'audio' && action.preferredAudioFormat && !action.preview.options.some((option) => option.mediaType === 'audio' && option.format === action.preferredAudioFormat)
        ? `Formato ${action.preferredAudioFormat.toUpperCase()} indisponível; usando ${selectedOption?.format.toUpperCase() ?? 'a melhor opção'}.`
        : action.preview.mediaType === 'video' && action.preferredVideoQuality && selectedOption?.quality !== action.preferredVideoQuality
          ? `Qualidade ${action.preferredVideoQuality}p indisponível; usando ${selectedOption?.quality === 'best' ? 'a melhor disponível' : selectedOption?.quality ? `${selectedOption.quality}p` : 'a melhor opção'}.`
          : undefined
      return {
        ...state,
        phase: 'ready',
        preview: action.preview,
        duplicate: action.duplicate,
        preferenceNotice,
        mediaType: action.preview.mediaType,
        format: preferredOption?.format ?? firstOption?.format ?? state.format,
        quality: preferredOption?.quality ?? firstOption?.quality ?? 'best',
        jobId: undefined,
        result: undefined,
        error: undefined,
      }
    }
    case 'analysisFailed':
      return { ...state, phase: 'error', error: action.error, preview: undefined, duplicate: undefined, preferenceNotice: undefined, jobId: undefined, result: undefined }
    case 'optionChanged':
      return { ...state, mediaType: action.mediaType ?? state.mediaType, format: action.format ?? state.format, quality: action.quality ?? state.quality }
    case 'downloadStarted':
      if (state.jobId === action.jobId && ['completed', 'cancelled', 'error'].includes(state.phase)) return state
      return { ...state, phase: 'preparing', jobId: action.jobId, error: undefined, result: undefined, logs: [] }
    case 'downloadEvent': {
      if (state.jobId && state.jobId !== action.event.jobId) return state
      return {
        ...state,
        phase: action.event.state,
        jobId: action.event.jobId || state.jobId,
        progress: action.event,
        result: action.event.result ?? state.result,
        error: action.event.error,
      }
    }
    case 'technicalLog':
      if (state.jobId && state.jobId !== action.line.jobId) return state
      return { ...state, logs: [...state.logs, `[${action.line.stream}] ${action.line.message}`].slice(-200) }
    case 'reset':
      return { ...initialState, url: state.url }
    default:
      return state
  }
}

export type ProviderAuthMap = Record<Service, ProviderAuthStatus>

export const initialProviderAuthState: ProviderAuthMap = {
  apple_music: { provider: 'apple_music', state: 'idle', connected: false },
  youtube: { provider: 'youtube', state: 'idle', connected: false },
}

export type ProviderAuthAction =
  | { type: 'providerStatusUpdated'; status: ProviderAuthStatus }
  | { type: 'providerStatusReset'; provider: Service; connected: boolean }

export function providerAuthReducer(state: ProviderAuthMap, action: ProviderAuthAction): ProviderAuthMap {
  switch (action.type) {
    case 'providerStatusUpdated':
      return { ...state, [action.status.provider]: action.status }
    case 'providerStatusReset':
      return {
        ...state,
        [action.provider]: {
          ...state[action.provider],
          provider: action.provider,
          state: action.connected ? 'connected' : 'idle',
          connected: action.connected,
          error: undefined,
          message: action.connected ? 'Sessão conectada.' : 'Não conectado.',
        },
      }
    default:
      return state
  }
}
