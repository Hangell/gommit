// SPDX-License-Identifier: GPL-3.0-only
package profile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInfer(t *testing.T) {
	for _, tt := range []struct{ subject, format string }{
		{"fix(cli): correct output", Conventional}, {"feat: 💡 add option", Emoji}, {"feat 💡: add option", Emoji}, {"Update documentation", Plain}, {"fix: support unicode 中文", Conventional}, {"fix: support emoji later 💡", Conventional},
	} {
		t.Run(tt.subject, func(t *testing.T) {
			p := Infer([]string{tt.subject, tt.subject, tt.subject, tt.subject, "misc"})
			if p.Format != tt.format || p.Confidence != .8 && tt.format != Plain {
				t.Fatalf("%+v", p)
			}
		})
	}
	if p := Infer([]string{"feat: one", "feat: two"}); p.Source != "fallback" {
		t.Fatal(p)
	}
	if p := Infer([]string{"feat: one", "fix: two", "plain", "plain", "another"}); p.Source != "fallback" {
		t.Fatal(p)
	}
}
func TestConfigValidation(t *testing.T) {
	for _, data := range []string{`null`, `{"format":"unknown"}`, `{"unknown":1}`, `{"types":["bad type"]}`, `{"subject_limit":-1}`, `{} {}`, "{}" + strings.Repeat(" ", 1<<20)} {
		path := filepath.Join(t.TempDir(), ".gommit.json")
		if err := os.WriteFile(path, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := ReadConfig(path); err == nil {
			t.Fatalf("accepted %q", data[:min(len(data), 80)])
		}
	}
}
func TestDeclarativeCommitlint(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".commitlintrc.json")
	if err := os.WriteFile(path, []byte(`{"extends":["@commitlint/config-conventional"],"rules":{"type-enum":[2,"always",["task","fix"]]}}`), 0600); err != nil {
		t.Fatal(err)
	}
	p, ok, err := commitlintProfile(root)
	if err != nil || !ok || p.Format != Conventional || len(p.AllowedTypes) != 2 {
		t.Fatalf("%+v %v %v", p, ok, err)
	}
}
func FuzzInfer(f *testing.F) {
	for _, s := range []string{"feat: 💡 add", "fix(cli)!: update", "", "🔥: x", "\xff"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		p := Infer([]string{s, s, s, s, s})
		if !ValidFormat(p.Format) || p.Confidence < 0 || p.Confidence > 1 {
			t.Fatalf("invalid inference %+v", p)
		}
	})
}
