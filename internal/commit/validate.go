package commit

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidSubject  = errors.New("subject must be a single line")
	ErrEmptySubject    = errors.New("subject is required")
	ErrSubjectTooLong  = errors.New("subject must be <= 72 chars")
	ErrBreakingMissing = errors.New("header has '!' but missing 'BREAKING CHANGE:' in body or footer")
)

// Validate checa tamanho do subject e, se houver '!' no header, exige BREAKING CHANGE.
func (m Message) Validate() error {
	sub := strings.TrimSpace(m.Subject)
	if sub == "" {
		return ErrEmptySubject
	}
	if strings.ContainsAny(sub, "\r\n") {
		return ErrInvalidSubject
	}
	limit := m.SubjectLimit
	if limit == 0 && !m.Plain {
		limit = 72
	}
	if limit > 0 && runeCount(sub) > limit {
		if limit == 72 {
			return ErrSubjectTooLong
		}
		return fmt.Errorf("subject must be <= %d chars", limit)
	}

	if m.Breaking || strings.Contains(m.Type, "!") || strings.Contains(m.headerPrefix(), "!:") {
		if !containsBreakingMarker(m.Body) && !containsBreakingMarker(m.Footer) {
			return ErrBreakingMissing
		}
	}
	return nil
}

func (m Message) headerPrefix() string {
	t := m.Type
	if m.Scope != "" {
		t += "(" + strings.TrimSpace(m.Scope) + ")"
	}
	return t
}

func containsBreakingMarker(s string) bool {
	return strings.Contains(s, "BREAKING CHANGE:")
}

func runeCount(s string) int {
	return len([]rune(s))
}
