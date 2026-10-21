package layout

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
	"github.com/charmbracelet/x/ansi"
)

func layoutTestTheme() theme.Theme {
	return theme.New(theme.Options{Dark: true})
}

func TestClassifyBoundaries(t *testing.T) {
	cases := []struct {
		size Size
		want Class
	}{
		{Size{120, 30}, ClassWide},
		{Size{119, 30}, ClassRegular},
		{Size{80, 24}, ClassRegular},
		{Size{79, 24}, ClassCompact},
		{Size{60, 20}, ClassCompact},
		{Size{40, 12}, ClassCompact},
		{Size{39, 12}, ClassTiny},
		{Size{40, 11}, ClassTiny},
		{Size{0, 0}, ClassTiny},
	}
	for _, c := range cases {
		if got := Classify(c.size); got != c.want {
			t.Fatalf("Classify(%dx%d) = %d, esperado %d", c.size.Width, c.size.Height, got, c.want)
		}
	}
}

func TestRenderTemCabecalhoERodapeNasClassesNormalECompacta(t *testing.T) {
	th := layoutTestTheme()
	f := Frame{Title: "GEX Dev Kit", Context: "instalar", Version: "v0.3.0", Body: "corpo"}
	for _, s := range []Size{{80, 24}, {60, 20}, {40, 12}} {
		out := ansi.Strip(Render(th, s, f))
		lines := strings.Split(out, "\n")
		if len(lines) != s.Height {
			t.Fatalf("tamanho %dx%d: %d linhas, esperado %d", s.Width, s.Height, len(lines), s.Height)
		}
		if !strings.Contains(lines[0], "GEX Dev Kit") || !strings.Contains(lines[0], "v0.3.0") {
			t.Fatalf("tamanho %dx%d: cabeçalho sem título ou versão: %q", s.Width, s.Height, lines[0])
		}
		if last := lines[len(lines)-1]; !strings.Contains(last, "? ajuda") || !strings.Contains(last, "q sai") {
			t.Fatalf("tamanho %dx%d: rodapé sem atalhos ? e q: %q", s.Width, s.Height, last)
		}
	}
}

func TestRenderAvisoFicaAcimaDoRodape(t *testing.T) {
	th := layoutTestTheme()
	out := ansi.Strip(Render(th, Size{80, 24}, Frame{Title: "GEX", Notice: "salvo"}))
	lines := strings.Split(out, "\n")
	if got := lines[len(lines)-2]; !strings.Contains(got, "salvo") {
		t.Fatalf("aviso fora da linha acima do rodapé: %q", got)
	}
}

func TestBodySizeReservaCabecalhoAvisoERodape(t *testing.T) {
	if got := BodySize(Size{80, 24}); got != (Size{80, 20}) {
		t.Fatalf("BodySize(80x24) = %+v, esperado 80x20", got)
	}
	if got := BodySize(Size{3, 2}); got != (Size{3, 0}) {
		t.Fatalf("BodySize(3x2) = %+v, esperado altura 0", got)
	}
}

func TestRenderPequenoMostraSoAvisoComAtalhoQ(t *testing.T) {
	th := layoutTestTheme()
	for _, s := range []Size{{39, 12}, {20, 5}} {
		out := ansi.Strip(Render(th, s, Frame{Title: "GEX", Version: "v0.3.0"}))
		if !strings.Contains(out, "Terminal pequeno") || !strings.Contains(out, "q sai") {
			t.Fatalf("tamanho %dx%d: aviso sem texto ou atalho q: %q", s.Width, s.Height, out)
		}
	}
}

func TestTooSmallCabeNoTamanho(t *testing.T) {
	th := layoutTestTheme()
	for _, s := range []Size{{39, 12}, {40, 11}, {20, 5}, {10, 3}} {
		out := TooSmall(th, s)
		layoutTestLimites(t, out, s.Width, s.Height, "terminal pequeno")
	}
}

func TestFitCortaLarguraEAltura(t *testing.T) {
	if got, want := Fit("abcdef\nxyz\nfinal", 4, 2), "abc…\nxyz"; got != want {
		t.Fatalf("Fit = %q, esperado %q", got, want)
	}
}

func TestFitRespeitaLarguraComCJK(t *testing.T) {
	got := Fit(string(rune(0x6F22))+string(rune(0x5B57))+string(rune(0x6F22)), 3, 1)
	if w := lipgloss.Width(got); w > 3 {
		t.Fatalf("Fit com CJK teve largura %d, máximo 3: %q", w, got)
	}
}

func TestColumnsMantemLadoDireitoEmLarguraExata(t *testing.T) {
	th := layoutTestTheme()
	got := Columns(th, 30, "GEX Dev Kit · um título bem comprido", "v0.3.0")
	if w := lipgloss.Width(got); w != 30 {
		t.Fatalf("Columns teve largura %d, esperado 30: %q", w, got)
	}
	if !strings.HasSuffix(ansi.Strip(got), "v0.3.0") {
		t.Fatalf("Columns perdeu o lado direito: %q", ansi.Strip(got))
	}
}

func TestColumnsComDireitaMaiorQueLargura(t *testing.T) {
	th := layoutTestTheme()
	got := Columns(th, 5, "abc", "longo demais")
	if w := lipgloss.Width(got); w > 5 {
		t.Fatalf("Columns teve largura %d, máximo 5: %q", w, got)
	}
}

func TestCenterNaoExcedeTamanho(t *testing.T) {
	s := Size{10, 3}
	got := Center(s, "abcdefghijklmnop\nxyz\n1\n2\n3")
	layoutTestLimites(t, got, s.Width, s.Height, "centro")
	if n := strings.Count(got, "\n") + 1; n != s.Height {
		t.Fatalf("Center teve %d linhas, esperado %d", n, s.Height)
	}
}

func TestInvariantesDeTamanhoComTextoDificil(t *testing.T) {
	th := layoutTestTheme()
	frames := layoutTestFrames()
	for w := 20; w <= 200; w += 6 {
		for h := 5; h <= 60; h += 4 {
			for i, f := range frames {
				out := Render(th, Size{Width: w, Height: h}, f)
				layoutTestLimites(t, out, w, h, layoutTestFrameName(i))
			}
		}
	}
}

func TestInvariantesNasBordasECentro(t *testing.T) {
	th := layoutTestTheme()
	dificil := layoutTestFrames()[0].Body
	for _, s := range []Size{{80, 24}, {60, 20}, {40, 12}, {39, 11}, {20, 5}, {200, 60}} {
		for i, f := range layoutTestFrames() {
			layoutTestLimites(t, Render(th, s, f), s.Width, s.Height, layoutTestFrameName(i))
		}
		layoutTestLimites(t, Center(s, dificil), s.Width, s.Height, "centro")
		layoutTestLimites(t, TooSmall(th, s), s.Width, s.Height, "terminal pequeno")
	}
}

func layoutTestFrames() []Frame {
	esc := string(rune(27))
	nfc := "caf" + string(rune(0xE9))
	nfd := "cafe" + string(rune(0x301))
	cjk := string(rune(0x6F22)) + string(rune(0x5B57))
	emoji := string(rune(0x1F680)) + " " + string(rune(0x1F469)) + string(rune(0x200D)) + string(rune(0x1F4BB))
	colored := esc + "[1;31m" + "vermelho muito longo para estourar a largura" + esc + "[0m"
	mixed := strings.Repeat(colored+" "+nfc+" "+nfd+" "+cjk+" "+emoji+"\n", 8)
	open := esc + "[31m sem fechamento " + strings.Repeat("x", 300)
	return []Frame{
		{Title: mixed, Context: cjk, Version: emoji, Body: mixed, Footer: mixed, Notice: mixed},
		{Title: "GEX", Body: "", Footer: "", Notice: ""},
		{Title: nfd + " " + nfc, Context: open, Version: colored, Body: open, Footer: nfd, Notice: cjk},
	}
}

func layoutTestFrameName(i int) string {
	return "quadro " + string(rune('A'+i))
}

func layoutTestLimites(t *testing.T, out string, w, h int, caso string) {
	t.Helper()
	lines := strings.Split(out, "\n")
	if len(lines) > h {
		t.Fatalf("%dx%d (%s): %d linhas, máximo %d", w, h, caso, len(lines), h)
	}
	for i, line := range lines {
		if got := lipgloss.Width(line); got > w {
			t.Fatalf("%dx%d (%s): linha %d com largura %d, máximo %d", w, h, caso, i, got, w)
		}
	}
}
