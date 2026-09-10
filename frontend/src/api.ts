import {
  AnalyzeURL,
  CancelDownload,
  CancelProviderAuth,
  DisconnectProvider,
  GetEngineStatus,
  GetProviderAuthStatus,
  GetSettings,
  OpenDownloadedFile,
  OpenDownloadFolder,
  OpenGitHubIssues,
  OpenRelease,
  RecentDownloads,
  ClearHistory,
  CopyDiagnostic,
  Diagnostic,
  CheckForUpdate,
  SaveSettings,
  SelectAppleMusicCookies,
  SelectProviderCookies,
  SelectDownloadDirectory,
  StartProviderAuth,
  StartDownload,
  SetWindowFocused,
  UpdateEngines,
  VerifyProviderAuth,
} from '../wailsjs/go/main/App'
import { models } from '../wailsjs/go/models'
import type { ActionResponse, AnalyzeResponse, DiagnosticResponse, DownloadRequest, EngineResponse, HistoryResponse, ProviderAuthStatus, Service, Settings, StartDownloadResponse, UpdateResponse } from './types'

export const analyzeURL = (url: string) => AnalyzeURL(url) as Promise<AnalyzeResponse>
export const startDownload = (request: DownloadRequest) => StartDownload(new models.DownloadRequest(request)) as Promise<StartDownloadResponse>
export const setWindowFocused = (focused: boolean) => SetWindowFocused(focused)
export const cancelDownload = (jobId: string) => CancelDownload(jobId) as Promise<ActionResponse>
export const recentDownloads = (limit = 5) => RecentDownloads(limit) as Promise<HistoryResponse>
export const clearHistory = () => ClearHistory() as Promise<ActionResponse>
export const diagnostic = () => Diagnostic() as Promise<DiagnosticResponse>
export const copyDiagnostic = () => CopyDiagnostic() as Promise<ActionResponse>
export const openGitHubIssues = () => OpenGitHubIssues() as Promise<ActionResponse>
export const checkForUpdate = () => CheckForUpdate() as Promise<UpdateResponse>
export const openRelease = (url: string) => OpenRelease(url) as Promise<ActionResponse>
export const getSettings = () => GetSettings() as Promise<Settings>
export const saveSettings = (settings: Settings) => SaveSettings(new models.SettingsView(settings)) as Promise<ActionResponse>
export const selectDownloadDirectory = () => SelectDownloadDirectory() as Promise<ActionResponse>
export const selectAppleMusicCookies = () => SelectAppleMusicCookies() as Promise<ActionResponse>
export const selectProviderCookies = (provider: Service) => SelectProviderCookies(provider) as Promise<ActionResponse>
export const startProviderAuth = (provider: Service) => StartProviderAuth(provider) as Promise<ProviderAuthStatus>
export const verifyProviderAuth = (provider: Service) => VerifyProviderAuth(provider) as Promise<ProviderAuthStatus>
export const getProviderAuthStatus = (provider: Service) => GetProviderAuthStatus(provider) as Promise<ProviderAuthStatus>
export const cancelProviderAuth = (provider: Service) => CancelProviderAuth(provider) as Promise<ProviderAuthStatus>
export const disconnectProvider = (provider: Service) => DisconnectProvider(provider) as Promise<ProviderAuthStatus>
export const getEngineStatus = () => GetEngineStatus() as Promise<EngineResponse>
export const updateEngines = () => UpdateEngines() as Promise<EngineResponse>
export const openDownloadedFile = (path: string) => OpenDownloadedFile(path) as Promise<ActionResponse>
export const openDownloadFolder = (path: string) => OpenDownloadFolder(path) as Promise<ActionResponse>
