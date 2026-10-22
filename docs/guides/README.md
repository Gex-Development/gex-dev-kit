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

   O script cria dois HOME descartáveis dentro de `CAMINHO_DE_TRABALHO`. Ele não usa o HOME real. Antes de gravar cada PNG, ele procura nomes de usuário, caminhos do sistema e o nome do diretório de trabalho na tela. Se achar algo, ele para com erro.

3. Confira a composição e renderize o vídeo:

   ```bash
   cd docs/guides/composition
   npx --yes hyperframes@0.8.181 check
   npx --yes hyperframes@0.8.181 render --quality looks --output ../guia.mp4
   ```

## Tempo aproximado

- Captura das 17 telas: cerca de 3 a 4 minutos.
- Verificação da composição: alguns minutos na primeira execução, porque o `npx` baixa o pacote.
- Render do vídeo: depende da máquina. Use `--quality draft` para testar mais rápido.

## Limites conhecidos

- A tela inicial e a tela final mostram o caminho do HOME como `~`. O script faz essa troca porque a CLI imprime o caminho absoluto nessas telas. Veja o relatório da versão para o detalhe.
- O comando equivalente da tela final aparece como `--home ~`, porque a CLI o imprime assim.
