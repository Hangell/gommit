// SPDX-License-Identifier: GPL-3.0-only

package commit

import (
	"strings"
	"testing"
)

func TestMessageBuild(t *testing.T) {
	tests := []struct {
		name string
		msg  Message
		want string
	}{
		{"header", Message{Type: "feat", Subject: "add tests"}, "feat: add tests"},
		{"trim scope and subject", Message{Type: "fix", Scope: " cli ", Subject: " fix prompt \n"}, "fix(cli): fix prompt"},
		{"body and footer", Message{Type: "feat", Subject: "add tests", Body: "First line\nSecond line\n\n", Footer: "Closes #42\n"}, "feat: add tests\n\nFirst line\nSecond line\n\nCloses #42"},
		{"footer without body", Message{Type: "docs", Subject: "update guide", Footer: "Refs #12"}, "docs: update guide\n\nRefs #12"},
		{"unicode", Message{Type: "fix", Scope: "界面", Subject: "corrigir tradução 🚀"}, "fix(界面): corrigir tradução 🚀"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.msg.Build(); got != tt.want {
				t.Fatalf("Build() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBreakingScopeAndEmoji(t *testing.T) {
	m := Message{Type: "feat", Scope: "api", Subject: "change", Emoji: "💡", Breaking: true, Footer: "BREAKING CHANGE: new API"}
	if err := m.Validate(); err != nil {
		t.Fatal(err)
	}
	if got := m.Build(); got != "feat(api)!: 💡 change\n\nBREAKING CHANGE: new API" {
		t.Fatal(got)
	}
}
func TestConfigurableSubjectLimit(t *testing.T) {
	if err := (Message{Plain: true, Subject: strings.Repeat("x", 100)}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (Message{Type: "fix", Subject: "123456", SubjectLimit: 5}).Validate(); err == nil {
		t.Fatal("limit ignored")
	}
	if err := (Message{Plain: true, Subject: "two\nlines"}).Validate(); err == nil {
		t.Fatal("multiline subject accepted")
	}
}
