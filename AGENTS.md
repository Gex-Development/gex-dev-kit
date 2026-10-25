# AGENTS.md

## Propósito

CLI em Go que prepara Claude Code e Codex com instruções globais, skills GEX e hooks.
Addons aparecem só com o comando oficial exibido; o kit não os executa.
Escopo, estado e uso para usuários estão em `README.md`.

## Estrutura

- `cmd/gex-dev-kit`: entrada da CLI, comandos, sinais e saída JSON.
- `internal/app`: casos de uso compartilhados por CLI e TUI (instalar, diff, rollback, doctor).
- `internal/catalog`: catálogo embutido (Core, especializações, Matt, Addons, hooks).
- `internal/kit`: gravação segura e histórico no HOME.
- `internal/agents`: controle dos binários do Claude Code e do Codex.
- `internal/hooks`: lógica dos hooks de agente.
- `internal/githooks`: git hooks por repositório.
- `internal/ui/*`: tema, terminal, movimento, layout e componentes do chat; `screens.RunChat` conduz setup, catálogo, doctor, rollback, agentes, git hooks e addons. `ds` permanece separado.

## Comandos

- `scripts/check.sh`: portão completo (formato, comentários, `go vet` com Windows, testes com `-race`, build e smoke). Rode antes de entregar.
- `go test ./internal/ui/screens -run TestGoldenChat -update` e `go test ./internal/ui/theme -run TestGolden -update`: regravam os goldens do chat e da paleta. Só ao mudar o visual de propósito; revise o diff.
- `scripts/smoke.sh CAMINHO_DO_BINARIO`: instalação, doctor, hooks e rollback num HOME descartável.
- `scripts/release.sh vX.Y.Z`: gera `dist/` local. Publicar exige `--publish` e confirmação humana.
- `scripts/ux-audit-env.sh`: sandbox da auditoria de UX; `--run ARQUIVO.sh` roda um script dentro dele.
- `scripts/demo.sh`: compila e abre a TUI num HOME descartável, sem `CI` nem `NO_COLOR`; o HOME é apagado ao sair.

## Convenções

- Código novo sem comentários nem docstrings; preserve `//go:embed` e `//go:build`.
- Commits em Conventional Commits, em inglês, sem emoji e sem assinatura de IA.
- Nomes de código em inglês; mensagens ao usuário em português.
- Estilos só via `internal/ui/theme`; telas chamam `internal/app`, que fala com `kit` e `agents`.
- Função nova leva teste.

## Segurança

- Todo teste manual roda num HOME descartável: `--home "$(mktemp -d)"` ou `scripts/ux-audit-env.sh`. O HOME real fica intocado.
- `git push`, `gh pr create` e `gh release` só com confirmação humana.
- Segredos ficam fora da saída; testes não consultam APIs de gateway.

## Orquestração

- Tarefas em paralelo têm posse de arquivos sem sobreposição.
- Mesmo pacote em paralelo usa worktree separado.
- Quem faz o commit é o orquestrador.

## Pegadinhas

- Um hook antigo no HOME real bloqueia comandos Bash com a palavra-chave de commit, heredoc, comentário com apóstrofo, `$'…'`, barra no fim ou mais de 64 KiB. Grave esses arquivos com a ferramenta de edição.
- `huh` escreve em stderr por padrão.
- `tea.Println` não rebaixa cor.
- Em teatest, `WithProgramOptions` substitui as opções padrão; passe sempre `WithFinalTimeout`.
- Telas encerram com `tea.Quit` e marcam a interrupção no modelo; o `Run*` da tela converte em `SignalError`.

## Documentação

- `docs/ARCHITECTURE.md`: ao mudar pacote, fluxo de dados, Design System, modo CLI/TUI ou segurança.
- `docs/ADR.md`: antes de reverter uma decisão ou quando o código parece "estranho de propósito".
- `docs/CONTEXT.md`: ao nomear algo novo (glossário) ou decidir se algo está no escopo.
- `docs/AUDITORIA-UX.md`: ao mudar mensagens, ajuda ou códigos de saída.
