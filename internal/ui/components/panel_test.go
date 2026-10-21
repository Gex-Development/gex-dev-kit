package components

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
)

const panelTestBody = "texto longo para forcar quebra de linha dentro do painel com varias palavras"

func panelTestTheme(profile colorprofile.Profile, plain bool) theme.Theme {
	return theme.New(theme.Options{Dark: true, Profile: profile, Plain: plain})
}

func panelTestASCII(s string) bool {
	for _, r := range s {
		if r >= 0x80 {
			return false
		}
	}
	return true
}

func TestPanelCabeNaLargura(t *testing.T) {
	for _, width := range []int{20, 40, 80} {
		for _, plain := range []bool{false, true} {
			th := panelTestTheme(colorprofile.TrueColor, plain)
			out := Panel(th, PanelOptions{Title: "Painel", Width: width, Glow: true}, panelTestBody)
			if got := lipgloss.Width(out); got > width {
				t.Fatalf("largura pedida %d (plain=%v): painel com %d colunas", width, plain, got)
			}
		}
	}
}

func TestPanelBordasTemOsCantosCorretos(t *testing.T) {
	cases := []struct {
		name                                       string
		kind                                       BorderKind
		topLeft, topRight, bottomLeft, bottomRight string
	}{
		{"arredondada", BorderRounded, "╭", "╮", "╰", "╯"},
		{"normal", BorderNormal, "┌", "┐", "└", "┘"},
		{"grossa", BorderThick, "┏", "┓", "┗", "┛"},
		{"dupla", BorderDouble, "╔", "╗", "╚", "╝"},
		{"bloco", BorderBlock, "█", "█", "█", "█"},
		{"ascii", BorderASCII, "+", "+", "+", "+"},
	}
	th := panelTestTheme(colorprofile.TrueColor, false)
	for _, c := range cases {
		out := ansi.Strip(Panel(th, PanelOptions{Width: 20, Border: c.kind}, "x"))
		lines := strings.Split(out, "\n")
		top := []rune(lines[0])
		bottom := []rune(lines[len(lines)-1])
		got := string(top[0]) + string(top[len(top)-1]) + string(bottom[0]) + string(bottom[len(bottom)-1])
		want := c.topLeft + c.topRight + c.bottomLeft + c.bottomRight
		if got != want {
			t.Fatalf("borda %s: cantos = %q, esperado %q", c.name, got, want)
		}
	}
}

func TestPanelPlainUsaBordaAscii(t *testing.T) {
	th := panelTestTheme(colorprofile.TrueColor, true)
	out := ansi.Strip(Panel(th, PanelOptions{Width: 20, Border: BorderThick}, "x"))
	if first := strings.Split(out, "\n")[0]; !strings.HasPrefix(first, "+") {
		t.Fatalf("Plain com borda grossa deveria usar ASCII, linha = %q", first)
	}
}

func TestPanelGlowRicoDifereDeBordaSolida(t *testing.T) {
	rich := panelTestTheme(colorprofile.TrueColor, false)
	poor := panelTestTheme(colorprofile.ASCII, false)
	glow := Panel(rich, PanelOptions{Width: 40, Glow: true}, "corpo")
	solid := Panel(poor, PanelOptions{Width: 40, Glow: true}, "corpo")
	if glow == solid {
		t.Fatalf("painel com glow e cor rica renderizou igual ao sem cor rica")
	}
	if solid != Panel(poor, PanelOptions{Width: 40}, "corpo") {
		t.Fatalf("sem cor rica, Glow deveria cair na borda sólida Accent")
	}
	if Panel(rich, PanelOptions{Width: 40}, "corpo") != solid {
		t.Fatalf("com cor rica e sem Glow, a borda deveria ser a sólida Accent")
	}
}

func TestPanelPlainSoAscii(t *testing.T) {
	th := panelTestTheme(colorprofile.TrueColor, true)
	out := Panel(th, PanelOptions{Title: "Painel", Width: 40, Border: BorderThick, Glow: true}, "corpo ascii")
	if !panelTestASCII(out) {
		t.Fatalf("Plain com caractere fora de ASCII: %q", out)
	}
}

func TestPanelCardCabeNaLargura(t *testing.T) {
	th := panelTestTheme(colorprofile.TrueColor, false)
	for _, width := range []int{20, 40, 80} {
		out := Card(th, CardOptions{Title: "Pacote", Meta: "Core", Footer: "rodape", Width: width, Selected: true}, panelTestBody)
		if got := lipgloss.Width(out); got > width {
			t.Fatalf("largura pedida %d: card com %d colunas", width, got)
		}
	}
}

func TestPanelCardSelecionadoDifereDoNormal(t *testing.T) {
	th := panelTestTheme(colorprofile.TrueColor, false)
	selected := Card(th, CardOptions{Title: "Pacote", Meta: "Core", Width: 40, Selected: true}, "descricao")
	normal := Card(th, CardOptions{Title: "Pacote", Meta: "Core", Width: 40}, "descricao")
	if selected == normal {
		t.Fatalf("card selecionado renderizou igual ao normal")
	}
}

func TestPanelCardPlainSoAscii(t *testing.T) {
	th := panelTestTheme(colorprofile.TrueColor, true)
	out := Card(th, CardOptions{Title: "Pacote", Meta: "Core", Footer: "rodape", Width: 40, Selected: true}, "descricao")
	if !panelTestASCII(out) {
		t.Fatalf("card Plain com caractere fora de ASCII: %q", out)
	}
}
