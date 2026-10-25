# Guia em vídeo e capturas da CLI

Esta pasta guarda o roteiro em vídeo, as capturas reais da CLI e a página do guia.

## Conteúdo

- `composition/`: fonte do vídeo em Hyperframes. `package.json` fixa `hyperframes@0.8.181`.
- `guia.mp4`: vídeo final, de cerca de 105 segundos. Só fica versionado se tiver até 20 MB.
- `guia.html`: página com vídeo, capturas, comandos e tabela das telas. Usa `assets/tokens.css` e troca para o tema escuro pelo `prefers-color-scheme`.
- `assets/screens/`: 17 capturas em 100×30.
- `scripts/capture_cli.py`: roda a CLI numa pseudo-terminal e gera as capturas.

## Requisitos

- macOS, com a fonte Menlo em `/System/Library/Fonts/Menlo.ttc`.
- Python 3 com `pyte` e `Pillow`. Crie um ambiente isolado e instale os dois pacotes nele.
- Git no caminho `/usr/bin/git`.
- Binário da CLI gerado por `bash scripts/release.sh v0.3.0` na raiz do repositório. O script grava `dist/`, que o git ignora.
- Node.js 22 ou mais novo, e FFmpeg, para renderizar o vídeo.

## Como regenerar

1. Gere o binário:

   ```bash
   bash scripts/release.sh v0.3.0
   ```

2. Capture as telas. Use um diretório de trabalho novo, fora do HOME real:

   ```bash
   python3 docs/guides/scripts/capture_cli.py dist/gex-dev-kit_v0.3.0_darwin_arm64 --workdir CAMINHO_DE_TRABALHO
   ```

   O script cria dois HOME descartáveis dentro de `CAMINHO_DE_TRABALHO`. Ele não usa o HOME real. Antes de gravar cada PNG, ele procura nomes de usuário, caminhos do sistema e o nome do diretório de trabalho na tela. Se achar algo, ele para com erro e mostra a linha com os trechos trocados por `[REDIGIDO]`.

3. Confira a composição e renderize o vídeo:

   ```bash
   bash docs/guides/scripts/render.sh
   ```

   O script roda `check` e depois `render`, e grava `docs/guides/guia.mp4`. Por padrão usa `npx --yes hyperframes@0.8.181`. Para outro binário, defina `HYPERFRAMES`, por exemplo `HYPERFRAMES=/caminho/hyperframes bash docs/guides/scripts/render.sh`.

## Tempo aproximado

- Captura das 17 telas: cerca de 3 a 4 minutos.
- Verificação da composição: alguns minutos na primeira execução, porque o `npx` baixa o pacote.
- Render do vídeo: depende da máquina. Use `--quality draft` para testar mais rápido.

## Limites conhecidos

- A TUI mostra o HOME do usuário como `~` na tela inicial, nas linhas dos agentes e na tela final. O comando equivalente da tela final aparece como `--home ~` quando o destino é o HOME do usuário. O comando executado não muda.
- O script não faz mais máscara de substituição: qualquer caminho bruto na tela interrompe a captura.
