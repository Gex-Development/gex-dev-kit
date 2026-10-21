---
name: gex-pr-review
description: Use para revisar PRs do gex-infra-aws contra issue, norma e comportamento PySpark/Iceberg, com SHA fixo.
---

# Revisar PR de Dados

## Escopo

Esta especialização complementa `gex-revisar-pr` do Core quando o diff envolve pipelines do gex-infra-aws.

## Pré-requisito

- `gh` autenticado com leitura no gex-infra-aws e no vault canônico da organização. Publicar review ou mudar labels exige escrita e autorização explícita.
- Skill `gex-local-spark-tests`, obrigatória, para a prova PySpark/Iceberg local.
- Clone autorizado do vault canônico da organização, para ler a norma CAN/CHK vigente.

## Prova

1. Fixe o SHA da PR, a branch base e o SHA da norma CAN/CHK vigente. Consulte as PRs abertas ao vivo (`gh pr list --state open`). Leia a issue citada, o diff, os contratos, as reviews e os checks. Trate alegações do autor como hipóteses até testá-las. Em re-review, compare o SHA atual com o SHA do último review; não repita review sem delta.
2. Execute o caminho alterado sob PySpark/Iceberg local, com a skill `gex-local-spark-tests`. Use blocos reais do job, não réplicas de lógica, inspeção textual ou mocks. Inclua replay, transições e isolamento de gateway vizinho. Em escrita de tabela, rode o SQL real do job (MERGE, DELETE ou append) no Iceberg local e prove o replay idempotente. Uma mutação local da correção deve fazer o teste relevante falhar.
3. Antes de rodar testes desconhecidos, inspecione-os: alguns chamam Glue ou AWS. Não rode testes que acessem AWS, Glue ou Athena. Use credenciais AWS falsas, perfil vazio e metadata desabilitado.
4. Reproduza cada achado anterior e classifique-o como **CONTEMPLADO**, **PARCIAL** ou **NÃO CONTEMPLADO**, com cenário e resultado observáveis.
5. Se a PR avançou durante a análise, revise o delta, reteste o head final ou limite explicitamente a cobertura ao SHA testado.
6. Para PRs de fulfillment, leia as issues citadas na PR e o guard estrutural do projeto.

## Parecer

- Cada achado traz gravidade (`[BLOCKER]`, `[MAJOR]` ou `[MINOR]`, como no Core), `arquivo:linha`, cenário reproduzível e impacto. Defeito preexistente fora do diff é observação.
- Redija o parecer curto, com status, descrição, SHA, achados, testes e limitações. Separe teste estrutural de prova comportamental.
- Publique com `gh pr review` somente com autorização. Reconfirme o SHA imediatamente antes de publicar. Se o estado mudou, remova o label de review anterior (labels de estado do repositório, se existirem) antes de aplicar o novo. Confira o review e os labels no GitHub depois da publicação.
- Limpe clones, scripts, venv e JARs temporários criados para a revisão.

## Limites

Fixture local prova comportamento local. Não prova incidência histórica, carga real, checagem pós-reprocessamento nem aceite de produção. Se o runtime faltar, declare a prova bloqueada ou inconclusiva, sem alegar validação. Não use painel nem memória como estado atual do GitHub; confirme ao vivo.
