# Changelog

Todas as mudanças relevantes do Kit de Desenvolvimento GEX ficam neste arquivo. O formato segue o [Keep a Changelog](https://keepachangelog.com/pt-BR/1.1.0/) e as versões usam [SemVer](https://semver.org/lang/pt-BR/).

## [0.3.2] - 2026-10-25

As notas da publicação estão em [`docs/release-notes/v0.3.2.md`](docs/release-notes/v0.3.2.md).

### Adicionado

- **MCPs como addons opcionais**, instalados no chat com os comandos exatos, os riscos e a confirmação (ADR-0040): `context7` 4.4.4 (documentação atualizada de bibliotecas), `terraform` 1.4.1 (imagem oficial fixada por digest; exige Docker ligado), `brave-search` 2.1.4 e `mapbox-devkit` 0.8.2 (exigem chave de API; o Mapbox exige Node.js 22 ou mais novo). Cada versão foi conferida contra o registro do fornecedor.
- **Chaves de API de MCP pedidas no chat** (ADR-0041): campo mascarado, gravadas em texto puro em `~/.claude.json` e `~/.codex/config.toml`, ambos `0600`. Isso altera a regra anterior de o kit não gravar chaves, só para MCP. A chave não aparece na transcrição, na confirmação, em erros, no comando equivalente, em diff nem em golden, e a saída dos passos é redigida. Reinstalar substitui a chave. O `rollback` não remove MCP nem revoga a chave.
- `addons --json` lista `secrets` de cada addon: nome da variável, rótulo e URL, nunca o valor.
- O instalador de addons aceita binários exigidos em tempo de execução (`npx`, `docker`), um passo de pré-requisito que pula o addon inteiro nos dois clientes quando falha, e trata `claude mcp add` repetido como já configurado.

### Alterado

- **Claude Code e Codex passam à versão mais recente** pelos instaladores oficiais (ADR-0039, que substitui a ADR-0016). O kit consulta a versão publicada antes de instalar, usa `latest` nos instaladores e confere a versão observada. Sem rede, um agente já registrado não propõe atualização e um agente não registrado fica bloqueado com o motivo.
- Um Claude Code nativo oficial já instalado (com o link em `~/.local/share/claude/versions/`) é adotado em vez de bloqueado. O `rollback-agents` continua pela versão exata anterior.
- O banner mostra `versão mais recente: X` e o chat mostra `mais recente: X`.

### Corrigido

- O cartão do Doctor mostra primeiro os itens de atenção e fica em aviso.
- Sem registro do agente no kit, o banner diz `não instalado pelo kit`.
- Com dois agentes bloqueados, cada um mostra o próprio motivo.
- O ambiente dos addons não define mais `CLAUDE_CONFIG_DIR`. Com ela, o Claude Code gravava o MCP em `~/.claude/.claude.json`, um arquivo que ele não lê numa sessão normal, e o kit mostrava o MCP como instalado sem que ele aparecesse.
- Instalar só MCPs no Codex não pede mais para aprovar hooks em `/hooks`, porque MCPs não criam hooks.

## [0.3.1] - 2026-10-25

As notas da publicação estão em [`docs/release-notes/v0.3.1.md`](docs/release-notes/v0.3.1.md).

### Adicionado

- **Instaladores de uma linha** por sistema: `scripts/install.sh` (macOS e Linux) e `scripts/install.ps1` (Windows). Escolhem o binário certo para o sistema e a arquitetura, conferem o SHA-256 antes de instalar, instalam como `gex-dev-kit` (`~/.local/bin` no macOS e Linux; `%LOCALAPPDATA%\Programs\gex-dev-kit` no Windows, com a pasta no PATH do usuário) e mostram a linha de PATH quando falta. Substituem o passo a passo manual de `gh release download`, `sha256sum` e `chmod`. Exigem o `gh` autenticado, porque a release é privada. O `install.ps1` não foi executado em Windows.
- `scripts/test-install.sh`, com um `gh` falso, entra no `scripts/check.sh`.

### Corrigido

- **Colisão de executáveis no PATH:** quando há mais de uma cópia de `claude` ou `codex` no PATH, o motivo do bloqueio agora lista todos os caminhos encontrados, na ordem do PATH. Antes dizia só "múltiplos executáveis" e não mostrava quais remover. O bloqueio em si não mudou.
- **Frase sobre o Core no chat:** com o Core alterado, o cartão mostrava "0/7 skills · alterado" e logo depois o chat dizia "Ainda não há Core instalado". Agora o chat conta os arquivos que precisam de atenção e manda rodar `/doctor`.

### Alterado

- README reorganizado: instalação por sistema com badges de ícone (macOS, Linux, Windows), primeiro uso com diagrama do setup, receitas por tarefa, solução de problemas, glossário e referência recolhida. Texto reescrito em frases curtas; o maior parágrafo caiu de 124 para 37 palavras.
- Banner animado do README (`docs/assets/readme/banner.svg`): SVG sem fontes externas, estático quando o sistema pede movimento reduzido. Imagem para o social preview em `docs/assets/readme/social-preview.png`; o GitHub só aceita essa imagem em repositório público.

## [0.3.0] - 2026-10-25

As notas da publicação estão em [`docs/release-notes/v0.3.0.md`](docs/release-notes/v0.3.0.md).

### Adicionado

- **Chat com o assistente GEX no lugar das telas separadas.** Setup, catálogo, doctor, rollback, rollback dos agentes, git hooks e addons viram fluxos do chat, abertos por comandos com barra (`/setup`, `/catalogo`, `/doctor`, `/rollback`, `/agentes`, `/githooks`, `/addons`, `/menu`, `/voltar`, `/ajuda`, `/limpar`, `/sair`) e por uma paleta que abre ao digitar `/`. A entrada fica travada enquanto o assistente escreve ou executa.
- **Setup em sete etapas** (Harness, Agentes, Destino, Extras, Plugins, Plano e Instalação), com perguntas de múltipla escolha marcadas por espaço, navegador de pastas que aceita colar, `Esc` ou `/voltar` para voltar uma etapa e resumo antes de gravar.
- **Plugins instalados pelo kit:** `caveman`, `ponytail`, `superpowers` e `spec-kit`, no setup e em `/addons`, depois de mostrar os comandos exatos e os conflitos e de você confirmar. Passos sem equivalente de shell ficam como instrução manual (`superpowers` no Codex e o `specify init` de cada projeto).
- **Sete skills Core obrigatórias** (`gex-*`), instaladas em `~/.claude/skills/` e `~/.agents/skills/`, e **sete especializações opcionais** nos grupos Dados e Plataforma e Produto e Interface.
- **Explicação de cada skill e hook** no catálogo: o que faz, quando usar e uma analogia. Um teste de completude falha se um item novo não tiver explicação.
- **Interface do chat:** status line com GEX, harness, projeto, branch Git e andamento do setup; cartões de ferramenta com checklist de etapas ao vivo; escrita por palavras com cadência natural; barra lateral a partir de 120×30; mouse em opções, botões e páginas.
- **Listas em páginas** (`‹ página i/n ›`) em vez de rolagem de itens; a rolagem da conversa continua.
- **Hooks de política** de agente (`gex-pre-pr`, `gex-sensitive-action-warning`, além de `gex-conventional-commits`), com escopo por organização GEX, e **git hooks por projeto** (`githooks`, com `--repo`).
- **Comandos e opções:** `addons`, `githooks`, `ds` (vitrine do Design System), `--json` em `skills`, `hooks`, `addons` e `doctor`, `--ui=auto|tui|cli`, `--no-anim`, `--quiet`, `--repo` e códigos de saída documentados (129, 130 e 143 para sinais).
- Schema `gex-dev-kit.skills/v1` de `skills --json`. Ele estreia nesta versão e as especializações não têm campo `credentials`.
- Scripts `scripts/check.sh`, `scripts/smoke.sh`, `scripts/demo.sh` e `scripts/ux-audit-env.sh`; `scripts/release.sh` agora lê as notas de `docs/release-notes/vX.Y.Z.md` e recusa uma versão sem elas.
- Documentação: arquitetura, registro de decisões (ADR), glossário, auditoria de UX e guia em vídeo.

### Alterado

- `rollback` mostra a prévia com o diff e pede confirmação; em automação, use `--yes`. A prévia de `install` mostra o diff com dados sensíveis mascarados.
- `doctor` deixa de sair com código 1 por agente que o kit não gerencia; esse caso vira aviso.
- Reinstalar sempre atualiza o executável dos hooks. O merge de hooks é aditivo: hooks que já existiam não são removidos.
- O histórico guarda o conteúdo fora do JSON e alcança as 5 gravações mais recentes.
- Skills Matt Pocock: 18 cópias offline do commit `5beb35b0` (eram 17).
- Limite de tamanho dos binários sobe de 14 MiB para 16 MiB (ADR-0037). O maior binário, `windows_amd64`, tem 14,1 MiB.

### Corrigido

- Falso positivo do hook de commit e contornos do hook por `wrapper`, `eval`, `heredoc`, chaves e alvo de `rm`.
- Erros de uso saem com código 2 e sugerem nomes válidos; o JSON do `doctor` ficou consistente.

### Removido

- As telas separadas de home, setup, instalação, doctor, rollback, catálogo, addons e git hooks, e os goldens delas. O modo CLI, o modo acessível e `ds` continuam.
- As skills de API de gateway (PagAmerican, BuyGoods e ClickBank) e a gravação de chave de API nas configurações dos agentes. Elas existiram só durante o desenvolvimento e não fazem parte de nenhuma release (ADR-0035).

### Segurança

- Plugins são código de terceiros: só rodam depois da sua confirmação, sem shell, com timeout de 5 minutos por passo e entrada vazia. O rollback do kit não os desfaz.
- Revisão de segurança do instalador de plugins e correções do que ela provou: o instalador roda em grupo de processos próprio e sem terminal de controle (o timeout encerra os filhos), recebe só um conjunto reduzido de variáveis de ambiente e um HOME realmente isolado, roda no diretório do HOME alvo, usa versões fixadas por commit ou tag e tem `Cancelar` como primeira opção na confirmação (ADR-0038).
- Riscos aceitos: caveman e ponytail no Claude Code usam a tag da release (a CLI não aceita commit) e o `superpowers` segue o pin do marketplace oficial da Anthropic. Da saída dos instaladores só saem as sequências CSI e OSC e os padrões `token=`, `key=` e `Bearer`; outros formatos de segredo e caracteres de controle chegam ao terminal.

## [0.2.0] - 2026-10-21

### Alterado

- Interface interativa revisada: tela inicial com menu, prévia da instalação e `doctor` mais completo.
- O README passa a mostrar como conferir o `SHA256SUMS` ao baixar a release.

## [0.1.0] - 2026-10-21

Primeira entrega.

### Adicionado

- Comandos `install`, `doctor`, `rollback`, `skills` e `hooks`. `install` mostra a prévia dos arquivos e das versões afetadas e pede confirmação (`--yes` confirma em automação).
- Core com instruções globais em `~/.claude/CLAUDE.md` e `~/.codex/AGENTS.md`.
- 17 skills Matt Pocock opcionais e o hook `gex-conventional-commits`.
- Claude Code `2.1.306` e Codex `0.163.0` fixados; a CLI só instala ou atualiza executáveis standalone próprios, pelos instaladores oficiais com versão exata.
- Binários para macOS, Linux e Windows, em `arm64` e `amd64`.

[0.3.2]: https://github.com/Gex-Development/gex-dev-kit/compare/v0.3.1...v0.3.2
[0.3.1]: https://github.com/Gex-Development/gex-dev-kit/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/Gex-Development/gex-dev-kit/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/Gex-Development/gex-dev-kit/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/Gex-Development/gex-dev-kit/releases/tag/v0.1.0
