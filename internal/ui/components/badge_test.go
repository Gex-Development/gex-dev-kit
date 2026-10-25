package components

import (
	"strings"
	"testing"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
)

func badgeTestTheme(plain bool) theme.Theme {
	return theme.New(theme.Options{Dark: true, Profile: colorprofile.TrueColor, Plain: plain})
}

func TestStatusBadgeSimboloETexto(t *testing.T) {
	for _, plain := range []bool{false, true} {
		th := badgeTestTheme(plain)
		for state := theme.StatePending; state <= theme.StateInfo; state++ {
			want := state.Symbol(plain) + " instalado"
			if got := strings.TrimSpace(ansi.Strip(StatusBadge(th, state, "instalado"))); got != want {
				t.Fatalf("estado %d (plain=%v): selo = %q, esperado %q", state, plain, got, want)
			}
		}
	}
}

func TestStatusBadgeSemRotuloUsaRotuloDoEstado(t *testing.T) {
	th := badgeTestTheme(false)
	if got, want := strings.TrimSpace(ansi.Strip(StatusBadge(th, theme.StateError, ""))), "✗ erro"; got != want {
		t.Fatalf("selo sem rótulo = %q, esperado %q", got, want)
	}
}

func TestStatusBadgePlainSoAscii(t *testing.T) {
	th := badgeTestTheme(true)
	for state := theme.StatePending; state <= theme.StateInfo; state++ {
		for _, r := range ansi.Strip(StatusBadge(th, state, "ok")) {
			if r >= 0x80 {
				t.Fatalf("estado %d em Plain tem caractere fora de ASCII: %q", state, r)
			}
		}
	}
}
