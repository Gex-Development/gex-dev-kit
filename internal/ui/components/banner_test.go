package components

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
)

func bannerTestTheme(profile colorprofile.Profile, plain bool) theme.Theme {
	return theme.New(theme.Options{Dark: true, Profile: profile, Plain: plain})
}

func TestBannerCabeNaLargura(t *testing.T) {
	th := bannerTestTheme(colorprofile.TrueColor, false)
	if got := Banner(th, 0, "GEX", ""); strings.Contains(got, "\n") {
		t.Fatalf("banner com largura 0 = %q, esperado só o título", got)
	}
	for _, width := range []int{20, 40, 80} {
		out := Banner(th, width, "GEX Dev Kit · instalação guiada de especializações", "tela de instalação com prévia antes de gravar")
		if got := lipgloss.Width(out); got > width {
			t.Fatalf("largura pedida %d: banner com %d colunas", width, got)
		}
	}
}

func TestBannerRegraTemLarguraPedida(t *testing.T) {
	th := bannerTestTheme(colorprofile.TrueColor, false)
	lines := strings.Split(ansi.Strip(Banner(th, 40, "GEX", "subtitulo")), "\n")
	if rule := lines[len(lines)-1]; len([]rune(rule)) != 40 {
		t.Fatalf("regra do banner tem %d caracteres, esperado 40", len([]rune(rule)))
	}
}

func TestBannerRichUsaGradiente(t *testing.T) {
	rich := Banner(bannerTestTheme(colorprofile.TrueColor, false), 40, "GEX", "subtitulo")
	flat := Banner(bannerTestTheme(colorprofile.ASCII, false), 40, "GEX", "subtitulo")
	if rich == flat {
		t.Fatalf("banner com cor rica renderizou igual ao sem cor rica")
	}
}

func TestBannerPlainSoAscii(t *testing.T) {
	th := bannerTestTheme(colorprofile.TrueColor, true)
	out := ansi.Strip(Banner(th, 40, "GEX Dev Kit", "instalacao guiada"))
	for _, r := range out {
		if r >= 0x80 {
			t.Fatalf("banner em Plain tem caractere fora de ASCII: %q", out)
		}
	}
}
