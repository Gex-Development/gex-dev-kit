package main

import (
	"fmt"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/kit"
)

type doctorModel struct {
	statuses []kit.Status
	table    table.Model
}

func newDoctorModel(statuses []kit.Status) doctorModel {
	rows := make([]table.Row, 0, len(statuses))
	for _, status := range statuses {
		rows = append(rows, table.Row{status.State, status.Path})
	}
	styles := table.DefaultStyles()
	styles.Selected = lipgloss.NewStyle().Bold(true)
	model := table.New(
		table.WithColumns([]table.Column{{Title: "Estado", Width: 10}, {Title: "Arquivo", Width: 64}}),
		table.WithRows(rows),
		table.WithWidth(80),
		table.WithHeight(18),
		table.WithFocused(true),
		table.WithStyles(styles),
	)
	return doctorModel{statuses: statuses, table: model}
}

func (m doctorModel) Init() tea.Cmd { return nil }

func (m doctorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "q" || msg.String() == "esc" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		width := max(1, msg.Width)
		stateWidth := min(10, max(1, (width-4)/3))
		m.table.SetColumns([]table.Column{{Title: "Estado", Width: stateWidth}, {Title: "Arquivo", Width: max(1, width-stateWidth-4)}})
		m.table.SetWidth(width)
		m.table.SetHeight(max(3, msg.Height-4))
	}
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m doctorModel) View() tea.View {
	selected := m.table.Cursor()
	detail := "Nenhuma instalação registrada."
	if selected >= 0 && selected < len(m.statuses) {
		status := m.statuses[selected]
		detail = fmt.Sprintf("%s: %s", status.Path, status.Detail)
	}
	return tea.NewView(lipgloss.NewStyle().Bold(true).Render("Doctor GEX") + "\n" + m.table.View() + "\n" + detail + "\n↑/↓ navegar · q sair")
}
