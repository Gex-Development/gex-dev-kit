package app

import (
	"strings"

	"github.com/Gex-Development/gex-dev-kit/internal/catalog"
)

func InstallOptions() []OptionGroup {
	engineering, productivity := optionsMattSkills()
	groups := []OptionGroup{
		{
			ID:          "engenharia",
			Title:       "Engenharia",
			Description: "O Core GEX é obrigatório; estas skills são opcionais.",
			Options:     engineering,
			Validate:    catalog.ValidateEngineeringSelection,
		},
		{
			ID:          "produtividade",
			Title:       "Produtividade",
			Description: "Adicione somente o que usará no trabalho.",
			Options:     productivity,
			Validate:    optionsValidateSkills,
		},
	}
	for _, group := range catalog.SpecialtyGroups() {
		groups = append(groups, OptionGroup{
			ID:          group,
			Title:       "Especializações · " + group,
			Description: "Opt-in. Nenhuma especialização é selecionada automaticamente; credenciais ficam no ambiente.",
			Options:     optionsSpecialties(group),
			Validate:    optionsValidateSkills,
		})
	}
	groups = append(groups, OptionGroup{
		ID:          "hooks",
		Title:       "Hooks",
		Description: "Codex exigirá aprovação posterior em /hooks.",
		Options:     optionsHooks(),
		Validate:    optionsValidateHooks,
	})
	return groups
}

func optionsValidateSkills(chosen []string) error {
	return catalog.ValidateSelection(chosen, nil)
}

func optionsValidateHooks(chosen []string) error {
	return catalog.ValidateSelection(nil, chosen)
}

func optionsMattSkills() ([]Option, []Option) {
	var engineering, productivity []Option
	for _, name := range catalog.Skills() {
		option := Option{Value: name, Label: name, Hint: catalog.SkillDescription(name)}
		if strings.Contains(catalog.MattSource(name), "/skills/productivity/") {
			productivity = append(productivity, option)
		} else {
			engineering = append(engineering, option)
		}
	}
	return engineering, productivity
}

func optionsSpecialties(group string) []Option {
	var options []Option
	for _, specialty := range catalog.Specialties() {
		if specialty.Group != group {
			continue
		}
		options = append(options, Option{Value: specialty.Name, Label: specialty.Name, Hint: specialty.Warning})
	}
	return options
}

func optionsHooks() []Option {
	var options []Option
	for _, name := range catalog.Hooks() {
		options = append(options, optionsHookOption(name))
	}
	return options
}

func optionsHookOption(name string) Option {
	spec, ok := catalog.HookSpecByName(name)
	if !ok {
		return Option{Value: name, Label: name}
	}
	hint := spec.Description
	if !spec.Codex.Supported {
		hint += " Não disponível no Codex: " + spec.Codex.Reason
	}
	return Option{Value: name, Label: name, Hint: hint}
}
