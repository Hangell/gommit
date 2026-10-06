// SPDX-License-Identifier: GPL-3.0-only

package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Hangell/gommit/internal/i18n"
)

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// Replace only within sequential tests and restore even when a test fails.
// Any unexpected URL fails immediately instead of accessing the network.
func mockRelease(t *testing.T, status int, body string) *int {
	t.Helper()
	requests := new(int)
	previous := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previous })
	http.DefaultTransport = transportFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.String() != releasesAPI {
			return nil, fmt.Errorf("unexpected network request: %s", r.URL)
		}
		*requests++
		if r.Header.Get("User-Agent") != "gommit-update-check" {
			t.Error("missing release lookup user agent")
		}
		return &http.Response{StatusCode: status, Status: fmt.Sprintf("%d %s", status, http.StatusText(status)), Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body)), Request: r}, nil
	})
	return requests
}

func TestFetchRelease(t *testing.T) {
	for _, tt := range []struct {
		name, body, errorText string
		status                int
	}{
		{"valid", `{"tag_name":"v1.2.3","assets":[{"name":"archive","browser_download_url":"https://example.invalid/archive"}]}`, "", http.StatusOK},
		{"missing tag", `{"assets":[]}`, "no tag", http.StatusOK},
		{"invalid JSON", `{`, "unexpected EOF", http.StatusOK},
		{"HTTP failure", `{}`, "GitHub returned", http.StatusForbidden},
	} {
		t.Run(tt.name, func(t *testing.T) {
			mockRelease(t, tt.status, tt.body)
			r, err := fetchRelease(time.Second)
			if tt.errorText != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errorText) {
					t.Fatalf("fetchRelease() = %v; want %q", err, tt.errorText)
				}
				return
			}
			if err != nil || r.TagName != "v1.2.3" || len(r.Assets) != 1 || r.Assets[0].Name != "archive" {
				t.Fatalf("release = %+v, %v", r, err)
			}
		})
	}
}

func isolatedCache(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)
	t.Setenv("LOCALAPPDATA", dir)
	t.Setenv("HOME", dir)
	path, err := cachePath()
	if err != nil {
		t.Fatal(err)
	}
	if rel, err := filepath.Rel(dir, path); err != nil || strings.HasPrefix(rel, "..") {
		t.Fatalf("cache was not isolated: %s", path)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestCachedLatest(t *testing.T) {
	for _, tt := range []struct {
		name, data string
		wantFetch  bool
	}{
		{"fresh", cacheJSON(t, time.Now(), "1.2.3"), false},
		{"expired", cacheJSON(t, time.Now().Add(-2*checkTTL), "1.0.0"), true},
		{"corrupt", `{`, true},
		{"missing", "", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			path := isolatedCache(t)
			if tt.data != "" {
				if err := os.WriteFile(path, []byte(tt.data), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			requests := mockRelease(t, http.StatusOK, `{"tag_name":"v1.2.3"}`)
			if got, err := cachedLatest(); err != nil || got != "1.2.3" {
				t.Fatalf("cachedLatest() = %q, %v", got, err)
			}
			if (*requests == 1) != tt.wantFetch {
				t.Fatalf("network requests = %d, wantFetch = %v", *requests, tt.wantFetch)
			}
			if _, err := cachedLatest(); err != nil || (*requests == 1) != tt.wantFetch {
				t.Fatalf("second lookup failed or refetched: %v, requests=%d", err, *requests)
			}
		})
	}
}

func cacheJSON(t *testing.T, checked time.Time, latest string) string {
	t.Helper()
	data, err := json.Marshal(cache{CheckedAt: checked, Latest: latest})
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestNotice(t *testing.T) {
	previous := i18n.Language()
	t.Cleanup(func() { i18n.Set(previous) })
	i18n.Set("en")
	isolatedCache(t)
	requests := mockRelease(t, http.StatusOK, `{"tag_name":"v1.2.3"}`)
	if got := Notice("dev"); got != "" || *requests != 0 {
		t.Fatalf("development build fetched or displayed notice: %q", got)
	}
	if got := Notice("1.0.0"); got == "" || !strings.Contains(got, "1.2.3") {
		t.Fatalf("missing upgrade notice: %q", got)
	}
	if got := Notice("1.2.3"); got != "" {
		t.Fatalf("up-to-date version displayed notice: %q", got)
	}
}

func TestNoticeLookupFailure(t *testing.T) {
	isolatedCache(t)
	mockRelease(t, http.StatusServiceUnavailable, "unavailable")
	if got := Notice("1.0.0"); got != "" {
		t.Fatalf("lookup failure must not disrupt commit: %q", got)
	}
}

func TestStartWithoutDownload(t *testing.T) {
	previous := i18n.Language()
	t.Cleanup(func() { i18n.Set(previous) })
	i18n.Set("en")
	mockRelease(t, http.StatusOK, `{"tag_name":"v1.2.3"}`)
	if got, err := Start("1.2.3"); err != nil || !strings.Contains(got, "1.2.3") {
		t.Fatalf("up-to-date Start() = %q, %v", got, err)
	}
	if _, err := Start("1.0.0"); err == nil || !strings.Contains(err.Error(), "no package") {
		t.Fatalf("missing platform asset error = %v", err)
	}
}
