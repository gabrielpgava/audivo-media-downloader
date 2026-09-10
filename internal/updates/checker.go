package updates

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"audivo-media-downloader/internal/models"
)

const defaultURL = "https://api.github.com/repos/gabrielpgava/audivo-media-downloader/releases?per_page=20"

type release struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
	Draft   bool   `json:"draft"`
	Pre     bool   `json:"prerelease"`
}

type cache struct {
	CheckedAt time.Time             `json:"checkedAt"`
	Response  models.UpdateResponse `json:"response"`
}

type Checker struct {
	Path   string
	URL    string
	Client *http.Client
	Now    func() time.Time
}

func NewChecker(path string) *Checker {
	return &Checker{Path: path, URL: defaultURL, Client: &http.Client{Timeout: 10 * time.Second}, Now: time.Now}
}

func (c *Checker) Check(ctx context.Context, current string) models.UpdateResponse {
	if ctx == nil {
		ctx = context.Background()
	}
	if c == nil {
		return models.UpdateResponse{CurrentVersion: current, Error: unavailableError()}
	}
	now := c.Now()
	if cached, ok := c.load(); ok && cached.Response.CurrentVersion == current && now.Sub(cached.CheckedAt) < 24*time.Hour {
		return cached.Response
	}
	response := models.UpdateResponse{CurrentVersion: current}
	var releases []release
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL, nil)
	if err == nil {
		request.Header.Set("Accept", "application/vnd.github+json")
		request.Header.Set("User-Agent", "Audivo-Update-Check")
		client := c.Client
		if client == nil {
			client = http.DefaultClient
		}
		if httpResponse, requestErr := client.Do(request); requestErr == nil {
			defer httpResponse.Body.Close()
			if httpResponse.StatusCode >= 200 && httpResponse.StatusCode < 300 {
				err = json.NewDecoder(httpResponse.Body).Decode(&releases)
			} else {
				err = errors.New("release endpoint returned a non-success status")
			}
		} else {
			err = requestErr
		}
	}
	if err == nil {
		for _, candidate := range releases {
			if candidate.Draft || candidate.Pre || !stable(candidate.TagName) {
				continue
			}
			if newer(candidate.TagName, current) && (response.LatestVersion == "" || newer(candidate.TagName, response.LatestVersion)) {
				response.Available = true
				response.LatestVersion = trimVersion(candidate.TagName)
				response.ReleaseURL = candidate.HTMLURL
				break
			}
		}
	} else {
		response.Error = unavailableError()
	}
	_ = c.save(cache{CheckedAt: now, Response: response})
	return response
}

func (c *Checker) load() (cache, bool) {
	data, err := os.ReadFile(c.Path)
	if err != nil {
		return cache{}, false
	}
	var stored cache
	if json.Unmarshal(data, &stored) != nil || stored.CheckedAt.IsZero() {
		return cache{}, false
	}
	return stored, true
}

func (c *Checker) save(stored cache) error {
	if c.Path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(c.Path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(stored, "", "  ")
	if err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(c.Path), ".update-*.tmp")
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
	return os.Rename(tempPath, c.Path)
}

func stable(value string) bool {
	value = trimVersion(value)
	if value == "" || strings.Contains(value, "-") {
		return false
	}
	parts := strings.Split(value, ".")
	if len(parts) != 3 {
		return false
	}
	for _, part := range parts {
		if _, err := strconv.Atoi(part); err != nil {
			return false
		}
	}
	return true
}

func newer(candidate, current string) bool {
	if !stable(candidate) || !stable(current) {
		return false
	}
	left := versionParts(candidate)
	right := versionParts(current)
	for index := range left {
		if left[index] != right[index] {
			return left[index] > right[index]
		}
	}
	return false
}

func versionParts(value string) [3]int {
	parts := strings.Split(trimVersion(value), ".")
	var result [3]int
	for index := range result {
		result[index], _ = strconv.Atoi(parts[index])
	}
	return result
}

func trimVersion(value string) string {
	return strings.TrimPrefix(strings.TrimSpace(value), "v")
}

func unavailableError() *models.UserError {
	return &models.UserError{Code: "update_check_failed", Message: "Não foi possível verificar atualizações agora.", Retryable: true}
}
