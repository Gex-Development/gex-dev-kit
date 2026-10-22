# AGENTS.md

## Propósito

CLI em Go que prepara Claude Code e Codex com instruções globais, skills GEX e hooks.
Addons aparecem só com o comando oficial exibido; o kit não os executa.
Escopo, estado e uso para usuários estão em `README.md`.

## Estrutura

- `cmd/gex-dev-kit`: entrada da CLI, comandos, sinais e saída JSON.
- `internal/app`: casos de uso (instalar, diff, rollback, doctor).
- `internal/catalog`: catálogo embutido (Core, especializações, Matt, Addons, hooks).
- `internal/kit`: gravação segura e histórico no HOME.
- `internal/agents`: controle dos binários do Claude Code e do Codex.
- `internal/hooks`: lógica dos hooks de agente.
- `internal/githooks`: git hooks por repositório.
- `internal/ui/*`: tema, terminal, movimento, layout, componentes e telas.
- `scripts/`: `check.sh`, `smoke.sh`, `release.sh`, `ux-audit-env.sh`.
- `docs/`: documentação do kit.

## Comandos

- `scripts/check.sh`: roda formato, comentários, `go vet`, testes, build e smoke.
- `go test ./...`: testes de todos os pacotes.
- `go test ./internal/ui/screens -run TestGolden -update`: atualiza os golden das telas. Use só ao mudar a visual e revise o diff.
- `go test ./internal/ui/theme -run TestGolden -update`: atualiza os golden da paleta. Use só ao mudar a visual e revise o diff.
- `scripts/smoke.sh CAMINHO_DO_BINARIO`: instala, roda `doctor`, testa hook e rollback em HOME descartável.
- `scripts/release.sh vX.Y.Z`: gera `dist/`. Sem `--publish`.
- `scripts/ux-audit-env.sh`: sem argumentos, imprime as variáveis de um sandbox; `--run ARQUIVO.sh` roda um script dentro dele.

## Convenções

- Código novo sem comentários nem docstrings. Preserve `//go:embed` e `//go:build`.
- Commits em Conventional Commits, em inglês, sem emoji e sem assinatura de IA.
- Nomes de código em inglês; mensagens ao usuário em português.
- Estilos só via `internal/ui/theme`.
- Telas não chamam `kit` nem `agents`.
- Função nova leva teste.

## Segurança

- Teste sempre com `--home "$(mktemp -d)"`. Nunca use o HOME real.
- Só `git push`, `gh pr create` e `gh release` com confirmação humana.
- Não imprima segredo. Testes não consultam APIs de gateway.
- `--home` e `--repo` aceitam só caminho absoluto.

## Orquestração

- Tarefas em paralelo têm posse de arquivos sem sobreposição.
- Mesmo pacote em paralelo usa worktree separado.
- Quem faz o commit é o orquestrador.

## Pegadinhas

- Um hook antigo no HOME real bloqueia comandos Bash com a palavra-chave de commit, heredoc, comentário com apóstrofo, `$'…'`, barra no fim ou mais de 64 KiB. Grave esses arquivos com a ferramenta de edição.
- `huh` escreve em stderr por padrão.
- `tea.Println` não rebaixa cor.
- Em teatest, `WithProgramOptions` substitui as opções padrão. Passe sempre `WithFinalTimeout`.
- Telas nunca devolvem `tea.Interrupt`.
- Auditoria e testes manuais rodam só por `scripts/ux-audit-env.sh` ou com `--home "$(mktemp -d)"`.

## Documentação

- `docs/CONTEXT.md`: propósito, glossário e escopo do kit.
