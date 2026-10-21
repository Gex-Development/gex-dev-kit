package theme

import (
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
)

func TestNewChoosesPaletteAndFlags(t *testing.T) {
	dark := New(Options{Dark: true, Plain: true})
	if !dark.Dark || !dark.Plain {
		t.Fatalf("tema escuro sem Dark ou Plain: %+v", dark)
	}
	if got := tokensHex(dark.Palette.Accent); got != "#7fc3ef" {
		t.Fatalf("acento do tema escuro = %s, esperado #7fc3ef", got)
	}
	light := New(Options{})
	if light.Dark || light.Plain {
		t.Fatalf("tema claro com Dark ou Plain ligado: %+v", light)
	}
	if got := tokensHex(light.Palette.Accent); got != "#1a5fc0" {
		t.Fatalf("acento do tema claro = %s, esperado #1a5fc0", got)
	}
}

func TestRichOnlyAtANSI256OrMore(t *testing.T) {
	cases := []struct {
		profile colorprofile.Profile
		rich    bool
	}{
		{colorprofile.NoTTY, false},
		{colorprofile.ASCII, false},
		{colorprofile.ANSI, false},
		{colorprofile.ANSI256, true},
		{colorprofile.TrueColor, true},
	}
	for _, c := range cases {
		if got := New(Options{Profile: c.profile}).Rich; got != c.rich {
			t.Fatalf("perfil %s: Rich = %v, esperado %v", c.profile, got, c.rich)
		}
	}
}

func TestStateStyleAlwaysSymbolAndText(t *testing.T) {
	for _, plain := range []bool{true, false} {
		th := New(Options{Dark: true, Plain: plain})
		for state := StatePending; state <= StateInfo; state++ {
			symbol := state.Symbol(plain)
			if got := ansi.Strip(th.StateStyle(state).Render("texto")); got != symbol+" texto" {
				t.Fatalf("estado %d, Plain=%v: texto com estilo = %q, esperado %q", state, plain, got, symbol+" texto")
			}
			if got := ansi.Strip(th.StateStyle(state).Render("")); got != symbol+" "+state.Label() {
				t.Fatalf("estado %d, Plain=%v: sem texto = %q, esperado rótulo %q", state, plain, got, symbol+" "+state.Label())
			}
		}
	}
}

func TestStateStyleUsesStateColor(t *testing.T) {
	th := New(Options{Dark: true})
	if got, want := th.StateStyle(StateOK).Render(""), th.Success.Render("✓ ok"); got != want {
		t.Fatalf("estilo de ok = %q, esperado %q", got, want)
	}
	if got, want := th.StateStyle(StateError).Render(""), th.Error.Render("✗ erro"); got != want {
		t.Fatalf("estilo de erro = %q, esperado %q", got, want)
	}
}

func TestBlendReturnsNColors(t *testing.T) {
	th := New(Options{Dark: true})
	for _, n := range []int{0, 1, 3, 10} {
		if got := len(th.Blend(n)); got != n {
			t.Fatalf("Blend(%d) devolveu %d cores", n, got)
		}
	}
	colors := th.Blend(10)
	if tokensHex(colors[0]) != "#1e6fd9" || tokensHex(colors[9]) != "#7fc3ef" {
		t.Fatalf("extremos do Blend(10) = %s e %s, esperado #1e6fd9 e #7fc3ef", tokensHex(colors[0]), tokensHex(colors[9]))
	}
}

func TestProgressColorsGradientOnlyWhenRich(t *testing.T) {
	rich := New(Options{Dark: true, Profile: colorprofile.TrueColor}).ProgressColors()
	if len(rich) != 3 || tokensHex(rich[0]) != "#1e6fd9" {
		t.Fatalf("ProgressColors com TrueColor = %d cores, primeira %s", len(rich), tokensHex(rich[0]))
	}
	simple := New(Options{Dark: true, Profile: colorprofile.ANSI}).ProgressColors()
	if len(simple) != 1 || tokensHex(simple[0]) != "#7fc3ef" {
		t.Fatalf("ProgressColors sem cor rica = %d cores", len(simple))
	}
}

func TestSpinnerUsesAccent(t *testing.T) {
	th := New(Options{Dark: true})
	if got, want := th.Spinner().Render("x"), lipgloss.NewStyle().Foreground(th.Palette.Accent).Render("x"); got != want {
		t.Fatalf("spinner = %q, esperado %q", got, want)
	}
}

func TestHelpUsesThemeColors(t *testing.T) {
	th := New(Options{Dark: false})
	s := th.Help()
	if got := tokensHex(s.ShortKey.GetForeground()); got != "#1a5fc0" {
		t.Fatalf("tecla do help = %s, esperado #1a5fc0", got)
	}
	if got := tokensHex(s.ShortDesc.GetForeground()); got != "#4b5669" {
		t.Fatalf("descrição do help = %s, esperado #4b5669", got)
	}
}

func TestTableUsesThemeColors(t *testing.T) {
	th := New(Options{Dark: true})
	s := th.Table()
	if got := tokensHex(s.Selected.GetBackground()); got != "#14202e" {
		t.Fatalf("fundo da linha selecionada = %s, esperado #14202e", got)
	}
	if got := tokensHex(s.Cell.GetForeground()); got != "#f4f4f5" {
		t.Fatalf("texto da célula = %s, esperado #f4f4f5", got)
	}
}

func TestListItemUsesThemeColors(t *testing.T) {
	th := New(Options{Dark: false})
	s := th.ListItem()
	if got := tokensHex(s.SelectedTitle.GetForeground()); got != "#1a5fc0" {
		t.Fatalf("título selecionado = %s, esperado #1a5fc0", got)
	}
	if got := tokensHex(s.NormalTitle.GetForeground()); got != "#0b1526" {
		t.Fatalf("título normal = %s, esperado #0b1526", got)
	}
}
