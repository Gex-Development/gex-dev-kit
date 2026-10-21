package components

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

func confirmTestDialog() Confirm {
	return NewConfirm("Aplicar mudanças?", "Grava os arquivos listados.", "Aplicar", "Cancelar")
}

func TestConfirmEnterSozinhoNaoDecide(t *testing.T) {
	_, dec := confirmTestDialog().Update(keysPressFor("enter"), true)
	if dec != DecisionNone {
		t.Fatalf("enter sozinho não deveria decidir; decisão %v", dec)
	}
}

func TestConfirmSimENao(t *testing.T) {
	casos := []struct {
		tecla string
		quer  Decision
	}{
		{"s", DecisionYes},
		{"y", DecisionYes},
		{"n", DecisionNo},
		{"esc", DecisionNo},
		{"q", DecisionNone},
		{"x", DecisionNone},
		{"tab", DecisionNone},
	}
	for _, c := range casos {
		_, dec := confirmTestDialog().Update(keysPressFor(c.tecla), true)
		if dec != c.quer {
			t.Fatalf("tecla %q: decisão %v; quero %v", c.tecla, dec, c.quer)
		}
	}
}

func TestConfirmInvisivelNaoConsomeTeclas(t *testing.T) {
	for _, tecla := range []string{"s", "y", "n", "esc", "enter"} {
		_, dec := confirmTestDialog().Update(keysPressFor(tecla), false)
		if dec != DecisionNone {
			t.Fatalf("modal invisível decidiu com a tecla %q: %v", tecla, dec)
		}
	}
}

func TestConfirmVisibleExigeEspaco(t *testing.T) {
	c := confirmTestDialog()
	if c.Visible(23, 11) || c.Visible(24, 10) {
		t.Fatalf("modal não deveria caber abaixo de 24 colunas ou 11 linhas")
	}
	if !c.Visible(24, 11) {
		t.Fatalf("modal deveria caber em 24 por 11")
	}
}

func TestConfirmViewCabeNaLargura(t *testing.T) {
	tema := keysTheme()
	c := confirmTestDialog()
	for largura := 12; largura <= 80; largura++ {
		for _, linha := range strings.Split(c.View(tema, largura), "\n") {
			if w := lipgloss.Width(linha); w > largura {
				t.Fatalf("largura pedida %d, linha com %d colunas: %q", largura, w, ansi.Strip(linha))
			}
		}
	}
}

func TestConfirmViewMostraTituloEAtalhos(t *testing.T) {
	saida := ansi.Strip(confirmTestDialog().View(keysTheme(), 60))
	for _, trecho := range []string{"Aplicar mudanças?", "s/y", "n/esc", "Cancelar"} {
		if !strings.Contains(saida, trecho) {
			t.Fatalf("modal não mostra %q: %q", trecho, saida)
		}
	}
}

func TestConfirmBordaDuplaEmTelaRica(t *testing.T) {
	tema := theme.New(theme.Options{Dark: true, Profile: colorprofile.TrueColor})
	saida := ansi.Strip(confirmTestDialog().View(tema, 60))
	if !strings.Contains(saida, "╔") {
		t.Fatalf("modal em tela rica deveria ter borda dupla: %q", saida)
	}
}

func TestConfirmBordaASCIIEmModoSimples(t *testing.T) {
	tema := theme.New(theme.Options{Dark: true, Profile: colorprofile.ASCII, Plain: true})
	saida := ansi.Strip(confirmTestDialog().View(tema, 60))
	if strings.Contains(saida, "╔") || !strings.Contains(saida, "+") {
		t.Fatalf("modo simples deveria usar borda ASCII: %q", saida)
	}
}

func TestConfirmOverlayCabeNaTela(t *testing.T) {
	tema := keysTheme()
	c := confirmTestDialog()
	for _, tamanho := range [][2]int{{80, 24}, {60, 20}, {24, 11}, {40, 14}} {
		largura, altura := tamanho[0], tamanho[1]
		base := strings.TrimSuffix(strings.Repeat(strings.Repeat(".", largura)+"\n", altura), "\n")
		saida := c.Overlay(tema, base, largura, altura)
		linhas := strings.Split(saida, "\n")
		if len(linhas) > altura {
			t.Fatalf("overlay %dx%d saiu com %d linhas", largura, altura, len(linhas))
		}
		for _, linha := range linhas {
			if w := lipgloss.Width(linha); w > largura {
				t.Fatalf("overlay %dx%d saiu com linha de %d colunas: %q", largura, altura, w, ansi.Strip(linha))
			}
		}
		if !strings.Contains(ansi.Strip(saida), "Aplicar mudanças?") {
			t.Fatalf("overlay %dx%d não mostra o modal", largura, altura)
		}
	}
}

func TestConfirmOverlaySemEspacoMostraSoModal(t *testing.T) {
	tema := keysTheme()
	c := confirmTestDialog()
	if got, quer := c.Overlay(tema, "base", 20, 8), c.View(tema, 20); got != quer {
		t.Fatalf("sem espaço o overlay deveria mostrar só o modal")
	}
}
