# Repository commit profiles

Gommit supports three message formats:

| Format | Example |
| --- | --- |
| `plain` | `Update parser behavior` |
| `conventional` | `fix(parser): handle empty input` |
| `emoji` | `fix(parser): 🐛 handle empty input` |

Without configuration, interactive use retains emoji headers. `--format auto`,
`--non-interactive`, and `--json` discover the repository convention. Detection
reads at most 50 non-merge commit subjects; it requires at least five samples
and 80% agreement. Ambiguous history falls back to plain messages. Observed types
are suggestions, not a whitelist. Detection does not generate descriptions or
execute repository JavaScript configurations.

## Plain Git messages

```sh
gommit -m "Update parser behavior"
gommit --plain --subject "Update parser behavior"
gommit -F commit-message.txt
gommit --non-interactive -F - < commit-message.txt
```

`-m`/`--message` and `-F`/`--file` preserve a supplied message without adding
prefixes, emojis, or a title length limit. They cannot be combined with structured
fields. Git still applies its message cleanup and hooks. To avoid automatic
staging in interactive use, pass `--auto-stage=false`.

## Shared configuration

```sh
gommit init --format conventional
gommit doctor
gommit --detect --json
```

`init` creates `.gommit.json` at the current worktree root and refuses to overwrite
an existing file. Review and commit this file to share the policy. For example:

```json
{
  "format": "conventional",
  "emoji": false,
  "types": ["feat", "fix", "docs", "task"],
  "subject_limit": 72,
  "scope_required": true
}
```

All fields are optional. Formats are `plain`, `conventional`, `emoji`, and `auto`.
Unknown fields and invalid values are rejected. A positive `subject_limit` counts
Unicode code points in the structured subject, excluding type, scope and emoji.
Zero or omission uses 72 for typed messages and no limit for plain messages.
`scope_required` applies to typed messages. Raw `-m`/`-F` messages intentionally
bypass these structured-message checks; repository hooks remain authoritative.

Selection order: explicit format flags, merged Git `gommit.format`/`gommit.emoji`
settings, `.gommit.json`, declarative commitlint JSON, recent history. Git uses
its normal configuration scope precedence, so local settings override global
settings. A `.gommit.json` file takes precedence over commitlint discovery.
Supported commitlint inputs are `.commitlintrc.json` and `commitlint.config.json`,
with `@commitlint/config-conventional` or an enabled `type-enum` `always` rule.
Other commitlint rules are not imported; keep the project's hooks enabled.

```sh
git config --local gommit.format conventional
git config --local gommit.emoji false
gommit --format conventional --type task --scope cli --subject "Improve parser"
```

`--no-emoji` and `NO_EMOJI=1` disable generated header emojis. They do not strip
characters from supplied subjects or raw messages. Breaking changes use
`type(scope)!: subject` with a `BREAKING CHANGE:` footer.

## Automation and AI agents

```sh
gommit doctor --json
gommit --json --format auto --type fix --scope parser \
  --subject "Handle empty input" --dry-run
git add parser.go
gommit --json --format conventional --type fix --scope parser \
  --subject "Handle empty input"
```

`--json` implies `--non-interactive`: no prompts, no update notices, and no
automatic staging unless `--auto-stage` is explicit. Structured typed messages
require `--type` and `--subject`; plain messages require a subject or raw message.
The program does not infer the meaning of a change. When history is ambiguous,
inspect the profile before choosing a format.

`--dry-run` never stages files, writes editor files, or commits, including with
`--auto-stage`. Validation failures occur before staging. An existing staged
selection is preserved; automatic staging only applies when the index is empty.
Git hooks and signing settings remain active unless explicitly overridden.

JSON commit results contain `ok`, `message`, `profile`, and `committed`; previews
also contain `dry_run`. Errors return a nonzero exit status and an object with
`ok: false` and `error`. Git command output goes to stderr in JSON mode. Diagnostic
output is the profile itself, including format, source, confidence and sample size.
`doctor` without JSON also reports whether changes are staged. It does not certify
compatibility with every repository hook.
