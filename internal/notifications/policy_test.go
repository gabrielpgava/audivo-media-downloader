package notifications

import (
	"testing"

	"audivo-media-downloader/internal/models"
)

func TestShouldSendOnlyFinalBackgroundEvents(t *testing.T) {
	tests := []struct {
		name      string
		state     models.JobState
		focused   bool
		available bool
		want      bool
	}{
		{name: "single completed in background", state: models.StateCompleted, available: true, want: true},
		{name: "partial collection in background", state: models.StatePartial, available: true, want: true},
		{name: "foreground completed", state: models.StateCompleted, focused: true, available: true},
		{name: "unsupported platform", state: models.StateCompleted, available: false},
		{name: "progress is not final", state: models.StateDownloading, available: true},
		{name: "cancel is not notification", state: models.StateCancelled, available: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := ShouldSend(models.DownloadEvent{State: test.state}, test.focused, test.available); got != test.want {
				t.Fatalf("ShouldSend() = %t, want %t", got, test.want)
			}
		})
	}
}

func TestContentForReducesErrorToSafeUserMessage(t *testing.T) {
	content := ContentFor(models.DownloadEvent{State: models.StateError, Error: &models.UserError{Message: "Atualize a sessão do provedor."}})
	if content.Title != "Download não concluído" || content.Body != "Atualize a sessão do provedor." {
		t.Fatalf("unexpected error notification: %+v", content)
	}
}
