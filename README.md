# Kit de Desenvolvimento GEX

CLI para preparar Claude Code e Codex com instruções globais e sete skills de trabalho GEX obrigatórias. Skills Matt Pocock, especializações e um hook de Conventional Commits entram por escolha. Addons, MCPs e arquivos de projetos não são alterados.

## Estado

- **v0.2.0:** publicada na release do GitHub.
- **v0.3.0:** preparada localmente e não publicada. Não há release, tag, push nem PR. Publicar exige confirmação explícita do responsável.

## Baixar a v0.2.0

Repositório privado da organização: é preciso acesso à Gex-Development e autenticação no `gh`. Baixe o executável correspondente ao sistema na [release](https://github.com/Gex-Development/gex-dev-kit/releases). Os alvos publicados são `darwin_{arm64,amd64}`, `linux_{arm64,amd64}` (Ubuntu/Fedora) e `windows_{arm64,amd64}` (PowerShell). Confira `SHA256SUMS` antes de executar.

```sh
gh release download v0.2.0 --repo Gex-Development/gex-dev-kit --dir .
sha256sum -c SHA256SUMS
chmod +x gex-dev-kit_v0.2.0_linux_amd64
./gex-dev-kit_v0.2.0_linux_amd64
```

No macOS, confira checksums com `shasum -a 256 -c SHA256SUMS` e troque `linux_amd64` por `darwin_arm64` ou `darwin_amd64`. No PowerShell, baixe o executável Windows correspondente, confira hash com `Get-FileHash` e execute-o.

A release v0.2.0 não contém as sete skills Core, as especializações nem as telas novas. Para avaliar essas mudanças, use o código da branch, não o binário publicado.

## Piloto local (v0.3.0)

Use sempre um HOME descartável. Estes comandos só alteram o HOME temporário:

```sh
go build -o ./gex-dev-kit ./cmd/gex-dev-kit
PILOT_HOME="$(mktemp -d)"
HOME="$PILOT_HOME" ./gex-dev-kit
./gex-dev-kit doctor --home "$PILOT_HOME"
./gex-dev-kit rollback --home "$PILOT_HOME"
```

No terminal interativo, o painel mostra o estado do Core com spinner, menu navegável e atalhos `?`, `↑`/`↓`, `Enter` e `q`. Na instalação guiada, escolha skills, especializações e hook. No último passo, responda **Agora não** para não atualizar executáveis durante o piloto. A prévia rolável mostra os arquivos afetados: `Enter` abre a confirmação, `s` aplica e `Esc` cancela.

`ACCESSIBLE=1`, `NO_COLOR=1` ou `TERM=dumb` trocam o painel por menu e confirmação em texto. `NO_COLOR=1` tira só as cores.

Para escolher opções sem o formulário:

```sh
./gex-dev-kit install --home "$PILOT_HOME" --skip-agent-update --skill tdd --hook gex-conventional-commits
```

A prévia aparece antes de qualquer gravação. Confirme no terminal. Em automação isolada, `--yes` pula a pergunta, mas a prévia continua sendo mostrada.

`doctor` pode sair com código 1 quando encontrar alteração em arquivo do kit. Agente externo que o kit não gerencia aparece como aviso e não muda o código de saída.

## Usar

```sh
./gex-dev-kit_v0.2.0_linux_amd64 skills
./gex-dev-kit_v0.2.0_linux_amd64 install --skill tdd --hook gex-conventional-commits
./gex-dev-kit_v0.2.0_linux_amd64 doctor
./gex-dev-kit_v0.2.0_linux_amd64 rollback
```

Os exemplos acima são da v0.2.0. Para a branch local, troque o executável por `./gex-dev-kit`.

### Modos de interface

| Situação | Interface usada |
|---|---|
| Terminal interativo, sem opção extra | Painel (TUI) |
| `--yes`, `--quiet` ou `--json` | CLI |
| `--ui=cli` | CLI |
| `--ui=tui` sem terminal interativo | Erro de uso (código 2) |
| `ACCESSIBLE=1`, `NO_COLOR=1` ou `TERM=dumb` | Menu e confirmação em texto |

### Comandos

| Comando | Para que serve |
|---|---|
| `install` | Mostra a prévia, pede confirmação e grava o Core e as escolhas |
| `doctor` | Confere o HOME (`--home`) ou os git hooks de um repositório (`--repo`) |
| `rollback` | Desfaz a última gravação do HOME ou do repositório |
| `rollback-agents` | Restaura a versão anterior de Claude Code e Codex, com confirmação |
| `skills` | Lista as skills do Core, as especializações e as referências |
| `hooks` | Lista os hooks de agente, os git hooks e as referências |
| `addons [NOME]` | Lista os Addons ou mostra o comando oficial de um deles (não executa) |
| `githooks --repo CAMINHO` | Mostra a prévia dos git hooks; grava só com `--repo` e confirmação |
| `ds` | Abre a vitrine do Design System (`--ui=cli` mostra em texto) |
| `version` | Mostra a versão |
| `help [COMANDO]` | Mostra a ajuda geral ou a de um comando |

Opções globais: `--ui=auto|tui|cli`, `--no-anim`, `--quiet`, `--json` (em `skills`, `hooks`, `addons` e `doctor`), `-h` e `-v`.

Regras de uso:

- `--home` e `--repo` aceitam só caminho absoluto. Caminho relativo sai com código 2.
- `--home` vale para `install`, `doctor`, `rollback` e `rollback-agents`.
- `install` mostra arquivos e versões afetadas e depois pede confirmação. `--yes` confirma em automação, depois da prévia.
- `--skill` e `--hook` podem ser repetidos.
- Sem seleção extra, `install` grava o Core e tenta atualizar os agentes elegíveis. `--skip-agent-update` não toca em rede nem nos binários dos agentes.
- `--json` em `skills`, `hooks`, `addons` e `doctor` imprime só JSON na saída padrão.

### Códigos de saída

| Código | Quando acontece |
|---|---|
| `0` | Concluído sem falha |
| `1` | Falha de execução ou de verificação (por exemplo, `doctor` com arquivo alterado) |
| `2` | Uso inválido: opção desconhecida, caminho relativo em `--home`, comando desconhecido |
| `129` | SIGHUP (terminal fechado) |
| `130` | Ctrl+C (SIGINT) |
| `143` | SIGTERM |

Com sinal durante uma gravação, o kit termina a gravação em curso antes de sair.

## O que mudou desde a v0.2.0

- `rollback` mostra prévia com diff e pede confirmação. Em automação, use `--yes`.
- A prévia de `install` mostra o diff com dados sensíveis mascarados.
- `doctor` não sai mais com código 1 por agente que o kit não gerencia. Esse caso vira aviso.
- Reinstalar sempre atualiza o executável dos hooks.
- O histórico guarda o conteúdo fora do JSON e alcança as 5 gravações mais recentes.
- Merge de hooks é aditivo: o kit não remove hooks que já existiam.
- Hooks de política (`gex-conventional-commits`, `gex-pre-pr`, `gex-sensitive-action-warning`) só agem em repositório com remote de uma destas organizações: `Gex-Development`, `Gex-Integrations`, `gex-data-architecture-org` ou `gex-data-architecture` (sem diferenciar maiúsculas). Também agem com `git config gex.policy true`; `false` desliga. Outras organizações podem ser incluídas no arquivo `orgs` da pasta de configuração do usuário. `doctor` mostra a lista efetiva.
- Corrigido o falso positivo do hook de commit.
- Em `claude -p`, as perguntas de `gex-sensitive-action-warning` e `gex-pre-pr` viram negação (comportamento do Claude Code). Não instale esses hooks em máquinas de automação.

## Core, especializações e skills Matt

O Core escreve instruções em `~/.claude/CLAUDE.md` e `~/.codex/AGENTS.md`. Instala 7 skills `gex-*` em `~/.claude/skills/` e `~/.agents/skills/`.

São 10 especializações opcionais, por grupo. Quatro delas são skills de gateway (`gex-pagamerican-api`, `gex-buygoods-api`, `gex-clickbank-api` e `gex-local-spark-tests`). Os clientes de gateway são só leitura e exigem credenciais no ambiente na hora da seleção:

- `PAGAMERICAN_API_KEY`
- `BUYGOODS_ACCOUNT_ID` e `BUYGOODS_API_TOKEN`
- `CLICKBANK_API_KEY` e `CLICKBANK_ACCOUNT`

Instalar uma skill não executa consulta. Acesso real depende de credenciais e permissões aprovadas.

O hook `gex-conventional-commits` usa `PreToolUse` nos dois agentes. Ele bloqueia commit com mensagem fora de Conventional Commits, quando a mensagem é literal. É orientação local: não substitui validação Git por projeto nem proteção remota.

No Codex, abra `/hooks`, revise e confie na definição instalada. O kit não faz essa confirmação por você.

Configurações existentes são preservadas. Login, modelo e credenciais não são alterados.

O catálogo tem 4 hooks de agente, 3 git hooks e 25 referências de hooks de terceiros. As referências não são instaladas.

As 18 skills Matt Pocock são cópias offline do commit [`5beb35b0`](https://github.com/mattpocock/skills/tree/5beb35b040e3de97e54fa9e386d7e42d15c295c2), sob [licença MIT](https://github.com/mattpocock/skills/blob/5beb35b040e3de97e54fa9e386d7e42d15c295c2/LICENSE). Instalar uma skill só copia os arquivos; não a executa. Há 11 referências de skill que ficam só no catálogo, como `setup-matt-pocock-skills`, `implement-spec` e `ask-matt`. Elas alteram projeto, labels ou dependem de outras partes do catálogo. Skills que criam issues, labels, branches ou alteram projetos exigem autorização separada quando forem usadas.

Os 4 Addons (`caveman`, `ponytail`, `superpowers` e `spec-kit`) aparecem só com o comando oficial. O kit não os instala nem os executa.

## Agentes

Claude Code `2.1.306` e Codex `0.163.0` são as versões fixadas nesta release. A CLI só instala ou atualiza executáveis standalone próprios, no perfil do usuário, pelos instaladores oficiais com versão exata.

Brew, npm, winget, apt, dnf, binários não gerenciados, PATH ambíguo e atualização automática nativa que saia do pin ficam bloqueados. Nesses casos, o Core ainda pode ser instalado, mas o comando retorna erro e não afirma atualização.

`rollback-agents` pede confirmação e reverte só um upgrade já gerenciado, com versão anterior registrada. A primeira instalação não tem versão anterior e não é desinstalada automaticamente.

## Git hooks por projeto

`githooks` mostra a prévia dos git hooks de um repositório. Só grava com `--repo` e confirmação. `rollback --repo` desfaz a gravação.

## Desenvolvimento

```sh
go test ./...
go vet ./...
go build -o /tmp/gex-dev-kit-piloto ./cmd/gex-dev-kit
./scripts/smoke.sh /tmp/gex-dev-kit-piloto
```

- `scripts/check.sh` roda o conjunto completo: formato, comentários, `go vet`, testes, build e smoke.
- `scripts/smoke.sh CAMINHO_DO_BINARIO` instala, roda `doctor`, testa hook e rollback em HOME descartável.
- `scripts/release.sh vX.Y.Z` compila seis alvos e gera `dist/SHA256SUMS` e `dist/THIRD_PARTY_LICENSES.txt`. Não publica. Publicar exige a opção `--publish`, versão aprovada e confirmação explícita.

Mudanças seguem Conventional Commits. Não há CI remoto nesta fase: testes locais não substituem a proteção contínua das branches.

Fontes de compatibilidade: [hooks Claude Code](https://code.claude.com/docs/en/hooks), [hooks Codex](https://developers.openai.com/codex/hooks), [instalação Claude Code](https://code.claude.com/docs/en/setup), [instalação Codex](https://developers.openai.com/codex/cli). Estado e decisões do kit: `kit-desenvolvimento-gex-acompanhamento.html` na documentação visual da equipe.
