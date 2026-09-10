import { useEffect, useMemo, useReducer, useState } from 'react'
import type { FormEvent } from 'react'
import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  CardMedia,
  CircularProgress,
  Container,
  CssBaseline,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Divider,
  FormControl,
  FormControlLabel,
  IconButton,
  InputLabel,
  LinearProgress,
  MenuItem,
  Select,
  Stack,
  Switch,
  TextField,
  ThemeProvider,
  Tooltip,
  Typography,
  createTheme,
  useMediaQuery,
} from '@mui/material'
import ExpandMoreIcon from '@mui/icons-material/ExpandMore'
import FolderOpenOutlinedIcon from '@mui/icons-material/FolderOpenOutlined'
import ContentPasteOutlinedIcon from '@mui/icons-material/ContentPasteOutlined'
import DownloadOutlinedIcon from '@mui/icons-material/DownloadOutlined'
import ErrorOutlineOutlinedIcon from '@mui/icons-material/ErrorOutlineOutlined'
import OpenInNewOutlinedIcon from '@mui/icons-material/OpenInNewOutlined'
import RefreshOutlinedIcon from '@mui/icons-material/RefreshOutlined'
import SettingsOutlinedIcon from '@mui/icons-material/SettingsOutlined'
import StopCircleOutlinedIcon from '@mui/icons-material/StopCircleOutlined'
import CheckCircleOutlineOutlinedIcon from '@mui/icons-material/CheckCircleOutlineOutlined'
import { ClipboardGetText, EventsOn } from '../wailsjs/runtime'
import {
  analyzeURL,
  cancelProviderAuth,
  cancelDownload,
  clearHistory,
  checkForUpdate,
  copyDiagnostic,
  diagnostic,
  disconnectProvider,
  getEngineStatus,
  getSettings,
  openDownloadedFile,
  openDownloadFolder,
  openGitHubIssues,
  openRelease,
  recentDownloads,
  saveSettings,
  selectProviderCookies,
  selectDownloadDirectory,
  setWindowFocused,
  startProviderAuth,
  startDownload,
  updateEngines,
  verifyProviderAuth,
} from './api'
import { initialProviderAuthState, initialState, providerAuthReducer, reducer } from './state'
import type { AppState } from './state'
import type {
  ActionResponse,
  DiagnosticResponse,
  DownloadEvent,
  EngineResponse,
  Format,
  HistoryItem,
  MediaOption,
  ProviderAuthStatus,
  ProviderAuthState,
  Quality,
  Service,
  Settings,
  UpdateResponse,
  UserError,
} from './types'

const emptySettings: Settings = {
  downloadDirectory: '',
  theme: 'system',
  audioFormat: 'mp3',
  videoQuality: 'best',
  quickMode: false,
  appleMusicConnected: false,
  youtubeConnected: false,
}

const fallbackError = (message: string, code = 'operation_failed'): UserError => ({
  code,
  message,
  retryable: true,
})

const displayService = (service?: string) => service === 'apple_music' ? 'Apple Music' : 'YouTube'

const providerLabel: Record<Service, string> = {
  apple_music: 'Apple Music',
  youtube: 'YouTube',
}

const providerAuthLabel = (state: ProviderAuthState) => {
  switch (state) {
    case 'opening': return 'Abrindo sessão…'
    case 'waiting_login': return 'Aguardando login na janela do provedor'
    case 'connected': return 'Conectado'
    case 'cancelled': return 'Autenticação cancelada'
    case 'unavailable': return 'Navegador compatível indisponível'
    case 'error': return 'Não foi possível conectar'
    default: return 'Não conectado'
  }
}

const formatDuration = (seconds?: number) => {
  if (!seconds || seconds < 1) return ''
  const minutes = Math.floor(seconds / 60)
  const remainder = seconds % 60
  return `${minutes}:${String(remainder).padStart(2, '0')}`
}

const formatEta = (seconds?: number) => {
  if (!seconds || seconds < 1) return ''
  const minutes = Math.floor(seconds / 60)
  const remainder = seconds % 60
  return minutes > 0 ? `${minutes}m ${remainder}s restantes` : `${remainder}s restantes`
}

const formatHistoryDate = (value: string) => {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? '' : date.toLocaleString('pt-BR')
}

const optionLabel = (option: MediaOption) => {
  if (option.mediaType === 'audio') return option.label || option.format.toUpperCase()
  return option.label || `${option.quality === 'best' ? 'Melhor' : `${option.quality}p`} · ${option.format.toUpperCase()}`
}

const settingsError = (response: ActionResponse | undefined) => response?.error?.message ?? 'Não foi possível salvar esta alteração.'

const hasNativeBridge = () => {
  if (typeof window === 'undefined') return false
  const nativeWindow = window as Window & { go?: unknown; runtime?: unknown }
  return Boolean(nativeWindow.go && nativeWindow.runtime)
}

const looksLikeSupportedURL = (value: string) => {
  try {
    const host = new URL(value.trim()).hostname.toLowerCase().replace(/\.$/, '')
    return host === 'youtu.be' || host === 'youtube.com' || host.endsWith('.youtube.com') || host === 'music.apple.com' || host.endsWith('.music.apple.com')
  } catch {
    return false
  }
}

function App() {
  const [state, dispatch] = useReducer(reducer, initialState)
  const [providerAuth, dispatchProviderAuth] = useReducer(providerAuthReducer, initialProviderAuthState)
  const [settings, setSettings] = useState<Settings>(emptySettings)
  const [settingsDraft, setSettingsDraft] = useState<Settings>(emptySettings)
  const [settingsOpen, setSettingsOpen] = useState(false)
  const [settingsBusy, setSettingsBusy] = useState(false)
  const [settingsMessage, setSettingsMessage] = useState<string>()
  const [engineResponse, setEngineResponse] = useState<EngineResponse>()
  const [engineBusy, setEngineBusy] = useState(false)
  const [clipboardBusy, setClipboardBusy] = useState(false)
  const [clipboardSuggestion, setClipboardSuggestion] = useState<string>()
  const [authBusyProvider, setAuthBusyProvider] = useState<Service>()
  const [recent, setRecent] = useState<HistoryItem[]>([])
  const [diagnosticResponse, setDiagnosticResponse] = useState<DiagnosticResponse>()
  const [diagnosticBusy, setDiagnosticBusy] = useState(false)
  const [updateResponse, setUpdateResponse] = useState<UpdateResponse>()
  const [updateBusy, setUpdateBusy] = useState(false)
  const systemPrefersDark = useMediaQuery('(prefers-color-scheme: dark)')

  const darkMode = settings.theme === 'dark' || (settings.theme === 'system' && systemPrefersDark)
  const theme = useMemo(() => createTheme({
    palette: {
      mode: darkMode ? 'dark' : 'light',
      primary: { main: darkMode ? '#90caf9' : '#315c9b' },
      secondary: { main: darkMode ? '#ffb74d' : '#b45309' },
      background: {
        default: darkMode ? '#111827' : '#f4f6f8',
        paper: darkMode ? '#182232' : '#ffffff',
      },
    },
    shape: { borderRadius: 14 },
    typography: {
      fontFamily: 'Nunito, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif',
      h4: { fontWeight: 800, letterSpacing: '-0.03em' },
      h5: { fontWeight: 800, letterSpacing: '-0.02em' },
      button: { fontWeight: 700, textTransform: 'none' },
    },
    components: {
      MuiButton: { defaultProps: { disableElevation: true } },
      MuiCard: { styleOverrides: { root: { border: '1px solid', borderColor: darkMode ? '#2b3b51' : '#e4e8ee' } } },
    },
  }), [darkMode])

  useEffect(() => {
    if (!hasNativeBridge()) return
    let active = true
    getSettings()
      .then((loaded) => {
        if (!active) return
        setSettings(loaded)
        setSettingsDraft(loaded)
        dispatchProviderAuth({ type: 'providerStatusReset', provider: 'apple_music', connected: loaded.appleMusicConnected })
        dispatchProviderAuth({ type: 'providerStatusReset', provider: 'youtube', connected: loaded.youtubeConnected })
      })
      .catch(() => {
        // The browser preview can run without a Wails bridge. Native errors are surfaced on action.
      })
    return () => { active = false }
  }, [])

  useEffect(() => {
    if (!hasNativeBridge()) return
    const syncWindowState = () => {
      void setWindowFocused(document.visibilityState === 'visible' && document.hasFocus())
    }
    syncWindowState()
    window.addEventListener('focus', syncWindowState)
    window.addEventListener('blur', syncWindowState)
    document.addEventListener('visibilitychange', syncWindowState)
    return () => {
      window.removeEventListener('focus', syncWindowState)
      window.removeEventListener('blur', syncWindowState)
      document.removeEventListener('visibilitychange', syncWindowState)
    }
  }, [])

  useEffect(() => {
    if (!hasNativeBridge()) return
    const unsubscribeUpdated = EventsOn('download:updated', (payload: DownloadEvent) => {
      dispatch({ type: 'downloadEvent', event: payload })
    })
    const unsubscribeLog = EventsOn('download:log', (payload) => {
      dispatch({ type: 'technicalLog', line: payload })
    })
    const unsubscribeAuth = EventsOn('provider-auth:updated', (payload: ProviderAuthStatus) => {
      dispatchProviderAuth({ type: 'providerStatusUpdated', status: payload })
    })
    return () => {
      unsubscribeUpdated()
      unsubscribeLog()
      unsubscribeAuth()
    }
  }, [])

  useEffect(() => {
    if (!hasNativeBridge()) return
    void recentDownloads(5).then((response) => setRecent(response.items ?? [])).catch(() => {
      // History is a convenience and must not block the main flow.
    })
  }, [])

  useEffect(() => {
    if (!hasNativeBridge() || state.url.trim()) return
    const suggestClipboard = () => {
      void ClipboardGetText().then((value) => {
        const candidate = value.trim()
        if (looksLikeSupportedURL(candidate) && !state.url.trim()) setClipboardSuggestion(candidate)
      }).catch(() => undefined)
    }
    window.addEventListener('focus', suggestClipboard)
    return () => window.removeEventListener('focus', suggestClipboard)
  }, [state.url])

  const serviceOptions = useMemo(() => {
    const options = state.preview?.options ?? []
    return Array.from(new Set(options.map((option) => option.mediaType)))
  }, [state.preview])

  const filteredOptions = useMemo(
    () => state.preview?.options.filter((option) => option.mediaType === state.mediaType) ?? [],
    [state.preview, state.mediaType],
  )

  const reportResponseError = (response: { error?: UserError } | undefined, defaultMessage: string) => {
    if (response?.error) dispatch({ type: 'analysisFailed', error: response.error })
    else dispatch({ type: 'analysisFailed', error: fallbackError(defaultMessage) })
  }

  const handleAnalyze = async (event?: FormEvent) => {
    event?.preventDefault()
    const url = state.url.trim()
    if (!url) {
      dispatch({ type: 'analysisFailed', error: fallbackError('Cole um link do YouTube ou Apple Music.', 'url_required') })
      return
    }
    dispatch({ type: 'analysisStarted' })
    try {
      const response = await analyzeURL(url)
      if (response.error || !response.preview) {
        reportResponseError(response, 'Não foi possível analisar este link.')
        return
      }
      dispatch({
        type: 'analysisSucceeded',
        preview: response.preview,
        duplicate: response.duplicate,
        preferredAudioFormat: settings.audioFormat,
        preferredVideoQuality: settings.videoQuality,
      })
    } catch {
      dispatch({ type: 'analysisFailed', error: fallbackError('Não foi possível analisar este link.') })
    }
  }

  const handlePaste = async () => {
    setClipboardBusy(true)
    try {
      const text = await ClipboardGetText()
      const candidate = text.trim()
      if (!looksLikeSupportedURL(candidate)) {
        dispatch({ type: 'analysisFailed', error: fallbackError('A área de transferência não contém um link compatível.', 'unsupported_url') })
        return
      }
      setClipboardSuggestion(undefined)
      dispatch({ type: 'urlChanged', url: candidate })
    } catch {
      dispatch({ type: 'analysisFailed', error: fallbackError('Não foi possível ler a área de transferência.', 'clipboard_unavailable') })
    } finally {
      setClipboardBusy(false)
    }
  }

  const handleMediaType = (mediaType: AppState['mediaType']) => {
    const option = state.preview?.options.find((candidate) => candidate.mediaType === mediaType && (mediaType !== 'audio' || candidate.format === settings.audioFormat))
      ?? state.preview?.options.find((candidate) => candidate.mediaType === mediaType)
    dispatch({
      type: 'optionChanged',
      mediaType,
      format: option?.format,
      quality: option?.quality ?? 'best',
    })
  }

  const handleDownload = async () => {
    if (!state.preview) return
    if (!settings.downloadDirectory) {
      setSettingsOpen(true)
      setSettingsMessage('Escolha uma pasta de download antes de começar.')
      return
    }
    try {
      const response = await startDownload({
        url: state.preview.url || state.url,
        title: state.preview.title,
        mediaType: state.mediaType,
        format: state.format,
        quality: state.quality as Quality,
        outputDir: settings.downloadDirectory,
        collectionTitle: state.preview.kind === 'collection' ? state.preview.title : undefined,
        collectionItems: state.preview.kind === 'collection' ? state.preview.items : undefined,
      })
      if (response.error || !response.job) {
        dispatch({ type: 'analysisFailed', error: response.error ?? fallbackError('Não foi possível iniciar o download.') })
        return
      }
      dispatch({ type: 'downloadStarted', jobId: response.job.id })
    } catch {
      dispatch({ type: 'analysisFailed', error: fallbackError('Não foi possível iniciar o download.') })
    }
  }

  const handleCancel = async () => {
    if (!state.jobId) return
    if (state.phase === 'collection-downloading' && !window.confirm('Cancelar a coleção? Os itens já concluídos serão preservados.')) return
    try {
      const response = await cancelDownload(state.jobId)
      if (response.error) {
        dispatch({ type: 'downloadEvent', event: { ...state.progress, state: 'error', error: response.error } })
      }
    } catch {
      dispatch({ type: 'downloadEvent', event: { ...state.progress, state: 'error', error: fallbackError('Não foi possível cancelar o download.') } })
    }
  }

  const refreshSettings = async () => {
    const loaded = await getSettings()
    setSettings(loaded)
    setSettingsDraft(loaded)
    dispatchProviderAuth({ type: 'providerStatusReset', provider: 'apple_music', connected: loaded.appleMusicConnected })
    dispatchProviderAuth({ type: 'providerStatusReset', provider: 'youtube', connected: loaded.youtubeConnected })
    return loaded
  }

  const refreshEngines = async () => {
    setEngineBusy(true)
    try {
      setEngineResponse(await getEngineStatus())
    } catch {
      setEngineResponse({ engines: [], error: fallbackError('Não foi possível consultar os engines.') })
    } finally {
      setEngineBusy(false)
    }
  }

  const openSettings = async () => {
    setSettingsMessage(undefined)
    setSettingsOpen(true)
    setSettingsBusy(true)
    try {
      await refreshSettings()
      await refreshEngines()
      await loadUpdate()
    } catch {
      setSettingsMessage('Não foi possível carregar as configurações nativas.')
    } finally {
      setSettingsBusy(false)
    }
  }

  const handleUpdateEngines = async () => {
    setEngineBusy(true)
    setSettingsMessage(undefined)
    try {
      const response = await updateEngines()
      setEngineResponse(response)
      if (response.error) setSettingsMessage(response.error.message)
      else setSettingsMessage('Verificação dos engines concluída.')
    } catch {
      setSettingsMessage('Não foi possível atualizar os engines.')
    } finally {
      setEngineBusy(false)
    }
  }

  const chooseDirectory = async () => {
    setSettingsBusy(true)
    try {
      const response = await selectDownloadDirectory()
      if (response.error) setSettingsMessage(settingsError(response))
      else await refreshSettings()
    } catch {
      setSettingsMessage('Não foi possível escolher a pasta de downloads.')
    } finally {
      setSettingsBusy(false)
    }
  }

  const chooseCookies = async (provider: Service) => {
    setSettingsBusy(true)
    try {
      const response = await selectProviderCookies(provider)
      if (response.error) setSettingsMessage(settingsError(response))
      else await refreshSettings()
    } catch {
      setSettingsMessage(`Não foi possível importar a sessão do ${providerLabel[provider]}.`)
    } finally {
      setSettingsBusy(false)
    }
  }

  const updateProviderAuth = (status: ProviderAuthStatus) => {
    dispatchProviderAuth({ type: 'providerStatusUpdated', status })
    if (status.error) setSettingsMessage(status.error.message)
  }

  const runProviderAuthAction = async (provider: Service, action: (provider: Service) => Promise<ProviderAuthStatus>, fallback: string) => {
    setAuthBusyProvider(provider)
    try {
      updateProviderAuth(await action(provider))
    } catch {
      setSettingsMessage(fallback)
    } finally {
      setAuthBusyProvider(undefined)
    }
  }

  const handleConnect = (provider: Service) => runProviderAuthAction(provider, startProviderAuth, `Não foi possível abrir a sessão do ${providerLabel[provider]}.`)
  const handleVerify = (provider: Service) => runProviderAuthAction(provider, verifyProviderAuth, `Não foi possível verificar a sessão do ${providerLabel[provider]}.`)
  const handleCancelAuth = (provider: Service) => runProviderAuthAction(provider, cancelProviderAuth, `Não foi possível cancelar a sessão do ${providerLabel[provider]}.`)
  const handleDisconnect = (provider: Service) => runProviderAuthAction(provider, disconnectProvider, `Não foi possível desconectar o ${providerLabel[provider]}.`)

  const handleClearHistory = async () => {
    if (!window.confirm('Limpar o histórico local? Os arquivos baixados serão preservados.')) return
    setSettingsBusy(true)
    try {
      const response = await clearHistory()
      if (response.error) setSettingsMessage(response.error.message)
      else setRecent([])
    } catch {
      setSettingsMessage('Não foi possível limpar o histórico.')
    } finally {
      setSettingsBusy(false)
    }
  }

  const loadDiagnostic = async () => {
    setDiagnosticBusy(true)
    try {
      setDiagnosticResponse(await diagnostic())
    } catch {
      setDiagnosticResponse({ report: '', error: fallbackError('Não foi possível gerar o diagnóstico.') })
    } finally {
      setDiagnosticBusy(false)
    }
  }

  const handleCopyDiagnostic = async () => {
    setDiagnosticBusy(true)
    try {
      const response = await copyDiagnostic()
      setSettingsMessage(response.error?.message ?? 'Diagnóstico copiado para a área de transferência.')
    } catch {
      setSettingsMessage('Não foi possível copiar o diagnóstico.')
    } finally {
      setDiagnosticBusy(false)
    }
  }

  const loadUpdate = async () => {
    setUpdateBusy(true)
    try {
      setUpdateResponse(await checkForUpdate())
    } catch {
      setUpdateResponse({ currentVersion: '', available: false, error: fallbackError('Não foi possível verificar atualizações.') })
    } finally {
      setUpdateBusy(false)
    }
  }

  const authSessionActive = Object.values(providerAuth).some((status) => status.state === 'opening' || status.state === 'waiting_login')

  const savePreferences = async () => {
    setSettingsBusy(true)
    setSettingsMessage(undefined)
    try {
      const response = await saveSettings(settingsDraft)
      if (response.error) {
        setSettingsMessage(settingsError(response))
        return
      }
      setSettings(response.ok ? settingsDraft : settings)
      setSettingsOpen(false)
    } catch {
      setSettingsMessage('Não foi possível salvar as configurações.')
    } finally {
      setSettingsBusy(false)
    }
  }

  const primaryFile = state.result?.files[0]
  const busy = ['analyzing', 'queued', 'preparing', 'downloading', 'retrying', 'processing', 'collection-downloading'].includes(state.phase)
  const showPreview = Boolean(state.preview) && ['ready', 'queued', 'preparing', 'downloading', 'retrying', 'processing', 'collection-downloading', 'completed', 'collection-partial', 'collection-completed'].includes(state.phase)
  const progressPercent = state.progress.overallPercent ?? state.progress.percent

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Box className="app-shell">
        <Container maxWidth="md" sx={{ py: { xs: 2, sm: 4 } }}>
          <Stack spacing={3}>
            <Box component="header" className="app-header">
              <Box>
                <Typography variant="overline" color="primary" sx={{ fontWeight: 800, letterSpacing: '0.14em' }}>
                  AUDIVO
                </Typography>
                <Typography variant="h4">Baixe sua mídia</Typography>
                <Typography color="text.secondary" sx={{ mt: 0.5 }}>
                  Um fluxo simples para YouTube e Apple Music.
                </Typography>
              </Box>
              <Tooltip title="Configurações">
                <IconButton aria-label="Abrir configurações" onClick={openSettings} color="primary">
                  <SettingsOutlinedIcon />
                </IconButton>
              </Tooltip>
            </Box>

            <Card component="section" aria-labelledby="download-form-title">
              <CardContent sx={{ p: { xs: 2, sm: 3 } }}>
                <Stack spacing={2.5}>
                  <Box>
                    <Typography id="download-form-title" variant="h5">Comece com um link</Typography>
                    <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5 }}>
                      O serviço é identificado automaticamente e o link é validado antes de qualquer engine ser executado.
                    </Typography>
                  </Box>
                  <Box component="form" onSubmit={handleAnalyze}>
                    <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1.5}>
                      <TextField
                        fullWidth
                        label="URL do YouTube ou Apple Music"
                        value={state.url}
                        onChange={(event) => dispatch({ type: 'urlChanged', url: event.target.value })}
                        onDrop={(event) => {
                          event.preventDefault()
                          const dropped = event.dataTransfer.getData('text/plain').trim()
                          if (dropped) dispatch({ type: 'urlChanged', url: dropped })
                        }}
                        onDragOver={(event) => event.preventDefault()}
                        placeholder="https://..."
                        autoComplete="url"
                        disabled={busy}
                      />
                      <Button
                        variant="outlined"
                        startIcon={clipboardBusy ? <CircularProgress size={16} /> : <ContentPasteOutlinedIcon />}
                        onClick={handlePaste}
                        disabled={busy || clipboardBusy}
                        sx={{ minWidth: { sm: 118 } }}
                      >
                        Colar
                      </Button>
                      <Button
                        type="submit"
                        variant="contained"
                        startIcon={state.phase === 'analyzing' ? <CircularProgress color="inherit" size={16} /> : <RefreshOutlinedIcon />}
                        disabled={busy}
                        sx={{ minWidth: { sm: 136 } }}
                      >
                        Analisar
                      </Button>
                    </Stack>
                  </Box>
                  {clipboardSuggestion && !state.url.trim() && (
                    <Alert severity="info" action={<Button size="small" onClick={() => { dispatch({ type: 'urlChanged', url: clipboardSuggestion }); setClipboardSuggestion(undefined) }}>Usar link</Button>}>
                      Link compatível encontrado na área de transferência.
                    </Alert>
                  )}

                  {state.error && (
                    <Alert
                      severity="error"
                      icon={<ErrorOutlineOutlinedIcon />}
                      action={
                        <Stack direction={{ xs: 'column', sm: 'row' }} spacing={0.5}>
                          {state.error.retryable && <Button size="small" onClick={() => void handleAnalyze()}>Tentar novamente</Button>}
                          {['engine_missing', 'engine_invalid', 'cookies_missing', 'cookies_invalid', 'apple_music_cookies_invalid', 'youtube_cookies_invalid', 'output_directory_invalid'].includes(state.error.code) && <Button size="small" onClick={openSettings}>Configurações</Button>}
                        </Stack>
                      }
                    >
                      <Typography variant="body2" sx={{ fontWeight: 700 }}>{state.error.message}</Typography>
                      {state.error.details && <Typography variant="caption" sx={{ display: 'block' }}>{state.error.details}</Typography>}
                    </Alert>
                  )}

                  {showPreview && state.preview && (
                    <Stack spacing={2.5}>
                      <Divider />
                      <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2}>
                        {state.preview.thumbnailUrl ? (
                          <CardMedia
                            component="img"
                            image={state.preview.thumbnailUrl}
                            alt=""
                            sx={{ width: { xs: '100%', sm: 188 }, height: 118, borderRadius: 2, objectFit: 'cover' }}
                          />
                        ) : (
                          <Box className="preview-placeholder" aria-hidden="true"><DownloadOutlinedIcon /></Box>
                        )}
                        <Box sx={{ minWidth: 0 }}>
                          <Typography variant="overline" color="text.secondary">{displayService(state.preview.service)}</Typography>
                          <Typography variant="h6" sx={{ overflowWrap: 'anywhere' }}>{state.preview.title || 'Mídia sem título'}</Typography>
                          <Typography color="text.secondary">
                            {[state.preview.kind === 'collection' ? `${state.preview.itemCount || state.preview.items?.length || 0} itens` : state.preview.author, formatDuration(state.preview.durationSeconds)].filter(Boolean).join(' · ')}
                          </Typography>
                        </Box>
                      </Stack>

                      {state.preview.kind === 'collection' && state.preview.items?.length ? (
                        <Box sx={{ maxHeight: 176, overflow: 'auto', border: '1px solid', borderColor: 'divider', borderRadius: 2, px: 1.5, py: 0.75 }}>
                          {state.preview.items.slice(0, 5).map((item) => (
                            <Stack key={`${item.index}-${item.url}`} direction="row" spacing={1} sx={{ py: 0.5 }}>
                              <Typography variant="caption" color="text.secondary" sx={{ minWidth: 24 }}>{item.index}.</Typography>
                              <Typography variant="body2" noWrap title={item.title}>{item.title || item.url}</Typography>
                            </Stack>
                          ))}
                          {(state.preview.itemCount ?? 0) > 5 && <Typography variant="caption" color="text.secondary">+ {(state.preview.itemCount ?? 0) - 5} itens</Typography>}
                        </Box>
                      ) : null}

                      {state.duplicate && (
                        <Alert severity="warning" action={state.duplicate.filePath && !state.duplicate.missing ? <Button size="small" onClick={() => void openDownloadedFile(state.duplicate?.filePath ?? '')}>Abrir existente</Button> : undefined}>
                          Este link já aparece no histórico{state.duplicate.missing ? ', mas o arquivo não foi encontrado' : '.'}
                        </Alert>
                      )}

                      {state.preferenceNotice && <Alert severity="info">{state.preferenceNotice}</Alert>}

                      <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1.5} sx={{ display: settings.quickMode ? 'none' : 'flex' }}>
                        <FormControl fullWidth>
                          <InputLabel id="media-type-label">Tipo</InputLabel>
                          <Select
                            labelId="media-type-label"
                            label="Tipo"
                            value={state.mediaType}
                            onChange={(event) => handleMediaType(event.target.value as AppState['mediaType'])}
                            disabled={busy || serviceOptions.length < 2}
                          >
                            {serviceOptions.map((mediaType) => <MenuItem key={mediaType} value={mediaType}>{mediaType === 'audio' ? 'Áudio' : 'Vídeo'}</MenuItem>)}
                          </Select>
                        </FormControl>
                        <FormControl fullWidth>
                          <InputLabel id="format-label">Formato</InputLabel>
                          <Select
                            labelId="format-label"
                            label="Formato"
                            value={state.format}
                            onChange={(event) => dispatch({ type: 'optionChanged', format: event.target.value as Format })}
                            disabled={busy || filteredOptions.length < 2}
                          >
                            {filteredOptions.map((option) => <MenuItem key={`${option.format}-${option.quality}`} value={option.format}>{optionLabel(option)}</MenuItem>)}
                          </Select>
                        </FormControl>
                        {state.mediaType === 'video' && (
                          <FormControl fullWidth>
                            <InputLabel id="quality-label">Qualidade</InputLabel>
                            <Select
                              labelId="quality-label"
                              label="Qualidade"
                              value={state.quality}
                              onChange={(event) => dispatch({ type: 'optionChanged', quality: event.target.value })}
                              disabled={busy || filteredOptions.filter((option) => option.quality).length < 2}
                            >
                              {Array.from(new Set(filteredOptions.map((option) => option.quality).filter(Boolean))).map((quality) => <MenuItem key={quality} value={quality}>{quality === 'best' ? 'Melhor disponível' : `${quality}p`}</MenuItem>)}
                            </Select>
                          </FormControl>
                        )}
                      </Stack>
                      {settings.quickMode && <Alert severity="info">Modo rápido: suas preferências foram aplicadas. Você ainda precisa confirmar o download.</Alert>}

                      <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1.5} sx={{ alignItems: { sm: 'center' } }}>
                        <Box sx={{ flex: 1, minWidth: 0 }}>
                          <Typography variant="caption" color="text.secondary">Destino</Typography>
                          <Typography noWrap title={settings.downloadDirectory || 'Nenhuma pasta selecionada'}>
                            {settings.downloadDirectory || 'Configure uma pasta em Configurações'}
                          </Typography>
                        </Box>
                        <Button variant="contained" onClick={handleDownload} disabled={busy || state.phase !== 'ready'} startIcon={<DownloadOutlinedIcon />}>
                          {state.duplicate ? 'Baixar novamente' : 'Baixar agora'}
                        </Button>
                      </Stack>
                    </Stack>
                  )}

                  {busy && (
                    <Box aria-live="polite">
                      <Stack direction="row" spacing={2} sx={{ mb: 0.75, justifyContent: 'space-between' }}>
                        <Typography sx={{ fontWeight: 700 }}>{state.progress.message || stageLabel(state.phase)}</Typography>
                        <Typography variant="body2" color="text.secondary">{Math.round(progressPercent)}%</Typography>
                      </Stack>
                      <LinearProgress variant={progressPercent > 0 ? 'determinate' : 'indeterminate'} value={progressPercent} sx={{ height: 8, borderRadius: 8 }} />
                      <Stack direction="row" sx={{ mt: 0.75, justifyContent: 'space-between' }}>
                        <Typography variant="caption" color="text.secondary">{state.progress.currentIndex && state.progress.totalItems ? `${state.progress.currentIndex}/${state.progress.totalItems}${state.progress.itemTitle ? ` · ${state.progress.itemTitle}` : ''}` : state.progress.speed || stageLabel(state.progress.stage)}</Typography>
                        <Typography variant="caption" color="text.secondary">{formatEta(state.progress.etaSeconds)}</Typography>
                      </Stack>
                      {state.jobId && (
                        <Button color="error" variant="outlined" startIcon={<StopCircleOutlinedIcon />} onClick={handleCancel} sx={{ mt: 2 }}>
                          Cancelar download
                        </Button>
                      )}
                    </Box>
                  )}

                  {state.phase === 'completed' && (
                    <Alert severity="success" icon={<CheckCircleOutlineOutlinedIcon />} action={
                      <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1}>
                        {primaryFile && <Button size="small" onClick={() => void openDownloadedFile(primaryFile)} startIcon={<OpenInNewOutlinedIcon />}>Abrir arquivo</Button>}
                        {state.result?.directory && <Button size="small" onClick={() => void openDownloadFolder(state.result?.directory ?? '')} startIcon={<FolderOpenOutlinedIcon />}>Abrir pasta</Button>}
                      </Stack>
                    }>
                      <Typography sx={{ fontWeight: 700 }}>Download concluído.</Typography>
                      {primaryFile && <Typography variant="body2" sx={{ overflowWrap: 'anywhere' }}>{primaryFile}</Typography>}
                    </Alert>
                  )}

                  {(state.phase === 'collection-completed' || state.phase === 'collection-partial') && (
                    <Alert severity={state.phase === 'collection-partial' ? 'warning' : 'success'} icon={<CheckCircleOutlineOutlinedIcon />} action={
                      <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1}>
                        {state.result?.directory && <Button size="small" onClick={() => void openDownloadFolder(state.result?.directory ?? '')} startIcon={<FolderOpenOutlinedIcon />}>Abrir pasta</Button>}
                      </Stack>
                    }>
                      <Typography sx={{ fontWeight: 700 }}>{state.phase === 'collection-partial' ? 'Coleção concluída parcialmente.' : 'Coleção concluída.'}</Typography>
                      <Typography variant="body2">{state.result?.completedItems ?? 0} de {state.result?.totalItems ?? state.result?.completedItems ?? 0} itens baixados{state.result?.failedItems ? ` · ${state.result.failedItems} falharam` : ''}.</Typography>
                    </Alert>
                  )}

                  {state.phase === 'cancelled' && (
                    <Alert severity="info">Download cancelado. {state.result?.completedItems ? `${state.result.completedItems} itens concluídos foram preservados. ` : ''}Você pode analisar o link novamente quando quiser.</Alert>
                  )}

                  {(state.logs.length > 0 || state.progress.message) && (
                    <Accordion disableGutters elevation={0} sx={{ bgcolor: 'transparent', '&:before': { display: 'none' } }}>
                      <AccordionSummary expandIcon={<ExpandMoreIcon />}><Typography variant="body2" sx={{ fontWeight: 700 }}>Detalhes técnicos</Typography></AccordionSummary>
                      <AccordionDetails sx={{ p: 0 }}>
                        <Box className="log-panel" role="log" aria-label="Detalhes técnicos">
                          {state.logs.map((line, index) => <Typography component="div" variant="caption" key={`${line}-${index}`}>{line}</Typography>)}
                          {state.progress.message && <Typography component="div" variant="caption">{state.progress.message}</Typography>}
                        </Box>
                      </AccordionDetails>
                    </Accordion>
                  )}
                </Stack>
              </CardContent>
            </Card>

            {recent.length > 0 && (
              <Card component="section" aria-labelledby="recent-title">
                <CardContent sx={{ p: { xs: 2, sm: 3 } }}>
                  <Stack spacing={1.5}>
                    <Stack direction="row" sx={{ alignItems: 'center', justifyContent: 'space-between' }}>
                      <Box>
                        <Typography id="recent-title" variant="h6" sx={{ fontWeight: 800 }}>Recentes</Typography>
                        <Typography variant="body2" color="text.secondary">Os últimos downloads ficam somente neste computador.</Typography>
                      </Box>
                      <Button size="small" onClick={() => void handleClearHistory()}>Limpar</Button>
                    </Stack>
                    {recent.map((item) => (
                      <Stack key={item.id} direction={{ xs: 'column', sm: 'row' }} spacing={1} sx={{ alignItems: { sm: 'center' }, justifyContent: 'space-between', borderTop: '1px solid', borderColor: 'divider', pt: 1.25 }}>
                        <Box sx={{ minWidth: 0 }}>
                          <Typography sx={{ fontWeight: 700 }} noWrap title={item.title}>{item.title}</Typography>
                          <Typography variant="caption" color={item.missing ? 'warning.main' : 'text.secondary'}>{item.missing ? 'Arquivo não encontrado' : `${displayService(item.service)} · ${formatHistoryDate(item.completedAt)}`}</Typography>
                        </Box>
                        <Stack direction="row" spacing={0.5}>
                          {item.filePath && !item.missing && <Tooltip title="Abrir arquivo"><IconButton aria-label={`Abrir ${item.title}`} onClick={() => void openDownloadedFile(item.filePath ?? '')}><OpenInNewOutlinedIcon fontSize="small" /></IconButton></Tooltip>}
                          <Tooltip title="Abrir pasta"><IconButton aria-label={`Abrir pasta de ${item.title}`} onClick={() => void openDownloadFolder(item.outputDir)}><FolderOpenOutlinedIcon fontSize="small" /></IconButton></Tooltip>
                        </Stack>
                      </Stack>
                    ))}
                  </Stack>
                </CardContent>
              </Card>
            )}

            <Typography variant="caption" color="text.secondary" align="center">
              Os arquivos são processados localmente. O Audivo não envia seus links ou cookies para um serviço intermediário.
            </Typography>
          </Stack>
        </Container>
      </Box>

      <Dialog open={settingsOpen} onClose={() => !settingsBusy && setSettingsOpen(false)} fullWidth maxWidth="sm">
        <DialogTitle>Configurações</DialogTitle>
        <DialogContent dividers>
          <Stack spacing={2.5} sx={{ pt: 0.5 }}>
            {settingsMessage && <Alert severity="info">{settingsMessage}</Alert>}
            {settings.warning && <Alert severity="warning">{settings.warning}</Alert>}
            <Box>
              <Typography variant="subtitle1" sx={{ fontWeight: 800 }}>Destino e privacidade</Typography>
              <Typography variant="body2" color="text.secondary">Esses caminhos ficam salvos apenas no computador.</Typography>
            </Box>
            <Stack direction="row" spacing={1} sx={{ alignItems: 'center' }}>
              <TextField fullWidth label="Pasta de downloads" value={settingsDraft.downloadDirectory} slotProps={{ input: { readOnly: true } }} />
              <Button variant="outlined" onClick={chooseDirectory} disabled={settingsBusy} startIcon={<FolderOpenOutlinedIcon />}>Escolher</Button>
            </Stack>
            <Box>
              <Typography variant="subtitle1" sx={{ fontWeight: 800 }}>Contas e sessões</Typography>
              <Typography variant="body2" color="text.secondary" sx={{ mb: 1.5 }}>
                O login abre uma janela visível do provedor e a sessão fica armazenada somente neste computador.
              </Typography>
              <Stack spacing={1.5}>
                {(['apple_music', 'youtube'] as Service[]).map((provider) => {
                  const status = providerAuth[provider]
                  const actionBusy = authBusyProvider === provider
                  const anotherSessionActive = authSessionActive && !actionBusy && status.state !== 'opening' && status.state !== 'waiting_login'
                  const waiting = status.state === 'opening' || status.state === 'waiting_login'
                  return (
                    <Box key={provider} sx={{ p: 1.5, border: '1px solid', borderColor: 'divider', borderRadius: 2 }}>
                      <Stack spacing={1.25}>
                        <Stack direction="row" spacing={1} sx={{ alignItems: 'flex-start', justifyContent: 'space-between' }}>
                          <Box>
                            <Typography sx={{ fontWeight: 800 }}>{providerLabel[provider]}</Typography>
                            <Typography variant="body2" color={status.error ? 'error.main' : status.connected ? 'success.main' : 'text.secondary'} aria-live="polite">
                              {providerAuthLabel(status.state)}
                            </Typography>
                            {status.message && status.state !== 'connected' && <Typography variant="caption" color="text.secondary">{status.message}</Typography>}
                            {status.error && <Typography variant="caption" color="error.main">{status.error.message}</Typography>}
                          </Box>
                          {status.connected && <CheckCircleOutlineOutlinedIcon color="success" fontSize="small" aria-label="Conectado" />}
                        </Stack>
                        <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1} sx={{ flexWrap: 'wrap' }}>
                          {waiting ? (
                            <>
                              <Button size="small" variant="contained" onClick={() => void handleVerify(provider)} disabled={settingsBusy || actionBusy} startIcon={actionBusy ? <CircularProgress size={15} /> : <RefreshOutlinedIcon />}>
                                Verificar sessão
                              </Button>
                              <Button size="small" color="error" variant="outlined" onClick={() => void handleCancelAuth(provider)} disabled={settingsBusy || actionBusy} startIcon={<StopCircleOutlinedIcon />}>
                                Cancelar
                              </Button>
                            </>
                          ) : (
                            <Button size="small" variant="contained" onClick={() => void handleConnect(provider)} disabled={settingsBusy || actionBusy || anotherSessionActive} startIcon={actionBusy ? <CircularProgress size={15} /> : <OpenInNewOutlinedIcon />}>
                              Conectar
                            </Button>
                          )}
                          <Button size="small" variant="outlined" onClick={() => void chooseCookies(provider)} disabled={settingsBusy}>
                            Importar arquivo
                          </Button>
                          {status.connected && <Button size="small" color="error" onClick={() => void handleDisconnect(provider)} disabled={settingsBusy || actionBusy}>
                            Desconectar
                          </Button>}
                        </Stack>
                      </Stack>
                    </Box>
                  )
                })}
              </Stack>
            </Box>
            <Divider />
            <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1.5}>
              <FormControl fullWidth>
                <InputLabel id="theme-label">Tema</InputLabel>
                <Select labelId="theme-label" label="Tema" value={settingsDraft.theme} onChange={(event) => setSettingsDraft({ ...settingsDraft, theme: event.target.value as Settings['theme'] })}>
                  <MenuItem value="system">Usar tema do sistema</MenuItem>
                  <MenuItem value="light">Claro</MenuItem>
                  <MenuItem value="dark">Escuro</MenuItem>
                </Select>
              </FormControl>
              <FormControl fullWidth>
                <InputLabel id="audio-format-label">Formato de áudio</InputLabel>
                <Select labelId="audio-format-label" label="Formato de áudio" value={settingsDraft.audioFormat} onChange={(event) => setSettingsDraft({ ...settingsDraft, audioFormat: event.target.value as Settings['audioFormat'] })}>
                  <MenuItem value="mp3">MP3</MenuItem>
                  <MenuItem value="m4a">M4A</MenuItem>
                </Select>
              </FormControl>
              <FormControl fullWidth>
                <InputLabel id="video-quality-label">Qualidade de vídeo</InputLabel>
                <Select labelId="video-quality-label" label="Qualidade de vídeo" value={settingsDraft.videoQuality} onChange={(event) => setSettingsDraft({ ...settingsDraft, videoQuality: event.target.value as Settings['videoQuality'] })}>
                  <MenuItem value="best">Melhor disponível</MenuItem>
                  {(['2160', '1440', '1080', '720', '540', '480', '360'] as Quality[]).map((quality) => <MenuItem key={quality} value={quality}>{quality}p</MenuItem>)}
                </Select>
              </FormControl>
            </Stack>
            <FormControlLabel control={<Switch checked={settingsDraft.quickMode} onChange={(event) => setSettingsDraft({ ...settingsDraft, quickMode: event.target.checked })} />} label="Modo rápido (usa os padrões salvos)" />
            <Divider />
            <Stack direction="row" sx={{ alignItems: 'center', justifyContent: 'space-between' }}>
              <Box>
                <Typography variant="subtitle1" sx={{ fontWeight: 800 }}>Histórico local</Typography>
                <Typography variant="body2" color="text.secondary">Limpar remove apenas os registros, não os arquivos baixados.</Typography>
              </Box>
              <Button color="error" variant="outlined" onClick={() => void handleClearHistory()} disabled={settingsBusy}>Limpar histórico</Button>
            </Stack>
            <Divider />
            <Stack spacing={1}>
              <Stack direction="row" sx={{ alignItems: 'center', justifyContent: 'space-between' }}>
                <Box>
                  <Typography variant="subtitle1" sx={{ fontWeight: 800 }}>Atualizações</Typography>
                  <Typography variant="body2" color="text.secondary">Somente releases estáveis; nada é instalado automaticamente.</Typography>
                </Box>
                <Button onClick={() => void loadUpdate()} disabled={updateBusy} startIcon={updateBusy ? <CircularProgress size={16} /> : <RefreshOutlinedIcon />}>Verificar</Button>
              </Stack>
              {updateResponse?.available && <Alert severity="info" action={updateResponse.releaseUrl ? <Button size="small" onClick={() => void openRelease(updateResponse.releaseUrl ?? '')}>Ver atualização</Button> : undefined}>Nova versão disponível: {updateResponse.latestVersion}</Alert>}
              {updateResponse?.error && <Typography variant="caption" color="text.secondary">{updateResponse.error.message}</Typography>}
            </Stack>
            <Divider />
            <Stack direction="row" sx={{ justifyContent: 'space-between', alignItems: 'center' }}>
              <Box>
                <Typography variant="subtitle1" sx={{ fontWeight: 800 }}>Engines</Typography>
                <Typography variant="body2" color="text.secondary">Executáveis locais e versões detectadas.</Typography>
              </Box>
              <Button onClick={handleUpdateEngines} disabled={engineBusy} startIcon={engineBusy ? <CircularProgress size={16} /> : <RefreshOutlinedIcon />}>Atualizar</Button>
            </Stack>
            <Stack spacing={1}>
              {(engineResponse?.engines ?? []).map((engine) => (
                <Box key={engine.name} className="engine-row">
                  <Box>
                    <Typography variant="body2" sx={{ fontWeight: 700 }}>{engine.name}</Typography>
                    <Typography variant="caption" color="text.secondary">{engine.version || engine.message || 'Não detectado'}</Typography>
                  </Box>
                  <Typography variant="caption" color={engine.valid ? 'success.main' : 'warning.main'}>{engine.valid ? 'Pronto' : 'Ausente'}</Typography>
                </Box>
              ))}
              {!engineResponse?.engines.length && !engineBusy && <Typography variant="body2" color="text.secondary">Abra as configurações para consultar os engines.</Typography>}
            </Stack>
            <Divider />
            <Stack spacing={1.25}>
              <Stack direction="row" sx={{ alignItems: 'center', justifyContent: 'space-between' }}>
                <Box>
                  <Typography variant="subtitle1" sx={{ fontWeight: 800 }}>Diagnóstico</Typography>
                  <Typography variant="body2" color="text.secondary">Relatório local, resumido e sem cookies.</Typography>
                </Box>
                <Button onClick={() => void loadDiagnostic()} disabled={diagnosticBusy} startIcon={diagnosticBusy ? <CircularProgress size={16} /> : <RefreshOutlinedIcon />}>Atualizar</Button>
              </Stack>
              {diagnosticResponse?.report && <TextField value={diagnosticResponse.report} multiline minRows={5} slotProps={{ input: { readOnly: true } }} />}
              <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1}>
                <Button variant="outlined" onClick={() => void handleCopyDiagnostic()} disabled={diagnosticBusy}>Copiar diagnóstico</Button>
                <Button variant="outlined" onClick={() => void openGitHubIssues()} disabled={diagnosticBusy}>Abrir GitHub Issues</Button>
              </Stack>
            </Stack>
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setSettingsOpen(false)} disabled={settingsBusy}>Fechar</Button>
          <Button variant="contained" onClick={savePreferences} disabled={settingsBusy}>Salvar</Button>
        </DialogActions>
      </Dialog>
    </ThemeProvider>
  )
}

function stageLabel(stage?: string) {
  switch (stage) {
    case 'analyzing': return 'Analisando link…'
    case 'queued': return 'Na fila…'
    case 'preparing': return 'Preparando…'
    case 'retrying': return 'Tentando novamente…'
    case 'collection-downloading': return 'Baixando coleção…'
    case 'downloading':
    case 'downloading_audio': return 'Baixando…'
    case 'merging': return 'Unindo áudio e vídeo…'
    case 'converting': return 'Convertendo formato…'
    case 'finalizing': return 'Finalizando arquivo…'
    case 'completed': return 'Concluído'
    default: return 'Processando…'
  }
}

export default App
