package components

import (
	"slices"
	"strings"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/list"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

type ChecklistItem struct {
	Label  string
	State  theme.State
	Detail string
}

type Checklist struct {
	items    []ChecklistItem
	animated bool
	spin     spinner.Model
}

func NewChecklist(items []ChecklistItem, animated bool) Checklist {
	return Checklist{
		items:    slices.Clone(items),
		animated: animated,
		spin:     spinner.New(spinner.WithSpinner(spinner.MiniDot)),
	}
}

func (c Checklist) Tick() tea.Cmd {
	if !c.animated {
		return nil
	}
	return c.spin.Tick
}

func (c Checklist) Set(i int, s theme.State, detail string) Checklist {
	if i < 0 || i >= len(c.items) {
		return c
	}
	c.items = slices.Clone(c.items)
	c.items[i].State = s
	c.items[i].Detail = detail
	return c
}

func (c Checklist) Items() []ChecklistItem {
	return slices.Clone(c.items)
}

func (c Checklist) Update(msg tea.Msg) (Checklist, tea.Cmd) {
	tick, ok := msg.(spinner.TickMsg)
	if !ok || !c.animated {
		return c, nil
	}
	var cmd tea.Cmd
	c.spin, cmd = c.spin.Update(tick)
	return c, cmd
}

func (c Checklist) View(t theme.Theme, width int) string {
	marks := make([]string, len(c.items))
	rows := make([]any, len(c.items))
	markWidth := 0
	for i, item := range c.items {
		marks[i] = c.mark(t, item.State)
		markWidth = max(markWidth, lipgloss.Width(marks[i]))
		rows[i] = checklistRow(t, item)
	}
	enumerate := func(_ list.Items, i int) string {
		return marks[i] + strings.Repeat(" ", markWidth-lipgloss.Width(marks[i]))
	}
	body := list.New(rows...).Enumerator(enumerate).String()
	return t.Text.MaxWidth(width).Render(body)
}

func (c Checklist) mark(t theme.Theme, s theme.State) string {
	if s == theme.StateRunning {
		return c.runningSymbol(t) + " " + t.Spinner().Render(theme.StateRunning.Label())
	}
	return t.StateStyle(s).Render("")
}

func (c Checklist) runningSymbol(t theme.Theme) string {
	if !c.animated || t.Plain {
		return t.Spinner().Render(checklistStaticSymbol(t))
	}
	frame := c.spin
	frame.Style = t.Spinner()
	return frame.View()
}

func checklistStaticSymbol(t theme.Theme) string {
	if t.Plain {
		return "..."
	}
	return "…"
}

func checklistRow(t theme.Theme, item ChecklistItem) string {
	label := t.Text.Render(item.Label)
	if item.Detail == "" {
		return label
	}
	return label + "\n" + t.Secondary.Render(item.Detail)
}
