package logstore

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStoreRotatesAndRedacts(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audivo.log")
	store := New(path, 40, 2)
	if err := store.Append("stderr", "token=secret-value --cookies-path /tmp/cookies.txt"); err != nil {
		t.Fatal(err)
	}
	if err := store.Append("stderr", strings.Repeat("x", 50)); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path + ".1")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "secret-value") || strings.Contains(string(data), "cookies.txt") {
		t.Fatalf("log leaked secret: %s", data)
	}
}
