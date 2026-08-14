package i18n

import (
	"fmt"
	"os"
	"strings"
)

var (
	current = "en"
	locales = map[string]map[string]string{
		"en": english,
		"es": spanish,
		"pt": portuguese,
		"hi": hindi,
		"ru": russian,
		"zh": chinese,
	}
)

func Normalize(language string) string {
	language = strings.ToLower(strings.TrimSpace(language))
	if i := strings.IndexAny(language, "_-"); i >= 0 {
		language = language[:i]
	}
	switch language {
	case "en", "english":
		return "en"
	case "es", "spanish", "español", "espanol":
		return "es"
	case "pt", "portuguese", "português", "portugues":
		return "pt"
	case "hi", "hindi":
		return "hi"
	case "ru", "russian", "русский":
		return "ru"
	case "zh", "chinese", "中文":
		return "zh"
	default:
		return ""
	}
}

func Set(language string) bool {
	language = Normalize(language)
	if language == "" {
		return false
	}
	current = language
	return true
}

func Language() string  { return current }
func Supported() string { return "en, es, pt, hi, ru, zh" }

func SystemLanguage() string {
	for _, value := range []string{systemLocale(), os.Getenv("LC_ALL"), os.Getenv("LC_MESSAGES"), os.Getenv("LANG")} {
		if language := Normalize(value); language != "" {
			return language
		}
	}
	return "en"
}

func T(key string, args ...any) string {
	text, ok := locales[current][key]
	if !ok {
		text = english[key]
	}
	if text == "" {
		text = key
	}
	if len(args) > 0 {
		return fmt.Sprintf(text, args...)
	}
	return text
}
