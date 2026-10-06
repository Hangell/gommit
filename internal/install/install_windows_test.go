//go:build windows

// SPDX-License-Identifier: GPL-3.0-only

package install

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWindowsInstallPaths(t *testing.T) {
	home := t.TempDir()
	local := t.TempDir()
	t.Setenv("USERPROFILE", home)
	t.Setenv("LOCALAPPDATA", local)
	assertTarget := func(want string) {
		t.Helper()
		if got, err := targetBinDir(); err != nil || got != want {
			t.Fatalf("targetBinDir() = %q, %v; want %q", got, err, want)
		}
	}
	assertTarget(filepath.Join(local, "Programs", "gommit", "bin"))
	t.Setenv("LOCALAPPDATA", "")
	assertTarget(filepath.Join(home, "AppData", "Local", "Programs", "gommit", "bin"))
	goBin := filepath.Join(home, "go", "bin")
	if err := os.MkdirAll(goBin, 0o755); err != nil {
		t.Fatal(err)
	}
	assertTarget(goBin)
}
