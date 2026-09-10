import { describe, expect, it } from 'vitest'
import { initialProviderAuthState, initialState, providerAuthReducer, reducer } from './state'
import type { DownloadEvent, MediaPreview, ProviderAuthStatus } from './types'

const preview: MediaPreview = {
  url: 'https://youtu.be/example',
  service: 'youtube',
  title: 'Example video',
  mediaType: 'video',
  options: [
    { mediaType: 'video', format: 'mp4', quality: '720', label: '720p · MP4' },
    { mediaType: 'audio', format: 'mp3', quality: 'best', label: 'MP3' },
  ],
}

const event = (overrides: Partial<DownloadEvent> = {}): DownloadEvent => ({
  jobId: 'job-1',
  state: 'downloading',
  percent: 42,
  ...overrides,
})

describe('download workflow reducer', () => {
  it('selects the first compatible option after analysis', () => {
    const next = reducer(initialState, { type: 'analysisSucceeded', preview })

    expect(next.phase).toBe('ready')
    expect(next.mediaType).toBe('video')
    expect(next.format).toBe('mp4')
    expect(next.quality).toBe('720')
  })

  it('honors the saved audio format preference when available', () => {
    const audioPreview: MediaPreview = {
      ...preview,
      mediaType: 'audio',
      options: [
        { mediaType: 'audio', format: 'mp3', label: 'MP3' },
        { mediaType: 'audio', format: 'm4a', label: 'M4A' },
      ],
    }
    const next = reducer(initialState, { type: 'analysisSucceeded', preview: audioPreview, preferredAudioFormat: 'm4a' })

    expect(next.format).toBe('m4a')
  })

  it('explains when the saved video quality is unavailable', () => {
    const next = reducer(initialState, { type: 'analysisSucceeded', preview, preferredVideoQuality: '1080' })

    expect(next.quality).toBe('720')
    expect(next.preferenceNotice).toContain('1080p indisponível')
  })

  it('does not let an old job overwrite the active job', () => {
    const downloading = reducer(
      reducer(initialState, { type: 'downloadStarted', jobId: 'job-2' }),
      { type: 'downloadEvent', event: event({ jobId: 'job-2' }) },
    )

    const stale = reducer(downloading, { type: 'downloadEvent', event: event({ jobId: 'job-1', percent: 99 }) })

    expect(stale.progress.percent).toBe(42)
    expect(stale.jobId).toBe('job-2')
  })

  it('keeps terminal result and exposes backend errors', () => {
    const started = reducer(initialState, { type: 'downloadStarted', jobId: 'job-1' })
    const completed = reducer(started, {
      type: 'downloadEvent',
      event: event({
        state: 'completed',
        percent: 100,
        result: { files: ['/tmp/example.mp3'], directory: '/tmp' },
      }),
    })

    expect(completed.phase).toBe('completed')
    expect(completed.result?.files).toEqual(['/tmp/example.mp3'])
  })

  it('does not overwrite a terminal event that arrived before the start response', () => {
    const terminal = reducer(
      { ...initialState, jobId: 'job-fast', phase: 'error', error: { code: 'engine_missing', message: 'Missing', retryable: true } },
      { type: 'downloadStarted', jobId: 'job-fast' },
    )

    expect(terminal.phase).toBe('error')
    expect(terminal.error?.code).toBe('engine_missing')
  })

  it('limits technical log growth and clears it on a new analysis', () => {
    let state = reducer(initialState, { type: 'downloadStarted', jobId: 'job-1' })
    for (let index = 0; index < 240; index += 1) {
      state = reducer(state, { type: 'technicalLog', line: { jobId: 'job-1', stream: 'stderr', message: `line-${index}` } })
    }

    expect(state.logs).toHaveLength(200)
    expect(state.logs[0]).toContain('line-40')
    expect(reducer(state, { type: 'analysisStarted' }).logs).toEqual([])
  })

  it('clears the previous job before a new analysis', () => {
    const state = reducer({ ...initialState, phase: 'completed', jobId: 'job-old' }, { type: 'analysisStarted' })

    expect(state.phase).toBe('analyzing')
    expect(state.jobId).toBeUndefined()
    expect(state.result).toBeUndefined()
  })
})

describe('provider authentication reducer', () => {
  const connected: ProviderAuthStatus = {
    provider: 'apple_music',
    state: 'connected',
    connected: true,
    message: 'Sessão conectada.',
  }

  it('updates one provider without changing the other provider', () => {
    const next = providerAuthReducer(initialProviderAuthState, { type: 'providerStatusUpdated', status: connected })

    expect(next.apple_music.state).toBe('connected')
    expect(next.youtube.state).toBe('idle')
    expect(next.youtube.connected).toBe(false)
  })

  it('preserves connection state when a reconnect is cancelled', () => {
    const waiting = providerAuthReducer(
      providerAuthReducer(initialProviderAuthState, { type: 'providerStatusUpdated', status: connected }),
      { type: 'providerStatusUpdated', status: { ...connected, state: 'waiting_login', message: 'Faça login.' } },
    )
    const cancelled = providerAuthReducer(waiting, { type: 'providerStatusUpdated', status: { ...connected, state: 'cancelled', message: 'Autenticação cancelada.' } })

    expect(cancelled.apple_music.state).toBe('cancelled')
    expect(cancelled.apple_music.connected).toBe(true)
  })

  it('renders an error status without exposing a cookie path', () => {
    const next = providerAuthReducer(initialProviderAuthState, {
      type: 'providerStatusUpdated',
      status: { provider: 'youtube', state: 'error', connected: false, error: { code: 'youtube_cookies_invalid', message: 'Importe cookies válidos.', retryable: false } },
    })

    expect(next.youtube.error?.message).toBe('Importe cookies válidos.')
    expect(JSON.stringify(next)).not.toContain('cookies.txt')
  })
})
