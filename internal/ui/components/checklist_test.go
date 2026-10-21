package components

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/spinner"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

func checklistTheme(plain bool) theme.Theme {
	return theme.New(theme.Options{Dark: true, Plain: plain, Profile: colorprofile.ASCII})
}

func checklistItems() []ChecklistItem {
	return []ChecklistItem{
		{Label: "Baixar pacote", State: theme.StateOK},
		{Label: "Instalar agentes", State: theme.StateRunning},
		{Label: "Configurar hooks", State: theme.StatePending},
		{Label: "Validar credenciais", State: theme.StateWarn, Detail: "GH_TOKEN ausente"},
		{Label: "Rede", State: theme.StateError},
	}
}

func TestChecklistEstaticoNaoAgendaTick(t *testing.T) {
	c := NewChecklist(checklistItems(), false)
	if c.Tick() != nil {
		t.Fatalf("checklist estático agendou Tick")
	}
	next, cmd := c.Update(spinner.TickMsg{})
	if cmd != nil {
		t.Fatalf("checklist estático devolveu comando após TickMsg")
	}
	th := checklistTheme(false)
	if next.View(th, 60) != c.View(th, 60) {
		t.Fatalf("checklist estático mudou a tela após TickMsg")
	}
	if got := ansi.Strip(c.View(th, 60)); !strings.Contains(got, "… rodando") {
		t.Fatalf("símbolo fixo de rodando ausente:\n%s", got)
	}
}

func TestChecklistAnimadoGiraOSpinner(t *testing.T) {
	th := checklistTheme(false)
	c := NewChecklist(checklistItems(), true)
	if c.Tick() == nil {
		t.Fatalf("checklist animado não devolveu Tick")
	}
	msg := c.Tick()()
	primeiro := ansi.Strip(c.View(th, 60))
	if !strings.Contains(primeiro, spinner.MiniDot.Frames[0]) {
		t.Fatalf("primeiro quadro sem o spinner:\n%s", primeiro)
	}
	proximo, cmd := c.Update(msg)
	if cmd == nil {
		t.Fatalf("checklist animado não agendou o próximo quadro")
	}
	segundo := ansi.Strip(proximo.View(th, 60))
	if !strings.Contains(segundo, spinner.MiniDot.Frames[1]) {
		t.Fatalf("segundo quadro não girou o spinner:\n%s", segundo)
	}
}

func TestChecklistSinalizaEstadoComSimboloETexto(t *testing.T) {
	casos := []struct {
		plain bool
		want  []string
	}{
		{false, []string{"✓ ok", "… rodando", "○ pendente", "! atenção", "✗ erro"}},
		{true, []string{"[ok] ok", "... rodando", "[ ] pendente", "[!] atenção", "[x] erro"}},
	}
	for _, caso := range casos {
		view := ansi.Strip(NewChecklist(checklistItems(), false).View(checklistTheme(caso.plain), 60))
		for _, want := range caso.want {
			if !strings.Contains(view, want) {
				t.Fatalf("Plain=%v: tela sem %q:\n%s", caso.plain, want, view)
			}
		}
	}
}

func TestChecklistRespeitaLargura(t *testing.T) {
	itens := []ChecklistItem{{
		Label:  "um rótulo longo demais para caber na largura pedida",
		State:  theme.StateOK,
		Detail: "detalhe igualmente longo que também precisa caber",
	}}
	view := NewChecklist(itens, false).View(checklistTheme(false), 20)
	for _, linha := range strings.Split(view, "\n") {
		if w := lipgloss.Width(linha); w > 20 {
			t.Fatalf("linha com %d colunas; largura pedida 20: %q", w, linha)
		}
	}
}

func TestChecklistSetNaoAlteraOValorAnterior(t *testing.T) {
	antigo := NewChecklist(checklistItems(), false)
	novo := antigo.Set(2, theme.StateOK, "pronto")
	if antigo.Items()[2].State != theme.StatePending {
		t.Fatalf("Set alterou o checklist anterior")
	}
	if item := novo.Items()[2]; item.State != theme.StateOK || item.Detail != "pronto" {
		t.Fatalf("Set não aplicou estado e detalhe: %+v", item)
	}
	if fora := antigo.Set(99, theme.StateOK, ""); fora.Items()[0] != antigo.Items()[0] {
		t.Fatalf("índice fora da lista alterou o checklist")
	}
	copia := novo.Items()
	copia[0].Label = "mudado"
	if novo.Items()[0].Label == "mudado" {
		t.Fatalf("Items devolveu o slice interno")
	}
}
