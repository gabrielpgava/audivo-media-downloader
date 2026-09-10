package download

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTempWorkspacePreservesResumeMarkersUntilSuccess(t *testing.T) {
	root := t.TempDir()
	workspace, err := NewTempWorkspace(root, "job-1")
	if err != nil {
		t.Fatal(err)
	}
	partial := filepath.Join(workspace, "track.part")
	discard := filepath.Join(workspace, "debug.json")
	if err := os.WriteFile(partial, []byte("partial"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(discard, []byte("details"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := FinalizeTempWorkspace(workspace, TempCancelled); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(partial); err != nil {
		t.Fatalf("resume marker was removed: %v", err)
	}
	if _, err := os.Stat(discard); !os.IsNotExist(err) {
		t.Fatalf("unsafe temporary still exists: %v", err)
	}
	if err := FinalizeTempWorkspace(workspace, TempSucceeded); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(workspace); !os.IsNotExist(err) {
		t.Fatalf("completed workspace still exists: %v", err)
	}
}

func TestTempWorkspaceRejectsUnownedPath(t *testing.T) {
	if err := FinalizeTempWorkspace(filepath.Join(t.TempDir(), "not-a-job"), TempSucceeded); err == nil {
		t.Fatal("expected arbitrary workspace to be rejected")
	}
}
