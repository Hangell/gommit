package ui

import "strings"

func Types() []CommitType {
	return commitTypes
}

func FindType(key string) (CommitType, bool) {
	for _, ct := range commitTypes {
		if equalFold(ct.Key, key) {
			return ct, true
		}
	}
	return CommitType{}, false
}

func EmojiFor(key string) string {
	switch strings.ToLower(key) {
	case "wip":
		return IconWIP
	case "feat":
		return IconFeat
	case "fix":
		return IconFix
	case "chore":
		return IconChore
	case "refactor":
		return IconRefactor
	case "prune":
		return IconPrune
	case "docs":
		return IconDocs
	case "perf":
		return IconPerf
	case "test":
		return IconTest
	case "build":
		return IconBuild
	case "ci":
		return IconCI
	case "style":
		return IconStyle
	case "revert":
		return IconRevert
	default:
		return ""
	}
}

func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return toLower(a) == toLower(b)
	}
	return toLower(a) == toLower(b)
}

func toLower(s string) string {
	return strings.ToLower(s)
}
