package components

import (
	"bytes"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/teatest/v2"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

const stepsTestFill = "▌"

func stepsTheme() theme.Theme {
	return theme.New(theme.Options{Dark: true, Profile: colorprofile.ASCII})
}

type stepsAdvance struct{ state theme.State }

type stepsHarness struct {
	steps StepProgress
	start tea.Cmd
}

func (h stepsHarness) Init() tea.Cmd {
	return h.start
}

func (h stepsHarness) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	if adv, ok := msg.(stepsAdvance); ok {
		h.steps, cmd = h.steps.Advance(adv.state)
		return h, cmd
	}
	h.steps, cmd = h.steps.Update(msg)
	return h, cmd
}

func (h stepsHarness) View() tea.View {
	return tea.NewView(h.steps.View(stepsTheme(), 60))
}

func TestStepProgressBarraChegaAoAlvoNoPrograma(t *testing.T) {
	steps, start := NewStepProgress(stepsTheme(), []string{"baixar", "instalar", "validar"}, true).Start()
	tm := teatest.NewTestModel(t, stepsHarness{steps: steps, start: start},
		teatest.WithInitialTermSize(80, 24),
		teatest.WithProgramOptions(
			tea.WithColorProfile(colorprofile.ASCII),
			tea.WithEnvironment([]string{"TERM=xterm-256color"}),
		),
	)
	tm.Send(stepsAdvance{state: theme.StateOK})
	celulas := 60 - stepsClockCells - 1
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Count(bts, []byte(stepsTestFill)) >= celulas/3
	}, teatest.WithDuration(5*time.Second))
	tm.Quit()
	final := tm.FinalModel(t, teatest.WithFinalTimeout(5*time.Second)).(stepsHarness)
	if alvo := final.steps.bar.Percent(); alvo != 1.0/3 {
		t.Fatalf("alvo da barra = %v; quero 1/3", alvo)
	}
	if final.steps.Elapsed() <= 0 {
		t.Fatalf("cronômetro não contou depois de quadros reais: %v", final.steps.Elapsed())
	}
}

func TestStepProgressEstaticoNaoAgendaComandos(t *testing.T) {
	s, cmd := NewStepProgress(stepsTheme(), []string{"a", "b", "c"}, false).Start()
	if cmd != nil {
		t.Fatalf("Start estático devolveu comando")
	}
	s, cmd = s.Advance(theme.StateOK)
	if cmd != nil {
		t.Fatalf("Advance estático devolveu comando")
	}
	if s.Elapsed() != 0 {
		t.Fatalf("cronômetro estático contou: %v", s.Elapsed())
	}
	linha := strings.Split(ansi.Strip(s.View(stepsTheme(), 40)), "\n")[0]
	if got := strings.Count(linha, stepsTestFill); got != 10 {
		t.Fatalf("barra com %d células cheias; quero 10 de 31 (um terço)", got)
	}
}

func TestStepProgressMarcaEstadoDasEtapas(t *testing.T) {
	s, _ := NewStepProgress(stepsTheme(), []string{"baixar", "instalar", "validar"}, false).Start()
	s, _ = s.Advance(theme.StateOK)
	s, _ = s.Advance(theme.StateWarn)
	view := ansi.Strip(s.View(stepsTheme(), 60))
	for _, want := range []string{"✓ baixar", "! instalar", "▸ validar"} {
		if !strings.Contains(view, want) {
			t.Fatalf("tela sem %q:\n%s", want, view)
		}
	}
}

func TestStepProgressAdvanceAposFimNaoAvanca(t *testing.T) {
	s, _ := NewStepProgress(stepsTheme(), []string{"a"}, false).Start()
	s, _ = s.Advance(theme.StateOK)
	depois, cmd := s.Advance(theme.StateError)
	if cmd != nil {
		t.Fatalf("Advance além do fim devolveu comando")
	}
	if depois.states[0] != theme.StateOK {
		t.Fatalf("Advance além do fim trocou o estado da etapa: %v", depois.states[0])
	}
}

func TestStepProgressWithThemeMantemEtapaAtual(t *testing.T) {
	s, _ := NewStepProgress(stepsTheme(), []string{"a", "b"}, false).Start()
	s, _ = s.Advance(theme.StateOK)
	claro := s.WithTheme(theme.New(theme.Options{Profile: colorprofile.ASCII}))
	if claro.next != 1 || claro.bar.Percent() != 0.5 {
		t.Fatalf("WithTheme perdeu o passo: próximo=%d, alvo=%v", claro.next, claro.bar.Percent())
	}
	if got := ansi.Strip(claro.View(stepsTheme(), 60)); !strings.Contains(got, "▸ b") {
		t.Fatalf("etapa atual não aparece depois de WithTheme:\n%s", got)
	}
}

func TestStepProgressPlainUsaCaracteresAscii(t *testing.T) {
	plain := theme.New(theme.Options{Plain: true, Profile: colorprofile.ASCII})
	s, _ := NewStepProgress(plain, []string{"a", "b"}, false).Start()
	s, _ = s.Advance(theme.StateOK)
	linha := strings.Split(ansi.Strip(s.View(plain, 40)), "\n")[0]
	if strings.Contains(linha, stepsTestFill) || !strings.Contains(linha, "#") {
		t.Fatalf("barra em modo plain não usa ASCII: %q", linha)
	}
}

func TestStepProgressFormataTempoComVirgula(t *testing.T) {
	if got := stepsClockText(1400 * time.Millisecond); got != "1,4 s" {
		t.Fatalf("tempo de 1,4 s = %q; quero \"1,4 s\"", got)
	}
	if got := stepsClockText(0); got != "0,0 s" {
		t.Fatalf("tempo zero = %q; quero \"0,0 s\"", got)
	}
}
