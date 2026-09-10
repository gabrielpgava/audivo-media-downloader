package updates

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestCheckerUsesStableReleaseAndTwentyFourHourCache(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		_, _ = writer.Write([]byte(`[
      {"tag_name":"v0.2.0-rc.1","html_url":"https://github.com/example/rc","prerelease":true},
      {"tag_name":"v0.1.5","html_url":"https://github.com/example/stable","prerelease":false},
      {"tag_name":"v0.1.4","html_url":"https://github.com/example/old","prerelease":false}
    ]`))
	}))
	defer server.Close()

	checker := NewChecker(filepath.Join(t.TempDir(), "update.json"))
	checker.URL = server.URL
	checker.Now = func() time.Time { return time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC) }
	first := checker.Check(context.Background(), "0.1.0")
	second := checker.Check(context.Background(), "0.1.0")
	if !first.Available || first.LatestVersion != "0.1.5" || second.LatestVersion != "0.1.5" {
		t.Fatalf("unexpected update responses: %+v / %+v", first, second)
	}
	if requests.Load() != 1 {
		t.Fatalf("expected cached second check, got %d requests", requests.Load())
	}
}
