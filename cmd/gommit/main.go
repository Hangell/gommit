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

	"github.com/Hangell/gommit/internal/commit"
	"github.com/Hangell/gommit/internal/git"
	"github.com/Hangell/gommit/internal/i18n"
	"github.com/Hangell/gommit/internal/install"
	"github.com/Hangell/gommit/internal/ui"
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

func main() {
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
	showVersion := fs.Bool("version", false, i18n.T("help.version"))
	doInstall := fs.Bool("install", false, i18n.T("help.install"))
	doUpdate := fs.Bool("update", false, i18n.T("help.update"))
	dryRun := fs.Bool("dry-run", false, i18n.T("help.dry_run"))
	modeFlag := fs.String("mode", "", i18n.T("help.mode"))
	setMode := fs.String("set-mode", "", i18n.T("help.set_mode"))
	languageFlag := fs.String("language", "", i18n.T("help.language"))
	setLanguage := fs.String("set-language", "", i18n.T("help.set_language"))

	typeFlag := fs.String("type", "", i18n.T("help.type"))
	scopeFlag := fs.String("scope", "", i18n.T("help.scope"))
	subjectFlag := fs.String("subject", "", i18n.T("help.subject"))
	bodyFlag := fs.String("body", "", i18n.T("help.body"))
	footerFlag := fs.String("footer", "", i18n.T("help.footer"))
	asEditor := fs.Bool("as-editor", false, i18n.T("help.editor"))

	allowEmpty := fs.Bool("allow-empty", false, i18n.T("help.allow_empty"))
	amend := fs.Bool("amend", false, i18n.T("help.amend"))
	noVerify := fs.Bool("no-verify", false, i18n.T("help.no_verify"))
	signoff := fs.Bool("signoff", false, i18n.T("help.signoff"))
	autoStage := fs.Bool("auto-stage", true, i18n.T("help.auto_stage"))
	showStatus := fs.Bool("show-status", true, i18n.T("help.show_status"))
	fs.Lookup("auto-stage").DefValue = ""
	fs.Lookup("show-status").DefValue = ""
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), i18n.T("help.usage"))
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

	mode := strings.ToLower(strings.TrimSpace(*modeFlag))
	if mode == "" {
		var err error
		mode, err = git.ConfiguredMode()
		if err != nil {
			log.Fatalf("could not read gommit.mode: %v", err)
		}
		if mode == "" {
			mode = "simple"
		}
	}
	if mode != "simple" && mode != "full" {
		log.Fatalf(i18n.T("mode.invalid"), mode)
	}

	// ===== editor mode =====
	if *asEditor {
		if fs.NArg() < 1 {
			log.Fatal("editor mode: missing COMMIT_EDITMSG path")
		}
		path := fs.Arg(0)
		msg := buildOrPromptMessage(*typeFlag, *scopeFlag, *subjectFlag, *bodyFlag, *footerFlag, true, mode)
		if err := git.WriteCommitEditMsg(path, msg); err != nil {
			log.Fatalf("failed to write commit message: %v", err)
		}
		return
	}

	// ===== precisa ser repo git =====
	if !git.InRepo() {
		log.Fatal("not a git repository (or any of the parent directories)")
	}

	// ===== staged / auto-stage / amend =====
	if *amend {
		if subj, err := git.LastCommitSubject(); err == nil && subj != "" {
			fmt.Println(i18n.T("amend.last", subj))
		} else {
			fmt.Println(i18n.T("amend.update"))
		}
	}
	if !*allowEmpty && !*amend {
		staged, err := git.HasStagedChanges()
		if err != nil {
			log.Fatalf("failed to check staged changes: %v", err)
		}
		if !staged {
			dirty, err := git.WorkingTreeDirty()
			if err != nil {
				log.Fatalf("failed to check working tree: %v", err)
			}
			if dirty {
				if *autoStage {
					fmt.Println(i18n.T("autostage"))
					if err := git.StageAll(); err != nil {
						log.Fatalf("git add -A failed: %v", err)
					}
					if *showStatus {
						if sum, err := git.StagedSummary(); err == nil && strings.TrimSpace(sum) != "" {
							fmt.Println("\n" + i18n.T("staged") + "\n" + strings.TrimRight(sum, "\n"))
						}
					}
				} else {
					log.Fatal("no staged changes. Run 'git add .' or pass --auto-stage")
				}
			} else {
				log.Fatal("nothing to commit. Working tree clean (use --allow-empty or --amend)")
			}
		} else if *showStatus {
			if sum, err := git.StagedSummary(); err == nil && strings.TrimSpace(sum) != "" {
				fmt.Println(i18n.T("staged") + "\n" + strings.TrimRight(sum, "\n"))
			}
		}
	}

	// ===== wizard / mensagem =====
	msg := buildOrPromptMessage(*typeFlag, *scopeFlag, *subjectFlag, *bodyFlag, *footerFlag, false, mode)

	if *dryRun {
		fmt.Println()
		fmt.Println("────────────────────────────────────────────────")
		fmt.Println(i18n.T("preview"))
		fmt.Println("────────────────────────────────────────────────")
		fmt.Println(msg)
		return
	}

	// ===== commit =====
	if err := git.CommitWithMessage(msg, git.Options{
		AllowEmpty: *allowEmpty,
		Amend:      *amend,
		NoVerify:   *noVerify,
		Signoff:    *signoff,
	}); err != nil {
		log.Fatalf("git commit failed: %v", err)
	}
	if notice := gommitupdate.Notice(version); notice != "" {
		fmt.Println(notice)
	}
}

// --- wizard / montagem ---------------------------------------------------

func buildOrPromptMessage(typeFlag, scopeFlag, subjectFlag, bodyFlag, footerFlag string, quiet bool, mode string) string {
	// TYPE
	var selected ui.CommitType
	if typeFlag == "" {
		var err error
		selected, err = ui.SelectCommitType()
		if err != nil {
			log.Fatalf("error selecting commit type: %v", err)
		}
	} else {
		if t, ok := ui.FindType(typeFlag); ok {
			selected = t
		} else {
			log.Fatalf("invalid --type '%s'", typeFlag)
		}
	}
	// SCOPE
	scope := scopeFlag
	if scope == "" && !quiet && mode == "full" {
		scope = promptLine(i18n.T("prompt.scope"))
	}

	// SUBJECT
	subject := subjectFlag
	if subject == "" {
		label := i18n.T("prompt.subject")
		if mode == "simple" {
			label = i18n.T("prompt.description")
		}
		subject = promptLine(label)
	}
	subject = strings.TrimSpace(subject)
	if subject == "" {
		log.Fatal("subject is required")
	}
	if len([]rune(subject)) > 72 {
		log.Fatal("subject must be <= 72 chars")
	}

	// BODY
	body := unescapeNewlines(bodyFlag)
	if body == "" && !quiet && mode == "full" {
		fmt.Println(i18n.T("prompt.body"))
		body = readMultiline()
	}

	// Breaking changes?
	breaking := false
	breakingDesc := ""
	if !quiet && mode == "full" {
		breaking = promptYesNo(i18n.T("prompt.breaking"), false)
		if breaking {
			breakingDesc = promptLine(i18n.T("prompt.breaking_desc"))
		}
	}

	// Issues (Closes / Refs)?
	var closes, refs []string
	if !quiet && mode == "full" && promptYesNo(i18n.T("prompt.issues"), false) {
		c := promptLine(i18n.T("prompt.closes"))
		r := promptLine(i18n.T("prompt.refs"))
		closes = splitCSVNums(c)
		refs = splitCSVNums(r)
	}
	issueLines := buildIssueFooter(closes, refs)

	// FOOTER manual (se passado por flag)
	footer := strings.TrimSpace(unescapeNewlines(footerFlag))

	// Montagem final
	headerType := selected.Key
	if breaking && !strings.HasSuffix(headerType, "!") {
		headerType += "!"
	}

	pre := headerType
	if scope != "" {
		pre += "(" + strings.TrimSpace(scope) + ")"
	}
	if ico := ui.EmojiFor(selected.Key); ico != "" {
		pre += ": " + ico
	}
	header := pre + " " + subject

	var b strings.Builder
	b.WriteString(header)

	if body != "" {
		b.WriteString("\n\n")
		b.WriteString(strings.TrimRight(body, "\n"))
	}

	if breaking && strings.TrimSpace(breakingDesc) != "" {
		b.WriteString("\n\n")
		b.WriteString("BREAKING CHANGE: ")
		b.WriteString(strings.TrimSpace(breakingDesc))
	}

	if footer != "" {
		b.WriteString("\n\n")
		b.WriteString(footer)
	}

	if issueLines != "" {
		b.WriteString("\n\n")
		b.WriteString(issueLines)
	}

	// valida
	typeForValidation := selected.Key
	if err := (commit.Message{
		Type:    typeForValidation,
		Scope:   scope,
		Subject: subject,
		Body:    body,
		Footer:  strings.TrimSpace(strings.Join([]string{footer, issueLines}, "\n\n")),
	}).Validate(); err != nil {
		log.Fatalf("invalid commit message: %v", err)
	}

	return b.String()
}

func promptLine(label string) string {
	fmt.Print(label)
	r := bufio.NewReader(os.Stdin)
	s, _ := r.ReadString('\n')
	return strings.TrimSpace(s)
}

func readMultiline() string {
	r := bufio.NewScanner(os.Stdin)
	r.Buffer(make([]byte, 0, 64*1024), 1_000_000)

	var lines []string
	emptyStreak := 0

	for {
		if !r.Scan() {
			break // EOF
		}
		line := r.Text()
		trim := strings.TrimSpace(line)

		if trim == "." {
			break // ponto sozinho
		}
		if trim == "" {
			emptyStreak++
			if emptyStreak >= 2 {
				break // enter duas vezes
			}
		} else {
			emptyStreak = 0
		}
		lines = append(lines, line)
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
	r := bufio.NewReader(os.Stdin)
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
