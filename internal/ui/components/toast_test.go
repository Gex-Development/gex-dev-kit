package components

import (
	"bytes"
	"testing"
	"time"

	"charm.land/bubbles/v2/timer"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/teatest/v2"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

func toastTheme() theme.Theme {
	return theme.New(theme.Options{Dark: true, Profile: colorprofile.ASCII})
}

type toastHarness struct {
	toast Toast
	start tea.Cmd
}

func (h toastHarness) Init() tea.Cmd {
	return h.start
}

func (h toastHarness) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	h.toast, cmd = h.toast.Update(msg)
	if !h.toast.Visible() {
		return h, tea.Quit
	}
	return h, cmd
}

func (h toastHarness) View() tea.View {
	return tea.NewView(h.toast.View(toastTheme(), 60))
}

func TestToastSomeAposOTimeoutDoTimer(t *testing.T) {
	inicio := time.Now()
	toast, start := NewToast().Show("interrupção adiada", theme.StateWarn, 300*time.Millisecond)
	tm := teatest.NewTestModel(t, toastHarness{toast: toast, start: start},
		teatest.WithInitialTermSize(80, 24),
		teatest.WithProgramOptions(
			tea.WithColorProfile(colorprofile.ASCII),
			tea.WithEnvironment([]string{"TERM=xterm-256color"}),
		),
	)
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("interrupção adiada"))
	}, teatest.WithDuration(2*time.Second))
	final := tm.FinalModel(t, teatest.WithFinalTimeout(5*time.Second)).(toastHarness)
	if final.toast.Visible() {
		t.Fatalf("toast continuou visível depois do timeout")
	}
	if decorrido := time.Since(inicio); decorrido < 300*time.Millisecond {
		t.Fatalf("toast sumiu em %v; quero no mínimo 300ms", decorrido)
	}
}

func TestToastIgnoraTimeoutDeToastAnterior(t *testing.T) {
	primeiro, _ := NewToast().Show("um", theme.StateInfo, time.Second)
	segundo, _ := primeiro.Show("dois", theme.StateInfo, time.Second)
	segundo, _ = segundo.Update(timer.TimeoutMsg{ID: primeiro.clock.ID()})
	if !segundo.Visible() {
		t.Fatalf("timeout de toast anterior escondeu o atual")
	}
	segundo, _ = segundo.Update(timer.TimeoutMsg{ID: segundo.clock.ID()})
	if segundo.Visible() {
		t.Fatalf("timeout do toast atual não escondeu o toast")
	}
}

func TestToastDuracaoZeroAindaEsconde(t *testing.T) {
	x, _ := NewToast().Show("copie o comando abaixo", theme.StateInfo, 0)
	x, _ = x.Update(timer.TickMsg{ID: x.clock.ID()})
	if !x.clock.Timedout() {
		t.Fatalf("toast com duração zero não chegou ao fim no primeiro tick")
	}
	x, _ = x.Update(timer.TimeoutMsg{ID: x.clock.ID()})
	if x.Visible() {
		t.Fatalf("toast com duração zero continuou visível")
	}
}

func TestToastMostraSimboloETextoOuNada(t *testing.T) {
	x, _ := NewToast().Show("interrupção adiada", theme.StateWarn, time.Second)
	if got := ansi.Strip(x.View(toastTheme(), 60)); got != "! interrupção adiada" {
		t.Fatalf("toast = %q; quero \"! interrupção adiada\"", got)
	}
	if largura := lipgloss.Width(x.View(toastTheme(), 10)); largura > 10 {
		t.Fatalf("toast com %d colunas; largura pedida 10", largura)
	}
	x, _ = x.Update(timer.TimeoutMsg{ID: x.clock.ID()})
	if got := x.View(toastTheme(), 60); got != "" {
		t.Fatalf("toast escondido ainda desenha %q", got)
	}
}
