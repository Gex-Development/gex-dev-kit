---
name: gex-verificar-mudanca
description: Use após mudança GEX e antes de afirmar conclusão para verificar critérios, caminhos de erro, testes e evidência executada, com bloqueios explícitos.
---

# Verificar Mudança GEX

## Quando usar
Antes de declarar pronta uma implementação ou preencher evidência de PR. Esta skill não transforma leitura de código em teste executado.

## Passos
1. Recupere demanda, critérios e diff. Mapeie cada critério ao comportamento alterado, teste ou verificação manual observável. Inclua consumidores, config, migração, documentação e casos de borda afetados.
2. Revise caminhos normais e de erro: validação de entrada, falhas externas, autorização, integridade e rollback quando cabível. Confira que teste novo cobre regressão relevante e que nenhum dado sensível vai a logs ou saída.
3. Execute testes da aplicação, build e lint definidos e aplicáveis pelo projeto, ou equivalentes da stack. Guarde comando, alvo, resultado e falha exata; se faltar comando, ambiente ou aplicabilidade, explique. Não troque suíte integral por teste estreito sem avisar.
4. Cruze resultados com critérios. Falha exige correção e nova execução, ou exceção explícita decidida pelo responsável antes da publicação; sem isso, PR bloqueada. Mudança posterior invalida evidência afetada: reexecute.
5. Entregue síntese dos critérios atendidos e não verificados, com evidência e motivos. Diga "passou" só para execução observada.

## Limites
Não invente resultado, não esconda falha e não rode operação destrutiva só para conferir. Se dependência, credencial ou ambiente faltar, declare não verificado e o requisito necessário; sem ambiente não há atestado de sucesso.
