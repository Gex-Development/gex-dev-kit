package components

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
)

func buttonsTestSet() Buttons {
	return NewButtons(Button{ID: "novo", Label: "Setup do zero"}, Button{ID: "atualizar", Label: "Atualizar"}).Place(4, 7)
}

func buttonsPressed(t *testing.T, cmd tea.Cmd) ButtonPressMsg {
	t.Helper()
	if cmd == nil {
		t.Fatal("esperava um comando de pressionar")
	}
	press, ok := cmd().(ButtonPressMsg)
	if !ok {
		t.Fatalf("comando devolveu %T, esperado ButtonPressMsg", cmd())
	}
	return press
}

func TestButtonsClicaNoBotaoCertoEIgnoraForaDele(t *testing.T) {
	b := buttonsTestSet()
	secondX := 4 + len("Setup do zero") + 2*buttonsPadding + buttonsGap
	next, cmd := b.Update(tea.MouseClickMsg{X: secondX, Y: 7, Button: tea.MouseLeft})
	if press := buttonsPressed(t, cmd); press.Button.ID != "atualizar" || next.Focus() != 1 {
		t.Fatalf("clique no segundo botão = %+v, foco %d", press, next.Focus())
	}
	for _, miss := range []tea.MouseClickMsg{
		{X: 4, Y: 8, Button: tea.MouseLeft},
		{X: 3, Y: 7, Button: tea.MouseLeft},
		{X: secondX - 1, Y: 7, Button: tea.MouseLeft},
		{X: 4, Y: 7, Button: tea.MouseRight},
	} {
		if _, cmd := b.Update(miss); cmd != nil {
			t.Fatalf("clique %+v não deveria pressionar", miss)
		}
	}
}

func TestButtonsTecladoGiraFocoEConfirma(t *testing.T) {
	b := buttonsTestSet()
	b, _ = b.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	if b.Focus() != 1 {
		t.Fatalf("direita deveria focar o segundo, foco %d", b.Focus())
	}
	b, _ = b.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	if b.Focus() != 0 {
		t.Fatalf("direita no último deveria voltar ao primeiro, foco %d", b.Focus())
	}
	b, _ = b.Update(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	if b.Focus() != 1 {
		t.Fatalf("shift+tab no primeiro deveria ir ao último, foco %d", b.Focus())
	}
	_, cmd := b.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if press := buttonsPressed(t, cmd); press.Index != 1 {
		t.Fatalf("enter pressionou o índice %d", press.Index)
	}
}

func TestButtonsLarguraDaViewConfereComAreaClicavel(t *testing.T) {
	b := buttonsTestSet()
	th := wordmarkTestTheme(colorprofile.TrueColor, false)
	if got := ansi.StringWidth(b.View(th)); got != b.Width() {
		t.Fatalf("view tem %d colunas, Width() diz %d", got, b.Width())
	}
	plain := wordmarkTestTheme(colorprofile.ASCII, true)
	if got := ansi.StringWidth(b.View(plain)); got != b.Width() {
		t.Fatalf("view plain tem %d colunas, Width() diz %d", got, b.Width())
	}
}
