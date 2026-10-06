package git

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

const configModeKey = "gommit.mode"
const configLanguageKey = "gommit.language"

func ConfiguredLanguage() (string, error) { return ConfigValue(configLanguageKey) }
func SetConfiguredLanguage(language string) error {
	cmd := exec.Command("git", "config", "--global", configLanguageKey, language)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

func ConfiguredMode() (string, error) {
	return ConfigValue(configModeKey)
}

func ConfigValue(key string) (string, error) {
	out, err := exec.Command("git", "config", "--get", key).Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 1 {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func SetConfiguredMode(mode string) error {
	if mode != "simple" && mode != "full" {
		return fmt.Errorf("invalid mode %q (use simple or full)", mode)
	}
	cmd := exec.Command("git", "config", "--global", configModeKey, mode)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type Options struct {
	AllowEmpty bool
	Amend      bool
	NoVerify   bool
	Signoff    bool
	Output     io.Writer
}

func InRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

// Há mudanças stageadas?
func HasStagedChanges() (bool, error) {
	cmd := exec.Command("git", "diff", "--cached", "--quiet")
	if err := cmd.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 1 {
			return true, nil // existem mudanças stageadas
		}
		return false, err
	}
	return false, nil // sem mudanças stageadas
}

// Working tree tem alterações (untracked/modified) não stageadas?
func WorkingTreeDirty() (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return len(bytes.TrimSpace(out)) > 0, nil
}

// Faz git add -A (tudo)
func StageAll() error {
	cmd := exec.Command("git", "add", "-A")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Lista o que está stageado (name-status)
func StagedSummary() (string, error) {
	cmd := exec.Command("git", "diff", "--cached", "--name-status")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// Último commit (assunto & corpo)
func LastCommitSubject() (string, error) {
	cmd := exec.Command("git", "log", "-1", "--pretty=%s")
	out, err := cmd.Output()
	return string(bytes.TrimSpace(out)), err
}

func LastCommitMessage() (string, error) {
	cmd := exec.Command("git", "log", "-1", "--pretty=%B")
	out, err := cmd.Output()
	return string(bytes.TrimSpace(out)), err
}

// Realiza o commit com -F <arquivo temporário>
func CommitWithMessage(msg string, opts Options) error {
	f, err := os.CreateTemp("", "gommit-*.txt")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())

	if _, err := f.WriteString(msg); err != nil {
		f.Close()
		return err
	}
	f.Close()

	args := []string{"commit", "-F", f.Name()}
	if opts.AllowEmpty {
		args = append(args, "--allow-empty")
	}
	if opts.Amend {
		args = append(args, "--amend")
	}
	if opts.NoVerify {
		args = append(args, "--no-verify")
	}
	if opts.Signoff {
		args = append(args, "-s")
	}

	cmd := exec.Command("git", args...)
	cmd.Stdout = opts.Output
	if cmd.Stdout == nil {
		cmd.Stdout = os.Stdout
	}
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Usado no modo --as-editor
func WriteCommitEditMsg(path, msg string) error {
	return os.WriteFile(path, []byte(msg), 0o644)
}

// Root resolves the worktree root even when invoked from a subdirectory.
func Root() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("not a git repository: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func RecentSubjects(limit int) ([]string, error) {
	out, err := exec.Command("git", "log", "--no-merges", fmt.Sprintf("-%d", limit), "--format=%s").Output()
	if err != nil {
		// A valid repository with an unborn branch has no history to infer.
		if exec.Command("git", "rev-parse", "--verify", "HEAD").Run() != nil {
			return nil, nil
		}
		return nil, err
	}
	return strings.Split(strings.TrimSpace(string(out)), "\n"), nil
}

func ConfigBool(key string) (*bool, error) {
	out, err := exec.Command("git", "config", "--bool", "--get", key).Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 1 {
			return nil, nil
		}
		return nil, fmt.Errorf("invalid boolean setting %s: %w", key, err)
	}
	value := strings.TrimSpace(string(out)) == "true"
	return &value, nil
}
