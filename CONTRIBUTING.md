# Contribuindo para o **gommit**

Obrigado por dedicar seu tempo para contribuir! Este guia explica como planejar mudanças, abrir issues/PRs, configurar o ambiente, seguir o padrão de commits e publicar alterações com segurança.

> **Resumo rápido:**
> - Linguagem: **Go 1.22+**
> - Sistema: macOS, Linux e Windows
> - Estilo de commit: **Conventional Commits**
> - Licença: **GPL-3.0-only** (todas as contribuições seguem esta licença)
> - Sem dependências de Node/NPM/NVM — é um **CLI em Go** (binário único)

---

## Sumário
- [Código de Conduta](#código-de-conduta)
- [Como posso ajudar?](#como-posso-ajudar)
- [Ambiente de desenvolvimento](#ambiente-de-desenvolvimento)
- [Build, testes e lint](#build-testes-e-lint)
- [Executando localmente](#executando-localmente)
- [Padrão de commits](#padrão-de-commits)
- [Estratégia de branches](#estratégia-de-branches)
- [Processo de Pull Request](#processo-de-pull-request)
- [Estrutura do projeto](#estrutura-do-projeto)
- [Versionamento e releases](#versionamento-e-releases)
- [Segurança e divulgação responsável](#segurança-e-divulgação-responsável)
- [Licença e cabeçalho SPDX](#licença-e-cabeçalho-spdx)

---

## Código de Conduta
Adotamos o **Contributor Covenant**. A participação na comunidade implica concordância com um ambiente acolhedor e respeitoso. Veja `CODE_OF_CONDUCT.md` (se ausente, será adicionado em breve).

---

## Como posso ajudar?
- **Bugs**: abra uma *issue* com passos para reproduzir, logs, plataforma (SO/arquitetura) e versão do `gommit`.
- **Features**: descreva o problema real a ser resolvido, alternativas consideradas e impacto esperado. Propostas maiores → use o rótulo **RFC**.
- **Docs**: melhorias no `README`, `CONTRIBUTING`, exemplos, gifs e correções gramaticais são bem-vindas.

> Dica: problemas “good first issue” e “help wanted” facilitam a entrada.

---

## Ambiente de desenvolvimento
**Pré‑requisitos**
- Go **1.22+**
- Git
- (Opcional) `golangci-lint`, `staticcheck`, `goreleaser`

**Clonar e preparar**
```bash
# 1) Fork no GitHub e clone seu fork
 git clone https://github.com/<seu-usuario>/gommit
 cd gommit

# 2) Configure o módulo (se necessário) e baixe deps
 go mod tidy
```

> Este projeto **não** depende de Node/NPM/NVM.

---

## Build, testes e lint
**Build local**
```bash
# compila o binário para sua plataforma
go build ./cmd/gommit
```

**Testes**
```bash
go test ./...
# cobertura (opcional)
go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out
```

**Lint (mínimo razoável)**
```bash
# verificação estática padrão do Go
go vet ./...
```

> Se usar `golangci-lint` ou `staticcheck`, mantenha os findings relevantes no PR.

---

## Executando localmente
**Modo assistente manual**
```bash
# no diretório do projeto (após build)
./gommit commit
```

**Integrar como editor do Git (recomendado para testes)**
- macOS/Linux:
  ```bash
  git config --global core.editor "$(pwd)/gommit --as-editor"
  ```
- Windows (CMD):
  ```bat
  git config --global core.editor "\"%CD%\\gommit.exe\" --as-editor"
  ```
- Windows (PowerShell):
  ```powershell
  git config --global core.editor '"%CD%\gommit.exe" --as-editor'
  ```

Agora, em qualquer repositório:
```bash
git add .
git commit  # abre o wizard do gommit
```

> Para reverter: `git config --global --unset core.editor`.

---

## Padrão de commits
Seguimos **Conventional Commits**. O `gommit` ajuda nesse fluxo, mas PRs devem respeitar o formato:

**Formato**
```
<type>(<scope>)!: <subject>

<body>

<footer>
```

**Tipos aceitos (MVP)**: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, `revert`.

**Regras rápidas**
- `<subject>` no imperativo e preferencialmente ≤ **72** caracteres.
- Use `!` em *breaking changes* e inclua `BREAKING CHANGE:` no body/footer.
- Relacione issues no footer: `Closes #123`, `Refs #456`.
- Co-autoria: `Co-authored-by: Nome <email>`.

**Exemplos**
```
feat(ui): add scope suggestions from staged paths

Suggest top-level folders as commit scope based on `git diff --cached --name-only`.

Closes #42
```
```
fix(editor)!: prevent overwriting merge commit messages

BREAKING CHANGE: editor flow now aborts when COMMIT_EDITMSG contains meaningful content.
```

---

## Estratégia de branches
- **main**: estável.
- **feature branches**: `feat/<slug>`, `fix/<slug>`, `chore/<slug>`.
- Evite PRs gigantes. Prefira mudanças pequenas e revisáveis.

---

## Processo de Pull Request
1. **Abra uma issue** (quando a mudança não for trivial) e alinhe o escopo.
2. **Implemente** com testes quando fizer sentido.
3. **Rodar checks locais**: `go vet`, `go test ./...`.
4. **Commits**: use o `gommit` ou siga o padrão manualmente.
5. **Assinatura (DCO)**: assine seus commits com `-s` (Developer Certificate of Origin).
   ```bash
   git commit -s -m "feat: ..."
   ```
6. **PR**: descreva o *rationale*, evidencie impactos, screenshots/gifs quando UI do terminal mudar.
7. **Review**: esteja aberto a feedbacks (pequenos refinamentos são comuns).

Checklist para o PR:
- [ ] Tests passam em `go test ./...`
- [ ] Mensagem de commit segue o padrão
- [ ] Cobertura razoável (quando aplicável)
- [ ] Documentação/README atualizada (se mudou UX/flags)

---

## Estrutura do projeto
```
cmd/gommit/main.go   # ponto de entrada do CLI
internal/commit      # montagem da mensagem, validações
internal/ui          # prompts/wizard (stdin/stdout), i18n
internal/git         # integrações com Git (rev-parse, diff, commit -F)
internal/config      # ./.gommit.json e ~/.gommit.json (futuro)
internal/editor      # fluxo --as-editor
internal/hook        # instalação de hooks (opcional)
assets/              # logo.svg, demo.gif
```

---

## Versionamento e releases
- **SemVer** para tags: `vMAJOR.MINOR.PATCH`.
- Releases oficiais são criadas pelos mantenedores (GoReleaser).
- Para propor uma release, abra uma issue com changelog proposto.

---

## Segurança e divulgação responsável
Se você encontrar uma vulnerabilidade, **não** abra uma issue pública de imediato. Envie um e‑mail para o(s) mantenedor(es) com detalhes e passos de reprodução. Daremos retorno e coordenaremos a correção antes da divulgação.

---

## Licença e cabeçalho SPDX
- Este projeto é licenciado sob **GPL-3.0-only** (veja `LICENSE`).
- Inclua o cabeçalho SPDX nos arquivos novos/modificados:
  ```go
  // SPDX-License-Identifier: GPL-3.0-only
  ```

Obrigado por contribuir! 🎉

