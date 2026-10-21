package components

import (
	"runtime"
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

func keysPressFor(tecla string) tea.KeyPressMsg {
	switch tecla {
	case "ctrl+c":
		return tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}
	case "ctrl+z":
		return tea.KeyPressMsg{Code: 'z', Mod: tea.ModCtrl}
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEscape}
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "up":
		return tea.KeyPressMsg{Code: tea.KeyUp}
	case "down":
		return tea.KeyPressMsg{Code: tea.KeyDown}
	case "tab":
		return tea.KeyPressMsg{Code: tea.KeyTab}
	}
	r := []rune(tecla)[0]
	return tea.KeyPressMsg{Code: r, Text: tecla}
}

func keysTheme() theme.Theme {
	return theme.New(theme.Options{Dark: true, Profile: colorprofile.ASCII})
}

func TestKeysPadraoReconheceTeclas(t *testing.T) {
	k := DefaultKeys()
	casos := []struct {
		tecla   string
		binding key.Binding
	}{
		{"up", k.Up}, {"k", k.Up},
		{"down", k.Down}, {"j", k.Down},
		{"enter", k.Enter}, {"esc", k.Back},
		{"q", k.Quit}, {"ctrl+c", k.Quit},
		{"?", k.Help}, {"/", k.Filter}, {"tab", k.Tab},
		{"s", k.Yes}, {"y", k.Yes}, {"n", k.No},
	}
	for _, c := range casos {
		if !key.Matches(keysPressFor(c.tecla), c.binding) {
			t.Fatalf("a tecla %q deveria acionar o atalho, mas não acionou", c.tecla)
		}
	}
}

func TestKeysTextInputDeixaLetrasParaOCampo(t *testing.T) {
	original := DefaultKeys()
	k := original.TextInput()
	for _, tecla := range []string{"q", "?", "s", "n", "t", "m", "j", "k", "/", "tab", "ctrl+z"} {
		if key.Matches(keysPressFor(tecla), k.Quit, k.Help, k.Yes, k.No, k.Up, k.Down, k.Filter, k.Tab, k.Suspend) {
			t.Fatalf("a tecla %q virou atalho dentro de campo de texto", tecla)
		}
	}
	for _, tecla := range []string{"ctrl+c", "esc", "enter"} {
		if !key.Matches(keysPressFor(tecla), k.Quit, k.Back, k.Enter) {
			t.Fatalf("a tecla %q deveria continuar global dentro de campo de texto", tecla)
		}
	}
	if !key.Matches(keysPressFor("q"), original.Quit) {
		t.Fatalf("TextInput alterou o mapa original; q deixou de sair")
	}
}

func TestKeysSuspendDesligadoQuandoSistemaNaoSuporta(t *testing.T) {
	k := keysBuild(func() bool { return false })
	if k.Suspend.Enabled() {
		t.Fatalf("ctrl+z deveria estar desabilitado quando o sistema não suporta suspensão")
	}
	if key.Matches(keysPressFor("ctrl+z"), k.Suspend) {
		t.Fatalf("ctrl+z acionou suspensão mesmo desabilitado")
	}
}

func TestKeysSuspendLigadoQuandoSistemaSuporta(t *testing.T) {
	k := keysBuild(func() bool { return true })
	if !k.Suspend.Enabled() {
		t.Fatalf("ctrl+z deveria estar habilitado quando o sistema suporta suspensão")
	}
	if !key.Matches(keysPressFor("ctrl+z"), k.Suspend) {
		t.Fatalf("ctrl+z não acionou suspensão habilitado")
	}
}

func TestKeysSuspendUsaSistemaAtual(t *testing.T) {
	esperado := runtime.GOOS != "windows"
	if got := DefaultKeys().Suspend.Enabled(); got != esperado {
		t.Fatalf("suspensão habilitada = %v; quero %v em %s", got, esperado, runtime.GOOS)
	}
}
