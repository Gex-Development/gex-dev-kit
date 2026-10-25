package screens

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/components"
)

func TestChatCatalogShowsSkillOrigin(t *testing.T) {
	m := chatTestModel(false).flowCatalog()
	m = chatSetupQuestion(t, m, "O que você quer ver?")
	m = chatSetupChoose(t, m, "skills")
	m = chatSetupQuestion(t, m, "Qual grupo")
	m = chatSetupChoose(t, m, "Core")
	m = chatSetupQuestion(t, m, "Qual item")
	if m.ask == nil || len(m.ask.options) == 0 {
		t.Fatal("catálogo sem skills")
	}
	name := m.ask.options[0].Label
	m = chatSetupChoose(t, m, m.ask.options[0].ID)
	transcript := chatTestTranscript(m)
	if !strings.Contains(transcript, name) || !strings.Contains(transcript, "Core") {
		t.Fatalf("skill e origem ausentes: %s", transcript)
	}
}

func TestChatCatalogDetailExplainsCoreSkill(t *testing.T) {
	m := chatTestModel(false).flowCatalog()
	m = chatSetupQuestion(t, m, "O que você quer ver?")
	m = chatSetupChoose(t, m, "skills")
	m = chatSetupQuestion(t, m, "Qual grupo")
	m = chatSetupChoose(t, m, "Core")
	m = chatSetupQuestion(t, m, "Qual item")
	m = chatSetupChoose(t, m, "gex-abrir-issue")
	transcript := chatTestTranscript(m)
	for _, label := range []string{"O que faz:", "Quando usar:", "Analogia:"} {
		if !strings.Contains(transcript, label) {
			t.Fatalf("%s ausente: %s", label, transcript)
		}
	}
}

func TestChatCatalogReferenceKeepsOldDetail(t *testing.T) {
	items := chatCatalogRefItems()
	if len(items) == 0 {
		t.Fatal("referências ausentes")
	}
	detail := chatCatalogDetail(items[0])
	if strings.Contains(detail, "O que faz:") || !strings.Contains(detail, "Fonte: ") {
		t.Fatalf("formato de referência mudou: %s", detail)
	}
}

func TestChatCatalogGroupsAndEscape(t *testing.T) {
	m := chatTestModel(false).flowCatalog()
	m = chatSetupQuestion(t, m, "O que você quer ver?")
	m = chatSetupChoose(t, m, "skills")
	m = chatSetupQuestion(t, m, "Qual grupo")
	for _, name := range []string{"Core", "Engenharia", "Produtividade", "Especializações"} {
		if !strings.Contains(m.ask.choice.View(m.theme, 80), name) {
			t.Fatalf("grupo %s ausente", name)
		}
	}
	m = chatSetupChoose(t, m, "Especializações")
	m = chatSetupQuestion(t, m, "Qual grupo")
	m = chatSetupChoose(t, m, "Dados e Plataforma")
	m = chatSetupQuestion(t, m, "Qual item")
	if len(m.ask.options) > 8 {
		t.Fatalf("subgrupo longo: %d", len(m.ask.options))
	}
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyEscape})
	if m.ask == nil || !strings.Contains(m.ask.question, "Qual grupo") {
		t.Fatal("esc não voltou para subgrupos")
	}
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyEscape})
	if m.ask == nil || !strings.Contains(m.ask.choice.View(m.theme, 80), "Engenharia") {
		t.Fatal("esc não voltou aos grupos")
	}
}

func TestChatCatalogReferencesSplit(t *testing.T) {
	m := chatTestModel(false).flowCatalog()
	m = chatSetupQuestion(t, m, "O que você quer ver?")
	m = chatSetupChoose(t, m, "refs")
	m = chatSetupQuestion(t, m, "Qual grupo")
	for _, name := range []string{"Skills de terceiros", "Hooks candidatos"} {
		if !strings.Contains(m.ask.choice.View(m.theme, 80), name) {
			t.Fatalf("grupo %s ausente", name)
		}
	}
	m = chatSetupChoose(t, m, "Skills de terceiros")
	m = chatSetupQuestion(t, m, "Qual item")
	if len(m.ask.options) > 12 {
		t.Fatalf("referências não divididas: %d", len(m.ask.options))
	}
}

func TestChatAddonsShowsOfficialClaudeCommand(t *testing.T) {
	m := chatTestModel(false).flowAddons()
	m = chatSetupQuestion(t, m, "Qual addon")
	m, _ = chatTestUpdate(m, components.ChoiceDoneMsg{IDs: []string{"caveman"}})
	if !strings.Contains(chatTestTranscript(m), "Comando oficial no Claude Code") {
		t.Fatal("comando oficial ausente")
	}
}
