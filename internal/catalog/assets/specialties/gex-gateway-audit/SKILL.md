---
name: gex-gateway-audit
description: Use para auditar um gateway no vault canônico, investigar FAILs e separar defeitos reais de falhas de medição.
---

# Auditar Gateway

## Quando usar

Em auditoria profunda de gateway autorizado no vault canônico de gateways. Leia o contrato, os checks e as instruções locais do vault antes do primeiro comando.

## Pré-requisito

- Clone autorizado do vault canônico de gateways. Execute os comandos na raiz do clone, depois de ler o `CONTEXT.md` e a skill `auditar` do próprio vault.
- Perfil AWS somente leitura de validação, configurado no vault.
- Para `--api`, credenciais de gateway do próprio vault, recebidas por integração segura.
- Esta skill não traz ferramenta nem credencial. O comando depende do vault e de permissões específicas, não da CLI global.

## Passos

1. Gere o inventário de cargas: `python3 -m tools.audit cargas --gateways <gateway>`. Rode em segundo plano, se o agente suportar. Anote os jobs em execução e `pos_publicacao`: jobs por loja podem ser posteriores à publicação mesmo quando o job global não é.
2. Rode a auditoria profunda: `python3 -m tools.audit run --gateways <gateway> --modo profundo --api`. Sem datas, a ferramenta usa o período padrão dela. A saída fica em `05_Auditoria/runs/<data>-<gateway>-historia/`, com `relatorio.md`, `resultados.json` e `sql/`.
3. Faça a triagem de `resultados.json`. Agrupe os ERRO pela mensagem. Muitos são candidatos a falso FAIL: `CONTRATO_SAIDA` com o placeholder `CAN-REFAC-FASE5` e `COLUMN_NOT_FOUND` com coluna fora do schema. Não investigue esses grupos agora; liste-os na retro. Checks vigentes com `NAO_APLICA` costumam ser checks de outro gateway.
4. Investigue por fatias. Se o agente tiver subagentes, divida por tema: total e rateio; IVA e câmbio; fulfillment; URL, dimensões e chaves estrangeiras; API de pedidos-prova. Sem subagentes, faça em sequência. Cada fatia recebe estas regras: consulte com o perfil somente leitura de validação; rode SQL só por `python3 -m tools.audit sql --arquivo X --run <run>`; grave o SQL em `<run>/investigacao/<fatia>/`; leia o script publicado no S3; use valores em centavos inteiros; faça CAST para double antes de dividir; não inclua dados pessoais; não edite checks, norma ou registro; não abra issue. O contrato de saída é `<run>/investigacao/<fatia>.md`, com a classificação (`defeito`, `decisao`, `limitacao`, `saneamento`, `regua_errada` ou `inconclusivo`), o n depurado, o impacto e um rascunho de issue.
5. Se uma fatia achar bug no instrumento, como fonte duplicada, avise as outras frentes pelo agente principal para conferirem a mesma causa.
6. Registre cada achado com `python3 -m tools.audit achado --id 'CHK-X|<gateway>' --classe <defeito|decisao|limitacao|saneamento> --nota '...'`. O registro não aceita a classe `regua_errada`: use `limitacao` e escreva "régua errada" na nota.
7. Escreva `<run>/retro.md`: placar bruto e depurado, contraprovas, correções propostas para a régua, decisões humanas, lacunas do catálogo e lista de rascunhos de issue. Antes de terminar, busque no diretório da run e-mails, CPF e tokens de autorização (`Bearer`).
8. Responda ao usuário em português simples: veredito curto, tabela de defeitos reais, erros de régua, decisões e caminhos. Não crie issue sem autorização do responsável.

## Sinais conhecidos de falso FAIL

- 100% de violações (n/n); tolerância mais rígida que a norma (CAN-FX-001 = R$ 0,03 por linha); mesma fonte contada duas vezes; valores `NaN` no payload; check global replicado por gateway; checks de continuidade mensal sem fonte independente.
- CartPanda: deduplicação por `(shop_id,id)`, não por `(loja,id)`, porque a mesma loja pode aparecer em dois prefixos; `exchange_rate_usd='NaN'`; `requires_shipping=1` mesmo em linhas de frete; `created_at_*` com `-03:00` desloca a janela em 3 horas.

## Limites

Separe defeito real de falha de medição antes de propor correção. Um FAIL é hipótese até a contraprova read-only.
