package catalog

import "strings"

const SecretPlaceholder = "{{valor}}"

type AddonSecret struct{ Env, Label, URL string }

type AddonStep struct {
	Argv         []string
	Prerequisite bool
	AlreadyOK    string
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
	Requires                     []string
	RequiresNote                 string
	Secrets                      []AddonSecret
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
		{
			Name: "context7", Kind: "mcp",
			Summary: "Documentação atualizada de bibliotecas (Spring, Hibernate, JUnit e outras) direto no contexto do agente.",
			Source:  "https://github.com/upstash/context7", License: "MIT",
			Revision: "c7f80ce5af1ff1403d167579b76784482cf7f3e2", Release: "4.4.4", CheckedAt: "2026-10-25",
			InstallClaude: "claude mcp add --scope user context7 -- npx -y @upstash/context7-mcp@4.4.4",
			InstallCodex:  "codex mcp add context7 -- npx -y @upstash/context7-mcp@4.4.4",
			Pinned:        "Pacote npm @upstash/context7-mcp@4.4.4; tag @upstash/context7-mcp@4.4.4 conferida no GitHub (integridade sha512-T2DkjhweDnC2mRAclDiTrPG3KI1zYKhhjgY2SMKD9PVrbEaGD8x35VsIOqpwwDc2qiKBvfYCF/jqr+jQMu5PUA==).",
			Requires:      []string{"npx"},
			RequiresNote:  "Instale Node.js (npx) e rode de novo.",
			Install: map[string]AddonInstall{
				"claude": {Steps: []AddonStep{{Argv: []string{"claude", "mcp", "add", "--scope", "user", "context7", "--", "npx", "-y", "@upstash/context7-mcp@4.4.4"}, AlreadyOK: "already exists"}}},
				"codex":  {Steps: []AddonStep{{Argv: []string{"codex", "mcp", "add", "context7", "--", "npx", "-y", "@upstash/context7-mcp@4.4.4"}}}},
			},
			Installs: []string{
				"Exige Node.js com npx; o npx baixa o pacote na primeira execução.",
				"Funciona sem chave de API, com limite anônimo.",
				"Envia o nome da biblioteca e a pergunta ao serviço Context7.",
				"Registro global (escopo de usuário) para Claude Code e Codex.",
			},
			Conflicts: []string{
				"Envia perguntas e nomes de bibliotecas ao serviço Context7 (Upstash).",
				"O rollback do kit não remove MCP: rode claude mcp remove --scope user context7 e codex mcp remove context7.",
			},
			CoreCompat: "Extra opt-in. Não altera o Core nem grava chaves de API.",
		},
		{
			Name: "terraform", Kind: "mcp",
			Summary: "Consulta documentação e recursos Terraform pelo servidor oficial.",
			Source:  "https://github.com/hashicorp/terraform-mcp-server", License: "MPL-2.0",
			Revision: "3781509a674482ca7e2821fab8159a25e3b29bdf", Release: "1.4.1", CheckedAt: "2026-10-25",
			InstallClaude: "docker pull hashicorp/terraform-mcp-server:1.4.1@sha256:11dd418bba6003dd5777e9d0979b68de59929a155b355f338f407947e00e9d80\nclaude mcp add --scope user terraform -- docker run -i --rm hashicorp/terraform-mcp-server:1.4.1@sha256:11dd418bba6003dd5777e9d0979b68de59929a155b355f338f407947e00e9d80",
			InstallCodex:  "docker pull hashicorp/terraform-mcp-server:1.4.1@sha256:11dd418bba6003dd5777e9d0979b68de59929a155b355f338f407947e00e9d80\ncodex mcp add terraform -- docker run -i --rm hashicorp/terraform-mcp-server:1.4.1@sha256:11dd418bba6003dd5777e9d0979b68de59929a155b355f338f407947e00e9d80",
			Pinned:        "Imagem oficial 1.4.1 fixada pelo digest sha256:11dd418bba6003dd5777e9d0979b68de59929a155b355f338f407947e00e9d80.",
			Requires:      []string{"docker"}, RequiresNote: "Instale e ligue o Docker e rode de novo.",
			Install: map[string]AddonInstall{
				"claude": {Steps: []AddonStep{
					{Argv: []string{"docker", "pull", "hashicorp/terraform-mcp-server:1.4.1@sha256:11dd418bba6003dd5777e9d0979b68de59929a155b355f338f407947e00e9d80"}, Prerequisite: true},
					{Argv: []string{"claude", "mcp", "add", "--scope", "user", "terraform", "--", "docker", "run", "-i", "--rm", "hashicorp/terraform-mcp-server:1.4.1@sha256:11dd418bba6003dd5777e9d0979b68de59929a155b355f338f407947e00e9d80"}, AlreadyOK: "already exists"},
				}},
				"codex": {Steps: []AddonStep{
					{Argv: []string{"docker", "pull", "hashicorp/terraform-mcp-server:1.4.1@sha256:11dd418bba6003dd5777e9d0979b68de59929a155b355f338f407947e00e9d80"}, Prerequisite: true},
					{Argv: []string{"codex", "mcp", "add", "terraform", "--", "docker", "run", "-i", "--rm", "hashicorp/terraform-mcp-server:1.4.1@sha256:11dd418bba6003dd5777e9d0979b68de59929a155b355f338f407947e00e9d80"}},
				}},
			},
			Installs:   []string{"Roda em contêiner Docker com imagem oficial fixada por digest; Docker precisa estar ligado.", "Consulta o Registry público; sem token HCP Terraform."},
			Conflicts:  []string{"Pode expor dados do Terraform ao agente; use só com clientes e modelos confiáveis.", "O rollback do kit não remove MCP: rode claude mcp remove --scope user terraform e codex mcp remove terraform."},
			CoreCompat: "Extra opt-in. Não altera o Core.",
		},
		{
			Name: "brave-search", Kind: "mcp",
			Summary: "Pesquisa na web com a API Brave Search.",
			Source:  "https://github.com/brave/brave-search-mcp-server", License: "MIT",
			Revision: "8bcb302657b0f072612c2f3e53182ff17912013c", Release: "2.1.4", CheckedAt: "2026-10-25",
			InstallClaude: "claude mcp remove --scope user brave-search\nclaude mcp add --scope user brave-search -e BRAVE_API_KEY=<sua chave> -- npx -y @brave/brave-search-mcp-server@2.1.4 --transport stdio",
			InstallCodex:  "codex mcp add brave-search --env BRAVE_API_KEY=<sua chave> -- npx -y @brave/brave-search-mcp-server@2.1.4 --transport stdio",
			Pinned:        "Pacote npm @brave/brave-search-mcp-server@2.1.4; tag v2.1.4 conferida no GitHub.",
			Requires:      []string{"npx"}, RequiresNote: "Instale Node.js (npx) e rode de novo.",
			Secrets: []AddonSecret{{Env: "BRAVE_API_KEY", Label: "Chave Brave Search", URL: "https://api-dashboard.search.brave.com/app/keys"}},
			Install: map[string]AddonInstall{
				"claude": {Steps: []AddonStep{
					{Argv: []string{"claude", "mcp", "remove", "--scope", "user", "brave-search"}, AlreadyOK: "No MCP server named"},
					{Argv: []string{"claude", "mcp", "add", "--scope", "user", "brave-search", "-e", "BRAVE_API_KEY={{valor}}", "--", "npx", "-y", "@brave/brave-search-mcp-server@2.1.4", "--transport", "stdio"}},
				}},
				"codex": {Steps: []AddonStep{{Argv: []string{"codex", "mcp", "add", "brave-search", "--env", "BRAVE_API_KEY={{valor}}", "--", "npx", "-y", "@brave/brave-search-mcp-server@2.1.4", "--transport", "stdio"}}}},
			},
			Installs:   []string{"Envia buscas à API do Brave.", "Chave em TEXTO PURO em ~/.claude.json e ~/.codex/config.toml (arquivos 0600); aparece na lista de processos durante o registro.", "Para trocar a chave, reinstale."},
			Conflicts:  []string{"O rollback do kit não remove MCP: rode claude mcp remove --scope user brave-search e codex mcp remove brave-search; revogue a chave no painel do Brave."},
			CoreCompat: "Extra opt-in. Chave fornecida pelo usuário.",
		},
		{
			Name: "mapbox-devkit", Kind: "mcp",
			Summary: "Acessa ferramentas de mapas e recursos da conta Mapbox.",
			Source:  "https://github.com/mapbox/mcp-devkit-server", License: "MIT",
			Revision: "502932367434321ce0cb99ab91a9ecfa9e0214ac", Release: "0.8.2", CheckedAt: "2026-10-25",
			InstallClaude: "claude mcp remove --scope user mapbox-devkit\nclaude mcp add --scope user mapbox-devkit -e MAPBOX_ACCESS_TOKEN=<sua chave> -- npx -y @mapbox/mcp-devkit-server@0.8.2",
			InstallCodex:  "codex mcp add mapbox-devkit --env MAPBOX_ACCESS_TOKEN=<sua chave> -- npx -y @mapbox/mcp-devkit-server@0.8.2",
			Pinned:        "Pacote npm @mapbox/mcp-devkit-server@0.8.2; tag v0.8.2 conferida no GitHub.",
			Requires:      []string{"npx"}, RequiresNote: "Instale Node.js 22 ou mais novo (npx) e rode de novo.",
			Secrets: []AddonSecret{{Env: "MAPBOX_ACCESS_TOKEN", Label: "Token Mapbox", URL: "https://account.mapbox.com/"}},
			Install: map[string]AddonInstall{
				"claude": {Steps: []AddonStep{
					{Argv: []string{"claude", "mcp", "remove", "--scope", "user", "mapbox-devkit"}, AlreadyOK: "No MCP server named"},
					{Argv: []string{"claude", "mcp", "add", "--scope", "user", "mapbox-devkit", "-e", "MAPBOX_ACCESS_TOKEN={{valor}}", "--", "npx", "-y", "@mapbox/mcp-devkit-server@0.8.2"}},
				}},
				"codex": {Steps: []AddonStep{{Argv: []string{"codex", "mcp", "add", "mapbox-devkit", "--env", "MAPBOX_ACCESS_TOKEN={{valor}}", "--", "npx", "-y", "@mapbox/mcp-devkit-server@0.8.2"}}}},
			},
			Installs:   []string{"Exige Node.js 22 ou mais novo; ferramentas usam a conta Mapbox.", "Token em TEXTO PURO em ~/.claude.json e ~/.codex/config.toml (arquivos 0600); aparece na lista de processos durante o registro.", "Para trocar o token, reinstale."},
			Conflicts:  []string{"O token define permissões: com escopos de escrita o agente cria e apaga estilos e tokens da conta; use o mínimo de escopos.", "O rollback do kit não remove MCP: rode claude mcp remove --scope user mapbox-devkit e codex mcp remove mapbox-devkit; revogue o token no painel Mapbox."},
			CoreCompat: "Extra opt-in. Token fornecido pelo usuário.",
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
		commands = append(commands, strings.ReplaceAll(strings.Join(step.Argv, " "), SecretPlaceholder, "<sua chave>"))
	}
	return commands
}
