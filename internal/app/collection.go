package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"audivo-media-downloader/internal/download"
	"audivo-media-downloader/internal/models"
)

const collectionRetryLimit = 3

type collectionItemDownloader func(context.Context, string, models.DownloadRequest, map[string]struct{}, string, func(models.DownloadEvent), func(models.TechnicalLogLine)) (models.DownloadResult, error)

func (s *Service) downloadCollection(ctx context.Context, jobID string, request models.DownloadRequest, tempDir string, emit func(models.DownloadEvent), log func(models.TechnicalLogLine)) (models.DownloadResult, error) {
	return s.downloadCollectionWith(ctx, jobID, request, tempDir, emit, log, s.downloadOne)
}

func (s *Service) downloadCollectionWith(ctx context.Context, jobID string, request models.DownloadRequest, tempDir string, emit func(models.DownloadEvent), log func(models.TechnicalLogLine), downloadItem collectionItemDownloader) (models.DownloadResult, error) {
	collectionDir := download.CollectionDirectory(request.OutputDir, request.CollectionTitle, "", "")
	if err := prepareOutputDirectory(collectionDir); err != nil {
		return models.DownloadResult{}, download.NewDetailedError("output_directory_invalid", "Escolha uma pasta de destino válida e gravável.", err.Error(), false)
	}
	items := request.CollectionItems
	if len(items) == 0 {
		// Apple Music's native collection URL is already a single gamdl job. It
		// still gets the collection result contract and a predictable directory.
		request.OutputDir = collectionDir
		before := snapshotMediaFiles(collectionDir)
		result, err := s.downloadOne(ctx, jobID, request, before, tempDir, emit, log)
		if err != nil {
			return models.DownloadResult{}, err
		}
		result.TotalItems = 1
		result.CompletedItems = 1
		return result, nil
	}

	total := len(items)
	result := models.DownloadResult{Directory: collectionDir, TotalItems: total}
	for index, item := range items {
		if ctx.Err() != nil {
			result.Partial = result.CompletedItems > 0
			return result, ctx.Err()
		}
		itemTitle := item.Title
		if itemTitle == "" {
			itemTitle = fmt.Sprintf("Item %d", index+1)
		}
		if existing := existingCollectionFile(collectionDir, itemTitle, request); existing != "" {
			result.CompletedItems++
			result.Files = append(result.Files, existing)
			emit(models.DownloadEvent{JobID: jobID, State: models.StateCollection, Stage: models.StageFinalizing, CurrentIndex: index + 1, TotalItems: total, ItemTitle: itemTitle, ItemPercent: 100, OverallPercent: overallPercent(index, 100, total), Percent: overallPercent(index, 100, total), Message: fmt.Sprintf("%d de %d já concluídos", index+1, total)})
			continue
		}
		requestForItem := request
		requestForItem.URL = item.URL
		requestForItem.CollectionItems = nil
		requestForItem.CollectionTitle = ""
		requestForItem.OutputDir = collectionDir
		var itemResult models.DownloadResult
		var itemErr error
		for attempt := 1; attempt <= collectionRetryLimit; attempt++ {
			if attempt > 1 {
				emit(models.DownloadEvent{JobID: jobID, State: models.StateRetrying, Stage: models.StageDownloading, CurrentIndex: index + 1, TotalItems: total, ItemTitle: itemTitle, RetryAttempt: attempt, RetryLimit: collectionRetryLimit, OverallPercent: overallPercent(index, 0, total), Message: fmt.Sprintf("Reconectando… tentativa %d de %d", attempt, collectionRetryLimit)})
			}
			before := snapshotMediaFiles(collectionDir)
			itemResult, itemErr = downloadItem(ctx, jobID, requestForItem, before, tempDir, func(event models.DownloadEvent) {
				event.JobID = jobID
				event.CurrentIndex = index + 1
				event.TotalItems = total
				event.ItemTitle = itemTitle
				event.ItemPercent = event.Percent
				event.OverallPercent = overallPercent(index, event.Percent, total)
				event.Percent = event.OverallPercent
				emit(event)
			}, log)
			if itemErr == nil {
				break
			}
			if errors.Is(itemErr, context.Canceled) || ctx.Err() != nil {
				break
			}
			if !retryableCollectionError(itemErr) || attempt == collectionRetryLimit {
				break
			}
		}
		if itemErr != nil {
			if errors.Is(itemErr, context.Canceled) || ctx.Err() != nil {
				result.Partial = result.CompletedItems > 0
				return result, ctx.Err()
			}
			result.FailedItems++
			log(models.TechnicalLogLine{JobID: jobID, Stream: "audivo", Message: fmt.Sprintf("%s: %s", itemTitle, safeErrorMessage(itemErr))})
			continue
		}
		result.CompletedItems++
		result.Files = append(result.Files, itemResult.Files...)
		emit(models.DownloadEvent{JobID: jobID, State: models.StateCollection, Stage: models.StageFinalizing, CurrentIndex: index + 1, TotalItems: total, ItemTitle: itemTitle, ItemPercent: 100, OverallPercent: overallPercent(index, 100, total), Percent: overallPercent(index, 100, total), Message: fmt.Sprintf("%d de %d concluídos", index+1, total)})
	}
	result.Partial = result.FailedItems > 0
	return result, nil
}

// ponytail: title-based resume is intentionally conservative; a provider can
// use a richer filename template later if duplicate titles need disambiguation.
func existingCollectionFile(directory, title string, request models.DownloadRequest) string {
	if title == "" {
		return ""
	}
	prefix := download.SanitizeComponent(title) + "."
	entries, err := os.ReadDir(directory)
	if err != nil {
		return ""
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), prefix) {
			continue
		}
		extension := strings.ToLower(filepath.Ext(entry.Name()))
		if request.MediaType == models.MediaTypeAudio && (extension == ".mp3" || extension == ".m4a" || extension == ".aac" || extension == ".mp4") {
			return filepath.Join(directory, entry.Name())
		}
		if request.MediaType == models.MediaTypeVideo && (extension == ".mp4" || extension == ".m4v" || extension == ".mkv" || extension == ".webm") {
			return filepath.Join(directory, entry.Name())
		}
	}
	return ""
}

func (s *Service) downloadOne(ctx context.Context, jobID string, request models.DownloadRequest, before map[string]struct{}, tempDir string, emit func(models.DownloadEvent), log func(models.TechnicalLogLine)) (models.DownloadResult, error) {
	service, userErr := s.registry.Detect(request.URL)
	if userErr != nil {
		return models.DownloadResult{}, &download.UserError{Value: userErr}
	}
	switch service {
	case models.ServiceYouTube:
		return s.downloadYouTube(ctx, jobID, request, before, emit, log)
	case models.ServiceAppleMusic:
		return s.downloadAppleMusic(ctx, jobID, request, before, tempDir, emit, log)
	default:
		return models.DownloadResult{}, download.NewError("unsupported_url", "Cole um link do YouTube ou Apple Music.", false)
	}
}

func overallPercent(index int, itemPercent float64, total int) float64 {
	if total <= 0 {
		return itemPercent
	}
	return (float64(index) + itemPercent/100) / float64(total) * 100
}

func retryableCollectionError(err error) bool {
	if err == nil {
		return false
	}
	userErr := download.AsUserError(err)
	if !userErr.Retryable {
		return false
	}
	switch userErr.Code {
	case "unsupported_url", "cookies_missing", "cookies_invalid", "output_directory_invalid", "ffmpeg_missing":
		return false
	default:
		return true
	}
}

func safeErrorMessage(err error) string {
	message := strings.TrimSpace(err.Error())
	if message == "" {
		return "falha sem detalhes"
	}
	return filepath.Base(message)
}
