package platform

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const applicationName = "Audivo"

type Paths struct {
	ConfigDir       string
	CacheDir        string
	EngineDir       string
	TempDir         string
	CookieDir       string
	AuthProfileDir  string
	DefaultDownload string
}

func ResolvePaths() (Paths, error) {
	configRoot, err := os.UserConfigDir()
	if err != nil {
		return Paths{}, err
	}
	cacheRoot, err := os.UserCacheDir()
	if err != nil {
		return Paths{}, err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return Paths{}, err
	}

	downloads := resolveDownloads(home)
	paths := Paths{
		ConfigDir:       filepath.Join(configRoot, applicationName),
		CacheDir:        filepath.Join(cacheRoot, applicationName),
		EngineDir:       filepath.Join(cacheRoot, applicationName, "engines", runtime.GOOS+"-"+runtime.GOARCH),
		TempDir:         filepath.Join(cacheRoot, applicationName, "tmp"),
		CookieDir:       filepath.Join(configRoot, applicationName, "cookies"),
		AuthProfileDir:  filepath.Join(configRoot, applicationName, "auth-profiles"),
		DefaultDownload: downloads,
	}
	for _, path := range []string{paths.ConfigDir, paths.CacheDir, paths.EngineDir, paths.TempDir, paths.CookieDir, paths.AuthProfileDir} {
		if !filepath.IsAbs(path) {
			return Paths{}, errors.New("resolved application path is not absolute")
		}
	}
	return paths, nil
}

func resolveDownloads(home string) string {
	if runtime.GOOS == "linux" {
		if configured := linuxUserDir(home, "XDG_DOWNLOAD_DIR"); configured != "" {
			return configured
		}
	}
	return filepath.Join(home, "Downloads")
}

func linuxUserDir(home, key string) string {
	configPath := filepath.Join(home, ".config", "user-dirs.dirs")
	file, err := os.Open(configPath)
	if err != nil {
		return ""
	}
	defer file.Close()

	prefix := key + "="
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		value := strings.Trim(strings.TrimPrefix(line, prefix), `"`)
		value = strings.ReplaceAll(value, "$HOME", home)
		if filepath.IsAbs(value) {
			return filepath.Clean(value)
		}
	}
	return ""
}

func EnsureDirectory(path string) error {
	if path == "" || !filepath.IsAbs(path) {
		return errors.New("directory must be an absolute path")
	}
	return os.MkdirAll(path, 0o700)
}

func IsWritableDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return false
	}
	testFile, err := os.CreateTemp(path, ".audivo-write-test-*")
	if err != nil {
		return false
	}
	testPath := testFile.Name()
	_ = testFile.Close()
	_ = os.Remove(testPath)
	return true
}
