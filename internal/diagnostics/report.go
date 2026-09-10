package diagnostics

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"audivo-media-downloader/internal/models"
)

var sensitiveValue = regexp.MustCompile(`(?i)(cookie|token|authorization|password|secret)(?:\s*[:=]\s*|\s+)\S+`)

type Input struct {
	Version   string
	Providers map[models.Service]string
	Engines   []models.EngineStatus
	Operation string
	Result    string
	Error     string
	Paths     []string
}

func Build(input Input) string {
	home, _ := os.UserHomeDir()
	var output strings.Builder
	line := func(name, value string) {
		output.WriteString(name)
		output.WriteString(": ")
		output.WriteString(Redact(value, home, input.Paths...))
		output.WriteByte('\n')
	}

	line("Audivo", input.Version)
	line("OS", runtime.GOOS)
	line("Arch", runtime.GOARCH)
	line("Wails", "v2")
	for _, provider := range []models.Service{models.ServiceYouTube, models.ServiceAppleMusic} {
		line("Provider "+string(provider), input.Providers[provider])
	}
	for _, engine := range input.Engines {
		status := "unavailable"
		if engine.Valid {
			status = "ready"
		}
		value := status
		if engine.Version != "" {
			value += " (" + engine.Version + ")"
		}
		if engine.Message != "" {
			value += ": " + engine.Message
		}
		if engine.Path != "" {
			value += " [" + engine.Path + "]"
		}
		line("Engine "+engine.Name, value)
	}
	if input.Operation != "" {
		line("Operation", input.Operation)
	}
	if input.Result != "" {
		line("Result", input.Result)
	}
	if input.Error != "" {
		line("Error", input.Error)
	}
	return strings.TrimSpace(output.String()) + "\n"
}

func Redact(value string, home string, sensitivePaths ...string) string {
	for _, path := range sensitivePaths {
		if path != "" {
			value = strings.ReplaceAll(value, path, "[redacted-path]")
		}
	}
	if home != "" {
		cleanHome := filepath.Clean(home)
		value = strings.ReplaceAll(value, cleanHome, "~")
	}
	return sensitiveValue.ReplaceAllString(value, "$1=[redacted]")
}
