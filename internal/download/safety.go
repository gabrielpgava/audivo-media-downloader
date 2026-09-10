package download

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"audivo-media-downloader/internal/platform"
)

const (
	maxPathComponentRunes = 120
	spaceSafetyMultiplier = 1.20
)

// SanitizeComponent keeps provider titles useful to people while removing the
// characters and reserved names that make paths fail on Windows.
func SanitizeComponent(value string) string {
	value = strings.TrimSpace(value)
	var builder strings.Builder
	for _, r := range value {
		switch {
		case r < 0x20 || strings.ContainsRune(`/\\:*?"<>|`, r):
			builder.WriteRune('-')
		case unicode.IsSpace(r):
			builder.WriteRune(' ')
		default:
			builder.WriteRune(r)
		}
	}
	value = strings.Trim(strings.Join(strings.Fields(builder.String()), " "), " .")
	if value == "" {
		value = "arquivo"
	}
	if isWindowsReserved(value) {
		value = "_" + value
	}
	runes := []rune(value)
	if len(runes) > maxPathComponentRunes {
		value = string(runes[:maxPathComponentRunes])
		value = strings.TrimRight(value, " .")
	}
	return value
}

func isWindowsReserved(value string) bool {
	base := strings.ToUpper(strings.TrimSpace(strings.TrimSuffix(value, filepath.Ext(value))))
	if base == "CON" || base == "PRN" || base == "AUX" || base == "NUL" {
		return true
	}
	if len(base) == 4 && (strings.HasPrefix(base, "COM") || strings.HasPrefix(base, "LPT")) {
		return base[3] >= '1' && base[3] <= '9'
	}
	return false
}

func CollectionDirectory(root, collection, artist, album string) string {
	root = filepath.Clean(root)
	if album != "" && artist != "" {
		return filepath.Join(root, SanitizeComponent(artist), SanitizeComponent(album))
	}
	if collection != "" {
		return filepath.Join(root, SanitizeComponent(collection))
	}
	return filepath.Join(root, "Audivo")
}

// UniquePath never overwrites an existing file. It is intentionally small and
// serial callers are expected; the engine itself still receives no-overwrite.
func UniquePath(directory, filename string) (string, error) {
	if directory == "" || !filepath.IsAbs(directory) {
		return "", errors.New("directory must be absolute")
	}
	filename = SanitizeComponent(strings.TrimSuffix(filename, filepath.Ext(filename))) + filepath.Ext(filename)
	path := filepath.Join(directory, filename)
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return path, nil
	} else if err != nil {
		return "", err
	}
	base, ext := strings.TrimSuffix(filename, filepath.Ext(filename)), filepath.Ext(filename)
	for index := 2; index < 10000; index++ {
		candidate := filepath.Join(directory, base+" ("+strconv.Itoa(index)+")"+ext)
		if _, err := os.Stat(candidate); errors.Is(err, os.ErrNotExist) {
			return candidate, nil
		} else if err != nil {
			return "", err
		}
	}
	return "", errors.New("could not resolve a free destination name")
}

// HasEnoughSpace reports false only when the platform can answer the query and
// the estimate plus a conversion margin does not fit. Unknown space is safe to
// continue because some virtual/network filesystems do not expose a value.
func HasEnoughSpace(path string, estimatedBytes uint64) (bool, uint64, error) {
	if estimatedBytes == 0 {
		return true, 0, nil
	}
	free, err := platform.FreeBytes(path)
	if err != nil || free == 0 {
		return true, free, err
	}
	needed := uint64(float64(estimatedBytes) * spaceSafetyMultiplier)
	return free >= needed, free, nil
}
