---
name: gex-buygoods-api
description: Consulte metadados de orders ou transactions da BuyGoods por conta e janela curta, sem imprimir pedidos nem dados pessoais.
---

# Consulta BuyGoods

Use para conferir se orders ou extrato transactions respondem e qual total de registros a paginação declara para uma conta. **Não** equivale a reconciliar vendas, refunds ou chargebacks no canônico.

## Pré-requisito

Python 3.10+. Obtenha autorização de leitura e defina `BUYGOODS_ACCOUNT_ID` e `BUYGOODS_API_TOKEN` no ambiente usando canal corporativo aprovado; são valores próprios de cada conta. Para selecionar o pack, a CLI confere presença dessas variáveis no processo; não grava token, testa validade nem faz GET durante a instalação. Não passe token em argumento nem registre saída de dados brutos. Sem credencial, não rode consulta. Permissões e limites dependem do contrato atual da conta.

## Executar, só após autorização

```bash
python3 "$HOME/.claude/skills/gex-buygoods-api/scripts/metadata.py" orders 2026-10-01 2026-10-02
python3 "$HOME/.claude/skills/gex-buygoods-api/scripts/metadata.py" transactions 2026-10-01 2026-10-02
```

Alternativa: `$HOME/.agents/skills/gex-buygoods-api/`. Substitua janela pelas datas autorizadas. Cada execução faz um GET com `limit=1`, `offset=0`, timeout 15 s; mostra só `pagination.total`. Não imprime linha, token, ID ou corpo de erro. URLs: `https://admin3.buygoods.com/public/v2/orders.php` e `/transactions.php`. O endpoint de customers está fora do escopo; o antigo host de staging não é usado. Se o gateway retornar 403, verifique permissão/contrato com responsável, não tente contornar bloqueio.

## Interpretação

`orders` e `transactions` não possuem o mesmo grão: transações são lançamentos no ledger, inclusive liquidação; não compare totais brutos como se fossem número de vendas. Para reconciliação verdadeira no `gex-infra-aws`, fixe conta, período, regras do adaptador e SQL de leitura de Athena em ambiente aprovado. Não execute writes nem use data de criação como substituto automático de `updated_at`. Conteúdo é orientação original baseada em contrato observado da API de integração; código privado, dados de contas e credenciais não foram incorporados. Confirme especificação vigente com o mantenedor antes de uso financeiro.
