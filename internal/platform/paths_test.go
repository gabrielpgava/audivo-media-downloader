package platform

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolvePathsReturnsAbsoluteApplicationPaths(t *testing.T) {
	paths, err := ResolvePaths()
	if err != nil {
		t.Fatal(err)
	}
	for name, path := range map[string]string{
		"config":    paths.ConfigDir,
		"cache":     paths.CacheDir,
		"engines":   paths.EngineDir,
		"temp":      paths.TempDir,
		"cookies":   paths.CookieDir,
		"profiles":  paths.AuthProfileDir,
		"downloads": paths.DefaultDownload,
	} {
		if !filepath.IsAbs(path) {
			t.Errorf("%s path %q is not absolute", name, path)
		}
	}
}

func TestEnsureDirectoryAndWritable(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested")
	if err := EnsureDirectory(dir); err != nil {
		t.Fatal(err)
	}
	if !IsWritableDirectory(dir) {
		t.Fatal("expected directory to be writable")
	}
	if err := EnsureDirectory("relative"); err == nil {
		t.Fatal("expected relative directory to fail")
	}
	_ = os.RemoveAll(dir)
}
