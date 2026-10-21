---
name: gex-preparar-pr
description: Use antes de abrir PR GEX para confirmar issue real, assignee, reviewers, auto-revisão e verificações aplicáveis.
---

# Preparar PR GEX

## Quando usar
Ao entregar mudança revisável. Siga template, branch e instruções do repositório; obtenha autorização antes da publicação. Esta skill não publica nem faz merge por conta própria.

## Escrita
Abra pela mudança e decisão necessária; depois contexto e evidência. Use frases curtas, verbos concretos, tópicos pequenos e siglas explicadas quando necessário. Evite densidade, jargão gratuito e duplicação; adapte ao leitor sem perder rigor.

## Passos
1. Leia diff contra base pretendida, issue e aceite. Mapeie arquivos, consumidores, contratos, migrações e documentação. Retire apenas resíduos da própria mudança; não descarte trabalho alheio.
2. Faça gex-auto-revisao antes da PR; corrija problemas reais. Execute testes da aplicação, build e lint definidos e aplicáveis pelo projeto (ou equivalentes da stack). Registre comando, resultado e critérios cobertos. Comando ausente, ambiente indisponível ou verificação não aplicável exigem motivo explícito, nunca sucesso inferido. Mudança posterior que invalide evidência exige repetir verificação afetada. Falha bloqueia publicação até correção ou exceção explícita do responsável.
3. Prepare quatro itens obrigatórios: (a) título e descrição enxutos, com o que mudou, evidência, riscos e pendências relevantes; (b) reviewers elegíveis segundo CODEOWNERS/convenção e acesso, nunca o autor; (c) assignee igual ao usuário GitHub autenticado, não o agente; (d) vínculo com issue real conferida, sem inventar número. Use referência explícita que não feche issue por engano; `Closes #` apenas se PR realmente a resolve.
4. Confirme permissão de escrita para solicitar revisão, capacidade de atribuir assignee e acesso dos reviewers. Se faltar issue, reviewer elegível, assignee ou permissão, relate bloqueio antes de publicar; não deixe pendência silenciosa nem simule configuração.
5. Confira destino, base, dados sensíveis, autorização e consistência com diff. Ao publicar por fluxo autorizado, confirme link, vínculo, assignee, reviewers e estado efetivo; se solicitação falhar, informe o resultado real.

## Saída e limites
PR explica a mudança com evidência concisa. Conventional Commits valem para commits GEX; título de PR segue regra local. Não invente issue, revisor, resultado de teste ou autoria; preserve divulgação legal e autoria real. Sem condições de publicação, entregue rascunho e bloqueio exato.
