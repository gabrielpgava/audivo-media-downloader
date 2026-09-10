package main

import "testing"

func TestIsOfficialReleaseURL(t *testing.T) {
	valid := "https://github.com/gabrielpgava/audivo-media-downloader/releases/tag/v1.2.3"
	if !isOfficialReleaseURL(valid) {
		t.Fatal("expected official release URL")
	}
	for _, candidate := range []string{
		"http://github.com/gabrielpgava/audivo-media-downloader/releases/tag/v1.2.3",
		"https://github.com.evil.example/gabrielpgava/audivo-media-downloader/releases/tag/v1.2.3",
		"https://github.com@evil.example/gabrielpgava/audivo-media-downloader/releases/tag/v1.2.3",
		"https://github.com/gabrielpgava/other/releases/tag/v1.2.3",
	} {
		if isOfficialReleaseURL(candidate) {
			t.Fatalf("unexpectedly accepted release URL %q", candidate)
		}
	}
}
