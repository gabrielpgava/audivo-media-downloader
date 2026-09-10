package notifications

import "audivo-media-downloader/internal/models"

type Content struct {
	Title string
	Body  string
}

// ShouldSend keeps notifications limited to one final job event and avoids
// duplicate feedback while the Audivo window is visible.
func ShouldSend(event models.DownloadEvent, focused, available bool) bool {
	if focused || !available {
		return false
	}
	switch event.State {
	case models.StateCompleted, models.StateCollectionCompleted, models.StatePartial, models.StateError:
		return true
	default:
		return false
	}
}

func ContentFor(event models.DownloadEvent) Content {
	if event.State == models.StateError {
		body := "O download não foi concluído."
		if event.Error != nil && event.Error.Message != "" {
			body = event.Error.Message
		}
		return Content{Title: "Download não concluído", Body: body}
	}
	body := event.Message
	if body == "" {
		body = "O download foi concluído."
	}
	if event.State == models.StatePartial {
		return Content{Title: "Coleção parcialmente concluída", Body: body}
	}
	return Content{Title: "Download concluído", Body: body}
}
