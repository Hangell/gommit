package commit

import (
	"strings"
)

type Message struct {
	Type    string
	Scope   string
	Subject string
	Body    string
	Footer  string
}

func (m Message) Build() string {
	header := m.Type
	if m.Scope != "" {
		header += "(" + strings.TrimSpace(m.Scope) + ")"
	}
	header += ": " + strings.TrimSpace(m.Subject)

	var b strings.Builder
	b.WriteString(header)
	if m.Body != "" {
		b.WriteString("\n\n")
		b.WriteString(strings.TrimRight(m.Body, "\n"))
	}
	if m.Footer != "" {
		b.WriteString("\n\n")
		b.WriteString(strings.TrimRight(m.Footer, "\n"))
	}
	return b.String()
}
