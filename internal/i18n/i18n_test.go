package i18n

import (
	"strings"
	"testing"
)

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
	previous := Language()
	t.Cleanup(func() { Set(previous) })
	Set("pt")
	if got := T("update.current", "1.2.3"); got == "" || got == "update.current" {
		t.Fatalf("unexpected translation: %q", got)
	}
	if got := T("unknown.key"); got != "unknown.key" {
		t.Fatalf("unexpected missing-key fallback: %q", got)
	}
	Set("en")
}

func TestLocalizedFormatArguments(t *testing.T) {
	previous := Language()
	t.Cleanup(func() { Set(previous) })
	for language := range locales {
		Set(language)
		for key, args := range map[string][]any{
			"language.saved":    {"pt"},
			"language.invalid":  {"invalid-language", "en, pt"},
			"mode.saved":        {"full"},
			"mode.invalid":      {"invalid-mode"},
			"amend.last":        {"previous subject"},
			"update.notice":     {"2.0.0", "1.0.0"},
			"update.current":    {"1.0.0"},
			"update.downloaded": {"2.0.0"},
		} {
			got := T(key, args...)
			if strings.Contains(got, "%!") {
				t.Errorf("%s/%s has invalid format placeholders: %s", language, key, got)
			}
			for _, arg := range args {
				if !strings.Contains(got, arg.(string)) {
					t.Errorf("%s/%s omitted argument %q: %s", language, key, arg, got)
				}
			}
		}
	}
}
