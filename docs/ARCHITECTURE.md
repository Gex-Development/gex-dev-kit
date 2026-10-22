# Arquitetura

O kit é um programa em Go (`cmd/gex-dev-kit`). A tela (TUI) e o modo linha de comando (CLI) usam os mesmos casos de uso em `internal/app`. Assim, as duas formas de uso fazem a mesma coisa. A gravação nos arquivos do usuário fica só em `internal/kit`.

Termos do projeto estão em [CONTEXT.md](CONTEXT.md). Decisões de projeto estão em [ADR.md](ADR.md).

## Pacotes e dependências

| Pacote | Responsabilidade | Importa (pacotes internos) |
|---|---|---|
| `cmd/gex-dev-kit` | Entrada, opções, escolha entre TUI e CLI, sinais, adaptador `hook` | `app`, `agents`, `catalog`, `githooks`, `hooks`, `kit`, `ui/components`, `ui/screens`, `ui/terminal`, `ui/theme` |
| `internal/app` | Casos de uso: prévia, aplicar, doctor, rollback, git hooks, explicação de erro | `agents`, `catalog`, `githooks`, `hooks`, `kit` |
| `internal/kit` | Gravação no HOME ou no repositório, histórico, lock, doctor, rollback | `catalog` |
| `internal/githooks` | Repositório alvo, conflitos e scripts dos git hooks | `catalog` |
| `internal/hooks` | Hooks de agente: eventos, handlers e escopo GEX | nenhum |
| `internal/agents` | Versões fixas de Claude Code e Codex, instalador oficial em grupo de processos próprio | nenhum |
| `internal/catalog` | Dados embutidos: Core, skills GEX, skills Matt, referências e addons | nenhum |
| `internal/ui/screens` | Telas Bubble Tea, formulário huh, vitrine `ds`, executor de programa | `app`, `catalog`, `ui/components`, `ui/layout`, `ui/motion`, `ui/terminal`, `ui/theme` |
| `internal/ui/components` | Componentes do Design System (18 arquivos) | `ui/motion`, `ui/theme` |
| `internal/ui/layout` | Classes de tamanho, moldura comum e corte de texto sem quebrar códigos ANSI | `ui/theme` |
| `internal/ui/motion` | Molas e partículas, com regra de desligamento | nenhum |
| `internal/ui/terminal` | Detecção de terminal, modo, cor e tamanho | nenhum |
| `internal/ui/theme` | Tokens, tema, estilos para huh e bubbles | nenhum |

Desenho das dependências (seta = "importa"):

```
cmd/gex-dev-kit
 ├─> internal/app ───┬─> internal/kit ──────> internal/catalog
 │                   ├─> internal/githooks ─> internal/catalog
 │                   ├─> internal/agents
 │                   └─> internal/hooks
 ├─> internal/ui/screens ─┬─> internal/app
 │                        ├─> internal/catalog
 │                        ├─> internal/ui/components ─┬─> internal/ui/motion
 │                        │                           └─> internal/ui/theme
 │                        ├─> internal/ui/layout ─────> internal/ui/theme
 │                        ├─> internal/ui/motion
 │                        ├─> internal/ui/terminal
 │                        └─> internal/ui/theme
 ├─> internal/ui/terminal
 └─> internal/ui/theme
```

## Fluxo de dados

### Instalar

1. O usuário escolhe o Core, as especializações, as skills Matt e os hooks.
2. `app.PlanInstall` monta a prévia: a lista exata do que muda e o diff dos arquivos do usuário. Nada é gravado nesta etapa.
3. A pessoa confirma com `s` na TUI, com `--yes`, ou pelo prompt do modo acessível.
4. `app.ApplyInstall` grava os arquivos por `internal/kit`, usando `os.Root`.
5. A gravação entra no histórico com os hashes dos arquivos. O conteúdo anterior fica só nas 5 gravações mais recentes.
6. `doctor` compara o estado atual dos arquivos com o histórico.
7. `rollback` restaura a última gravação. Ele aborta se algum arquivo mudou depois da gravação.

### Raiz da gravação

- Sem `--repo`, a raiz é o HOME (ou o HOME passado em `--home`).
- Com `--repo`, a raiz é o git common dir do repositório, obtido com `git rev-parse --git-common-dir`. O histórico e os git hooks ficam nessa pasta.
- Os fluxos são os mesmos nos dois casos: prévia, confirmação, gravação, histórico, doctor e rollback.

### Git hooks

- `app.PlanGitHooks` monta a prévia. `app.ApplyGitHooks` grava só com `--repo` e confirmação.
- Os scripts instalados chamam o executável do kit com `exec "$bin" hook <nome>` (`internal/githooks/scripts.go`).

### Agentes

- `rollback-agents` restaura versões anteriores de Claude Code e Codex.
- O instalador oficial roda em grupo de processos próprio. Assim, o Ctrl+C do terminal não chega a ele.

## Design System

Resumo da decisão de forma (ADR-0003):

- O tema é um valor imutável, criado por `theme.New(Options)`. Não existe tema global mutável.
- Componente sem estado é função pura: recebe o tema e devolve texto.
- Componente com estado é uma struct com `Update` e `View`.
- Telas recebem funções (`Deps`) para falar com `kit` e `agents`. Elas não chamam esses pacotes direto. Assim, os testes não tocam no HOME real.
- O terminal é lido uma vez, em `internal/ui/terminal`, antes de qualquer tela.

Regras verificadas por `internal/ui/arch_test.go` (função `TestArquiteturaDaUI`):

1. Fora de `internal/ui/theme`, nenhum arquivo de `internal/ui` usa `lipgloss.NewStyle(`, `lipgloss.Color(` ou cor hexadecimal literal.
2. Nenhum pacote importa `charm.land/lipgloss/v2/compat`, porque ele consulta o terminal no `init`.
3. Nenhum arquivo de `internal/ui` chama `lipgloss.Print*`, `Fprint*` ou `Sprint*`. A saída passa por `terminal.Env.Writer`.
4. `cmd/gex-dev-kit` não importa `lipgloss` nem `huh`, salvo `accessible.go`.
5. `internal/ui` não importa `internal/kit` nem `internal/agents`.

## CLI e TUI

### Escolha do modo (`--ui=auto`)

| Situação | Modo | Cor | Animação |
|---|---|---|---|
| stdin e stdout são terminal, `TERM` diferente de `dumb`, sem `ACCESSIBLE` e sem `CI` | TUI | perfil detectado; `NO_COLOR` remove a cor | sim, salvo `--no-anim` ou `GEX_NO_ANIM` |
| Terminal com `ACCESSIBLE=1` | CLI interativo acessível, em português | sem cor | não |
| `TERM=dumb` | CLI | sem cor | não |
| stdin ou stdout sem terminal (pipe, redirecionamento) | CLI | sem ANSI | não |
| `CI` com valor, ou `--yes`, `--quiet`, `--json` | CLI | conforme as regras acima | não |

`--ui=tui` sem terminal, com `TERM=dumb` ou junto de `--yes`, `--quiet` ou `--json` sai com código 2.

### Códigos de saída

| Código | Quando |
|---|---|
| 0 | Sucesso, nada a alterar, resposta "não" na confirmação, catálogo, ajuda, versão, `doctor` sem problema do kit |
| 1 | Falha operacional; `doctor` com item do kit ausente, alterado, incompleto ou bloqueado; conflito de git hook; confirmação necessária sem terminal e sem `--yes` |
| 2 | Uso inválido: comando, opção, valor ou caminho relativo |
| 130 | SIGINT (Ctrl+C) |
| 143 | SIGTERM |
| 129 | SIGHUP (terminal fechado) |

Sinais recebidos durante uma gravação são adiados até a gravação terminar. Depois, o programa sai com o código do sinal.

### Equivalência entre TUI e CLI

| Ação na TUI | Comando equivalente na CLI |
|---|---|
| Instalar | `install` (com `--skill`, `--hook`, `--skip-agent-update`, `--home`) |
| Prévia com diff e confirmação | `install` (sem terminal, use `--yes`) |
| Doctor | `doctor` (com `--json`) |
| Rollback | `rollback` (com `--yes`) |
| Reverter agentes | `rollback-agents` (com `--yes`) |
| Catálogo de skills e hooks | `skills` e `hooks` (com `--json`) |
| Addons | `addons` (com `--json`) |
| Git hooks do projeto | `githooks --repo` (com `--yes`); `doctor --repo`; `rollback --repo` |
| Vitrine | `ds --ui=cli` |
| Ajuda | `help` ou `comando --help` |

A tela final da TUI mostra o comando equivalente. A função `app.EquivalentCommand` gera o texto.

## Mapa de cobertura Charm

A fonte real é `internal/ui/screens/coverage.go`. Cada linha liga um recurso de uma biblioteca a uma página da vitrine e a um teste.

| Biblioteca | Linhas em `coverage.go` | Observação |
|---|---|---|
| bubbletea | 10 | Modelo, tick, tamanho da janela, alt screen, mouse, foco, Batch e Sequence, Println, suspensão, erros tipados |
| bubbles | 13 | Spinner, progresso, tabela, lista, campos de texto, viewport, paginador, ajuda, seletor de arquivos, timer, cronômetro, cursor |
| huh | 10 | Nota, seleção, seletor de pasta, campo com validação, múltipla escolha, confirmação, texto, grupos, tema, modo acessível |
| lipgloss | 10 | Junção, posicionamento, bordas, gradiente, tabela, lista, árvore, cores adaptativas, medidas, compositor |
| harmonica | 2 | Mola e partícula |
| x | 4 | `x/ansi`, `x/term`, teatest e golden. Teatest e golden são só de teste (`TestOnly`) |

Os testes que verificam a tabela estão em `internal/ui/screens/coverage_test.go`: `TestCoverageMapIsComplete`, `TestCoverageTestOnlyEhSoParaTesteDeTUI` e `TestCoverageCobreCadaLibDaTabela`.

## Segurança

- Histórico e lock: arquivos novos com permissão 0600 e pastas novas com 0700 (`internal/kit/kit.go`). No Windows, o histórico é cifrado com DPAPI (`internal/kit/privacy_windows.go`).
- Lock com PID: o arquivo `.lock` guarda o PID e a hora de início. Se o processo não existe mais, o doctor pede para remover o lock. Se ele existe, a execução é bloqueada.
- Escrita por `os.Root`: a gravação não sai da raiz escolhida.
- Symlink recusado: o kit não segue link simbólico no caminho de gravação (`kit.ErrSymlink`).
- Git hooks: os scripts chamam o executável com `exec`. As mensagens do adaptador `hook` são textos fixos e não repetem o comando do commit (`cmd/gex-dev-kit/hook.go`).
- Addons: o kit mostra o comando oficial de instalação e não executa nenhum addon (`executed_by_kit` é `false` para todos).
- Dados pessoais: o teste `scripts/smoke.sh` procura dados pessoais nos arquivos instalados. Na execução desta revisão (macOS arm64), ele falhou (ver Matriz de plataformas).

## Matriz de plataformas

| Plataforma | Testes | Smoke | Build | Situação |
|---|---|---|---|---|
| macOS arm64 | `go test ./...` passou (14 pacotes) | falhou no passo de dados pessoais | ok | executado |
| Linux amd64 e arm64 | não executado | não executado | `go build` ok | só compilado |
| Windows amd64 e arm64 | não executado | não executado | `go build` ok; `go vet` ok (com `GOOS=windows`) | só compilado |
| Execução real em Windows e Linux | não verificado | não verificado | não verificado | não verificado |

Os resultados de macOS vêm de `go test ./...` e `go build` executados nesta revisão, e de `bash scripts/smoke.sh`, que falhou. Os resultados de Linux e Windows vêm só de `go build` e `go vet`.

## Como testar

- [`scripts/check.sh`](../scripts/check.sh): roda gofmt, a verificação de comentários em código, a verificação de rótulos de IA, `go test -race ./...`, `go build ./...` e o smoke. Rode antes de abrir PR.
- [`scripts/smoke.sh`](../scripts/smoke.sh) `CAMINHO_DO_BINARIO`: instala o kit em um HOME temporário e confere os arquivos gerados.
- Golden da vitrine: `go test ./internal/ui/theme -run TestGolden -update`. Use só quando o visual mudar de propósito e revise os arquivos `.golden` gerados antes de commitar.
- [`scripts/release.sh`](../scripts/release.sh) `vX.Y.Z`: gera binários, `SHA256SUMS` e `THIRD_PARTY_LICENSES.txt` em `dist/`. Sem `--publish`, nada é publicado.
