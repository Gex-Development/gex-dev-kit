package components

import (
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
)

func wordmarkTestTheme(profile colorprofile.Profile, plain bool) theme.Theme {
	return theme.New(theme.Options{Dark: true, Profile: profile, Plain: plain})
}

func wordmarkTestASCII(s string) bool {
	for _, r := range s {
		if r >= 0x80 {
			return false
		}
	}
	return true
}

func TestWordmarkCabeNaLargura(t *testing.T) {
	for _, width := range []int{20, 40, 80} {
		for _, plain := range []bool{false, true} {
			th := wordmarkTestTheme(colorprofile.TrueColor, plain)
			if got := lipgloss.Width(Wordmark(th, width, 1)); got > width {
				t.Fatalf("largura pedida %d (plain=%v): saída com %d colunas", width, plain, got)
			}
		}
	}
}

func TestWordmarkLetreiroSomenteAcimaDeSessenta(t *testing.T) {
	th := wordmarkTestTheme(colorprofile.TrueColor, false)
	big := ansi.Strip(Wordmark(th, 80, 1))
	if n := len(wordmarkTestLines(big)); n != 4 {
		t.Fatalf("letreiro com largura 80 tem %d linhas, esperado 4 (3 de blocos e DEV KIT)", n)
	}
	small := ansi.Strip(Wordmark(th, 40, 1))
	if small != "GEX · DEV KIT" {
		t.Fatalf("letreiro compacto = %q, esperado %q", small, "GEX · DEV KIT")
	}
}

func TestWordmarkCompactoPlainUsaHifen(t *testing.T) {
	th := wordmarkTestTheme(colorprofile.ASCII, true)
	if got := ansi.Strip(Wordmark(th, 40, 1)); got != "GEX - DEV KIT" {
		t.Fatalf("letreiro compacto em Plain = %q, esperado %q", got, "GEX - DEV KIT")
	}
}

func TestWordmarkRevelaColunasDaEsquerda(t *testing.T) {
	th := wordmarkTestTheme(colorprofile.TrueColor, false)
	if got := ansi.Strip(Wordmark(th, 80, 0)); !wordmarkTestBlank(got) {
		t.Fatalf("com reveal 0 ainda aparece conteúdo: %q", got)
	}
	half := wordmarkTestLines(ansi.Strip(Wordmark(th, 80, 0.5)))[0]
	runes := []rune(half)
	if runes[0] != '█' {
		t.Fatalf("com reveal 0.5 a primeira coluna deveria estar visível, linha = %q", half)
	}
	if runes[len(runes)-1] != ' ' {
		t.Fatalf("com reveal 0.5 a última coluna deveria estar oculta, linha = %q", half)
	}
	full := wordmarkTestLines(ansi.Strip(Wordmark(th, 80, 1)))[0]
	if []rune(full)[len([]rune(full))-1] != '█' {
		t.Fatalf("com reveal 1 a última coluna deveria estar visível, linha = %q", full)
	}
}

func TestWordmarkRichUsaGradiente(t *testing.T) {
	rich := Wordmark(wordmarkTestTheme(colorprofile.TrueColor, false), 80, 1)
	flat := Wordmark(wordmarkTestTheme(colorprofile.ASCII, false), 80, 1)
	if rich == flat {
		t.Fatalf("saída com cor rica e sem cor rica são iguais")
	}
}

func TestWordmarkPlainSoAscii(t *testing.T) {
	th := wordmarkTestTheme(colorprofile.TrueColor, true)
	for _, width := range []int{20, 40, 80} {
		if out := Wordmark(th, width, 1); !wordmarkTestASCII(out) {
			t.Fatalf("Plain com largura %d tem caractere fora de ASCII: %q", width, out)
		}
	}
}

func wordmarkTestLines(s string) []string {
	var lines []string
	start := 0
	for i, r := range s {
		if r == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	return append(lines, s[start:])
}

func wordmarkTestBlank(s string) bool {
	for _, r := range s {
		if r != ' ' && r != '\n' {
			return false
		}
	}
	return true
}
