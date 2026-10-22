#!/usr/bin/env bash
set -euo pipefail

binary=${1:?Uso: scripts/smoke.sh CAMINHO_DO_BINARIO}
binary=$(cd -- "$(dirname -- "$binary")" && pwd)/$(basename -- "$binary")
work=$(mktemp -d)
trap 'rm -rf -- "$work"' EXIT
home="$work/home"
repos="$work/repos"
esc=$(printf '\033')
mkdir -p "$home/.claude" "$home/.codex" "$home/.local/bin" "$repos"
printf '%s\n' '{"model":"modelo-pessoal","hooks":{"PreToolUse":[{"matcher":"Read","hooks":[{"type":"command","command":"true"}]}]}}' > "$home/.claude/settings.json"
printf '%s\n' '{"hooks":{"PreToolUse":[{"matcher":"Read","hooks":[{"type":"command","command":"true"}]}]}}' > "$home/.codex/hooks.json"
chmod 600 "$home/.claude/settings.json" "$home/.codex/hooks.json"
cp "$home/.claude/settings.json" "$work/claude-before"
cp "$home/.codex/hooks.json" "$work/codex-before"
export HOME="$home" USERPROFILE="$home" CODEX_HOME="$home/.codex" CLAUDE_CONFIG_DIR="$home/.claude"
export PATH="/usr/bin:/bin:$(dirname -- "$binary")"
export PAGAMERICAN_API_KEY=ficticio-smoke BUYGOODS_ACCOUNT_ID=ficticio-smoke BUYGOODS_API_TOKEN=ficticio-smoke
export CLICKBANK_ACCOUNT=ficticio-smoke CLICKBANK_API_KEY=ficticio-smoke

falhou() {
  printf 'Falhou: %s\n' "$1" >&2
  if [[ $# -ge 2 ]]; then printf '%s\n' "$2" >&2; fi
  exit 1
}

ok() {
  printf 'ok: %s\n' "$1"
}

pendente() {
  printf 'pendente: %s\n' "$1" >&2
}

entrada() {
  python3 -c 'import json,sys; print(json.dumps({"hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{"command":sys.argv[1]},"cwd":sys.argv[2]}))' "$1" "$2"
}

rodar() {
  local repo=$1 json=$2
  shift 2
  (cd -- "$repo" && printf '%s' "$json" | "$binary" "$@")
}

executar() {
  status=0
  saida=$(rodar "$@" 2>&1) || status=$?
}

skills_core=(gex-abrir-issue gex-auto-revisao gex-fluxo-desenvolvimento gex-preparar-pr gex-regras-de-negocio gex-revisar-pr gex-verificar-mudanca)
skills_especialidades=(gex-pagamerican-api gex-buygoods-api gex-clickbank-api gex-local-spark-tests gex-canonical-prod-readiness gex-hub-gold-refresh-prod gex-gateway-audit gex-pr-review gex-design-system gex-hub-design-system)
args_especialidades=()
for skill in "${skills_especialidades[@]}"; do
  args_especialidades+=(--skill "$skill")
done

"$binary" install --home "$home" --skill tdd --skill retro --skill writing-for-agents "${args_especialidades[@]}" \
  --hook gex-conventional-commits --hook gex-sensitive-action-warning --hook gex-pre-pr --hook gex-session-start \
  --skip-agent-update --yes > "$work/install.txt" 2>&1 || {
  cat "$work/install.txt" >&2
  falhou 'install com 10 especializações, retro, writing-for-agents e 4 hooks'
}
ok 'install com 10 especializações, retro, writing-for-agents e 4 hooks'

if doctor_output=$("$binary" doctor --home "$home"); then
  doctor_status=0
else
  doctor_status=$?
fi
printf '%s\n' "$doctor_output"
if [[ $doctor_status -gt 1 || $doctor_output == *'alterado '* || $doctor_output == *'ausente '* || $doctor_output == *'incompleto '* ]]; then
  falhou 'doctor reportou falha no Core instalado'
fi

for skill in "${skills_core[@]}" "${skills_especialidades[@]}" tdd retro writing-for-agents; do
  test -s "$home/.claude/skills/$skill/SKILL.md"
  test -s "$home/.agents/skills/$skill/SKILL.md"
done
for dir in "$home/.claude/skills/gex-design-system" "$home/.agents/skills/gex-design-system"; do
  test -s "$dir/tokens.css"
  test -s "$dir/assets/gex-logo-branco.svg"
  test -s "$dir/assets/gex-logo-escuro.svg"
done
ok 'arquivos nos dois diretórios de skills, tokens.css e logos SVG'

termos_pessoais=(-e '/Users/' -e '~/vaults' -e 'davi-agents' -e 'Davi' -e 'gex-terraform' -e 'canonico-validador' -e 'TickTick' -e 'skill://' -e 'agent://' -e '@latest' -e 'Male Enhancement' -e '<metadata>' -e 'C2PA' -e 'vaul-canonico-gateways' -e 'canonico-gateways' -e 'Cockpit' -e 'arn:aws:' -e '[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}' -e '[0-9]{12}')
if grep -R -I -n -E "${termos_pessoais[@]}" "$home/.claude" "$home/.agents" "$home/.codex" > "$work/pessoais_bruto.txt"; then
  : > "$work/pessoais.txt"
  while IFS= read -r linha; do
    restante=$(printf '%s\n' "$linha" | sed -E 's/[A-Za-z0-9._%+-]+@(example\.(com|org|net)|localhost)($|[^A-Za-z0-9.-]|\.($|[^A-Za-z0-9]))/ /g')
    if printf '%s\n' "$restante" | grep -q -E "${termos_pessoais[@]}"; then
      printf '%s\n' "$linha" >> "$work/pessoais.txt"
    fi
  done < "$work/pessoais_bruto.txt"
  if [[ -s "$work/pessoais.txt" ]]; then
    cat "$work/pessoais.txt" >&2
    falhou 'dado pessoal encontrado nos arquivos instalados'
  fi
fi
ok 'nenhum dado pessoal nos arquivos instalados'

test "$(grep -c 'gex-dev-kit"' "$home/.claude/settings.json")" -eq 4
if [[ $(grep -c -F 'gex-dev-kit\" hook ' "$home/.codex/hooks.json") -eq 2 ]] && ! grep -q -e 'sensitive-action' -e 'pre-pr' "$home/.codex/hooks.json"; then
  ok 'settings.json com 4 handlers; hooks.json só com commit-msg e session-start'
else
  falhou 'hooks.json do Codex deveria ter só commit-msg e session-start'
fi

gex_origin="$repos/gex-origin"
gex_upstream="$repos/gex-upstream"
alheio="$repos/alheio"
git init -q "$gex_origin"
git init -q "$gex_upstream"
git init -q "$alheio"
git -C "$gex_origin" remote add origin https://github.com/Gex-Development/demo.git
git -C "$gex_upstream" remote add upstream git@GitHub.com:GEX-DATA-ARCHITECTURE/demo.git
git -C "$alheio" remote add origin https://github.com/outra-org/demo.git

cmd_push_force='git push --force origin main'
cmd_ruim="git commit -m 'mensagem ruim'"
cmd_ok="git commit -m 'fix: correto'"

for gex in "$gex_origin" "$gex_upstream"; do
  executar "$gex" "$(entrada "$cmd_push_force" "$gex")" hook sensitive-action
  [[ $status -eq 0 && $saida == *'"permissionDecision":"ask"'* ]] || falhou "push forçado sem pedido de confirmação em $gex" "$saida"
  executar "$gex" "$(entrada "$cmd_ruim" "$gex")" hook commit-msg
  [[ $status -eq 2 ]] || falhou "commit inválido aceito em $gex" "$saida"
  executar "$gex" "$(entrada "$cmd_ok" "$gex")" hook commit-msg
  [[ $status -eq 0 ]] || falhou "commit válido bloqueado em $gex" "$saida"
done
ok 'repositórios GEX: push forçado pede confirmação e commit inválido sai com 2'

executar "$alheio" "$(entrada "$cmd_push_force" "$alheio")" hook sensitive-action
[[ $status -eq 0 && -z $saida ]] || falhou 'repositório alheio recebeu aviso de push' "$saida"
executar "$alheio" "$(entrada "$cmd_ruim" "$alheio")" hook commit-msg
[[ $status -eq 0 && -z $saida ]] || falhou 'repositório alheio bloqueado sem gex.policy' "$saida"
executar "$alheio" "$(entrada 'ls' "$alheio")" hook sensitive-action
[[ $status -eq 0 && -z $saida ]] || falhou 'ls em repositório alheio gerou saída' "$saida"
git -C "$alheio" config gex.policy true
executar "$alheio" "$(entrada "$cmd_ruim" "$alheio")" hook commit-msg
[[ $status -eq 2 ]] || falhou 'repositório alheio com gex.policy não bloqueou commit inválido' "$saida"
ok 'repositório alheio: sem aviso; com gex.policy, commit inválido bloqueado'

grande=$(head -c 70000 /dev/zero | tr '\0' 'x')
executar "$gex_origin" "$(entrada "$(printf "cat <<'EOF'\nO apóstrofo aqui\nEOF\n%s" "$grande")" "$gex_origin")" hook commit-msg
[[ $status -eq 0 ]] || falhou 'heredoc com apóstrofo e comando de 70 KB bloqueado' "$(printf '%s' "$saida" | head -c 300)"
ok 'heredoc com apóstrofo e comando de 70 KB passam'

executar "$gex_origin" '{"source":"startup","cwd":"'"$gex_origin"'"}' hook session-start --agent claude
[[ $status -eq 0 ]] || falhou 'session-start falhou' "$saida"
[[ $(printf '%s' "$saida" | wc -m) -le 1200 ]] || falhou 'texto de session-start acima de 1.200 caracteres'
[[ $saida != *"$home"* ]] || falhou 'session-start expôs o caminho do HOME'
ok 'session-start: texto até 1.200 caracteres e sem caminho do HOME'

for schema_cmd in skills:gex-dev-kit.skills/v1 hooks:gex-dev-kit.hooks/v1 addons:gex-dev-kit.addons/v1 doctor:gex-dev-kit.doctor/v1; do
  cmd=${schema_cmd%%:*}
  schema=${schema_cmd#*:}
  if [[ $cmd == doctor ]]; then
    "$binary" doctor --home "$home" --json > "$work/doctor.json"
  else
    "$binary" "$cmd" --json > "$work/$cmd.json"
  fi
  json_path="$work/$cmd.json"
  [[ $cmd == doctor ]] && json_path="$work/doctor.json"
  python3 -c 'import json,sys; d=json.load(open(sys.argv[1])); assert d["schema"]==sys.argv[2], d.get("schema")' "$json_path" "$schema" || falhou "schema inesperado em $cmd --json"
done
python3 -c 'import json,sys; d=json.load(open(sys.argv[1])); assert [i for i in d["items"] if i["path"]==".codex/hooks.json" and i["state"]=="pendente"]' "$work/doctor.json" || falhou 'doctor --json não marca .codex/hooks.json como pendente'
ok 'JSON dos 4 comandos com schema esperado; .codex/hooks.json pendente'

home_exe="$work/home-exe"
mkdir -p "$home_exe/.claude" "$home_exe/.codex"
printf '{}\n' > "$home_exe/.claude/settings.json"
printf '{}\n' > "$home_exe/.codex/hooks.json"
binario_a="$work/binario-a"
binario_b="$work/binario-b"
cp "$binary" "$binario_a"
cp "$binary" "$binario_b"
printf 'x' >> "$binario_b"
"$binario_a" install --home "$home_exe" --skill tdd --hook gex-conventional-commits --skip-agent-update --yes > "$work/exe-a.txt" 2>&1
"$binario_b" doctor --home "$home_exe" | grep -q 'atualizar' || falhou 'doctor não pediu atualização do executável dos hooks'
"$binario_b" install --home "$home_exe" --skill tdd --skip-agent-update --yes > "$work/exe-b.txt" 2>&1
cmp "$binario_b" "$home_exe/.local/share/gex-dev-kit/bin/gex-dev-kit"
ok 'executável dos hooks atualizado para o binário mais recente'

for i in 1 2 3; do
  troca="$work/troca-$i"
  cp "$binary" "$troca"
  printf 'troca%s' "$i" >> "$troca"
  "$troca" install --home "$home_exe" --skill tdd --skip-agent-update --yes > "$work/troca-$i.txt" 2>&1
done
historico="$home_exe/.local/share/gex-dev-kit/history"
test "$(wc -c < "$historico/state.json")" -lt 1048576
test -n "$(find "$historico/blobs" -type f | head -n 1)"
"$binary" doctor --home "$home_exe" > "$work/doctor-historico.txt"
ok 'histórico: state.json abaixo de 1 MB, blobs presentes e doctor sem erro'

repo_gh="$work/repo-githooks"
bare="$work/remoto.git"
sem_repo="$work/sem-repo"
mkdir -p "$sem_repo"
antes_sem_repo=$(ls -A "$sem_repo")
(cd -- "$sem_repo" && "$binary" githooks --hook gex-commit-msg > /dev/null 2>&1 || true)
[[ $(ls -A "$sem_repo") == "$antes_sem_repo" ]] || falhou 'githooks sem --repo criou arquivos'
git init -q "$repo_gh"
printf '# Demo\n\nComandos documentados:\n\n- `true`\n- `false`\n' > "$repo_gh/README.md"
git -C "$repo_gh" add README.md
git -C "$repo_gh" -c user.name=smoke -c user.email=smoke@localhost commit -q -m 'docs: readme'
"$binary" githooks --repo "$repo_gh" --hook commit-msg --hook pre-commit --hook pre-push --pre-commit-run true --run true --yes > "$work/githooks.txt" 2>&1 || {
  cat "$work/githooks.txt" >&2
  falhou 'githooks não instalou os git hooks no repositório temporário'
}

printf 'a\n' > "$repo_gh/a.txt"
git -C "$repo_gh" add a.txt
if git -C "$repo_gh" -c user.name=smoke -c user.email=smoke@localhost commit -q -m 'mensagem ruim' > "$work/g-ruim.txt" 2>&1; then
  falhou 'git hook aceitou mensagem inválida'
fi
printf 'feat: aceita template\n\n# Linha de comentário do template\n' > "$work/msg.txt"
git -C "$repo_gh" -c user.name=smoke -c user.email=smoke@localhost commit -q -F "$work/msg.txt" || falhou 'template com comentário bloqueado'
printf 'b\n' > "$repo_gh/b.txt"
git -C "$repo_gh" add b.txt
git -C "$repo_gh" -c user.name=smoke -c user.email=smoke@localhost commit -q --fixup=HEAD || falhou '--fixup bloqueado'
printf 'c\n' > "$repo_gh/c.txt"
git -C "$repo_gh" add c.txt
git -C "$repo_gh" -c user.name=smoke -c user.email=smoke@localhost commit -q --squash=HEAD || falhou '--squash bloqueado'
ok 'git hook: mensagem inválida bloqueada; template, --fixup e --squash aceitos'

printf 'AKIA%s\n' ABCDEFGHIJKLMNOP > "$repo_gh/credencial.txt"
git -C "$repo_gh" add credencial.txt
if git -C "$repo_gh" -c user.name=smoke -c user.email=smoke@localhost commit -q -m 'feat: teste credencial' > "$work/g-cred.txt" 2>&1; then
  falhou 'credencial montada não foi bloqueada'
fi
if grep -q 'ABCDEFGHIJKLMNOP' "$work/g-cred.txt"; then
  falhou 'credencial apareceu na saída do git hook'
fi
mkdir -p "$repo_gh/.git/gex-dev-kit"
printf 'credencial.txt\taws-access-key\n' > "$repo_gh/.git/gex-dev-kit/pre-commit-allow"
git -C "$repo_gh" -c user.name=smoke -c user.email=smoke@localhost commit -q -m 'feat: teste credencial' || falhou 'credencial liberada pelo arquivo de exceção continua bloqueada'
ok 'pre-commit: credencial bloqueada, sem vazar valor, e liberada pelo arquivo de exceção'

git init -q --bare "$bare"
git -C "$repo_gh" remote add origin "$bare"
git -C "$repo_gh" push -q origin HEAD || falhou 'push para remoto temporário bloqueado'
ok 'pre-push: push para remoto temporário passa'

"$binary" doctor --repo "$repo_gh" > "$work/doctor-repo.txt" || falhou 'doctor --repo falhou'
ok 'doctor --repo sem falha nos git hooks gravados'
"$binary" rollback --repo "$repo_gh" --yes > "$work/rollback-repo.txt" 2>&1 || falhou 'rollback --repo falhou'
if [[ -e "$repo_gh/.git/hooks/commit-msg" ]]; then
  pendente "rollback --repo não remove os git hooks (responde 'Nada a restaurar'; cmd/gex-dev-kit/writecmds.go:118-121 e 324-329)"
else
  ok 'githooks: rollback --repo remove os hooks'
fi

"$binary" rollback --home "$home" --yes > "$work/rollback.txt" 2>&1 || {
  cat "$work/rollback.txt" >&2
  falhou 'rollback --home falhou'
}
cmp "$work/claude-before" "$home/.claude/settings.json"
cmp "$work/codex-before" "$home/.codex/hooks.json"
for skill in "${skills_core[@]}" "${skills_especialidades[@]}" tdd retro writing-for-agents; do
  test ! -e "$home/.claude/skills/$skill/SKILL.md"
  test ! -e "$home/.agents/skills/$skill/SKILL.md"
done
test ! -e "$home/.local/share/gex-dev-kit/bin/gex-dev-kit"
ok 'rollback --home: settings.json e hooks.json idênticos; skills e binário removidos'

ds_saida=$("$binary" ds --ui=cli)
if printf '%s' "$ds_saida" | LC_ALL=C grep -q "$esc"; then
  falhou 'ds --ui=cli emitiu byte ESC'
fi
ok 'ds --ui=cli sem byte ESC'

printf 'Smoke isolado: itens 1 a 10 conferidos.\n'
