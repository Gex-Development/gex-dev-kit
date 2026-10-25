#!/usr/bin/env bash
set -euo pipefail

repo=Gex-Development/gex-dev-kit
version=${GEX_VERSION:-}
install_dir=${GEX_INSTALL_DIR:-$HOME/.local/bin}
temp_dir=
trap 'if [[ -n $temp_dir ]]; then rm -rf -- "$temp_dir"; fi' EXIT

fail() {
  if [[ -t 2 && -z ${NO_COLOR:-} ]]; then printf '✗ %s\n' "$1" >&2
  else printf 'erro: %s\n' "$1" >&2; fi
  exit 1
}
step() {
  if [[ -t 1 && -z ${NO_COLOR:-} ]]; then printf '\033[2m[%s/5]\033[0m %s\n' "$1" "$2"
  else printf '[%s/5] %s\n' "$1" "$2"; fi
}
mark() {
  if [[ -t 1 && -z ${NO_COLOR:-} ]]; then printf '\033[32m✓\033[0m %s\n' "$1"; else printf 'ok: %s\n' "$1"; fi
}
usage() {
  printf 'Uso: bash install.sh [--version vX.Y.Z] [--dir CAMINHO]\n'
  printf 'Padrão: última release, pasta ~/.local/bin; GEX_VERSION e GEX_INSTALL_DIR alteram os padrões.\n'
  printf 'Teste: GEX_TEST_OS e GEX_TEST_ARCH simulam uname (somente para testes).\n'
}
parse_args() {
  while [[ $# -gt 0 ]]; do
    case $1 in
      --version|--dir)
        [[ $# -ge 2 && -n $2 ]] || fail "Falta valor para $1. Use --help."
        if [[ $1 == --version ]]; then version=$2; else install_dir=$2; fi
        shift 2 ;;
      -h|--help) usage; exit 0 ;;
      *) fail "Opção desconhecida: $1. Use --help." ;;
    esac
  done
}
detect_os() {
  case ${GEX_TEST_OS:-$(uname -s)} in
    Darwin|darwin) os=darwin ;;
    Linux|linux) os=linux ;;
    *) fail 'SO sem build. Suportados: darwin/arm64, darwin/amd64, linux/arm64, linux/amd64, windows/arm64, windows/amd64.' ;;
  esac
}
detect_arch() {
  local machine=${GEX_TEST_ARCH:-$(uname -m)}
  if [[ -z ${GEX_TEST_ARCH:-} && $os == darwin && $machine == x86_64 ]] && [[ $(sysctl -n hw.optional.arm64 2>/dev/null || true) == 1 ]]; then machine=arm64; fi
  case $machine in
    arm64|aarch64) arch=arm64 ;;
    amd64|x86_64) arch=amd64 ;;
    *) fail 'Arquitetura sem build. Suportados: darwin/arm64, darwin/amd64, linux/arm64, linux/amd64, windows/arm64, windows/amd64.' ;;
  esac
}
check_gh() {
  if ! command -v gh >/dev/null 2>&1; then
    if [[ $os == darwin ]]; then
      fail 'Instale gh com brew install gh e entre com gh auth login.'
    fi
    fail 'Instale gh via apt/dnf: https://github.com/cli/cli/blob/trunk/docs/install_linux.md; depois gh auth login.'
  fi
  gh auth status >/dev/null 2>&1 || fail 'GitHub CLI sem autenticação. Rode gh auth login.'
}
resolve_version() {
  if [[ -z $version ]]; then
    version=$(gh release view --repo "$repo" --json tagName --jq .tagName 2>/dev/null) || fail 'Release indisponível. Peça acesso à organização Gex-Development.'
  fi
  [[ $version =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] || fail 'Versão inválida. Use vX.Y.Z.'
}
download() {
  temp_dir=$(mktemp -d) || fail 'Não foi possível criar pasta temporária.'
  gh release download "$version" --repo "$repo" --pattern "$asset" --dir "$temp_dir" >/dev/null 2>&1 || fail 'Binário indisponível. Confira a versão e peça acesso à organização Gex-Development.'
  gh release download "$version" --repo "$repo" --pattern SHA256SUMS --dir "$temp_dir" >/dev/null 2>&1 || fail 'SHA256SUMS indisponível. Confira a release e peça acesso à organização Gex-Development.'
  [[ -f $temp_dir/$asset && -f $temp_dir/SHA256SUMS ]] || fail 'Release incompleta: binário ou SHA256SUMS ausente.'
}
verify_checksum() {
  local line expected actual matches=0
  while IFS= read -r line; do
    if [[ ${#line} -ge 66 && ( ${line:64} == "  $asset" || ${line:64} == " *$asset" ) && ${line:0:64} =~ ^[[:xdigit:]]{64}$ ]]; then
      expected=${line:0:64}
      matches=$((matches + 1))
    fi
  done < "$temp_dir/SHA256SUMS"
  [[ $matches -eq 1 ]] || fail "SHA256SUMS não contém uma linha única para $asset. Instalação cancelada."
  if command -v sha256sum >/dev/null 2>&1; then
    actual=$(sha256sum "$temp_dir/$asset")
  else
    actual=$(shasum -a 256 "$temp_dir/$asset")
  fi
  actual=${actual%% *}
  expected=$(printf '%s' "$expected" | tr 'ABCDEF' 'abcdef')
  [[ $actual == "$expected" ]] || fail 'SHA-256 diferente do SHA256SUMS. Instalação cancelada.'
}
install_binary() {
  local old='' installed
  installed="$install_dir/gex-dev-kit"
  if [[ -f $installed ]]; then old=$("$installed" --version 2>/dev/null || true); fi
  mkdir -p -- "$install_dir" || fail "Não foi possível criar $install_dir."
  local staged
  staged=$(mktemp "$install_dir/.gex-dev-kit.XXXXXX") || fail "Não foi possível gravar em $install_dir."
  if ! cp -- "$temp_dir/$asset" "$staged" || ! chmod +x "$staged"; then rm -f -- "$staged"; fail 'Falha ao preparar binário.'; fi
  local observed
  observed=$("$staged" --version 2>/dev/null) || { rm -f -- "$staged"; fail 'Binário baixado não executou --version.'; }
  [[ $observed == "$version" ]] || { rm -f -- "$staged"; fail "Versão divergente: $observed (esperada: $version)."; }
  if ! mv -f -- "$staged" "$installed"; then rm -f -- "$staged"; fail 'Falha ao instalar binário.'; fi
  observed=$("$installed" --version 2>/dev/null) || fail 'Binário instalado não executou --version.'
  [[ $observed == "$version" ]] || fail "Versão instalada divergente: $observed (esperada: $version)."
  if [[ -n $old ]]; then printf 'Versão: %s → %s\n' "$old" "$observed"; else printf 'Versão: %s\n' "$observed"; fi
}
path_hint() {
  case :$PATH: in *:"$install_dir":*) return ;; esac
  printf '\nPATH\n'
  case ${SHELL:-} in
    */zsh) printf '  echo '\''export PATH="%s:$PATH"'\'' >> ~/.zshrc\n' "$install_dir" ;;
    */fish) printf '  fish_add_path -- "%s"\n' "$install_dir" ;;
    *)
      if [[ $os == darwin ]]; then printf '  echo '\''export PATH="%s:$PATH"'\'' >> ~/.bash_profile\n' "$install_dir"
      else printf '  echo '\''export PATH="%s:$PATH"'\'' >> ~/.bashrc\n' "$install_dir"; fi ;;
  esac
}

main() {
  parse_args "$@"
  printf 'Kit GEX · instalador\n'
  detect_os
  detect_arch
  step 1 'Conferindo GitHub CLI e acesso…'
  check_gh
  step 2 'Identificando release e plataforma…'
  resolve_version
  asset="gex-dev-kit_${version}_${os}_${arch}"
  mark "$version ($os/$arch)"
  step 3 'Baixando binário e SHA256SUMS…'
  download
  step 4 'Conferindo SHA-256…'
  verify_checksum
  mark 'integridade confirmada'
  step 5 'Instalando e verificando versão…'
  install_binary
  local display_dir=$install_dir
  case $display_dir in "$HOME"/*) display_dir="~${display_dir#$HOME}" ;; esac
  mark "gex-dev-kit $version instalado em $display_dir/gex-dev-kit"
  printf '\nPróximos passos\n'
  if [[ -t 1 && -z ${NO_COLOR:-} ]]; then
    printf '  \033[1m%-21s\033[0m%s\n' 'gex-dev-kit' 'abre o assistente'
    printf '  \033[1m%-21s\033[0m%s\n' 'gex-dev-kit doctor' 'confere a instalação'
  else
    printf '  %-21s%s\n' 'gex-dev-kit' 'abre o assistente'
    printf '  %-21s%s\n' 'gex-dev-kit doctor' 'confere a instalação'
  fi
  path_hint
}

main "$@"
