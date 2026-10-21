---
name: gex-hub-gold-refresh-prod
description: Use para atualizar Gold existente em produção e publicar snapshot no Hub após mudança Canonical autorizada.
---

# Atualizar Gold e Publicar Snapshot

## Quando usar

Depois de mudança na Canonical ou no Glue, para atualizar a tabela Gold do Hub em produção e publicar o snapshot no worker do GEX Hub. Para promover código por PR, siga o fluxo de release do projeto.

## Pré-requisito

- Perfil AWS de produção autorizado, na região definida no projeto, com permissão de execução.
- Autorização explícita para cada execução. Sem ela, entregue só diagnóstico.

## Antes de executar

1. Valide a origem e o frescor. Consulte a última execução completa da state machine diária da Canonical e inspecione os jobs Glue separadamente. Um job Glue manual bem-sucedido não muda uma execução com falha da Step Functions.
2. Consulte a tabela Glue da Gold (nome no projeto): parâmetros e localização. `silver_status=INCOMPLETA` é o marcador que a extração do Hub lê e envia adiante; ele pode continuar depois de reexecuções manuais.
3. Compare o horário de conclusão do último job da Canonical com os horários do job Gold, da state machine e dos crawlers. Gold gerada antes do último job da Canonical não é o snapshot final.
4. Se o status continuar `INCOMPLETA`, peça decisão explícita: atualizar o dashboard com aviso ou concluir antes a orquestração completa. Nunca reescreva metadados de status para esconder a pendência.

## Execução autorizada

Só depois de autorização explícita:

1. Execute a state machine da Gold (nome no projeto) depois que todos os jobs manuais aprovados da Canonical terminarem. Use o input definido no projeto. Aguarde o estado terminal `SUCCEEDED`.
2. Releia a tabela Glue da Gold: `UpdateTime`, `StorageDescriptor.Location`, `silver_status` e `silver_pendencias`. Uma Gold bem-sucedida ainda pode trazer `INCOMPLETA` da última orquestração completa da Canonical. Avise o usuário quando o aviso continuar.
3. Inicie a extração do Hub somente depois de a Gold terminar com sucesso. Use um horário de agendamento novo, em UTC. Não reutilize horário anterior: a idempotência da consulta no Athena e a ordem de geração dependem dele. Aguarde o estado terminal e registre `messageId`, `queryExecutionId` e `runId` da saída.

## Verificação honesta

1. Confira os contadores da fila principal do SQS depois que o worker consumir. Não receba mensagens manualmente.
2. Execução bem-sucedida da state machine e um `messageId` provam apenas o enfileiramento. Fila vazia com API saudável são evidências de consumo e disponibilidade, não prova de que o banco de reporting gravou a publicação como `PUBLISHED`.
3. Confirme a ingestão real pelos registros de carga e publicação do banco de reporting ou por tela autenticada do Hub. Se não houver acesso a esses registros, declare o limite em vez de afirmar publicação. Resposta `healthy` do `/health` da API não é checagem de ingestão.

## Limite

Sem permissão de produção, entregue diagnóstico somente leitura e o bloqueio. Não anuncie dados novos no dashboard sem confirmação no consumidor.
