package components

import (
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/stopwatch"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/teatest/v2"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

func loadingTheme() theme.Theme {
	return theme.New(theme.Options{Dark: true, Profile: colorprofile.ASCII})
}

type loadingHarness struct {
	loading Loading
	ticks   int
	start   tea.Cmd
}

func (h loadingHarness) Init() tea.Cmd {
	return h.start
}

func (h loadingHarness) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(stopwatch.TickMsg); ok {
		h.ticks++
	}
	var cmd tea.Cmd
	h.loading, cmd = h.loading.Update(msg)
	if h.ticks >= 3 {
		return h, tea.Quit
	}
	return h, cmd
}

func (h loadingHarness) View() tea.View {
	return tea.NewView(h.loading.View(loadingTheme(), 30, 2))
}

func TestLoadingAnimadoAvancaComCronometro(t *testing.T) {
	l := NewLoading(true)
	tm := teatest.NewTestModel(t, loadingHarness{loading: l, start: l.Tick()},
		teatest.WithInitialTermSize(80, 24),
		teatest.WithProgramOptions(
			tea.WithColorProfile(colorprofile.ASCII),
			tea.WithEnvironment([]string{"TERM=xterm-256color"}),
		),
	)
	final := tm.FinalModel(t, teatest.WithFinalTimeout(5*time.Second)).(loadingHarness)
	if final.ticks < 3 {
		t.Fatalf("cronômetro deu %d ticks; quero 3", final.ticks)
	}
	if final.loading.clock.Elapsed() < 3*loadingTick {
		t.Fatalf("cronômetro do carregamento = %v; quero no mínimo %v", final.loading.clock.Elapsed(), 3*loadingTick)
	}
}

func TestLoadingEstaticoNaoAgendaTick(t *testing.T) {
	l := NewLoading(false)
	if l.Tick() != nil {
		t.Fatalf("carregamento estático agendou Tick")
	}
	next, cmd := l.Update(stopwatch.TickMsg{})
	if cmd != nil {
		t.Fatalf("carregamento estático devolveu comando após TickMsg")
	}
	if next.View(loadingTheme(), 30, 2) != l.View(loadingTheme(), 30, 2) {
		t.Fatalf("carregamento estático mudou a tela após TickMsg")
	}
}

func TestLoadingDesenhaLinhasDentroDaLargura(t *testing.T) {
	linhas := strings.Split(ansi.Strip(NewLoading(false).View(loadingTheme(), 30, 3)), "\n")
	esperadas := []int{30, 20, 10}
	if len(linhas) != len(esperadas) {
		t.Fatalf("carregamento com %d linhas; quero %d", len(linhas), len(esperadas))
	}
	for i, linha := range linhas {
		if got := lipgloss.Width(linha); got != esperadas[i] {
			t.Fatalf("linha %d com %d colunas; quero %d", i, got, esperadas[i])
		}
	}
	if got := NewLoading(false).View(loadingTheme(), 0, 3); got != "" {
		t.Fatalf("largura zero desenhou %q", got)
	}
}

func TestLoadingPlainUsaCaracteresAscii(t *testing.T) {
	plain := theme.New(theme.Options{Plain: true, Profile: colorprofile.ASCII})
	if got := ansi.Strip(loadingRow(plain, 5, 0, false)); got != "#####" {
		t.Fatalf("esqueleto em modo plain = %q; quero \"#####\"", got)
	}
}

func TestLoadingBrilhoPercorreSemMudarTamanho(t *testing.T) {
	th := loadingTheme()
	inicial := loadingRow(th, 20, 0, true)
	depois := loadingRow(th, 20, 4, true)
	if inicial == depois {
		t.Fatalf("brilho não mudou de posição entre dois quadros")
	}
	if got := ansi.Strip(depois); got != strings.Repeat("█", 20) {
		t.Fatalf("brilho alterou o número de células: %q", got)
	}
}
