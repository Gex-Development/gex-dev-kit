package components

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
)

func emptyTestTheme(plain bool) theme.Theme {
	return theme.New(theme.Options{Dark: true, Profile: colorprofile.TrueColor, Plain: plain})
}

func TestEmptyStateCentralizaOTitulo(t *testing.T) {
	th := emptyTestTheme(false)
	lines := strings.Split(ansi.Strip(EmptyState(th, 40, "Nada aqui", "Instale um pacote")), "\n")
	if len(lines) != 2 {
		t.Fatalf("estado vazio com dica tem %d linhas, esperado 2", len(lines))
	}
	title := lines[0]
	lead := len(title) - len(strings.TrimLeft(title, " "))
	trail := len(title) - len(strings.TrimRight(title, " "))
	if lead == 0 || lead-trail > 1 || trail-lead > 1 {
		t.Fatalf("título não centralizado: %d espaços à esquerda e %d à direita", lead, trail)
	}
}

func TestEmptyStateCabeNaLargura(t *testing.T) {
	th := emptyTestTheme(false)
	for _, width := range []int{20, 40, 80} {
		out := EmptyState(th, width, "Nada instalado", "Escolha uma especialização na tela de instalação para ver o resultado aqui")
		if got := lipgloss.Width(out); got > width {
			t.Fatalf("largura pedida %d: estado vazio com %d colunas", width, got)
		}
	}
}

func TestEmptyStateSemDicaTemUmaLinha(t *testing.T) {
	th := emptyTestTheme(false)
	if got := strings.Count(ansi.Strip(EmptyState(th, 40, "Nada", "")), "\n"); got != 0 {
		t.Fatalf("estado vazio sem dica tem %d quebras de linha, esperado 0", got)
	}
}

func TestEmptyStatePlainSoAscii(t *testing.T) {
	th := emptyTestTheme(true)
	out := ansi.Strip(EmptyState(th, 40, "Nada aqui", "Instale um pacote"))
	for _, r := range out {
		if r >= 0x80 {
			t.Fatalf("estado vazio em Plain tem caractere fora de ASCII: %q", out)
		}
	}
}
