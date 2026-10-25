package components

import (
	"strings"
	"testing"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
)

func TestStepperMarcaFeitosAtualEPendentes(t *testing.T) {
	th := wordmarkTestTheme(colorprofile.TrueColor, false)
	out := ansi.Strip(Stepper(th, 80, []string{"Repositório", "Hooks", "Prévia"}, 1))
	for _, want := range []string{"✓ Repositório", "2 Hooks", "3 Prévia"} {
		if !strings.Contains(out, want) {
			t.Fatalf("etapa %q ausente: %q", want, out)
		}
	}
}

func TestStepperNaoPassaDaLargura(t *testing.T) {
	th := wordmarkTestTheme(colorprofile.TrueColor, false)
	for width := 1; width <= 60; width++ {
		if w := ansi.StringWidth(ansi.Strip(Stepper(th, width, []string{"Repositório", "Hooks", "Comandos", "Prévia"}, 2))); w > width {
			t.Fatalf("largura %d, saída com %d colunas", width, w)
		}
	}
}

func TestStepperPlainSoAscii(t *testing.T) {
	th := wordmarkTestTheme(colorprofile.ASCII, true)
	if out := Stepper(th, 80, []string{"A", "B", "C"}, 1); !wordmarkTestASCII(out) {
		t.Fatalf("Plain com caractere fora de ASCII: %q", out)
	}
}
