package components

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func diffviewPlainLines(view string) []string {
	return strings.Split(ansi.Strip(view), "\n")
}

func diffviewWords(view string) string {
	return strings.Join(strings.Fields(ansi.Strip(view)), " ")
}

func TestDiffViewKeepsBlockedAgentVisibleOnNarrowTerminal(t *testing.T) {
	groups := []ChangeGroup{
		{Title: "Repositório", Changes: []Change{{Action: ChangeCreate, Path: ".claude/skills/very-long-skill-name/SKILL.md"}}},
		{Title: "Agentes", Changes: []Change{{Action: ChangeBlocked, Path: "claude 2.1.306", Detail: "outro gestor de pacotes. Ação manual: confira versão atual"}}},
	}
	view := DiffView(filetreeTestTheme(), 40, groups, nil)
	for _, line := range diffviewPlainLines(view) {
		if width := ansi.StringWidth(line); width > 40 {
			t.Fatalf("prévia excedeu 40 colunas (%d): %q", width, line)
		}
	}
	if text := diffviewWords(view); !strings.Contains(text, "bloqueado") || !strings.Contains(text, "Ação manual: confira versão atual") {
		t.Fatalf("prévia ocultou bloqueio e orientação: %s", view)
	}
}

func TestDiffViewPrefixesLinesAndKeepsHunkHeaders(t *testing.T) {
	change := Change{Action: ChangeUpdate, Path: "CLAUDE.md", Lines: []DiffLine{
		{Kind: DiffHunk, Text: "@@ -1,2 +1,2 @@"},
		{Kind: DiffContext, Text: "contexto"},
		{Kind: DiffRemoved, Text: "linha antiga"},
		{Kind: DiffAdded, Text: "linha nova"},
		{Kind: DiffHunk, Text: "-5 +5"},
	}}
	view := DiffView(filetreeTestTheme(), 60, []ChangeGroup{{Title: "HOME", Changes: []Change{change}}}, map[string]bool{"CLAUDE.md": true})
	want := []string{"@@ -1,2 +1,2 @@", " contexto", "-linha antiga", "+linha nova", "@@ -5 +5"}
	lines := diffviewPlainLines(view)
	for _, w := range want {
		if !diffviewContainsLine(lines, w) {
			t.Fatalf("linha %q ausente da prévia:\n%s", w, ansi.Strip(view))
		}
	}
}

func diffviewContainsLine(lines []string, want string) bool {
	for _, line := range lines {
		if line == want {
			return true
		}
	}
	return false
}

func TestDiffViewShowsOmittedAndMaskedNotes(t *testing.T) {
	change := Change{Action: ChangeUpdate, Path: "hooks.json", Lines: []DiffLine{{Kind: DiffAdded, Text: "x"}}, Omitted: 12, Masked: 3}
	view := ansi.Strip(DiffBody(filetreeTestTheme(), 80, change))
	for _, want := range []string{"… 12 linhas omitidas", "3 valores mascarados"} {
		if !strings.Contains(view, want) {
			t.Fatalf("nota %q ausente:\n%s", want, view)
		}
	}
	single := ansi.Strip(DiffBody(filetreeTestTheme(), 80, Change{Omitted: 1, Masked: 1}))
	for _, want := range []string{"… 1 linha omitida", "1 valor mascarado"} {
		if !strings.Contains(single, want) {
			t.Fatalf("nota singular %q ausente:\n%s", want, single)
		}
	}
}

func TestDiffViewNeverExceedsWidthWithLongContent(t *testing.T) {
	change := Change{
		Action:  ChangeUpdate,
		Path:    ".agents/skills/gex-muito-longa/referencias/arquivo-com-nome-enorme.md",
		Detail:  "+12 −3 com um detalhe muito longo que precisa quebrar em várias linhas",
		Lines:   []DiffLine{{Kind: DiffAdded, Text: strings.Repeat("abcdefghij", 9)}, {Kind: DiffRemoved, Text: "\tcom\ttabs\x1b[31m e escape\r"}, {Kind: DiffContext, Text: "日本語の文字列テキスト長い行"}},
		Omitted: 5,
		Masked:  2,
	}
	view := DiffView(filetreeTestTheme(), 30, []ChangeGroup{{Title: "Agentes e skills", Changes: []Change{change}}}, map[string]bool{change.Path: true})
	for _, line := range diffviewPlainLines(view) {
		if width := ansi.StringWidth(line); width > 30 {
			t.Fatalf("linha excedeu 30 colunas (%d): %q", width, line)
		}
	}
	if strings.Contains(view, "\x1b[31m") {
		t.Fatal("escape vindo do conteúdo do arquivo chegou à tela")
	}
}

func TestDiffViewShowsDiffOnlyForExpandedChanges(t *testing.T) {
	change := Change{Action: ChangeUpdate, Path: "a.md", Lines: []DiffLine{{Kind: DiffAdded, Text: "conteúdo oculto"}}}
	groups := []ChangeGroup{{Title: "Repositório", Changes: []Change{change}}}
	collapsed := ansi.Strip(DiffView(filetreeTestTheme(), 80, groups, nil))
	if strings.Contains(collapsed, "conteúdo oculto") {
		t.Fatalf("arquivo recolhido mostrou o diff:\n%s", collapsed)
	}
	if !strings.Contains(diffviewWords(collapsed), "atualizar a.md") {
		t.Fatalf("resumo do arquivo recolhido ausente:\n%s", collapsed)
	}
	expanded := ansi.Strip(DiffView(filetreeTestTheme(), 80, groups, map[string]bool{"a.md": true}))
	if !strings.Contains(expanded, "+conteúdo oculto") {
		t.Fatalf("arquivo expandido não mostrou o diff:\n%s", expanded)
	}
}

func TestDiffViewWithoutGroupsIsEmpty(t *testing.T) {
	if view := DiffView(filetreeTestTheme(), 40, nil, nil); view != "" {
		t.Fatalf("sem grupos a prévia deveria ser vazia: %q", view)
	}
}

func TestChangeLabelCoversEveryAction(t *testing.T) {
	want := map[ChangeAction]string{
		ChangeCreate:    "criar",
		ChangeUpdate:    "atualizar",
		ChangeRemove:    "remover",
		ChangeRestore:   "restaurar",
		ChangeBlocked:   "bloqueado",
		ChangeUnchanged: "sem mudança",
	}
	for action, label := range want {
		if got := ChangeLabel(action); got != label {
			t.Fatalf("rótulo da ação %d: %q, esperado %q", action, got, label)
		}
	}
	if got := ChangeLabel(ChangeAction(-1)); got != "sem mudança" {
		t.Fatalf("rótulo desconhecido: %q", got)
	}
}

func TestDiffviewPluralChoosesByCount(t *testing.T) {
	if got := diffviewPlural(1, "um", "muitos"); got != "um" {
		t.Fatalf("plural de 1: %q", got)
	}
	if got := diffviewPlural(0, "um", "muitos"); got != "muitos" {
		t.Fatalf("plural de 0: %q", got)
	}
}

func TestDiffviewTextRemovesEscapesTabsAndReturns(t *testing.T) {
	got := diffviewText("\x1b[31ma\x1b[0m\tb\r")
	if got != "a    b" {
		t.Fatalf("texto sanitizado errado: %q", got)
	}
}

func TestDiffviewLineUsesSemanticPrefixes(t *testing.T) {
	theme := filetreeTestTheme()
	cases := map[DiffLineKind]string{DiffAdded: "+x", DiffRemoved: "-x", DiffContext: " x", DiffHunk: "@@ x"}
	for kind, want := range cases {
		if got := ansi.Strip(diffviewLine(theme, 40, DiffLine{Kind: kind, Text: "x"})); got != want {
			t.Fatalf("prefixo do tipo %d: %q, esperado %q", kind, got, want)
		}
	}
	if got := lipgloss.Width(diffviewLine(theme, 3, DiffLine{Kind: DiffAdded, Text: "abcdef"})); got > 3 {
		t.Fatalf("linha de diff excedeu 3 colunas: %d", got)
	}
}
