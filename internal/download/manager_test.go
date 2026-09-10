package download

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"audivo-media-downloader/internal/models"
)

func TestJobManagerQueuesFIFOAndRestartsAfterCancel(t *testing.T) {
	var mu sync.Mutex
	var events []models.DownloadEvent
	manager := NewJobManager(func(event models.DownloadEvent) {
		mu.Lock()
		events = append(events, event)
		mu.Unlock()
	}, nil)

	started := make(chan struct{})
	first, userErr := manager.Start(context.Background(), func(ctx context.Context, jobID string, emit func(models.DownloadEvent), _ func(models.TechnicalLogLine)) (models.DownloadResult, error) {
		close(started)
		<-ctx.Done()
		return models.DownloadResult{}, ctx.Err()
	})
	if userErr != nil || first == nil {
		t.Fatalf("Start() = job %+v, error %+v", first, userErr)
	}
	<-started
	second, err := manager.Start(context.Background(), func(context.Context, string, func(models.DownloadEvent), func(models.TechnicalLogLine)) (models.DownloadResult, error) {
		return models.DownloadResult{}, nil
	})
	if err != nil || second == nil || second.State != models.StateQueued {
		t.Fatalf("expected queued result, got job %+v error %+v", second, err)
	}
	if err := manager.Cancel(first.ID); err != nil {
		t.Fatal(err)
	}
	waitForEvent(t, manager, func() bool {
		mu.Lock()
		defer mu.Unlock()
		for _, event := range events {
			if event.JobID == first.ID && event.State == models.StateCancelled {
				return true
			}
		}
		return false
	})
	waitForEvent(t, manager, func() bool {
		mu.Lock()
		defer mu.Unlock()
		for _, event := range events {
			if event.JobID == second.ID && event.State == models.StateCompleted {
				return true
			}
		}
		return false
	})
	if manager.IsActive() {
		t.Fatal("expected queued job to finish")
	}
}

func TestJobManagerPublishesDomainError(t *testing.T) {
	events := make(chan models.DownloadEvent, 4)
	manager := NewJobManager(func(got models.DownloadEvent) { events <- got }, nil)
	job, err := manager.Start(context.Background(), func(context.Context, string, func(models.DownloadEvent), func(models.TechnicalLogLine)) (models.DownloadResult, error) {
		return models.DownloadResult{}, NewError("engine_missing", "Engine ausente.", true)
	})
	if err != nil || job == nil {
		t.Fatal(err)
	}
	deadline := time.After(time.Second)
	for {
		select {
		case event := <-events:
			if event.State == models.StateError {
				if event.Error == nil || event.Error.Code != "engine_missing" {
					t.Fatalf("unexpected error event: %+v", event)
				}
				return
			}
		case <-deadline:
			t.Fatal("timed out waiting for domain error")
		}
	}
}

func TestJobManagerPublishesPartialCollectionResult(t *testing.T) {
	events := make(chan models.DownloadEvent, 4)
	manager := NewJobManager(func(event models.DownloadEvent) { events <- event }, nil)
	if _, err := manager.Start(context.Background(), func(context.Context, string, func(models.DownloadEvent), func(models.TechnicalLogLine)) (models.DownloadResult, error) {
		return models.DownloadResult{Directory: "/tmp/collection", TotalItems: 2, CompletedItems: 1, FailedItems: 1, Partial: true}, nil
	}); err != nil {
		t.Fatal(err)
	}
	deadline := time.After(time.Second)
	for {
		select {
		case event := <-events:
			if event.State != models.StatePartial {
				continue
			}
			if event.Result == nil || event.Result.CompletedItems != 1 || event.Result.FailedItems != 1 {
				t.Fatalf("unexpected partial result: %+v", event.Result)
			}
			return
		case <-deadline:
			t.Fatal("timed out waiting for partial result")
		}
	}
}

func TestAsUserError(t *testing.T) {
	if got := AsUserError(context.Canceled); got.Code != "operation_failed" {
		t.Fatalf("unexpected generic error: %+v", got)
	}
	if got := AsUserError(errors.New("boom")); got.Message == "" {
		t.Fatal("expected generic message")
	}
}

func waitForEvent(t *testing.T, _ *JobManager, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for !condition() && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if !condition() {
		t.Fatal("timed out waiting for event")
	}
}
