package download

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// TempOutcome controls which operation-owned temporary files can survive a
// job. Only resumable engine markers survive an interrupted operation.
type TempOutcome uint8

const (
	TempSucceeded TempOutcome = iota
	TempPermanentFailure
	TempRecoverableFailure
	TempPartial
	TempCancelled
)

// NewTempWorkspace creates an isolated workspace for one download job.
func NewTempWorkspace(root, jobID string) (string, error) {
	if root == "" || !filepath.IsAbs(root) {
		return "", errors.New("temporary root must be absolute")
	}
	if strings.TrimSpace(jobID) == "" {
		return "", errors.New("job id is required")
	}
	if err := os.MkdirAll(root, 0o700); err != nil {
		return "", err
	}
	return os.MkdirTemp(root, ".audivo-"+SanitizeComponent(jobID)+"-*")
}

// FinalizeTempWorkspace never accepts an arbitrary path. It only operates on
// workspaces created by NewTempWorkspace and never touches the output folder.
func FinalizeTempWorkspace(path string, outcome TempOutcome) error {
	if path == "" {
		return nil
	}
	if !filepath.IsAbs(path) || !strings.HasPrefix(filepath.Base(filepath.Clean(path)), ".audivo-") {
		return errors.New("temporary workspace is not Audivo-owned")
	}
	switch outcome {
	case TempSucceeded, TempPermanentFailure:
		return os.RemoveAll(path)
	case TempRecoverableFailure, TempPartial, TempCancelled:
		return removeUnsafeTempFiles(path)
	default:
		return errors.New("unknown temporary cleanup outcome")
	}
}

func removeUnsafeTempFiles(directory string) error {
	entries, err := os.ReadDir(directory)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		path := filepath.Join(directory, entry.Name())
		if entry.IsDir() {
			if err := removeUnsafeTempFiles(path); err != nil {
				return err
			}
			children, err := os.ReadDir(path)
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}
			if len(children) == 0 {
				if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
					return err
				}
			}
			continue
		}
		if isResumableTemp(path) {
			continue
		}
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return nil
}

func isResumableTemp(path string) bool {
	lower := strings.ToLower(filepath.Base(path))
	for _, suffix := range []string{".part", ".ytdl", ".ytdl.part", ".download", ".partial"} {
		if strings.HasSuffix(lower, suffix) {
			return true
		}
	}
	return false
}
