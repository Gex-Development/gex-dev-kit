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

func tokensCheck(t *testing.T, p Palette, expected map[string]string) {
	t.Helper()
	items := tokensList(p)
	if len(items) != len(expected) {
		t.Fatalf("paleta tem %d tokens, esperado %d", len(items), len(expected))
	}
	for _, item := range items {
		if got := tokensHex(item.value); got != expected[item.name] {
			t.Fatalf("token %s = %s, esperado %s", item.name, got, expected[item.name])
		}
	}
}

func TestTokensDark(t *testing.T) {
	tokensCheck(t, Dark(), map[string]string{
		"BrandBlue":     "#1e6fd9",
		"BrandBlueDeep": "#0b3d86",
		"BrandCyan":     "#38bdf8",
		"BrandCyanSoft": "#7fc3ef",
		"Accent":        "#7fc3ef",
		"Navy":          "#0b1526",
		"Text":          "#f4f4f5",
		"TextSecondary": "#a1a1aa",
		"TextFaint":     "#3f3f46",
		"SurfaceActive": "#14202e",
		"BorderCard":    "#26272d",
		"Positive":      "#34d399",
		"Negative":      "#f87171",
		"Warning":       "#fcd34d",
	})
}

func TestTokensLight(t *testing.T) {
	tokensCheck(t, Light(), map[string]string{
		"BrandBlue":     "#1e6fd9",
		"BrandBlueDeep": "#0b3d86",
		"BrandCyan":     "#38bdf8",
		"BrandCyanSoft": "#7fc3ef",
		"Accent":        "#1a5fc0",
		"Navy":          "#0b1526",
		"Text":          "#0b1526",
		"TextSecondary": "#4b5669",
		"TextFaint":     "#b6bfcc",
		"SurfaceActive": "#e8f0fc",
		"BorderCard":    "#d0d7e1",
		"Positive":      "#047857",
		"Negative":      "#b91c1c",
		"Warning":       "#92400e",
	})
}

func TestGradients(t *testing.T) {
	cases := []struct {
		name     string
		palette  Palette
		expected []string
	}{
		{"escuro", Dark(), []string{"#1e6fd9", "#38bdf8", "#7fc3ef"}},
		{"claro", Light(), []string{"#0b3d86", "#1e6fd9", "#38bdf8"}},
	}
	for _, c := range cases {
		if len(c.palette.Gradient) != len(c.expected) {
			t.Fatalf("gradiente %s tem %d cores, esperado %d", c.name, len(c.palette.Gradient), len(c.expected))
		}
		for i, stop := range c.palette.Gradient {
			if got := tokensHex(stop); got != c.expected[i] {
				t.Fatalf("gradiente %s, cor %d = %s, esperado %s", c.name, i, got, c.expected[i])
			}
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
