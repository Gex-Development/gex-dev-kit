package screens

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"charm.land/huh/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/app"
	"github.com/Gex-Development/gex-dev-kit/internal/catalog"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

type Selection struct {
	Home            string
	Skills          []string
	Hooks           []string
	SkipAgentUpdate bool
}

const (
	installDestinationCurrent  = "atual"
	installDestinationIsolated = "isolado"
	installHooksGroup          = "hooks"
)

func NewInstallForm(t theme.Theme, s *Selection, groups []app.OptionGroup) *huh.Form {
	return huh.NewForm(newInstallGroups(s, groups)...).
		WithTheme(t.Huh()).
		WithKeyMap(t.HuhKeyMap()).
		WithShowHelp(true)
}

type installAccessor[T any] struct {
	value T
	onSet func(T)
}

func (a *installAccessor[T]) Get() T {
	return a.value
}

func (a *installAccessor[T]) Set(value T) {
	a.value = value
	if a.onSet != nil {
		a.onSet(value)
	}
}

type installGroupValue struct {
	hooks  bool
	chosen []string
}

type installState struct {
	selection    *Selection
	current      string
	isolated     bool
	home         string
	update       bool
	groups       []installGroupValue
	inputField   *huh.Input
	presetSkills []string
	presetHooks  []string
}

type installHomePicker struct{ state *installState }

func (p installHomePicker) Get() string {
	return p.state.home
}

func (p installHomePicker) Set(value string) {
	p.state.home = value
	p.state.flush()
	if p.state.inputField != nil {
		p.state.inputField.Accessor(installHomeInput{p.state})
	}
}

type installHomeInput struct{ state *installState }

func (p installHomeInput) Get() string {
	return p.state.home
}

func (p installHomeInput) Set(value string) {
	p.state.home = value
	p.state.flush()
}

func newInstallGroups(s *Selection, groups []app.OptionGroup) []*huh.Group {
	st := &installState{
		selection:    s,
		current:      s.Home,
		update:       !s.SkipAgentUpdate,
		groups:       make([]installGroupValue, len(groups)),
		presetSkills: slices.Clone(s.Skills),
		presetHooks:  slices.Clone(s.Hooks),
	}
	homeGroup := huh.NewGroup(st.homePicker(), st.homeInput()).WithHideFunc(func() bool {
		return !st.isolated
	})
	result := []*huh.Group{
		huh.NewGroup(st.noteField(), st.destinationField()),
		homeGroup,
	}
	for i, group := range groups {
		result = append(result, huh.NewGroup(st.optionField(i, group)))
	}
	result = append(result, huh.NewGroup(st.updateField()))
	st.flush()
	return result
}

func (st *installState) flush() {
	st.selection.Home = st.current
	if st.isolated {
		st.selection.Home = st.home
	}
	st.selection.Skills = nil
	st.selection.Hooks = nil
	for _, group := range st.groups {
		if group.hooks {
			st.selection.Hooks = append(st.selection.Hooks, group.chosen...)
			continue
		}
		st.selection.Skills = append(st.selection.Skills, group.chosen...)
	}
	st.selection.SkipAgentUpdate = !st.update
}

func (st *installState) noteField() huh.Field {
	description := fmt.Sprintf("O Core GEX é instalado sempre: instruções e %d skills. As opções abaixo são opcionais.", len(catalog.CoreSkills()))
	return huh.NewNote().
		Title("Core obrigatório").
		Description(description).
		NextLabel("Próximo")
}

func (st *installState) destinationField() huh.Field {
	accessor := &installAccessor[string]{
		value: installDestinationCurrent,
		onSet: func(value string) {
			st.isolated = value == installDestinationIsolated
			st.flush()
		},
	}
	return huh.NewSelect[string]().
		Title("Destino da instalação").
		Description("HOME atual grava no seu perfil. HOME isolado grava numa pasta de teste.").
		Options(
			huh.NewOption("HOME atual", installDestinationCurrent),
			huh.NewOption("HOME isolado (pasta de teste)", installDestinationIsolated),
		).
		Accessor(accessor)
}

func (st *installState) homePicker() huh.Field {
	picker := huh.NewFilePicker().
		Title("Pasta do HOME isolado").
		Description("Escolha uma pasta. O kit grava os arquivos dentro dela.").
		DirAllowed(true).
		FileAllowed(false).
		Accessor(installHomePicker{st}).
		Validate(st.validateHome)
	if dir, err := os.Getwd(); err == nil {
		picker = picker.CurrentDirectory(dir)
	}
	return picker
}

func (st *installState) homeInput() huh.Field {
	input := huh.NewInput().
		Title("Caminho do HOME isolado").
		Description("Ou digite o caminho absoluto da pasta.").
		Placeholder("/caminho/absoluto").
		Accessor(installHomeInput{st}).
		Validate(st.validateHome)
	st.inputField = input
	return input
}

func (st *installState) validateHome(value string) error {
	if !st.isolated {
		return nil
	}
	return installAbsolutePath(value)
}

func installAbsolutePath(value string) error {
	if strings.TrimSpace(value) == "" {
		return errors.New("informe uma pasta para o HOME isolado")
	}
	if !filepath.IsAbs(value) {
		return errors.New("use caminho absoluto, começando na raiz do sistema ou na unidade (Windows)")
	}
	return nil
}

func (st *installState) optionField(index int, group app.OptionGroup) huh.Field {
	hooks := group.ID == installHooksGroup
	preset := st.presetSkills
	if hooks {
		preset = st.presetHooks
	}
	chosen := installPreselected(preset, group)
	st.groups[index] = installGroupValue{hooks: hooks, chosen: chosen}
	accessor := &installAccessor[[]string]{
		value: chosen,
		onSet: func(value []string) {
			st.groups[index].chosen = slices.Clone(value)
			st.flush()
		},
	}
	options := make([]huh.Option[string], 0, len(group.Options))
	for _, option := range group.Options {
		options = append(options, huh.NewOption(installOptionLabel(option), option.Value))
	}
	field := huh.NewMultiSelect[string]().
		Title(group.Title).
		Description(group.Description).
		Options(options...).
		Filterable(true).
		Accessor(accessor)
	if group.Validate != nil {
		field = field.Validate(group.Validate)
	}
	return field
}

func installPreselected(preset []string, group app.OptionGroup) []string {
	var chosen []string
	for _, option := range group.Options {
		if slices.Contains(preset, option.Value) {
			chosen = append(chosen, option.Value)
		}
	}
	return chosen
}

func installOptionLabel(option app.Option) string {
	if option.Hint == "" {
		return option.Label
	}
	return option.Label + " · " + option.Hint
}

func (st *installState) updateField() huh.Field {
	accessor := &installAccessor[bool]{
		value: st.update,
		onSet: func(value bool) {
			st.update = value
			st.flush()
		},
	}
	return huh.NewConfirm().
		Title("Atualizar agentes instalados?").
		Description("Atualiza Claude Code e Codex para as versões fixadas pelo kit.").
		Affirmative("Atualizar").
		Negative("Agora não").
		Accessor(accessor)
}
