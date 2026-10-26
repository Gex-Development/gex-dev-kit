# Kit de Desenvolvimento GEX

CLI que prepara Claude Code e Codex com o Core GEX e com escolhas opcionais, sempre com prévia, confirmação e histórico reversível.

## Propósito

- Instalar e manter o Core GEX nos dois agentes: instruções globais e skills `gex-*`.
- Deixar a pessoa escolher especializações, skills Matt e hooks. Nada opcional é instalado sem escolha.
- Mostrar e instalar plugins e MCPs opcionais de terceiros na TUI, após revisão dos comandos e confirmação.

## Quem usa

Pessoas de desenvolvimento da GEX que usam Claude Code ou Codex em repositórios da organização.

## Language

**Setup**:
Fluxo do chat da TUI com sete etapas fixas (Harness, Agentes, Destino, Extras, Plugins, Plano e Instalação): escolhe extras e plugins, mostra prévia e diff, pede confirmação e executa a instalação na conversa.
_Avoid_: segundo instalador, modelo de IA

**Extras**:
Grupos opcionais oferecidos além do Core obrigatório; só os grupos marcados abrem seleção de itens. Sem grupo marcado, não há skills nem hooks opcionais.
_Avoid_: parte obrigatória do Core

**Harness**:
Ferramenta de desenvolvimento que lê instruções e skills: Claude Code, Codex ou OMP. OMP usa `~/.agents/skills`; escolher apenas OMP não atualiza binários de Claude Code nem Codex.
_Avoid_: agente instalado pelo kit

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
Skill GEX opcional, escolhida por função: Dados e Plataforma ou Produto e Interface.
_Avoid_: pack extra, skill avançada

**Skill Matt**:
Cópia offline de uma skill de mattpocock/skills, no commit fixado.
_Avoid_: skill externa

**Referência**:
Item do catálogo mostrado com motivo e fonte, que o kit nunca instala.
_Avoid_: candidato, sugestão

**Addon**:
Método, estilo ou MCP de terceiros instalado opcionalmente pelo chat, com comandos oficiais exibidos antes da confirmação; passos interativos ficam manuais.
_Avoid_: plugin do kit, extensão

**MCP**:
Servidor ou cliente de ferramentas que o agente chama; o kit apenas o registra nos clientes.
_Avoid_: plugin

**Chave de API de MCP**:
Credencial que o usuário cola no chat para um MCP; fica em texto puro na configuração do cliente.
_Avoid_: segredo do kit


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
Resposta explícita depois da prévia: opção **Sim, aplicar** no chat, `--yes` na CLI ou resposta no modo acessível.
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

**Versão mais recente**:
Versão publicada mais nova de um agente, consultada antes de instalar e instalada só pelo instalador oficial.
_Avoid_: versão fixada, versão pinada

### Interface

**Modo TUI**:
Chat de tela cheia em terminal interativo; boas-vindas ao entrar sem subcomando.
_Avoid_: GUI, app

**Chat**:
Conversa local do assistente GEX, com respostas e comandos fixos, sem modelo de IA; a entrada fica travada enquanto ele escreve ou executa ferramentas.
_Avoid_: chat com IA, fila de mensagens do usuário

**Fluxo**:
Sequência de perguntas e ferramentas dentro do chat para setup, catálogo, doctor, rollback, agentes, git hooks ou addons.
_Avoid_: tela independente

**Ferramenta (ToolBlock)**:
Cartão emoldurado no histórico que mostra título, checklist de fases durante execução, estado e resultado de uma operação do kit; clique na borda superior alterna o corpo expandido e recolhido.
_Avoid_: comando de shell executado pelo usuário

**Linha de trabalho**:
Linha entre histórico e entrada que exibe operação em curso, spinner, rótulo com brilho móvel e tempo decorrido.
_Avoid_: mensagem enviada pelo usuário

**Linha de status**:
Barra escura do chat com GEX, harness, caminho do projeto e Git; o HOME de instalação aparece somente quando difere do HOME real. Core/versão ou porcentagem/etapa ficam na borda superior rosa da caixa de entrada (`PromptBox`).
_Avoid_: menu de navegação, caminho do projeto como HOME de instalação

**Barra lateral**:
Painel do chat em terminais de pelo menos 120 colunas por 30 linhas, com estado do ambiente e etapas do setup.
_Avoid_: menu separado

**Paleta de comandos**:
Lista filtrável que aparece ao digitar `/` no chat ocioso; `/menu` mostra opções como lista sob demanda.
_Avoid_: menu permanente

**Página de escolha**:
Parte de uma pergunta longa que cabe por inteiro na área de entrada; setas horizontais ou `PageUp`/`PageDown` trocam a página sem perder marcações. Não substitui a rolagem da transcrição.
_Avoid_: janela rolável de itens

**Ocioso**:
Estado do chat sem pergunta ou ferramenta ativa, aguardando comando na barra de entrada.
_Avoid_: fim do programa

**Banner**:
Arte do mascote do harness exibida no histórico durante sua verificação, como a do Claude Code no setup.
_Avoid_: tela separada

**Navegador de pastas**:
Sugestões de diretórios para perguntas de caminho: `Tab` completa e `Enter` desce na pasta destacada.
_Avoid_: seletor de arquivos do projeto

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
- Catálogo de addons com fonte, licença, versão conferida e comandos oficiais; instalação opcional na TUI depois de confirmar.
- Instalação da versão mais recente de Claude Code e Codex pelo instalador oficial, com reversão pela versão exata anterior.

## Fora de escopo

- Servidor MCP. O kit não implementa um servidor; somente registra servidores MCP opcionais nos clientes Claude Code e Codex.
- Alterar arquivos do projeto fora de `.git`.
- Publicação: push, PR, release ou merge feitos pelo kit.
- GSD, que não entra no kit.
