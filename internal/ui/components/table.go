package components

import (
	bubbletable "charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	lgtable "charm.land/lipgloss/v2/table"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

type Column struct {
	Title string
	Width int
}

type Table struct {
	model   bubbletable.Model
	columns []Column
	padding int
}

func NewTable(t theme.Theme, columns []Column, rows [][]string, height int) Table {
	styles := t.Table()
	data := make([]bubbletable.Row, 0, len(rows))
	for _, row := range rows {
		data = append(data, bubbletable.Row(row))
	}
	x := Table{columns: columns, padding: styles.Cell.GetHorizontalFrameSize()}
	natural := tableNaturalWidth(columns, x.padding)
	x.model = bubbletable.New(
		bubbletable.WithColumns(tableColumns(columns, tableFitWidths(columns, x.padding, natural))),
		bubbletable.WithRows(data),
		bubbletable.WithFocused(true),
		bubbletable.WithStyles(styles),
	)
	return x.Resize(natural, height)
}

func (x Table) Update(msg tea.Msg) (Table, tea.Cmd) {
	var cmd tea.Cmd
	x.model, cmd = x.model.Update(msg)
	return x, cmd
}

func (x Table) Resize(width, height int) Table {
	widths := tableFitWidths(x.columns, x.padding, width)
	x.model.SetColumns(tableColumns(x.columns, widths))
	x.model.SetWidth(width)
	x.model.SetHeight(height)
	return x
}

func (x Table) Cursor() int {
	return x.model.Cursor()
}

func (x Table) View() string {
	return x.model.View()
}

func StaticTable(t theme.Theme, width int, headers []string, rows [][]string) string {
	return lgtable.New().
		Headers(headers...).
		Rows(rows...).
		Width(width).
		Border(lipgloss.NormalBorder()).
		BorderStyle(t.Divider).
		StyleFunc(func(row, _ int) lipgloss.Style {
			if row == lgtable.HeaderRow {
				return t.Title.Padding(0, 1)
			}
			return t.Text.Padding(0, 1)
		}).
		String()
}

func tableColumns(columns []Column, widths []int) []bubbletable.Column {
	out := make([]bubbletable.Column, len(columns))
	for i, column := range columns {
		out[i] = bubbletable.Column{Title: column.Title, Width: widths[i]}
	}
	return out
}

func tableNaturalWidth(columns []Column, padding int) int {
	total := 0
	for _, column := range columns {
		total += max(1, column.Width) + padding
	}
	return total
}

func tableFitWidths(columns []Column, padding, width int) []int {
	widths := make([]int, len(columns))
	visible := len(columns)
	for visible > 0 && width-padding*visible < visible {
		visible--
	}
	if visible == 0 {
		return widths
	}
	budget := width - padding*visible
	preferred := 0
	for i := range visible {
		widths[i] = max(1, columns[i].Width)
		preferred += widths[i]
	}
	if preferred <= budget {
		widths[visible-1] += budget - preferred
		return widths
	}
	for tableSum(widths) > budget {
		widest := 0
		for i := range widths {
			if widths[i] > widths[widest] {
				widest = i
			}
		}
		widths[widest]--
	}
	return widths
}

func tableSum(values []int) int {
	total := 0
	for _, value := range values {
		total += value
	}
	return total
}
