# Gommit [![Go Report Card](https://goreportcard.com/badge/github.com/Hangell/gommit)](https://goreportcard.com/report/github.com/Hangell/gommit) [![GitHub tag](https://img.shields.io/github/v/tag/Hangell/gommit?label=version&color=orange)](https://github.com/Hangell/gommit/tags) [![Build](https://github.com/Hangell/gommit/actions/workflows/ci.yml/badge.svg)](https://github.com/Hangell/gommit/actions/workflows/ci.yml) [![License](https://img.shields.io/github/license/Hangell/gommit)](LICENSE) [![Contributing](https://img.shields.io/badge/contributions-welcome-brightgreen)](CONTRIBUTING.md) [![Go Reference](https://pkg.go.dev/badge/github.com/Hangell/gommit.svg)](https://pkg.go.dev/github.com/Hangell/gommit)


**gommit** é um assistente de linha de comando rápido e sem dependências para _Conventional Commits_, escrito em Go.  
Ele abre um assistente interativo (similar ao Commitizen/cz) e executa `git commit` com a mensagem devidamente formatada.

> 💡 Por padrão, **emojis do tipo de commit** são incluídos **no cabeçalho** (ex: `feat 💡: ...`) para torná-los visíveis nas listagens de arquivos/pastas do GitHub. Os emojis são adicionados apenas ao título, não ao corpo ou rodapé.

---

## ✨ Funcionalidades

- ✅ Assistente interativo com busca/atalhos (números, palavras-chave, `q` para sair)
- ✅ Tipos padrão do Conventional Commits: **feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert**  
  Mais extras: **WIP, prune**
- ✅ Emojis no menu (com fallback ASCII automático) e **emoji no cabeçalho do commit**
- ✅ Verificações prévias do Git:
  - Validação de repositório (`not a git repository…` se não estiver em um)
  - Auto-stage (`git add -A`) quando nada está staged (configurável)
  - Lista mudanças staged antes de commitar (configurável)
  - Modo `--amend` com banner do último commit
- ✅ Campos completos no estilo Commitizen:
  - `scope` (opcional)
  - `subject` (obrigatório, `<=72` caracteres)
  - `body` multilinha (termine com **ponto único em uma linha** **ou** **pressione Enter duas vezes**)
  - `BREAKING CHANGE` (pergunta e pede descrição)
  - Issues: `Closes #123` / `Refs #45` (se desejado)
- ✅ Modo **`--install`** integrado que instala/atualiza o binário no PATH do usuário
  - Windows: `%LOCALAPPDATA%\Programs\gommit\bin`
  - Linux/macOS: `~/.local/bin` (adiciona ao PATH se estiver faltando)

---

## 🚀 Instalação

### Método 1: Scripts de Instalação de Uma Linha **(Recomendado)**

**Windows:**
```powershell
powershell -ExecutionPolicy Bypass -NoProfile -Command "irm https://raw.githubusercontent.com/Hangell/gommit/main/scripts/install.ps1 | iex"
```

**Linux/macOS:**
```bash
curl -fsSL https://raw.githubusercontent.com/Hangell/gommit/main/scripts/install.sh | bash
```

### Método 2: Download do Binário da Release
1. Baixe o arquivo para seu OS/arquitetura da página de **Releases**
2. Extraia o pacote (`.zip` no Windows / `.tar.gz` no Linux/macOS)
3. Execute o binário extraído com a flag `--install`:

**Windows (PowerShell):**
```powershell
.\gommit.exe --install
# Feche e reabra o terminal
gommit --version
```

**Linux/macOS:**
```bash
chmod +x ./gommit
./gommit --install
# Reinicie o shell
gommit --version
```

> Executar `--install` novamente irá atualizar a instalação.

---

## 💻 Uso Rápido

### Idioma da interface

Comandos e flags continuam em inglês. Menus, descrições, perguntas e o `--help`
podem ser exibidos em inglês, espanhol, português, hindi, russo ou chinês:

```bash
gommit --language pt          # somente nesta execução
gommit --set-language pt      # salva para o usuário atual
```

Códigos: `en`, `es`, `pt`, `hi`, `ru`, `zh`. O padrão é inglês.

Na primeira instalação, o Gommit detecta e salva o idioma do sistema quando houver
suporte. Uma escolha feita com `--set-language` nunca é sobrescrita.

### Commit automático com Gemini

Instale e autentique o Gemini CLI oficial e configure `GEMINI_API_KEY`. O Gommit
também aceita `GEMINI_KEY` como apelido. Selecione `AI` no menu de tipos (ou use
`--type AI`); o Gemini escolherá o tipo Conventional Commit e criará um assunto
técnico com no máximo 72 caracteres.

Somente `git diff --cached` é enviado, com limite de 512 KiB; conteúdo binário e
arquivos fora do stage não são incluídos. Revise o stage para evitar o envio de
segredos. Os hooks Git/Husky continuam sendo executados normalmente.

Dentro de um repositório Git com mudanças:

```bash
gommit
```

O modo simples é o padrão:
1. Selecione o **tipo** usando `↑`/`↓` e confirme com Enter.
2. Descreva a mudança uma única vez e pressione Enter.

Para usar temporariamente o assistente completo antigo, execute `gommit --mode full`.
Para salvá-lo como preferência do usuário, execute `gommit --set-mode full`. Volte ao
fluxo simples com `gommit --set-mode simple`. A preferência é armazenada em
`git config --global gommit.mode`.

No final, `gommit` executa `git commit` normalmente, portanto hooks do Git, Husky e
`commit-msg` são respeitados. Eles só são ignorados quando `--no-verify` é informado.
Após um commit bem-sucedido, o Gommit verifica atualizações no máximo uma vez a cada
24 horas. Se houver uma nova release, ele apenas avisa; o commit nunca é bloqueado.

### Exemplos

Commit normal (com auto-stage):
```bash
gommit
```

Permitir commit vazio:
```bash
gommit --allow-empty
```

Emendar último commit:
```bash
gommit --amend
```

Mostrar apenas a mensagem (não commitar):
```bash
gommit --dry-run
```

Baixar e instalar a release mais recente no equipamento:
```bash
gommit --update
```

Passar tudo via flags (sem assistente):
```bash
gommit --type feat --scope ui --subject "adicionar seletor de tipo" --body "Detalhes da implementação..."
```

Modo não-interativo com todos os parâmetros:
```bash
gommit --type fix --scope api --subject "resolver problema de autenticação" --body "Corrigida validação de token JWT\n\nIsto resolve o problema onde tokens expirados\nnão estavam sendo tratados adequadamente." --footer "Closes #42"
```

---

## 📝 Formato do Cabeçalho do Commit

```
<type>[(!)][(scope)] <emoji>: <subject>
```

Exemplos:
```
feat 💡: adicionar comando install
fix(api) 🐛: corrigir nil pointer no carregamento de config
refactor(core)! 🎨: unificar message builder
```

> O `!` aparece apenas no **cabeçalho**; a validação interna usa o tipo "puro" (`feat`, `fix`, etc.) para manter compatibilidade com ferramentas de Conventional Commits.

---

## ⚙️ Flags de Linha de Comando

| Flag | Descrição |
|---|---|
| `--version` | Mostra versão e sai |
| `--install` | Instala/atualiza binário no PATH do usuário e sai |
| `--update` | Baixa e instala a release mais recente do GitHub |
| `--dry-run` | Apenas imprime a mensagem gerada (não chama `git commit`) |
| `--language` | Usa `en`, `es`, `pt`, `hi`, `ru` ou `zh` nesta execução |
| `--set-language` | Salva globalmente o idioma da interface do usuário |
| `--mode` | Usa `simple` ou `full` somente nesta execução |
| `--set-mode` | Salva `simple` ou `full` como padrão global do usuário |
| `--type` | Tipo do commit (feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert, WIP, prune) |
| `--scope` | Scope opcional (ex: `ui`, `api`) |
| `--subject` | Linha de assunto (modo imperativo, `<=72` chars) |
| `--body` | Texto do corpo (use `\n` para novas linhas) |
| `--footer` | Rodapé manual (ex: `Closes #123`) |
| `--as-editor` | Executar como editor do Git (modo `COMMIT_EDITMSG`) |
| `--allow-empty` | Permitir commit vazio |
| `--amend` | Emendar o último commit |
| `--no-verify` | Pular hooks (`pre-commit`, `commit-msg`) |
| `--signoff` | Adicionar trailer `Signed-off-by` |
| `--auto-stage` | Auto `git add -A` quando nada está staged (padrão: `true`) |
| `--show-status` | Mostrar resumo de mudanças staged antes do commit (padrão: `true`) |

---

## 📋 Tipos de Commit Suportados

| Tipo | Emoji | Descrição |
|---|---|---|
| `WIP` | 🚧 | Trabalho em progresso |
| `feat` | 💡 | Uma nova funcionalidade |
| `fix` | 🐛 | Correção de um bug |
| `chore` | 📦 | Atualização de dependências, deploys, arquivos de configuração |
| `refactor` | 🎨 | Melhoria da estrutura/formato do código |
| `prune` | 🔥 | Remoção de código ou arquivos |
| `docs` | 📝 | Escrita de documentação |
| `perf` | ⚡ | Melhoria de performance |
| `test` | ✅ | Adição de testes |
| `build` | 🔧 | Mudanças no sistema de build ou dependências |
| `ci` | 🤖 | Mudanças na configuração de CI/CD |
| `style` | 💅 | Mudanças que não afetam o significado do código |
| `revert` | ⏪ | Reverter para um commit |

---

## 🔧 Modo Editor do Git

Você pode configurar `gommit` como seu editor do Git para mensagens de commit consistentes:

```bash
git config --global core.editor "gommit --as-editor"
```

Isso abrirá o assistente `gommit` sempre que o Git precisar de uma mensagem de commit (incluindo durante `git commit`, `git merge`, `git rebase`, etc.).

---

## 🛠️ Desenvolvimento

### Construindo do Código Fonte

Requisitos: Go 1.22+

```bash
# Clone o repositório
git clone https://github.com/Hangell/gommit.git
cd gommit

# Construir
go build -o gommit ./cmd/gommit

# Executar testes
go test ./...

# Instalar localmente
./gommit --install
```

### Construindo com Informações de Versão

Para releases com injeção de versão:
```bash
go build -ldflags "-s -w -X main.version=0.2.0" -o gommit ./cmd/gommit
```

---

## 🎨 Variáveis de Ambiente

| Variável | Descrição |
|---|---|
| `NO_COLOR=1` | Desabilita cores ANSI no menu |
| `NO_EMOJI=1` | Força fallback ASCII no menu (cabeçalho ainda usa emojis unicode) |

---

## 🤝 Contribuindo

1. Faça fork do repositório
2. Crie sua branch de feature (`git checkout -b feature/funcionalidade-incrivel`)
3. Commit suas mudanças usando `gommit` 😉
4. Push para a branch (`git push origin feature/funcionalidade-incrivel`)
5. Abra um Pull Request

---

## 📄 Licença

GPL-3.0-only — veja o arquivo [`LICENSE`](LICENSE) para detalhes.

---

## ❓ FAQ

**O gommit substitui o Commitizen?** É uma **alternativa**. Se seu ambiente já depende de Node/npm e você está confortável com adaptadores cz, continue com ele. Se você quer **menos dependências**, **instalação única**, e **uso consistente** em múltiplos repositórios/servidores, gommit é provavelmente mais simples.

**É mais rápido?** Não publicamos benchmarks aqui, mas como um binário nativo, gommit **tende a iniciar mais rápido** e usar menos memória que ferramentas CLI tradicionais do Node — especialmente em ambientes frios (CI, containers) onde iniciar o runtime do Node adiciona latência.

**Funciona offline?** Sim. Uma vez instalado, não precisa de acesso à rede para executar o assistente ou fazer commit.

**Por que o gommit foi criado?** A ideia nasceu da frustração com ambientes nvm onde tínhamos que instalar `git-cz` para cada versão do nvm ou diretamente em cada projeto. O gommit é muito mais simples e fornece uma solução única e unificada.

---

## 👨‍💻 Autor

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

## 🙏 Agradecimentos

- Inspirado pelo [Commitizen](https://github.com/commitizen/cz-cli)
- Construído com a especificação [Conventional Commits](https://www.conventionalcommits.org/)
- Emojis ajudam a tornar o histórico de commits mais visual e divertido! 🎉

---

<div align="center">
  <strong>Feito com ❤️ para a comunidade de desenvolvedores</strong>
</div>
