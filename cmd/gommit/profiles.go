// SPDX-License-Identifier: GPL-3.0-only
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Hangell/gommit/internal/commit"
	"github.com/Hangell/gommit/internal/git"
	"github.com/Hangell/gommit/internal/i18n"
	"github.com/Hangell/gommit/internal/profile"
	"github.com/Hangell/gommit/internal/ui"
	gommitupdate "github.com/Hangell/gommit/internal/update"
)

// optionTokens skips values, so messages such as "doctor" and "--update"
// cannot accidentally dispatch a command.
func optionTokens(args []string) []int {
	var indices []int
	values := map[string]bool{"format": true, "type": true, "scope": true, "subject": true, "body": true, "footer": true, "message": true, "m": true, "file": true, "F": true, "mode": true, "language": true, "set-mode": true, "set-language": true}
	for i := 0; i < len(args); i++ {
		if args[i] == "--" {
			break
		}
		indices = append(indices, i)
		key := strings.TrimLeft(args[i], "-")
		if !strings.Contains(key, "=") && values[key] {
			i++
		}
	}
	return indices
}

func main() {
	args := os.Args[1:]
	for _, index := range optionTokens(args) {
		arg := args[index]
		key := strings.TrimLeft(strings.SplitN(arg, "=", 2)[0], "-")
		if strings.HasSuffix(arg, "=false") {
			continue
		}
		switch key {
		case "apply-update", "install", "update", "version", "set-mode", "set-language", "help", "h":
			administrativeMain()
			return
		}
	}
	configureLanguage(args)
	machine := false
	for _, index := range optionTokens(args) {
		arg := args[index]
		arg = strings.TrimLeft(arg, "-")
		if arg == "json" || arg == "json=true" {
			machine = true
		} else if arg == "json=false" {
			machine = false
		}
	}
	if err := runProfiles(args); err != nil {
		if machine {
			_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"ok": false, "error": err.Error()})
		} else {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}

func runProfiles(args []string) error {
	command := ""
	editorRequested := false
	for _, i := range optionTokens(args) {
		if args[i] == "--as-editor" || args[i] == "-as-editor" || args[i] == "--as-editor=true" {
			editorRequested = true
		}
	}
	// Subcommands may follow global flags, as in gommit --language=en doctor.
	for _, i := range optionTokens(args) {
		arg := args[i]
		if !editorRequested && (arg == "doctor" || arg == "init") {
			command = arg
			args = append(append([]string{}, args[:i]...), args[i+1:]...)
			break
		}
	}
	fs := flag.NewFlagSet("gommit", flag.ContinueOnError)
	format := fs.String("format", "", "plain, conventional, emoji, auto")
	plain := fs.Bool("plain", false, "Plain message")
	noEmoji := fs.Bool("no-emoji", false, "Disable emojis")
	nonInteractive := fs.Bool("non-interactive", false, "Never prompt")
	machine := fs.Bool("json", false, "JSON output")
	detect := fs.Bool("detect", false, "Inspect convention")
	dry := fs.Bool("dry-run", false, "Preview without staging")
	kind := fs.String("type", "", "Commit type")
	scope := fs.String("scope", "", "Scope")
	subject := fs.String("subject", "", "Subject")
	body := fs.String("body", "", "Body")
	footer := fs.String("footer", "", "Footer")
	var raw, file string
	fs.StringVar(&raw, "message", "", "Plain message")
	fs.StringVar(&raw, "m", "", "Plain message")
	fs.StringVar(&file, "file", "", "Message file")
	fs.StringVar(&file, "F", "", "Message file")
	mode := fs.String("mode", "", "simple or full")
	language := fs.String("language", "", "Language")
	editor := fs.Bool("as-editor", false, "Edit message file")
	allowEmpty := fs.Bool("allow-empty", false, "Allow empty commit")
	amend := fs.Bool("amend", false, "Amend")
	noVerify := fs.Bool("no-verify", false, "Skip hooks")
	signoff := fs.Bool("signoff", false, "Sign off")
	stage := fs.Bool("auto-stage", true, "Stage changes if index is empty")
	status := fs.Bool("show-status", true, "Show staged summary")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *language != "" {
		if i18n.Normalize(*language) == "" {
			return fmt.Errorf("invalid language %q", *language)
		}
		i18n.Set(*language)
	}
	if *mode == "" {
		var err error
		*mode, err = git.ConfiguredMode()
		if err != nil {
			return err
		}
		if *mode == "" {
			*mode = "simple"
		}
	}
	if *mode != "simple" && *mode != "full" {
		return fmt.Errorf("invalid mode %q", *mode)
	}
	if *editor && fs.NArg() != 1 {
		return errors.New("editor mode: missing COMMIT_EDITMSG path (exactly one required)")
	}
	if !*editor && fs.NArg() != 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}
	*nonInteractive = *nonInteractive || *machine
	stageExplicit := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "auto-stage" {
			stageExplicit = true
		}
	})
	if *nonInteractive && !stageExplicit {
		*stage = false
	}
	p := profile.Profile{Format: profile.Emoji, Emoji: true, Source: "default", ObservedTypes: []string{}}
	if git.InRepo() {
		discovered, err := profile.Discover()
		if err != nil {
			return err
		}
		if discovered.Explicit || *format == profile.Auto || *nonInteractive || *detect || command != "" {
			p = discovered
		}
	} else if !*editor {
		return errors.New("not a git repository (or any of the parent directories)")
	}
	if *format != "" {
		if !profile.ValidFormat(*format) {
			return fmt.Errorf("invalid format %q", *format)
		}
		if *format != profile.Auto {
			p.Format = *format
			p.Emoji = *format == profile.Emoji
			p.Source = "flag"
			p.Confidence = 1
		}
	}
	if *plain {
		if *format != "" && *format != profile.Plain {
			return errors.New("--plain conflicts with --format")
		}
		p.Format = profile.Plain
		p.Emoji = false
		p.Source = "flag"
	}
	if *noEmoji || os.Getenv("NO_EMOJI") != "" {
		p.Emoji = false
		if p.Format == profile.Emoji {
			p.Format = profile.Conventional
		}
	}
	if *detect || command == "doctor" {
		if *machine {
			return json.NewEncoder(os.Stdout).Encode(p)
		}
		fmt.Printf("Format: %s\nSource: %s\nConfidence: %.0f%% (%d commits)\nObserved types: %s\nAllowed types: %s\n", p.Format, p.Source, p.Confidence*100, p.SampleSize, strings.Join(p.ObservedTypes, ", "), strings.Join(p.AllowedTypes, ", "))
		staged, err := git.HasStagedChanges()
		if err != nil {
			return err
		}
		fmt.Printf("Staged changes: %t\n", staged)
		return nil
	}
	if command == "init" {
		if *dry {
			return errors.New("init writes configuration; --dry-run is not supported")
		}
		root, err := git.Root()
		if err != nil {
			return err
		}
		path := filepath.Join(root, ".gommit.json")
		data, err := json.MarshalIndent(profile.Config{Format: p.Format, Types: p.AllowedTypes, SubjectLimit: p.SubjectLimit, ScopeRequired: p.ScopeRequired}, "", "  ")
		if err != nil {
			return err
		}
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if err != nil {
			return fmt.Errorf("configuration not overwritten: %w", err)
		}
		_, err = f.Write(append(data, '\n'))
		closeErr := f.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
		if *machine {
			return json.NewEncoder(os.Stdout).Encode(map[string]any{"ok": true, "path": path, "profile": p})
		}
		fmt.Println("Created", path)
		return nil
	}
	if raw != "" && file != "" {
		return errors.New("--message and --file are mutually exclusive")
	}
	if raw != "" || file != "" {
		if *kind != "" || *scope != "" || *subject != "" || *body != "" || *footer != "" || (*format != "" && *format != profile.Plain) {
			return errors.New("raw messages require plain format and cannot be mixed with structured fields")
		}
		p.Format = profile.Plain
		p.Emoji = false
		p.Source = "message"
		if file != "" {
			var data []byte
			var err error
			if file == "-" {
				data, err = io.ReadAll(os.Stdin)
			} else {
				data, err = os.ReadFile(file)
			}
			if err != nil {
				return err
			}
			raw = string(data)
		}
		if strings.TrimSpace(raw) == "" {
			return errors.New("message is required")
		}
	} else {
		if p.Format == profile.Plain && (*kind != "" || *scope != "") {
			return errors.New("plain format does not accept --type or --scope; select --format conventional")
		}
		if *nonInteractive && strings.TrimSpace(*subject) == "" {
			return errors.New("--subject or --message is required in non-interactive mode")
		}
		if p.Format != profile.Plain {
			if *kind == "" {
				if *nonInteractive {
					return errors.New("--type is required for conventional and emoji formats")
				}
				if len(p.AllowedTypes) > 0 {
					*kind = promptLine("Type (" + strings.Join(p.AllowedTypes, ", ") + "): ")
				} else {
					t, err := ui.SelectCommitTypeWithReader(promptInput)
					if err != nil {
						return err
					}
					*kind = t.Key
				}
			}
			if !profile.ValidType(*kind) {
				return fmt.Errorf("invalid --type %q", *kind)
			}
			if p.Source == "default" {
				if _, ok := ui.FindType(*kind); !ok {
					return fmt.Errorf("invalid --type %q", *kind)
				}
			}
			if len(p.AllowedTypes) > 0 {
				found := false
				for _, v := range p.AllowedTypes {
					if v == *kind {
						found = true
					}
				}
				if !found {
					return fmt.Errorf("type %q is not allowed by repository configuration", *kind)
				}
			}
			if strings.ContainsAny(*scope, "()\r\n") {
				return errors.New("scope cannot contain parentheses or newlines")
			}
		}
		if *subject == "" {
			*subject = promptLine(i18n.T("prompt.subject"))
		}
		if !*nonInteractive && !*editor && *mode == "full" {
			if p.Format != profile.Plain && *scope == "" {
				*scope = promptLine(i18n.T("prompt.scope"))
			}
			if *body == "" {
				fmt.Println(i18n.T("prompt.body"))
				*body = readMultiline()
			}
			if promptYesNo(i18n.T("prompt.breaking"), false) {
				desc := promptLine(i18n.T("prompt.breaking_desc"))
				if desc == "" {
					return errors.New("breaking description is required")
				}
				*footer = strings.TrimSpace(*footer + "\n\nBREAKING CHANGE: " + desc)
			}
			if promptYesNo(i18n.T("prompt.issues"), false) {
				c := promptLine(i18n.T("prompt.closes"))
				r := promptLine(i18n.T("prompt.refs"))
				*footer = strings.TrimSpace(*footer + "\n\n" + buildIssueFooter(splitCSVNums(c), splitCSVNums(r)))
			}
		}
		if strings.ContainsAny(*scope, "()\r\n") {
			return errors.New("scope cannot contain parentheses or newlines")
		}
		if p.ScopeRequired && p.Format != profile.Plain && strings.TrimSpace(*scope) == "" {
			return errors.New("repository requires a scope")
		}
		m := commit.Message{SubjectLimit: p.SubjectLimit, Plain: p.Format == profile.Plain, Type: *kind, Scope: *scope, Subject: *subject, Body: unescapeNewlines(*body), Footer: unescapeNewlines(*footer)}
		if p.Emoji {
			m.Emoji = ui.EmojiFor(*kind)
		}
		m.Breaking = !m.Plain && strings.Contains(m.Footer, "BREAKING CHANGE:")
		if err := m.Validate(); err != nil {
			return err
		}
		raw = m.Build()
	}
	if *dry {
		if *machine {
			return json.NewEncoder(os.Stdout).Encode(map[string]any{"ok": true, "message": raw, "profile": p, "dry_run": true, "committed": false})
		}
		fmt.Println(i18n.T("preview"))
		fmt.Println(raw)
		return nil
	}
	if *editor {
		if err := git.WriteCommitEditMsg(fs.Arg(0), raw); err != nil {
			return err
		}
		if *machine {
			return json.NewEncoder(os.Stdout).Encode(map[string]any{"ok": true, "message": raw, "committed": false})
		}
		return nil
	}
	output := io.Writer(os.Stdout)
	if *machine {
		output = os.Stderr
	}
	if !*allowEmpty && !*amend {
		staged, err := git.HasStagedChanges()
		if err != nil {
			return err
		}
		if !staged {
			dirty, err := git.WorkingTreeDirty()
			if err != nil {
				return err
			}
			if !dirty {
				return errors.New("nothing to commit. Working tree clean (use --allow-empty or --amend)")
			}
			if !*stage {
				return errors.New("no staged changes. Run 'git add .' or pass --auto-stage")
			}
			if err := git.StageAll(); err != nil {
				return err
			}
		}
		if *status && !*machine {
			summary, err := git.StagedSummary()
			if err != nil {
				return err
			}
			fmt.Fprintln(output, summary)
		}
	}
	if err := git.CommitWithMessage(raw, git.Options{AllowEmpty: *allowEmpty, Amend: *amend, NoVerify: *noVerify, Signoff: *signoff, Output: output}); err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}
	if *machine {
		return json.NewEncoder(os.Stdout).Encode(map[string]any{"ok": true, "message": raw, "profile": p, "committed": true})
	}
	if !*nonInteractive {
		if notice := gommitupdate.Notice(version); notice != "" {
			fmt.Println(notice)
		}
	}
	return nil
}
