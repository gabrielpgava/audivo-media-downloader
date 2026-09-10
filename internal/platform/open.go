package platform

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func OpenPath(path string, reveal bool) error {
	if path == "" || !filepath.IsAbs(path) {
		return errors.New("path must be absolute")
	}
	if _, err := os.Stat(path); err != nil {
		return err
	}

	var command string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		command = "open"
		if reveal {
			args = append(args, "-R")
		}
		args = append(args, path)
	case "windows":
		command = "explorer.exe"
		if reveal {
			args = append(args, "/select,"+path)
		} else {
			args = append(args, path)
		}
	default:
		command = "xdg-open"
		args = append(args, path)
	}
	return exec.Command(command, args...).Run()
}
