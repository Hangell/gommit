package ai

import "testing"

func TestParseResponse(t *testing.T) {
	input := []byte(`{"response":"{\"type\":\"fix\",\"subject\":\"handle empty configuration safely\"}"}`)
	got, err := parseResponse(input)
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != "fix" || got.Subject != "handle empty configuration safely" {
		t.Fatalf("unexpected suggestion: %#v", got)
	}
}

func TestParseResponseRejectsLongSubject(t *testing.T) {
	input := []byte(`{"response":"{\"type\":\"feat\",\"subject\":\"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\"}"}`)
	if _, err := parseResponse(input); err == nil {
		t.Fatal("expected long subject to fail")
	}
}

func TestGeminiCommandCanBeResolved(t *testing.T) {
	command, _, err := geminiCommand()
	if err != nil {
		t.Skip("Gemini CLI is not installed on this test machine")
	}
	if command == "" {
		t.Fatal("resolved an empty Gemini command")
	}
}
