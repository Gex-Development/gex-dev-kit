package catalog

import (
	"go/parser"
	"go/token"
	"net/url"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestAddonCatalogoTemOitoEntradas(t *testing.T) {
	want := []string{"caveman", "ponytail", "superpowers", "spec-kit", "context7", "terraform", "brave-search", "mapbox-devkit"}
	var got []string
	for _, addon := range Addons() {
		got = append(got, addon.Name)
	}
	if !slices.Equal(got, want) {
		t.Fatalf("addons = %v, esperado %v", got, want)
	}
}

func TestAddonCamposObrigatoriosPreenchidos(t *testing.T) {
	for _, addon := range Addons() {
		campos := map[string]string{
			"Name":          addon.Name,
			"Kind":          addon.Kind,
			"Summary":       addon.Summary,
			"Source":        addon.Source,
			"License":       addon.License,
			"Revision":      addon.Revision,
			"Release":       addon.Release,
			"CheckedAt":     addon.CheckedAt,
			"InstallClaude": addon.InstallClaude,
			"InstallCodex":  addon.InstallCodex,
			"Pinned":        addon.Pinned,
			"CoreCompat":    addon.CoreCompat,
		}
		for campo, valor := range campos {
			if strings.TrimSpace(valor) == "" {
				t.Fatalf("%s: campo %s vazio", addon.Name, campo)
			}
		}
		if len(addon.Installs) == 0 || len(addon.Conflicts) == 0 {
			t.Fatalf("%s: Installs e Conflicts precisam ter ao menos um item", addon.Name)
		}
		for _, item := range append(slices.Clone(addon.Installs), addon.Conflicts...) {
			if strings.TrimSpace(item) == "" {
				t.Fatalf("%s: item vazio em Installs ou Conflicts", addon.Name)
			}
		}
		if addon.Kind != "estilo" && addon.Kind != "metodo" && addon.Kind != "mcp" {
			t.Fatalf("%s: Kind inválido: %s", addon.Name, addon.Kind)
		}
	}
}

func TestAddonURLsDeFonteSaoHTTPS(t *testing.T) {
	for _, addon := range Addons() {
		u, err := url.Parse(addon.Source)
		if err != nil {
			t.Fatalf("%s: fonte inválida: %v", addon.Name, err)
		}
		if u.Scheme != "https" || u.Host == "" {
			t.Fatalf("%s: fonte deve ser https com host: %s", addon.Name, addon.Source)
		}
	}
}

func TestAddonRevisaoEhSHADeQuarentaHex(t *testing.T) {
	sha := regexp.MustCompile(`^[0-9a-f]{40}$`)
	for _, addon := range Addons() {
		if addon.Kind == "mcp" && addon.Revision == "" && addon.Pinned != "" {
			continue
		}
		if !sha.MatchString(addon.Revision) {
			t.Fatalf("%s: revisão deve ser um SHA de 40 caracteres hex: %s", addon.Name, addon.Revision)
		}
	}
}

func TestAddonLicencaEstaNaListaAprovada(t *testing.T) {
	for _, addon := range Addons() {
		if addon.License != "MIT" && addon.License != "Apache-2.0" && addon.License != "GPL-3.0-or-later" && addon.License != "MPL-2.0" {
			t.Fatalf("%s: licença fora da lista aprovada: %s", addon.Name, addon.License)
		}
	}
}

func TestAddonDataDeConferenciaEhData(t *testing.T) {
	for _, addon := range Addons() {
		if _, err := time.Parse("2006-01-02", addon.CheckedAt); err != nil {
			t.Fatalf("%s: CheckedAt deve estar no formato AAAA-MM-DD: %v", addon.Name, err)
		}
	}
}

func TestAddonByNameEncontraEAusenteRetornaFalso(t *testing.T) {
	addon, ok := AddonByName("ponytail")
	if !ok || addon.Name != "ponytail" {
		t.Fatalf("AddonByName(ponytail) = %q, %v", addon.Name, ok)
	}
	if _, ok := AddonByName("gsd"); ok {
		t.Fatal("GSD não deve estar no catálogo de addons")
	}
}

func TestAddonArquivoNaoImportaExecucaoDeComando(t *testing.T) {
	arquivo, err := parser.ParseFile(token.NewFileSet(), "addons.go", nil, parser.ImportsOnly)
	if err != nil {
		t.Fatal(err)
	}
	for _, imp := range arquivo.Imports {
		caminho, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			t.Fatal(err)
		}
		if caminho == "os/exec" || caminho == "syscall" {
			t.Fatalf("addons.go importa %s; o catálogo guarda só texto e não executa comandos", caminho)
		}
	}
}

func TestAddonInstallStepsAndDisplay(t *testing.T) {
	want := map[string]map[string][][]string{
		"caveman": {
			"claude": {{"claude", "plugin", "marketplace", "add", "JuliusBrussee/caveman#v4.0.1"}, {"claude", "plugin", "install", "caveman@caveman"}},
			"codex":  {{"npx", "-y", "github:JuliusBrussee/caveman#bad48a0e1df55e05e3f52249afa35ecc1909ee76", "--", "--only", "codex"}},
		},
		"ponytail": {
			"claude": {{"claude", "plugin", "marketplace", "add", "DietrichGebert/ponytail#v5.3.2"}, {"claude", "plugin", "install", "ponytail@ponytail"}},
			"codex":  {{"codex", "plugin", "marketplace", "add", "DietrichGebert/ponytail@09c434ab73ca43761c5d8b0c5d341926c2fba710"}, {"codex", "plugin", "add", "ponytail@ponytail"}},
		},
		"superpowers": {
			"claude": {{"claude", "plugin", "install", "superpowers@claude-plugins-official"}},
			"codex":  {},
		},
		"spec-kit": {
			"claude": {{"uv", "tool", "install", "specify-cli", "--from", "git+https://github.com/github/spec-kit.git@dc2cec4f568d0f6e2cdfb8571c69116bc55c1aaf"}},
			"codex":  {{"uv", "tool", "install", "specify-cli", "--from", "git+https://github.com/github/spec-kit.git@dc2cec4f568d0f6e2cdfb8571c69116bc55c1aaf"}},
		},
		"context7": {
			"claude": {{"claude", "mcp", "add", "--scope", "user", "context7", "--", "npx", "-y", "@upstash/context7-mcp@4.4.4"}},
			"codex":  {{"codex", "mcp", "add", "context7", "--", "npx", "-y", "@upstash/context7-mcp@4.4.4"}},
		},
		"terraform": {
			"claude": {{"docker", "pull", "hashicorp/terraform-mcp-server:1.4.1@sha256:11dd418bba6003dd5777e9d0979b68de59929a155b355f338f407947e00e9d80"}, {"claude", "mcp", "add", "--scope", "user", "terraform", "--", "docker", "run", "-i", "--rm", "hashicorp/terraform-mcp-server:1.4.1@sha256:11dd418bba6003dd5777e9d0979b68de59929a155b355f338f407947e00e9d80"}},
			"codex":  {{"docker", "pull", "hashicorp/terraform-mcp-server:1.4.1@sha256:11dd418bba6003dd5777e9d0979b68de59929a155b355f338f407947e00e9d80"}, {"codex", "mcp", "add", "terraform", "--", "docker", "run", "-i", "--rm", "hashicorp/terraform-mcp-server:1.4.1@sha256:11dd418bba6003dd5777e9d0979b68de59929a155b355f338f407947e00e9d80"}},
		},
		"brave-search": {
			"claude": {{"claude", "mcp", "remove", "--scope", "user", "brave-search"}, {"claude", "mcp", "add", "--scope", "user", "brave-search", "-e", "BRAVE_API_KEY={{valor}}", "--", "npx", "-y", "@brave/brave-search-mcp-server@2.1.4", "--transport", "stdio"}},
			"codex":  {{"codex", "mcp", "add", "brave-search", "--env", "BRAVE_API_KEY={{valor}}", "--", "npx", "-y", "@brave/brave-search-mcp-server@2.1.4", "--transport", "stdio"}},
		},
		"mapbox-devkit": {
			"claude": {{"claude", "mcp", "remove", "--scope", "user", "mapbox-devkit"}, {"claude", "mcp", "add", "--scope", "user", "mapbox-devkit", "-e", "MAPBOX_ACCESS_TOKEN={{valor}}", "--", "npx", "-y", "@mapbox/mcp-devkit-server@0.8.2"}},
			"codex":  {{"codex", "mcp", "add", "mapbox-devkit", "--env", "MAPBOX_ACCESS_TOKEN={{valor}}", "--", "npx", "-y", "@mapbox/mcp-devkit-server@0.8.2"}},
		},
	}
	for _, addon := range Addons() {
		for _, harness := range []string{"claude", "codex"} {
			install, ok := addon.Install[harness]
			if !ok || len(install.Steps) == 0 && install.Manual == "" {
				t.Fatalf("%s/%s sem instalação", addon.Name, harness)
			}
			var got [][]string
			for _, step := range install.Steps {
				got = append(got, step.Argv)
				for _, arg := range step.Argv {
					if strings.ContainsAny(arg, ";&|") {
						t.Fatalf("%s/%s: argumento de shell: %q", addon.Name, harness, arg)
					}
					if strings.HasPrefix(arg, "github:") || strings.HasPrefix(arg, "git+") {
						if !strings.HasSuffix(arg, addon.Revision) {
							t.Fatalf("%s/%s: referência não fixa revisão conferida: %q", addon.Name, harness, arg)
						}
					}
					if len(step.Argv) >= 4 && step.Argv[1] == "plugin" && step.Argv[2] == "marketplace" && step.Argv[3] == "add" {
						pin := "@" + addon.Revision
						if harness == "claude" {
							pin = "#" + addon.Release
						}
						if !strings.HasSuffix(arg, pin) && arg == step.Argv[len(step.Argv)-1] {
							t.Fatalf("%s/%s: marketplace não fixado: %q", addon.Name, harness, arg)
						}
					}
				}
			}
			if !reflect.DeepEqual(got, want[addon.Name][harness]) {
				if len(got) != 0 || len(want[addon.Name][harness]) != 0 {
					t.Fatalf("%s/%s: argv = %v; quer %v", addon.Name, harness, got, want[addon.Name][harness])
				}
			}
			var display []string
			for _, argv := range want[addon.Name][harness] {
				display = append(display, strings.ReplaceAll(strings.Join(argv, " "), SecretPlaceholder, "<sua chave>"))
			}
			if !slices.Equal(addon.DisplayCommands(harness), display) {
				t.Fatalf("%s/%s: display = %v; quer %v", addon.Name, harness, addon.DisplayCommands(harness), display)
			}
		}
	}
}

func TestAddonDeterministicoNaoDependeDeTag(t *testing.T) {
	for _, addon := range Addons() {
		for harness, install := range addon.Install {
			for _, step := range install.Steps {
				for _, arg := range step.Argv {
					if addon.Deterministic && strings.HasSuffix(arg, "#"+addon.Release) {
						t.Fatalf("%s/%s é marcado como determinístico mas fixa a tag mutável %s", addon.Name, harness, addon.Release)
					}
				}
			}
		}
	}
}

func TestMCPAddonMetadataAndSteps(t *testing.T) {
	addon, ok := AddonByName("context7")
	if !ok || addon.Kind != "mcp" || addon.CheckedAt != "2026-10-25" {
		t.Fatalf("context7: metadados = %+v", addon)
	}
	for _, harness := range []string{"claude", "codex"} {
		steps := addon.Install[harness].Steps
		if len(steps) != 1 {
			t.Fatalf("%s: passos = %+v", harness, steps)
		}
		if harness == "claude" && steps[0].AlreadyOK != "already exists" {
			t.Fatalf("%s: resposta existente = %+v", harness, steps)
		}
	}
}

func TestMCPAddonsRequireCorrectRuntimeAndProtectSecrets(t *testing.T) {
	for _, tc := range []struct {
		name, runtime, note, env, url string
	}{
		{"terraform", "docker", "Instale e ligue o Docker e rode de novo.", "", ""},
		{"brave-search", "npx", "Instale Node.js (npx) e rode de novo.", "BRAVE_API_KEY", "https://api-dashboard.search.brave.com/app/keys"},
		{"mapbox-devkit", "npx", "Instale Node.js 22 ou mais novo (npx) e rode de novo.", "MAPBOX_ACCESS_TOKEN", "https://account.mapbox.com/"},
	} {
		addon, ok := AddonByName(tc.name)
		if !ok || addon.Kind != "mcp" || addon.CheckedAt != "2026-10-25" || !slices.Equal(addon.Requires, []string{tc.runtime}) || addon.RequiresNote != tc.note {
			t.Fatalf("%s: requisito inválido: %+v", tc.name, addon)
		}
		if tc.env == "" {
			if len(addon.Secrets) != 0 || !addon.Install["claude"].Steps[0].Prerequisite || !addon.Install["codex"].Steps[0].Prerequisite || addon.Install["claude"].Steps[1].AlreadyOK != "already exists" {
				t.Fatalf("terraform: passos ou chave inválidos: %+v", addon)
			}
			continue
		}
		if len(addon.Secrets) != 1 || addon.Secrets[0].Env != tc.env || addon.Secrets[0].URL != tc.url || addon.Secrets[0].Label == "" || addon.Install["claude"].Steps[0].AlreadyOK != "No MCP server named" {
			t.Fatalf("%s: segredo ou remoção inválidos: %+v", tc.name, addon)
		}
		for _, command := range append(addon.DisplayCommands("claude"), addon.DisplayCommands("codex")...) {
			if strings.Contains(command, SecretPlaceholder) || strings.Contains(command, "fake-key-123") {
				t.Fatalf("%s: comando exibido inseguro: %s", tc.name, command)
			}
		}
	}
}
