# Kit de Desenvolvimento GEX

CLI para preparar Claude Code e Codex com instruções globais e sete skills de trabalho GEX obrigatórias. Skills Matt Pocock, especializações e um hook de Conventional Commits entram por escolha; outros hooks ficam apenas no catálogo até existir implementação compatível e segura. Addons, MCPs e arquivos de projetos não são alterados.

## Baixar

Repositório privado da organização: é preciso acesso à Gex-Development e autenticação no `gh`. Baixe o executável correspondente ao sistema na [release](https://github.com/Gex-Development/gex-dev-kit/releases). Os alvos publicados são `darwin_{arm64,amd64}`, `linux_{arm64,amd64}` (Ubuntu/Fedora) e `windows_{arm64,amd64}` (PowerShell). Confira `SHA256SUMS` antes de executar.

```sh
gh release download v0.2.0 --repo Gex-Development/gex-dev-kit --dir .
sha256sum -c SHA256SUMS
chmod +x gex-dev-kit_v0.2.0_linux_amd64
./gex-dev-kit_v0.2.0_linux_amd64
```

No macOS, confira checksums com `shasum -a 256 -c SHA256SUMS` e troque `linux_amd64` por `darwin_arm64` ou `darwin_amd64`. No PowerShell, baixe o executável Windows correspondente, confira hash com `Get-FileHash` e execute-o.

Na release v0.2.0, abrir em terminal interativo mostra painel GEX e instalação guiada; essa release **não contém** as sete skills Core, especializações ou as novas telas desta branch local. Use o código atual para avaliar essas mudanças, não o binário publicado.

## Piloto da branch local (ainda não publicada)

```sh
go build -o ./gex-dev-kit ./cmd/gex-dev-kit
PILOT_HOME="$(mktemp -d)"
HOME="$PILOT_HOME" ./gex-dev-kit
./gex-dev-kit doctor --home "$PILOT_HOME"
./gex-dev-kit rollback --home "$PILOT_HOME"
```

O painel mostra estado do Core com spinner, menu navegável e atalhos `?`, `↑`/`↓`, `Enter` e `q`. Na instalação guiada, escolha skills, especializações e hook; no último passo, responda **Agora não** para não atualizar executáveis durante o piloto. A prévia rolável mostra arquivos afetados; `Enter` abre confirmação, `s` aplica e `Esc` cancela. `ACCESSIBLE=1`, `NO_COLOR=1` ou `TERM=dumb` oferecem menu e confirmação textuais. O ciclo acima só altera o HOME temporário, mas mantenha o caminho para conferir o resultado antes de descartá-lo. `doctor` pode sair com código 1 quando encontrar agentes externos bloqueados, mesmo com arquivos do Core íntegros.

Para escolher opções sem o formulário, use `./gex-dev-kit install --home "$PILOT_HOME" --skip-agent-update --skill tdd --hook gex-conventional-commits`; revise a prévia e confirme no terminal. Em automação isolada, `--yes` substitui a confirmação, **não** a prévia.

## Usar

```sh
./gex-dev-kit_v0.2.0_linux_amd64 skills
./gex-dev-kit_v0.2.0_linux_amd64 install --skill tdd --hook gex-conventional-commits
./gex-dev-kit_v0.2.0_linux_amd64 doctor
./gex-dev-kit_v0.2.0_linux_amd64 rollback
```

`install` mostra arquivos e versões afetadas, depois exige confirmação. `--yes` confirma em automação após prévia; `--skill` e `--hook` aceitam repetição. `skills`, `hooks` e `--help` listam opções. Sem seleção adicional, instala Core e tenta atualizar agentes elegíveis. `--home CAMINHO_ABSOLUTO` instala em HOME isolado para teste; `--skip-agent-update` evita rede e alterações nos binários nesse piloto. `doctor` mostra arquivos íntegros, alterações posteriores, versões e confiança pendente de hooks Codex. `rollback` restaura arquivos e modos anteriores; se algum alvo foi editado depois, aborta antes de restaurar qualquer arquivo. Histórico de arquivos do kit fica no perfil do usuário com acesso restrito (DPAPI no Windows).

O Core escreve instruções em `~/.claude/CLAUDE.md` e `~/.codex/AGENTS.md` e instala sete skills `gex-*` em `~/.claude/skills/` e `~/.agents/skills/`. Quatro especializações adicionais são opt-in: `gex-pagamerican-api`, `gex-buygoods-api`, `gex-clickbank-api` e `gex-local-spark-tests`. As três skills gateway incluem clientes Python somente leitura e exigem credenciais no ambiente no momento da seleção (`PAGAMERICAN_API_KEY`; `BUYGOODS_ACCOUNT_ID` e `BUYGOODS_API_TOKEN`; `CLICKBANK_API_KEY` e `CLICKBANK_ACCOUNT`). Instalar uma skill não executa sua consulta. Acesso real depende de credenciais e permissões aprovadas. O hook selecionável `gex-conventional-commits` usa `PreToolUse` nos dois agentes e evita commits com cabeçalho fora de Conventional Commits quando a mensagem é literal. É orientação local, não substitui validação Git por projeto ou proteção remota. No Codex, abra `/hooks`, revise e confie na definição instalada: o kit nunca contorna essa etapa. Configurações existentes são preservadas semanticamente; login, modelo e credenciais não são alterados. O Core limita convenções GEX ao trabalho GEX e respeita instruções de projeto.

Claude Code `2.1.306` e Codex `0.163.0` são versões fixadas nesta release. A CLI só instala ou atualiza executáveis standalone próprios, no perfil do usuário, via instaladores oficiais com versão exata. Brew, npm, winget, apt, dnf, binários não gerenciados, PATH ambíguo e desvios por auto-update nativo ficam bloqueados com ação manual; o Core ainda pode ser instalado, mas o comando retorna erro e não afirma atualização. Login e modelo não mudam. `rollback-agents` tem confirmação e reverte apenas upgrade de binário previamente gerenciado com versão anterior registrada; a primeira instalação não tem versão anterior e não é desinstalada automaticamente. Atualizações nativas posteriores podem desviar do pin e pedir revisão manual.

As 17 skills Matt Pocock são cópias offline do commit [`5beb35b0`](https://github.com/mattpocock/skills/tree/5beb35b040e3de97e54fa9e386d7e42d15c295c2), sob [licença MIT](https://github.com/mattpocock/skills/blob/5beb35b040e3de97e54fa9e386d7e42d15c295c2/LICENSE). Instalar uma skill só copia seus arquivos; não a executa. `setup-matt-pocock-skills` (altera projeto e labels), `implement-spec` (depende de tracker e outras skills) e `ask-matt` (depende do catálogo completo) não foram empacotadas nesta seleção. Skills que criam issues, labels, branches ou alteram projetos exigem autorização separada quando invocadas.

## Desenvolvimento e release

```sh
go test ./...
go vet ./...
go build -o /tmp/gex-dev-kit-piloto ./cmd/gex-dev-kit
./scripts/smoke.sh /tmp/gex-dev-kit-piloto
```

O smoke executa instalação, `doctor`, hook e rollback em HOME descartável. O script de release compila seis alvos e gera `dist/SHA256SUMS`; publicar requer versão nova aprovada e confirmação explícita. Gum pode ser usado no fluxo manual sem virar dependência de instalação. Mudanças seguem Conventional Commits. Sem CI remoto nesta fase: testes locais não equivalem à proteção contínua de branches.

Fontes de compatibilidade: [hooks Claude Code](https://code.claude.com/docs/en/hooks), [hooks Codex](https://developers.openai.com/codex/hooks), [instalação Claude Code](https://code.claude.com/docs/en/setup), [instalação Codex](https://developers.openai.com/codex/cli). Estado e decisões do kit: `kit-desenvolvimento-gex-acompanhamento.html` na documentação visual da equipe.
