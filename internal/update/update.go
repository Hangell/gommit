package update

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Hangell/gommit/internal/i18n"
	"github.com/Hangell/gommit/internal/install"
)

const (
	releasesAPI = "https://api.github.com/repos/Hangell/gommit/releases/latest"
	checkTTL    = 24 * time.Hour
)

type release struct {
	TagName string  `json:"tag_name"`
	Assets  []asset `json:"assets"`
}

type asset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

type cache struct {
	CheckedAt time.Time `json:"checked_at"`
	Latest    string    `json:"latest"`
}

// Notice returns a non-blocking-friendly update message. Network and cache
// failures are intentionally silent because they must never fail a commit.
func Notice(current string) string {
	if !isReleaseVersion(current) {
		return ""
	}
	latest, err := cachedLatest()
	if err != nil || !newer(latest, current) {
		return ""
	}
	return i18n.T("update.notice", latest, current)
}

func cachedLatest() (string, error) {
	path, err := cachePath()
	if err == nil {
		var c cache
		if data, readErr := os.ReadFile(path); readErr == nil && json.Unmarshal(data, &c) == nil && time.Since(c.CheckedAt) < checkTTL {
			return c.Latest, nil
		}
	}

	r, err := fetchRelease(3 * time.Second)
	if err != nil {
		return "", err
	}
	latest := normalizeVersion(r.TagName)
	if path != "" {
		_ = os.MkdirAll(filepath.Dir(path), 0o755)
		if data, marshalErr := json.Marshal(cache{CheckedAt: time.Now(), Latest: latest}); marshalErr == nil {
			_ = os.WriteFile(path, data, 0o644)
		}
	}
	return latest, nil
}

// Start downloads the latest release and starts a helper that installs it
// after this process exits (required on Windows, harmless elsewhere).
func Start(current string) (string, error) {
	r, err := fetchRelease(30 * time.Second)
	if err != nil {
		return "", err
	}
	latest := normalizeVersion(r.TagName)
	if isReleaseVersion(current) && !newer(latest, current) {
		return i18n.T("update.current", current), nil
	}

	wanted := "_" + runtime.GOOS + "_" + runtime.GOARCH
	ext := ".tar.gz"
	if runtime.GOOS == "windows" {
		ext = ".zip"
	}
	var selected asset
	var checksums asset
	for _, a := range r.Assets {
		if strings.Contains(a.Name, wanted) && strings.HasSuffix(a.Name, ext) {
			selected = a
		}
		if a.Name == "SHA256SUMS.txt" {
			checksums = a
		}
	}
	if selected.URL == "" {
		return "", fmt.Errorf("release %s has no package for %s/%s", latest, runtime.GOOS, runtime.GOARCH)
	}

	dir, err := os.MkdirTemp("", "gommit-update-*")
	if err != nil {
		return "", err
	}
	archive := filepath.Join(dir, selected.Name)
	if err := download(selected.URL, archive); err != nil {
		_ = os.RemoveAll(dir)
		return "", err
	}
	if checksums.URL == "" {
		_ = os.RemoveAll(dir)
		return "", errors.New("release has no SHA256SUMS.txt")
	}
	if err := verifyChecksum(checksums.URL, selected.Name, archive); err != nil {
		_ = os.RemoveAll(dir)
		return "", err
	}
	bin, err := extractBinary(archive, dir)
	if err != nil {
		_ = os.RemoveAll(dir)
		return "", err
	}
	_ = os.Chmod(bin, 0o755)

	cmd := exec.Command(bin, "--apply-update", strconv.Itoa(os.Getpid()), latest, dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = nil
	if err := cmd.Start(); err != nil {
		_ = os.RemoveAll(dir)
		return "", err
	}
	return i18n.T("update.downloaded", latest), nil
}

func verifyChecksum(url, assetName, archive string) error {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "gommit-updater")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("checksum download returned %s", resp.Status)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}
	var expected string
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && strings.TrimPrefix(fields[len(fields)-1], "*") == assetName {
			expected = strings.ToLower(fields[0])
			break
		}
	}
	if expected == "" {
		return fmt.Errorf("checksum for %s not found", assetName)
	}
	f, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	actual := fmt.Sprintf("%x", h.Sum(nil))
	if actual != expected {
		return fmt.Errorf("checksum mismatch for %s", assetName)
	}
	return nil
}

func Apply(parentPID int, version, tempDir string) error {
	// Give the parent enough time to exit and release the installed .exe on Windows.
	// The PID remains part of the internal protocol for future platform-specific waits.
	_ = parentPID
	time.Sleep(1500 * time.Millisecond)
	res, err := install.InstallSelf(version)
	if err != nil {
		return err
	}
	fmt.Println(res.Message)
	_ = os.RemoveAll(tempDir)
	return nil
}

func fetchRelease(timeout time.Duration) (release, error) {
	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequest(http.MethodGet, releasesAPI, nil)
	if err != nil {
		return release{}, err
	}
	req.Header.Set("User-Agent", "gommit-update-check")
	resp, err := client.Do(req)
	if err != nil {
		return release{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return release{}, fmt.Errorf("GitHub returned %s", resp.Status)
	}
	var r release
	if err := json.NewDecoder(io.LimitReader(resp.Body, 2<<20)).Decode(&r); err != nil {
		return release{}, err
	}
	if r.TagName == "" {
		return release{}, errors.New("latest release has no tag")
	}
	return r, nil
}

func download(url, dst string) error {
	client := &http.Client{Timeout: 2 * time.Minute}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "gommit-updater")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned %s", resp.Status)
	}
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(f, io.LimitReader(resp.Body, 100<<20))
	closeErr := f.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func extractBinary(archive, dir string) (string, error) {
	name := "gommit"
	if runtime.GOOS == "windows" {
		name += ".exe"
		z, err := zip.OpenReader(archive)
		if err != nil {
			return "", err
		}
		defer z.Close()
		for _, f := range z.File {
			if filepath.Base(f.Name) != name || f.FileInfo().IsDir() {
				continue
			}
			r, err := f.Open()
			if err != nil {
				return "", err
			}
			return writeExtracted(filepath.Join(dir, "new-"+name), r)
		}
		return "", errors.New("gommit.exe not found in release package")
	}

	f, err := os.Open(archive)
	if err != nil {
		return "", err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", err
		}
		if filepath.Base(h.Name) == name && h.Typeflag == tar.TypeReg {
			return writeExtracted(filepath.Join(dir, "new-"+name), io.NopCloser(tr))
		}
	}
	return "", errors.New("gommit not found in release package")
}

func writeExtracted(dst string, src io.ReadCloser) (string, error) {
	defer src.Close()
	f, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	_, copyErr := io.Copy(f, src)
	closeErr := f.Close()
	if copyErr != nil {
		return "", copyErr
	}
	if closeErr != nil {
		return "", closeErr
	}
	return dst, nil
}

func cachePath() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "gommit", "update-check.json"), nil
}

func normalizeVersion(v string) string { return strings.TrimPrefix(strings.TrimSpace(v), "v") }

func isReleaseVersion(v string) bool {
	parts := strings.Split(normalizeVersion(v), ".")
	if len(parts) < 2 {
		return false
	}
	for _, p := range parts {
		if _, err := strconv.Atoi(p); err != nil {
			return false
		}
	}
	return true
}

func newer(latest, current string) bool {
	a, b := strings.Split(normalizeVersion(latest), "."), strings.Split(normalizeVersion(current), ".")
	for len(a) < 3 {
		a = append(a, "0")
	}
	for len(b) < 3 {
		b = append(b, "0")
	}
	for i := 0; i < 3; i++ {
		ai, errA := strconv.Atoi(a[i])
		bi, errB := strconv.Atoi(b[i])
		if errA != nil || errB != nil {
			return false
		}
		if ai != bi {
			return ai > bi
		}
	}
	return false
}
