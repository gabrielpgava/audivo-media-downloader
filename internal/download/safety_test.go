package download

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSanitizeComponentCoversCrossPlatformNames(t *testing.T) {
	for _, input := range []string{"AC/DC - Thunderstruck", "What's Up?", "Artist: Song", "<Official Video>", "CON"} {
		got := SanitizeComponent(input)
		if got == "" || got == input && input != "CON" {
			t.Fatalf("SanitizeComponent(%q) = %q", input, got)
		}
	}
	if got := SanitizeComponent("CON"); got == "CON" {
		t.Fatal("reserved Windows name was not changed")
	}
	if got := SanitizeComponent("Álbum 日本語 " + strings.Repeat("x", 200)); len([]rune(got)) > maxPathComponentRunes {
		t.Fatalf("long Unicode component was not bounded: %d", len([]rune(got)))
	}
}

func TestUniquePathDoesNotOverwrite(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "Musica.mp3")
	if err := os.WriteFile(existing, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := UniquePath(dir, "Musica.mp3")
	if err != nil {
		t.Fatal(err)
	}
	if got == existing || filepath.Base(got) != "Musica (2).mp3" {
		t.Fatalf("unexpected collision path: %q", got)
	}
}
