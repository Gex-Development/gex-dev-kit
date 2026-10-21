package components

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestKeyHintsRespeitaLargura(t *testing.T) {
	tema := keysTheme()
	for largura := 1; largura <= 90; largura++ {
		for _, tudo := range []bool{false, true} {
			saida := KeyHints(tema, largura, tudo, DefaultKeys())
			if w := lipgloss.Width(saida); w > largura {
				t.Fatalf("largura pedida %d (tudo=%v), saída com %d colunas: %q", largura, tudo, w, ansi.Strip(saida))
			}
		}
	}
}

func TestKeyHintsMostraAtalhosEmPortugues(t *testing.T) {
	saida := ansi.Strip(KeyHints(keysTheme(), 100, false, DefaultKeys()))
	for _, trecho := range []string{"↑/k", "subir", "voltar", "sair"} {
		if !strings.Contains(saida, trecho) {
			t.Fatalf("ajuda curta não mostra %q: %q", trecho, saida)
		}
	}
}

func TestKeyHintsAjudaCompletaTemMaisAtalhos(t *testing.T) {
	curta := ansi.Strip(KeyHints(keysTheme(), 100, false, DefaultKeys()))
	completa := ansi.Strip(KeyHints(keysTheme(), 100, true, DefaultKeys()))
	if !strings.Contains(completa, "filtrar") {
		t.Fatalf("ajuda completa não mostra o atalho de filtro: %q", completa)
	}
	if strings.Contains(curta, "filtrar") {
		t.Fatalf("ajuda curta não deveria mostrar o atalho de filtro: %q", curta)
	}
}

func TestKeyHintsSegueMapaDeCampoDeTexto(t *testing.T) {
	saida := ansi.Strip(KeyHints(keysTheme(), 100, false, DefaultKeys().TextInput()))
	if strings.Contains(saida, "subir") {
		t.Fatalf("dentro de campo de texto a ajuda não deveria mostrar 'subir': %q", saida)
	}
	if !strings.Contains(saida, "voltar") {
		t.Fatalf("dentro de campo de texto a ajuda deveria mostrar 'voltar': %q", saida)
	}
}

func TestKeyHintsSemLarguraNaoDesenha(t *testing.T) {
	if saida := KeyHints(keysTheme(), 0, false, DefaultKeys()); saida != "" {
		t.Fatalf("largura 0 deveria devolver vazio, veio %q", saida)
	}
}
