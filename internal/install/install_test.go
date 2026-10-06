// SPDX-License-Identifier: GPL-3.0-only

package install

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCopyFile(t *testing.T) {
	dir := t.TempDir()
	src, dst := filepath.Join(dir, "source"), filepath.Join(dir, "destination")
	payload := []byte("binary\x00payload\n")
	if err := os.WriteFile(src, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, []byte("long previous binary contents"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := copyFile(src, dst); err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(dst); err != nil || string(got) != string(payload) {
		t.Fatalf("copied data = %q, %v", got, err)
	}
	if err := copyFile(filepath.Join(dir, "missing"), dst); err == nil {
		t.Fatal("expected missing source error")
	}
	if err := copyFile(src, dir); err == nil {
		t.Fatal("expected destination directory error")
	}
	if err := copyFile(src, filepath.Join(dir, "missing", "binary")); err == nil {
		t.Fatal("expected missing destination parent error")
	}
}

func TestDirOnPath(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "bin")
	other := filepath.Join(t.TempDir(), "other")
	t.Setenv("PATH", strings.Join([]string{other, bin}, string(os.PathListSeparator)))
	if got, err := dirOnPath(bin); err != nil || !got {
		t.Fatalf("existing directory = %v, %v", got, err)
	}
	if got, err := dirOnPath(bin + string(os.PathSeparator)); err != nil || !got {
		t.Fatalf("trailing separator = %v, %v", got, err)
	}
	if got, err := dirOnPath(bin + "-other"); err != nil || got {
		t.Fatalf("partial path matched = %v, %v", got, err)
	}
}
