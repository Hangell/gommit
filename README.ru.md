# Gommit [![Go Report Card](https://goreportcard.com/badge/github.com/Hangell/gommit)](https://goreportcard.com/report/github.com/Hangell/gommit) [![GitHub tag](https://img.shields.io/github/v/tag/Hangell/gommit?label=version&color=orange)](https://github.com/Hangell/gommit/tags) [![Build](https://github.com/Hangell/gommit/actions/workflows/ci.yml/badge.svg)](https://github.com/Hangell/gommit/actions/workflows/ci.yml) [![License](https://img.shields.io/github/license/Hangell/gommit)](LICENSE) [![Contributing](https://img.shields.io/badge/contributions-welcome-brightgreen)](CONTRIBUTING.md) [![Go Reference](https://pkg.go.dev/badge/github.com/Hangell/gommit.svg)](https://pkg.go.dev/github.com/Hangell/gommit)

**gommit** — это быстрый CLI‑помощник для _Conventional Commits_, написанный на Go и **не требующий внешних зависимостей**.  
Он запускает интерактивный мастер (похож на Commitizen/cz) и выполняет `git commit` с корректно отформатированным сообщением.

> 💡 По умолчанию **эмодзи типа коммита** добавляется **в заголовок** (например, `feat 💡: ...`), чтобы иконка была видна в списках файлов/папок GitHub. Эмодзи добавляется только в заголовок, не в тело/футер.

---

## ✨ Возможности

- ✅ Интерактивный мастер с поиском/ярлыками (цифры, ключевые слова, `q` — выход)
- ✅ Стандартные типы Conventional Commits: **feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert**  
  Плюс дополнительные: **WIP, prune**
- ✅ Эмодзи в меню (с автоматическим ASCII‑фолбэком) и **эмодзи в заголовке коммита**
- ✅ Пред‑проверки Git:
  - Проверка репозитория (`not a git repository…`, если это не репо)
  - Auto‑stage (`git add -A`), когда ничего не в staged (переключаемо)
  - Вывод списка staged‑изменений перед коммитом (переключаемо)
  - Режим `--amend` с баннером последнего коммита
- ✅ Поля в стиле Commitizen:
  - `scope` (необязательно)
  - `subject` (обязательно, `<=72` символов)
  - Многострочный `body` (завершите **одной точкой в отдельной строке** **или** **дважды Enter**)
  - `BREAKING CHANGE` (вопрос и описание)
  - Issues: `Closes #123` / `Refs #45` (по желанию)
- ✅ Встроенный режим **`--install`** — устанавливает/обновляет бинарник в PATH пользователя
  - Windows: `%LOCALAPPDATA%\Programs\gommitin`
  - Linux/macOS: `~/.local/bin` (добавляет в PATH при необходимости)

---

## 🚀 Установка

### Метод 1: Однострочные скрипты **(рекомендуется)**

**Windows:**
```powershell
powershell -ExecutionPolicy Bypass -NoProfile -Command "irm https://raw.githubusercontent.com/Hangell/gommit/main/scripts/install.ps1 | iex"
```

**Linux/macOS:**
```bash
curl -fsSL https://raw.githubusercontent.com/Hangell/gommit/main/scripts/install.sh | bash
```

### Метод 2: Скачать бинарник из Release
1. Скачайте архив для вашей ОС/архитектуры со страницы **Releases**
2. Распакуйте пакет (`.zip` на Windows / `.tar.gz` на Linux/macOS)
3. Запустите бинарник с флагом `--install`:

**Windows (PowerShell):**
```powershell
.\gommit.exe --install
# Закройте и снова откройте терминал
gommit --version
```

**Linux/macOS:**
```bash
chmod +x ./gommit
./gommit --install
# Перезапустите shell
gommit --version
```

> Повторный запуск `--install` обновит установленную версию.

---

## 💻 Быстрый старт

### Язык интерфейса

Команды и флаги остаются на английском. Меню, описания, вопросы и `--help` доступны
на английском, испанском, португальском, хинди, русском и китайском:

```bash
gommit --language ru          # только для этого запуска
gommit --set-language ru      # сохранить для текущего пользователя
```

Коды: `en`, `es`, `pt`, `hi`, `ru`, `zh`. По умолчанию используется английский.

При первой установке Gommit определяет и сохраняет поддерживаемый язык системы.
Выбор через `--set-language` никогда не перезаписывается.

Внутри Git‑репозитория с изменениями:

```bash
gommit
```

Типовой сценарий мастера:
1. Выберите **type** (номер, имя или поиск)
2. Укажите **scope** (необязательно)
3. Напишите **subject** (повелительное наклонение, `<=72` символов)
4. Добавьте **body** (необязательно, многострочно) — завершите `.` в отдельной строке **или** нажмите Enter **два раза**
5. **Есть breaking‑изменения?** (если «да», опишите изменение)
6. **Issues?** (Closes/Refs)

В конце `gommit` выполнит `git commit` (или покажет предпросмотр с `--dry-run`).

### Примеры

Обычный коммит (с авто‑индексацией):
```bash
gommit
```

Пустой коммит:
```bash
gommit --allow-empty
```

Изменить последний коммит:
```bash
gommit --amend
```

Только показать сообщение (без коммита):
```bash
gommit --dry-run
```

Все через флаги (без мастера):
```bash
gommit --type feat --scope ui --subject "добавить выбор типа" --body "Детали реализации..."
```

Неинтерактивно со всеми параметрами:
```bash
gommit --type fix --scope api --subject "исправить авторизацию" --body "Исправлена проверка JWT

Раньше просроченные токены обрабатывались неверно." --footer "Closes #42"
```

---

## 📝 Формат заголовка коммита

```
<type>[(!)][(scope)] <emoji>: <subject>
```

Примеры:
```
feat 💡: добавить команду install
fix(api) 🐛: исправить nil pointer при загрузке конфигурации
refactor(core)! 🎨: унифицировать сборщик сообщения
```

> `!` отображается только в **заголовке**; внутренняя валидация использует «чистый» тип (`feat`, `fix` и т.д.) для совместимости с инструментами Conventional Commits.

---

## ⚙️ Флаги командной строки

| Флаг | Описание |
|---|---|
| `--version` | Показать версию и выйти |
| `--install` | Установить/обновить бинарник в PATH и выйти |
| `--dry-run` | Только вывести сгенерированное сообщение (без `git commit`) |
| `--language` | Использовать `en`, `es`, `pt`, `hi`, `ru` или `zh` для запуска |
| `--set-language` | Сохранить язык интерфейса пользователя |
| `--type` | Тип коммита (feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert, WIP, prune) |
| `--scope` | Необязательный scope (например, `ui`, `api`) |
| `--subject` | Заголовок (повелительное наклонение, `<=72` символов) |
| `--body` | Текст тела (используйте `
` для переноса строк) |
| `--footer` | Футер (например, `Closes #123`) |
| `--as-editor` | Режим редактора Git (`COMMIT_EDITMSG`) |
| `--allow-empty` | Разрешить пустой коммит |
| `--amend` | Изменить предыдущий коммит |
| `--no-verify` | Пропустить хуки (`pre-commit`, `commit-msg`) |
| `--signoff` | Добавить трейлер `Signed-off-by` |
| `--auto-stage` | Делать `git add -A`, если ничего не staged (по умолчанию: `true`) |
| `--show-status` | Показывать сводку staged‑изменений перед коммитом (по умолчанию: `true`) |

---

## 📋 Поддерживаемые типы коммитов

| Тип | Эмодзи | Описание |
|---|---|---|
| `WIP` | 🚧 | Работа в процессе |
| `feat` | 💡 | Новая функциональность |
| `fix` | 🐛 | Исправление бага |
| `chore` | 📦 | Зависимости, деплой, конфигурация |
| `refactor` | 🎨 | Улучшение структуры/формата кода |
| `prune` | 🔥 | Удаление кода/файлов |
| `docs` | 📝 | Документация |
| `perf` | ⚡ | Производительность |
| `test` | ✅ | Тесты |
| `build` | 🔧 | Система сборки/зависимости |
| `ci` | 🤖 | Конфигурация CI/CD |
| `style` | 💅 | Стиль без изменения смысла кода |
| `revert` | ⏪ | Откат к коммиту |

---

## 🔧 Режим редактора Git

Настройте `gommit` как редактор сообщений:

```bash
git config --global core.editor "gommit --as-editor"
```

Мастер `gommit` будет открываться, когда Git запрашивает сообщение (в т.ч. при `git commit`, `git merge`, `git rebase`).

---

## 🛠️ Разработка

### Сборка из исходников

Требования: Go 1.22+

```bash
git clone https://github.com/Hangell/gommit.git
cd gommit

go build -o gommit ./cmd/gommit
go test ./...

./gommit --install
```

### Сборка с версией

```bash
go build -ldflags "-s -w -X main.version=0.2.0" -o gommit ./cmd/gommit
```

---

## 🎨 Переменные окружения

| Переменная | Описание |
|---|---|
| `NO_COLOR=1` | Отключить ANSI‑цвета в меню |
| `NO_EMOJI=1` | Принудительный ASCII‑фолбэк в меню (в заголовке по‑прежнему emoji) |

---

## 🤝 Вклад

1. Форкните репозиторий  
2. Создайте ветку (`git checkout -b feature/awesome-feature`)  
3. Делайте коммиты через `gommit` 😉  
4. `git push origin feature/awesome-feature`  
5. Откройте Pull Request

---

## 📄 Лицензия

GPL‑3.0‑only — см. файл [`LICENSE`](LICENSE) для подробностей.

---

## ❓ FAQ

**Заменяет ли gommit Commitizen?** Это **альтернатива**. Если у вас уже налажен процесс на Node/npm и вы довольны адаптерами cz — продолжайте. Если хотите **меньше зависимостей**, **единую установку** и **одинаковое поведение** на разных репозиториях/серверах, gommit чаще всего проще.

**Быстрее ли он?** Мы не приводим бенчмарки здесь, но как нативный бинарник gommit обычно **стартует быстрее** и потребляет меньше памяти, чем Node‑CLI, особенно в CI/контейнерах (холодный старт Node добавляет задержку).

**Работает офлайн?** Да. После установки сети не требуется.

**Почему появился gommit?** Из‑за рутины с nvm/множественными установками cz на разных версиях/проектах. gommit — более простой единый инструмент.

---

## 👨‍💻 Автор

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

## 🙏 Благодарности

- Вдохновлено [Commitizen](https://github.com/commitizen/cz-cli)
- Основано на спецификации [Conventional Commits](https://www.conventionalcommits.org/)
- Эмодзи делают историю коммитов наглядней и веселее! 🎉

---

<div align="center">
  <strong>Сделано с ❤️ для сообщества разработчиков</strong>
</div>
