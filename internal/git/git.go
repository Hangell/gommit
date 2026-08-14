package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const configModeKey = "gommit.mode"
const configLanguageKey = "gommit.language"

func ConfiguredLanguage() (string, error) { return globalConfig(configLanguageKey) }
func SetConfiguredLanguage(language string) error {
	cmd := exec.Command("git", "config", "--global", configLanguageKey, language)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

func ConfiguredMode() (string, error) {
	return globalConfig(configModeKey)
}

func globalConfig(key string) (string, error) {
	out, err := exec.Command("git", "config", "--global", "--get", key).Output()
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
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Usado no modo --as-editor
func WriteCommitEditMsg(path, msg string) error {
	return os.WriteFile(path, []byte(msg), 0o644)
}
