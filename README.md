# Kit de Desenvolvimento GEX

CLI para preparar Claude Code e Codex com instruções e skills globais de trabalho GEX. O Core instala sempre; skills de Matt Pocock e hooks compatíveis entram por escolha. Addons, MCPs e arquivos de projetos não são alterados.

## Baixar

Repositório privado da organização: é preciso acesso à Gex-Development e autenticação no `gh`. Baixe o executável correspondente ao sistema na [release](https://github.com/Gex-Development/gex-dev-kit/releases). Os alvos publicados são `darwin_{arm64,amd64}`, `linux_{arm64,amd64}` (Ubuntu/Fedora) e `windows_{arm64,amd64}` (PowerShell). Confira `SHA256SUMS` antes de executar.

```sh
gh release download v0.1.0 --repo Gex-Development/gex-dev-kit --pattern 'gex-dev-kit_v0.1.0_linux_amd64' --dir .
chmod +x gex-dev-kit_v0.1.0_linux_amd64
./gex-dev-kit_v0.1.0_linux_amd64
```

No macOS, troque `linux_amd64` por `darwin_arm64` ou `darwin_amd64`. No PowerShell, baixe `gex-dev-kit_v0.1.0_windows_amd64.exe` (ou `windows_arm64.exe`) e execute `./gex-dev-kit_v0.1.0_windows_amd64.exe`. O executável abre escolha guiada em terminal interativo; fora dele, use `install` com opções explícitas.

## Usar

```sh
gex-dev-kit install --skill tdd --hook gex-conventional-commits
gex-dev-kit doctor
gex-dev-kit rollback
```

`install` mostra arquivos e versões afetadas, depois exige confirmação. `--yes` confirma em automação após prévia; `--skill` e `--hook` aceitam repetição. `skills`, `hooks` e `--help` listam opções. Sem seleção adicional, instala Core e tenta atualizar agentes elegíveis. `--home CAMINHO_ABSOLUTO` instala em HOME isolado para teste; `--skip-agent-update` evita rede e alterações nos binários nesse piloto. `doctor` mostra arquivos íntegros, alterações posteriores, versões e confiança pendente de hooks Codex. `rollback` restaura arquivos e modos anteriores; se algum alvo foi editado depois, aborta antes de restaurar qualquer arquivo. Histórico de arquivos do kit fica no perfil do usuário com acesso restrito (DPAPI no Windows).

O Core escreve instruções em `~/.claude/CLAUDE.md` e `~/.codex/AGENTS.md`; skills GEX ficam em `~/.claude/skills/` e `~/.agents/skills/`. O hook selecionável `gex-conventional-commits` usa `PreToolUse` nos dois agentes e evita commits com cabeçalho fora de Conventional Commits quando a mensagem é literal. É orientação local, não substitui validação Git por projeto ou proteção remota. No Codex, abra `/hooks`, revise e confie na definição instalada: o kit nunca contorna essa etapa. Configurações existentes são preservadas semanticamente; nenhum login, modelo ou credencial é solicitado. O Core limita convenções GEX ao trabalho GEX e respeita instruções de projeto.

Claude Code `2.1.306` e Codex `0.163.0` são versões fixadas nesta release. A CLI só instala ou atualiza executáveis standalone próprios, no perfil do usuário, via instaladores oficiais com versão exata. Brew, npm, winget, apt, dnf, binários não gerenciados, PATH ambíguo e desvios por auto-update nativo ficam bloqueados com ação manual; o Core ainda pode ser instalado, mas o comando retorna erro e não afirma atualização. Login e modelo não mudam. `rollback-agents` tem confirmação e reverte apenas upgrade de binário previamente gerenciado com versão anterior registrada; a primeira instalação não tem versão anterior e não é desinstalada automaticamente. Atualizações nativas posteriores podem desviar do pin e pedir revisão manual.

As 17 skills Matt Pocock são cópias offline do commit [`5beb35b0`](https://github.com/mattpocock/skills/tree/5beb35b040e3de97e54fa9e386d7e42d15c295c2), sob [licença MIT](https://github.com/mattpocock/skills/blob/5beb35b040e3de97e54fa9e386d7e42d15c295c2/LICENSE). Instalar uma skill só copia seus arquivos; não a executa. `setup-matt-pocock-skills` (altera projeto e labels), `implement-spec` (depende de tracker e outras skills) e `ask-matt` (depende do catálogo completo) não foram empacotadas nesta seleção. Skills que criam issues, labels, branches ou alteram projetos exigem autorização separada quando invocadas.

## Desenvolvimento e release

```sh
go test ./...
./scripts/release.sh v0.1.0
./scripts/smoke.sh ./dist/gex-dev-kit_v0.1.0_darwin_arm64
```

O script compila seis binários e gera `dist/SHA256SUMS`. Publicação manual requer `./scripts/release.sh v0.1.0 --publish` e confirmação no terminal; usa Gum se estiver instalado, sem exigir Gum de quem instala. Mudanças seguem Conventional Commits. Sem CI remoto nesta fase: teste local e evidência de release não devem ser confundidos com proteção contínua de branches.

Fontes de compatibilidade: [hooks Claude Code](https://code.claude.com/docs/en/hooks), [hooks Codex](https://developers.openai.com/codex/hooks), [instalação Claude Code](https://code.claude.com/docs/en/setup), [instalação Codex](https://developers.openai.com/codex/cli). Estado e decisões do kit: `kit-desenvolvimento-gex-acompanhamento.html` na documentação visual da equipe.
