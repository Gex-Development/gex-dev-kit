---
name: gex-fluxo-desenvolvimento
description: Use ao iniciar mudança GEX para escolher fluxo proporcional, localizar instruções do projeto e coordenar issue, implementação, verificação e PR sem impor Addons.
---

# Fluxo de Desenvolvimento GEX

## Quando usar
No início de feature, correção ou ajuste de código. A skill organiza trabalho; não substitui regras locais nem decide escopo de negócio sozinha.

## Passos
1. Leia demanda, instruções do repositório e estado atual. Defina resultado observável, limites, dependências e riscos antes de editar. Para bug, reproduza ou isole causa com evidência; para mudança pequena, registre critério conciso.
2. Escolha tamanho do trabalho pelo risco e alcance, não por ritual único. Use issue existente ou gex-abrir-issue antes de qualquer PR. Se requisitos ou autorização forem insuficientes, explicite lacuna e obtenha decisão antes de ação irreversível.
3. Planeje passos mínimos para mudança multifile ou contratual; preserve convenções e resolva causa, não só sintoma. Se Addons aprovados estiverem disponíveis, siga o método pertinente sem duplicar instruções; Superpowers e Spec Kit são opcionais.
4. Implemente mudanças e testes proporcionais. Use gex-regras-de-negocio para invariantes e gex-verificar-mudanca para critérios/evidências. Não confunda TDD com hook; quando exigido, observe e registre ciclo red/green/refactor.
5. Faça gex-auto-revisao antes de gex-preparar-pr. Corrija achados, reexecute verificações afetadas, prepare PR com issue real, assignee e reviewers elegíveis; relate riscos e bloqueios sem anunciar conclusão prematura.

## Limites
Não instalar Addons, conectores ou hooks por inferência. Não publicar sem autorização. Se um fluxo local divergir, siga instruções aplicáveis e exponha conflito real; não apague decisões do projeto.
