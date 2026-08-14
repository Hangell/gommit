
# Gommit [![Go Report Card](https://goreportcard.com/badge/github.com/Hangell/gommit)](https://goreportcard.com/report/github.com/Hangell/gommit) [![GitHub tag](https://img.shields.io/github/v/tag/Hangell/gommit?label=version&color=orange)](https://github.com/Hangell/gommit/tags) [![Build](https://github.com/Hangell/gommit/actions/workflows/ci.yml/badge.svg)](https://github.com/Hangell/gommit/actions/workflows/ci.yml) [![License](https://img.shields.io/github/license/Hangell/gommit)](LICENSE) [![Contributing](https://img.shields.io/badge/contributions-welcome-brightgreen)](CONTRIBUTING.md) [![Go Reference](https://pkg.go.dev/badge/github.com/Hangell/gommit.svg)](https://pkg.go.dev/github.com/Hangell/gommit)



[![pt-BR](https://img.shields.io/badge/🇧🇷-Português-black)](README.pt-BR.md)
[![ru](https://img.shields.io/badge/🇷🇺-Русский-black)](README.ru.md)
[![hi](https://img.shields.io/badge/🇮🇳-Hindi-black)](README.hi.md)
[![zh-CN](https://img.shields.io/badge/🇨🇳-中文-black)](README.zh-CN.md)
[![es](https://img.shields.io/badge/🇪🇸-Español-black)](README.es.md)
[![en](https://img.shields.io/badge/🇺🇸-English-green)](README.md)

**gommit** is a fast, zero-dependency command-line assistant for _Conventional Commits_, written in Go.  
It opens an interactive wizard (similar to Commitizen/cz) and executes `git commit` with the properly formatted message.

> 💡 By default, **commit type emojis** are included **in the header** (e.g., `feat 💡: ...`) to make them visible in GitHub's file/folder listings. Emojis are only added to the title, not to the body or footer.

---

## ✨ Features

- ✅ Interactive wizard with search/shortcuts (numbers, keywords, `q` to quit)
- ✅ Standard Conventional Commits types: **feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert**  
  Plus extras: **WIP, prune**
- ✅ Emojis in menu (with automatic ASCII fallback) and **emoji in commit header**
- ✅ Git pre-checks:
  - Repository validation (`not a git repository…` if not in one)
  - Auto-stage (`git add -A`) when nothing is staged (toggleable)
  - List staged changes before committing (toggleable)
  - `--amend` mode with last commit banner
- ✅ Full Commitizen-style fields:
  - `scope` (optional)
  - `subject` (required, `<=72` characters)
  - `body` multiline (end with **single dot on a line** **or** **press Enter twice**)
  - `BREAKING CHANGE` (prompts and asks for description)
  - Issues: `Closes #123` / `Refs #45` (if desired)
- ✅ Built-in **`--install`** mode that installs/updates the binary in user PATH
  - Windows: `%LOCALAPPDATA%\Programs\gommit\bin`
  - Linux/macOS: `~/.local/bin` (adds to PATH if missing)

---

## 🚀 Installation

### Method 1: One-liner Installation Scripts **(Recommended)**

**Windows:**
```powershell
powershell -ExecutionPolicy Bypass -NoProfile -Command "irm https://raw.githubusercontent.com/Hangell/gommit/main/scripts/install.ps1 | iex"
```

**Linux/macOS:**
```bash
curl -fsSL https://raw.githubusercontent.com/Hangell/gommit/main/scripts/install.sh | bash
```


### Method 2: Download Release Binary
1. Download the artifact for your OS/architecture from the **Releases** page
2. Extract the package (`.zip` on Windows / `.tar.gz` on Linux/macOS)
3. Run the extracted binary with the `--install` flag:

**Windows (PowerShell):**
```powershell
.\gommit.exe --install
# Close and reopen terminal
gommit --version
```

**Linux/macOS:**
```bash
chmod +x ./gommit
./gommit --install
# Restart shell
gommit --version
```

> Running `--install` again will update the installation.

---

## 💻 Quick Usage

### Interface language

Commands and flags remain in English, while menus, descriptions, prompts and `--help`
can use English, Spanish, Portuguese, Hindi, Russian or Chinese:

```bash
gommit --language pt          # only this run
gommit --set-language pt      # save for the current user
```

Language codes: `en`, `es`, `pt`, `hi`, `ru`, `zh`. English is the default.

Inside a Git repository with changes:

```bash
gommit
```

Simple mode is the default:
1. Select the **type** with `↑`/`↓` and confirm with Enter.
2. Describe the change once and press Enter.

Run `gommit --mode full` to use the previous full wizard once. Save it as the user
default with `gommit --set-mode full`, or return to the simple flow with
`gommit --set-mode simple`. The preference is stored in `git config --global gommit.mode`.

At the end, `gommit` runs `git commit` normally, so Git hooks, Husky, and `commit-msg`
hooks run as usual. They are skipped only when `--no-verify` is explicitly passed.
After a successful commit, Gommit checks for updates at most once every 24 hours.
When a release is available it only prints a notice; the commit is never blocked.

### Examples

Normal commit (with auto-stage):
```bash
gommit
```

Allow empty commit:
```bash
gommit --allow-empty
```

Amend last commit:
```bash
gommit --amend
```

Show message only (don't commit):
```bash
gommit --dry-run
```

Download and install the latest release on this machine:
```bash
gommit --update
```

Pass everything via flags (no wizard):
```bash
gommit --type feat --scope ui --subject "add type picker" --body "Implementation details..."
```

Non-interactive mode with all parameters:
```bash
gommit --type fix --scope api --subject "resolve authentication issue" --body "Fixed JWT token validation\n\nThis resolves the issue where expired tokens\nwere not being properly handled." --footer "Closes #42"
```

---

## 📝 Commit Header Format

```
<type>[(!)][(scope)] <emoji>: <subject>
```

Examples:
```
feat 💡: add install command
fix(api) 🐛: correct nil pointer on config load
refactor(core)! 🎨: unify message builder
```

> The `!` appears only in the **header**; internal validation uses the "pure" type (`feat`, `fix`, etc.) to maintain compatibility with Conventional Commits tools.

---

## ⚙️ Command Line Flags

| Flag | Description |
|---|---|
| `--version` | Show version and exit |
| `--install` | Install/update binary to user PATH and exit |
| `--update` | Download and install the latest GitHub release |
| `--dry-run` | Only print the generated message (don't call `git commit`) |
| `--language` | Use `en`, `es`, `pt`, `hi`, `ru` or `zh` for this run |
| `--set-language` | Save the user's global interface language |
| `--mode` | Use `simple` or `full` for this run only |
| `--set-mode` | Save `simple` or `full` as the user's global default |
| `--type` | Commit type (feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert, WIP, prune) |
| `--scope` | Optional scope (e.g., `ui`, `api`) |
| `--subject` | Subject line (imperative mood, `<=72` chars) |
| `--body` | Body text (use `\n` for new lines) |
| `--footer` | Manual footer (e.g., `Closes #123`) |
| `--as-editor` | Run as Git editor (`COMMIT_EDITMSG` mode) |
| `--allow-empty` | Allow empty commit |
| `--amend` | Amend the last commit |
| `--no-verify` | Skip hooks (`pre-commit`, `commit-msg`) |
| `--signoff` | Add `Signed-off-by` trailer |
| `--auto-stage` | Auto `git add -A` when nothing is staged (default: `true`) |
| `--show-status` | Show staged changes summary before commit (default: `true`) |

---

## 📋 Supported Commit Types

| Type | Emoji | Description |
|---|---|---|
| `WIP` | 🚧 | Work in progress |
| `feat` | 💡 | A new feature |
| `fix` | 🐛 | Fixing a bug |
| `chore` | 📦 | Updating dependencies, deployments, configuration files |
| `refactor` | 🎨 | Improving structure/format of the code |
| `prune` | 🔥 | Removing code or files |
| `docs` | 📝 | Writing documentation |
| `perf` | ⚡ | Improving performance |
| `test` | ✅ | Adding tests |
| `build` | 🔧 | Changes to build system or dependencies |
| `ci` | 🤖 | Changes to CI/CD configuration |
| `style` | 💅 | Changes that do not affect the meaning of the code |
| `revert` | ⏪ | Revert to a commit |

---

## 🔧 Git Editor Mode

You can configure `gommit` as your Git editor for consistent commit messages:

```bash
git config --global core.editor "gommit --as-editor"
```

This will open the `gommit` wizard whenever Git needs a commit message (including during `git commit`, `git merge`, `git rebase`, etc.).

---

## 🛠️ Development

### Building from Source

Requirements: Go 1.22+

```bash
# Clone the repository
git clone https://github.com/Hangell/gommit.git
cd gommit

# Build
go build -o gommit ./cmd/gommit

# Run tests
go test ./...

# Install locally
./gommit --install
```

### Building with Version Information

For releases with version injection:
```bash
go build -ldflags "-s -w -X main.version=0.2.0" -o gommit ./cmd/gommit
```

---

## 🎨 Environment Variables

| Variable | Description |
|---|---|
| `NO_COLOR=1` | Disable ANSI colors in menu |
| `NO_EMOJI=1` | Force ASCII fallback in menu (header still uses unicode emojis) |

---

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes using `gommit` 😉
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## 📄 License

GPL-3.0-only — see [`LICENSE`](LICENSE) file for details.

---

## ❓ FAQ

**Does gommit replace Commitizen?** It's an **alternative**. If your environment already depends on Node/npm and you're comfortable with cz adapters, stick with it. If you want **fewer dependencies**, **single installation**, and **consistent usage** across multiple repositories/servers, gommit is probably simpler.

**Is it faster?** We don't publish benchmarks here, but as a native binary, gommit **tends to start faster** and use less memory than traditional Node CLI tools — especially in cold environments (CI, containers) where starting the Node runtime adds latency.

**Does it work offline?** Yes. Once installed, it doesn't need network access to run the wizard or commit.

**Why was gommit created?** The idea was born from frustration with nvm environments where we had to install `git-cz` for each nvm version or directly in each project. gommit is much simpler and provides a single, unified solution.

---

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes using `gommit` 😉
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 👨‍💻 Author

<div align="center">
  <img src="https://avatars.githubusercontent.com/u/53544561?v=4" width="150" style="border-radius: 50%;" />

**Rodrigo Rangel**

  <div>
    <a href="https://hangell.org" target="_blank">
      <img src="https://img.shields.io/badge/website-000000?style=for-the-badge&logo=About.me&logoColor=white" alt="Website" />
    </a>
    <a href="https://play.google.com/store/apps/dev?id=5606456325281613718" target="_blank">
      <img src="https://img.shields.io/badge/Google_Play-414141?style=for-the-badge&logo=google-play&logoColor=white" alt="Google Play" />
    </a>
    <a href="https://www.youtube.com/channel/UC8_zG7RFM2aMhI-p-6zmixw" target="_blank">
      <img src="https://img.shields.io/badge/YouTube-FF0000?style=for-the-badge&logo=youtube&logoColor=white" alt="YouTube" />
    </a>
    <a href="https://www.facebook.com/hangell.org" target="_blank">
      <img src="https://img.shields.io/badge/Facebook-1877F2?style=for-the-badge&logo=facebook&logoColor=white" alt="Facebook" />
    </a>
    <a href="https://www.linkedin.com/in/rodrigo-rangel-a80810170" target="_blank">
      <img src="https://img.shields.io/badge/-LinkedIn-%230077B5?style=for-the-badge&logo=linkedin&logoColor=white" alt="LinkedIn" />
    </a>
  </div>
</div>

---

## 📄 License

GPL-3.0-only — see [`LICENSE`](LICENSE) file for details.

---

## 🙏 Acknowledgments

- Inspired by [Commitizen](https://github.com/commitizen/cz-cli)
- Built with the [Conventional Commits](https://www.conventionalcommits.org/) specification
- Emojis help make commit history more visual and fun! 🎉

---

<div align="center">
  <strong>Made with ❤️ for the developer community</strong>
</div>
