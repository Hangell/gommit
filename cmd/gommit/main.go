// SPDX-License-Identifier: GPL-3.0-only

package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/Hangell/gommit/internal/git"
	"github.com/Hangell/gommit/internal/i18n"
	"github.com/Hangell/gommit/internal/install"
	gommitupdate "github.com/Hangell/gommit/internal/update"
)

var version = "dev"

func configureLanguage(args []string) {
	language, _ := git.ConfiguredLanguage()
	if language == "" {
		language = i18n.SystemLanguage()
	}
	for i, arg := range args {
		if strings.HasPrefix(arg, "--language=") {
			language = strings.TrimPrefix(arg, "--language=")
		}
		if arg == "--language" && i+1 < len(args) {
			language = args[i+1]
		}
	}
	if language != "" {
		i18n.Set(language)
	}
}

func administrativeMain() {
	if len(os.Args) == 5 && os.Args[1] == "--apply-update" {
		pid, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatalf("invalid updater parent PID: %v", err)
		}
		if err := gommitupdate.Apply(pid, os.Args[3], os.Args[4]); err != nil {
			log.Fatalf("update installation failed: %v", err)
		}
		return
	}
	configureLanguage(os.Args[1:])

	fs := flag.NewFlagSet("gommit", flag.ContinueOnError)
	fs.String("format", "", "Commit format: plain, conventional, emoji, auto")
	fs.Bool("plain", false, "Use a plain Git message")
	fs.Bool("no-emoji", false, "Disable header emojis")
	fs.Bool("non-interactive", false, "Never prompt; only commit staged changes by default")
	fs.Bool("json", false, "Machine-readable output; implies non-interactive")
	fs.Bool("detect", false, "Inspect repository convention without changing files")
	fs.String("message", "", "Plain Git message (alias -m)")
	fs.String("file", "", "Plain message file (alias -F; - reads stdin)")
	fs.String("m", "", "Alias for --message")
	fs.String("F", "", "Alias for --file")

	showVersion := fs.Bool("version", false, i18n.T("help.version"))
	doInstall := fs.Bool("install", false, i18n.T("help.install"))
	doUpdate := fs.Bool("update", false, i18n.T("help.update"))
	fs.Bool("dry-run", false, i18n.T("help.dry_run"))
	fs.String("mode", "", i18n.T("help.mode"))
	setMode := fs.String("set-mode", "", i18n.T("help.set_mode"))
	languageFlag := fs.String("language", "", i18n.T("help.language"))
	setLanguage := fs.String("set-language", "", i18n.T("help.set_language"))

	fs.String("type", "", i18n.T("help.type"))
	fs.String("scope", "", i18n.T("help.scope"))
	fs.String("subject", "", i18n.T("help.subject"))
	fs.String("body", "", i18n.T("help.body"))
	fs.String("footer", "", i18n.T("help.footer"))
	fs.Bool("as-editor", false, i18n.T("help.editor"))

	fs.Bool("allow-empty", false, i18n.T("help.allow_empty"))
	fs.Bool("amend", false, i18n.T("help.amend"))
	fs.Bool("no-verify", false, i18n.T("help.no_verify"))
	fs.Bool("signoff", false, i18n.T("help.signoff"))
	fs.Bool("auto-stage", true, i18n.T("help.auto_stage"))
	fs.Bool("show-status", true, i18n.T("help.show_status"))
	fs.Lookup("auto-stage").DefValue = ""
	fs.Lookup("show-status").DefValue = ""
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), i18n.T("help.usage"))
		fmt.Fprintln(fs.Output(), "Commands: gommit init [--format FORMAT]; gommit doctor [--json]")
		fs.PrintDefaults()
	}

	fs.SetOutput(os.Stderr)
	if err := fs.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		os.Exit(2)
	}

	if *showVersion {
		fmt.Println("gommit version", version)
		return
	}
	if *languageFlag != "" {
		language := i18n.Normalize(*languageFlag)
		if language == "" {
			log.Fatalf(i18n.T("language.invalid"), *languageFlag, i18n.Supported())
		}
		i18n.Set(language) // help was already localized by configureLanguage
	}
	if *setLanguage != "" {
		language := i18n.Normalize(*setLanguage)
		if language == "" {
			log.Fatalf(i18n.T("language.invalid"), *setLanguage, i18n.Supported())
		}
		if err := git.SetConfiguredLanguage(language); err != nil {
			log.Fatal(err)
		}
		i18n.Set(language)
		fmt.Println(i18n.T("language.saved", language))
		return
	}
	if *doUpdate {
		message, err := gommitupdate.Start(version)
		if err != nil {
			log.Fatalf("update failed: %v", err)
		}
		fmt.Println(message)
		return
	}
	if *setMode != "" {
		mode := strings.ToLower(strings.TrimSpace(*setMode))
		if err := git.SetConfiguredMode(mode); err != nil {
			log.Fatalf("could not save mode: %v", err)
		}
		fmt.Println(i18n.T("mode.saved", mode))
		return
	}
	// Installation does not require Git or a gommit configuration.
	if *doInstall {
		if saved, _ := git.ConfiguredLanguage(); saved == "" {
			_ = git.SetConfiguredLanguage(i18n.SystemLanguage())
		}
		res, err := install.InstallSelf(version)
		if err != nil {
			log.Fatalf("install failed: %v", err)
		}
		fmt.Println(res.Message)
		return
	}

}

// --- wizard / montagem ---------------------------------------------------

var promptInput = bufio.NewReader(os.Stdin)

func promptLine(label string) string {
	fmt.Print(label)
	r := promptInput
	s, _ := r.ReadString('\n')
	return strings.TrimSpace(s)
}

func readMultiline() string {
	var lines []string
	emptyStreak := 0
	for {
		line, err := promptInput.ReadString('\n')
		if err != nil && line == "" {
			break
		}
		line = strings.TrimRight(line, "\r\n")
		trim := strings.TrimSpace(line)
		if trim == "." {
			break
		}
		if trim == "" {
			emptyStreak++
			if emptyStreak >= 2 {
				break
			}
		} else {
			emptyStreak = 0
		}
		lines = append(lines, line)
		if err != nil {
			break
		}
	}
	return strings.Join(lines, "\n")
}

func unescapeNewlines(s string) string {
	return strings.ReplaceAll(s, `\n`, "\n")
}

// helpers cz-like

func promptYesNo(label string, defYes bool) bool {
	suf := "y/N"
	if defYes {
		suf = "Y/n"
	}
	fmt.Printf("%s (%s): ", label, suf)
	r := promptInput
	in, _ := r.ReadString('\n')
	in = strings.TrimSpace(strings.ToLower(in))
	if in == "" {
		return defYes
	}
	return in == "y" || in == "yes" || in == "s"
}

func splitCSVNums(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if !strings.HasPrefix(p, "#") {
			p = "#" + p
		}
		out = append(out, p)
	}
	return out
}

func buildIssueFooter(closes, refs []string) string {
	var lines []string
	for _, id := range closes {
		lines = append(lines, "Closes "+id)
	}
	for _, id := range refs {
		lines = append(lines, "Refs "+id)
	}
	return strings.Join(lines, "\n")
}
