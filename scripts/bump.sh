#!/usr/bin/env bash
set -euo pipefail

version=${1:-}
if [[ $# -ne 1 || ! $version =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  printf 'Uso: %s vX.Y.Z\n' "$0" >&2
  exit 2
fi
root=$(git rev-parse --show-toplevel)
notes="$root/docs/release-notes/$version.md"
changelog="$root/CHANGELOG.md"
main="$root/cmd/gex-dev-kit/main.go"
readme="$root/README.md"
number=${version#v}
if [[ ! -f $notes ]]; then printf 'Falta o arquivo de notas da release: %s\n' "$notes" >&2; exit 1; fi
if [[ ! -f $changelog || ! -f $main || ! -f $readme ]]; then printf 'Faltam arquivos da release (CHANGELOG, main.go ou README).\n' >&2; exit 1; fi
if ! grep -q '^## \[Não publicada\]$' "$changelog"; then printf 'Falta seção ## [Não publicada] no CHANGELOG.\n' >&2; exit 1; fi
if grep -q "^## \[$number\]" "$changelog"; then printf 'Versão %s já consta no CHANGELOG.\n' "$number" >&2; exit 1; fi
previous=$(awk '/^## \[[0-9]+\.[0-9]+\.[0-9]+\]/ { v=$2; gsub(/\[|\]/,"",v); print v; exit }' "$changelog")
if [[ -z $previous ]] || ! grep -q "^\[$previous\]: https://github.com/Gex-Development/gex-dev-kit/" "$changelog"; then
  printf 'Falta versão anterior com link no CHANGELOG.\n' >&2
  exit 1
fi
if [[ $(grep -c '^var version = "v[0-9]*\.[0-9]*\.[0-9]*"$' "$main") -ne 1 ]] ||
   [[ $(grep -c '^!\[Versão v[0-9]*\.[0-9]*\.[0-9]*\]\[badge-version\]' "$readme") -ne 1 ]] ||
   [[ $(grep -c '^\[badge-version\]: https://img.shields.io/badge/Vers%C3%A3o-v[0-9]*\.[0-9]*\.[0-9]*-6f42c1?style=flat-square$' "$readme") -ne 1 ]]; then
  printf 'Formato esperado de versão ou badge não encontrado.\n' >&2
  exit 1
fi
stamp=${GEX_BUMP_DATE:-$(date +%F)}
tmp=$(mktemp -d)
trap 'rm -rf -- "$tmp"' EXIT
awk -v v="$version" '/^var version = "v[0-9]+\.[0-9]+\.[0-9]+"$/ { print "var version = \"" v "\""; next } { print }' "$main" > "$tmp/main.go"
awk -v v="$version" '/^!\[Versão v[0-9]+\.[0-9]+\.[0-9]+\]\[badge-version\]/ { sub(/v[0-9]+\.[0-9]+\.[0-9]+/,v) } /^\[badge-version\]:/ { sub(/v[0-9]+\.[0-9]+\.[0-9]+/,v) } { print }' "$readme" > "$tmp/README.md"
awk -v n="$number" -v v="$version" -v p="$previous" -v d="$stamp" '
  /^## \[Não publicada\]$/ { print "## [" n "] - " d; print ""; print "As notas da publicação estão em [`docs/release-notes/" v ".md`](docs/release-notes/" v ".md)."; next }
  /^\[[0-9]+\.[0-9]+\.[0-9]+\]: https:\/\/github.com\/Gex-Development\/gex-dev-kit\// && !linked++ { print "[" n "]: https://github.com/Gex-Development/gex-dev-kit/compare/v" p "..." v }
  { print }
' "$changelog" > "$tmp/CHANGELOG.md"
mv "$tmp/main.go" "$main"
mv "$tmp/README.md" "$readme"
mv "$tmp/CHANGELOG.md" "$changelog"
printf 'Versão %s: main.go, CHANGELOG.md, badge README.md atualizados. Rode scripts/check.sh.\n' "$version"
