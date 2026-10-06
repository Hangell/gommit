// SPDX-License-Identifier: GPL-3.0-only

package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestBuildMessageFromFlags(t *testing.T) {
	got := buildOrPromptMessage("feat", " cli ", " add tests ", `first\nsecond`, "Closes #42", true, "full")
	want := "feat(cli): 💡 add tests\n\nfirst\nsecond\n\nCloses #42"
	if got != want {
		t.Fatalf("message = %q, want %q", got, want)
	}
}

func TestIssueFooter(t *testing.T) {
	got := splitCSVNums(" 12, #34, ,56 ")
	if want := []string{"#12", "#34", "#56"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("issue IDs = %v, want %v", got, want)
	}
	if got := buildIssueFooter(got, []string{"#78"}); got != "Closes #12\nCloses #34\nCloses #56\nRefs #78" {
		t.Fatalf("issue footer = %q", got)
	}
	if got := buildIssueFooter(nil, nil); got != "" {
		t.Fatalf("empty footer = %q", got)
	}
}

// Exercise main in a subprocess so exit codes and log.Fatal cannot terminate
// the test runner. Each process gets its own cwd and Git configuration.
func TestCLIProcess(t *testing.T) {
	if os.Getenv("GOMMIT_TEST_CLI") != "1" {
		return
	}
	for i, arg := range os.Args {
		if arg == "--" {
			os.Args = append([]string{"gommit"}, os.Args[i+1:]...)
			break
		}
	}
	version = "dev"
	main()
	os.Exit(0)
}

func runCLI(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], append([]string{"-test.run=^TestCLIProcess$", "--", "--language=en"}, args...)...)
	cmd.Dir = dir
	cmd.Env = append(isolatedGitEnv(dir), "GOMMIT_TEST_CLI=1")
	out, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("CLI did not exit: %v\n%s", ctx.Err(), out)
	}
	return string(out), err
}

func TestCLIFlags(t *testing.T) {
	for _, tt := range []struct {
		name string
		args []string
		want string
		fail bool
	}{
		{"version", []string{"--version"}, "gommit version dev", false},
		{"help", []string{"--help"}, "-as-editor", false},
		{"unknown flag", []string{"--unknown-option"}, "flag provided but not defined", true},
		{"invalid language", []string{"--language=invalid"}, "invalid", true},
		{"invalid mode", []string{"--mode=invalid"}, "invalid", true},
		{"missing editor file", []string{"--as-editor"}, "missing COMMIT_EDITMSG path", true},
		{"outside repository", []string{"--type=feat", "--subject=example"}, "not a git repository", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			out, err := runCLI(t, t.TempDir(), tt.args...)
			if (err != nil) != tt.fail || !strings.Contains(out, tt.want) {
				t.Fatalf("CLI = %q, %v; want %q, failure=%v", out, err, tt.want, tt.fail)
			}
		})
	}
}

func TestCLIEditor(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "COMMIT_EDITMSG")
	if err := os.WriteFile(path, []byte("old message"), 0o600); err != nil {
		t.Fatal(err)
	}
	out, err := runCLI(t, dir, "--as-editor", "--type=fix", "--scope=cli", "--subject=fix prompt", path)
	if err != nil {
		t.Fatalf("editor failed: %v\n%s", err, out)
	}
	if data, err := os.ReadFile(path); err != nil || string(data) != "fix(cli): 🐛 fix prompt" {
		t.Fatalf("editor message = %q, %v", data, err)
	}
	for _, args := range [][]string{
		{"--type=unknown", "--subject=example"},
		{"--type=fix", "--subject=" + strings.Repeat("界", 73)},
	} {
		out, err := runCLI(t, dir, append([]string{"--as-editor"}, append(args, path)...)...)
		if err == nil {
			t.Fatalf("accepted invalid message: %s", out)
		}
		if data, err := os.ReadFile(path); err != nil || string(data) != "fix(cli): 🐛 fix prompt" {
			t.Fatalf("invalid input changed editor file: %q, %v", data, err)
		}
	}
}

func isolatedGitEnv(dir string) []string {
	var env []string
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "GIT_") && !strings.HasPrefix(entry, "GOMMIT_TEST_CLI=") {
			env = append(env, entry)
		}
	}
	return append(env, "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_COUNT=0", "GIT_CONFIG_GLOBAL="+filepath.Join(dir, "gitconfig"))
}

func TestCLICommitAndPreview(t *testing.T) {
	dir := t.TempDir()
	git := func(args ...string) string {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "git", args...)
		cmd.Dir, cmd.Env = dir, isolatedGitEnv(dir)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	git("init", "--initial-branch=main", "--template="+t.TempDir())
	git("config", "user.name", "Gommit CLI Test")
	git("config", "user.email", "cli@example.invalid")
	git("config", "commit.gpgsign", "false")
	git("config", "core.hooksPath", t.TempDir())
	if err := os.WriteFile(filepath.Join(dir, "example.txt"), []byte("example"), 0o600); err != nil {
		t.Fatal(err)
	}
	out, err := runCLI(t, dir, "--dry-run", "--allow-empty", "--auto-stage=false", "--type=feat", "--subject=add example")
	if err != nil || !strings.Contains(out, "feat: 💡 add example") {
		t.Fatalf("preview = %q, %v", out, err)
	}
	if staged := git("diff", "--cached", "--name-only"); staged != "" {
		t.Fatalf("preview staged files: %s", staged)
	}
	if status := git("status", "--porcelain"); !strings.Contains(status, "?? example.txt") {
		t.Fatalf("preview changed worktree: %s", status)
	}
	out, err = runCLI(t, dir, "--type=feat", "--subject=add example", "--signoff")
	if err != nil {
		t.Fatalf("commit failed: %v\n%s", err, out)
	}
	if msg := git("log", "-1", "--pretty=%B"); !strings.HasPrefix(msg, "feat: 💡 add example") || !strings.Contains(msg, "Signed-off-by: Gommit CLI Test <cli@example.invalid>") {
		t.Fatalf("committed message = %q", msg)
	}
	if status := git("status", "--porcelain"); status != "" {
		t.Fatalf("commit left dirty worktree: %s", status)
	}
	if count := git("rev-list", "--count", "HEAD"); count != "1" {
		t.Fatalf("preview created a commit: commit count %s", count)
	}
}
