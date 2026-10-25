---
name: gex-design-system
description: Use ao criar qualquer entregável visual com identidade GEX: dashboard, página, slide, relatório, mock ou protótipo, nos temas claro e escuro. Para editar código dentro de gex-hub-dashboards, use gex-hub-design-system.
---

# GEX Design System

Este é o design system fechado da GEX, em versão exportada em 08/10/2026 do GEX Hub (o hub de dashboards da empresa). Vale para tudo o que a GEX coloca na frente de alguém: dashboards, apresentações, relatórios, telas de sistema e páginas.

Para editar código **dentro** de `gex-hub-dashboards`, use `gex-hub-design-system` e confira os tokens CSS locais atuais. Este export serve para entregáveis visuais GEX em geral e não substitui a fonte de verdade do repo.

Os arquivos:

| Arquivo | O que é |
| --- | --- |
| `SKILL.md` | Este documento: as regras e a anatomia de cada peça |
| `tokens.css` | Todos os valores (cor, tipo, espaço, raio, sombra) como variáveis CSS, nos dois temas |
| `assets/gex-logo-branco.svg` | Logo monocromática branca, para fundo escuro |
| `assets/gex-logo-escuro.svg` | Logo monocromática azul-marinho, para fundo claro |

As logos coloridas e o símbolo existem só em PNG e não fazem parte deste pacote. Peça-os ao responsável pela marca se a peça precisar deles.

## Como usar

1. Leia este documento inteiro antes de desenhar.
2. **HTML (dashboard, página, apresentação em HTML):** carregue a fonte Geist no `<head>` e cole o conteúdo de `tokens.css` num `<style>`. Use as variáveis (`var(--surface-card)`), nunca hex solto. Para a logo, cole o SVG inline.

```html
<link href="https://fonts.googleapis.com/css2?family=Geist:wght@400;500;600;700;800&family=Geist+Mono:wght@500;700&display=swap" rel="stylesheet">
```

3. **PowerPoint, Google Slides, Keynote ou PDF:** use os valores de `tokens.css` e a logo SVG convertida para PNG. A fonte Geist é gratuita no Google Fonts. Se não der para instalar, use Inter.
4. Na dúvida entre duas opções, fique com a mais sóbria: a marca aparece pela cor azul, pelo degradê do topo e pela tipografia, nunca por enfeite.

## O que nunca muda

- A marca é **azul `#1e6fd9` com ciano `#38bdf8`**. Roxo e índigo não são cor de marca.
- A assinatura é o **degradê azul-marinho no topo**, com um brilho ciano. Efeito de vidro (fundo translúcido com desfoque) só vai em cima desse degradê.
- O **tema escuro é o padrão** e o claro é completo: o que um tema tem, o outro tem também, com valores próprios.
- **Geist** em tudo. **Geist Mono** só em rótulos pequenos em caixa alta (eyebrow, grupo de colunas da tabela, cabeçalho das colunas do tooltip).
- Títulos com peso 500 e espaçamento de letras −0,03em, em qualquer tamanho grande.
- **Números sempre tabulares** (`font-variant-numeric: tabular-nums`), no padrão brasileiro: vírgula decimal, ponto de milhar.
- **A cor carrega significado:** verde é positivo, vermelho é negativo ou fora da meta, amarelo é aviso. Cor não é decoração.
- Ícones **Lucide**, traço 2. **Sem emoji.**
- Texto em português do Brasil, curto, direto e sem exclamação.

## Temas e contraste

No `tokens.css`, o claro fica em `:root` e o escuro em `.dark`. Para o tema padrão, coloque `class="dark"` no `<html>`.

**Qual tema usar:** dashboard, sistema e apresentação na tela usam o escuro. Relatório, documento para imprimir e deck que vai por PDF ou e-mail usam o claro. Se a peça tiver os dois temas, ofereça um botão de sol/lua no topo.

**Contraste.** Os tokens já foram auditados: use-os como estão, sem ajustar valor. Texto que precisa ser lido vai em `--text-primary` ou `--text-secondary`. `--text-muted` é só para rótulo auxiliar (eixo, placeholder, nota de rodapé). Cor nova que você precisar criar (uma família de calor, uma série) tem que dar pelo menos 4,5:1 se for texto e 3:1 se for marca de gráfico, nos dois temas. No claro, verde, vermelho, laranja e azul de texto e de gráfico usam o tom mais escuro do par (os tokens já trazem isso). Se faltar contraste, ajuste a cor nova. Não tire a peça.

## Marca e logo por fundo

As cores de marca estão nos tokens `--brand-*`, `--accent-ink` e `--brand-gradient-logo`.

**Logo.** Sobre fundo escuro (o topo azul-marinho ou o preto), use a branca (`gex-logo-branco.svg`). Sobre fundo claro, use a azul-marinho (`gex-logo-escuro.svg`). Nunca use a logo branca em fundo claro. A altura padrão é 22 px na barra do topo e 18 px no rodapé. Na barra do topo, a logo vem seguida de um divisor vertical fino (1 × 18 px, `rgba(255,255,255,.28)`) e do nome da peça em 15 px, peso 600.

**Degradê do topo (hero).** É uma faixa no alto da página com o título, onde os filtros ficam sobre vidro.

- Escuro: `--hero-gradient`, que morre no preto da página por trás dos primeiros cards.
- Claro (dashboard): `--hero-gradient-dash`. Todo o conteúdo do topo fica sobre o azul-marinho, e logo abaixo um desbotamento de 260 px (`--hero-skirt-dash`) vai até a cor da página, por trás dos primeiros cards. Nunca reserve espaço vazio só para o degradê.
- Por cima, nos dois temas, vai o brilho `--hero-glow`, um radial ciano no centro do topo.
- O texto sobre o hero é branco (`--hero-text`), e o texto secundário é `rgba(255,255,255,.78)`.
- Vidro sobre o hero: fundo `rgba(255,255,255,.07)`, borda `rgba(255,255,255,.16)` e desfoque de 12 px. No escuro, a opção escolhida é `rgba(255,255,255,.16)` com texto branco. No claro, pílula **branca** com texto `#0b1526`.

## Uso da marca

A licença formal de uso das logos em materiais externos está a confirmar. Até a confirmação, use a marca só em materiais internos e consulte o responsável pela marca antes de publicar em outro contexto.

## Cores

Use sempre o token semântico de `tokens.css` (`--surface-*`, `--text-*`, `--border-*`, `--positive*`, `--negative*`, `--warning*`). Não escreva hex solto em HTML.

- O vermelho só aparece no que é o problema. Em chip de alerta, use vermelho claro sobre fundo vermelho translúcido, nunca vermelho escuro chapado nem rosa.
- Texto de aviso sobre `--warning-bg` usa `--warning-ink`, não `--warning`.
- Métrica que chega fora do padrão de vendas usa a cor do papel que ela tem (ver Dados e gráficos).

## Dados e gráficos

### Uma cor por indicador

Cada indicador principal tem a sua cor, usada no ícone, no fundo do chip do ícone e na minilinha do card:

- Receita: `--kpi-revenue`
- Pedidos: `--kpi-orders`
- Ticket médio (AOV): `--kpi-aov`
- CPA (custo por aquisição): `--kpi-cpa`
- Reembolso: `--kpi-refund`
- Custo de produto: `--kpi-product-cost`
- Margem: `--kpi-margin`
- Valor negativo: `--kpi-negative`

**Métrica que não é de vendas** (RH, financeiro, operação) pega emprestada a cor pelo papel que ela tem, e a cor é a mesma em todos os lugares da peça. Volume ou contagem usa `--kpi-orders` (headcount, chamados, entregas). Entrada, ganho ou resultado bom usa `--kpi-revenue` (admissões, receita, economia). Taxa ou média usa `--kpi-aov` (ticket, salário médio, tempo médio). Custo usa `--kpi-cpa` ou `--kpi-product-cost` (custo por contratação, folha). Saída ou perda usa `--kpi-refund` (desligamentos, cancelamentos).

Duas métricas na mesma peça nunca dividem a mesma cor. Se faltar cor, use as `--stage-*` na ordem.

### Séries e categorias

- **Barra:** degradê vertical de `--chart-bar` (embaixo) até `--chart-bar-top` (em cima). Barra de outra cor (a cor de uma métrica): degradê da cor cheia embaixo até ela mesma clareada em cima. Cantos de cima com raio 4 e largura máxima de 56 px.
- **Linha da segunda métrica** (eixo direito, por exemplo margem %): `--chart-mc-line`, laranja, com curva suave e traço de 2,25 px. A linha tem a cor do indicador dela. Se essa cor for da mesma família das barras (azul sobre azul), use o laranja.
- **Categorias** (afiliado, nicho, plataforma, área), nesta ordem: `--cat-1` a `--cat-8`. Nunca invente uma cor fora da lista.
- **Etapas de um funil** (até 7, em ordem): `--stage-1` a `--stage-7`.
- **Grade:** `--chart-grid`, só horizontal. A linha do zero é tracejada (5 4) em `--chart-zero`.
- **Eixos:** sem linha de eixo e sem tracinho. O texto é de 10 px em `--chart-axis-text`.
- **Hover:** guia vertical tracejada (3 3) e um tooltip.
- **Meta ou média de referência:** linha horizontal tracejada (5 4) em `--text-muted`, com o rótulo na legenda ("Meta de 2,0%"). A série mantém a cor dela. O que passou da meta ganha o vermelho no **valor** (rótulo da barra, selo, célula), nunca na série inteira.

### Calor em tabela (mapa de calor)

Uma coluna de números pode ganhar cor de fundo pela distância até o Total da tabela. Os tokens `--heat-*` de `tokens.css` definem as famílias e as proporções.

- **Uma família de cor por grupo de colunas.** Nunca misture matizes dentro de uma coluna. As famílias são azul→índigo, verde-água→verde, azul→roxo e amarelo→laranja.
- **Sete níveis**, de −3 a +3, contra o Total: desvio abaixo de 4% é o nível 0, abaixo de 10% é ±1, abaixo de 20% é ±2 e acima disso é ±3. Quando menor é melhor, inverta o sinal.
- No claro, o fundo é tingido e o texto é escuro da mesma família. Nunca texto branco sobre pastel.
- A célula é um chip: `inline-block`, raio 6, preenchimento de 2 × 4 px.
- Explique a régua numa legenda embaixo ou na descrição do card. Tom forte não pode ser lido como "valor alto" sem saber a régua.

### Régua fixa de meta

Quando a métrica tem meta definida pelo negócio, a cor segue uma régua **fixa**, que não muda com o período. A régua vem da regra de negócio. Confirme a meta vigente antes de aplicar as faixas.

### Regras de gráfico

- Todo gráfico tem tooltip. A única exceção é a minilinha do card de indicador, que é estática: sem hover, sem ponto e sem clique.
- **Eixo em números redondos, em quatro partes iguais** ($600 mil, $1,2 mi, $1,8 mi; 8%, 16%, 24%), com até uma casa e sem ",0" sobrando. Com dois eixos, os dois caem nas mesmas linhas de grade. O número e a unidade do eixo nunca quebram linha.
- **Rótulo de dado compacto**, com uma casa no milhão: `1,5M` e `850k`. O valor completo fica no tooltip.
- Valor em cima da barra só onde cabe, medido pela largura real: em toda barra, numa a cada duas ou em nenhuma.
- Num gráfico de barra e linha, a linha nunca passa sobre os rótulos das barras. Dê folga ao eixo das barras para a linha correr acima delas.
- Num gráfico de uma linha só, o valor vai em cada ponto que cabe. Se não couber em todos, vai em passo igual (um a cada N pontos), sempre com o primeiro e o último. Com várias séries, o valor fica só no tooltip. Nada de texto solto no meio do gráfico.
- **Taxa sem base fica vazia, não zero.** Num dia sem venda, a linha do ticket médio para no buraco e volta no próximo ponto. Zero desenharia uma queda que não houve.
- Período sem dado mostra "Sem dado no período" no lugar do gráfico, com a mesma altura.
- Série no tempo agrupa por dia até 3 meses, por semana até 1 ano e por mês acima disso. Métrica que só existe numa granularidade (turnover mensal, por exemplo) fica na dela.
- No eixo de datas, mostre a data de cada ponto que cabe. Se não couber, pule em passo igual. Deixe 16 px de folga nas pontas do eixo.
- Animação curta só quando o dado muda (500 ms). Passar o mouse não reanima nada. Quem pede menos movimento no sistema (`prefers-reduced-motion`) não vê animação.

## Tipografia

Fonte: **Geist** (`--font-ui`). Rótulos em caixa alta: **Geist Mono** (`--font-mono`). Use os tokens `--fs-*` da escala de dashboard, apresentação, página ou relatório. Não escreva tamanho fora da escala.

Pesos: 400 normal, 500 título, 600 título de card e número, 700 cabeçalho e destaque, 800 linha de total.

## Espaço, raio e sombra

Os valores foram medidos nas telas e não seguem uma grade de 4/8 px. Use os tokens `--space-*`, `--radius-*` e `--shadow-*` de `tokens.css`. Foco por teclado: `--focus-ring`. Movimento curto e sem quique (`--ease`, 120 a 180 ms). O estado ativo aparece pela cor, não por movimento.

## Componentes

A anatomia de cada peça, com os valores reais do sistema.

**Eyebrow.** Rótulo acima do título. Geist Mono, 10 px, peso 500, caixa alta, espaçamento 0,1em, com "• " na frente. A cor é `--text-muted` sobre card e `--hero-text-muted` sobre o topo. Exemplo: `• PAINEL DIÁRIO · HORÁRIO DE BRASÍLIA`.

**Barra do topo.** Uma pílula de vidro sobre o hero. À esquerda vão a logo, o divisor, o nome da peça e, se houver, um selo pequeno de vidro ("Dados Fictícios", "Em Teste"). À direita vão as abas, a moeda e o botão de tema (sol/lua, 42 px, mesma variante de vidro).

**Título do topo.** Eyebrow, depois um título de 36 px peso 500 com −0,03em, depois uma linha de descrição de 13,5 px em `--hero-text-muted`. À direita (ou embaixo, em tela estreita) ficam os filtros em pílulas de vidro.

**Card de indicador (big number).**

- Card com raio 18, borda de 1 px `--border-card`, fundo `--surface-card` e sombra `--shadow-kpi`. Preenchimento de 14 px em cima e 16 px nos lados.
- Linha 1: chip de ícone de 26 × 26 px, raio 10, com fundo na cor do indicador a 13% sobre o card e ícone de 14 px na cor do indicador. Ao lado, o rótulo em 12 px peso 500, `--text-secondary`.
- Linha 2: o valor em 21 px peso 600, −0,025em. Ele fica verde (`--positive-strong`) ou vermelho (`--negative`) só quando o sinal importa, como na margem.
- Linha 3: o selo de variação e uma nota curta de 11 px, `--text-secondary`.
- No fundo do card, encostada na borda de baixo: uma minilinha da série na cor do indicador. Traço de 1,75 px, área em degradê de 22% a 4% de opacidade, estática.
- Fileira: até 7 numa linha em tela larga, e 4 + 3 abaixo disso. Nunca deixe um card sozinho na última linha (estique o primeiro para fechar a conta).

**Selo de variação.** Texto de 10,5 px peso 600, preenchimento de 2 × 6 px e raio 8. **A cor diz se foi bom ou ruim, a seta diz se subiu ou desceu:** quando menor é melhor (custo, desligamento, turnover, absenteísmo), ▲ fica vermelho e ▼ fica verde. Positivo: fundo `--positive-bg` com texto `--positive-strong`. Negativo: `--negative-bg` com `--negative-strong`. Neutro: `--surface-card-3` com `--text-secondary`. Texto: `▲ 10,0%`, `▼ 3,4%`, `▲ 0,6 pp` (variação de percentual é em pontos percentuais), e `—` ou `Sem Comparativo` quando não há base. Variação que arredonda para zero fica neutra, sem seta.

**Card de seção.**

- Raio 22, borda de 1 px, fundo `--surface-card`, sombra `--shadow-card` e preenchimento de 20 × 22 px. O espaço entre cards é de 16 px.
- Cabeçalho: chip de ícone de 32 × 32 px, raio 10, fundo `--surface-active` e ícone de 16 px em `--brand-blue`. Num card de alerta, o chip usa fundo `--negative-bg` e ícone `--negative-strong`.
- Ao lado do chip: título de 16 px peso 600 e, embaixo, a descrição em 12,5 px `--text-secondary`.
- Controles (seletor, busca, download) à direita, na linha do título. Depois vêm 12 px de espaço e o conteúdo.

**Bloco de resumo.** Um número dentro de card, só para leitura. Raio 14, borda de 1 px, fundo `--surface-card-2` e preenchimento de 9 × 12 px. Em cima, o rótulo de 11,5 px com uma bolinha de 8 px na cor da série. No meio, o valor de 18 px peso 600. Embaixo, um contexto de 11 px em `--accent-ink`.

**Tabela.**

- Texto de 11,5 px em tabela densa (muitas colunas) e 12 px (`--fs-table`) em tabela com poucas colunas. Opcionalmente, uma linha de grupo de colunas em Geist Mono 9,5 px, caixa alta, `--brand-blue`, centralizada, com borda à esquerda entre os grupos.
- Cabeçalho de coluna: 10,5 px peso 700, `--text-secondary`, quebra em até duas linhas e tem borda de baixo `--border-strong`. Números alinhados à direita.
- Células com 3 px em cima e embaixo, 2 px nos lados, e separador `--border-row`. A coluna do nome corta com reticências e mostra o nome inteiro ao passar o mouse.
- **Linha de total** fixa no rodapé, com fundo `--surface-card-3` e peso 800. Cabeçalho fixo no topo. Altura máxima de cerca de 15 linhas, e o resto rola por dentro, com as últimas linhas visíveis esmaecendo sobre o total.
- Linha selecionada: `--surface-active`. Toda coluna ordena ao clicar no cabeçalho, com uma seta indicando o sentido.
- Tabela com só duas ou três colunas de número: a primeira ganha uma barra proporcional atrás do valor, que começa na borda esquerda da célula, com o número alinhado à esquerda por cima (degradê de `--brand-blue` de 32% a 8%, raio 8, proporcional ao maior valor da tabela).
- A tabela cabe na largura. Se não couber, rola para o lado dentro do card, com a coluna do nome parada. A página nunca rola para o lado.

**Seletor (abas, moeda, visão).** Opções coladas dentro de um trilho, com a escolhida marcada por uma pílula que desliza até ela (250 ms).

- Dentro de card: trilho `--surface-card-3` (escuro: `--surface-card-2`), raio 10 e 3 px de folga. A opção marcada é branca com contorno fino (escuro: `--surface-control`), raio 8, texto de 12,5 px peso 500.
- Sobre o topo: trilho de vidro em pílula, com a opção marcada em pílula branca (claro) ou translúcida (escuro).
- Opção que troca o que a seção mostra leva ícone. A moeda vai só com texto.

**Filtro em pílula.** Uma pílula de vidro sobre o topo, com ícone, nome do filtro e valor ("Todos" quando não há recorte), e 12 px entre o ícone e o texto. O painel do filtro tem busca, uma caixa de seleção por opção e "Só Este" ao passar o mouse.

**Chip de filtro aplicado.** Pílula em `--surface-card` com texto de 12 px peso 500 em `--brand-blue` e um × num círculo `--surface-active`. Os chips ficam numa linha "Filtrado por", com um "Limpar" no fim.

**Tooltip.** Raio 14, borda `--border-strong`, fundo `--surface-card` (nunca preto chapado) e sombra `--shadow-tooltip`. Largura mínima de 280 px.

- Cabeçalho: título de 13 px peso 600, com subtítulo à direita, e linha de baixo `--border-row`.
- Linhas: bolinha de 8 px na cor da série, nome em `--text-secondary` e valor peso 600 à direita. Os grupos são separados por uma linha tracejada.

**Etiqueta.** Texto de 10,5 px peso 500, raio 8 e preenchimento de 2 × 7 px. Neutra: fundo de 5% de preto (no escuro, 5% de branco) com `--text-secondary`. Alerta: vermelho claro sobre vermelho translúcido.

**Aviso.** Pílula âmbar (`--warning-bg`, borda `--warning-border`, texto `--warning-ink`) com ícone de relógio, para "dado parcial", "atualiza às 6h" etc.

**Botão.** Altura de 36 px, raio 10, texto peso 500 com −0,01em e ícone de 15 px.

- Primário: `--brand-blue` com texto branco (hover um tom mais claro).
- Secundário: borda `--border-card` e fundo branco (escuro: `--surface-card-2`).
- Fantasma: sem fundo, em `--text-secondary`.
- Vidro: só sobre o topo.
- Perigo: `--negative-strong`.

**Legenda.** Quadradinho de 9 px com raio 3 e o nome em 12 px `--text-secondary`, em linha, com 18 px entre os itens.

**Download.** Num dashboard interativo, todo gráfico e toda tabela têm um ícone de download no canto do card, que baixa em CSV o que está na tela. Card de indicador não leva. Apresentação e relatório não levam.

## Ícones

**Lucide** (lucide.dev), traço 2, em 16 px no chip de seção, 14 px no chip de indicador e 15 px em botão. Sem emoji. Unicode só em seta de variação (▲ ▼), ordenação (↓ ↑) e separador (· —).

Mapa de uso: `dollar-sign` receita · `shopping-cart` pedidos · `receipt` ticket · `hand-coins` CPA · `undo-2` reembolso · `package` produto · `piggy-bank` margem · `chart-column` por dia · `trophy` ranking · `tags` categoria · `globe` país · `funnel` funil · `triangle-alert` alerta · `flask-conical` teste · `users` pessoas · `calendar` período · `search` busca · `filter` filtro · `download` baixar · `moon`/`sun` tema.

## Texto e números

- **Português do Brasil**, informal mas preciso. Frases curtas, sem exclamação e sem emoji. Exemplo: "Clique na linha para filtrar o painel."
- **Title Case em texto curto que nomeia algo:** título, rótulo, coluna, legenda, botão, chip, opção e placeholder ("Receita por Dia", "Só Este", "Incluir Hoje"). Os conectivos ficam minúsculos (de, da, do, em, por, com, sem, vs). Frase continua frase: descrição de card, tooltip explicativo e a linha embaixo do título.
- **Dinheiro sempre com símbolo** ($ ou R$), nunca convertido sem avisar. Em tabela densa, o valor vai sem centavos. Em número grande de card, com centavos.
- **Percentual diz a base:** "10,0% da Receita". A variação de um percentual é em pontos percentuais (12,0% → 12,6% é ▲ 0,6 pp). Dinheiro e contagem variam em % relativo.
- **Compacto** com uma casa: `$1,5M`, `850k`, `$23,6M`. No eixo, "$600 mil" e "$1,2 mi".
- **Data** em dd/mm/aaaa ou "23 set – 6 out". O ano só aparece quando o intervalo atravessa a virada.
- Sem dado, use "—". Nunca escreva 0 onde não houve base.
- Num dashboard, embaixo do título vai a descrição do que o dash é, não o estado do filtro. O período aparece no filtro.
- Uma seção de alerta abre com uma frase de resumo (quanto, quantos, a concentração) e depois mostra os maiores casos. Sem alerta, a seção diz isso em vez de ficar vazia.

## Por tipo de peça

A base é a mesma em tudo: cor, tipo, ícone e voz. O que muda é a densidade e o peso da marca.

**Dashboard e painel.** Densidade alta, tema escuro e a cor carregando a informação.

- Topo: barra em pílula, título com filtros em vidro, a linha "Filtrado por" e uma fileira de cards de indicador.
- Corpo: cards de seção empilhados com 16 px de espaço. Dois cards lado a lado só quando cada um tem pelo menos 560 px.
- Largura máxima de 1560 px. A página nunca rola para o lado.
- Do celular (320 px) ao monitor, cada bloco se arruma no espaço que tem, sem texto sobreposto e sem item sozinho numa linha.

**Apresentação (slides).** Slide de 1280 × 720 px com 64 px de margem.

- **Capa:** o degradê ocupando o slide inteiro, nos dois temas azul-marinho por trás do texto (escuro: `--hero-gradient`; claro: `--hero-gradient-dash`, que não desbota, porque o `--hero-gradient` do claro clareia até o cinza e o texto branco some). Barra do topo em cima, e embaixo o eyebrow em mono, o título de 48 px peso 500 com −0,03em e uma linha de descrição. Logo branca ou colorida.
- **Slides internos:** fundo da página (claro para PDF e impressão, escuro para projetar). No alto, eyebrow e título do slide de 32 px. Embaixo, o conteúdo em cards de seção, que mantêm o título deles de 16 px (o do slide diz o assunto; o do card diz o que o gráfico mostra). Rodapé discreto com a logo pequena, o nome da peça e o número do slide. No máximo uma ideia por slide.
- **Ordem de uma "apresentação em formato de dashboard":** capa, depois um slide de indicadores (a fileira de cards de indicador, até 5 por linha, nos tamanhos do dashboard, com o gráfico principal embaixo), depois um slide por assunto, cada um com o seu card de seção.
- O botão de tema fica na barra do topo da capa. Gráficos seguem as regras desta skill.

**Relatório e documento.** Fundo claro, coluna de texto de até 760 px, corpo de 15 px com entrelinha 1,6 e títulos de 32 px. Tabelas e gráficos seguem as mesmas regras do dashboard.

**Sistema (telas de aplicação).** Densidade média, com os mesmos cards e controles. Tema escuro padrão, com o claro completo. O topo pode virar uma faixa mais baixa ou só a barra de navegação.

**Página e peça de marca.** É onde a marca aparece com mais força.

- O degradê azul-marinho com brilho ciano é a assinatura. O botão principal vai em `--brand-blue`.
- Mais respiro, título de 48 a 64 px (sempre peso 500 e −0,03em) e eyebrow em mono com "•".
- A cor continua tendo função: sem arco-íris decorativo, sem roxo como marca e sem emoji.

## Antes de entregar

- [ ] Tema escuro como padrão e claro completo (ou só claro, em relatório e impressão)
- [ ] Contraste: texto ≥ 4,5:1, gráfico e controle ≥ 3:1
- [ ] Só cores dos tokens, e cada métrica com a mesma cor em todo lugar
- [ ] Geist em tudo, mono só nos rótulos em caixa alta, números tabulares
- [ ] Logo certa para o fundo (branca no escuro, azul-marinho no claro)
- [ ] Dinheiro com símbolo, número no padrão brasileiro, variação com seta
- [ ] Title Case em rótulo e título, frase em descrição
- [ ] Eixo em números redondos, rótulo compacto e tooltip em todo gráfico
- [ ] Nenhuma rolagem lateral na página, nenhum texto cortado ou sobreposto
- [ ] Sem emoji e sem exclamação
