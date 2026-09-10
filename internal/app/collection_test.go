package app

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"audivo-media-downloader/internal/download"
	"audivo-media-downloader/internal/models"
)

func TestExistingCollectionFileSupportsResumeWithoutScanningOutsideDestination(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "First - Song.mp3")
	if err := os.WriteFile(path, []byte("done"), 0o600); err != nil {
		t.Fatal(err)
	}
	request := models.DownloadRequest{MediaType: models.MediaTypeAudio, Format: models.FormatMP3}
	if got := existingCollectionFile(directory, "First - Song", request); got != path {
		t.Fatalf("existing collection file = %q, want %q", got, path)
	}
	if got := existingCollectionFile(directory, "Missing", request); got != "" {
		t.Fatalf("missing collection file = %q", got)
	}
}

func TestRetryableCollectionErrorClassification(t *testing.T) {
	if retryableCollectionError(download.NewError("cookies_missing", "missing", false)) {
		t.Fatal("missing cookies must not retry")
	}
	if !retryableCollectionError(download.NewError("engine_failed", "failed", true)) {
		t.Fatal("transient engine failure should retry")
	}
}

func TestCollectionFixtureProducesPartialResultAndPreservesFourFiles(t *testing.T) {
	root := t.TempDir()
	items := make([]models.CollectionItem, 0, 5)
	for index := 1; index <= 5; index++ {
		items = append(items, models.CollectionItem{URL: "https://www.youtube.com/watch?v=item-" + string(rune('0'+index)), Title: "Item " + string(rune('0'+index)), Index: index})
	}
	var events []models.DownloadEvent
	service := &Service{}
	result, err := service.downloadCollectionWith(
		context.Background(),
		"fixture-partial",
		models.DownloadRequest{OutputDir: root, CollectionTitle: "Summer 2026", MediaType: models.MediaTypeAudio, Format: models.FormatMP3, CollectionItems: items},
		filepath.Join(root, "tmp"),
		func(event models.DownloadEvent) { events = append(events, event) },
		func(models.TechnicalLogLine) {},
		func(_ context.Context, jobID string, request models.DownloadRequest, _ map[string]struct{}, _ string, emit func(models.DownloadEvent), _ func(models.TechnicalLogLine)) (models.DownloadResult, error) {
			itemID := request.URL[len(request.URL)-6:]
			if itemID == "item-3" {
				return models.DownloadResult{}, download.NewError("content_removed", "Conteúdo removido.", false)
			}
			path := filepath.Join(request.OutputDir, itemID+".mp3")
			if err := os.WriteFile(path, []byte(itemID), 0o600); err != nil {
				return models.DownloadResult{}, err
			}
			emit(models.DownloadEvent{JobID: jobID, State: models.StateDownloading, Percent: 100})
			return models.DownloadResult{Files: []string{path}, Directory: request.OutputDir}, nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Partial || result.TotalItems != 5 || result.CompletedItems != 4 || result.FailedItems != 1 {
		t.Fatalf("unexpected partial result: %+v", result)
	}
	if len(result.Files) != 4 {
		t.Fatalf("expected four preserved files, got %+v", result.Files)
	}
	for _, path := range result.Files {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("preserved file %q is missing: %v", path, err)
		}
	}
	if len(events) != 8 {
		t.Fatalf("expected progress and final events for four successful items, got %d", len(events))
	}
	if events[len(events)-1].CurrentIndex != 5 || events[len(events)-1].TotalItems != 5 || events[len(events)-1].OverallPercent != 100 {
		t.Fatalf("unexpected final collection progress: %+v", events[len(events)-1])
	}
}

func TestCollectionFixtureCancellationDoesNotStartThirdItem(t *testing.T) {
	root := t.TempDir()
	items := []models.CollectionItem{
		{URL: "https://www.youtube.com/watch?v=item-1", Title: "Item 1", Index: 1},
		{URL: "https://www.youtube.com/watch?v=item-2", Title: "Item 2", Index: 2},
		{URL: "https://www.youtube.com/watch?v=item-3", Title: "Item 3", Index: 3},
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	itemTwoStarted := make(chan struct{})
	itemThreeStarted := make(chan struct{})
	done := make(chan struct{})
	var result models.DownloadResult
	var resultErr error
	go func() {
		result, resultErr = (&Service{}).downloadCollectionWith(
			ctx,
			"fixture-cancel",
			models.DownloadRequest{OutputDir: root, CollectionTitle: "Cancel fixture", MediaType: models.MediaTypeAudio, Format: models.FormatMP3, CollectionItems: items},
			filepath.Join(root, "tmp"),
			func(models.DownloadEvent) {},
			func(models.TechnicalLogLine) {},
			func(ctx context.Context, _ string, request models.DownloadRequest, _ map[string]struct{}, _ string, _ func(models.DownloadEvent), _ func(models.TechnicalLogLine)) (models.DownloadResult, error) {
				switch request.URL[len(request.URL)-6:] {
				case "item-1":
					path := filepath.Join(request.OutputDir, "item-1.mp3")
					if err := os.WriteFile(path, []byte("done"), 0o600); err != nil {
						return models.DownloadResult{}, err
					}
					return models.DownloadResult{Files: []string{path}, Directory: request.OutputDir}, nil
				case "item-2":
					close(itemTwoStarted)
					<-ctx.Done()
					return models.DownloadResult{}, ctx.Err()
				default:
					close(itemThreeStarted)
					return models.DownloadResult{}, nil
				}
			},
		)
		close(done)
	}()
	select {
	case <-itemTwoStarted:
	case <-done:
		t.Fatal("collection finished before item two started")
	}
	cancel()
	<-done
	if !errors.Is(resultErr, context.Canceled) || !result.Partial || result.CompletedItems != 1 {
		t.Fatalf("unexpected cancellation result=%+v err=%v", result, resultErr)
	}
	if _, err := os.Stat(filepath.Join(root, "Cancel fixture", "item-1.mp3")); err != nil {
		t.Fatalf("completed item was not preserved: %v", err)
	}
	select {
	case <-itemThreeStarted:
		t.Fatal("item three started after cancellation")
	default:
	}
}
