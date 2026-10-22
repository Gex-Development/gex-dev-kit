# Registro de decisões (ADR)

Este arquivo guarda todas as decisões em um só lugar. A GEX não usa um arquivo por decisão (`docs/adr/NNNN-*.md`), como sugere o modelo da skill Matt. Cada decisão segue o modelo de `ADR-FORMAT.md`: um título e de uma a três frases com o contexto, a decisão e o motivo.

## ADR-0001: Formato de commit

Commits usam Conventional Commits em inglês, sem emoji e sem assinatura de IA. Esta regra vence o hook do kit e o formato com emoji do CLAUDE.md pessoal.

## ADR-0002: Trabalho só local

Branches `feature/...` ficam no repositório local. Push, PR, release e publicação de binário só acontecem com confirmação explícita da pessoa.

## ADR-0003: Forma do Design System

O tema é um valor imutável, com `tokens` dentro de `theme`. Os pacotes são `terminal`, `theme`, `motion`, `layout`, `components` e `screens`, mais `internal/app`. Tema global mutável e telas que chamam `kit` direto foram descartados.

## ADR-0004: Movimento e redução de movimento

As molas e partículas usam harmonica. A animação é desligada por `--no-anim`, `GEX_NO_ANIM`, `ACCESSIBLE`, `CI`, saída sem terminal e `TERM=dumb`. Os quadros só são desenhados enquanto houver animação.

## ADR-0005: gum fora do runtime

O gum só é usado em scripts (release e comparação visual). O binário não depende dele.

## ADR-0006: Sentido de NO_COLOR

`NO_COLOR` só remove a cor. A TUI continua. `ACCESSIBLE` e `TERM=dumb` levam ao modo textual.

## ADR-0007: Rollback com prévia e confirmação

O rollback segue o mesmo contrato do install: mostra a prévia com diff e pede confirmação. `--yes` serve para automação. Scripts antigos que dependiam do comportamento anterior precisam ser ajustados.

## ADR-0008: Hooks implementados e referências

A decisão de cada hook depende do agente e da documentação oficial (tabela 5.2 do plano). No Codex, sem a opção `ask`, `sensitive-action` e `pre-pr` não são instalados, e `sensitive-action` cobre `Bash`, `PowerShell` e `Read`. Em `claude -p`, a pergunta vira negação.

## ADR-0009: gex-session-start opcional

O hook de início de sessão é opt-in. Ele mostra só fatos verificáveis, com até 1.200 caracteres, e nunca bloqueia.

## ADR-0010: Git hooks por --repo

Sem `--repo`, o comando só mostra a prévia, e o histórico fica no git common dir. Os scripts chamam o executável do HOME, o kit recusa `core.hooksPath` e hook alheio, e pre-commit e pre-push só rodam comandos que o projeto documenta. Exceções do pre-commit ficam em `<git common dir>/gex-dev-kit/pre-commit-allow`, e o kit nunca usa `--no-verify`.

## ADR-0011: Merge de hooks aditivo

Reinstalar substitui só o mesmo hook, e outros hooks do kit ou alheios ficam. A remoção acontece pelo rollback, e um arquivo de agente só recebe hook que o agente suporta.

## ADR-0012: Pack do Design System GEX

O `tokens.css` não tem comentários e os 2 SVG não têm `<metadata>`. Os PNGs ficam fora por causa da proveniência C2PA e do tamanho. A licença da marca ainda está pendente.

## ADR-0013: Packs de gateway mantidos (desvio)

As prévias dos packs de gateway podem divergir, aceito. Desvio do escopo aprovado: o cofre do sistema não foi implementado. Só se verifica se a variável de ambiente está presente.

## ADR-0014: Skills Matt fixadas

As skills Matt ficam no commit `5beb35b0` (licença MIT), e `retro` entra no kit. `wizard`, `to-spec`, `to-tickets`, `triage`, `wayfinder` e `implement` viram referência, e três skills ficam fora.

## ADR-0015: Addons só com comando exibido

O kit nunca executa a instalação de um addon. Nenhum addon fica ativo por padrão. O GSD fica fora do kit.

## ADR-0016: Versões fixas dos agentes

Claude Code 2.1.306 e Codex 0.163.0 são instalados só pelos instaladores oficiais, com versão exata. O processo roda em grupo de processos próprio.

## ADR-0017: Vídeo do guia

O arquivo `guia.mp4` só entra no repositório se tiver até 20 MB. Caso contrário, o repositório traz a instrução para gerar o vídeo.

## ADR-0018: Códigos de saída

Os códigos são 0, 1, 2, 130, 143 e 129, e caminho relativo sai com 2. `doctor` sai com 1 só por item do kit ou agente gerenciado inconsistente. Agente não gerenciado vira aviso.

## ADR-0019: Casos de uso compartilhados

`internal/app` serve a CLI e a TUI. A igualdade entre elas vem da estrutura do código e de um teste com HOME normalizado.

## ADR-0020: Texto do Core alinhado ao brief

O texto do Core usa "este texto" no lugar de "esta prévia". A expressão "esta prévia" só fazia sentido na página HTML.

## ADR-0021: Lacunas aceitas na Fase B

Ficam aceitas as lacunas da lista 11.1 do plano e as que o crítico apontar e o orquestrador aceitar nas rodadas seguintes.

## ADR-0022: Escopo GEX dos hooks de política

Os hooks `gex-conventional-commits`, `gex-pre-pr` e `gex-sensitive-action-warning` só agem se algum remote for de organização GEX ou se `git config gex.policy true` estiver ligado, e permitem a ação fora desse escopo. A lista de organizações só amplia pelo arquivo `<os.UserConfigDir()>/gex-dev-kit/orgs`, e `gex.policy false` desliga o escopo.

## ADR-0023: Histórico com blobs e retenção

Os hashes ficam no JSON do histórico, e o conteúdo anterior fica em `history/blobs/<sha256>` só para as 5 gravações mais recentes. Registros antigos continuam visíveis no doctor. Há migração do formato v0.2.0.

## ADR-0024: Sinais e processos

O kit usa `tea.WithoutSignalHandler()` com tratador próprio, e adia SIGINT, SIGTERM e SIGHUP durante a gravação (códigos 130, 143 e 129). O lock guarda o PID. Comandos externos são encerrados com o grupo de processos, com `WaitDelay` de 5 s.

## ADR-0025: Prévia com diff mascarado

A prévia mostra diff unificado por escopo, com limite de linhas. Valores de segredo aparecem mascarados. O diff usa `go-udiff`, que já está no grafo de módulos.

## ADR-0026: Prompts acessíveis próprios

Menu, escolha múltipla e confirmação no modo acessível são feitos pelo próprio kit, em português. O huh acessível só é usado para `Note`, `Input` e `Text`.

## ADR-0027: Exceção de cobertura da vitrine

`x/exp/teatest` e `x/exp/golden` não têm tela. Eles aparecem na página "Testes" da vitrine.

## ADR-0028: Versão no fonte

A versão fica no fonte como `var version = "v0.3.0"`, como pede o briefing. O release troca o valor por `-ldflags`. As capturas usam o binário de release.

## ADR-0029: Executável dos hooks sempre atualizado

Toda instalação atualiza o executável estável quando ele é diferente do kit em uso ou quando há handler do kit. O doctor avisa se houver diferença.

## ADR-0030: Esqueletos compiláveis e contratos por onda

Antes de cada onda, o orquestrador registra esqueletos com as assinaturas da seção 1 do plano. Nenhum esqueleto importa pacote de onda posterior. Os contratos entre tarefas paralelas ficam no esqueleto, e cada tarefa os verifica com dados ou funções falsas (tabela 10.7).
