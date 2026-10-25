# Kit de Desenvolvimento GEX

CLI para preparar Claude Code e Codex com instruções globais e sete skills de trabalho GEX obrigatórias. Skills Matt Pocock, especializações e um hook de Conventional Commits entram por escolha. A TUI também instala plugins opcionais; MCPs e arquivos de projetos não são alterados pelo Core.

## Estado

- **v0.2.0:** última release publicada no GitHub. Não tem as sete skills Core, as especializações nem o chat da TUI.
- **v0.3.0:** binários gerados localmente em `dist/` com `scripts/release.sh v0.3.0`, ainda não publicados. Não há tag, release, push nem PR. Os comandos de download abaixo só funcionam depois que a release existir; publicar exige confirmação explícita do responsável. As notas estão em [`docs/release-notes/v0.3.0.md`](docs/release-notes/v0.3.0.md).

## Baixar a v0.3.0

Repositório privado da organização: é preciso acesso à Gex-Development e autenticação no `gh`. Baixe o executável correspondente ao sistema na [release](https://github.com/Gex-Development/gex-dev-kit/releases). Os alvos publicados são `darwin_{arm64,amd64}`, `linux_{arm64,amd64}` (Ubuntu/Fedora) e `windows_{arm64,amd64}` (PowerShell). Confira `SHA256SUMS` antes de executar.

```sh
gh release download v0.3.0 --repo Gex-Development/gex-dev-kit --dir .
sha256sum -c SHA256SUMS
chmod +x gex-dev-kit_v0.3.0_linux_amd64
./gex-dev-kit_v0.3.0_linux_amd64
```

No macOS, confira checksums com `shasum -a 256 -c SHA256SUMS` e troque `linux_amd64` por `darwin_arm64` ou `darwin_amd64`. No PowerShell, baixe o executável Windows correspondente, confira hash com `Get-FileHash` e execute-o.

## Piloto local (v0.3.0)

Use sempre um HOME descartável. Estes comandos só alteram o HOME temporário:

```sh
go build -o ./gex-dev-kit ./cmd/gex-dev-kit
PILOT_HOME="$(mktemp -d)"
HOME="$PILOT_HOME" ./gex-dev-kit
./gex-dev-kit doctor --home "$PILOT_HOME"
./gex-dev-kit rollback --home "$PILOT_HOME"
```

No terminal interativo, `gex-dev-kit` abre boas-vindas e verifica o ambiente; `Enter` ou clique inicia o chat do assistente GEX. Quando está ocioso, digite `/` para abrir a paleta de comandos. `/setup` começa com um roteiro curto: Harness, Agentes, Destino, Extras, Plugins, Plano e Instalação. Após escolher os harnesses, o kit os reconhece numa frase e mostra um banner para cada Claude Code ou Codex escolhido, sem verificação separada por ferramenta. Para testar sem atualizar executáveis, escolha **Agora não** na pergunta sobre agentes.

Na conversa, a barra escura mostra GEX, o primeiro harness detectado e sua versão, o caminho do projeto (diretório atual) e, quando disponível, o branch Git (`*N` arquivos modificados, `?N` não rastreados). O HOME de instalação aparece apenas quando difere do HOME real. A régua fica na borda superior rosa da caixa de entrada: fora do setup mostra o Core instalado (`N/7`) e a versão do kit; durante o setup, porcentagem e `Etapa N/7`. Não há divisória separada acima da transcrição. `GEX_NO_POWERLINE=1` troca os separadores decorativos por `›`; terminais em modo Plain usam ASCII.

Com animação ativa, cada resposta chega por palavras inteiras num cronograma determinístico a 60 fps: base de 55 caracteres por segundo, intervalos irregulares, pausas de pontuação e aceleração proporcional para limitar textos longos a cerca de 9 s. Entre mensagens há uma batida de 450 ms (700 ms após mais de 300 runas). A linha de trabalho mostra spinner, tempo e brilho móvel no rótulo da fase atual. Cada ferramenta tem cartão de cantos arredondados, estado colorido, resultado alinhado e tempo na borda inferior; enquanto executa, seu corpo exibe checklist de fases de 1 s. Em “Verificar ambiente”, HOME, Claude Code, Codex e Core substituem suas fases pelas linhas reais ao fim de cada etapa; nas outras ferramentas, o checklist cede lugar ao resultado final. A retenção visual dura no máximo 6 s; progresso real chega imediatamente. `Enter` ou espaço pula a animação atual; a entrada permanece travada enquanto o assistente escreve ou uma ferramenta executa. Depois que a gravação real termina, a exibição pode continuar sem adiar `Ctrl+C`. Com `--no-anim`, respostas e resultados aparecem imediatamente, sem checklist.

`ACCESSIBLE=1` ou `TERM=dumb` usam menu e confirmação em texto. `NO_COLOR=1` remove apenas as cores; com terminal interativo, o chat continua.

Para escolher opções sem o formulário:

```sh
./gex-dev-kit install --home "$PILOT_HOME" --skip-agent-update --skill tdd --hook gex-conventional-commits
```

A prévia aparece antes de qualquer gravação. Confirme no terminal. Em automação isolada, `--yes` pula a pergunta, mas a prévia continua sendo mostrada.

`doctor` pode sair com código 1 quando encontrar alteração em arquivo do kit. Agente externo que o kit não gerencia aparece como aviso e não muda o código de saída.

## Usar

```sh
./gex-dev-kit_v0.3.0_linux_amd64 skills
./gex-dev-kit_v0.3.0_linux_amd64 install --skill tdd --hook gex-conventional-commits
./gex-dev-kit_v0.3.0_linux_amd64 doctor
./gex-dev-kit_v0.3.0_linux_amd64 rollback
```

Para o código da branch, troque o executável por `./gex-dev-kit`.

### Modos de interface

| Situação | Interface usada |
|---|---|
| Terminal interativo, sem opção extra | Chat (TUI) |
| `--yes`, `--quiet` ou `--json` | CLI |
| `--ui=cli` | CLI |
| `--ui=tui` sem terminal interativo | Erro de uso (código 2) |
| `ACCESSIBLE=1` ou `TERM=dumb` | Menu e confirmação em texto |
| `NO_COLOR=1` em terminal interativo | Chat sem cores |

### Comandos dentro do chat

| Comando | Ação |
|---|---|
| `/setup` | Configurar ou atualizar o kit |
| `/catalogo` | Ver skills e hooks com “O que faz”, “Quando usar” e “Analogia” |
| `/doctor` | Verificar a integridade |
| `/rollback` | Desfazer a última instalação |
| `/agentes` | Restaurar versões anteriores dos agentes |
| `/githooks` | Configurar git hooks |
| `/addons` | Ver e instalar plugins de terceiros após revisão e confirmação |
| `/menu` | Mostrar as opções como lista, quando necessário |
| `/voltar` | Voltar uma etapa |
| `/ajuda` | Listar comandos e atalhos |
| `/limpar` | Limpar a conversa atual |
| `/sair` | Encerrar |

Na paleta, `↑`/`↓` escolhem e `Tab` completa. Também é possível digitar palavras como “instalar”, “skill”, “verificar”, “desfazer”, “agente”, “git hook” e “addon”. Nas perguntas com várias opções, espaço e números `1`–`9` marcam itens; `Enter` confirma a seleção. Se uma escolha obrigatória ainda não estiver marcada, `Enter` escolhe o item sob o cursor. O botão **Confirmar (N)** ou **Pular (nenhuma marcada)** também confirma com clique; sem Extras selecionados, não há skills ou hooks opcionais.

Listas mostram só os itens que cabem, preservando ao menos quatro linhas da conversa. Escolhas longas exibem `‹ página i/n ›`: `←`/`→`, `PageUp`/`PageDown` ou clique em `‹`/`›` trocam a página; números `1`–`9` valem na página atual e marcações persistem. Com texto digitado, `←`/`→` movem o cursor. Sem escolha paginada, `PageUp`/`PageDown` e a roda continuam rolando a conversa. A paleta e o navegador de pastas mostram `… mais N` quando não cabem; digite para filtrar. Em `/catalogo`, Skills são divididas em Core, Engenharia, Produtividade e Especializações (Dados e Plataforma / Produto e Interface); Referências, em Skills de terceiros e Hooks candidatos. `Esc` volta um nível.

Em **Extras**, a pergunta `Além do Core (obrigatório), o que mais você quer instalar?` lista grupos opcionais; apenas os grupos marcados abrem suas opções. Depois, **Plugins** permite selecionar instaladores para Claude Code/Codex. O resumo antes da prévia mostra harness, destino, Core, extras, agentes e plugins escolhidos. Confira comandos, conflitos, prévia e diff; só **Sim, aplicar** executa. `Esc` limpa a entrada ou volta à etapa anterior; `/voltar` também volta. Perguntas de caminho oferecem navegador de pastas (`Tab` completa, `Enter` abre a pasta destacada); é possível colar com `Cmd+V` ou `Ctrl+V`, conforme o terminal. Clique no título de uma ferramenta ou use `Ctrl+O` para expandir o bloco; a roda navega escolhas ou conversa. A partir de 120 colunas e 30 linhas, a barra lateral mostra ambiente e sete etapas. Enquanto GEX escreve ou executa ferramenta, o campo fica travado: entradas digitadas não são guardadas. `Enter` ou espaço apenas termina de revelar texto em curso.

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
- O schema de `skills --json` é `gex-dev-kit.skills/v1`. Ele estreia na v0.3.0 e as especializações não têm o campo `credentials`: as skills de API de gateway ficaram fora do catálogo. A v0.2.0 não tinha `skills --json`.

### Códigos de saída

| Código | Quando acontece |
|---|---|
| `0` | Concluído sem falha |
| `1` | Falha de execução ou de verificação (por exemplo, `doctor` com arquivo alterado) |
| `2` | Uso inválido: opção ou comando desconhecido, valor de flag ausente, caminho relativo em `--home`, skill ou hook desconhecido, `--repo` que não é repositório Git, `--run` ou `--pre-commit-run` sem o hook correspondente |
| `129` | SIGHUP (terminal fechado) |
| `130` | Ctrl+C (SIGINT) |
| `143` | SIGTERM |

Com sinal durante uma gravação, o kit termina a gravação em curso antes de sair.

## O que mudou desde a v0.2.0

- A TUI virou um chat com o assistente GEX; setup, catálogo, doctor, rollback, git hooks e addons são fluxos dele, abertos por comandos com barra.
- O kit instala os plugins opcionais `caveman`, `ponytail`, `superpowers` e `spec-kit` depois da sua confirmação, no setup e em `/addons`.
- Sete skills Core obrigatórias e sete especializações opcionais, com explicação de cada skill e hook.
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

São 7 especializações opcionais, nos grupos Dados e Plataforma e Produto e Interface. Instalar uma skill apenas copia seus arquivos; atividades externas exigem permissões próprias.

O hook `gex-conventional-commits` usa `PreToolUse` nos dois agentes. Ele bloqueia commit com mensagem fora de Conventional Commits, quando a mensagem é literal. É orientação local: não substitui validação Git por projeto nem proteção remota.

No Codex, abra `/hooks`, revise e confie na definição instalada. O kit não faz essa confirmação por você.

Configurações existentes são preservadas. Login e modelo não são alterados.

O catálogo tem 4 hooks de agente, 3 git hooks e 25 referências de hooks de terceiros. As referências não são instaladas.

As 18 skills Matt Pocock são cópias offline do commit [`5beb35b0`](https://github.com/mattpocock/skills/tree/5beb35b040e3de97e54fa9e386d7e42d15c295c2), sob [licença MIT](https://github.com/mattpocock/skills/blob/5beb35b040e3de97e54fa9e386d7e42d15c295c2/LICENSE). Instalar uma skill só copia os arquivos; não a executa. Há 11 referências de skill que ficam só no catálogo, como `setup-matt-pocock-skills`, `implement-spec` e `ask-matt`. Elas alteram projeto, labels ou dependem de outras partes do catálogo. Skills que criam issues, labels, branches ou alteram projetos exigem autorização separada quando forem usadas.

Os quatro plugins opcionais são `caveman` (respostas curtas), `ponytail` (menor mudança completa), `superpowers` (métodos de brainstorming, testes e revisão) e `spec-kit` (fluxo de especificação). Na etapa **Plugins** do `/setup`, entre Extras e Plano, ou em `/addons`, selecione destinos Claude Code/Codex, confira os comandos exatos e os conflitos com o Core e confirme antes de executar; fonte e licença ficam no card de `/addons`. A CLI `addons` só consulta o catálogo; `install` não instala plugins. A instalação roda cada comando sem shell, em grupo próprio e sem terminal de controle, com timeout, ambiente reduzido, HOME escolhido e versões fixadas; Claude Code usa tag porque seu marketplace não aceita commit, e superpowers segue o pin do marketplace oficial da Anthropic, sujeito a atualizações. Se faltar um executável, o passo é pulado e mostrado como instrução manual. `superpowers` no Codex exige `/plugins` interativo; `spec-kit` exige `specify init` manual por projeto. Depois de instalar hooks no Codex, abra `/hooks` e aprove-os. O rollback do kit não desfaz plugins: remova-os no próprio Claude Code ou Codex.

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
- `scripts/release.sh vX.Y.Z` compila seis alvos (limite de 16 MiB por binário) e gera `dist/SHA256SUMS` e `dist/THIRD_PARTY_LICENSES.txt`. Exige `docs/release-notes/vX.Y.Z.md`. Não publica. Publicar exige a opção `--publish`, versão aprovada e confirmação explícita.

Mudanças seguem Conventional Commits. Não há CI remoto nesta fase: testes locais não substituem a proteção contínua das branches.

Fontes de compatibilidade: [hooks Claude Code](https://code.claude.com/docs/en/hooks), [hooks Codex](https://developers.openai.com/codex/hooks), [instalação Claude Code](https://code.claude.com/docs/en/setup), [instalação Codex](https://developers.openai.com/codex/cli). Estado e decisões do kit: `kit-desenvolvimento-gex-acompanhamento.html` na documentação visual da equipe.
