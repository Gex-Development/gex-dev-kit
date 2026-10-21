---
name: gex-revisar-pr
description: Use para revisar PR GEX publicado por SHA, confrontando issue, diff, contratos e evidência; comentar problemas reais nas linhas alteradas.
---

# Revisar PR GEX

## Quando usar
Quando revisão ou re-review de PR publicado for solicitada ou autorizada. É distinta da auto-revisão privada antes da PR; não é aprovação automática.

## Escrita
Comece pela conclusão e ação necessária; depois mostre cenário e evidência. Use frases curtas, verbos concretos e tópicos pequenos. Explique siglas quando necessário. Evite densidade, jargão gratuito e repetição; adapte ao leitor sem perder rigor.

## Passos
1. Identifique repositório, base, HEAD SHA, issue e instruções locais. Fixe SHA analisado. Sem diff completo ou contexto essencial, registre limite, não conclua revisão global.
2. Leia diff, consumidores, contratos, dados, permissões e testes. Confronte aceite e examine comportamento normal, erros, segurança, bordas, concorrência e reversibilidade onde relevante.
3. Registre somente achados concretos. Cada problema recebe gravidade [BLOCKER], [MAJOR] ou [MINOR] conforme impacto real; não gere apontamento gratuito nem preferência estética. Dê arquivo, linha, cenário, impacto e correção mais simples. Um problema por comentário; reúna causa única, não duplique efeitos do mesmo defeito.
4. Quando publicação da revisão estiver solicitada ou autorizada, publique comentários inline diretamente no GitHub, nas linhas alteradas, ancorados no menor bloco elegível. Se não existir linha elegível, use resumo com arquivo/linha e explique por que não coube inline. O resumo geral contém só veredito técnico curto e pontos não-inline; não replique achados inline nele.
5. Execute testes apenas com ambiente e autorização; relate comando e resultado observado. Não transforme ausência de teste em falha reproduzida nem exponha dados sensíveis.
6. Confira HEAD antes de publicar. Se mudou, reavalie trechos afetados e indique SHA revisado. Em re-review, confirme cada achado anterior no novo diff como resolvido, persistente ou inconclusivo, com evidência.

## Saída e limites
Relate SHA, veredito técnico, riscos e verificações reais. Não edite arquivos, não crie commits, não publique comentários/PR sem pedido ou autorização. Não submeta aprovação formal nem request changes formal sem autorização explícita. Não assine review, comentário ou PR nem acrescente rótulo de IA; preserve autoria real e divulgação obrigatória.
