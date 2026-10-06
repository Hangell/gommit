// SPDX-License-Identifier: GPL-3.0-only

package update

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestVerifyChecksum(t *testing.T) {
	payload := []byte("release binary payload")
	archive := filepath.Join(t.TempDir(), "release.tar.gz")
	if err := os.WriteFile(archive, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	digest := fmt.Sprintf("%x", sha256.Sum256(payload))
	for _, tt := range []struct {
		name, manifest, errorText string
		status                    int
	}{
		{"valid", digest + "  release.tar.gz\n", "", http.StatusOK},
		{"binary checksum format", strings.ToUpper(digest) + " *release.tar.gz\r\n", "", http.StatusOK},
		{"mismatch", strings.Repeat("0", 64) + "  release.tar.gz\n", "checksum mismatch", http.StatusOK},
		{"missing filename", digest + "  other.tar.gz\n", "not found", http.StatusOK},
		{"empty", "", "not found", http.StatusOK},
		{"HTTP error", "unavailable", "checksum download returned", http.StatusServiceUnavailable},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("User-Agent") != "gommit-updater" {
					t.Error("missing updater user agent")
				}
				w.WriteHeader(tt.status)
				fmt.Fprint(w, tt.manifest)
			}))
			defer server.Close()
			err := verifyChecksum(server.URL, "release.tar.gz", archive)
			if tt.errorText == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tt.errorText) {
				t.Fatalf("verifyChecksum() = %v; want %q", err, tt.errorText)
			}
		})
	}
}

func TestDownload(t *testing.T) {
	for _, status := range []int{http.StatusOK, http.StatusNotFound} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("User-Agent") != "gommit-updater" {
					t.Error("missing updater user agent")
				}
				w.WriteHeader(status)
				fmt.Fprint(w, "archive contents")
			}))
			defer server.Close()
			dst := filepath.Join(t.TempDir(), "archive")
			err := download(server.URL, dst)
			if status != http.StatusOK {
				if err == nil {
					t.Fatal("expected HTTP failure")
				}
				if _, err := os.Stat(dst); !os.IsNotExist(err) {
					t.Fatalf("failed HTTP request created a file: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if data, err := os.ReadFile(dst); err != nil || string(data) != "archive contents" {
				t.Fatalf("downloaded data = %q, %v", data, err)
			}
			if err := download(server.URL, filepath.Join(dst, "invalid")); err == nil {
				t.Fatal("expected destination failure")
			}
		})
	}
}

func releaseArchive(t *testing.T, dir string, includeBinary bool) string {
	t.Helper()
	name := "gommit"
	ext := ".tar.gz"
	if runtime.GOOS == "windows" {
		name += ".exe"
		ext = ".zip"
	}
	path := filepath.Join(dir, "release"+ext)
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	entries := map[string]string{"package/README.md": "documentation"}
	if includeBinary {
		entries["package/"+name] = "binary contents"
	}
	if runtime.GOOS == "windows" {
		w := zip.NewWriter(f)
		for name, contents := range entries {
			entry, err := w.Create(name)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := io.WriteString(entry, contents); err != nil {
				t.Fatal(err)
			}
		}
		if err := w.Close(); err != nil {
			t.Fatal(err)
		}
	} else {
		gz := gzip.NewWriter(f)
		w := tar.NewWriter(gz)
		for name, contents := range entries {
			if err := w.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(contents)), Typeflag: tar.TypeReg}); err != nil {
				t.Fatal(err)
			}
			if _, err := io.WriteString(w, contents); err != nil {
				t.Fatal(err)
			}
		}
		if err := w.Close(); err != nil {
			t.Fatal(err)
		}
		if err := gz.Close(); err != nil {
			t.Fatal(err)
		}
	}
	return path
}

func TestExtractBinary(t *testing.T) {
	for _, include := range []bool{true, false} {
		t.Run(fmt.Sprint(include), func(t *testing.T) {
			dir := t.TempDir()
			archive := releaseArchive(t, dir, include)
			path, err := extractBinary(archive, dir)
			if !include {
				if err == nil || !strings.Contains(err.Error(), "not found") {
					t.Fatalf("missing binary error = %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if filepath.Dir(path) != dir {
				t.Fatalf("extracted outside destination: %s", path)
			}
			if data, err := os.ReadFile(path); err != nil || string(data) != "binary contents" {
				t.Fatalf("extracted data = %q, %v", data, err)
			}
			if _, err := os.Stat(filepath.Join(dir, "package", "README.md")); !os.IsNotExist(err) {
				t.Fatalf("unexpected extraction of documentation: %v", err)
			}
		})
	}
	dir := t.TempDir()
	bad := filepath.Join(dir, "corrupt")
	if err := os.WriteFile(bad, []byte("not an archive"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := extractBinary(bad, dir); err == nil {
		t.Fatal("accepted corrupt archive")
	}
	if _, err := extractBinary(filepath.Join(dir, "missing"), dir); err == nil {
		t.Fatal("accepted missing archive")
	}
}
