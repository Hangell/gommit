// SPDX-License-Identifier: GPL-3.0-only

// Package profile discovers commit conventions without executing repository code.
package profile

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/Hangell/gommit/internal/git"
)

const (
	Plain        = "plain"
	Conventional = "conventional"
	Emoji        = "emoji"
	Auto         = "auto"
)

var typePattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9-]*$`)
var headerPattern = regexp.MustCompile(`^([a-zA-Z][a-zA-Z0-9-]*)(?:\([^()\r\n]+\))?!?:\s+\S`)

type Profile struct {
	Format        string   `json:"format"`
	Emoji         bool     `json:"emoji"`
	Source        string   `json:"source"`
	Confidence    float64  `json:"confidence"`
	SampleSize    int      `json:"sample_size"`
	ObservedTypes []string `json:"observed_types"`
	AllowedTypes  []string `json:"allowed_types,omitempty"`
	SubjectLimit  int      `json:"subject_limit,omitempty"`
	ScopeRequired bool     `json:"scope_required,omitempty"`
	Explicit      bool     `json:"-"`
}

type Config struct {
	Format        string   `json:"format"`
	Emoji         *bool    `json:"emoji,omitempty"`
	SubjectLimit  int      `json:"subject_limit,omitempty"`
	ScopeRequired bool     `json:"scope_required,omitempty"`
	Types         []string `json:"types,omitempty"`
}

func ValidFormat(format string) bool {
	return format == Plain || format == Conventional || format == Emoji || format == Auto
}

func ValidType(kind string) bool { return typePattern.MatchString(kind) }

// Infer uses a bounded sample and requires at least five subjects and 80% agreement.
// Ambiguous and empty histories use plain messages rather than inventing a convention.
func Infer(subjects []string) Profile {
	p := Profile{Format: Plain, Source: "fallback", ObservedTypes: []string{}}
	counts := map[string]int{}
	types := map[string]bool{}
	for _, subject := range subjects {
		if strings.TrimSpace(subject) == "" {
			continue
		}
		p.SampleSize++
		clean := strings.TrimSpace(strings.Map(func(r rune) rune {
			if emojiRune(r) || r == '\ufe0f' || r == '\u200d' {
				return -1
			}
			return r
		}, subject))
		// Older gommit headers put the emoji between the type and colon.
		clean = strings.ReplaceAll(clean, " :", ":")
		match := headerPattern.FindStringSubmatch(clean)
		if match == nil {
			counts[Plain]++
			continue
		}
		types[strings.ToLower(match[1])] = true
		format := Conventional
		// Only consider the header decoration, not emojis in the description.
		colon := strings.Index(subject, ":")
		if hasEmoji(subject[:colon+1]) || hasLeadingEmoji(strings.TrimSpace(subject[colon+1:])) {
			format = Emoji
		}
		counts[format]++
	}
	for kind := range types {
		p.ObservedTypes = append(p.ObservedTypes, kind)
	}
	sort.Strings(p.ObservedTypes)
	if p.SampleSize < 5 {
		return p
	}
	for _, format := range []string{Plain, Conventional, Emoji} {
		ratio := float64(counts[format]) / float64(p.SampleSize)
		if ratio >= 0.8 {
			p.Format, p.Emoji, p.Source, p.Confidence = format, format == Emoji, "history", ratio
			break
		}
	}
	return p
}

func emojiRune(r rune) bool {
	return (r >= 0x1f000 && r <= 0x1faff) || (r >= 0x2600 && r <= 0x27ff) || r == 0x23e9 || r == 0x23ea
}

func hasEmoji(s string) bool {
	for _, r := range s {
		if emojiRune(r) {
			return true
		}
	}
	return false
}

func hasLeadingEmoji(s string) bool {
	for _, r := range s {
		return emojiRune(r)
	}
	return false
}

func ReadConfig(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, (1<<20)+1))
	if err != nil {
		return Config{}, err
	}
	if len(data) > 1<<20 {
		return Config{}, fmt.Errorf("configuration exceeds 1 MiB")
	}
	if !strings.HasPrefix(strings.TrimSpace(string(data)), "{") {
		return Config{}, fmt.Errorf("expected JSON object in %s", path)
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	var c Config
	if err := decoder.Decode(&c); err != nil {
		return c, fmt.Errorf("invalid %s: %w", path, err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return c, fmt.Errorf("invalid %s: expected one JSON object", path)
	}
	if c.Format != "" && !ValidFormat(c.Format) {
		return c, fmt.Errorf("invalid format %q in %s", c.Format, path)
	}
	if c.SubjectLimit < 0 {
		return c, fmt.Errorf("subject_limit must be nonnegative")
	}
	for _, kind := range c.Types {
		if !ValidType(kind) {
			return c, fmt.Errorf("invalid commit type %q in %s", kind, path)
		}
	}
	return c, nil
}

// Discover honors Git's scoped configuration before the shared .gommit.json file,
// then declarative commitlint JSON, and finally recent local commit history.
func Discover() (Profile, error) {
	root, err := git.Root()
	if err != nil {
		return Profile{}, err
	}
	subjects, err := git.RecentSubjects(50)
	if err != nil {
		return Profile{}, err
	}
	p := Infer(subjects)
	configPath := filepath.Join(root, ".gommit.json")
	c, err := ReadConfig(configPath)
	if err != nil && !os.IsNotExist(err) {
		return p, err
	}
	if err == nil {
		p.AllowedTypes = c.Types
		p.SubjectLimit, p.ScopeRequired = c.SubjectLimit, c.ScopeRequired
		p.Explicit = true
		if c.Format != "" && c.Format != Auto {
			p.Format, p.Source, p.Confidence, p.Explicit = c.Format, ".gommit.json", 1, true
			p.Emoji = c.Format == Emoji
		}
		if c.Emoji != nil {
			p.Emoji, p.Explicit, p.Source, p.Confidence = *c.Emoji, true, ".gommit.json", 1
			if p.Format != Plain {
				if *c.Emoji {
					p.Format = Emoji
				} else {
					p.Format = Conventional
				}
			}
		}
	} else {
		if declared, ok, err := commitlintProfile(root); err != nil {
			return p, err
		} else if ok {
			declared.SampleSize, declared.ObservedTypes = p.SampleSize, p.ObservedTypes
			p = declared
		}
	}
	format, err := git.ConfigValue("gommit.format")
	if err != nil {
		return p, err
	}
	if format != "" {
		if !ValidFormat(format) {
			return p, fmt.Errorf("invalid gommit.format %q", format)
		}
		p.Explicit = true
		if format != Auto {
			p.Format, p.Source, p.Confidence, p.Emoji = format, "git-config", 1, format == Emoji
		}
	}
	emoji, err := git.ConfigBool("gommit.emoji")
	if err != nil {
		return p, err
	}
	if emoji != nil {
		p.Emoji, p.Explicit, p.Source, p.Confidence = *emoji, true, "git-config", 1
		if *emoji && p.Format != Plain {
			p.Format = Emoji
		}
		if !*emoji && p.Format == Emoji {
			p.Format = Conventional
		}
	}
	if p.Format == Plain {
		p.Emoji = false
	}
	return p, nil
}

func commitlintProfile(root string) (Profile, bool, error) {
	for _, name := range []string{".commitlintrc.json", "commitlint.config.json"} {
		data, err := os.ReadFile(filepath.Join(root, name))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return Profile{}, false, err
		}
		var c struct {
			Extends json.RawMessage            `json:"extends"`
			Rules   map[string]json.RawMessage `json:"rules"`
		}
		if err := json.Unmarshal(data, &c); err != nil {
			return Profile{}, false, fmt.Errorf("invalid %s: %w", name, err)
		}
		var names []string
		if err := json.Unmarshal(c.Extends, &names); err != nil {
			var single string
			if json.Unmarshal(c.Extends, &single) == nil {
				names = []string{single}
			}
		}
		known := false
		for _, value := range names {
			if value == "@commitlint/config-conventional" {
				known = true
			}
		}
		var types []string
		if raw := c.Rules["type-enum"]; raw != nil {
			var rule []json.RawMessage
			if json.Unmarshal(raw, &rule) != nil || len(rule) == 0 {
				return Profile{}, false, fmt.Errorf("invalid type-enum rule in %s", name)
			}
			var level int
			if json.Unmarshal(rule[0], &level) != nil || level < 0 || level > 2 {
				return Profile{}, false, fmt.Errorf("invalid type-enum severity in %s", name)
			}
			if level > 0 {
				var when string
				if len(rule) != 3 || json.Unmarshal(rule[1], &when) != nil || when != "always" {
					return Profile{}, false, fmt.Errorf("unsupported type-enum rule in %s; use .gommit.json", name)
				}
				if json.Unmarshal(rule[2], &types) != nil {
					return Profile{}, false, fmt.Errorf("invalid type-enum values in %s", name)
				}
				for _, kind := range types {
					if !ValidType(kind) {
						return Profile{}, false, fmt.Errorf("invalid type %q in %s", kind, name)
					}
				}
				known = true
			}
		}

		if known {
			return Profile{Format: Conventional, Source: name, Confidence: 1, Explicit: true, AllowedTypes: types}, true, nil
		}
	}
	return Profile{}, false, nil
}
