# Kit de Desenvolvimento GEX

CLI que prepara Claude Code e Codex com o Core GEX e com escolhas opcionais, sempre com prévia, confirmação e histórico reversível.

## Propósito

- Instalar e manter o Core GEX nos dois agentes: instruções globais e skills `gex-*`.
- Deixar a pessoa escolher especializações, skills Matt e hooks. Nada opcional é instalado sem escolha.
- Mostrar addons de terceiros com o comando oficial de instalação. O kit não executa esse comando.

## Quem usa

Pessoas de desenvolvimento da GEX que usam Claude Code ou Codex em repositórios da organização.

## Language

### Conteúdo

**Core**:
Conjunto obrigatório de instruções globais GEX e sete skills `gex-*`, instalado nos dois agentes.
_Avoid_: base, pacote padrão

**Instruções globais**:
Bloco entre marcadores do kit em `~/.claude/CLAUDE.md` e `~/.codex/AGENTS.md`.
_Avoid_: prompt global, system prompt

**Skill**:
Pasta com `SKILL.md` que o agente carrega quando a tarefa pede.
_Avoid_: plugin, comando

**Especialização**:
Skill GEX opcional, escolhida por função: Gateways, Dados e Plataforma, ou Produto e Interface.
_Avoid_: pack extra, skill avançada

**Skill Matt**:
Cópia offline de uma skill de mattpocock/skills, no commit fixado.
_Avoid_: skill externa

**Referência**:
Item do catálogo mostrado com motivo e fonte, que o kit nunca instala.
_Avoid_: candidato, sugestão

**Addon**:
Método ou estilo de terceiros que a pessoa instala por conta própria, com o comando oficial exibido pelo kit.
_Avoid_: plugin do kit, extensão

### Hooks

**Hook de agente**:
Handler em `settings.json` ou `hooks.json` que o agente executa num evento.
_Avoid_: gatilho, script do agente

**Hook de política**:
Hook de agente que aplica regra GEX (formato de commit, preflight de PR, ação sensível) e só age em repositório GEX.
_Avoid_: hook obrigatório

**Git hook**:
Script em `.git/hooks` de um repositório, gravado só com `--repo` e confirmação.
_Avoid_: hook global

**Adaptador**:
Subcomando `gex-dev-kit hook <nome>` que traduz o protocolo do agente ou do Git.
_Avoid_: wrapper

**Executável dos hooks**:
Cópia do kit em `~/.local/share/gex-dev-kit/bin` que os hooks de agente e os Git hooks chamam.
_Avoid_: binário global

**Confiança do Codex**:
Aprovação de um hook em `/hooks`, registrada pelo hash da definição do hook.
_Avoid_: ativação

### Operação

**Repositório GEX**:
Repositório com algum remote numa organização GEX (Gex-Development, Gex-Integrations, gex-data-architecture-org, gex-data-architecture ou outra listada pelo usuário), ou com `git config gex.policy true`.
_Avoid_: projeto da empresa

**Escopo GEX**:
Regra que limita os hooks de política a repositórios GEX. Fora do escopo, o hook não interfere e a ação segue.
_Avoid_: modo GEX, filtro da empresa

**Prévia**:
Lista exata do que será criado, atualizado, restaurado ou removido, com o diff do que muda nos arquivos do usuário. É mostrada antes de qualquer gravação.
_Avoid_: dry-run, simulação

**Confirmação**:
Resposta explícita depois da prévia: tecla `s`, `--yes` ou resposta no modo acessível.
_Avoid_: aceite automático

**Histórico**:
Registro privado das gravações do kit, com hashes. O conteúdo anterior fica só nas 5 gravações mais recentes.
_Avoid_: backup, log

**Raiz**:
Diretório absoluto onde o kit grava e guarda o histórico: o HOME ou o git common dir de um repositório.
_Avoid_: destino

**HOME isolado**:
HOME temporário passado em `--home` para testar sem tocar no HOME real.
_Avoid_: sandbox, HOME fake

**Doctor**:
Diagnóstico do que o kit gerencia, comparado com o histórico.
_Avoid_: health check

**Rollback**:
Restauração do último registro do histórico. Aborta se algum alvo mudou depois.
_Avoid_: desfazer, revert

**Agente**:
Claude Code ou Codex.
_Avoid_: IA, assistente, ferramenta

**Agente gerenciado**:
Agente cuja versão foi instalada pelo kit e está registrada no histórico de agentes.
_Avoid_: agente oficial

**Versão fixada**:
Versão exata de um agente, instalada só pelo instalador oficial.
_Avoid_: última versão

### Interface

**Modo TUI**:
Interface de tela cheia em terminal interativo.
_Avoid_: GUI, app

**Modo CLI**:
Saída linha a linha, scriptável, sem animação.
_Avoid_: modo texto, headless

**Modo acessível**:
Modo CLI interativo com perguntas numeradas em português, para leitor de tela e terminal simples.
_Avoid_: modo simples

**Design System**:
Tokens, tema, movimento, layout e componentes da TUI.
_Avoid_: skin, UI kit

**Vitrine**:
Tela `ds` que mostra todos os componentes e a cobertura Charm.
_Avoid_: galeria, demo

**Comando equivalente**:
Linha de CLI que reproduz uma ação feita na TUI.
_Avoid_: atalho

## Escopo aprovado

- Core GEX nos dois agentes, com prévia, confirmação e histórico.
- Especializações GEX e skills Matt, em cópia offline no commit fixado.
- Hooks de política GEX, que só agem em repositório GEX, e hook de início de sessão opcional.
- Git hooks por projeto, só com `--repo` e confirmação.
- `doctor` e `rollback` para ver e desfazer o que o kit gravou.
- Catálogo de addons com fonte, licença, versão conferida e comando oficial. Nada é executado.
- Instalação e reversão da versão fixada de Claude Code e Codex, só pelo instalador oficial.

## Fora de escopo

- Servidor MCP.
- Instalação de addons. O kit só mostra o comando oficial.
- Alterar arquivos do projeto fora de `.git`.
- Publicação: push, PR, release ou merge feitos pelo kit.
- GSD, que não entra no kit.
