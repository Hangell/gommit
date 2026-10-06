// SPDX-License-Identifier: GPL-3.0-only

package ui

import (
	"os"
	"strings"
	"testing"

	"github.com/Hangell/gommit/internal/i18n"
)

func TestSelectCommitTypeFromPipe(t *testing.T) {
	previousLanguage := i18n.Language()
	t.Cleanup(func() { i18n.Set(previousLanguage) })
	i18n.Set("en")
	t.Setenv("NO_COLOR", "1")
	for _, tt := range []struct {
		name, input, want string
		fail              bool
	}{
		{"number", "2\n", "feat", false},
		{"case insensitive key", " FIX \n", "fix", false},
		{"unique partial match", "typ\n", "typo", false},
		{"invalid then valid", "not-a-type\n\nfeat\n", "feat", false},
		{"ambiguous then valid", "f\n\nfeat\n", "feat", false},
		{"quit", "q\n", "", true},
		{"EOF", "", "", true},
		{"no final newline", "feat", "feat", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			input, err := os.CreateTemp(t.TempDir(), "stdin")
			if err != nil {
				t.Fatal(err)
			}
			defer input.Close()
			if _, err := input.WriteString(tt.input); err != nil {
				t.Fatal(err)
			}
			if _, err := input.Seek(0, 0); err != nil {
				t.Fatal(err)
			}
			output, err := os.CreateTemp(t.TempDir(), "stdout")
			if err != nil {
				t.Fatal(err)
			}
			defer output.Close()
			stdin, stdout := os.Stdin, os.Stdout
			os.Stdin, os.Stdout = input, output
			t.Cleanup(func() { os.Stdin, os.Stdout = stdin, stdout })
			got, err := SelectCommitType()
			if (err != nil) != tt.fail || got.Key != tt.want {
				t.Fatalf("selection = %q, %v; want %q, failure=%v", got.Key, err, tt.want, tt.fail)
			}
		})
	}
}

func TestCommitTypeCatalog(t *testing.T) {
	seen := make(map[string]bool)
	for _, ct := range Types() {
		key := strings.ToLower(ct.Key)
		if seen[key] {
			t.Fatalf("duplicate commit type: %s", ct.Key)
		}
		seen[key] = true
		if EmojiFor(ct.Key) == "" {
			t.Errorf("missing header icon for %s", ct.Key)
		}
		if got, ok := FindType(strings.ToUpper(ct.Key)); !ok || got.Key != ct.Key {
			t.Errorf("case insensitive lookup failed for %s", ct.Key)
		}
	}
	if _, ok := FindType("unknown"); ok || EmojiFor("unknown") != "" {
		t.Fatal("unknown type accepted")
	}
}
