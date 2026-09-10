package logstore

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"
)

type Store struct {
	path     string
	maxBytes int64
	maxFiles int
	mu       sync.Mutex
}

var sensitiveArgument = regexp.MustCompile(`(?i)(--cookies(?:-path)?|token|password|secret|authorization)(?:=|\s+)\S+`)

func New(path string, maxBytes int64, maxFiles int) *Store {
	if maxBytes < 1 {
		maxBytes = 1 << 20
	}
	if maxFiles < 1 {
		maxFiles = 3
	}
	return &Store{path: path, maxBytes: maxBytes, maxFiles: maxFiles}
}

func (s *Store) Append(stream, message string) error {
	if s == nil || s.path == "" {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}
	line := fmt.Sprintf("%s [%s] %s\n", time.Now().UTC().Format(time.RFC3339), stream, Redact(message))
	if info, err := os.Stat(s.path); err == nil && info.Size()+int64(len(line)) > s.maxBytes {
		if err := s.rotate(); err != nil {
			return err
		}
	}
	file, err := os.OpenFile(s.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(line)
	return err
}

func (s *Store) rotate() error {
	for index := s.maxFiles - 1; index >= 1; index-- {
		from := fmt.Sprintf("%s.%d", s.path, index)
		to := fmt.Sprintf("%s.%d", s.path, index+1)
		if err := os.Rename(from, to); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	if err := os.Rename(s.path, s.path+".1"); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func Redact(message string) string {
	return sensitiveArgument.ReplaceAllString(message, "$1=[redacted]")
}
