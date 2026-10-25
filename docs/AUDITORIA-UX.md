# Auditoria de UX do CLI gex-dev-kit

Este arquivo registra o resultado da auditoria de UX do PLAN.md, seção 8.8 (tarefa W5-T7).

- Alvo da avaliação: v0.3.0, HEAD ddcea84, feita em 21/10/2026.
- Correções: commit 3b68ba7 (branch worktree-w5t7).
- Avaliação anterior: 20/10/2026, sobre a v0.2.0.

Esta auditoria documenta uma execução anterior da CLI; as notas e resultados de testes abaixo não foram refeitos para o chat atual. O estado da TUI nesta branch está descrito na seção 9.

## 1. Resumo

- A avaliação de 21/10 deu nota 3.4 de 5 no núcleo (critérios 1 a 8) e 3.5 de 5 no geral (critérios 1 a 11).
- Foram achados 13 problemas de UX (UX-001 a UX-013) e 1 problema de processo (PROC-001).
- O commit 3b68ba7 corrige 7 dos 13 achados (UX-001, UX-002, UX-003, UX-005, UX-006, UX-008 e UX-013). Ele também corrige dois itens que não estavam na lista: valor de flag que é outra flag e `--ui=` vazio.
- Justificamos 4 achados (UX-007, UX-009, UX-010 e UX-012, este último só em parte).
- UX-004 e UX-011 foram corrigidos depois, junto com as sentinelas de erro.
- Regressão (seção 8.8, item 7) executada em 22/10 com sandbox novo e binário da HEAD ddd926e: 26 de 27 testes do `test.sh` da avaliação passaram. O único que falhou exigia o comportamento antigo (git hook desconhecido com exit 1); agora sai com exit 2, que é a correção de UX-002.
- PROC-001: na regressão, hash e data de modificação dos arquivos do HOME real ficaram iguais antes e depois. O incidente de 21/10 continua sem origem confirmada e não se repetiu.

## 2. Verificação desta rodada

Comandos rodados na worktree, no HEAD 3b68ba7:

- `go build ./...`: passou.
- Os 13 testes novos ligados às correções (listados na seção 3): todos passaram.
- `go test ./...`: 927 testes passaram em 14 pacotes.

Não rodado nesta rodada:

- A regressão `scripts/ux-audit-env.sh --run <test.sh>` (seção 8.8, item 7). Ela deve rodar depois desta entrega.
- A TUI com terminal real, `NO_COLOR` e `rollback-agents`. Eles não estão cobertos pela avaliação de 21/10 (seção 5).

## 3. Notas por critério

A nota é a da avaliação de 21/10. Ela não foi refeita. A coluna "Achados ligados" mostra o status atual de cada item.

| # | Critério | Nota 21/10 | Achados ligados | Situação hoje |
|---|---|---|---|---|
| 1 | Descoberta | 4.0 | falta de sugestão para comando errado (UX-004) | UX-004 corrigido |
| 2 | Nomes e flags | 3.5 | UX-009 | justificado |
| 3 | Erros e mensagens | 3.0 | UX-002, UX-004, UX-008 | corrigido (UX-002, UX-004, UX-008) |
| 4 | Ajuda e documentação | 3.5 | UX-005 | corrigido |
| 5 | Consistência | 3.0 | UX-006, UX-007, UX-010 | corrigido (UX-006); UX-007 e UX-010 justificados |
| 6 | Visual e saída | 3.0 | UX-007, UX-012 | UX-007 justificado; UX-012 justificado em parte |
| 7 | Desempenho | 4.0 | sem achado | sem mudança |
| 8 | Acessibilidade | 3.5 | TUI com TTY e leitor de tela não testados | lacuna aberta (seção 5) |
| 9 | Integração | 3.0 | UX-001 | corrigido |
| 10 | Segurança | 4.0 | sem achado de falha | sem mudança |
| 11 | Orientação | 3.5 | UX-004, UX-011 | UX-004 e UX-011 corrigidos |

## 4. Achados e status

Status possíveis: "corrigido com evidência", "justificado" e "pendente".

| ID | Prioridade (21/10) | Achado | Status | Evidência ou justificativa |
|---|---|---|---|---|
| UX-001 | Alta | `doctor --json` podia dar `ok: false` com `exit_code: 0` | corrigido com evidência | Commit 3b68ba7. Regra única de nível: `ok` é verdadeiro quando o código de saída é 0. Testes: TestDoctorJSONSemInstalacao, TestDoctorJSONInstalacaoLimpa, TestDoctorJSONArquivoAlterado. Passaram nesta rodada. |
| UX-002 | Alta | Erro de entrada saía como "Falha inesperada" com exit 1 | corrigido com evidência | Commit 3b68ba7. Agora sai com exit 2, com título e ação. Testes: TestUnknownSkillSaiComDoisEExplica, TestUnknownHookInstallSaiComDoisEExplica, TestGitHooksRunSemPrePushSaiComDois. Passaram nesta rodada. |
| UX-003 | Alta | `githooks --repo` em pasta que não é Git escondia a causa | corrigido com evidência | Commit 3b68ba7. A checagem de repositório Git vem antes da seleção de hooks. A mesma mensagem aparece em `doctor`, `rollback` e `githooks`. Teste: TestGitHooksRepoNaoGitVemAntesDaSelecao. Passou nesta rodada. |
| UX-004 | Média | Comando, flag, hook, skill ou addon desconhecido sem sugestão | corrigido | lista de válidos e "Você quis dizer" (TestComandoDesconhecidoSugereProximo, TestAddonsUnknownListsValidosESugere) |
| UX-005 | Média | Ajuda de `hook` incompleta e `--ui` sem explicação | corrigido com evidência | Commit 3b68ba7. A ajuda de `hook` lista os cinco subcomandos, e a ajuda global explica `--ui` com o modo `auto`. Testes: TestAjudaDoHookListaCincoSubcomandos, TestUsoGlobalExplicaUIAuto. Passaram nesta rodada. |
| UX-006 | Média | Mesmo erro de entrada com códigos de saída diferentes | corrigido com evidência | Commit 3b68ba7. Valor inválido de flag agora é erro de uso (exit 2), como flag sem valor. Testes: TestValorDeFlagNaoPodeSerOutraFlag, TestUIVazioEhErroDeUso. Passaram nesta rodada. |
| UX-007 | Média | `--quiet` com efeito diferente entre comandos | justificado | O plano 8.4 define que `install --quiet --yes` mostra só a prévia e o id do histórico, com stderr vazio. A prévia garante que nada é gravado às cegas. |
| UX-008 | Baixa | Capitalização e prefixo de mensagens diferentes | corrigido com evidência | Commit 3b68ba7. Mensagens de uso e de erro seguem o mesmo formato. Testes: TestReadFailNaoUsaPrefixoErro, TestHookMensagensDeUsoComecamComPrefixo. Passaram nesta rodada. |
| UX-009 | Baixa | Nomes `hook`, `hooks` e `githooks` parecidos e `rollback-agents` com hífen | justificado | Os nomes já foram publicados na v0.2.0. `hook` é uso interno chamado pelos agentes. Renomear quebraria configurações já instaladas. |
| UX-010 | Baixa | `--home` com sentido diferente em `githooks` | justificado | Em `githooks`, o HOME é onde fica o executável estável que os hooks chamam. A ajuda descreve isso. |
| UX-011 | Baixa | Mensagens de estado contraditórias no `doctor` | corrigido | frase única "Nenhuma instalação registrada." |
| UX-012 | Baixa | Prévia repete a lista de arquivos e caminho aparece com `•••` | justificado (em parte) | A tabela resume e a lista detalha. O mascaramento do caminho segue a regra de não exibir valores que parecem segredo (ADR-0025). Melhoria futura: mostrar a lista uma vez só. |
| UX-013 | Baixa | `hook commit-msg` aceitava argumento extra | corrigido com evidência | Commit 3b68ba7. Argumento extra sai com exit 2. |
| PROC-001 | Alta (processo) | Mudança de arquivos fora da sandbox durante a auditoria | aberto (processo) | Ver seção 6. |

Itens fora da lista de 21/10, também corrigidos em 3b68ba7:

- Valor de flag que é outra flag é recusado. Teste: TestValorDeFlagNaoPodeSerOutraFlag.
- `--ui=` vazio é erro de uso. Teste: TestUIVazioEhErroDeUso.

## 5. Lacunas da avaliação de 21/10

- TUI com terminal real não foi exercitada. A cobertura vem só dos testes golden.
- `NO_COLOR` e `--no-color` não foram verificados.
- `rollback-agents` não foi testado, porque o wrapper de auditoria bloqueia o comando.
- Install interativo com confirmação não foi testado.
- Nenhuma operação longa foi cronometrada. Não há evidência sobre progresso (spinner ou barra).
- O achado UX-001 não foi reproduzido na avaliação. Ele veio do relatório de um agente ("Teste B"). Os testes novos cobrem o caso.
- O achado UX-003 também veio do "Teste B".
- O resumo da avaliação de 21/10 diz 5 achados baixos e 12 no total. A lista de achados tem 6 baixos (UX-008 a UX-013) e 13 no total. Esta auditoria segue a lista.

## 6. Incidente de isolamento (PROC-001)

Fatos, conforme o arquivo de relatórios da auditoria:

- Em 21/10/2026, às 23:31:05, o arquivo de configuração do Claude (`~/.claude/settings.json`) e o arquivo de hooks do Codex (`~/.codex/hooks.json`) mudaram. Os dois estão fora do sandbox da auditoria.
- O estado do histórico do kit (`~/.local/share/gex-dev-kit/history/state.json`) tem data de modificação de 23:11:43. Essa data é anterior à janela do incidente. Ou seja, o kit não gravou histórico nesse período.
- O relatório atribui a mudança a um escritor externo ao kit. Essa atribuição é coerente com as provas, mas a origem não foi confirmada.
- O arquivo `~/.codex/hooks.json` agora tem nove hooks de outro aplicativo.

Ação:

- A regressão da seção 8.8 deve repetir a checagem de hash e de data de modificação antes e depois da auditoria.
- Se o kit for confirmado como origem, a regra de isolamento deve ser revisada.
- Até a regressão rodar, o status segue como aberto.

## 7. Comparação com a avaliação de 20/10

A avaliação de 20/10 foi feita sobre a v0.2.0, com correção testada só em binário local. A avaliação de 21/10 foi feita sobre a v0.3.0 (HEAD ddcea84), com sandbox isolado. Por isso as duas não cobrem as mesmas coisas.

| Tema | Avaliação de 20/10 | Avaliação de 21/10 | Situação atual |
|---|---|---|---|
| Nota | Sem nota numérica (checklist não foi executado por inteiro) | Núcleo 3.4 e geral 3.5 | Nota não refeita |
| Ajuda com exemplos | Ajuda sem exemplos | Exemplos em 12 de 12 comandos | Resolvido segundo a avaliação de 21/10 (não reverificado aqui) |
| Menu com `TERM=dumb` ou `ACCESSIBLE=1` | `q` não encerrava e o menu caía direto em skills | Não testado (TUI com TTY fora do escopo) | Pendente de teste. A correção de 20/10 não foi publicada |
| `NO_COLOR` | Não impedia cor truecolor no caminho do menu | Não verificado | Pendente de teste |
| Texto em inglês da biblioteca Huh no menu de skills | Apontado como prioridade | Não citado | Não avaliado |
| PATH ambíguo e confiança do hook Codex | Causou código 1 no `doctor` | Não citado | Não avaliado |
| Erros de entrada | Comando e flag inválidos saíam com código 2 | `githooks --hook` e `install --skill` inválidos saíam com código 1 | Corrigido em 3b68ba7 (UX-002, UX-006) |
| Prévia e segurança de `install` | Sem TTY, `install` sem `--yes` saiu com código 1 e não alterou arquivos | Prévia antes de aplicar e recusa de caminhos fora do sandbox | Mantido. Não houve mudança nesse ponto |
| Rollback | Arquivo editado foi preservado e o original voltou com modo `0640` | `rollback --yes` saiu com código 0 na sandbox | Mantido |

Conclusão da comparação: a avaliação de 21/10 achou falhas novas em erros de entrada e em saída de `doctor --json`. Os temas de acessibilidade da TUI, `NO_COLOR` e PATH da avaliação de 20/10 não foram retestados em 21/10. Eles continuam como lacuna.

## 8. Itens pendentes

- UX-004 e UX-011: corrigidos.
- Regressão da seção 8.8, item 7: executada (26 de 27; a diferença é a correção de UX-002).
- PROC-001: hash do HOME real inalterado na regressão; origem do incidente de 21/10 não confirmada.
- Lacunas da seção 5 e os temas da avaliação de 20/10 que não foram retestados (seção 7).

## 9. Interface atual da TUI (não reavaliada)

- Sem subcomando, `gex-dev-kit` mostra boas-vindas, verifica o ambiente e pede `Enter` ou clique para iniciar o chat. Os comandos TUI abrem diretamente o fluxo correspondente em `RunChat`; `ds` continua como vitrine separada. CLI, JSON e modo acessível continuam disponíveis.
- O chat fica ocioso após a saudação ou ao concluir fluxos; `/` abre a paleta filtrável, sem menu permanente. `/menu` mostra a lista quando solicitado. `/ajuda` lista `/setup`, `/catalogo`, `/doctor`, `/rollback`, `/agentes`, `/githooks`, `/addons`, `/menu`, `/voltar`, `/ajuda`, `/limpar` e `/sair`, além de `esc`, `ctrl+o` e `ctrl+c`. Texto não reconhecido responde `Não entendi. Digite / para ver os comandos.`.
- A conversa usa uma barra escura de contexto: GEX, harness detectado e versão, caminho do projeto (diretório atual) e branch Git quando disponível; o HOME de instalação aparece apenas quando difere do HOME real. Core instalado e versão do kit, ou progresso/etapa do setup, ficam na borda superior rosa da caixa de entrada (`PromptBox`), sem divisória separada acima da transcrição. `GEX_NO_POWERLINE=1` substitui glifos Powerline por `›` e o modo Plain usa ASCII.
- Durante mensagens animadas e ferramentas, o campo mostra `GEX está trabalhando…`, a linha de trabalho mostra spinner, tempo e brilho móvel no rótulo, e texto digitado não é enfileirado. Respostas chegam por palavras inteiras em intervalos variados. Ferramentas ocupam cartões de cantos arredondados: checklist de fases enquanto executam, resultado alinhado ao terminar e tempo na borda inferior. `Enter` ou espaço durante a animação apenas termina de revelar o texto. `Esc` limpa a entrada ou retorna à pergunta anterior; `/voltar` também volta. Nas perguntas múltiplas, espaço ou dígitos marcam e `Enter` confirma (sem marca numa pergunta obrigatória, confirma a opção do cursor). O botão clicável `Confirmar (N)` ou `Pular (nenhuma marcada)` confirma; perguntas de caminho aceitam `Tab` para completar e `Enter` para entrar na pasta destacada. Cliques em opções e na borda superior da ferramenta e roda do mouse são aceitos; a barra lateral aparece a partir de 120×30.
- Escolhas longas são paginadas e mantêm ao menos quatro linhas da transcrição; `←`/`→`, `PageUp`/`PageDown` e `‹`/`›` trocam páginas com o campo vazio. Dígitos valem na página atual. Com texto digitado, setas horizontais editam o campo; fora de lista paginada, `PageUp`/`PageDown` e roda rolam a conversa. Paleta e sugestões de pasta mostram `… mais N` ao exceder o espaço disponível. Catálogo separa Skills por grupos e Referências por tipo.
- O setup anuncia harness, destino, extras, plugins, revisão e instalação. A barra lateral tem sete etapas fixas: Harness, Agentes, Destino, Extras, Plugins, Plano e Instalação. Após Extras, Plugins permite marcar addons para Claude Code e Codex; só OMP pula a etapa. O chat revisa comandos, conflitos e passos manuais antes do resumo e de **Sim, aplicar**. `/addons` oferece revisão e confirmação por plugin e destino; o rollback do kit não desfaz esses plugins.
- A confirmação do setup é **Sim, aplicar** depois da ferramenta **Planejar instalação**; cancelar mostra `Nada foi gravado.`. `doctor` mostra `Tudo íntegro: N item(ns) ok.` ou `Encontrei N item(ns) com atenção.` e usa `report.ExitCode()`. Conflito no rollback mostra `Não restauro: estes arquivos mudaram depois da instalação.` e marca código 1; ausência de histórico mostra `Nada a restaurar.`. Em gravação, `Ctrl+C` é adiado até concluir.
- A cobertura visual usa `TestGoldenChat` (boas-vindas, chat ocioso, setup, ferramenta em execução, confirmação e resultado), no lugar dos goldens individuais das telas removidas. Esta seção descreve código, não prova teste manual de terminal, mouse ou acessibilidade.
