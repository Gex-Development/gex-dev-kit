package theme

import (
	"fmt"
	"image/color"
	"math"
	"testing"
)

type tokensItem struct {
	name  string
	value color.Color
}

func tokensList(p Palette) []tokensItem {
	return []tokensItem{
		{"BrandBlue", p.BrandBlue},
		{"BrandBlueDeep", p.BrandBlueDeep},
		{"BrandCyan", p.BrandCyan},
		{"BrandCyanSoft", p.BrandCyanSoft},
		{"Accent", p.Accent},
		{"Navy", p.Navy},
		{"Text", p.Text},
		{"TextSecondary", p.TextSecondary},
		{"TextFaint", p.TextFaint},
		{"SurfaceActive", p.SurfaceActive},
		{"BorderCard", p.BorderCard},
		{"Surface", p.Surface},
		{"SurfaceRaised", p.SurfaceRaised},
		{"BorderStrong", p.BorderStrong},
		{"Glow", p.Glow},
		{"Positive", p.Positive},
		{"Negative", p.Negative},
		{"Warning", p.Warning},
	}
}

func tokensHex(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02x%02x%02x", r>>8, g>>8, b>>8)
}

func tokensChannel(v uint32) float64 {
	x := float64(v) / 65535
	if x <= 0.03928 {
		return x / 12.92
	}
	return math.Pow((x+0.055)/1.055, 2.4)
}

func tokensLuminance(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	return 0.2126*tokensChannel(r) + 0.7152*tokensChannel(g) + 0.0722*tokensChannel(b)
}

func tokensRatio(a, b color.Color) float64 {
	la, lb := tokensLuminance(a), tokensLuminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

func TestTokensAreSetAndSurfacesAreDistinct(t *testing.T) {
	for name, p := range map[string]Palette{"escuro": Dark(), "claro": Light()} {
		for _, item := range tokensList(p) {
			if item.value == nil {
				t.Fatalf("tema %s: token %s sem valor", name, item.name)
			}
		}
		if len(p.Gradient) < 2 {
			t.Fatalf("tema %s: gradiente precisa de ao menos 2 cores", name)
		}
		if tokensHex(p.Surface) == tokensHex(p.SurfaceRaised) || tokensHex(p.SurfaceRaised) == tokensHex(p.SurfaceActive) {
			t.Fatalf("tema %s: superfícies sem hierarquia visual", name)
		}
	}
}

func TestSelectedTextContrastsWithItsBar(t *testing.T) {
	for name, dark := range map[string]bool{"escuro": true, "claro": false} {
		th := New(Options{Dark: dark})
		fg, bg := th.Selected.GetForeground(), th.Selected.GetBackground()
		if ratio := tokensRatio(fg, bg); ratio < 4.5 {
			t.Fatalf("tema %s: item selecionado tem contraste %.2f, abaixo de 4,5", name, ratio)
		}
	}
}

func TestContrastRatioBlackAndWhite(t *testing.T) {
	if ratio := tokensRatio(color.Black, color.White); math.Abs(ratio-21) > 0.01 {
		t.Fatalf("razão entre preto e branco = %.2f, esperado 21", ratio)
	}
}

func TestWCAGContrastOfTokens(t *testing.T) {
	cases := []struct {
		name       string
		palette    Palette
		background color.Color
	}{
		{"escuro", Dark(), color.Black},
		{"claro", Light(), color.White},
	}
	for _, c := range cases {
		tokens := []tokensItem{
			{"Text", c.palette.Text},
			{"TextSecondary", c.palette.TextSecondary},
			{"Accent", c.palette.Accent},
			{"Positive", c.palette.Positive},
			{"Negative", c.palette.Negative},
			{"Warning", c.palette.Warning},
		}
		for _, token := range tokens {
			if ratio := tokensRatio(token.value, c.background); ratio < 4.5 {
				t.Fatalf("tema %s: %s tem contraste %.2f, abaixo de 4,5", c.name, token.name, ratio)
			}
		}
	}
}

func TestStateSymbolsAndLabels(t *testing.T) {
	cases := []struct {
		state   State
		plain   string
		unicode string
		label   string
	}{
		{StatePending, "[ ]", "○", "pendente"},
		{StateRunning, ">", "▸", "rodando"},
		{StateOK, "[ok]", "✓", "ok"},
		{StateWarn, "[!]", "!", "atenção"},
		{StateError, "[x]", "✗", "erro"},
		{StateSkipped, "[-]", "–", "ignorado"},
		{StateInfo, "[i]", "i", "info"},
	}
	for _, c := range cases {
		if got := c.state.Symbol(true); got != c.plain {
			t.Fatalf("estado %d: símbolo Plain = %q, esperado %q", c.state, got, c.plain)
		}
		if got := c.state.Symbol(false); got != c.unicode {
			t.Fatalf("estado %d: símbolo Unicode = %q, esperado %q", c.state, got, c.unicode)
		}
		if got := c.state.Label(); got != c.label {
			t.Fatalf("estado %d: rótulo = %q, esperado %q", c.state, got, c.label)
		}
		for i := 0; i < len(c.plain); i++ {
			if c.plain[i] >= 0x80 {
				t.Fatalf("estado %d: símbolo Plain %q não é ASCII", c.state, c.plain)
			}
		}
	}
}
