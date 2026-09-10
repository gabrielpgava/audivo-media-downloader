package history

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"audivo-media-downloader/internal/download"
	"audivo-media-downloader/internal/models"
	"audivo-media-downloader/internal/platform"
)

const (
	fileVersion = 1
	maxEntries  = 50
)

type fileData struct {
	Version int                  `json:"version"`
	Items   []models.HistoryItem `json:"items"`
}

type Store struct {
	path string
	mu   sync.Mutex
}

func NewStore(paths platform.Paths) *Store {
	return &Store{path: filepath.Join(paths.ConfigDir, "history.json")}
}

func NewStoreAt(path string) *Store { return &Store{path: path} }

func (s *Store) Path() string { return s.path }

func (s *Store) Load() ([]models.HistoryItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadLocked()
}

func (s *Store) Add(item models.HistoryItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	items, err := s.loadLocked()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if item.ID == "" {
		item.ID = time.Now().UTC().Format("20060102T150405.000000000Z07:00")
	}
	if item.CompletedAt == "" {
		item.CompletedAt = time.Now().UTC().Format(time.RFC3339)
	}
	items = append([]models.HistoryItem{item}, items...)
	if len(items) > maxEntries {
		items = items[:maxEntries]
	}
	return s.saveLocked(items)
}

func (s *Store) Recent(limit int) ([]models.HistoryItem, error) {
	items, err := s.Load()
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > len(items) {
		limit = len(items)
	}
	items = items[:limit]
	for index := range items {
		items[index].Missing = items[index].FilePath != "" && !fileExists(items[index].FilePath)
	}
	return items, nil
}

func (s *Store) Find(sourceURL string) (*models.HistoryItem, error) {
	normalized := download.NormalizeInput(sourceURL)
	items, err := s.Load()
	if err != nil {
		return nil, err
	}
	for index := range items {
		if download.NormalizeInput(items[index].SourceURL) != normalized {
			continue
		}
		items[index].Missing = items[index].FilePath != "" && !fileExists(items[index].FilePath)
		return &items[index], nil
	}
	return nil, nil
}

func (s *Store) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := os.Remove(s.path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func (s *Store) loadLocked() ([]models.HistoryItem, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []models.HistoryItem{}, nil
		}
		return nil, err
	}
	var stored fileData
	if err := json.Unmarshal(data, &stored); err != nil {
		return []models.HistoryItem{}, nil
	}
	if stored.Version != fileVersion {
		return []models.HistoryItem{}, nil
	}
	if len(stored.Items) > maxEntries {
		stored.Items = stored.Items[:maxEntries]
	}
	return stored.Items, nil
}

func (s *Store) saveLocked(items []models.HistoryItem) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(fileData{Version: fileVersion, Items: items}, "", "  ")
	if err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(s.path), ".history-*.tmp")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if err := temp.Chmod(0o600); err != nil {
		_ = temp.Close()
		return err
	}
	if _, err := temp.Write(append(data, '\n')); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(tempPath, s.path)
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
