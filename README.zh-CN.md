# Gommit [![Go Report Card](https://goreportcard.com/badge/github.com/Hangell/gommit)](https://goreportcard.com/report/github.com/Hangell/gommit) [![GitHub tag](https://img.shields.io/github/v/tag/Hangell/gommit?label=version&color=orange)](https://github.com/Hangell/gommit/tags) [![Build](https://github.com/Hangell/gommit/actions/workflows/ci.yml/badge.svg)](https://github.com/Hangell/gommit/actions/workflows/ci.yml) [![License](https://img.shields.io/github/license/Hangell/gommit)](LICENSE) [![Contributing](https://img.shields.io/badge/contributions-welcome-brightgreen)](CONTRIBUTING.md) [![Go Reference](https://pkg.go.dev/badge/github.com/Hangell/gommit.svg)](https://pkg.go.dev/github.com/Hangell/gommit)

**gommit** 是一个由 Go 编写、**零运行时依赖** 的 _Conventional Commits_ 命令行助手。  
它提供类似 Commitizen/cz 的交互式向导，并以正确的格式执行 `git commit`。

> 💡 默认在提交 **标题** 中加入 **类型表情**（例如 `feat 💡: ...`），这样图标会在 GitHub 的文件/文件夹列表中可见。表情仅添加到标题，不会进入正文或页脚。

---

## ✨ 功能特性

- ✅ 交互式向导，支持搜索/快捷键（数字、关键字、`q` 退出）
- ✅ 标准 Conventional Commits 类型：**feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert**  
  以及额外类型：**WIP, prune**
- ✅ 菜单表情（自动 ASCII 回退）与 **标题表情**
- ✅ Git 预检查：
  - 仓库校验（若不是 Git 仓库则给出 `not a git repository…`）
  - 当没有 staged 变更时自动 `git add -A`（可配置）
  - 提交前显示 staged 变更（可配置）
  - `--amend` 模式并显示上次提交标题
- ✅ Commitizen 风格字段：
  - `scope`（可选）
  - `subject`（必填，`<=72` 字符）
  - 多行 `body`（以**单独一行的句点 `.`**结束，或**回车两次**）
  - `BREAKING CHANGE`（询问并填写描述）
  - Issues：`Closes #123` / `Refs #45`（可选）
- ✅ 内置 **`--install`** 安装模式，将二进制安装/更新到用户 PATH
  - Windows：`%LOCALAPPDATA%\Programs\gommitin`
  - Linux/macOS：`~/.local/bin`（缺失时自动加入 PATH）

---

## 🚀 安装

### 方法一：一键脚本 **（推荐）**

**Windows：**
```powershell
powershell -ExecutionPolicy Bypass -NoProfile -Command "irm https://raw.githubusercontent.com/Hangell/gommit/main/scripts/install.ps1 | iex"
```

**Linux/macOS：**
```bash
curl -fsSL https://raw.githubusercontent.com/Hangell/gommit/main/scripts/install.sh | bash
```

### 方法二：从 Release 下载二进制
1. 前往 **Releases** 页面下载适合你系统/架构的包
2. 解压（Windows 为 `.zip`，Linux/macOS 为 `.tar.gz`）
3. 使用 `--install` 进行安装：

**Windows (PowerShell)：**
```powershell
.\gommit.exe --install
# 重新打开终端
gommit --version
```

**Linux/macOS：**
```bash
chmod +x ./gommit
./gommit --install
# 重新载入 Shell
gommit --version
```

> 再次执行 `--install` 将更新已安装版本。

---

## 💻 快速使用

### 界面语言

命令和 flags 始终保持英文。菜单、说明、提示和 `--help` 可显示为英语、西班牙语、
葡萄牙语、印地语、俄语或中文：

```bash
gommit --language zh          # 仅本次运行
gommit --set-language zh      # 为当前用户保存
```

语言代码：`en`、`es`、`pt`、`hi`、`ru`、`zh`。默认语言为英语。

首次安装时，Gommit 会检测并保存系统支持的语言。通过 `--set-language` 设置的
语言永远不会被覆盖。

### 使用 Gemini 自动提交

安装并认证官方 Gemini CLI，然后设置 `GEMINI_API_KEY`（也支持 `GEMINI_KEY`）。
在菜单中选择 `AI`，即可仅根据 `git diff --cached`（最大 512 KiB）生成提交类型
和不超过 72 个字符的技术主题。使用前请检查 stage 中是否有密钥；Git/Husky
hooks 仍会正常执行。

在有变更的 Git 仓库中：

```bash
gommit
```

典型流程：
1. 选择 **type**（数字、名称或搜索）
2. 输入 **scope**（可选）
3. 填写 **subject**（祈使语气，`<=72` 字符）
4. 添加 **body**（可选，多行）——以 `.` 单独一行结束，**或** 连续按两次 Enter  
5. **有 breaking change 吗？**（若是，请描述）
6. **有关联的 issues 吗？**（Closes/Refs）

最后，`gommit` 会执行 `git commit`（或用 `--dry-run` 仅展示消息）。

### 示例

普通提交（自动 stage）：
```bash
gommit
```

允许空提交：
```bash
gommit --allow-empty
```

修改上次提交：
```bash
gommit --amend
```

只显示消息（不提交）：
```bash
gommit --dry-run
```

通过参数直接提交（无向导）：
```bash
gommit --type feat --scope ui --subject "添加类型选择器" --body "实现细节..."
```

全参数非交互示例：
```bash
gommit --type fix --scope api --subject "修复认证问题" --body "修复 JWT 令牌校验

此前过期令牌未被正确处理。" --footer "Closes #42"
```

---

## 📝 标题格式

```
<type>[(!)][(scope)] <emoji>: <subject>
```

示例：
```
feat 💡: 添加安装命令
fix(api) 🐛: 修正配置加载中的空指针
refactor(core)! 🎨: 统一消息构建器
```

> `!` 只出现在**标题**里；内部校验仍使用“纯类型”（`feat`、`fix` 等），与 Conventional Commits 工具保持兼容。

---

## ⚙️ 命令行参数

| 参数 | 说明 |
|---|---|
| `--version` | 显示版本并退出 |
| `--install` | 安装/更新二进制到用户 PATH 后退出 |
| `--dry-run` | 仅打印生成的消息（不执行 `git commit`） |
| `--language` | 本次运行使用 `en`、`es`、`pt`、`hi`、`ru` 或 `zh` |
| `--set-language` | 保存用户的全局界面语言 |
| `--type` | 提交类型（feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert, WIP, prune） |
| `--scope` | 可选作用域（如 `ui`、`api`） |
| `--subject` | 标题（祈使语气，`<=72` 字符） |
| `--body` | 正文文本（用 `
` 表示换行） |
| `--footer` | 页脚（如 `Closes #123`） |
| `--as-editor` | 以 Git 编辑器模式运行（`COMMIT_EDITMSG`） |
| `--allow-empty` | 允许空提交 |
| `--amend` | 修改上一次提交 |
| `--no-verify` | 跳过钩子（`pre-commit`、`commit-msg`） |
| `--signoff` | 添加 `Signed-off-by` 尾注 |
| `--auto-stage` | 当没有 staged 变更时自动 `git add -A`（默认：`true`） |
| `--show-status` | 提交前显示 staged 变更摘要（默认：`true`） |

---

## 📋 支持的提交类型

| 类型 | 表情 | 说明 |
|---|---|---|
| `WIP` | 🚧 | 进行中 |
| `feat` | 💡 | 新功能 |
| `fix` | 🐛 | 修复 Bug |
| `chore` | 📦 | 依赖、部署、配置等更新 |
| `refactor` | 🎨 | 重构（不改变行为） |
| `prune` | 🔥 | 删除代码或文件 |
| `docs` | 📝 | 文档 |
| `perf` | ⚡ | 性能优化 |
| `test` | ✅ | 测试 |
| `build` | 🔧 | 构建系统/依赖变更 |
| `ci` | 🤖 | CI/CD 配置变更 |
| `style` | 💅 | 代码风格（不影响含义） |
| `revert` | ⏪ | 回退提交 |

---

## 🔧 Git 编辑器模式

将 `gommit` 设为 Git 编辑器：

```bash
git config --global core.editor "gommit --as-editor"
```

当 Git 需要提交信息时（包括 `git commit`、`git merge`、`git rebase` 等），会弹出 `gommit` 向导。

---

## 🛠️ 开发

### 从源码构建

要求：Go 1.22+

```bash
git clone https://github.com/Hangell/gommit.git
cd gommit

go build -o gommit ./cmd/gommit
go test ./...

./gommit --install
```

### 带版本信息的构建

```bash
go build -ldflags "-s -w -X main.version=0.2.0" -o gommit ./cmd/gommit
```

---

## 🎨 环境变量

| 变量 | 说明 |
|---|---|
| `NO_COLOR=1` | 禁用菜单中的 ANSI 颜色 |
| `NO_EMOJI=1` | 强制菜单使用 ASCII 回退（标题仍使用 Unicode 表情） |

---

## 🤝 贡献

1. Fork 项目  
2. 创建分支（`git checkout -b feature/amazing-feature`）  
3. 使用 `gommit` 提交 😉  
4. 推送分支（`git push origin feature/amazing-feature`）  
5. 发起 Pull Request

---

## 📄 许可证

GPL-3.0-only —— 见 [`LICENSE`](LICENSE)。

---

## ❓ 常见问题

**gommit 替代 Commitizen 吗？** 它是一个**替代方案**。如果你的流程强依赖 Node/npm 与 cz 适配器，可以继续使用。若你更希望**更少依赖**、**一次安装**并在多仓库/服务器**一致**使用，gommit 往往更简单。

**更快吗？** 未在此给出基准，但作为原生二进制，gommit 通常**启动更快**、占用更小，尤其适合 CI/容器等冷启动场景（Node 启动会增加时延）。

**可离线使用吗？** 可以。安装完成后无需网络。

**为什么要做 gommit？** 为避免在 nvm/不同项目里反复安装 cz 与适配器，提供一个更简单统一的方案。

---

## 👨‍💻 作者

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

## 🙏 致谢

- 灵感来自 [Commitizen](https://github.com/commitizen/cz-cli)
- 遵循 [Conventional Commits](https://www.conventionalcommits.org/) 规范
- 表情让提交历史更直观更有趣！🎉

---

<div align="center">
  <strong>用 ❤️ 为开发者社区打造</strong>
</div>
