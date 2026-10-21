package screens

import (
	"slices"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/app"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

func installFormTestTheme() theme.Theme {
	return theme.New(theme.Options{})
}

func installFormGrupoProdutividade(t *testing.T) app.OptionGroup {
	t.Helper()
	for _, grupo := range app.InstallOptions() {
		if grupo.ID == "produtividade" {
			return grupo
		}
	}
	t.Fatalf("grupo produtividade não encontrado em InstallOptions")
	return app.OptionGroup{}
}

func TestInstallFormDevolveFormCom3FixosMaisUmGrupoPorOpcao(t *testing.T) {
	grupos := app.InstallOptions()
	var sel Selection
	form := NewInstallForm(installFormTestTheme(), &sel, grupos)
	if form == nil {
		t.Fatalf("NewInstallForm devolveu nil")
	}
	if n := len(newInstallGroups(&sel, grupos)); n != len(grupos)+3 {
		t.Fatalf("grupos do formulário: %d, esperado %d (3 fixos + %d opções)", n, len(grupos)+3, len(grupos))
	}
}

func TestInstallFormRejeitaDependenciaAusente(t *testing.T) {
	sel := Selection{Skills: []string{"grill-me"}}
	form := NewInstallForm(installFormTestTheme(), &sel, app.InstallOptions())
	form.Init()
	for i := 0; i < 10 && len(form.Errors()) == 0; i++ {
		form.NextGroup()
	}
	erros := form.Errors()
	if len(erros) == 0 {
		t.Fatalf("grill-me sem grilling deveria gerar erro de validação")
	}
	if !strings.Contains(erros[0].Error(), "grill-me requer selecionar grilling") {
		t.Fatalf("mensagem de erro inesperada: %v", erros[0])
	}
}

func TestInstallFormTeclasAtualizamSelection(t *testing.T) {
	grupos := app.InstallOptions()
	var engenharia app.OptionGroup
	for _, grupo := range grupos {
		if grupo.ID == "engenharia" {
			engenharia = grupo
		}
	}
	if len(engenharia.Options) == 0 {
		t.Fatalf("grupo engenharia sem opções")
	}
	primeira := engenharia.Options[0].Value
	sel := Selection{Home: "/casa/atual", SkipAgentUpdate: true}
	form := NewInstallForm(installFormTestTheme(), &sel, grupos)
	form.Init()
	form.NextGroup()
	form.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
	form.NextGroup()
	if !slices.Contains(sel.Skills, primeira) {
		t.Fatalf("Skills %v deveria conter %q após marcar no formulário", sel.Skills, primeira)
	}
	if sel.Home != "/casa/atual" {
		t.Fatalf("destino atual deveria manter o HOME %q; obtido %q", "/casa/atual", sel.Home)
	}
	if !sel.SkipAgentUpdate {
		t.Fatalf("SkipAgentUpdate deveria refletir o valor inicial")
	}
}

func TestInstallFormCaminhoAbsoluto(t *testing.T) {
	casos := []struct {
		valor    string
		esperaOK bool
	}{
		{valor: "", esperaOK: false},
		{valor: "   ", esperaOK: false},
		{valor: "relativo/pasta", esperaOK: false},
		{valor: t.TempDir(), esperaOK: true},
	}
	for _, caso := range casos {
		erro := installAbsolutePath(caso.valor)
		if caso.esperaOK && erro != nil {
			t.Fatalf("%q deveria ser aceito: %v", caso.valor, erro)
		}
		if !caso.esperaOK && erro == nil {
			t.Fatalf("%q deveria ser rejeitado", caso.valor)
		}
	}
}

func TestInstallFormFlushSeparaSkillsDeHooks(t *testing.T) {
	var sel Selection
	st := &installState{
		selection: &sel,
		current:   "/casa/atual",
		isolated:  true,
		home:      "/tmp/teste",
		update:    false,
		groups: []installGroupValue{
			{chosen: []string{"tdd"}},
			{chosen: []string{"grill-me", "grilling"}},
			{hooks: true, chosen: []string{"gex-conventional-commits"}},
		},
	}
	st.flush()
	if !slices.Equal(sel.Skills, []string{"tdd", "grill-me", "grilling"}) {
		t.Fatalf("Skills = %v", sel.Skills)
	}
	if !slices.Equal(sel.Hooks, []string{"gex-conventional-commits"}) {
		t.Fatalf("Hooks = %v", sel.Hooks)
	}
	if sel.Home != "/tmp/teste" {
		t.Fatalf("Home isolado deveria ser %q; obtido %q", "/tmp/teste", sel.Home)
	}
	if !sel.SkipAgentUpdate {
		t.Fatalf("com update=false, SkipAgentUpdate deveria ser verdadeiro")
	}
}

func TestInstallFormFlushUsaHomeAtualSemIsolamento(t *testing.T) {
	var sel Selection
	st := &installState{selection: &sel, current: "/casa/atual", home: "/tmp/ignorado", update: true}
	st.flush()
	if sel.Home != "/casa/atual" || sel.SkipAgentUpdate {
		t.Fatalf("HOME atual deveria ser %q e atualizar agentes; obtido %+v", "/casa/atual", sel)
	}
}

func TestInstallFormOpcaoComDicaUsaRotuloEDica(t *testing.T) {
	grupo := installFormGrupoProdutividade(t)
	rotulo := installOptionLabel(app.Option{Label: "grill-me", Hint: "Entrevista sobre o plano."})
	if rotulo != "grill-me · Entrevista sobre o plano." {
		t.Fatalf("rótulo com dica = %q", rotulo)
	}
	if installOptionLabel(grupo.Options[0]) == "" {
		t.Fatalf("opção sem rótulo")
	}
}
