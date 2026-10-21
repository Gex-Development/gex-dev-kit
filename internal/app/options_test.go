package app

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Gex-Development/gex-dev-kit/internal/catalog"
)

func optionsGroupByID(t *testing.T, id string) OptionGroup {
	t.Helper()
	for _, group := range InstallOptions() {
		if group.ID == id {
			return group
		}
	}
	t.Fatalf("grupo %q ausente", id)
	return OptionGroup{}
}

func TestInstallOptionsKeepsGroupOrder(t *testing.T) {
	var titles []string
	for _, group := range InstallOptions() {
		titles = append(titles, group.Title)
	}
	want := []string{
		"Engenharia", "Produtividade",
		"Especializações · Gateways", "Especializações · Dados e Plataforma", "Especializações · Produto e Interface",
		"Hooks",
	}
	if !reflect.DeepEqual(titles, want) {
		t.Fatalf("grupos %v, esperado %v", titles, want)
	}
}

func TestInstallOptionsSplitMattSkillsByGroup(t *testing.T) {
	engineering := optionsValues(optionsGroupByID(t, "engenharia"))
	productivity := optionsValues(optionsGroupByID(t, "produtividade"))
	if !optionsContains(engineering, "tdd") || optionsContains(engineering, "grill-me") {
		t.Fatalf("engenharia mal dividida: %v", engineering)
	}
	if !optionsContains(productivity, "grill-me") || optionsContains(productivity, "tdd") {
		t.Fatalf("produtividade mal dividida: %v", productivity)
	}
}

func TestInstallOptionsSpecialtyGroupsMatchCatalog(t *testing.T) {
	for _, group := range catalog.SpecialtyGroups() {
		option := optionsGroupByID(t, group)
		for _, name := range optionsValues(option) {
			specialty, ok := catalog.SpecialtyByName(name)
			if !ok || specialty.Group != group {
				t.Fatalf("%s listado em %q, grupo real %q", name, group, specialty.Group)
			}
		}
	}
}

func TestInstallOptionsValidateDependencies(t *testing.T) {
	dados := optionsGroupByID(t, "Dados e Plataforma")
	if err := dados.Validate([]string{"gex-pr-review"}); err == nil || !strings.Contains(err.Error(), "gex-local-spark-tests") {
		t.Fatalf("pr-review sem runtime local aceito: %v", err)
	}
	if err := dados.Validate([]string{"gex-pr-review", "gex-local-spark-tests"}); err != nil {
		t.Fatalf("par válido recusado: %v", err)
	}
	produtividade := optionsGroupByID(t, "produtividade")
	if err := produtividade.Validate([]string{"grill-me"}); err == nil {
		t.Fatal("grill-me sem grilling aceito")
	}
	if err := produtividade.Validate([]string{"grill-me", "grilling"}); err != nil {
		t.Fatalf("grill-me com grilling recusado: %v", err)
	}
	engenharia := optionsGroupByID(t, "engenharia")
	if err := engenharia.Validate([]string{"grill-with-docs"}); err == nil || !strings.Contains(err.Error(), "domain-modeling") {
		t.Fatalf("grill-with-docs sem domain-modeling aceito: %v", err)
	}
	if err := engenharia.Validate([]string{"grill-with-docs", "domain-modeling"}); err != nil {
		t.Fatalf("grill-with-docs com dependência recusado: %v", err)
	}
}

func TestInstallOptionHintCarriesCodexReason(t *testing.T) {
	spec, ok := catalog.HookSpecByName("gex-sensitive-action-warning")
	if !ok {
		t.Fatal("especificação de ação sensível ausente")
	}
	option := optionsHookOption(spec.Name)
	if !strings.Contains(option.Hint, spec.Codex.Reason) {
		t.Fatalf("motivo do Codex ausente na dica: %q", option.Hint)
	}
	if plain := optionsHookOption("gex-conventional-commits"); strings.Contains(plain.Hint, "Não disponível no Codex") {
		t.Fatalf("hook com suporte Codex marcado como indisponível: %q", plain.Hint)
	}
}

func optionsValues(group OptionGroup) []string {
	values := make([]string, 0, len(group.Options))
	for _, option := range group.Options {
		values = append(values, option.Value)
	}
	return values
}

func optionsContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
