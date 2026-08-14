package ui

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/Hangell/gommit/internal/i18n"
	"github.com/Hangell/gommit/platform"
	"golang.org/x/term"
)

const (
	IconWIP      = "🚧"
	IconPrune    = "🔥"
	IconFeat     = "💡"
	IconFix      = "🐛"
	IconDocs     = "📝"
	IconStyle    = "💅"
	IconRefactor = "🎨"
	IconPerf     = "⚡"
	IconTest     = "✅"
	IconBuild    = "🔧"
	IconCI       = "🤖"
	IconChore    = "📦"
	IconRevert   = "⏪"
)

const (
	FAWIP      = "[WIP]"
	FAPrune    = "[-]"
	FAFeat     = "[+]"
	FAFix      = "[fix]"
	FADocs     = "[doc]"
	FAStyle    = "[fmt]"
	FARefactor = "[ref]"
	FAPerf     = "[perf]"
	FATest     = "[test]"
	FABuild    = "[build]"
	FACI       = "[ci]"
	FAChore    = "[chore]"
	FARevert   = "[revert]"
)

var (
	useColor = func() bool {
		return os.Getenv("NO_COLOR") == ""
	}

	clrDim = func(s string) string {
		if !useColor() {
			return s
		}
		return "\x1b[2m" + s + "\x1b[0m"
	}
	clrType = func(s string) string {
		if !useColor() {
			return s
		}
		return "\x1b[36m" + s + "\x1b[0m"
	} // cyan
	clrIcon = func(s string) string {
		if !useColor() {
			return s
		}
		return "\x1b[33m" + s + "\x1b[0m"
	} // yellow
	clrError = func(s string) string {
		if !useColor() {
			return s
		}
		return "\x1b[31m" + s + "\x1b[0m"
	} // red
)

type CommitType struct {
	Key         string
	Icon        string
	Description string
}

var commitTypes = []CommitType{
	{"WIP", icon(IconWIP, FAWIP), "Work in progress"},
	{"feat", icon(IconFeat, FAFeat), "A new feature"},
	{"fix", icon(IconFix, FAFix), "Fixing a bug"},
	{"chore", icon(IconChore, FAChore), "Updating Dependencies, Deployments, Configuration files"},
	{"refactor", icon(IconRefactor, FARefactor), "Improving structure / format of the code"},
	{"prune", icon(IconPrune, FAPrune), "Removing code or files"},
	{"docs", icon(IconDocs, FADocs), "Writing documentation"},
	{"perf", icon(IconPerf, FAPerf), "Improving performance"},
	{"test", icon(IconTest, FATest), "Adding tests"},
	{"build", icon(IconBuild, FABuild), "Changes to build system or dependencies"},
	{"ci", icon(IconCI, FACI), "Changes to CI/CD configuration"},
	{"style", icon(IconStyle, FAStyle), "Changes that do not affect the meaning of the code"},
	{"revert", icon(IconRevert, FARevert), "Revert to a commit"},
}

func SelectCommitType() (CommitType, error) {
	if term.IsTerminal(int(os.Stdin.Fd())) {
		return selectCommitTypeInteractive()
	}

	// Se stdin não é TTY (pipe), ainda vamos tentar ler uma linha.
	reader := bufio.NewReader(os.Stdin)

	for {
		clearScreen()
		displayCommitTypes()

		fmt.Print("\n" + i18n.T("menu.prompt"))
		input, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				return CommitType{}, fmt.Errorf("selection aborted (EOF)")
			}
			return CommitType{}, fmt.Errorf("error reading input: %w", err)
		}

		input = strings.TrimSpace(input)

		switch strings.ToLower(input) {
		case "q", "quit", "exit":
			return CommitType{}, fmt.Errorf("selection aborted by user")
		}

		if num, err := strconv.Atoi(input); err == nil && num > 0 && num <= len(commitTypes) {
			return commitTypes[num-1], nil
		}

		for _, ct := range commitTypes {
			if strings.EqualFold(input, ct.Key) {
				return ct, nil
			}
		}

		var matches []CommitType
		inputLower := strings.ToLower(input)
		for _, ct := range commitTypes {
			if strings.Contains(strings.ToLower(ct.Key), inputLower) ||
				strings.Contains(strings.ToLower(i18n.T("type."+ct.Key)), inputLower) {
				matches = append(matches, ct)
			}
		}

		if len(matches) == 1 {
			return matches[0], nil
		}

		if len(matches) > 1 {
			fmt.Printf("\n%s\n", clrDim(i18n.T("menu.multiple")))
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			for i, m := range matches {
				fmt.Fprintf(w, "  %2d)\t%s\t%s\t- %s\n", i+1, clrIcon(m.Icon), clrType(m.Key), i18n.T("type."+m.Key))
			}
			w.Flush()
			fmt.Print(clrDim(i18n.T("menu.continue")))
			reader.ReadString('\n')
			continue
		}

		fmt.Printf("\n%s '%s'. %s\n", clrError(i18n.T("menu.invalid")), input, clrDim(i18n.T("menu.retry")))
		fmt.Print(clrDim(i18n.T("menu.continue")))
		reader.ReadString('\n')
	}
}

func selectCommitTypeInteractive() (CommitType, error) {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return CommitType{}, fmt.Errorf("could not enable keyboard navigation: %w", err)
	}
	defer term.Restore(fd, oldState)

	selected := 0
	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h")
	renderInteractiveTypes(selected, false)

	one := make([]byte, 1)
	for {
		if _, err := os.Stdin.Read(one); err != nil {
			return CommitType{}, fmt.Errorf("error reading key: %w", err)
		}
		switch one[0] {
		case '\r', '\n':
			fmt.Print("\r\n")
			return commitTypes[selected], nil
		case 'q', 'Q', 3:
			fmt.Print("\r\n")
			return CommitType{}, fmt.Errorf("selection aborted by user")
		case 27:
			seq := make([]byte, 2)
			if _, err := io.ReadFull(os.Stdin, seq); err == nil && seq[0] == '[' {
				if seq[1] == 'A' {
					selected = (selected - 1 + len(commitTypes)) % len(commitTypes)
				} else if seq[1] == 'B' {
					selected = (selected + 1) % len(commitTypes)
				}
				renderInteractiveTypes(selected, true)
			}
		case 0, 224:
			if _, err := os.Stdin.Read(one); err == nil {
				if one[0] == 72 {
					selected = (selected - 1 + len(commitTypes)) % len(commitTypes)
				} else if one[0] == 80 {
					selected = (selected + 1) % len(commitTypes)
				}
				renderInteractiveTypes(selected, true)
			}
		}
	}
}

func renderInteractiveTypes(selected int, redraw bool) {
	if redraw {
		fmt.Printf("\033[%dA", len(commitTypes)+3)
	}
	fmt.Printf("%s\033[K\r\n\r\n", clrDim(i18n.T("menu.interactive")))
	for i, ct := range commitTypes {
		cursor := "  "
		if i == selected {
			cursor = "> "
		}
		fmt.Printf("%s%s %-10s %s\033[K\r\n", cursor, clrIcon(ct.Icon), clrType(ct.Key), i18n.T("type."+ct.Key))
	}
	fmt.Print("\033[K")
}

func displayCommitTypes() {
	title := i18n.T("menu.title")
	fmt.Println(clrDim(title))
	fmt.Println()

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	for i, ct := range commitTypes {
		fmt.Fprintf(w, "  %2d)\t%s\t%-10s\t%s\n", i+1, clrIcon(ct.Icon), clrType(ct.Key), i18n.T("type."+ct.Key))
	}
	w.Flush()

	fmt.Println()
	fmt.Println(clrDim(i18n.T("menu.hint")))
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func icon(emoji, fallback string) string {
	if platform.SupportsUnicode() {
		return emoji
	}
	return fallback
}
