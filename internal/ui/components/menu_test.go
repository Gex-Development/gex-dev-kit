package components

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/motion"
)

func menuTestItems() []MenuItem {
	return []MenuItem{
		{ID: "a", Title: "Alfa"},
		{ID: "b", Title: "Beta", Disabled: true},
		{ID: "c", Title: "Gama", Description: "terceiro"},
		{ID: "d", Title: "Delta"},
	}
}

func TestMenuInicialPulaItemDesabilitado(t *testing.T) {
	m := NewMenu([]MenuItem{{ID: "x", Disabled: true}, {ID: "y"}}, false)
	if m.Index() != 1 || m.Selected().ID != "y" {
		t.Fatalf("cursor inicial deveria ir para o primeiro item habilitado; índice %d", m.Index())
	}
}

func TestMenuTecladoMoveEscolhe(t *testing.T) {
	m := NewMenu(menuTestItems(), false)
	m, _ = m.Update(keysPressFor("down"))
	if m.Index() != 2 {
		t.Fatalf("↓ deveria pular o item desabilitado; índice %d", m.Index())
	}
	m, _ = m.Update(keysPressFor("k"))
	if m.Index() != 0 {
		t.Fatalf("k deveria subir para o índice 0; índice %d", m.Index())
	}
	m, _ = m.Update(keysPressFor("up"))
	if m.Index() != 0 {
		t.Fatalf("↑ no topo deveria ficar em 0; índice %d", m.Index())
	}
	m, _ = m.Update(keysPressFor("j"))
	m, _ = m.Update(keysPressFor("down"))
	m, _ = m.Update(keysPressFor("down"))
	if m.Index() != 3 {
		t.Fatalf("depois de j e ↓ deveria estar no último item; índice %d", m.Index())
	}
	m, cmd := m.Update(keysPressFor("enter"))
	if cmd == nil {
		t.Fatalf("enter sobre item habilitado deveria devolver escolha")
	}
	escolha, ok := cmd().(MenuChoiceMsg)
	if !ok || escolha.Index != 3 || escolha.Item.ID != "d" {
		t.Fatalf("escolha inesperada: %#v", cmd())
	}
}

func TestMenuVazioNaoEscolhe(t *testing.T) {
	m := NewMenu(nil, false)
	if _, cmd := m.Update(keysPressFor("enter")); cmd != nil {
		t.Fatalf("menu vazio não deveria devolver escolha")
	}
	if m.Selected() != (MenuItem{}) {
		t.Fatalf("menu vazio deveria devolver item zerado")
	}
}

func TestMenuCliqueSelecionaSoItemValido(t *testing.T) {
	m := NewMenu(menuTestItems(), false).Place(4, 2)
	m, _ = m.Update(tea.MouseClickMsg{X: 5, Y: 5, Button: tea.MouseLeft})
	if m.Index() != 3 {
		t.Fatalf("clique na linha 3 deveria selecionar o índice 3; índice %d", m.Index())
	}
	m, _ = m.Update(tea.MouseClickMsg{X: 5, Y: 3, Button: tea.MouseLeft})
	if m.Index() != 3 {
		t.Fatalf("clique em item desabilitado não deveria mudar a seleção; índice %d", m.Index())
	}
	m, _ = m.Update(tea.MouseClickMsg{X: 5, Y: 1, Button: tea.MouseLeft})
	if m.Index() != 3 {
		t.Fatalf("clique acima do menu não deveria mudar a seleção; índice %d", m.Index())
	}
	m, _ = m.Update(tea.MouseClickMsg{X: 5, Y: 4, Button: tea.MouseRight})
	if m.Index() != 3 {
		t.Fatalf("clique com botão direito não deveria mudar a seleção; índice %d", m.Index())
	}
	m, _ = m.Update(tea.MouseClickMsg{X: 3, Y: 2, Button: tea.MouseLeft})
	if m.Index() != 3 {
		t.Fatalf("clique à esquerda do menu não deveria mudar a seleção; índice %d", m.Index())
	}
}

func TestMenuRodaDoMouseMove(t *testing.T) {
	m := NewMenu(menuTestItems(), false)
	m, _ = m.Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	if m.Index() != 2 {
		t.Fatalf("roda para baixo deveria ir ao índice 2; índice %d", m.Index())
	}
	m, _ = m.Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	if m.Index() != 3 {
		t.Fatalf("segunda roda para baixo deveria ir ao índice 3; índice %d", m.Index())
	}
	m, _ = m.Update(tea.MouseWheelMsg{Button: tea.MouseWheelUp})
	if m.Index() != 2 {
		t.Fatalf("roda para cima deveria voltar ao índice 2; índice %d", m.Index())
	}
}

func TestMenuMovimentoLigadoPersegueEAssenta(t *testing.T) {
	m := NewMenu(menuTestItems(), true)
	m, cmd := m.Update(keysPressFor("down"))
	if cmd == nil || !m.Animating() {
		t.Fatalf("com movimento ligado, mudar a seleção deveria agendar quadros")
	}
	m, cmd = m.Update(motion.FrameMsg{ID: m.frame})
	if v := m.spring.Value(); v <= 0 || v >= 2 {
		t.Fatalf("depois do primeiro quadro a mola deveria estar entre 0 e 2; valor %v", v)
	}
	quadros := 1
	for m.Animating() && quadros < 300 {
		if cmd == nil {
			t.Fatalf("mola em movimento sem quadro agendado no quadro %d", quadros)
		}
		m, cmd = m.Update(motion.FrameMsg{ID: m.frame})
		quadros++
	}
	if m.Animating() {
		t.Fatalf("a mola não assentou em 300 quadros")
	}
	if v := m.spring.Value(); v != 2 {
		t.Fatalf("a mola assentou em %v; quero o alvo 2", v)
	}
	if cmd != nil {
		t.Fatalf("mola assentada não deveria pedir mais quadros")
	}
}

func TestMenuMovimentoDesligadoSaltaParaAlvo(t *testing.T) {
	m := NewMenu(menuTestItems(), false)
	m, cmd := m.Update(keysPressFor("down"))
	if cmd != nil || m.Animating() {
		t.Fatalf("com movimento desligado não deveria agendar quadros")
	}
	if v := m.spring.Value(); v != 2 {
		t.Fatalf("com movimento desligado o realce deveria saltar ao alvo 2; valor %v", v)
	}
	if _, cmd := m.Update(motion.FrameMsg{ID: m.frame}); cmd != nil {
		t.Fatalf("quadro avulso com movimento desligado não deveria agendar outro")
	}
}

func TestMenuViewCabeNaLargura(t *testing.T) {
	m := NewMenu(menuTestItems(), false)
	tema := keysTheme()
	for largura := 1; largura <= 40; largura++ {
		for _, linha := range strings.Split(m.View(tema, largura, true), "\n") {
			if w := lipgloss.Width(linha); w > largura {
				t.Fatalf("largura pedida %d, linha com %d colunas: %q", largura, w, ansi.Strip(linha))
			}
		}
	}
}

func TestMenuMarcaItemAtivoEMostraDescricaoSoQuandoPedida(t *testing.T) {
	m := NewMenu(menuTestItems(), false)
	tema := keysTheme()
	comDescricao := ansi.Strip(m.View(tema, 60, true))
	if !strings.Contains(comDescricao, "▸ Alfa") {
		t.Fatalf("item ativo deveria ter o marcador ▸: %q", comDescricao)
	}
	if !strings.Contains(comDescricao, "terceiro") {
		t.Fatalf("descrição pedida não apareceu: %q", comDescricao)
	}
	semDescricao := ansi.Strip(m.View(tema, 60, false))
	if strings.Contains(semDescricao, "terceiro") {
		t.Fatalf("descrição apareceu sem ser pedida: %q", semDescricao)
	}
}
