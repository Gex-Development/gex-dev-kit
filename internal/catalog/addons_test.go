package catalog

import (
	"go/parser"
	"go/token"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestAddonCatalogoTemQuatroEntradas(t *testing.T) {
	want := []string{"caveman", "ponytail", "superpowers", "spec-kit"}
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
		if addon.Kind != "estilo" && addon.Kind != "metodo" {
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
		if !sha.MatchString(addon.Revision) {
			t.Fatalf("%s: revisão deve ser um SHA de 40 caracteres hex: %s", addon.Name, addon.Revision)
		}
	}
}

func TestAddonLicencaEstaNaListaAprovada(t *testing.T) {
	for _, addon := range Addons() {
		if addon.License != "MIT" && addon.License != "Apache-2.0" {
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
