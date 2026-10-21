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
mkdir -p "$root/dist"
for target in darwin/arm64 darwin/amd64 linux/arm64 linux/amd64 windows/arm64 windows/amd64; do
  os=${target%/*}
  arch=${target#*/}
  extension=
  if [[ $os == windows ]]; then extension=.exe; fi
  printf 'Compilando %s/%s\n' "$os" "$arch"
  (cd "$root" && CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build -trimpath -ldflags "-s -w -X main.version=$version" -o "dist/gex-dev-kit_${version}_${os}_${arch}${extension}" ./cmd/gex-dev-kit)
done
(
  cd "$root/dist"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "gex-dev-kit_${version}_"* > SHA256SUMS
  else
    shasum -a 256 "gex-dev-kit_${version}_"* > SHA256SUMS
  fi
)

if [[ ${2:-} != '--publish' ]]; then
  printf 'Binários e SHA256SUMS em %s/dist. Publique com --publish após conferir.\n' "$root"
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
gh release create "$version" "$root/dist/gex-dev-kit_${version}_"* "$root/dist/SHA256SUMS" --repo Gex-Development/gex-dev-kit --title "$version" --notes "Kit GEX: Core global Claude Code/Codex, skills Matt Pocock opcionais, hooks auditados, doctor e rollback. Confira README e hashes antes de instalar."
