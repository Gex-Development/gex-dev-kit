#!/usr/bin/env bash
set -euo pipefail

binary=${1:?Uso: scripts/smoke.sh CAMINHO_DO_BINARIO}
work=$(mktemp -d)
trap 'rm -rf -- "$work"' EXIT
home="$work/home"
mkdir -p "$home/.claude" "$home/.codex" "$home/.local/bin"
printf '%s\n' '{"model":"modelo-pessoal","hooks":{"PreToolUse":[{"matcher":"Read","hooks":[{"type":"command","command":"true"}]}]}}' > "$home/.claude/settings.json"
printf '%s\n' '{"hooks":{"PreToolUse":[{"matcher":"Read","hooks":[{"type":"command","command":"true"}]}]}}' > "$home/.codex/hooks.json"
chmod 600 "$home/.claude/settings.json" "$home/.codex/hooks.json"
cp "$home/.claude/settings.json" "$work/claude-before"
cp "$home/.codex/hooks.json" "$work/codex-before"
export HOME="$home" USERPROFILE="$home" CODEX_HOME="$home/.codex" CLAUDE_CONFIG_DIR="$home/.claude"
export PATH="$home/.local/bin:$PATH"
"$binary" install --home "$home" --skill tdd --hook gex-conventional-commits --skip-agent-update --yes
if doctor_output=$("$binary" doctor --home "$home"); then
  doctor_status=0
else
  doctor_status=$?
fi
printf '%s\n' "$doctor_output"
if [[ $doctor_status -gt 1 || $doctor_output == *'alterado '* || $doctor_output == *'ausente '* || $doctor_output == *'incompleto '* ]]; then
  printf 'Doctor reportou falha no Core instalado\n' >&2
  exit 1
fi
test -f "$home/.claude/skills/tdd/SKILL.md"
test -f "$home/.agents/skills/tdd/SKILL.md"
test -x "$home/.local/share/gex-dev-kit/bin/gex-dev-kit"
printf '%s' '{"tool_name":"Bash","tool_input":{"command":"git commit -m '\''fix: correto'\''"}}' | "$binary" hook commit-msg
if printf '%s' '{"tool_name":"Bash","tool_input":{"command":"git commit -m '\''mensagem ruim'\''"}}' | "$binary" hook commit-msg; then
  printf 'Hook aceitou mensagem inválida\n' >&2
  exit 1
fi
"$binary" rollback --home "$home"
cmp "$work/claude-before" "$home/.claude/settings.json"
cmp "$work/codex-before" "$home/.codex/hooks.json"
test ! -e "$home/.claude/skills/tdd/SKILL.md"
test ! -e "$home/.agents/skills/tdd/SKILL.md"
test ! -e "$home/.local/share/gex-dev-kit/bin/gex-dev-kit"
printf 'Smoke isolado: Core, skill, hook, doctor e rollback conferidos.\n'
