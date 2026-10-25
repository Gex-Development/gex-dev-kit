package main

import (
	"fmt"
	"strings"
)

const usage = `gex-dev-kit [opções globais] [comando] [opções]

Comandos:
  install                 prévia, confirmação e gravação do Core e das escolhas
  doctor                  estado do HOME (--home) ou dos git hooks de um repositório (--repo)
  rollback                desfaz a última gravação do HOME ou do repositório
  rollback-agents         restaura versões anteriores dos agentes
  skills                  catálogo de skills
  hooks                   catálogo de hooks e referências
  addons [NOME]           catálogo de Addons; mostra o comando oficial e não o executa
  githooks                prévia dos git hooks; grava só com --repo e confirmação
  ds                      vitrine do Design System
  hook commit-msg         adaptador chamado por Claude Code, Codex e Git (uso interno)
  hook sensitive-action   adaptador chamado por Claude Code, Codex e Git (uso interno)
  hook pre-pr             adaptador chamado por Claude Code, Codex e Git (uso interno)
  hook session-start      adaptador chamado por Claude Code, Codex e Git (uso interno)
  hook pre-commit         adaptador chamado por Claude Code, Codex e Git (uso interno)
  version                 mostra a versão
  help [COMANDO]          mostra esta ajuda ou a de um comando

Opções globais (antes ou depois do comando):
  --ui=auto|tui|cli      escolhe a interface
                         auto: TUI em terminal interativo; texto em pipe, CI, --yes, --quiet ou --json
  --no-anim              desliga animações
  --quiet                remove mensagens de progresso e de sucesso
  --json                 saída JSON (skills, hooks, addons, doctor)
  -h, --help             mostra a ajuda
  -v, --version          mostra a versão

Sem comando, em um terminal interativo, abre o menu inicial.
`

type helpSpec struct {
	uso      string
	resumo   string
	opcoes   []string
	exemplos []string
}

var helpSpecs = map[string]helpSpec{
	"install": {
		uso:    "gex-dev-kit install [opções]",
		resumo: "Mostra a prévia com o diff, pede confirmação e grava o Core e as escolhas.",
		opcoes: []string{
			"--home CAMINHO        grava em outro HOME (caminho absoluto)",
			"--skill NOME          instala uma skill opcional (pode repetir)",
			"--hook NOME           instala um hook (pode repetir)",
			"--skip-agent-update   não atualiza Claude Code nem Codex",
			"--yes                 confirma depois de mostrar a prévia",
		},
		exemplos: []string{
			"gex-dev-kit install",
			"gex-dev-kit install --skill tdd --hook gex-conventional-commits",
			"gex-dev-kit install --home \"$(mktemp -d)\" --skip-agent-update --yes",
		},
	},
	"doctor": {
		uso:    "gex-dev-kit doctor [--home CAMINHO | --repo CAMINHO] [--json]",
		resumo: "Confere o estado do kit no HOME ou os git hooks de um repositório.",
		opcoes: []string{
			"--home CAMINHO   confere outro HOME (caminho absoluto)",
			"--repo CAMINHO   confere os git hooks do repositório (caminho absoluto)",
			"--json           saída JSON",
		},
		exemplos: []string{
			"gex-dev-kit doctor",
			"gex-dev-kit doctor --json",
			"gex-dev-kit doctor --repo \"$(pwd)\"",
		},
	},
	"rollback": {
		uso:    "gex-dev-kit rollback [--home CAMINHO | --repo CAMINHO] [--yes]",
		resumo: "Mostra a prévia com o diff e desfaz a última gravação do HOME ou do repositório.",
		opcoes: []string{
			"--home CAMINHO   desfaz no HOME indicado (caminho absoluto)",
			"--repo CAMINHO   desfaz nos git hooks do repositório (caminho absoluto)",
			"--yes            confirma depois de mostrar a prévia",
		},
		exemplos: []string{
			"gex-dev-kit rollback",
			"gex-dev-kit rollback --home \"$(mktemp -d)\" --yes",
			"gex-dev-kit rollback --repo \"$(pwd)\"",
		},
	},
	"rollback-agents": {
		uso:    "gex-dev-kit rollback-agents [--home CAMINHO] [--yes]",
		resumo: "Restaura as versões anteriores de Claude Code e Codex, com confirmação.",
		opcoes: []string{
			"--home CAMINHO   HOME usado no controle dos agentes (caminho absoluto)",
			"--yes            confirma sem perguntar",
		},
		exemplos: []string{
			"gex-dev-kit rollback-agents",
			"gex-dev-kit rollback-agents --yes",
			"gex-dev-kit rollback-agents --home \"$(mktemp -d)\" --yes",
		},
	},
	"skills": {
		uso:    "gex-dev-kit skills [--json]",
		resumo: "Lista as skills do Core, as opcionais e as referências que não são instaladas.",
		opcoes: []string{"--json   saída JSON"},
		exemplos: []string{
			"gex-dev-kit skills",
			"gex-dev-kit skills --json",
		},
	},
	"hooks": {
		uso:    "gex-dev-kit hooks [--json]",
		resumo: "Lista os hooks de agente e de Git e as referências.",
		opcoes: []string{"--json   saída JSON"},
		exemplos: []string{
			"gex-dev-kit hooks",
			"gex-dev-kit hooks --json",
		},
	},
	"addons": {
		uso:    "gex-dev-kit addons [NOME] [--json]",
		resumo: "Lista os Addons ou mostra o comando oficial de um deles. O comando nunca é executado.",
		opcoes: []string{"--json   saída JSON"},
		exemplos: []string{
			"gex-dev-kit addons",
			"gex-dev-kit addons caveman",
			"gex-dev-kit addons --json",
		},
	},
	"githooks": {
		uso:    "gex-dev-kit githooks --repo CAMINHO [--hook NOME] [--pre-commit-run COMANDO] [--run COMANDO] [--yes]",
		resumo: "Mostra a prévia dos git hooks do repositório. Só grava com --repo e confirmação.",
		opcoes: []string{
			"--home CAMINHO            HOME usado como referência (caminho absoluto)",
			"--repo CAMINHO            repositório Git alvo (caminho absoluto)",
			"--hook NOME               hook a instalar (pode repetir)",
			"--pre-commit-run COMANDO  comando do pre-commit, já documentado no repositório (pode repetir)",
			"--run COMANDO             comando do pre-push, já documentado no repositório (pode repetir)",
			"--yes                     confirma depois de mostrar a prévia",
		},
		exemplos: []string{
			"gex-dev-kit githooks --repo \"$(pwd)\"",
			"gex-dev-kit githooks --repo \"$(pwd)\" --hook gex-pre-commit --pre-commit-run \"go vet ./...\"",
			"gex-dev-kit githooks --repo \"$(pwd)\" --hook gex-pre-push --run \"go test ./...\" --yes",
		},
	},
	"ds": {
		uso:    "gex-dev-kit ds [--ui=tui|cli]",
		resumo: "Abre a vitrine do Design System. Com --ui=cli, mostra a listagem em texto.",
		opcoes: []string{"--ui=tui|cli   escolhe a interface"},
		exemplos: []string{
			"gex-dev-kit ds",
			"gex-dev-kit ds --ui=cli",
			"gex-dev-kit ds --ui=cli --no-anim",
		},
	},
	"hook": {
		uso:    "gex-dev-kit hook commit-msg|sensitive-action|pre-pr|session-start|pre-commit [ARGS]",
		resumo: "Adaptador chamado por Claude Code, Codex e Git. Uso interno; não é necessário rodar à mão.",
		opcoes: []string{
			"commit-msg [CAMINHO]                  valida a mensagem de commit (arquivo ou entrada do agente)",
			"sensitive-action                      avisa antes de ação sensível",
			"pre-pr                                pede confirmação antes de abrir PR",
			"session-start --agent claude|codex    monta o contexto do início de sessão",
			"pre-commit                            procura credenciais no diff staged",
		},
		exemplos: []string{
			"gex-dev-kit hook commit-msg .git/COMMIT_EDITMSG",
			"gex-dev-kit hook commit-msg < entrada.json",
		},
	},
	"version": {
		uso:    "gex-dev-kit version",
		resumo: "Mostra a versão do gex-dev-kit.",
		exemplos: []string{
			"gex-dev-kit version",
			"gex-dev-kit -v",
		},
	},
	"help": {
		uso:    "gex-dev-kit help [COMANDO]",
		resumo: "Mostra a ajuda geral ou a de um comando.",
		exemplos: []string{
			"gex-dev-kit help",
			"gex-dev-kit help install",
			"gex-dev-kit help githooks",
		},
	},
}

func commandHelp(command string) string {
	spec, ok := helpSpecs[command]
	if !ok {
		return usage
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Uso: %s\n\n%s\n", spec.uso, spec.resumo)
	if len(spec.opcoes) > 0 {
		b.WriteString("\nOpções:\n")
		for _, linha := range spec.opcoes {
			fmt.Fprintf(&b, "  %s\n", linha)
		}
	}
	b.WriteString("\nExemplos:\n")
	for _, exemplo := range spec.exemplos {
		fmt.Fprintf(&b, "  %s\n", exemplo)
	}
	return b.String()
}
