package auth

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"audivo-media-downloader/internal/models"
	"audivo-media-downloader/internal/platform"
)

const netscapeHeader = "# Netscape HTTP Cookie File\n"

// Cookie is the small, provider-neutral representation exchanged between a
// browser adapter and the local Netscape cookie store.
type Cookie struct {
	Domain            string
	IncludeSubdomains bool
	Path              string
	Secure            bool
	HTTPOnly          bool
	Expires           int64
	Name              string
	Value             string
}

type CookieStore struct {
	paths platform.Paths
	mu    sync.Mutex
}

func NewCookieStore(paths platform.Paths) *CookieStore {
	return &CookieStore{paths: paths}
}

func (s *CookieStore) CookiePath(provider models.Service) string {
	if s == nil || !filepath.IsAbs(s.paths.CookieDir) {
		return ""
	}
	slug := providerSlug(provider)
	if slug == "" {
		return ""
	}
	return filepath.Join(s.paths.CookieDir, slug+".cookies.txt")
}

func (s *CookieStore) ProfilePath(provider models.Service) string {
	if s == nil || !filepath.IsAbs(s.paths.AuthProfileDir) {
		return ""
	}
	slug := providerSlug(provider)
	if slug == "" {
		return ""
	}
	return filepath.Join(s.paths.AuthProfileDir, slug)
}

func (s *CookieStore) Validate(provider models.Service, path string) error {
	if path == "" || !filepath.IsAbs(path) {
		return errors.New("cookie path must be absolute")
	}
	cookies, err := readNetscapeFile(path)
	if err != nil {
		return err
	}
	return ValidateCookies(provider, cookies)
}

func (s *CookieStore) Import(provider models.Service, source string) (string, error) {
	if err := s.Validate(provider, source); err != nil {
		return "", err
	}
	cookies, err := readNetscapeFile(source)
	if err != nil {
		return "", err
	}
	return s.Save(provider, cookies)
}

func (s *CookieStore) Save(provider models.Service, cookies []Cookie) (string, error) {
	if err := ValidateCookies(provider, cookies); err != nil {
		return "", err
	}
	target := s.CookiePath(provider)
	if target == "" {
		return "", errors.New("unsupported cookie provider")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := platform.EnsureDirectory(filepath.Dir(target)); err != nil {
		return "", err
	}
	if err := os.Chmod(filepath.Dir(target), 0o700); err != nil {
		return "", err
	}
	temp, err := os.CreateTemp(filepath.Dir(target), ".cookies-*.tmp")
	if err != nil {
		return "", err
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if err := temp.Chmod(0o600); err != nil {
		_ = temp.Close()
		return "", err
	}
	if _, err := temp.Write(serializeNetscape(cookies)); err != nil {
		_ = temp.Close()
		return "", err
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return "", err
	}
	if err := temp.Close(); err != nil {
		return "", err
	}
	if err := os.Rename(tempPath, target); err != nil {
		return "", err
	}
	return target, nil
}

func (s *CookieStore) Remove(provider models.Service) error {
	target := s.CookiePath(provider)
	if target == "" {
		return errors.New("unsupported cookie provider")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := os.Remove(target); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func (s *CookieStore) RemoveProfile(provider models.Service) error {
	profile := s.ProfilePath(provider)
	if profile == "" {
		return errors.New("unsupported cookie provider")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return os.RemoveAll(profile)
}

func (s *CookieStore) IsConnected(provider models.Service, configuredPath string) bool {
	return configuredPath != "" && s.Validate(provider, configuredPath) == nil
}

func ValidateCookies(provider models.Service, cookies []Cookie) error {
	if !supportedProvider(provider) {
		return errors.New("unsupported cookie provider")
	}
	if len(cookies) == 0 {
		return errors.New("cookie file is empty or not in Netscape format")
	}
	allowed := false
	for _, cookie := range cookies {
		if err := validateCookieRecord(cookie); err != nil {
			return err
		}
		if allowedCookieDomain(provider, cookie.Domain) {
			allowed = true
		}
	}
	if !allowed {
		return fmt.Errorf("cookie file has no %s provider domain", providerLabel(provider))
	}
	return nil
}

func readNetscapeFile(path string) ([]Cookie, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var cookies []Cookie
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSuffix(scanner.Text(), "\r")
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "#HttpOnly_") {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) != 7 {
			return nil, errors.New("cookies file is not in Netscape format")
		}
		domain := parts[0]
		httpOnly := false
		if strings.HasPrefix(domain, "#HttpOnly_") {
			httpOnly = true
			domain = strings.TrimPrefix(domain, "#HttpOnly_")
		}
		expires, err := strconv.ParseInt(parts[4], 10, 64)
		if err != nil || expires < 0 {
			expires = 0
		}
		cookies = append(cookies, Cookie{
			Domain:            domain,
			IncludeSubdomains: strings.EqualFold(parts[1], "TRUE"),
			Path:              parts[2],
			Secure:            strings.EqualFold(parts[3], "TRUE"),
			HTTPOnly:          httpOnly,
			Expires:           expires,
			Name:              parts[5],
			Value:             parts[6],
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(cookies) == 0 {
		return nil, errors.New("cookie file is empty or not in Netscape format")
	}
	return cookies, nil
}

func serializeNetscape(cookies []Cookie) []byte {
	var output bytes.Buffer
	output.WriteString(netscapeHeader)
	for _, cookie := range cookies {
		domain := cookie.Domain
		if cookie.HTTPOnly {
			domain = "#HttpOnly_" + domain
		}
		expires := cookie.Expires
		if expires < 0 {
			expires = 0
		}
		fmt.Fprintf(&output, "%s\t%s\t%s\t%s\t%d\t%s\t%s\n",
			domain, boolToken(cookie.IncludeSubdomains), cookie.Path, boolToken(cookie.Secure), expires, cookie.Name, cookie.Value)
	}
	return output.Bytes()
}

func validateCookieRecord(cookie Cookie) error {
	if strings.TrimSpace(cookie.Domain) == "" || cookie.Name == "" {
		return errors.New("cookies file is not in Netscape format")
	}
	if cookie.Path == "" || !strings.HasPrefix(cookie.Path, "/") {
		return errors.New("cookies file is not in Netscape format")
	}
	for _, field := range []string{cookie.Domain, cookie.Path, cookie.Name, cookie.Value} {
		if strings.ContainsAny(field, "\r\n\t") {
			return errors.New("cookies file is not in Netscape format")
		}
	}
	return nil
}

func allowedCookieDomain(provider models.Service, domain string) bool {
	domain = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(domain)), ".")
	switch provider {
	case models.ServiceAppleMusic:
		return domain == "apple.com" || strings.HasSuffix(domain, ".apple.com")
	case models.ServiceYouTube:
		return domain == "youtube.com" || strings.HasSuffix(domain, ".youtube.com") || domain == "google.com" || strings.HasSuffix(domain, ".google.com")
	default:
		return false
	}
}

func supportedProvider(provider models.Service) bool {
	return provider == models.ServiceAppleMusic || provider == models.ServiceYouTube
}

func providerSlug(provider models.Service) string {
	switch provider {
	case models.ServiceAppleMusic:
		return "apple-music"
	case models.ServiceYouTube:
		return "youtube"
	default:
		return ""
	}
}

func providerLabel(provider models.Service) string {
	if provider == models.ServiceAppleMusic {
		return "Apple Music"
	}
	return "YouTube"
}

func boolToken(value bool) string {
	if value {
		return "TRUE"
	}
	return "FALSE"
}
