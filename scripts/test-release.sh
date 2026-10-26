#!/usr/bin/env bash
set -euo pipefail

source_root=$(cd -- "$(dirname -- "$0")/.." && pwd)
work=$(mktemp -d)
trap 'rm -rf -- "$work"' EXIT
repo="$work/repo"
mkdir -p "$repo/scripts" "$repo/docs/release-notes" "$repo/cmd/gex-dev-kit" "$work/bin" "$work/home"
repo=$(cd "$repo" && pwd -P)
cp "$source_root/scripts/release.sh" "$source_root/scripts/bump.sh" "$repo/scripts/"
cp "$source_root/README.md" "$repo/README.md"
printf '## [Não publicada]\n\n### Adicionado\n\n- Mudança humana.\n\n## [9.9.8] - 2026-01-01\n\n[9.9.8]: https://github.com/Gex-Development/gex-dev-kit/compare/v9.9.7...v9.9.8\n' > "$repo/CHANGELOG.md"
printf 'package main\n\nvar version = "v9.9.8"\n' > "$repo/cmd/gex-dev-kit/main.go"
printf 'Notas humanas.\n' > "$repo/docs/release-notes/v9.9.9.md"
export HOME="$work/home" PATH="$work/bin:/usr/bin:/bin" FAKE_GH_LOG="$work/gh-args" FAKE_LICENSE_DIR="$work"
(cd "$repo" && git init -q)
printf '%s\n' '#!/usr/bin/env bash' 'set -euo pipefail' 'case $1 in' '  build)' '    while [[ $1 != -o ]]; do shift; done' '    file=$2' '    size=8388608' '    if [[ $GOOS == windows && $GOARCH == amd64 ]]; then size=15099495; fi' '    dd if=/dev/zero of="$file" bs=1 count=0 seek="$size" 2>/dev/null ;;' '  list)' '    if [[ $2 == -m ]]; then printf "example.com/kit\n"; else printf "example.com/kit %s\n" "$FAKE_LICENSE_DIR"; fi ;;' '  *) exit 3 ;;' 'esac' > "$work/bin/go"
printf '%s\n' '#!/usr/bin/env bash' 'set -euo pipefail' 'printf "%s\n" "$@" > "$FAKE_GH_LOG"' > "$work/bin/gh"
chmod +x "$work/bin/go" "$work/bin/gh"
printf 'Licença de teste.\n' > "$work/LICENSE"
fail() { printf 'Falhou: %s\n%s\n' "$1" "${output:-}" >&2; exit 1; }
run_release() { status=0; output=$(cd "$repo" && /bin/bash scripts/release.sh "$@" 2>&1) || status=$?; }
run_bump() { status=0; output=$(cd "$repo/docs" && GEX_BUMP_DATE=2026-02-03 /bin/bash ../scripts/bump.sh "$@" 2>&1) || status=$?; }
reject() { [[ $status -eq $1 && $output == *"$2"* ]] || fail "$3 (status $status)"; }

run_bump invalid
reject 2 'Uso:' 'versão inválida'
[[ ! -d $repo/dist ]] || fail 'versão inválida gerou dist'
run_release v9.9.9 --publish
reject 1 'Publicação sem terminal exige --confirm vX.Y.Z' 'gate sem TTY'
[[ ! -d $repo/dist && ! -e $FAKE_GH_LOG ]] || fail 'gate compilou ou publicou'
run_release v9.9.9 --publish --confirm v9.9.8
reject 2 'Confirmação divergente' 'confirmação divergente'
[[ ! -d $repo/dist ]] || fail 'confirmação divergente compilou'
run_release v9.9.9 --other
reject 2 'Opção desconhecida' 'opção desconhecida'
[[ ! -d $repo/dist ]] || fail 'opção desconhecida compilou'
mv "$repo/docs/release-notes/v9.9.9.md" "$work/notes"
run_release v9.9.9 --confirm v9.9.9 --publish
reject 1 'Falta o arquivo de notas' 'notas ausentes'
[[ ! -d $repo/dist ]] || fail 'notas ausentes compilaram'
mv "$work/notes" "$repo/docs/release-notes/v9.9.9.md"
run_release v9.9.9 --confirm v9.9.9 --publish
[[ $status -eq 0 ]] || fail 'publicação confirmada'
[[ $output == *'50% do limite de 16777216'* && $output == *'90% do limite de 16777216'* && $output == *'aviso: gex-dev-kit_v9.9.9_windows_amd64.exe usa 90% do limite; veja a ADR-0037'* ]] || fail 'percentual e aviso'
[[ $output != *'usa 50% do limite'* ]] || fail 'aviso indevido a 50%'
[[ $(find "$repo/dist" -type f | wc -l | tr -d ' ') -eq 8 ]] || fail 'oito anexos'
[[ $(wc -l < "$FAKE_GH_LOG" | tr -d ' ') -eq 17 ]] || fail 'gh não recebeu oito anexos e opções'
[[ $(cat "$FAKE_GH_LOG") == *$'create\nv9.9.9\n'* && $(cat "$FAKE_GH_LOG") == *'THIRD_PARTY_LICENSES.txt'* ]] || fail 'gh release create'
for file in "$repo/dist/gex-dev-kit_v9.9.9_"* "$repo/dist/SHA256SUMS" "$repo/dist/THIRD_PARTY_LICENSES.txt"; do
  grep -Fxq -- "$file" "$FAKE_GH_LOG" || fail "anexo ausente: $file"
done

cp "$repo/CHANGELOG.md" "$work/changelog-original"
cp "$repo/README.md" "$work/readme-original"
cp "$repo/cmd/gex-dev-kit/main.go" "$work/main-original"
run_bump v9.9.9
[[ $status -eq 0 && $output == *'scripts/check.sh'* ]] || fail 'bump válido'
[[ $(cat "$repo/cmd/gex-dev-kit/main.go") == *'var version = "v9.9.9"'* ]] || fail 'versão Go'
[[ $(cat "$repo/CHANGELOG.md") == *'## [9.9.9] - 2026-02-03'* && $(cat "$repo/CHANGELOG.md") == *'docs/release-notes/v9.9.9.md'* && $(cat "$repo/CHANGELOG.md") == *'[9.9.9]: https://github.com/Gex-Development/gex-dev-kit/compare/v9.9.8...v9.9.9'* ]] || fail 'promoção e comparação'
[[ $(cat "$repo/README.md") == *'![Versão v9.9.9][badge-version]'* && $(cat "$repo/README.md") == *'Vers%C3%A3o-v9.9.9-6f42c1'* ]] || fail 'badge'
run_bump v9.9.9
reject 1 'Falta seção ## [Não publicada]' 'segunda execução'
for missing in notes section version previous badge; do
  cp "$work/changelog-original" "$repo/CHANGELOG.md"
  cp "$work/main-original" "$repo/cmd/gex-dev-kit/main.go"
  cp "$work/readme-original" "$repo/README.md"
  case $missing in
    notes) mv "$repo/docs/release-notes/v9.9.9.md" "$work/notes" ;;
    section) printf 'Sem seção.\n' > "$repo/CHANGELOG.md" ;;
    version) printf '## [Não publicada]\n## [9.9.9]\n' > "$repo/CHANGELOG.md" ;;
    previous) printf '## [Não publicada]\n\n### Adicionado\n' > "$repo/CHANGELOG.md" ;;
    badge) printf 'Sem badge.\n' > "$repo/README.md" ;;
  esac
  cp "$repo/CHANGELOG.md" "$work/expected-changelog"
  cp "$repo/README.md" "$work/expected-readme"
  run_bump v9.9.9
  [[ $status -eq 1 ]] || fail "pré-condição $missing"
  cmp -s "$repo/CHANGELOG.md" "$work/expected-changelog" || fail "pré-condição $missing alterou changelog"
  cmp -s "$repo/README.md" "$work/expected-readme" || fail "pré-condição $missing alterou README"
  cmp -s "$repo/cmd/gex-dev-kit/main.go" "$work/main-original" || fail "pré-condição $missing alterou Go"
  if [[ $missing == notes ]]; then mv "$work/notes" "$repo/docs/release-notes/v9.9.9.md"; fi
done
for missing in CHANGELOG.md README.md cmd/gex-dev-kit/main.go; do
  cp "$work/changelog-original" "$repo/CHANGELOG.md"
  cp "$work/main-original" "$repo/cmd/gex-dev-kit/main.go"
  cp "$work/readme-original" "$repo/README.md"
  mv "$repo/$missing" "$work/missing-file"
  run_bump v9.9.9
  [[ $status -eq 1 && ! -e $repo/$missing ]] || fail "arquivo ausente $missing"
  mv "$work/missing-file" "$repo/$missing"
  cmp -s "$repo/CHANGELOG.md" "$work/changelog-original" || fail "arquivo ausente $missing alterou changelog"
  cmp -s "$repo/README.md" "$work/readme-original" || fail "arquivo ausente $missing alterou README"
  cmp -s "$repo/cmd/gex-dev-kit/main.go" "$work/main-original" || fail "arquivo ausente $missing alterou Go"
done
cp "$repo/scripts/release.sh" "$work/release-original"
cp "$repo/scripts/bump.sh" "$work/bump-original"
awk 'index($0, "$confirm != \"$version\"") { sub(/if \[\[.*; then/, "if false; then") } { print }' "$work/release-original" > "$repo/scripts/release.sh"
run_release v9.9.9 --publish --confirm v9.9.8
[[ $status -eq 0 ]] || fail 'mutação de confirmação não foi aplicada'
cp "$work/release-original" "$repo/scripts/release.sh"
run_release v9.9.9 --publish --confirm v9.9.8
reject 2 'Confirmação divergente' 'teste não detecta mutação de confirmação'

awk 'index($0, "percentual >= 85") { sub(/if \(\(.*; then/, "if false; then") } { print }' "$work/release-original" > "$repo/scripts/release.sh"
run_release v9.9.9
[[ $status -eq 0 && $output != *'aviso: gex-dev-kit_v9.9.9_windows_amd64.exe usa 90%'* ]] || fail 'mutação de aviso não foi aplicada'
cp "$work/release-original" "$repo/scripts/release.sh"
run_release v9.9.9
[[ $status -eq 0 && $output == *'aviso: gex-dev-kit_v9.9.9_windows_amd64.exe usa 90%'* ]] || fail 'teste não detecta mutação de aviso'

for mutation in promotion comparison; do
  cp "$work/changelog-original" "$repo/CHANGELOG.md"
  cp "$work/main-original" "$repo/cmd/gex-dev-kit/main.go"
  cp "$work/readme-original" "$repo/README.md"
  if [[ $mutation == promotion ]]; then
    awk 'index($0, "print \"## [\" n \"] - \" d") { print "  /^## \\[Não publicada\\]$/ { next }"; next } { print }' "$work/bump-original" > "$repo/scripts/bump.sh"
  else
    awk 'index($0, "print \"[\" n \"]: https://github.com/") { print "  /^\\[[0-9]+\\.[0-9]+\\.[0-9]+\\]:/ && !linked++ { next }"; next } { print }' "$work/bump-original" > "$repo/scripts/bump.sh"
  fi
  run_bump v9.9.9
  [[ $status -eq 0 ]] || fail "mutação $mutation não executou"
  if [[ $mutation == promotion ]]; then
    [[ $(cat "$repo/CHANGELOG.md") != *'## [9.9.9] - 2026-02-03'* ]] || fail 'mutação de promoção não foi aplicada'
  else
    [[ $(cat "$repo/CHANGELOG.md") != *'[9.9.9]: https://github.com/Gex-Development/gex-dev-kit/compare/v9.9.8...v9.9.9'* ]] || fail 'mutação de comparação não foi aplicada'
  fi
  cp "$work/bump-original" "$repo/scripts/bump.sh"
  cp "$work/changelog-original" "$repo/CHANGELOG.md"
  cp "$work/main-original" "$repo/cmd/gex-dev-kit/main.go"
  cp "$work/readme-original" "$repo/README.md"
  run_bump v9.9.9
  [[ $status -eq 0 && $(cat "$repo/CHANGELOG.md") == *'## [9.9.9] - 2026-02-03'* && $(cat "$repo/CHANGELOG.md") == *'[9.9.9]: https://github.com/Gex-Development/gex-dev-kit/compare/v9.9.8...v9.9.9'* ]] || fail "teste não detecta mutação $mutation"
done
printf 'test-release: ok\n'
