package screens

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"
	"time"
	"unicode"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/teatest/v2"

	"github.com/Gex-Development/gex-dev-kit/internal/catalog"
)

const addonsCavemanClaude = "claude plugin marketplace add JuliusBrussee/caveman && claude plugin install caveman@caveman"

const addonsPonytailClaude = "/plugin marketplace add DietrichGebert/ponytail"

func addonsTestCommon() Common {
	return Common{Version: "v0.3.0", Profile: colorprofile.ASCII}
}

func addonsTestProgram(t *testing.T, width, height int) *teatest.TestModel {
	t.Helper()
	return teatest.NewTestModel(t, newAddonsModel(addonsTestCommon(), catalog.Addons()),
		teatest.WithInitialTermSize(width, height),
		teatest.WithProgramOptions(
			tea.WithColorProfile(colorprofile.ASCII),
			tea.WithEnvironment([]string{"TERM=xterm-256color"}),
		),
	)
}

func addonsFinalView(t *testing.T, tm *teatest.TestModel, teclas ...tea.KeyPressMsg) string {
	t.Helper()
	for _, tecla := range teclas {
		tm.Send(tecla)
	}
	tm.Send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	tm.WaitFinished(t, teatest.WithFinalTimeout(5*time.Second))
	return tm.FinalModel(t, teatest.WithFinalTimeout(5*time.Second)).View().Content
}

func addonsCompacto(texto string) string {
	var b strings.Builder
	for _, r := range ansi.Strip(texto) {
		if unicode.IsSpace(r) || strings.ContainsRune("│─╭╮╰╯|+", r) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func TestAddonsPrimeiroCardMostraFonteLicencaEComando(t *testing.T) {
	view := addonsFinalView(t, addonsTestProgram(t, 80, 24))
	texto := ansi.Strip(view)
	for _, trecho := range []string{"Fonte:", "Licença:", "1 de 4"} {
		if !strings.Contains(texto, trecho) {
			t.Fatalf("primeiro addon sem %q:\n%s", trecho, texto)
		}
	}
	if !strings.Contains(addonsCompacto(view), addonsCompacto(addonsCavemanClaude)) {
		t.Fatalf("comando oficial do caveman não aparece inteiro:\n%s", texto)
	}
}

func TestAddonsSetaDireitaTrocaParaSegundoAddon(t *testing.T) {
	view := addonsFinalView(t, addonsTestProgram(t, 80, 24), tea.KeyPressMsg{Code: tea.KeyRight})
	texto := ansi.Strip(view)
	if !strings.Contains(texto, "2 de 4") {
		t.Fatalf("→ não foi para a segunda página:\n%s", texto)
	}
	if !strings.Contains(addonsCompacto(view), addonsCompacto(addonsPonytailClaude)) {
		t.Fatalf("comando do ponytail não apareceu após →:\n%s", texto)
	}
}

func TestAddonsCabeNa60x20SemExceder(t *testing.T) {
	view := addonsFinalView(t, addonsTestProgram(t, 60, 20))
	linhas := strings.Split(view, "\n")
	if len(linhas) > 20 {
		t.Fatalf("60×20 produziu %d linhas", len(linhas))
	}
	for _, linha := range linhas {
		if ansi.StringWidth(linha) > 60 {
			t.Fatalf("linha com %d colunas em 60×20: %q", ansi.StringWidth(linha), ansi.Strip(linha))
		}
	}
}

func TestAddonsNaoUsaMouseMode(t *testing.T) {
	m := newAddonsModel(addonsTestCommon(), catalog.Addons())
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	if v := m.View(); v.MouseMode != tea.MouseModeNone {
		t.Fatalf("addons deve manter o mouse desligado para permitir copiar o comando; modo %v", v.MouseMode)
	}
}

func TestAddonsNaoImportaExecucaoDeComando(t *testing.T) {
	arquivos := token.NewFileSet()
	arvore, err := parser.ParseFile(arquivos, "addons.go", nil, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("não consegui ler os imports de addons.go: %v", err)
	}
	for _, imp := range arvore.Imports {
		if imp.Path.Value == `"os/exec"` {
			t.Fatalf("addons.go não pode importar os/exec")
		}
	}
}

func TestAddonsCtrlCNaoDevolveInterrupt(t *testing.T) {
	m := newAddonsModel(addonsTestCommon(), catalog.Addons())
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	if cmd == nil {
		t.Fatalf("ctrl+c deveria devolver comando de saída")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("ctrl+c deveria devolver tea.Quit, não %T", cmd())
	}
	if !m.interrupted {
		t.Fatalf("ctrl+c deveria marcar a tela como interrompida")
	}
}

func TestAddonsSemAnimacaoAgendada(t *testing.T) {
	if cmd := newAddonsModel(addonsTestCommon(), catalog.Addons()).Init(); cmd != nil {
		t.Fatalf("addons não tem animação e não deveria agendar comando inicial")
	}
}

func TestAddonsPaginaNaoPassaDosLimites(t *testing.T) {
	m := newAddonsModel(addonsTestCommon(), catalog.Addons())
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	if m.pager.Page != 0 {
		t.Fatalf("← na primeira página foi para %d", m.pager.Page)
	}
	for i := 0; i < 10; i++ {
		m.Update(tea.KeyPressMsg{Code: 'l', Text: "l"})
	}
	if m.pager.Page != len(catalog.Addons())-1 {
		t.Fatalf("l além da última página parou em %d", m.pager.Page)
	}
}

func TestAddonsTamanhoPequenoIgnoraSetas(t *testing.T) {
	m := newAddonsModel(addonsTestCommon(), catalog.Addons())
	m.Update(tea.WindowSizeMsg{Width: 30, Height: 10})
	m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	if m.pager.Page != 0 {
		t.Fatalf("tela pequena não deveria trocar de página; página %d", m.pager.Page)
	}
}

func TestAddonsInterrogacaoMostraAjuda(t *testing.T) {
	m := newAddonsModel(addonsTestCommon(), catalog.Addons())
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m.Update(tea.KeyPressMsg{Code: '?', Text: "?"})
	if !strings.Contains(ansi.Strip(m.View().Content), "trocar de addon") {
		t.Fatalf("? deveria mostrar a ajuda")
	}
}

func TestAddonsRevisaoCurtaTiraSeteCaracteres(t *testing.T) {
	casos := map[string]string{
		"bad48a0e1df55e05e3f52249afa35ecc1909ee76": "bad48a0",
		"abc": "abc",
	}
	for entrada, esperado := range casos {
		if obtido := addonsShortRevision(entrada); obtido != esperado {
			t.Fatalf("addonsShortRevision(%q) = %q, esperado %q", entrada, obtido, esperado)
		}
	}
}

func TestAddonsRotuloDoTipo(t *testing.T) {
	if addonsKindLabel("metodo") != "método" || addonsKindLabel("estilo") != "estilo" {
		t.Fatalf("rótulos de tipo inesperados")
	}
}
