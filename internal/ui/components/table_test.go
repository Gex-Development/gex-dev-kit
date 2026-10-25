package components

import (
	"reflect"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func tableTestColumns() []Column {
	return []Column{{Title: "Estado", Width: 10}, {Title: "Arquivo", Width: 64}}
}

func tableTestRows() [][]string {
	return [][]string{
		{"alterado", ".claude/CLAUDE.md"},
		{"ok", ".agents/skills/gex-muito-longa/referencias/arquivo-com-nome-enorme.md"},
	}
}

func TestTableResizeKeepsColumnsInsideWidth(t *testing.T) {
	for _, width := range []int{40, 80} {
		table := NewTable(filetreeTestTheme(), tableTestColumns(), tableTestRows(), 9).Resize(width, 9)
		if got := lipgloss.Width(table.View()); got > width {
			t.Fatalf("tabela de %d colunas ficou com %d colunas", width, got)
		}
		used := 0
		for _, column := range table.model.Columns() {
			used += column.Width + table.padding
		}
		if used > width {
			t.Fatalf("colunas ocupam %d em largura %d: %v", used, width, table.model.Columns())
		}
	}
}

func TestTableResizeKeepsPathVisibleOnNarrowTerminal(t *testing.T) {
	table := NewTable(filetreeTestTheme(), tableTestColumns(), tableTestRows(), 9).Resize(40, 9)
	if view := ansi.Strip(table.View()); !strings.Contains(view, ".claude/CLAUDE.md") {
		t.Fatalf("caminho sumiu da tabela de 40 colunas:\n%s", view)
	}
}

func TestTableResizeExpandsLastColumnOnWideTerminal(t *testing.T) {
	table := NewTable(filetreeTestTheme(), tableTestColumns(), tableTestRows(), 9).Resize(100, 9)
	widths := []int{}
	for _, column := range table.model.Columns() {
		widths = append(widths, column.Width)
	}
	if want := []int{10, 86}; !reflect.DeepEqual(widths, want) {
		t.Fatalf("larguras em 100 colunas: %v, esperado %v", widths, want)
	}
}

func TestTableNewTableUsesNaturalWidth(t *testing.T) {
	table := NewTable(filetreeTestTheme(), []Column{{Title: "a", Width: 4}, {Title: "b", Width: 6}}, nil, 5)
	if got := table.model.Width(); got != 14 {
		t.Fatalf("largura natural da tabela: %d, esperado 14", got)
	}
}

func TestTableHidesTrailingColumnsWhenTooNarrow(t *testing.T) {
	table := NewTable(filetreeTestTheme(), tableTestColumns(), tableTestRows(), 9).Resize(5, 9)
	columns := table.model.Columns()
	if columns[0].Width != 3 || columns[1].Width != 0 {
		t.Fatalf("colunas em largura 5: %v", columns)
	}
	if got := lipgloss.Width(table.View()); got > 5 {
		t.Fatalf("tabela de 5 colunas ficou com %d", got)
	}
}

func TestTableCursorMovesWithKeys(t *testing.T) {
	table := NewTable(filetreeTestTheme(), tableTestColumns(), tableTestRows(), 9)
	table, _ = table.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	if table.Cursor() != 1 {
		t.Fatalf("cursor após j: %d, esperado 1", table.Cursor())
	}
	table, _ = table.Update(tea.KeyPressMsg{Code: 'k', Text: "k"})
	if table.Cursor() != 0 {
		t.Fatalf("cursor após k: %d, esperado 0", table.Cursor())
	}
}

func TestTableFitWidthsGivesRemainderToLastColumn(t *testing.T) {
	got := tableFitWidths([]Column{{Width: 4}, {Width: 6}}, 2, 30)
	if want := []int{4, 22}; !reflect.DeepEqual(got, want) {
		t.Fatalf("sobra não foi para a última coluna: %v, esperado %v", got, want)
	}
}

func TestTableFitWidthsShrinksWidestColumnsFirst(t *testing.T) {
	got := tableFitWidths([]Column{{Width: 10}, {Width: 30}}, 2, 24)
	if want := []int{10, 10}; !reflect.DeepEqual(got, want) {
		t.Fatalf("encolhimento das colunas largas: %v, esperado %v", got, want)
	}
	doctor := tableFitWidths(tableTestColumns(), 2, 40)
	if want := []int{10, 26}; !reflect.DeepEqual(doctor, want) {
		t.Fatalf("coluna de estado encolheu em 40 colunas: %v, esperado %v", doctor, want)
	}
}

func TestTableFitWidthsNeverExceedsBudgetWhenColumnsAreMinimal(t *testing.T) {
	got := tableFitWidths([]Column{{Width: 1}, {Width: 1}, {Width: 100}}, 2, 9)
	if want := []int{1, 1, 1}; !reflect.DeepEqual(got, want) {
		t.Fatalf("colunas mínimas em largura 9: %v, esperado %v", got, want)
	}
}

func TestTableFitWidthsWithoutRoomReturnsZeroWidths(t *testing.T) {
	got := tableFitWidths(tableTestColumns(), 2, 0)
	if want := []int{0, 0}; !reflect.DeepEqual(got, want) {
		t.Fatalf("sem largura as colunas deveriam ficar ocultas: %v", got)
	}
}

func TestStaticTableFitsWidth(t *testing.T) {
	headers := []string{"Estado", "Arquivo"}
	rows := [][]string{{"alterado", ".agents/skills/gex-muito-longa/referencias/arquivo-com-nome-enorme.md"}}
	for _, width := range []int{40, 80} {
		view := StaticTable(filetreeTestTheme(), width, headers, rows)
		if got := lipgloss.Width(view); got > width {
			t.Fatalf("tabela estática de %d colunas ficou com %d", width, got)
		}
		if plain := ansi.Strip(view); !strings.Contains(plain, "Estado") || !strings.Contains(plain, "alterado") {
			t.Fatalf("tabela estática perdeu cabeçalho ou linha:\n%s", plain)
		}
	}
}
