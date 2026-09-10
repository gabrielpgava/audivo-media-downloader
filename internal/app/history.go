package app

import (
	"path/filepath"
	"time"

	"audivo-media-downloader/internal/download"
	"audivo-media-downloader/internal/models"
)

func (s *Service) recordHistory(request models.DownloadRequest, service models.Service, result models.DownloadResult) {
	s.mu.RLock()
	store := s.history
	s.mu.RUnlock()
	if store == nil {
		return
	}
	kind := models.MediaKindSingle
	if request.CollectionTitle != "" || len(request.CollectionItems) > 0 || result.CompletedItems > 0 {
		kind = models.MediaKindCollection
	}
	title := request.Title
	if title == "" {
		title = filepath.Base(request.URL)
	}
	status := "completed"
	if result.Partial {
		status = "partial"
	}
	_ = store.Add(models.HistoryItem{
		SourceURL:      download.NormalizeInput(request.URL),
		Service:        service,
		Title:          title,
		Type:           kind,
		FilePath:       firstFile(result.Files),
		OutputDir:      result.Directory,
		CompletedAt:    time.Now().UTC().Format(time.RFC3339),
		Status:         status,
		ItemCount:      result.TotalItems,
		CompletedItems: result.CompletedItems,
	})
}

func firstFile(files []string) string {
	if len(files) == 0 {
		return ""
	}
	return files[0]
}

func (s *Service) RecentDownloads(limit int) models.HistoryResponse {
	s.mu.RLock()
	store := s.history
	s.mu.RUnlock()
	if store == nil {
		return models.HistoryResponse{Items: []models.HistoryItem{}}
	}
	items, err := store.Recent(limit)
	if err != nil {
		return models.HistoryResponse{Items: []models.HistoryItem{}}
	}
	return models.HistoryResponse{Items: items}
}

func (s *Service) ClearHistory() models.ActionResponse {
	s.mu.RLock()
	store := s.history
	s.mu.RUnlock()
	if store == nil {
		return models.ActionResponse{Error: &models.UserError{Code: "history_unavailable", Message: "O histórico local não está disponível.", Retryable: true}}
	}
	if err := store.Clear(); err != nil {
		return models.ActionResponse{Error: &models.UserError{Code: "history_clear_failed", Message: "Não foi possível limpar o histórico.", Details: err.Error(), Retryable: true}}
	}
	return models.ActionResponse{OK: true}
}
