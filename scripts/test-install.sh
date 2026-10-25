#!/usr/bin/env bash
set -euo pipefail

root=$(cd -- "$(dirname -- "$0")/.." && pwd)
work=$(mktemp -d)
trap 'rm -rf -- "$work"' EXIT
mkdir -p "$work/bin" "$work/assets" "$work/home" "$work/tmp"
export HOME="$work/home" TMPDIR="$work/tmp" FAKE_ASSETS="$work/assets" PATH="$work/bin:/usr/bin:/bin"

fail() { printf 'Falhou: %s\n%s\n' "$1" "${output:-}" >&2; exit 1; }
run() { status=0; output=$(/bin/bash "$root/scripts/install.sh" "$@" 2>&1) || status=$?; }
run_pipe() { status=0; output=$(cat "$root/scripts/install.sh" | /bin/bash -s -- "$@" 2>&1) || status=$?; }
expect() { [[ $status -eq 0 && $output == *"$1"* ]] || fail "$2"; }
reject() { [[ $status -ne 0 && $output == *"$1"* ]] || fail "$2"; }

printf '%s\n' '#!/usr/bin/env bash' 'set -euo pipefail' 'if [[ ${FAKE_READ_STDIN:-} == 1 && $1 == auth ]]; then cat >/dev/null; fi' 'if [[ ${FAKE_AUTH:-ok} != ok && $1 == auth ]]; then exit 1; fi' 'case "$1 $2" in' '  "auth status") exit 0 ;;' '  "release view") printf "v0.3.0\n" ;;' '  "release download")' '    shift 2' '    while [[ $# -gt 0 ]]; do' '      case $1 in --pattern) pattern=$2; shift 2 ;; --dir) dir=$2; shift 2 ;; *) shift ;; esac' '    done' '    cp -- "$FAKE_ASSETS/$pattern" "$dir/" ;;' '  *) exit 1 ;;' 'esac' > "$work/bin/gh"
chmod +x "$work/bin/gh"
for target in darwin_arm64 darwin_amd64 linux_arm64 linux_amd64; do
  printf '%s\n' '#!/usr/bin/env bash' "printf 'v0.3.0\\n'" ": $target" > "$work/assets/gex-dev-kit_v0.3.0_$target"
  chmod +x "$work/assets/gex-dev-kit_v0.3.0_$target"
done
(
  cd "$work/assets"
  if command -v sha256sum >/dev/null 2>&1; then sha256sum gex-dev-kit_v0.3.0_* > SHA256SUMS; else shasum -a 256 gex-dev-kit_v0.3.0_* > SHA256SUMS; fi
)

FAKE_READ_STDIN=1 GEX_TEST_OS=darwin GEX_TEST_ARCH=arm64 run_pipe --dir "$work/piped"
expect 'Próximos passos' 'execução por pipe com gh lendo stdin'
[[ $output == *'Kit GEX · instalador'* && $output == *'gex-dev-kit          abre o assistente'* && $output == *'gex-dev-kit doctor   confere a instalação'* ]] || fail 'resumo da instalação'
cmp -s "$work/piped/gex-dev-kit" "$work/assets/gex-dev-kit_v0.3.0_darwin_arm64" || fail 'ativo por pipe'
printf 'ok: instalação via pipe mesmo com gh lendo stdin\n'

for target in darwin_arm64 darwin_amd64 linux_arm64 linux_amd64; do
  dir="$work/$target"
  GEX_TEST_OS=${target%_*} GEX_TEST_ARCH=${target#*_} run --dir "$dir"
  expect 'Próximos passos' "instalação $target"
  [[ -x "$dir/gex-dev-kit" && $("$dir/gex-dev-kit" --version) == v0.3.0 ]] || fail "execução $target"
  cmp -s "$dir/gex-dev-kit" "$work/assets/gex-dev-kit_v0.3.0_$target" || fail "ativo $target"
done
[[ -z $(find "$work/tmp" -mindepth 1 -print) ]] || fail 'temporários não apagados'
printf 'ok: instalação e seleção de quatro ativos; temporários apagados\n'

GEX_TEST_OS=linux GEX_TEST_ARCH=amd64 run --version v0.3.0 --dir "$work/explicit"
expect 'v0.3.0' 'versão explícita'
GEX_TEST_OS=linux GEX_TEST_ARCH=amd64 run --dir "$work/explicit"
expect 'v0.3.0 → v0.3.0' 'versão anterior'
GEX_INSTALL_DIR="$work/from-env" GEX_TEST_OS=linux GEX_TEST_ARCH=amd64 run
[[ $status -eq 0 && -x $work/from-env/gex-dev-kit ]] || fail 'GEX_INSTALL_DIR'
NO_COLOR=1 GEX_TEST_OS=linux GEX_TEST_ARCH=amd64 run
expect '~/.local/bin/gex-dev-kit' 'HOME abreviado no resumo'
[[ -x $HOME/.local/bin/gex-dev-kit && $output == *'Kit GEX · instalador'* && $output == *$'\nPATH\n'* && $output != *"$(printf '\033')"* ]] || fail 'saída sem ANSI e PATH em bloco'

for shell in zsh bash fish; do
  SHELL="/bin/$shell" GEX_TEST_OS=linux GEX_TEST_ARCH=amd64 run --dir "$work/path-$shell"
  case $shell in zsh) expected='.zshrc' ;; bash) expected='.bashrc' ;; fish) expected='fish_add_path' ;; esac
  expect "$expected" "dica PATH $shell"
  [[ $output == *$'\nPATH\n'* ]] || fail "título PATH $shell"
done
printf 'ok: opções, atualização e dicas PATH\n'

cp "$work/assets/SHA256SUMS" "$work/original-sums"
printf '0%.0s' {1..64} > "$work/assets/SHA256SUMS"
printf '  gex-dev-kit_v0.3.0_linux_amd64\n' >> "$work/assets/SHA256SUMS"
GEX_TEST_OS=linux GEX_TEST_ARCH=amd64 run --dir "$work/bad-hash"
reject 'SHA-256' 'checksum adulterado'
[[ ! -e $work/bad-hash/gex-dev-kit ]] || fail 'instalou checksum adulterado'
printf '0%.0s' {1..64} > "$work/assets/SHA256SUMS"
printf '  arquivo-errado\n' >> "$work/assets/SHA256SUMS"
GEX_TEST_OS=linux GEX_TEST_ARCH=amd64 run --dir "$work/missing-line"
reject 'SHA256SUMS' 'linha ausente'
[[ ! -e $work/missing-line/gex-dev-kit ]] || fail 'instalou sem linha SHA256SUMS'
cp "$work/original-sums" "$work/assets/SHA256SUMS"
printf 'ok: checksum inválido e linha ausente bloqueados\n'

FAKE_AUTH=invalid GEX_TEST_OS=linux GEX_TEST_ARCH=amd64 run --dir "$work/no-auth"
reject 'gh auth login' 'gh sem autenticação'
[[ ! -e $work/no-auth/gex-dev-kit ]] || fail 'instalou sem autenticação'
mv "$work/bin/gh" "$work/gh-hidden"
GEX_TEST_OS=linux GEX_TEST_ARCH=amd64 run --dir "$work/no-gh"
reject 'gh auth login' 'gh ausente'
[[ ! -e $work/no-gh/gex-dev-kit ]] || fail 'instalou sem gh'
mv "$work/gh-hidden" "$work/bin/gh"
GEX_TEST_OS=plan9 GEX_TEST_ARCH=amd64 run --dir "$work/bad-os"
reject 'darwin/arm64' 'SO sem suporte'
GEX_TEST_OS=linux GEX_TEST_ARCH=mips run --dir "$work/bad-arch"
reject 'linux/amd64' 'arquitetura sem suporte'
[[ -z $(find "$work/tmp" -mindepth 1 -print) ]] || fail 'temporários não apagados após erro'
printf 'ok: dependências e plataformas sem suporte\n'
printf 'test-install: ok\n'
