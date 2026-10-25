---
name: gex-canonical-prod-readiness
description: Use para verificar se Canonical de produção está pronta antes de regenerar Gold ou publicar dashboards.
---

# Verificar Prontidão Canonical

## Quando usar

Antes de afirmar que a Silver foi aprovada ou de liberar atualização da Gold e do Hub. Use em verificação somente leitura.

## Pré-requisito

- Perfil AWS somente leitura de produção, para Step Functions e Glue, obtido por canal aprovado. A presença de variável de ambiente não prova acesso.
- Sem esse perfil, a skill não conclui prontidão. Entregue o que foi possível verificar e declare o limite.

## Prova somente leitura

1. Consulte o workflow completo: a state machine diária da Canonical (nome no Terraform do gex-infra-aws). Veja `list-executions` e `describe-execution` da execução mais recente, com `status`, início, fim, tarefa que falhou e causa. Sucesso de um job Glue isolado não prova que o workflow inteiro passou.
2. Diferencie reprocessamentos manuais da execução integral. Se o engenheiro relatar reprocessamento manual, consulte `get-job-runs` dos jobs Glue exatos. Se `glue:GetJobs` estiver negado, extraia os nomes dos jobs da definição da state machine, quando `DescribeStateMachine` estiver permitido, e consulte cada job pelo nome. Registre estado e horários. Um job `SUCCEEDED` isolado não altera o status da state machine.
3. Compare a conclusão de cada job da Canonical com o início e o fim da execução da Gold. A Gold pode começar enquanto algum job ainda escreve; por isso, o sucesso da Gold não prova que ela leu todos os dados reprocessados.
4. Leia os marcadores publicados na tabela Gold do Glue Catalog: `silver_status`, `silver_verificada_em`, `silver_pendencias`, `UpdateTime` e localização. `INCOMPLETA`, ou pendência `FAILED` da state machine diária, significa que a Gold continua marcada como parcial, mesmo se alguns jobs manuais ficaram verdes.
5. Confirme a propagação para o Hub. Consulte a execução mais recente da state machine de extração do Hub (nome na infraestrutura do Hub) e a confirmação de ingestão e publicação no backend. Fila vazia sozinha não prova que o dashboard recebeu a geração atual.
6. Decida a publicação. Para limpar o aviso de dados parciais, exija: execução completa da Canonical com `SUCCEEDED`; Silver marcada como `COMPLETA`; Gold executada depois da conclusão da Canonical; extração do Hub depois da Gold; e publicação confirmada no backend ou no dashboard. Não dispare a extração como solução final enquanto a Silver estiver `INCOMPLETA`: o snapshot pode atualizar e o aviso continua.

## Decisão

Se a Silver estiver `INCOMPLETA`, informe a causa e bloqueie a alegação de prontidão. Não dispare jobs nem altere dados numa verificação somente leitura.

## Limites

Relate claramente as diferenças entre o workflow do Step Functions, os jobs Glue individuais, a Gold e a publicação no Hub. Esta verificação mostra o estado publicado; não corrige dados.
