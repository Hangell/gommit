//go:build !windows
// +build !windows

package platform

import (
	"os"
	"strings"
)

// SupportsUnicode informa se o terminal deve renderizar UTF-8/emoji.
// Mantemos heurística simples e rápida: respeita NO_EMOJI, ignora TERM=dumb
// e confia na locale UTF-8 (LANG/LC_ALL). Em Unix moderno isso costuma bastar.
func SupportsUnicode() bool {
	// Opt-out explícito
	if os.Getenv("NO_EMOJI") == "1" {
		return false
	}

	// Terminais "burros"
	if os.Getenv("TERM") == "dumb" {
		return false
	}

	// Locale em UTF-8? (ex.: "pt_BR.UTF-8", "C.UTF-8")
	if isUTF8Locale(os.Getenv("LC_ALL")) ||
		isUTF8Locale(os.Getenv("LC_CTYPE")) ||
		isUTF8Locale(os.Getenv("LANG")) {
		return true
	}

	// Fallback otimista: Unix/macOS geralmente são UTF-8 hoje
	return true
}

func isUTF8Locale(s string) bool {
	s = strings.ToLower(s)
	return strings.Contains(s, "utf-8") || strings.Contains(s, "utf8")
}
