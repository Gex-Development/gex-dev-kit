package components

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
)

func errorboxTestTheme(plain bool) theme.Theme {
	return theme.New(theme.Options{Dark: true, Profile: colorprofile.TrueColor, Plain: plain})
}

func TestErrorBoxSempreMostraCausaEAção(t *testing.T) {
	th := errorboxTestTheme(false)
	out := ansi.Strip(ErrorBox(th, 60, "falha ao gravar", "permissao negada", "Rode o doctor"))
	for _, want := range []string{"Causa:", "Ação:", "permissao negada", "Rode o doctor"} {
		if !strings.Contains(out, want) {
			t.Fatalf("caixa de erro sem %q:\n%s", want, out)
		}
	}
	vazio := ansi.Strip(ErrorBox(th, 60, "falha", "", ""))
	for _, want := range []string{"Causa:", "Ação:", "não informada", "Rode gex-dev-kit doctor"} {
		if !strings.Contains(vazio, want) {
			t.Fatalf("caixa de erro com causa e ação vazias sem %q:\n%s", want, vazio)
		}
	}
}

func TestErrorBoxPlainMantemAcentos(t *testing.T) {
	th := errorboxTestTheme(true)
	out := ErrorBox(th, 60, "falha", "", "")
	for _, want := range []string{"Causa:", "Ação:", "não informada"} {
		if !strings.Contains(ansi.Strip(out), want) {
			t.Fatalf("caixa de erro em Plain sem %q:\n%s", want, out)
		}
	}
	if !strings.HasPrefix(ansi.Strip(out), "+") {
		t.Fatalf("caixa de erro em Plain sem borda ASCII (+): %q", ansi.Strip(out))
	}
}

func TestErrorBoxCabeNaLargura(t *testing.T) {
	th := errorboxTestTheme(false)
	for _, width := range []int{20, 40, 80} {
		out := ErrorBox(th, width, "falha ao aplicar a instalação", "o arquivo de configuração do agente não pôde ser gravado no diretório pessoal", "feche o agente e rode de novo")
		if got := lipgloss.Width(out); got > width {
			t.Fatalf("largura pedida %d: caixa de erro com %d colunas", width, got)
		}
	}
}

func TestErrorBoxBordaGrossaECantos(t *testing.T) {
	lines := strings.Split(ansi.Strip(ErrorBox(errorboxTestTheme(false), 40, "falha", "x", "y")), "\n")
	if !strings.HasPrefix(lines[0], "┏") || !strings.HasSuffix(lines[0], "┓") {
		t.Fatalf("topo da caixa de erro = %q, esperado ┏ e ┓", lines[0])
	}
	last := lines[len(lines)-1]
	if !strings.HasPrefix(last, "┗") || !strings.HasSuffix(last, "┛") {
		t.Fatalf("base da caixa de erro = %q, esperado ┗ e ┛", last)
	}
}

func TestErrorBoxTituloComEstadoErro(t *testing.T) {
	out := ansi.Strip(ErrorBox(errorboxTestTheme(false), 40, "falha", "x", "y"))
	if !strings.Contains(out, "✗ falha") {
		t.Fatalf("título sem símbolo de erro: %q", out)
	}
	plain := ansi.Strip(ErrorBox(errorboxTestTheme(true), 40, "falha", "x", "y"))
	if !strings.Contains(plain, "[x] falha") {
		t.Fatalf("título em Plain sem símbolo de erro: %q", plain)
	}
}
