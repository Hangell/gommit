// SPDX-License-Identifier: GPL-3.0-only

package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func isolatedGit(t *testing.T) {
	t.Helper()
	configDir := t.TempDir()
	// Ignore the contributor's identity, signing settings, hooks, and repo overrides.
	for _, key := range []string{"GIT_DIR", "GIT_WORK_TREE", "GIT_INDEX_FILE", "GIT_COMMON_DIR", "GIT_CONFIG", "GIT_CONFIG_PARAMETERS", "GIT_AUTHOR_NAME", "GIT_AUTHOR_EMAIL", "GIT_COMMITTER_NAME", "GIT_COMMITTER_EMAIL"} {
		t.Setenv(key, "")
		if err := os.Unsetenv(key); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("GIT_CONFIG_COUNT", "0")
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(configDir, "gitconfig"))
	t.Setenv("GIT_TEMPLATE_DIR", configDir)
	t.Chdir(t.TempDir())
}

func runGit(t *testing.T, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", args...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func TestCommitWorkflow(t *testing.T) {
	isolatedGit(t)
	runGit(t, "init", "--initial-branch=main")
	runGit(t, "config", "user.name", "Gommit Test")
	runGit(t, "config", "user.email", "gommit@example.invalid")
	runGit(t, "config", "commit.gpgsign", "false")
	if !InRepo() {
		t.Fatal("temporary repository was not detected")
	}
	assertState := func(wantDirty, wantStaged bool) {
		t.Helper()
		if got, err := WorkingTreeDirty(); err != nil || got != wantDirty {
			t.Fatalf("WorkingTreeDirty() = %v, %v; want %v", got, err, wantDirty)
		}
		if got, err := HasStagedChanges(); err != nil || got != wantStaged {
			t.Fatalf("HasStagedChanges() = %v, %v; want %v", got, err, wantStaged)
		}
	}
	assertState(false, false)
	if err := os.WriteFile("example.txt", []byte("example\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	assertState(true, false)
	if err := StageAll(); err != nil {
		t.Fatal(err)
	}
	assertState(true, true)
	if summary, err := StagedSummary(); err != nil || !strings.Contains(summary, "A\texample.txt") {
		t.Fatalf("StagedSummary() = %q, %v", summary, err)
	}
	message := "feat: add example\n\nBody with Unicode: tradução 🚀"
	if err := CommitWithMessage(message, Options{Signoff: true}); err != nil {
		t.Fatal(err)
	}
	if got, err := LastCommitSubject(); err != nil || got != "feat: add example" {
		t.Fatalf("LastCommitSubject() = %q, %v", got, err)
	}
	if got, err := LastCommitMessage(); err != nil || !strings.Contains(got, message) || !strings.Contains(got, "Signed-off-by: Gommit Test <gommit@example.invalid>") {
		t.Fatalf("LastCommitMessage() = %q, %v", got, err)
	}
	assertState(false, false)
	if err := CommitWithMessage("fix: amend example", Options{Amend: true}); err != nil {
		t.Fatal(err)
	}
	if count := runGit(t, "rev-list", "--count", "HEAD"); count != "1" {
		t.Fatalf("amend created an extra commit: %s", count)
	}
	if got, err := LastCommitSubject(); err != nil || got != "fix: amend example" {
		t.Fatalf("amended subject = %q, %v", got, err)
	}
	if err := CommitWithMessage("chore: empty commit", Options{AllowEmpty: true}); err != nil {
		t.Fatal(err)
	}
	if count := runGit(t, "rev-list", "--count", "HEAD"); count != "2" {
		t.Fatalf("allow-empty did not create a commit: %s", count)
	}
}

func TestOutsideRepository(t *testing.T) {
	isolatedGit(t)
	if InRepo() {
		t.Fatal("unexpected repository")
	}
	if _, err := HasStagedChanges(); err == nil {
		t.Fatal("expected error outside repository")
	}
	if _, err := WorkingTreeDirty(); err == nil {
		t.Fatal("expected error outside repository")
	}
	if err := CommitWithMessage("fix: example", Options{}); err == nil {
		t.Fatal("expected commit error outside repository")
	}
}

func TestWriteCommitEditMsg(t *testing.T) {
	path := filepath.Join(t.TempDir(), "COMMIT_EDITMSG")
	message := "docs: update tradução\n\nCloses #42"
	if err := os.WriteFile(path, []byte("old message"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := WriteCommitEditMsg(path, message); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != message {
		t.Fatalf("editor message = %q, %v; want %q", got, err, message)
	}
	if err := WriteCommitEditMsg(filepath.Join(path, "missing"), message); err == nil {
		t.Fatal("expected error writing to invalid path")
	}
}

func TestConfiguredPreferences(t *testing.T) {
	isolatedGit(t)
	if got, err := ConfiguredMode(); err != nil || got != "" {
		t.Fatalf("unset mode = %q, %v", got, err)
	}
	if got, err := ConfiguredLanguage(); err != nil || got != "" {
		t.Fatalf("unset language = %q, %v", got, err)
	}
	for _, mode := range []string{"simple", "full"} {
		if err := SetConfiguredMode(mode); err != nil {
			t.Fatal(err)
		}
		if got, err := ConfiguredMode(); err != nil || got != mode {
			t.Fatalf("configured mode = %q, %v; want %q", got, err, mode)
		}
	}
	if err := SetConfiguredMode("invalid"); err == nil {
		t.Fatal("expected invalid mode error")
	}
	if got, err := ConfiguredMode(); err != nil || got != "full" {
		t.Fatalf("invalid mode changed config: %q, %v", got, err)
	}
	if err := SetConfiguredLanguage("pt"); err != nil {
		t.Fatal(err)
	}
	if got, err := ConfiguredLanguage(); err != nil || got != "pt" {
		t.Fatalf("configured language = %q, %v", got, err)
	}
}
