# Kit de Desenvolvimento GEX

Prepara Claude Code e Codex para o trabalho GEX: instala instruções e sete skills Core no HOME, oferece skills e hooks opcionais e conduz a configuração por chat ou CLI.

![Versão v0.3.0](https://img.shields.io/badge/Vers%C3%A3o-v0.3.0-6f42c1?style=flat-square) ![Go 1.25.8](https://img.shields.io/badge/Go-1.25.8-00ADD8?style=flat-square) ![Plataformas macOS Linux Windows](https://img.shields.io/badge/Plataformas-macOS%20%7C%20Linux%20%7C%20Windows-555555?style=flat-square) ![Arquiteturas arm64 amd64](https://img.shields.io/badge/Arquiteturas-arm64%20%7C%20amd64-555555?style=flat-square) ![Interface TUI CLI](https://img.shields.io/badge/Interface-TUI%20%2B%20CLI-2088aa?style=flat-square) ![Agentes Claude Code e Codex](https://img.shields.io/badge/Agentes-Claude%20Code%20%2B%20Codex-2088aa?style=flat-square) ![Conventional Commits](https://img.shields.io/badge/Commits-Conventional-2088aa?style=flat-square)

[Instalar](#instalação) · [Primeiro uso](#primeiro-uso) · [Comandos](#comandos-no-chat) · [Plugins](#plugins-opcionais) · [Desenvolvimento](#desenvolvimento) · [Changelog](CHANGELOG.md)

## Visão geral

- **Core obrigatório:** instruções globais para Claude Code/Codex e sete skills `gex-*`. Opcionalmente, sete especializações, 18 skills Matt Pocock, hooks de agente e git hooks por repositório.
- **Sem alterações implícitas no projeto:** o Core não modifica MCPs nem arquivos de projeto. Git hooks exigem seleção do repositório; plugins de terceiros só rodam após prévia e confirmação no **chat**. A CLI `install` não instala plugins.
- **Antes de gravar:** prévia de arquivos, versões, conflitos e diff; `--yes` dispensa a pergunta, não a prévia. Há histórico das cinco gravações mais recentes e `rollback` com prévia e confirmação. O rollback do kit **não remove plugins**.
- **Execução dos plugins:** comandos sem shell, em grupo de processos próprio, HOME escolhido, ambiente reduzido e timeout de cinco minutos por passo. Ainda são código de terceiros executado com seus privilégios: revise fonte, licença e comandos antes de confirmar. O chat segue roteiro local; não chama um modelo de IA.

## Instalação

A [release privada v0.3.0](https://github.com/Gex-Development/gex-dev-kit/releases) exige acesso à organização Gex-Development e autenticação no `gh`. Há seis binários: `darwin_{arm64,amd64}`, `linux_{arm64,amd64}` e `windows_{arm64,amd64}`. Os alvos Linux e Windows foram compilados e verificados com `go vet`, **não exercitados em execução**; os testes rodaram em macOS. Confira o checksum antes de executar.

```sh
# Linux amd64; troque linux_amd64 pelo alvo do seu sistema.
gh release download v0.3.0 --repo Gex-Development/gex-dev-kit --dir .
sha256sum -c SHA256SUMS
chmod +x gex-dev-kit_v0.3.0_linux_amd64
./gex-dev-kit_v0.3.0_linux_amd64
```

No macOS, use `shasum -a 256 -c SHA256SUMS` e execute `gex-dev-kit_v0.3.0_darwin_arm64` ou `darwin_amd64`. No PowerShell, baixe o binário `windows_arm64.exe` ou `windows_amd64.exe`, compare o resultado de `Get-FileHash .\gex-dev-kit_v0.3.0_windows_amd64.exe -Algorithm SHA256` com a linha correspondente de `SHA256SUMS` e execute o `.exe`. Avisos das dependências acompanham a release em `THIRD_PARTY_LICENSES.txt`.

### Primeiro uso

Execute o binário em um terminal interativo. Boas-vindas → `Enter` ou clique → verificação do ambiente → digite `/setup`. O roteiro tem **Harness, Agentes, Destino, Extras, Plugins, Plano e Instalação**. Escolha Claude Code/Codex, revise destino e opcionais, confira comandos, conflitos e diff e confirme **Sim, aplicar**. **Agora não** na pergunta sobre agentes evita atualizar seus executáveis. A entrada fica travada enquanto o assistente escreve ou executa uma ferramenta.

Para experimentar sem tocar no seu HOME, compile e use um HOME descartável:

```sh
go build -o /tmp/gex-dev-kit-piloto ./cmd/gex-dev-kit
PILOT_HOME="$(mktemp -d)"
HOME="$PILOT_HOME" /tmp/gex-dev-kit-piloto
/tmp/gex-dev-kit-piloto doctor --home "$PILOT_HOME"
/tmp/gex-dev-kit-piloto rollback --home "$PILOT_HOME"
```

Se o ambiente exportar `CI` ou `NO_COLOR`, remova-os ao testar as cores da TUI: `env -u CI -u NO_COLOR HOME="$PILOT_HOME" /tmp/gex-dev-kit-piloto`. `scripts/demo.sh` também abre a TUI com HOME descartável.

### Demonstração

Trecho de uma sessão real da v0.3.0 em terminal, com HOME descartável (sem Core instalado); resultados de agentes dependem do PATH de cada máquina:

```text
◆ Olá! Eu sou o assistente do Kit GEX. Vou configurar seu ambiente de desenvolvimento com você.

╭─ ✓ Verificar ambiente ~ ─────────────────────────────────────────────────────────────────────────╮
│ ✓ HOME         ~                                                                                 │
│ ! Claude Code  bloqueado · múltiplos executáveis no PATH; resolver colisão antes da instalação   │
│ ! Codex        bloqueado · múltiplos executáveis no PATH; resolver colisão antes da instalação   │
│ i Core         0/7 skills · não instalado                                                        │
╰───────────────────────────────────────────────────────────────────────────────────────────── 4s ─╯

◆ Ainda não há Core instalado neste HOME.

◆ Digite / para ver o que posso fazer. Para configurar o ambiente, use /setup.
```

Sem capturas antigas: elas mostram a interface anterior. No chat, a barra mostra harness detectado, projeto e branch Git (com contagens `*N` modificados e `?N` não rastreados); o HOME de instalação aparece quando difere do HOME real. A borda do prompt indica `Core N/7` ou progresso `Etapa N/7`. `GEX_NO_POWERLINE=1` troca os separadores por `›`; modo Plain usa ASCII. Com animação, texto chega por palavras (~55 caracteres/s, com pausas e teto aproximado de 9 s para textos longos), há intervalo entre mensagens e cartões de ferramenta exibem fases, resultado e duração. `Enter` ou espaço pula a animação; `--no-anim` exibe tudo imediatamente.

## Comandos no chat

Digite `/` para abrir a paleta (`↑`/`↓` escolhem, `Tab` completa). Também aceita buscas como “instalar”, “skill”, “verificar”, “desfazer”, “agente”, “git hook” e “addon”.

| Comando | Ação |
|---|---|
| `/setup` | Configurar ou atualizar o kit |
| `/catalogo` | Explorar skills e hooks com “O que faz”, “Quando usar” e “Analogia” |
| `/doctor` | Verificar a integridade |
| `/rollback` | Desfazer a última instalação |
| `/agentes` | Restaurar versões anteriores dos agentes |
| `/githooks` | Configurar git hooks |
| `/addons` | Ver e instalar plugins de terceiros após revisão e confirmação |
| `/menu` | Mostrar as opções como lista |
| `/voltar` | Voltar uma etapa |
| `/ajuda` | Listar comandos e atalhos |
| `/limpar` | Limpar a conversa atual |
| `/sair` | Encerrar |

**Atalhos:** espaço ou `1`–`9` marca itens, `Enter` confirma (ou escolhe o item sob o cursor quando obrigatório). Clique em **Confirmar (N)** ou **Pular (nenhuma marcada)** para avançar; Extras vazios não instalam opcionais. `Esc` limpa a entrada ou volta; `/voltar` volta etapa. `←`/`→`, `PageUp`/`PageDown` ou clique nas setas trocam a página de escolhas longas (`‹ página i/n ›`); marcações persistem, números valem na página atual. Com texto digitado, setas movem o cursor. Sem escolha paginada, roda e `PageUp`/`PageDown` rolam a conversa. Paleta e navegador de pastas mostram `… mais N`: digite para filtrar. `Tab` completa caminhos, `Enter` abre pasta destacada e é possível colar pelo terminal. `Ctrl+O` ou clique no título expande um cartão; `Ctrl+C` sai. O catálogo separa Core, Engenharia, Produtividade, Especializações e Referências.

## CLI e automação

```sh
gex-dev-kit skills
gex-dev-kit install --skill tdd --hook gex-conventional-commits
gex-dev-kit doctor
gex-dev-kit rollback
```

| Comando | Para que serve |
|---|---|
| `install` | Mostra prévia, pede confirmação e grava Core e escolhas |
| `doctor` | Confere HOME (`--home`) ou git hooks de repositório (`--repo`) |
| `rollback` | Desfaz última gravação do HOME ou repositório, com prévia |
| `rollback-agents` | Restaura versão anterior de Claude Code/Codex, com confirmação |
| `skills` | Lista Core, opcionais e referências |
| `hooks` | Lista hooks de agente, git hooks e referências |
| `addons [NOME]` | Lista plugins ou mostra comando oficial; **não instala** |
| `githooks --repo CAMINHO` | Prévia dos git hooks; grava com confirmação |
| `ds` | Vitrine do Design System (`--ui=cli` mostra em texto) |
| `version` | Mostra versão |
| `help [COMANDO]` | Ajuda geral ou por comando |

| Situação | Interface |
|---|---|
| Terminal interativo, sem opção extra | Chat (TUI) |
| `--yes`, `--quiet`, `--json` ou `--ui=cli` | CLI |
| `--ui=tui` sem terminal interativo | Erro de uso (código 2) |
| `ACCESSIBLE=1` ou `TERM=dumb` | Menu e confirmação em texto |
| `NO_COLOR=1` em terminal interativo | Chat sem cores |

Opções globais: `--ui=auto|tui|cli`, `--no-anim`, `--quiet`, `--json` (em `skills`, `hooks`, `addons`, `doctor`), `-h`/`--help` e `-v`/`--version`. `--home` (em `install`, `doctor`, `rollback`, `rollback-agents`, `githooks`) e `--repo` aceitam apenas caminho absoluto. `--skill` e `--hook` podem ser repetidos. Sem Extras, `install` grava o Core e tenta atualizar os agentes elegíveis; `--skip-agent-update` evita rede e mudanças nos binários dos agentes. `--yes` confirma após a prévia. `--json` imprime apenas JSON no stdout; `skills --json` usa `gex-dev-kit.skills/v1`, introduzido na v0.3.0, sem `credentials` nas especializações (skills de API de gateway não fazem parte do catálogo; o kit não pede nem grava essas chaves).

| Saída | Significado |
|---|---|
| `0` | Concluído sem falha |
| `1` | Falha de execução ou verificação, por exemplo arquivo alterado no `doctor` |
| `2` | Uso inválido: comando/flag desconhecido, valor ausente, caminho relativo, skill/hook desconhecido, `--repo` sem Git ou `--run`/`--pre-commit-run` sem hook correspondente |
| `129` | SIGHUP (terminal fechado) |
| `130` | Ctrl+C (SIGINT) |
| `143` | SIGTERM |

Durante gravação, um sinal aguarda a gravação em curso terminar antes de encerrar. Agente externo não gerenciado gera aviso no `doctor`, sem alterar seu código de saída.

## O que o kit instala

**Core:** instruções em `~/.claude/CLAUDE.md` e `~/.codex/AGENTS.md`; sete skills em `~/.claude/skills/` e `~/.agents/skills/`:

`gex-abrir-issue`, `gex-auto-revisao`, `gex-fluxo-desenvolvimento`, `gex-preparar-pr`, `gex-regras-de-negocio`, `gex-revisar-pr`, `gex-verificar-mudanca`.

**Especializações opcionais (7):**

| Grupo | Skills |
|---|---|
| Dados e Plataforma | `gex-local-spark-tests`, `gex-canonical-prod-readiness`, `gex-hub-gold-refresh-prod`, `gex-gateway-audit`, `gex-pr-review` |
| Produto e Interface | `gex-design-system`, `gex-hub-design-system` |

Algumas especializações exigem credenciais/permissões externas e autorização própria para agir em produção; copiar a skill não executa essas ações. `gex-pr-review` requer `gex-local-spark-tests`.

**Matt Pocock (18, opcionais):** cópias offline do [commit `5beb35b0`](https://github.com/mattpocock/skills/tree/5beb35b040e3de97e54fa9e386d7e42d15c295c2), sob [MIT da fonte](https://github.com/mattpocock/skills/blob/5beb35b040e3de97e54fa9e386d7e42d15c295c2/LICENSE): `code-review`, `codebase-design`, `diagnosing-bugs`, `domain-modeling`, `grill-me`, `grill-with-docs`, `grilling`, `handoff`, `improve-codebase-architecture`, `pr`, `prototype`, `research`, `retro`, `tdd`, `teach`, `to-questionnaire`, `wait-what`, `writing-for-agents`. Instalá-las apenas copia arquivos; usar skills que criam issues, labels, branches ou modificam projetos exige autorização separada. Outras 11 (`setup-matt-pocock-skills`, `implement-spec`, `ask-matt` etc.) são **somente referências**, não instaláveis.

**Hooks de agente (4, por escolha):** `gex-conventional-commits` (bloqueia mensagem literal fora do padrão), `gex-sensitive-action-warning` (pede confirmação de ação sensível), `gex-pre-pr` (confirma PR), `gex-session-start` (contexto ao iniciar, opt-in). Os hooks de política só agem com remote de `Gex-Development`, `Gex-Integrations`, `gex-data-architecture-org`, `gex-data-architecture` (sem distinguir maiúsculas) ou `git config gex.policy true`; `false` desliga. Organizações extras entram no arquivo `orgs` da configuração do usuário; `doctor` mostra a lista efetiva. No Codex, os dois hooks que pediriam confirmação não são suportados; aprove hooks instalados em `/hooks`. Em `claude -p`, perguntas de `gex-sensitive-action-warning` e `gex-pre-pr` viram negação: não instale esses hooks em máquinas de automação. Há ainda **25 referências** de hooks de terceiros, não instaladas.

**Git hooks (3, por repositório):** `gex-commit-msg` valida Conventional Commits, `gex-pre-commit` procura possíveis credenciais no diff e roda comandos rápidos documentados pelo projeto, `gex-pre-push` roda comandos documentados passados com `--run`. O hook de agente `gex-conventional-commits` não substitui validação por Git nem proteção remota. Configurações existentes são preservadas; merge de hooks é aditivo e reinstalação atualiza o executável dos hooks.

## Plugins opcionais

Só a etapa **Plugins** de `/setup` ou `/addons` no chat instala plugins. Escolha Claude Code e/ou Codex, confira comandos e conflitos com o Core e confirme. A CLI `addons` apenas consulta o catálogo; a CLI `install` não instala plugins.

| Plugin | Para quê | Instalação pelo chat | Passo manual |
|---|---|---|---|
| `caveman` | Respostas curtas | Claude Code por tag `v4.0.1`; Codex via `npx` por commit | Aprovar hook no Codex em `/hooks` |
| `ponytail` | Menor mudança completa | Claude Code por tag `v5.3.2`; Codex por commit | Aprovar hooks no Codex em `/hooks` |
| `superpowers` | Brainstorming, TDD, planos e revisão | Claude Code via marketplace oficial | No Codex, `/plugins` → superpowers → Install Plugin |
| `spec-kit` | Fluxo de especificação | `uv tool install` por commit verificado | `specify init . --integration claude` ou `codex` em cada projeto |

As tags de `caveman` e `ponytail` no Claude Code **não são** os commits conferidos do catálogo; o marketplace não aceita commit como referência. O pin de `superpowers` depende do marketplace oficial e pode mudar. `spec-kit` requer Python 3.11+ e `uv`. Instaladores ausentes viram instrução manual. Um plugin pode alterar `settings.json`/`hooks.json` após o Core, causando conflito no `rollback` ou aviso no `doctor`. O kit reduz as variáveis passadas (tokens como `GH_TOKEN`/`NPM_TOKEN` não chegam ao instalador), mas a filtragem da saída não cobre todos os formatos de segredo/caracteres de controle. Revise a [análise completa de segurança e fixação nas notas v0.3.0](docs/release-notes/v0.3.0.md). **Rollback não desfaz plugins**: remova-os no próprio Claude Code/Codex.

## Agentes e git hooks por projeto

Versões fixadas nesta release: **Claude Code 2.1.306** e **Codex 0.163.0**. O kit só instala/atualiza executáveis standalone que gerencia, no perfil do usuário, por instaladores oficiais e versão exata. Brew, npm, winget, apt, dnf, binários não gerenciados, PATH ambíguo e atualização nativa fora do pin bloqueiam essa operação: o Core ainda pode ser instalado, mas o comando retorna erro e não afirma atualização. `rollback-agents` só reverte upgrade gerenciado com versão anterior registrada; a primeira instalação não é desinstalada automaticamente. Login e modelo não são alterados.

`githooks --repo CAMINHO_ABSOLUTO` mostra prévia e grava somente após confirmação; `rollback --repo CAMINHO_ABSOLUTO` desfaz a gravação. Use `--hook gex-pre-commit --pre-commit-run "go vet ./..."` ou `--hook gex-pre-push --run "go test ./..."` apenas quando esses comandos estiverem documentados no projeto.

## Desenvolvimento

```sh
go test ./...
go vet ./...
go build -o /tmp/gex-dev-kit-piloto ./cmd/gex-dev-kit
./scripts/smoke.sh /tmp/gex-dev-kit-piloto
```

- `scripts/check.sh`: formato, comentários, vet, testes, build e smoke local.
- `scripts/smoke.sh CAMINHO_DO_BINARIO`: instalação, doctor, hook e rollback em HOME descartável.
- `scripts/demo.sh`: compila e abre a TUI em HOME descartável.
- `scripts/release.sh vX.Y.Z`: compila seis alvos (máximo 16 MiB cada), gera `dist/SHA256SUMS` e avisos das dependências; exige `docs/release-notes/vX.Y.Z.md`. Não publica sem `--publish` e confirmação explícita.

| Pasta | Responsabilidade |
|---|---|
| `cmd/gex-dev-kit` | Entrada da CLI, comandos, sinais e JSON |
| `internal/app` | Casos de uso compartilhados por CLI e TUI |
| `internal/catalog` | Catálogo Core, opcionais, plugins e hooks |
| `internal/kit` | Gravação segura e histórico no HOME |
| `internal/agents` | Controle dos binários dos agentes |
| `internal/hooks` | Hooks de agente |
| `internal/githooks` | Git hooks por projeto |
| `internal/ui` | Tema, componentes e chat; `ds` separado |

Commits seguem Conventional Commits. Não há CI remoto nesta fase. Consulte [CHANGELOG](CHANGELOG.md), [notas da v0.3.0](docs/release-notes/v0.3.0.md), [decisões (ADR)](docs/ADR.md), [arquitetura](docs/ARCHITECTURE.md), [contexto](docs/CONTEXT.md) e [auditoria de UX](docs/AUDITORIA-UX.md) para histórico e detalhes.

Fontes de compatibilidade: [hooks Claude Code](https://code.claude.com/docs/en/hooks), [hooks Codex](https://developers.openai.com/codex/hooks), [instalação Claude Code](https://code.claude.com/docs/en/setup), [instalação Codex](https://developers.openai.com/codex/cli).
