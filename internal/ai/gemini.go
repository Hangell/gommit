package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unicode/utf8"
)

const maxDiffBytes = 512 * 1024

type Suggestion struct {
	Type    string `json:"type"`
	Subject string `json:"subject"`
}

type cliResponse struct {
	Response string `json:"response"`
	Error    any    `json:"error,omitempty"`
}

func GenerateCommit() (Suggestion, error) {
	command, prefix, err := geminiCommand()
	if err != nil {
		return Suggestion{}, errors.New("Gemini CLI was not found in PATH")
	}
	diff, err := stagedDiff()
	if err != nil {
		return Suggestion{}, err
	}
	if len(diff) == 0 {
		return Suggestion{}, errors.New("there are no staged changes for Gemini to analyze")
	}
	if len(diff) > maxDiffBytes {
		return Suggestion{}, fmt.Errorf("staged diff is too large for AI (%d KiB; maximum %d KiB)", len(diff)/1024, maxDiffBytes/1024)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	prompt := `Analyze only the staged Git diff supplied on stdin. Do not use tools, read files, or modify anything. Return one Conventional Commit suggestion as raw JSON only, with exactly this schema: {"type":"feat","subject":"short imperative technical description"}. Allowed types: WIP, feat, fix, chore, refactor, prune, docs, perf, test, build, ci, style, revert. The subject must be meaningful, must not include the type, emoji, scope, period, markdown, or quotes, and must contain at most 72 Unicode characters.`
	args := append(prefix, "--prompt", prompt, "--output-format", "json")
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = os.TempDir() // prevent workspace files/instructions from becoming implicit context
	cmd.Stdin = bytes.NewReader(diff)
	cmd.Env = geminiEnvironment()
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if ctx.Err() != nil {
		return Suggestion{}, errors.New("Gemini timed out after 90 seconds")
	}
	if err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		return Suggestion{}, fmt.Errorf("Gemini CLI failed: %s", message)
	}
	return parseResponse(out)
}

func geminiCommand() (string, []string, error) {
	if runtime.GOOS != "windows" {
		path, err := exec.LookPath("gemini")
		return path, nil, err
	}
	launcher, err := exec.LookPath("gemini.cmd")
	if err != nil {
		return "", nil, err
	}
	psScript := strings.TrimSuffix(launcher, filepath.Ext(launcher)) + ".ps1"
	if _, err := os.Stat(psScript); err != nil {
		return "", nil, err
	}
	powershell, err := exec.LookPath("powershell.exe")
	if err != nil {
		return "", nil, err
	}
	return powershell, []string{"-NoProfile", "-ExecutionPolicy", "Bypass", "-File", psScript}, nil
}

func stagedDiff() ([]byte, error) {
	cmd := exec.Command("git", "diff", "--cached", "--no-ext-diff", "--unified=3", "--")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("could not read staged diff: %w", err)
	}
	return out, nil
}

func geminiEnvironment() []string {
	env := os.Environ()
	if os.Getenv("GEMINI_API_KEY") == "" && os.Getenv("GEMINI_KEY") != "" {
		env = append(env, "GEMINI_API_KEY="+os.Getenv("GEMINI_KEY"))
	}
	return env
}

func parseResponse(data []byte) (Suggestion, error) {
	var envelope cliResponse
	if err := json.Unmarshal(data, &envelope); err != nil {
		return Suggestion{}, fmt.Errorf("invalid Gemini CLI response: %w", err)
	}
	if envelope.Error != nil {
		return Suggestion{}, errors.New("Gemini CLI returned an API error")
	}
	text := strings.TrimSpace(envelope.Response)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)
	var suggestion Suggestion
	if err := json.Unmarshal([]byte(text), &suggestion); err != nil {
		return Suggestion{}, fmt.Errorf("Gemini returned an invalid commit suggestion: %w", err)
	}
	suggestion.Type = strings.TrimSpace(suggestion.Type)
	suggestion.Subject = strings.TrimSpace(suggestion.Subject)
	if suggestion.Type == "" || suggestion.Subject == "" {
		return Suggestion{}, errors.New("Gemini returned an empty type or subject")
	}
	if strings.ContainsAny(suggestion.Subject, "\r\n") {
		return Suggestion{}, errors.New("Gemini returned a multiline subject")
	}
	if utf8.RuneCountInString(suggestion.Subject) > 72 {
		return Suggestion{}, errors.New("Gemini returned a subject longer than 72 characters")
	}
	return suggestion, nil
}
