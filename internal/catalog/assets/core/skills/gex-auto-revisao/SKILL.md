---
name: gex-auto-revisao
description: Use antes de abrir PR GEX para revisar a própria mudança contra issue, aceite, código, riscos e evidências, sem editar nem publicar.
---

# Auto-Revisão GEX

## Quando usar
Depois da implementação, antes de preparar a PR. Revise o próprio diff. A revisão de uma PR já publicada usa gex-revisar-pr.

## Escrita
Comece pela conclusão e ação necessária. Use frases curtas e tópicos pequenos. Explique termos quando ajudarem o leitor; mantenha cenário, evidência e correção claros.

## 1. Entenda o pedido
- Leia issue, critérios de aceite, referências e instruções locais. Sem acesso à issue, declare o limite; não presuma requisitos.
- Compare cada critério com a entrega. Identifique ausência, atendimento parcial e mudanças fora do escopo. Confira se a solução resolve a causa, não só o sintoma.

## 2. Procure problemas reais
- Prefira a menor solução correta. Aponte abstrações, dependências, configurações, fallbacks e duplicações sem necessidade demonstrada. Evite refatoração estética ou arquitetura hipotética.
- Confira nomes, legibilidade, padrões do projeto, código morto, comentários explicativos novos, logs e resíduos de desenvolvimento.
- Verifique lógica, entradas inválidas, erros, segurança, casos de borda, contratos, regressões e integridade. Considere concorrência, idempotência e efeitos colaterais quando couber.
- Avalie consultas, memória e operações repetidas se houver impacto concreto. Não invente defeitos hipotéticos.

## 3. Verifique o comportamento
- Confronte critérios com testes úteis de sucesso, falha e regressão. Não trate ausência de teste como defeito reproduzido.
- Execute verificações pertinentes disponíveis e autorizadas. Registre comando, resultado e limites; sem execução, não diga que passou. Falha exige correção e nova verificação antes da PR.

## 4. Relate o resultado
- Liste só achados fundamentados: [BLOCKER] para perda de dados, falha grave ou requisito essencial; [MAJOR] para bug ou regressão relevante; [MINOR] para melhoria objetiva de manutenção, legibilidade ou consistência. Não liste NIT nem preencha relatório com apontamentos artificiais.
- Para cada causa, indique arquivo/linha, cenário, impacto, critério relacionado e correção mais simples. Evite duplicar efeitos do mesmo defeito.
- Resuma critérios atendidos, ausentes e não verificados; verificações feitas e veredito técnico: aprovável, precisa correção ou falta verificação.

## Limites
Esta é análise privada anterior à PR. Não edite arquivos, crie commits, publique comentários ou PR, aprove ou faça merge. Não assine nem inclua referência à ferramenta de IA. Depois, o desenvolvedor corrige achados e repete verificações afetadas antes de preparar a PR.
