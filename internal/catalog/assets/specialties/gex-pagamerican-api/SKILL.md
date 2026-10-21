---
name: gex-pagamerican-api
description: Consulte metadados de uma janela curta da PagAmerican Data API v1, só leitura, sem expor registros de clientes.
---

# Consulta PagAmerican

Use para conferir disponibilidade, paginação e volume declarado de uma rota numa janela de **um dia**. Não confunda esta consulta de metadados com auditoria/reconciliação completa.

## Pré-requisito

Python 3.10+ e chave de leitura `PAGAMERICAN_API_KEY` fornecida no ambiente, por canal corporativo aprovado. Para selecionar o pack, a CLI confere presença da variável no processo; não guarda a chave, testa validade nem faz GET durante a instalação. Não ponha chave em argumento, arquivo versionado ou saída. Sem credencial, pare e peça acesso pelo canal aprovado.

## Executar, com autorização para consulta externa

De qualquer diretório, use o caminho da skill instalada; exemplo para Claude Code:

```bash
python3 "$HOME/.claude/skills/gex-pagamerican-api/scripts/metadata.py" purchases 2026-10-01
```

Também disponível em `$HOME/.agents/skills/gex-pagamerican-api/` para agentes que leem esse diretório. Substitua data pelo dia autorizado. Script faz **um GET**, primeira página de até 100 registros, timeout 15 s; imprime somente metadados numéricos da página, nunca corpo ou IDs. Rotas permitidas: purchases, orders, refunds, chargebacks, checkouts, abandoned-carts. Filtro `from` é inclusivo e `to` exclusivo; não mescle com `updatedFrom`/`updatedTo`. Retenção é móvel, então HTTP 422 não significa ausência histórica. `meta.truncated` sinaliza janela incompleta. `meta.count` pode ser apenas tamanho da página; não interprete como total do dia. Ordenação das páginas pode empatar no mesmo timestamp; para auditoria integral, use pipeline e contratos do projeto, com deduplicação por ID e reconciliação de contagens. Para refunds, separe alerta de reembolso consumado via `refundSource`; não agregue linhas brutas como dinheiro.

## Proveniência e limites

Orientação original baseada nos contratos da PagAmerican Data API v1 e no comportamento de integração documentado internamente; nenhum código ou dado pessoal foi copiado. Endpoint `https://external-api-service.pagamerican.app/api/v1`; valide documentação vigente e permissões antes da consulta. Não executa durante instalação, não modifica API nem consulta dados de detalhe. Script apenas metadados da primeira página; não serve para extrair dataset.
