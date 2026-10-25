#!/usr/bin/env bash
set -euo pipefail

guia=$(cd "$(dirname "$0")/.." && pwd)
composicao="$guia/composition"
hyperframes=${HYPERFRAMES:-npx --yes hyperframes@0.8.181}

cd "$composicao"
$hyperframes check
$hyperframes render --quality looks --output "$guia/guia.mp4"
printf 'Vídeo gerado em %s\n' "$guia/guia.mp4"
