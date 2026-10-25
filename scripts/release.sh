#!/usr/bin/env bash
set -euo pipefail

version=${1:-}
if [[ ! $version =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  printf 'Uso: %s vX.Y.Z [--publish]\n' "$0" >&2
  exit 2
fi
if [[ ${2:-} != '' && ${2:-} != '--publish' ]]; then
  printf 'Opção desconhecida: %s\n' "$2" >&2
  exit 2
fi

root=$(git rev-parse --show-toplevel)
notas="$root/docs/release-notes/$version.md"
if [[ ! -f $notas ]]; then
  printf 'Falta o arquivo de notas da release: %s\n' "$notas" >&2
  exit 1
fi
mkdir -p "$root/dist"
limite_mib=16
limite=$((limite_mib * 1024 * 1024))
for target in darwin/arm64 darwin/amd64 linux/arm64 linux/amd64 windows/arm64 windows/amd64; do
  os=${target%/*}
  arch=${target#*/}
  extension=
  if [[ $os == windows ]]; then extension=.exe; fi
  binario="$root/dist/gex-dev-kit_${version}_${os}_${arch}${extension}"
  printf 'Compilando %s/%s\n' "$os" "$arch"
  (cd "$root" && CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build -trimpath -ldflags "-s -w -X main.version=$version" -o "$binario" ./cmd/gex-dev-kit)
  tamanho=$(wc -c < "$binario" | tr -d ' ')
  printf '  %s: %s bytes (limite %s)\n' "$(basename -- "$binario")" "$tamanho" "$limite"
  if (( tamanho > limite )); then
    printf 'Binário acima de %s MiB (%s bytes): %s\n' "$limite_mib" "$limite" "$binario" >&2
    exit 1
  fi
done
(
  cd "$root/dist"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "gex-dev-kit_${version}_"* > SHA256SUMS
  else
    shasum -a 256 "gex-dev-kit_${version}_"* > SHA256SUMS
  fi
)

licencas="$root/dist/THIRD_PARTY_LICENSES.txt"
principal=$(cd "$root" && go list -m)
falta=0
printf 'Avisos de licença das dependências incluídas nos binários.\n' > "$licencas"
while IFS=' ' read -r modulo diretorio; do
  [[ -n $modulo && $modulo != "$principal" ]] || continue
  arquivos=$(find "$diretorio" -maxdepth 1 -type f \( -iname 'LICENSE*' -o -iname 'COPYING*' -o -iname 'NOTICE*' \) | sort)
  if [[ -z $arquivos ]]; then
    printf 'Módulo sem arquivo de licença: %s\n' "$modulo" >&2
    falta=1
    continue
  fi
  {
    printf '\n===== %s =====\n' "$modulo"
    while IFS= read -r arquivo; do
      printf -- '--- %s\n' "$(basename -- "$arquivo")"
      cat -- "$arquivo"
      printf '\n'
    done <<< "$arquivos"
  } >> "$licencas"
done < <(cd "$root" && go list -deps -f '{{with .Module}}{{.Path}} {{.Dir}}{{end}}' ./cmd/gex-dev-kit | sort -u)
if [[ $falta -ne 0 ]]; then
  printf 'Falta arquivo de licença; revise dist/THIRD_PARTY_LICENSES.txt antes de publicar.\n' >&2
  exit 1
fi

if [[ ${2:-} != '--publish' ]]; then
  printf 'Binários, SHA256SUMS e THIRD_PARTY_LICENSES.txt em %s/dist. Publique com --publish após conferir.\n' "$root"
  exit 0
fi
if [[ ! -t 0 ]]; then
  printf 'Publicação exige confirmação no terminal.\n' >&2
  exit 1
fi
if command -v gum >/dev/null 2>&1; then
  gum confirm "Publicar $version em Gex-Development/gex-dev-kit?" || exit 1
else
  read -r -p "Publicar $version em Gex-Development/gex-dev-kit? Digite sim: " answer
  [[ $answer == sim ]] || exit 1
fi
gh release create "$version" "$root/dist/gex-dev-kit_${version}_"* "$root/dist/SHA256SUMS" "$root/dist/THIRD_PARTY_LICENSES.txt" --repo Gex-Development/gex-dev-kit --title "$version" --notes-file "$notas"
