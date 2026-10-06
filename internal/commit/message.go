package commit

import (
	"strings"
)

type Message struct {
	Type         string
	Scope        string
	Subject      string
	Body         string
	Footer       string
	Plain        bool
	Emoji        string
	Breaking     bool
	SubjectLimit int
}

func (m Message) Build() string {
	header := strings.TrimSpace(m.Subject)
	if !m.Plain {
		header = strings.TrimSuffix(m.Type, "!")
		if m.Scope != "" {
			header += "(" + strings.TrimSpace(m.Scope) + ")"
		}
		if m.Breaking || strings.HasSuffix(m.Type, "!") {
			header += "!"
		}
		header += ": "
		if m.Emoji != "" {
			header += m.Emoji + " "
		}
		header += strings.TrimSpace(m.Subject)
	}

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
