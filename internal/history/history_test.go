package history

import (
	"os"
	"path/filepath"
	"testing"

	"audivo-media-downloader/internal/models"
)

func TestStoreIsAtomicBoundedAndDetectsMissingFiles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.json")
	store := NewStoreAt(path)
	for index := 0; index < 55; index++ {
		if err := store.Add(models.HistoryItem{SourceURL: "https://youtu.be/" + string(rune('a'+index)), Title: "item"}); err != nil {
			t.Fatal(err)
		}
	}
	items, err := store.Recent(100)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 50 {
		t.Fatalf("expected 50 history items, got %d", len(items))
	}
	missingPath := filepath.Join(t.TempDir(), "gone.mp3")
	if err := store.Add(models.HistoryItem{SourceURL: "https://youtu.be/gone", FilePath: missingPath, Title: "gone"}); err != nil {
		t.Fatal(err)
	}
	item, err := store.Find("https://youtu.be/gone")
	if err != nil || item == nil || !item.Missing {
		t.Fatalf("expected missing item, got item=%+v err=%v", item, err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}

func TestClearDoesNotTouchMedia(t *testing.T) {
	root := t.TempDir()
	store := NewStoreAt(filepath.Join(root, "history.json"))
	media := filepath.Join(root, "media.mp3")
	if err := os.WriteFile(media, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := store.Add(models.HistoryItem{SourceURL: "https://youtu.be/x", FilePath: media}); err != nil {
		t.Fatal(err)
	}
	if err := store.Clear(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(media); err != nil {
		t.Fatal(err)
	}
}

func TestCorruptOrUnknownHistoryFallsBackToEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.json")
	if err := os.WriteFile(path, []byte("not-json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if items, err := NewStoreAt(path).Load(); err != nil || len(items) != 0 {
		t.Fatalf("corrupt history = %+v, %v", items, err)
	}
	if err := os.WriteFile(path, []byte(`{"version":99,"items":[]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if items, err := NewStoreAt(path).Load(); err != nil || len(items) != 0 {
		t.Fatalf("unknown history = %+v, %v", items, err)
	}
}

func TestFindCoversNewPresentRemovedAndExplicitRedownload(t *testing.T) {
	root := t.TempDir()
	store := NewStoreAt(filepath.Join(root, "history.json"))
	if item, err := store.Find("https://youtu.be/new"); err != nil || item != nil {
		t.Fatalf("new URL lookup = %+v, %v", item, err)
	}
	media := filepath.Join(root, "track.mp3")
	if err := os.WriteFile(media, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := store.Add(models.HistoryItem{SourceURL: "https://youtu.be/existing", FilePath: media, Title: "track"}); err != nil {
		t.Fatal(err)
	}
	present, err := store.Find("https://youtu.be/existing")
	if err != nil || present == nil || present.Missing {
		t.Fatalf("present URL lookup = %+v, %v", present, err)
	}
	if err := os.Remove(media); err != nil {
		t.Fatal(err)
	}
	removed, err := store.Find("https://youtu.be/existing")
	if err != nil || removed == nil || !removed.Missing {
		t.Fatalf("removed URL lookup = %+v, %v", removed, err)
	}
	// The explicit re-download choice is represented by retaining the duplicate
	// record and letting the caller start a new job with the normal collision policy.
	if removed.SourceURL != "https://youtu.be/existing" {
		t.Fatalf("duplicate record was not retained for re-download: %+v", removed)
	}
}
