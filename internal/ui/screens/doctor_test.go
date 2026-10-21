package screens

import (
	"image/color"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/app"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
	"github.com/charmbracelet/colorprofile"
)

func doctorTestCommon(t *testing.T) Common {
	return Common{Version: "teste", UserHome: t.TempDir(), Dark: true, Profile: colorprofile.TrueColor}
}

func doctorTestReport(state, detail string) app.Report {
	return app.Report{Root: "/tmp/teste", Items: []app.Status{{Path: ".claude/CLAUDE.md", State: state, Detail: detail}}}
}

func doctorTestResize(t *testing.T, report app.Report, width, height int) doctorModel {
	t.Helper()
	updated, _ := newDoctorModel(doctorTestCommon(t), report).Update(tea.WindowSizeMsg{Width: width, Height: height})
	return updated.(doctorModel)
}

func TestDoctorTableFitsNarrowTerminal(t *testing.T) {
	screen := doctorTestResize(t, doctorTestReport("alterado", "Arquivo foi alterado."), 40, 12)
	view := screen.View().Content
	if lipgloss.Width(view) > 40 || lipgloss.Height(view) > 12 {
		t.Fatalf("tabela excedeu terminal: %dx%d", lipgloss.Width(view), lipgloss.Height(view))
	}
	if !strings.Contains(view, ".claude/CLAUDE.md") {
		t.Fatalf("caminho desapareceu da tabela:\n%s", view)
	}
}

func TestDoctorWrapsLongDetailInNarrowTerminal(t *testing.T) {
	screen := doctorTestResize(t, doctorTestReport("alterado", strings.Repeat("texto longo ", 12)), 40, 16)
	view := screen.View().Content
	if width := lipgloss.Width(view); width > 40 {
		t.Fatalf("detalhe do doctor excedeu 40 colunas: %d", width)
	}
	if height := lipgloss.Height(view); height > 16 {
		t.Fatalf("detalhe do doctor excedeu 16 linhas: %d", height)
	}
}

func TestDoctorShortListDoesNotFillScreenWithEmptyRows(t *testing.T) {
	screen := doctorTestResize(t, doctorTestReport("ok", "Íntegro."), 80, 30)
	if height := lipgloss.Height(screen.View().Content); height > 10 {
		t.Fatalf("uma linha ocupou tela com espaço vazio: %d linhas", height)
	}
}

func TestDoctorFitsTinyTerminal(t *testing.T) {
	sizes := []tea.WindowSizeMsg{{Width: 14, Height: 6}, {Width: 40, Height: 8}, {Width: 40, Height: 9}}
	for _, size := range sizes {
		screen := doctorTestResize(t, doctorTestReport("alterado", strings.Repeat("informação importante ", 6)), size.Width, size.Height)
		view := screen.View().Content
		if lipgloss.Width(view) > size.Width || lipgloss.Height(view) > size.Height {
			t.Fatalf("doctor ultrapassou %dx%d: %dx%d", size.Width, size.Height, lipgloss.Width(view), lipgloss.Height(view))
		}
		if !strings.Contains(view, "Terminal") {
			t.Fatalf("tamanho %dx%d devia mostrar TooSmall:\n%s", size.Width, size.Height, view)
		}
	}
}

func TestDoctorAdaptsPaletteToLightBackground(t *testing.T) {
	doctor := newDoctorModel(doctorTestCommon(t), doctorTestReport("ok", "Íntegro."))
	light, _ := doctor.Update(tea.BackgroundColorMsg{Color: color.White})
	if doctor.View().Content == light.View().Content {
		t.Fatal("doctor manteve cores de fundo escuro no terminal claro")
	}
}

func TestDoctorStateKindMapsKitStates(t *testing.T) {
	cases := map[string]theme.State{
		"ok":            theme.StateOK,
		"pendente":      theme.StatePending,
		"sem-historico": theme.StatePending,
		"atualizar":     theme.StateWarn,
		"aviso":         theme.StateInfo,
		"alterado":      theme.StateError,
		"ausente":       theme.StateError,
		"incompleto":    theme.StateError,
		"bloqueado":     theme.StateError,
		"desconhecido":  theme.StateWarn,
	}
	for state, want := range cases {
		if got := doctorStateKind(state); got != want {
			t.Fatalf("estado %q: esperado %v, veio %v", state, want, got)
		}
	}
}
