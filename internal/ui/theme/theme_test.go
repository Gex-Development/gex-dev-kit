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
	if got, want := tokensHex(dark.Palette.Accent), tokensHex(Dark().Accent); got != want {
		t.Fatalf("acento do tema escuro = %s, esperado %s", got, want)
	}
	light := New(Options{})
	if light.Dark || light.Plain {
		t.Fatalf("tema claro com Dark ou Plain ligado: %+v", light)
	}
	if got, want := tokensHex(light.Palette.Accent), tokensHex(Light().Accent); got != want {
		t.Fatalf("acento do tema claro = %s, esperado %s", got, want)
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
	grad := th.Palette.Gradient
	if tokensHex(colors[0]) != tokensHex(grad[0]) || tokensHex(colors[9]) != tokensHex(grad[len(grad)-1]) {
		t.Fatalf("extremos do Blend(10) = %s e %s, esperado %s e %s", tokensHex(colors[0]), tokensHex(colors[9]), tokensHex(grad[0]), tokensHex(grad[len(grad)-1]))
	}
}

func TestProgressColorsGradientOnlyWhenRich(t *testing.T) {
	dark := Dark()
	rich := New(Options{Dark: true, Profile: colorprofile.TrueColor}).ProgressColors()
	if len(rich) != len(dark.Gradient) || tokensHex(rich[0]) != tokensHex(dark.Gradient[0]) {
		t.Fatalf("ProgressColors com TrueColor = %d cores, primeira %s", len(rich), tokensHex(rich[0]))
	}
	simple := New(Options{Dark: true, Profile: colorprofile.ANSI}).ProgressColors()
	if len(simple) != 1 || tokensHex(simple[0]) != tokensHex(dark.Accent) {
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
	if got, want := tokensHex(s.ShortKey.GetForeground()), tokensHex(th.Palette.Accent); got != want {
		t.Fatalf("tecla do help = %s, esperado %s", got, want)
	}
	if got, want := tokensHex(s.ShortDesc.GetForeground()), tokensHex(th.Palette.TextSecondary); got != want {
		t.Fatalf("descrição do help = %s, esperado %s", got, want)
	}
}

func TestTableUsesThemeColors(t *testing.T) {
	th := New(Options{Dark: true})
	s := th.Table()
	if got, want := tokensHex(s.Selected.GetBackground()), tokensHex(th.Palette.SurfaceActive); got != want {
		t.Fatalf("fundo da linha selecionada = %s, esperado %s", got, want)
	}
	if got, want := tokensHex(s.Cell.GetForeground()), tokensHex(th.Palette.Text); got != want {
		t.Fatalf("texto da célula = %s, esperado %s", got, want)
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
