// SPDX-License-Identifier: GPL-3.0-only

package commit

import (
	"errors"
	"strings"
	"testing"
)

func TestMessageValidate(t *testing.T) {
	tests := []struct {
		name string
		msg  Message
		want error
	}{
		{"empty", Message{Type: "fix"}, ErrEmptySubject},
		{"whitespace only", Message{Type: "fix", Subject: " \t\n "}, ErrEmptySubject},
		{"valid", Message{Type: "feat", Subject: "add tests"}, nil},
		{"72 characters", Message{Type: "fix", Subject: strings.Repeat("a", 72)}, nil},
		{"73 characters", Message{Type: "fix", Subject: strings.Repeat("a", 73)}, ErrSubjectTooLong},
		{"trim before counting", Message{Type: "fix", Subject: " \t" + strings.Repeat("a", 72) + "\n"}, nil},
		{"72 Unicode characters", Message{Type: "fix", Subject: strings.Repeat("界", 72)}, nil},
		{"73 Unicode characters", Message{Type: "fix", Subject: strings.Repeat("🚀", 73)}, ErrSubjectTooLong},
		{"breaking marker required", Message{Type: "feat!", Subject: "change API"}, ErrBreakingMissing},
		{"breaking scoped header", Message{Type: "feat!", Scope: "cli", Subject: "change API"}, ErrBreakingMissing},
		{"marker in body", Message{Type: "feat!", Subject: "change API", Body: "BREAKING CHANGE: remove flag"}, nil},
		{"marker in footer", Message{Type: "feat!", Subject: "change API", Footer: "BREAKING CHANGE: remove flag"}, nil},
		{"wrong marker case", Message{Type: "feat!", Subject: "change API", Body: "breaking change: remove flag"}, ErrBreakingMissing},
		{"subject error takes precedence", Message{Type: "feat!"}, ErrEmptySubject},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.msg.Validate(); !errors.Is(err, tt.want) {
				t.Fatalf("Validate() = %v, want %v", err, tt.want)
			}
		})
	}
}
