package i18n

import "testing"

func TestSupportedLanguages(t *testing.T) {
	for _, language := range []string{"en", "es", "pt-BR", "hi", "ru_RU", "zh-CN"} {
		if Normalize(language) == "" {
			t.Errorf("expected %q to be supported", language)
		}
	}
}

func TestEveryLocaleHasEveryEnglishKey(t *testing.T) {
	for language, messages := range locales {
		for key := range english {
			if messages[key] == "" {
				t.Errorf("locale %s is missing %s", language, key)
			}
		}
	}
}

func TestFallbackAndFormatting(t *testing.T) {
	Set("pt")
	if got := T("update.current", "1.2.3"); got == "" || got == "update.current" {
		t.Fatalf("unexpected translation: %q", got)
	}
	if got := T("unknown.key"); got != "unknown.key" {
		t.Fatalf("unexpected missing-key fallback: %q", got)
	}
	Set("en")
}
