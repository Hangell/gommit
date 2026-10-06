# Contributing to gommit

Thank you for helping improve gommit! Contributions to code, tests, translations,
and documentation are welcome.

## Community and scope

Be respectful, constructive, and welcoming in issues and reviews. Search existing
issues before opening a new one. For a substantial change, open an issue first to
discuss the problem, proposed behavior, alternatives, and platform impact.

For bugs, include your gommit version (`gommit --version`), OS and architecture,
Git version, reproduction steps, and expected versus actual behavior. Remove
credentials and private repository information from logs.

## Development setup

You need Git and **Go 1.25 or newer**, as specified in [go.mod](go.mod). Use the
latest patch release of your Go version. No Node.js or npm installation is needed.
GNU Make and a POSIX shell are optional shortcuts for running checks; on Windows,
you can use Git Bash with Make or run the Go commands below directly.

Fork the repository, then clone your fork:

```sh
git clone https://github.com/<your-username>/gommit.git
cd gommit
git remote add upstream https://github.com/Hangell/gommit.git
go mod download
git switch -c feat/your-change
```

Use focused branches such as `feat/your-change`, `fix/your-change`, or
`docs/your-change`, and keep pull requests small enough to review.

## Build and run

```sh
go build -o bin/gommit ./cmd/gommit
# On Windows, use: go build -o bin/gommit.exe ./cmd/gommit
go run ./cmd/gommit --help
go run ./cmd/gommit --version
```

Test the commit wizard in a disposable Git repository. gommit can automatically
stage changes during a real commit. `--dry-run` validates and previews the message
without staging, writing editor files, or creating a commit:

```sh
/path/to/gommit/bin/gommit --dry-run --type feat --subject "add example"
```

For Git editor integration, set the editor only for one command in that temporary
repository (use the absolute path to the binary):

```sh
git -c core.editor='"/absolute/path/to/gommit/bin/gommit" --as-editor' commit
```

## Tests and validators

Run all checks before submitting a pull request:

```sh
make check
```

The individual targets are:

| Target | Purpose |
| --- | --- |
| `make fmt` | Format Go source files (writes changes). |
| `make fmt-check` | Reject unformatted Go source files. |
| `make mod-check` | Check module tidiness without changing files and verify downloaded modules. |
| `make lint` | Run `go vet` and validate GitHub Actions workflows with actionlint. |
| `make test` | Run uncached tests with the race detector and generate `coverage.out`. |
| `make build` | Build all packages for the current platform. |
| `make vuln-check` | Check reachable known vulnerabilities with govulncheck. |

Without Make, run these equivalent commands from the repository root:

```sh
gofmt -l cmd internal platform
# The command above must print no files. To fix formatting:
# gofmt -w cmd internal platform
go mod tidy -diff
go mod verify
go vet ./...
go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12 -shellcheck=""
go test -race -count=1 "-coverprofile=coverage.out" ./...
go build ./...
go run golang.org/x/vuln/cmd/govulncheck@v1.7.0 ./...
go tool cover -func=coverage.out
```

Validator versions are pinned in [Makefile](Makefile). Running them for the first
time requires network access; govulncheck also reads the Go vulnerability database.
The race detector requires a supported platform and C compiler (on Windows, a
compatible MinGW-w64 toolchain). If unavailable locally, run `go test -count=1
./...` and disclose that limitation in your PR; CI runs the race checks.

CI runs formatting, module, vet, workflow, and vulnerability checks, executes tests
on Linux, macOS, and Windows, and cross-compiles Linux/macOS/Windows binaries for
amd64 and arm64. Release builds depend on the quality and test jobs passing.
Linux coverage is available as the `coverage` workflow artifact.

### Writing tests

Add regression tests for bug fixes and tests for new behavior. Cover relevant
failure cases, Unicode input, and platform differences. Prefer table-driven tests
for message validation. Use temporary files and repositories for filesystem and
Git tests; isolate Git configuration and avoid network access, real commits in the
contributor's checkout, and changes to global user settings.

Translation changes should preserve all English keys and formatting placeholders.
Tests that change the active language must restore it before returning.

## Commit messages and sign-off

Use Conventional Commits:

```text
<type>(<optional scope>): <subject>

<optional body>

<optional footer>
```

Common types include `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`,
`build`, `ci`, `chore`, and `revert`. Write a clear imperative subject of at most
72 characters. For breaking changes, add `!` to the header and explain the change
in a `BREAKING CHANGE:` body or footer. Link issues with `Closes #123` or `Refs #123`.

Sign off your commits with `git commit -s` to certify the
[Developer Certificate of Origin](https://developercertificate.org/):

```sh
git commit -s -m "test(commit): cover Unicode subject limits"
```

A DCO sign-off is a `Signed-off-by` trailer, not a cryptographic signature. gommit
also supports `--signoff`.

## Pull requests and review

1. Explain the problem, resulting behavior, and any compatibility impact using the
   PR template. Link the relevant issue.
2. Include tests and update documentation or translations when behavior changes.
3. Run the checks above and describe the validation you performed.
4. Open your PR against `main`; use a draft while work is in progress.
5. Address review feedback and failing CI checks. Maintainers make merge decisions.

### CodeRabbit

[.coderabbit.yaml](.coderabbit.yaml) configures English reviews, automatic reviews
of non-draft PRs, incremental reviews after pushes, and project-specific guidance.
CodeRabbit complements CI and maintainer review. Its GitHub App must be installed
and enabled for `Hangell/gommit` by a repository administrator before reviews can
run; the YAML file alone does not install the app. See the
[official setup guide](https://docs.coderabbit.ai/getting-started/quickstart).

The active `Protect main` ruleset requires pull requests, resolved review
conversations, an up-to-date branch, and the GitHub Actions checks `Quality`,
`Test (ubuntu-latest)`, `Test (macos-latest)`, and `Test (windows-latest)`.
It blocks force pushes and deletion of the default branch (currently `main`).
There are no bypass actors. Human approvals are optional while the repository has
one maintainer; maintainers can require an approval when another reviewer is
available. These repository settings are separate from the checked-in workflow.
Dependabot proposes weekly Go module and GitHub Actions updates.

## Contributor recognition

The README includes contributor avatars from [contrib.rocks](https://contrib.rocks)
and links to the [GitHub contributor list](https://github.com/Hangell/gommit/graphs/contributors).
The image is generated from GitHub's contributor data and may take time to refresh.
Commit-based recognition includes code, tests, documentation, and translations.
Use a commit email associated with your GitHub account so GitHub can attribute your
commits. See [GitHub's contributor documentation](https://docs.github.com/en/repositories/viewing-activity-and-data-for-your-repository/viewing-a-projects-contributors).

To list contributors in a local checkout, run:

```sh
make contributors
# Or, without Make:
git shortlog --group=author --group=trailer:co-authored-by -sn HEAD
```

This lists names and commit counts from the checked-out history, including
`Co-authored-by` trailers. The [.mailmap](.mailmap) file consolidates author aliases;
propose an entry there if your own name appears more than once. For shared work,
include the co-author's name and associated commit email in the commit footer:

```text
Co-authored-by: Contributor Name <contributor@example.com>
```

Credit for co-authorship is separate from each contributor's DCO sign-off.
For contributions through issues, reviews, or other work without commits, describe
the contribution in the related issue or PR so maintainers can acknowledge it.

## Project layout

```text
cmd/gommit/       CLI entry point, flags, and wizard orchestration
internal/commit/  Commit message construction and validation
internal/git/     Git command integration
internal/i18n/    Locales, language selection, and translations
internal/install/ Binary installation for Unix and Windows
internal/ui/      Terminal prompts and commit type selection
internal/update/  Release lookup and binary updates
platform/         Platform-specific console handling
scripts/          Shell and PowerShell installers
.github/          CI, contribution template, and dependency updates
```

## Releases, security, and license

Maintainers publish SemVer tags (`vMAJOR.MINOR.PATCH`). The GitHub Actions release
job packages binaries and publishes checksums for version tags after checks pass.

Report vulnerabilities privately to the maintainer through a private contact
listed on their GitHub profile, or GitHub private vulnerability reporting if it is
enabled. Do not disclose exploit details in a public issue before coordination.

Contributions are licensed under **GPL-3.0-only**; see [LICENSE](LICENSE). Add this
header to new Go source files:

```go
// SPDX-License-Identifier: GPL-3.0-only
```

## Repository profiles

See [repository profiles and automation](docs/commit-profiles.md) for plain Git
messages, configuration, detection, and JSON output. Add regression tests for
changes to staging, hooks, editor mode, and configuration precedence. The profile
parser also has a fuzz target:

```sh
go test ./internal/profile -fuzz=FuzzInfer -fuzztime=10s -parallel=2
```
