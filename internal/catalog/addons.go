package catalog

import "strings"

type AddonStep struct {
	Argv []string
}

type AddonInstall struct {
	Steps  []AddonStep
	Manual string
}

type Addon struct {
	Name, Kind, Summary          string
	Source, License              string
	Revision, Release, CheckedAt string
	InstallClaude, InstallCodex  string
	Pinned                       string
	Install                      map[string]AddonInstall
	Deterministic                bool
	Installs, Conflicts          []string
	CoreCompat                   string
}

func Addons() []Addon {
	const (
		cavemanRevision  = "bad48a0e1df55e05e3f52249afa35ecc1909ee76"
		cavemanRelease   = "v4.0.1"
		ponytailRevision = "09c434ab73ca43761c5d8b0c5d341926c2fba710"
		ponytailRelease  = "v5.3.2"
		specKitRevision  = "dc2cec4f568d0f6e2cdfb8571c69116bc55c1aaf"
		specKitRelease   = "v1.2.2"
	)
	return []Addon{
		{
			Name:          "caveman",
			Kind:          "estilo",
			Summary:       "Respostas curtas que mantêm o conteúdo técnico.",
			Source:        "https://github.com/JuliusBrussee/caveman",
			License:       "Apache-2.0",
			Revision:      cavemanRevision,
			Release:       cavemanRelease,
			CheckedAt:     "2026-10-21",
			InstallClaude: "claude plugin marketplace add JuliusBrussee/caveman && claude plugin install caveman@caveman",
			InstallCodex:  "npx -y github:JuliusBrussee/caveman -- --only codex",
			Pinned:        "npx e Codex pelo commit conferido; Claude Code pela tag v4.0.1 porque a CLI não aceita commit.",
			Install: map[string]AddonInstall{
				"claude": {Steps: []AddonStep{
					{Argv: []string{"claude", "plugin", "marketplace", "add", "JuliusBrussee/caveman#" + cavemanRelease}},
					{Argv: []string{"claude", "plugin", "install", "caveman@caveman"}},
				}},
				"codex": {Steps: []AddonStep{{Argv: []string{"npx", "-y", "github:JuliusBrussee/caveman#" + cavemanRevision, "--", "--only", "codex"}}}},
			},
			Deterministic: false,
			Installs: []string{
				"No Claude Code: hooks SessionStart, SubagentStart, UserPromptSubmit e SessionEnd.",
				"Skills caveman*, como caveman-commit e caveman-review.",
				"Entradas de hook gravadas no settings.json do Claude Code.",
				"No Codex: hook SessionStart. Aprove o hook em /hooks.",
				"A CLI @caveman-ai/cli envia estatísticas de uso por padrão. O instalador não envia telemetria.",
			},
			Conflicts: []string{
				"Muda o estilo de todas as respostas. A Core pede texto sem densidade.",
				"caveman-commit e caveman-review sobrepõem as regras de commit e de revisão.",
				"Mexe no settings.json, arquivo que o merge de hooks do kit também altera.",
			},
			CoreCompat: "Opt-in. Desligue se conflitar com a escrita GEX. Não use a CLI nem o proxy sem revisão.",
		},
		{
			Name:      "ponytail",
			Kind:      "estilo",
			Summary:   "Menor mudança completa; evita código desnecessário.",
			Source:    "https://github.com/DietrichGebert/ponytail",
			License:   "MIT",
			Revision:  ponytailRevision,
			Release:   ponytailRelease,
			CheckedAt: "2026-10-21",
			InstallClaude: `/plugin marketplace add DietrichGebert/ponytail
/plugin install ponytail@ponytail`,
			InstallCodex: `codex plugin marketplace add DietrichGebert/ponytail
codex plugin add ponytail@ponytail`,
			Pinned: "Codex pelo commit conferido; Claude Code pela tag v5.3.2 porque a CLI não aceita commit.",
			Install: map[string]AddonInstall{
				"claude": {Steps: []AddonStep{
					{Argv: []string{"claude", "plugin", "marketplace", "add", "DietrichGebert/ponytail#" + ponytailRelease}},
					{Argv: []string{"claude", "plugin", "install", "ponytail@ponytail"}},
				}},
				"codex": {Steps: []AddonStep{
					{Argv: []string{"codex", "plugin", "marketplace", "add", "DietrichGebert/ponytail@" + ponytailRevision}},
					{Argv: []string{"codex", "plugin", "add", "ponytail@ponytail"}},
				}},
			},
			Deterministic: false,
			Installs: []string{
				"No Claude Code: hooks SessionStart, SubagentStart e UserPromptSubmit, e o painel band acima do prompt.",
				"6 skills: ponytail, ponytail-audit, ponytail-debt, ponytail-gain, ponytail-help e ponytail-review.",
				"Dados do painel local ficam em ~/.config/ponytail/band/.",
				"No Codex: dois hooks de ciclo de vida. Abra /hooks e confie neles.",
			},
			Conflicts: []string{
				"Escreve comentários shortcut: no código. A Core proíbe comentários em código novo.",
			},
			CoreCompat: "Conflito direto com a regra de comentários da Core. Use só com essa ressalva.",
		},
		{
			Name:          "superpowers",
			Kind:          "metodo",
			Summary:       "Brainstorming, TDD, diagnóstico, planos, revisão e verificação.",
			Source:        "https://github.com/obra/superpowers",
			License:       "MIT",
			Revision:      "2ce8aa15d05bd57213e6f746f953aafb7ded7a8b",
			Release:       "v7.0.1",
			CheckedAt:     "2026-10-21",
			InstallClaude: "/plugin install superpowers@claude-plugins-official",
			InstallCodex:  "/plugins → superpowers → Install Plugin",
			Pinned:        "Marketplace oficial fixa o commit bb92a77741419a4ab5f06e711a283343f1ada0c3 (tag v7.0.0). Muda quando o marketplace atualiza.",
			Install: map[string]AddonInstall{
				"claude": {Steps: []AddonStep{{Argv: []string{"claude", "plugin", "install", "superpowers@claude-plugins-official"}}}},
				"codex":  {Manual: "No Codex: abra /plugins, escolha superpowers e Install Plugin (menu interativo, o kit não consegue automatizar)."},
			},
			Deterministic: false,
			Installs: []string{
				"No Claude Code: hook SessionStart que carrega o bootstrap das skills.",
				"15 skills, como brainstorming, test-driven-development e systematic-debugging.",
				"Grava docs/superpowers/ (specs de design) e .worktrees/ no projeto.",
				"Ao terminar um trabalho, pede a escolha do usuário antes de qualquer push ou PR.",
			},
			Conflicts: []string{
				"TDD e brainstorming sobrepõem a skill tdd da Matt.",
				"A skill de TDD manda commitar a cada ciclo.",
			},
			CoreCompat: "Método opcional. Push e PR continuam exigindo autorização.",
		},
		{
			Name:      "spec-kit",
			Kind:      "metodo",
			Summary:   "Especificação formal: constitution, specify, plan, tasks, implement.",
			Source:    "https://github.com/github/spec-kit",
			License:   "MIT",
			Revision:  specKitRevision,
			Release:   specKitRelease,
			CheckedAt: "2026-10-21",
			InstallClaude: `uv tool install specify-cli
specify init my-project --integration claude`,
			InstallCodex: `uv tool install specify-cli
specify init my-project --integration codex`,
			Pinned: "uv instala o commit conferido; a tag v1.2.2 tem histórico divergente.",
			Install: map[string]AddonInstall{
				"claude": {Steps: []AddonStep{{Argv: []string{"uv", "tool", "install", "specify-cli", "--from", "git+https://github.com/github/spec-kit.git@" + specKitRevision}}}, Manual: "Em cada projeto, rode specify init . --integration claude (ou codex)."},
				"codex":  {Steps: []AddonStep{{Argv: []string{"uv", "tool", "install", "specify-cli", "--from", "git+https://github.com/github/spec-kit.git@" + specKitRevision}}}, Manual: "Em cada projeto, rode specify init . --integration claude (ou codex)."},
			},
			Deterministic: true,
			Installs: []string{
				"Exige Python 3.11 ou mais novo e uv.",
				"Cria .specify/ no projeto e os comandos /speckit-*.",
				"Extensões podem reagir a eventos do ciclo de vida do agente.",
				"Em projeto já existente: specify integration install claude (ou codex).",
			},
			Conflicts: []string{
				"Cria arquivos em cada repositório onde é iniciado.",
				"Os comandos speckit-* disputam contexto com as skills gex-*.",
				"A constitution pode repetir ou contrariar regras GEX. Revise antes de usar.",
			},
			CoreCompat: "O kit instala a CLI; specify init por projeto é manual e a decisão é do time.",
		},
	}
}

func AddonByName(name string) (Addon, bool) {
	for _, addon := range Addons() {
		if addon.Name == name {
			return addon, true
		}
	}
	return Addon{}, false
}

func (a Addon) DisplayCommands(harness string) []string {
	var commands []string
	for _, step := range a.Install[harness].Steps {
		commands = append(commands, strings.Join(step.Argv, " "))
	}
	return commands
}
