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
| `internal/ui/screens` | Chat Bubble Tea (`RunChat`), fluxos conversacionais, vitrine `ds` e executor de programa | `app`, `catalog`, `ui/components`, `ui/layout`, `ui/motion`, `ui/terminal`, `ui/theme` |
| `internal/ui/components` | Componentes do Design System, incluindo mensagens, escolhas, ferramentas, editor e barra lateral do chat | `ui/layout`, `ui/motion`, `ui/theme` |
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

1. `/setup` abre a jornada com o roteiro “harness, destino, extras, plugins, revisão e instalação”. A barra lateral mostra sete etapas fixas: Harness, Agentes, Destino, Extras, Plugins, Plano e Instalação. Depois da escolha dos harnesses, uma frase os reconhece e um banner por Claude Code ou Codex substitui a verificação longa em blocos separados. Se houver agente elegível, o chat pergunta sobre atualização; depois pergunta o destino.
2. Em Extras, “Além do Core (obrigatório), o que mais você quer instalar?” oferece um item por grupo. Só os grupos marcados abrem suas opções; sem grupos selecionados não há skills ou hooks opcionais. Caminhos podem ser digitados ou escolhidos no navegador de pastas. O resumo em várias linhas exibe harness, destino, quantidades opcionais e agentes.
3. `app.PlanInstall` monta a prévia: lista e diff do que muda; nada é gravado nesta etapa.
4. A pessoa confirma **Sim, aplicar** no chat, com `--yes` na CLI, ou pelo prompt do modo acessível.
5. `app.ApplyInstall` grava os arquivos por `internal/kit`, usando `os.Root`.
6. A gravação entra no histórico com os hashes dos arquivos. A retenção de conteúdo anterior é limitada às 5 gravações mais recentes.
7. `doctor` compara o estado atual dos arquivos com o histórico; `rollback` restaura a última gravação e os arquivos de configuração anteriores. Ele aborta se algum arquivo mudou depois da gravação.

`RunChat` recebe os casos de uso como funções em `ChatDeps`; o fluxo setup chama `PlanInstall`, mostra o resultado em um bloco de ferramenta, pede confirmação e só então chama `ApplyInstall`. A gravação em curso termina antes de sair, mesmo após sinal. `install` abre o mesmo fluxo quando usado como TUI.

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
- A transcrição usa `ChatMessage` sem caixas, `ToolBlock` para operações, `Choice` para opções numeradas, `Palette` para comandos, `PathPicker` para pastas, `PromptBox` para texto com borda superior de status e `WorkRow` para spinner e tempo. `StatusLine`, `HarnessBanner` e `Sidebar` compõem o contexto; estilos vêm de `theme`.

Regras verificadas por `internal/ui/arch_test.go` (função `TestArquiteturaDaUI`):

1. Fora de `internal/ui/theme`, nenhum arquivo de `internal/ui` usa `lipgloss.NewStyle(`, `lipgloss.Color(` ou cor hexadecimal literal.
2. Nenhum pacote importa `charm.land/lipgloss/v2/compat`, porque ele consulta o terminal no `init`.
3. Nenhum arquivo de `internal/ui` chama `lipgloss.Print*`, `Fprint*` ou `Sprint*`. A saída passa por `terminal.Env.Writer`.
4. `cmd/gex-dev-kit` não importa `lipgloss` nem `huh`, salvo `accessible.go`.
5. `internal/ui` não importa `internal/kit` nem `internal/agents`.

### Chat de harness

`gex-dev-kit` sem subcomando abre boas-vindas com verificação do ambiente; `Enter` ou clique inicia a conversa ociosa. Os intents de `RunChat` abrangem setup, catálogo, doctor, rollback, rollback de agentes, git hooks e addons. Não há menu permanente: `/` abre uma paleta filtrável com `/setup`, `/catalogo`, `/doctor`, `/rollback`, `/agentes`, `/githooks`, `/addons`, `/menu`, `/voltar`, `/ajuda`, `/limpar` e `/sair`. `/menu` mostra a lista de opções sob demanda. Palavras-chave fixas também escolhem fluxos; não há modelo de IA interpretando texto livre. `Esc` limpa a entrada ou retorna à pergunta anterior; na primeira pergunta cancela o fluxo e volta ao estado ocioso. `/voltar` tem a mesma navegação. O catálogo usa `catalog.Explain` para mostrar “O que faz”, “Quando usar” e “Analogia”; `TestExplainCoversCatalogWithoutOrphans` exige explicação para cada Core, skill Matt, especialização e hook.

`chatLayout` e `placeChoice` compartilham o cálculo da altura das listas: Choice usa páginas completas, incluindo arte, dica, botão e controles; Palette e PathPicker limitam linhas visíveis com `… mais N`. A transcrição mantém ao menos quatro linhas. `←`/`→` e `PageUp`/`PageDown` mudam a página apenas quando há Choice paginado e campo vazio; fora disso, `PageUp`/`PageDown` e a roda preservam a rolagem da transcrição. O catálogo divide Skills em Core, Engenharia, Produtividade e dois grupos de Especializações; Referências em skills externas e hooks candidatos.

O motor pré-calcula uma vez o cronograma de cada `say`: palavras inteiras aparecem de uma vez, com intervalos determinísticos variados, base de 55 runas/s, ajuste proporcional para textos longos caberem em cerca de 9 s e pausas escaladas de pontuação (sem pausa dentro de versões como `2.1.306`). Um tick próprio de 60 fps revela palavras pelo relógio; a mola, o spinner e as fases seguem os frames de 30 fps. Travadas do terminal limitam o avanço do stream a 250 ms por frame. A batida entre mensagens dura 450 ms, mais 250 ms quando a anterior excede 300 runas. Com `Motion=false`, stream, batida, fases e retenção são instantâneos. `WorkRow` mostra spinner, tempo e brilho móvel no rótulo da fase atual a cada 1 s. O título do `ToolBlock` não muda; dentro do cartão, cada fase tem uma linha (concluída, em andamento ou futura). Na verificação inicial, HOME, Claude Code, Codex e Core substituem as fases por resultados reais na mesma posição. Nas demais ferramentas, o checklist fica até terminar a retenção, quando o resultado substitui a lista. A janela das fases dura no máximo 6 s; progresso real chega imediatamente. `Busy()` cobre apenas gravação real, não a retenção visual. `locked()` trava a entrada durante stream, batida, ferramenta ou fila pendente: não acumula texto digitado; `Enter` e espaço pulam a animação atual, inclusive batida e retenção. Na escolha múltipla, espaço e dígitos marcam opções; `Enter` confirma os itens marcados (sem marca em pergunta obrigatória, confirma a opção sob o cursor). O botão **Confirmar (N)** ou **Pular (nenhuma marcada)** também confirma por clique. Perguntas de caminho usam `PathPicker`: `Tab` completa e `Enter` desce na pasta destacada; colagem por `Cmd+V` ou `Ctrl+V` passa pela entrada do terminal. `Ctrl+O` e cliques na primeira linha do cartão alternam a expansão da ferramenta; a roda navega opções ou rola a transcrição. No setup, um banner identifica cada harness Claude Code ou Codex escolhido. Em largura ≥ 120 e altura ≥ 30, a barra lateral de 32 colunas exibe ambiente e as sete etapas fixas do setup; no layout compacto, ela é omitida.

`Enter` e clique só atravessam as boas-vindas após leitura de Core e agentes. Intents diretos também aguardam essas leituras antes de preencher o resultado de `Verificar ambiente`, com limite de 8 s; `Enter` não pula essa espera. O `ToolBlock` é cartão de cantos arredondados com largura máxima de 100 colunas e tempo na borda inferior. Exibe linhas estruturadas (`Rows`) antes de texto (`Lines`), sem conectores: estados coloridos, detalhes alinhados, conteúdo cortado ao caber. Ambos os tipos compartilham o colapso após seis linhas; abaixo de 20 colunas, não há moldura, e Plain usa ASCII.

A barra escura de contexto (`components.StatusLine`) mostra GEX, primeiro harness detectado e versão, caminho do projeto (diretório atual) e branch Git quando disponível (com contagens de modificados e não rastreados). O HOME de instalação só aparece quando difere do HOME real. A borda superior rosa da caixa de entrada (`PromptBox`) reúne as métricas: `Core N/7` e versão do kit fora do setup, ou porcentagem e `Etapa N/7` durante o setup: Harness, Agentes, Destino, Extras, Plugins, Plano, Instalação. Não há divisória separada acima da transcrição. A barra usa separadores Powerline por padrão; `GEX_NO_POWERLINE=1` usa `›` e o perfil Plain usa ASCII. Em largura restrita, métricas e segmentos são reduzidos para caber.

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
| Setup e instalação no chat | `install` (mesmos `--skill`, `--hook`, `--skip-agent-update`, `--home`) |
| Prévia com diff e confirmação | `install` (sem terminal, use `--yes`) |
| Doctor | `doctor` (com `--json`) |
| Rollback | `rollback` (com `--yes`) |
| Reverter agentes | `rollback-agents` (com `--yes`) |
| Catálogo de skills e hooks | `skills` e `hooks` (com `--json`) |
| Addons | `addons` consulta catálogo; instalação só na TUI (`/setup` ou `/addons`) |
| Git hooks do projeto | `githooks --repo` (com `--yes`); `doctor --repo`; `rollback --repo` |
| Vitrine | `ds --ui=cli` |
| Ajuda do chat (`/ajuda`) | `help` ou `comando --help` |

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
- Addons: `ChatDeps.Addons` planeja com `app.PlanAddons` (instalar) ou `app.PlanAddonRemoval` (remover, pelo `/addons`) e executa com `app.InstallAddons` depois da confirmação, com HOME escolhido; argv sem shell, timeout e passos manuais visíveis. `Ctrl+C` espera terminar a gravação. O rollback do kit não os desfaz; a remoção é uma ação separada e confirmada.
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
- Golden do chat: `go test ./internal/ui/screens -run TestGoldenChat -update`. Use só quando o visual mudar de propósito e revise o `git diff` dos arquivos `.golden` antes de commitar.
- Golden da paleta: `go test ./internal/ui/theme -run TestGolden -update`. Use só quando o visual mudar de propósito e revise o `git diff` dos arquivos `.golden` antes de commitar.
- [`scripts/release.sh`](../scripts/release.sh) `vX.Y.Z`: gera binários (até 16 MiB cada), `SHA256SUMS` e `THIRD_PARTY_LICENSES.txt` em `dist/` e exige as notas em `docs/release-notes/vX.Y.Z.md`. Sem `--publish`, nada é publicado.
