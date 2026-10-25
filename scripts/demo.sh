#!/usr/bin/env bash
set -euo pipefail

root=$(cd -- "$(dirname -- "$0")/.." && pwd)
sandbox=$(mktemp -d "${TMPDIR:-/tmp}/gdk-demo-XXXXXX")
trap 'rm -rf "$sandbox"' EXIT

(cd -- "$root" && go build -o "$sandbox/gex-dev-kit" ./cmd/gex-dev-kit)

printf 'HOME descartável: %s (apagado ao sair)\n' "$sandbox/home" >&2
mkdir -p "$sandbox/home"
env -u CI -u NO_COLOR \
  HOME="$sandbox/home" \
  USERPROFILE="$sandbox/home" \
  XDG_CONFIG_HOME="$sandbox/home/.config" \
  XDG_DATA_HOME="$sandbox/home/.local/share" \
  COLORTERM="${COLORTERM:-truecolor}" \
  TERM="${TERM:-xterm-256color}" \
  "$sandbox/gex-dev-kit" --ui=tui "$@"
