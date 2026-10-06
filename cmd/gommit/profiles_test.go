// SPDX-License-Identifier: GPL-3.0-only
package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func profileRepo(t *testing.T) (string, func(...string) string) {
	t.Helper()
	dir := t.TempDir()
	git := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = isolatedGitEnv(dir)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	git("init", "--initial-branch=main", "--template="+t.TempDir())
	git("config", "user.name", "Test")
	git("config", "user.email", "test@example.invalid")
	git("config", "commit.gpgsign", "false")
	git("config", "core.hooksPath", t.TempDir())
	return dir, git
}
func TestProfileCLI(t *testing.T) {
	dir, git := profileRepo(t)
	if err := os.WriteFile(filepath.Join(dir, "file"), []byte("unstaged"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		args []string
		want string
		fail bool
	}{
		{[]string{"--plain", "--subject=Normal message", "--dry-run"}, "Normal message", false},
		{[]string{"--format=conventional", "--type=custom", "--subject=change", "--dry-run"}, "custom: change", false},
		{[]string{"--no-emoji", "--type=fix", "--subject=change", "--dry-run"}, "fix: change", false},
		{[]string{"--json", "--format=conventional", "--subject=change"}, "--type is required", true},
		{[]string{"--json"}, "required", true},
		{[]string{"--json", "--message=normal"}, "no staged changes", true},
		{[]string{"--json", "--dry-run", "--message=normal", "--auto-stage"}, `"committed":false`, false},
		{[]string{"--json", "--format=emoji", "--type=bad type", "--subject=change", "--auto-stage"}, "invalid --type", true},
		{[]string{"--json", "--detect"}, `"source":"fallback"`, false},
	} {
		out, err := runCLI(t, dir, tt.args...)
		if (err != nil) != tt.fail || !strings.Contains(out, tt.want) {
			t.Fatalf("%v: %s %v", tt.args, out, err)
		}
		if staged := git("diff", "--cached", "--name-only"); staged != "" {
			t.Fatal("unexpected staging", staged)
		}
	}
	// Invalid JSON automation errors must be one machine-readable object.
	out, err := runCLI(t, dir, "--json", "--subject=x", "--format=invalid")
	var result map[string]any
	if err == nil || json.Unmarshal([]byte(out), &result) != nil || result["ok"] != false {
		t.Fatalf("%s %v", out, err)
	}
	git("add", "file")
	if err := os.WriteFile(filepath.Join(dir, "file"), []byte("new unstaged content"), 0600); err != nil {
		t.Fatal(err)
	}
	out, err = runCLI(t, dir, "--non-interactive", "--message=Commit staged content")
	if err != nil {
		t.Fatalf("%s %v", out, err)
	}
	if data := git("show", "HEAD:file"); data != "unstaged" {
		t.Fatal("committed unstaged content", data)
	}
	if git("status", "--porcelain") == "" {
		t.Fatal("lost unstaged edits")
	}
}
func TestInitAndConfiguration(t *testing.T) {
	dir, git := profileRepo(t)
	out, err := runCLI(t, dir, "init", "--format=conventional")
	if err != nil {
		t.Fatalf("%s %v", out, err)
	}
	original, err := os.ReadFile(filepath.Join(dir, ".gommit.json"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runCLI(t, dir, "init", "--format=plain"); err == nil {
		t.Fatal("overwrote existing config")
	}
	after, _ := os.ReadFile(filepath.Join(dir, ".gommit.json"))
	if string(after) != string(original) {
		t.Fatal("config changed")
	}
	config := `{"format":"emoji","emoji":false,"types":["task"],"subject_limit":10,"scope_required":true}`
	if err := os.WriteFile(filepath.Join(dir, ".gommit.json"), []byte(config), 0600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"--type=fix", "--scope=cli", "--subject=ok"}, {"--type=task", "--subject=ok"}, {"--type=task", "--scope=cli", "--subject=too long subject"}} {
		if out, err := runCLI(t, dir, append(args, "--json", "--dry-run")...); err == nil {
			t.Fatalf("accepted %v %s", args, out)
		}
	}
	out, err = runCLI(t, dir, "--json", "--dry-run", "--type=task", "--scope=cli", "--subject=ok")
	if err != nil || !strings.Contains(out, "task(cli): ok") {
		t.Fatalf("%s %v", out, err)
	}
	git("config", "gommit.format", "plain")
	out, err = runCLI(t, dir, "doctor", "--json")
	if err != nil || !strings.Contains(out, `"format":"plain"`) || !strings.Contains(out, `"source":"git-config"`) {
		t.Fatalf("%s %v", out, err)
	}
}
func TestWorktreeAndHistory(t *testing.T) {
	_, git := profileRepo(t)
	for range 5 {
		git("commit", "--allow-empty", "-m", "fix: history")
	}
	worktree := filepath.Join(t.TempDir(), "linked")
	git("worktree", "add", "-b", "linked", worktree)
	out, err := runCLI(t, worktree, "--detect", "--json")
	if err != nil || !strings.Contains(out, `"format":"conventional"`) {
		t.Fatalf("%s %v", out, err)
	}
	out, err = runCLI(t, worktree, "--non-interactive", "--format=auto", "--allow-empty", "--type=fix", "--subject=worktree")
	if err != nil {
		t.Fatalf("%s %v", out, err)
	}
	out, err = runCLI(t, worktree, "--non-interactive", "--format=auto", "--amend", "--allow-empty", "--type=fix", "--subject=amended")
	if err != nil {
		t.Fatalf("%s %v", out, err)
	}
	if git("rev-list", "--count", "linked") != "6" {
		t.Fatal("amend created extra commit")
	}
}
func TestCommitHookRejection(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell hook fixture requires executable POSIX scripts")
	}
	dir, git := profileRepo(t)
	hooks := t.TempDir()
	git("config", "core.hooksPath", hooks)
	if err := os.WriteFile(filepath.Join(hooks, "commit-msg"), []byte("#!/bin/sh\necho rejected >&2\nexit 1\n"), 0700); err != nil {
		t.Fatal(err)
	}
	out, err := runCLI(t, dir, "--non-interactive", "--allow-empty", "--message=normal")
	if err == nil || !strings.Contains(out, "rejected") {
		t.Fatalf("%s %v", out, err)
	}
	out, err = runCLI(t, dir, "--non-interactive", "--allow-empty", "--message=normal", "--no-verify")
	if err != nil {
		t.Fatalf("%s %v", out, err)
	}
}

func TestPipedWizardAndJSONCommit(t *testing.T) {
	dir, _ := profileRepo(t)
	run := func(input string, args ...string) (string, error) {
		t.Helper()
		cmd := exec.Command(os.Args[0], append([]string{"-test.run=^TestCLIProcess$", "--", "--language=en"}, args...)...)
		cmd.Dir = dir
		cmd.Env = append(isolatedGitEnv(dir), "GOMMIT_TEST_CLI=1")
		cmd.Stdin = strings.NewReader(input)
		out, err := cmd.Output()
		return string(out), err
	}
	out, err := run("feat\nAdd a feature\n", "--dry-run")
	if err != nil || !strings.Contains(out, "feat: 💡 Add a feature") {
		t.Fatalf("pipe: %s %v", out, err)
	}
	out, err = run("", "--json", "--allow-empty", "--message", "doctor")
	var result map[string]any
	if err != nil || json.Unmarshal([]byte(out), &result) != nil || result["committed"] != true || result["message"] != "doctor" {
		t.Fatalf("JSON commit: %s %v", out, err)
	}
	out, err = run("--update\n\nMessage body", "--json", "--dry-run", "--file=-")
	if err != nil || json.Unmarshal([]byte(out), &result) != nil || result["message"] != "--update\n\nMessage body" {
		t.Fatalf("file input: %s %v", out, err)
	}
}
