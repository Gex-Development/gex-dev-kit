#!/usr/bin/env bash
set -euo pipefail

here=$(cd -- "$(dirname -- "$0")" && pwd)
root=$(cd -- "$here/.." && pwd)
sandbox=${UX_SANDBOX:-$(mktemp -d "${TMPDIR:-/tmp}/gdk-ux-XXXXXX")}
mkdir -p "$sandbox/home" "$sandbox/repo" "$sandbox/xdg" "$sandbox/bin" "$sandbox/tmp"
ambiente=(
  "HOME=$sandbox/home"
  "USERPROFILE=$sandbox/home"
  "CLAUDE_CONFIG_DIR=$sandbox/home/.claude"
  "CODEX_HOME=$sandbox/home/.codex"
  "XDG_CONFIG_HOME=$sandbox/xdg/config"
  "XDG_DATA_HOME=$sandbox/xdg/data"
  "XDG_STATE_HOME=$sandbox/xdg/state"
  "XDG_CACHE_HOME=$sandbox/xdg/cache"
  "TMPDIR=$sandbox/tmp"
  "PATH=/usr/bin:/bin:$sandbox/bin"
)
binary="$sandbox/bin/gex-dev-kit"

recusa() {
  printf 'ux-audit-env: %s\n' "$1" >&2
  exit 2
}

if [[ ! -x $binary ]]; then
  (cd -- "$root" && go build -o "$binary" ./cmd/gex-dev-kit)
fi

if [[ $# -eq 0 ]]; then
  printf 'export UX_SANDBOX=%q\n' "$sandbox"
  for kv in "${ambiente[@]}"; do
    printf 'export %s=%q\n' "${kv%%=*}" "${kv#*=}"
  done
  exit 0
fi

first=${1:-}
if [[ $first == --run ]]; then
  script=${2:?Uso: scripts/ux-audit-env.sh --run ARQUIVO.sh}
  script=$(cd -- "$(dirname -- "$script")" && pwd)/$(basename -- "$script")
  exec env -i "${ambiente[@]}" bash "$script"
fi

ajuda=0
for arg in "$@"; do
  case $arg in
    -h|--help) ajuda=1 ;;
  esac
done
if [[ $first == help || $ajuda -eq 1 ]]; then
  exec env -i "${ambiente[@]}" "$binary" "$@"
fi

case $first in
  rollback-agents) recusa 'rollback-agents não roda na auditoria' ;;
  install)
    tem_skip=0
    for arg in "$@"; do
      if [[ $arg == --skip-agent-update ]]; then tem_skip=1; fi
    done
    if [[ $tem_skip -eq 0 ]]; then recusa 'install exige --skip-agent-update na auditoria'; fi
    ;;
esac

anterior=''
for arg in "$@"; do
  valor=''
  case $anterior in
    --home|--repo) valor=$arg ;;
  esac
  case $arg in
    --home=*|--repo=*) valor=${arg#*=} ;;
  esac
  if [[ -n $valor ]]; then
    case $valor in
      "$sandbox"|"$sandbox"/*) ;;
      *) recusa "$valor está fora do sandbox da auditoria" ;;
    esac
  fi
  anterior=$arg
done

exec env -i "${ambiente[@]}" "$binary" "$@"
