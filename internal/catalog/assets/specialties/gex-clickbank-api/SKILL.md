---
name: gex-clickbank-api
description: Consulte totais agregados da ClickBank Quickstats por conta e período autorizado, sem extrair pedidos nem dados de clientes.
---

# Consulta ClickBank Quickstats

Use para obter valores agregados de sale, refund e chargeback via REST 1.3 e conferir janela da conta. Não use esses totais como prova de reconciliação diária contra `orders2`: eventos e períodos têm semânticas diferentes.

## Pré-requisito

Python 3.10+. Credencial `CLICKBANK_API_KEY` e nickname autorizado `CLICKBANK_ACCOUNT` fornecidos por canal corporativo aprovado no ambiente local. API key precisa de **API Order Read** para nickname. Não prefixe key com `Bearer`. Para selecionar o pack, a CLI confere presença dessas variáveis no processo; não guarda chave, testa validade nem faz GET durante a instalação. Não grave token em script, argumento ou log. Sem permissão, solicite acesso, não contorne 403.

## Executar, após autorização de consulta

```bash
python3 "$HOME/.claude/skills/gex-clickbank-api/scripts/quickstats.py" 2026-10-01 2026-10-02
```

Alternativa: `$HOME/.agents/skills/gex-clickbank-api/`. Datas inclusivas; escolha período pequeno e permitido. Um GET em `/rest/1.3/quickstats/count`, timeout 15 s, `Accept: application/json`; sem paginação nem pedidos individuais. Saída contém só três valores agregados; não publique números financeiros fora do canal autorizado. `quickStatDate = null` em `/count` é esperado. Se schema mudar, script falha em vez de imprimir resposta crua.

## Limites e fonte

`orders2/list` pagina pelo header `Page`; HTTP 206 indica mais páginas. Este pack **não** extrai `orders2`, Analytics nem Shipping. Para auditoria bronze/silver, leia contrato do projeto, compare grão de transação e trate BILL, RFND, CGBK e testes separadamente. Referência oficial: [ClickBank APIs](https://support.clickbank.com/en/articles/10535400-clickbank-apis). Texto e código originais, sem copiar scripts, dados pessoais ou credenciais da integração individual.
