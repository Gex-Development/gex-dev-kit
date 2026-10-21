---
name: gex-regras-de-negocio
description: Use em mudança GEX que afeta dados ou comportamento de domínio para rastrear invariantes da fonte até código, testes e cenários de erro.
---

# Regras de Negócio GEX

## Quando usar
Quando cálculo, estado, integração, permissão, join ou persistência puder alterar resultado de negócio. Sem fonte confiável, não declare regra como aprovada.

## Passos
1. Localize fonte vigente: issue, contrato, documentação do projeto e decisões explícitas do responsável. Marque versão e conflito entre fontes. Diferencie requisito confirmado, comportamento observado e hipótese.
2. Escreva invariantes verificáveis: entradas válidas, saídas, transições, cardinalidade, unidade, arredondamento, duplicidade, ausência de dados e idempotência onde relevantes. Identifique quem é dono do dado e consumidores afetados.
3. Trace cada invariante até ponto de implementação e teste. Examine transformações, joins, filtros, validações e tratamento de erro. Para integração externa, confira contrato real e condições de falha; não suponha que COALESCE corrige multiplicação de linhas por join.
4. Prepare cenários com base e bordas que distingam regra correta de resultado plausível errado. Verifique com dados não sensíveis e testes do projeto quando possível; reporte comando, resultado e limites. Não declare reconciliação sem comparar fontes autorizadas.
5. Liste lacunas, impacto e decisão necessária. Se regra contradiz fonte ou pode corromper dado, interrompa conclusão até responsável esclarecer. Leve critérios e evidência à issue ou PR.

## Limites
Não acessar produção, mudar dado real ou chamar API de gateway sem autorização e integração habilitada. Não inserir credenciais nem dados pessoais em exemplos. Especializações de gateway, Spark, Gold e dashboards dependem da função e ficam fora do Core.
