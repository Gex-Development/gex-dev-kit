---
name: gex-hub-design-system
description: Use para construir UI no gex-hub-dashboards com tokens do repo, shadcn/ui e Storybook.
---

# Trabalhar no Design System do Hub

Use esta skill ao criar ou estilizar componentes, stories e páginas dentro do gex-hub-dashboards: tokens CSS locais, Tailwind, shadcn/ui, Storybook e data-viz. Para visual GEX fora desse repo, use `gex-design-system`.

## Fonte de verdade

Leia `src/styles/tokens/*.css`, `src/styles/tailwind.css` e os componentes atuais antes de alterar qualquer coisa. Os tokens locais podem divergir do export estático `gex-design-system`; o export geral GEX não substitui os tokens locais do Hub. O tema escuro usa a classe `.dark`; preserve a acessibilidade dos componentes.

## Tokens

- Valores claros ficam em `:root` e escuros em `.dark`. O escuro é o padrão, aplicado pelo next-themes com a classe `.dark`. Não use `data-theme`.
- Utilities: `bg-surface-page|card|card-2|card-3|active|control`, `text-text-primary|secondary|muted`, `border-border-card|row|strong`, `text-positive*` e `text-negative*`, `bg-hero-glass`, `text-hero-text(-muted)`.
- Raio: `rounded-chip` (6), `tile` (8), `control` (10), `tile-lg` (12), `tooltip` (14), `kpi` (18), `card` (22) e `pill`.
- Sombra: `shadow-card`, `shadow-kpi`, `shadow-tooltip`.
- Para tokens sem utility, use a forma arbitrária, por exemplo `h-(--control-height)`, `bg-(--btn-primary-bg)`, `duration-(--dur-fast) ease-(--ease)` e `focus-visible:shadow-(--focus-ring)`.
- A escala tipográfica são variáveis CSS (`--fs-*`). Escreva tamanhos como `text-[12.5px]`. Não crie chaves `--text-*` no tema: o `cn` mistura tamanhos de texto com cores de texto.
- As variáveis do shadcn mapeiam para os tokens GEX: `--primary` é `--brand-blue` (#1E6FD9).

## Componentes

- Mantenha o shadcn em `src/components/core` e estenda com variantes GEX em vez de fazer fork.
  - `Button`: variantes `default`, `secondary`, `outline`, `ghost`, `destructive`, `glass` e `link`. Tamanhos: sm 28, default 36 e lg 44.
  - `Badge`: variantes `brand`, `danger`, `warning`, `neutral` e `diag`. Tamanhos `md` e `mono`.
- Primitivas de marca ficam em `src/components/brand`: `Eyebrow` (tom muted, brand ou onDark), `HeroBand` (altura full, short ou nav) e `Logo` (tom color, ink, white ou symbol).
- `glass` e `Logo tone="color"` funcionam só dentro de `HeroBand`. Em superfícies claras, use `Logo tone="ink"`.
- Stories de componentes hero ou glass precisam de um decorator `HeroBand`. Tipe-o como `Decorator` de `@storybook/react-vite`.

## Workflow shadcn + Storybook

Ao importar componentes shadcn/ui ou itens de registry para este repo, mantenha as implementações upstream intactas até haver pedido de customização. O Storybook é o inventário visual: crie `*.stories.tsx` ao lado de cada componente, com a versão, os decorators, os aliases e o tema Tailwind atuais. Preserve API e acessibilidade; uma story que só comprova o import não basta.

1. Inspecione `components.json`, os scripts do pacote, `.storybook/main` e `preview`, `src/components/core` e a versão da CLI shadcn. O repo usa pnpm, alias `@/` e primitivas em pastas próprias sob `src/components/core/`.
2. Busque itens exatos nos registries configurados. Antes de instalar, rode `pnpm dlx shadcn@<versão fixada no projeto> add <item> --dry-run --yes` e inspecione sobrescritas, dependências, alterações em variáveis CSS e caminhos gerados. Compare itens de terceiros com as versões locais. Preserve utilities e tokens existentes.
3. Instale os itens pedidos pela CLI shadcn. Adicione dependências com pnpm e as convenções do workspace, sem aceitar sobrescritas amplas às cegas.
4. Para cada componente, crie uma story colocalizada com variantes, tamanhos, estados e composições representativos, usando props válidas. Disponibilize Controls para interações quando fizer sentido.
5. Composições com dados ou providers exigem fixtures determinísticas e wrappers reais mínimos; não omita componentes difíceis de configurar. Reutilize o decorator `HeroBand` para hero e glass.
6. Rode o build do Storybook, abra-o e inspecione cada categoria no navegador; corrija erros de runtime, providers, estilos, imports e API das stories.
7. Preserve os componentes upstream sem customização até a revisão do inventário-base pelo responsável. Relate itens importados, indisponíveis, URL do Storybook e resultados observados. Não incorpore erros temporários do projeto como regra universal.

## Data viz

- Sem valores hex em componentes. Importe de `src/lib/data-viz.ts`:
  - `KPI_COLORS` e `KPI_NEGATIVE_COLOR`
  - `TOOLTIP_COLORS` e `STAGE_COLORS`
  - `heatStyle(family: aov|funnel|stage|pots, level -3..3)`
  - `roasStyle(value)`, para a régua fixa de ROAS
  - `FREIGHT_OK` e `FREIGHT_BAD`
- Cores de gráfico usam `var(--chart-bar|chart-bar-top|chart-mc-line|chart-grid|chart-zero|chart-label-text)`.
- Para achar hex sobrando, rode `grep -rnoE '#[0-9A-Fa-f]{6}' src/components src/prototypes` e ignore as pastas `/core/` e `/effects/`.

## Regras de interação

- Hover muda cor, não posição.
- Linhas de tabela não têm hover; só a linha selecionada é destacada (`bg-surface-active`).
- Estado desabilitado é `opacity-45`. Transições de 120 a 180 ms, sem quique.

## Verificação

Rode `pnpm exec tsc --noEmit`, `pnpm build`, `pnpm exec vitest run src/` e `pnpm storybook:build`. Depois confira o app em :5173 e o Storybook em :6006 nos dois temas; para o tema claro, use o parâmetro de URL `globals=theme:light`. Sem execução real, não declare a variante revisada.

## Referência de tema

No estado atual dos tokens, `--warning` é #fcd34d nos dois temas, mas `--warning-ink` é #92400e no claro e #fcd34d no escuro. Para texto sobre aviso claro, use `--warning-ink`, não `--warning` direto. `src/styles/tokens/surfaces.css` já define escalas para landing, relatório e deck; confirme a implementação dos templates antes de tratá-los como entregues.
