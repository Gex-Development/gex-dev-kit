#!/usr/bin/env bash
set -euo pipefail

root=$(git rev-parse --show-toplevel)
cd -- "$root"

passo() {
  printf '== %s\n' "$1"
}

verifica_gofmt() {
  passo 'gofmt'
  local pendentes
  pendentes=$(gofmt -l .)
  if [[ -n $pendentes ]]; then
    printf '%s\n' "$pendentes" >&2
    printf 'check: gofmt apontou arquivos fora do padrão\n' >&2
    exit 1
  fi
}

verifica_comentarios() {
  passo 'comentários em código'
  local falhou=0 arquivo linhas
  while IFS= read -r arquivo; do
    [[ -n $arquivo ]] || continue
    [[ -f $arquivo ]] || continue
    case $arquivo in
      internal/catalog/assets/matt/*) continue ;;
      *.go)
        linhas=$(grep -n -E '^[[:space:]]*//' "$arquivo" | grep -v -E '^[0-9]+:[[:space:]]*//go:(build|embed|generate)' || true)
        ;;
      *.sh)
        linhas=$(grep -n -E '^[[:space:]]*#' "$arquivo" | grep -v -E '^1:#!' || true)
        ;;
      *.py)
        linhas=$(grep -n -E '^[[:space:]]*#' "$arquivo" | grep -v -E '^1:#!' || true)
        linhas="$linhas$(grep -n -F '"""' "$arquivo" || true)"
        ;;
      *.html)
        linhas=$(grep -n -F '<!--' "$arquivo" || true)
        ;;
      *.css|*.js)
        linhas=$(grep -n -F '/*' "$arquivo" || true)
        ;;
      *)
        continue
        ;;
    esac
    if [[ -n $linhas ]]; then
      printf '%s\n%s\n' "$arquivo" "$linhas" >&2
      falhou=1
    fi
  done < <(git ls-files --cached --others --exclude-standard -- '*.go' '*.sh' '*.py' '*.html' '*.css' '*.js' ':!.claude/**')
  if [[ $falhou -ne 0 ]]; then
    printf 'check: comentários encontrados em código (ver linhas acima)\n' >&2
    exit 1
  fi
}

verifica_rotulos_ia() {
  passo 'rótulos de autoria de IA'
  local coautoria gerado padrao log_ausente=0 achados
  coautoria="Co-Authored""-By: Claude"
  gerado="Generated"" with .*Claude"
  padrao="$coautoria|$gerado"
  if ! git rev-parse --verify -q main >/dev/null; then
    log_ausente=1
  fi
  achados=""
  if [[ $log_ausente -eq 0 ]]; then
    achados=$(git log --format=%B main..HEAD | grep -n -i -E "$padrao" || true)
  fi
  if [[ -n $achados ]]; then
    printf 'Mensagens de commit da branch:\n%s\n' "$achados" >&2
    printf 'check: rótulo de autoria de IA em mensagem de commit\n' >&2
    exit 1
  fi
  achados=$(git ls-files --cached --others --exclude-standard -- '*.go' '*.sh' '*.py' '*.html' '*.js' '*.css' ':!.claude/**' | xargs -I{} grep -l -i -E "$padrao" {} 2>/dev/null || true)
  if [[ -n $achados ]]; then
    printf '%s\n' "$achados" >&2
    printf 'check: rótulo de autoria de IA em arquivo de código\n' >&2
    exit 1
  fi
}

verifica_gofmt

passo 'go vet'
go vet ./...
passo 'go vet windows/amd64'
GOOS=windows GOARCH=amd64 go vet ./...
passo 'go vet windows/arm64'
GOOS=windows GOARCH=arm64 go vet ./...

passo 'go test'
go test ./...
passo 'go test -race'
go test -race ./...
passo 'instalador'
/bin/bash "$root/scripts/test-install.sh"

passo 'release'
/bin/bash "$root/scripts/test-release.sh"

passo 'go build'
go build ./...

verifica_comentarios
verifica_rotulos_ia

passo 'smoke'
tmp=$(mktemp -d)
trap 'rm -rf -- "$tmp"' EXIT
go build -o "$tmp/gex-dev-kit" ./cmd/gex-dev-kit
bash "$root/scripts/smoke.sh" "$tmp/gex-dev-kit"

printf 'check: ok\n'
