package components

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
)

func sectionTestTheme(plain bool) theme.Theme {
	return theme.New(theme.Options{Dark: true, Profile: colorprofile.TrueColor, Plain: plain})
}

func TestSectionDividerTemLarguraPedida(t *testing.T) {
	th := sectionTestTheme(false)
	for _, width := range []int{20, 40, 80} {
		if got := lipgloss.Width(Divider(th, width)); got != width {
			t.Fatalf("divisor com largura %d tem %d colunas", width, got)
		}
	}
	if got := Divider(th, 0); got != "" {
		t.Fatalf("divisor com largura 0 = %q, esperado vazio", got)
	}
}

func TestSectionTituloCabeNaLargura(t *testing.T) {
	th := sectionTestTheme(false)
	if got := Section(th, 0, "Instalacao"); strings.Contains(got, "\n") {
		t.Fatalf("seção com largura 0 = %q, esperado só o título sem divisor", got)
	}
	for _, width := range []int{20, 40, 80} {
		out := Section(th, width, "Instalacao guiada de especializacoes do kit")
		if got := lipgloss.Width(out); got > width {
			t.Fatalf("seção com largura pedida %d tem %d colunas", width, got)
		}
	}
}

func TestSectionPlainSoAscii(t *testing.T) {
	th := sectionTestTheme(true)
	out := ansi.Strip(Section(th, 40, "Instalacao"))
	for _, r := range out {
		if r >= 0x80 {
			t.Fatalf("seção em Plain tem caractere fora de ASCII: %q", out)
		}
	}
	lines := strings.Split(out, "\n")
	if len(lines) != 2 || lines[1] != strings.Repeat("-", 40) {
		t.Fatalf("divisor de seção em Plain = %q, esperado uma linha de 40 traços", out)
	}
}
