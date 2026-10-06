//go:build !windows

// SPDX-License-Identifier: GPL-3.0-only

package install

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureLineInFile(t *testing.T) {
	for _, initial := range []string{"", "# existing settings\nexport EDITOR=vim\n"} {
		t.Run(initial, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "profile")
			if initial != "" {
				if err := os.WriteFile(path, []byte(initial), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			line := `export PATH="/example/bin:$PATH"`
			for i := 0; i < 2; i++ {
				if err := ensureLineInFile(path, line); err != nil {
					t.Fatal(err)
				}
			}
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.HasPrefix(string(data), initial) || strings.Count(string(data), line) != 1 {
				t.Fatalf("profile was overwritten or duplicated: %q", data)
			}
		})
	}
	dir := t.TempDir()
	if err := ensureLineInFile(dir, "line"); err == nil {
		t.Fatal("expected profile directory error")
	}
	if err := ensureLineInFile(filepath.Join(dir, "missing", "profile"), "line"); err == nil {
		t.Fatal("expected missing parent error")
	}
}

func TestUnixInstallPaths(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	want := filepath.Join(home, ".local", "bin")
	if got, err := targetBinDir(); err != nil || got != want {
		t.Fatalf("targetBinDir() = %q, %v; want %q", got, err, want)
	}
	for i := 0; i < 2; i++ {
		if err := addDirToPath(want); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{".zshrc", ".bashrc", ".profile"} {
		data, err := os.ReadFile(filepath.Join(home, name))
		if err != nil || strings.Count(string(data), want) != 1 {
			t.Fatalf("%s = %q, %v; expected one PATH entry", name, data, err)
		}
	}
}
