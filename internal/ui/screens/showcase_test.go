package screens

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/exp/teatest/v2"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/motion"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

func showcaseApply(t *testing.T, m showcaseModel, msgs ...tea.Msg) showcaseModel {
	t.Helper()
	for _, msg := range msgs {
		next, _ := m.Update(msg)
		m = next.(showcaseModel)
	}
	return m
}

func showcaseTecla(s string) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: []rune(s)[0], Text: s}
}

func showcaseIndicePagina(t *testing.T, nome string) int {
	t.Helper()
	for i, p := range showcasePages() {
		if p.Name == nome {
			return i
		}
	}
	t.Fatalf("página %q não existe", nome)
	return -1
}

func TestShowcasePagesFitTerminals(t *testing.T) {
	tamanhos := []struct{ w, h int }{{80, 24}, {60, 20}, {40, 12}}
	for _, tam := range tamanhos {
		m := showcaseApply(t, newShowcaseModel(Common{Version: "teste", Profile: colorprofile.TrueColor}),
			tea.WindowSizeMsg{Width: tam.w, Height: tam.h})
		for i := range showcasePages() {
			conteudo := m.View().Content
			if w := lipgloss.Width(conteudo); w > tam.w {
				t.Fatalf("%dx%d, página %d: largura %d passa do terminal", tam.w, tam.h, i, w)
			}
			if h := lipgloss.Height(conteudo); h > tam.h {
				t.Fatalf("%dx%d, página %d: altura %d passa do terminal", tam.w, tam.h, i, h)
			}
			m = showcaseApply(t, m, tea.KeyPressMsg{Code: tea.KeyDown})
		}
	}
}

func TestShowcaseTestFilesExistem(t *testing.T) {
	for _, f := range showcaseTestFiles {
		if _, err := os.Stat(showcaseRepoPath(f.Path)); err != nil {
			t.Fatalf("arquivo de teste citado na vitrine não existe: %s: %v", f.Path, err)
		}
	}
}

func showcaseRepoPath(rel string) string {
	return "../../../" + rel
}

func TestShowcaseTextTemTodasAsPaginasSemEscape(t *testing.T) {
	texto := ShowcaseText(theme.New(theme.Options{}), 80)
	if strings.Contains(texto, "\x1b") {
		t.Fatalf("ShowcaseText ainda tem byte ESC")
	}
	for _, p := range showcasePages() {
		if !strings.Contains(texto, p.Name) {
			t.Fatalf("ShowcaseText não mostra a página %q", p.Name)
		}
	}
}

func TestShowcaseMDesligaMovimento(t *testing.T) {
	m := showcaseApply(t, newShowcaseModel(Common{Motion: true}), showcaseTecla("m"))
	if m.motion {
		t.Fatalf("tecla m não desligou o movimento")
	}
	frames := m.live.Frames
	next, cmd := m.Update(motion.FrameMsg{ID: showcaseFrameID})
	if cmd != nil {
		t.Fatalf("com movimento desligado, quadro não pode agendar outro quadro")
	}
	if next.(showcaseModel).live.Frames != frames {
		t.Fatalf("contador de quadros subiu com movimento desligado")
	}
}

func TestShowcaseProgramaAlternaAltScreen(t *testing.T) {
	m := showcaseApply(t, newShowcaseModel(Common{}), tea.WindowSizeMsg{Width: 80, Height: 24})
	for i := 0; i < showcaseIndicePagina(t, showcasePagePrograma); i++ {
		m = showcaseApply(t, m, tea.KeyPressMsg{Code: tea.KeyDown})
	}
	m = showcaseApply(t, m, tea.KeyPressMsg{Code: tea.KeyTab}, showcaseTecla("a"))
	if !m.View().AltScreen {
		t.Fatalf("tecla a na página Programa não ligou o alt screen")
	}
	m = showcaseApply(t, m, showcaseTecla("a"))
	if m.View().AltScreen {
		t.Fatalf("segunda tecla a não desligou o alt screen")
	}
}

func showcaseSaidaTeatest(t *testing.T, teclas ...tea.Msg) []byte {
	t.Helper()
	tm := teatest.NewTestModel(t, newShowcaseModel(Common{Profile: colorprofile.TrueColor, Version: "teste"}),
		teatest.WithInitialTermSize(80, 24),
		teatest.WithProgramOptions(tea.WithColorProfile(colorprofile.TrueColor)),
	)
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Identidade"))
	}, teatest.WithDuration(5*time.Second))
	for _, tecla := range teclas {
		tm.Send(tecla)
	}
	tm.Send(showcaseTecla("q"))
	saida, err := io.ReadAll(tm.FinalOutput(t, teatest.WithFinalTimeout(5*time.Second)))
	if err != nil {
		t.Fatalf("ler saída da vitrine: %v", err)
	}
	return saida
}

func TestShowcaseNavegacaoPelaTeatest(t *testing.T) {
	tm := teatest.NewTestModel(t, newShowcaseModel(Common{Motion: true, Profile: colorprofile.TrueColor}),
		teatest.WithInitialTermSize(80, 24),
		teatest.WithProgramOptions(tea.WithColorProfile(colorprofile.TrueColor)),
	)
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Identidade"))
	}, teatest.WithDuration(5*time.Second))
	tm.Send(tea.KeyPressMsg{Code: tea.KeyDown})
	tm.Send(tea.KeyPressMsg{Code: tea.KeyTab})
	tm.Send(showcaseTecla("t"))
	tm.Send(showcaseTecla("m"))
	tm.Send(showcaseTecla("q"))
	final := tm.FinalModel(t, teatest.WithFinalTimeout(5*time.Second)).(showcaseModel)
	if final.list.Index() != 1 {
		t.Fatalf("seta para baixo não mudou a página: índice %d", final.list.Index())
	}
	if !final.demoFocus {
		t.Fatalf("Tab não trocou o foco para a demo")
	}
	if !final.dark {
		t.Fatalf("tecla t não alternou para o tema escuro")
	}
	if final.motion {
		t.Fatalf("tecla m não desligou o movimento")
	}
}

func TestShowcaseTemaMudaASaida(t *testing.T) {
	claro := showcaseSaidaTeatest(t)
	escuro := showcaseSaidaTeatest(t, showcaseTecla("t"))
	if bytes.Equal(claro, escuro) {
		t.Fatalf("saída do tema claro e do escuro saiu igual")
	}
}

func TestShowcaseRunShowcaseSemTerminalFalha(t *testing.T) {
	if err := RunShowcase(Runner{}, Common{}); err == nil {
		t.Fatalf("RunShowcase sem terminal deveria falhar")
	}
}
