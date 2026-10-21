package main

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/kit"
	"github.com/charmbracelet/x/ansi"
)

type doctorModel struct {
	statuses       []kit.Status
	table          table.Model
	width          int
	height         int
	detailLines    []string
	detailOffset   int
	summary        string
	compactSummary string
	summaryColor   color.Color
	attention      int
	pending        int
	dark           bool
}

func newDoctorModel(statuses []kit.Status) doctorModel {
	rows := make([]table.Row, 0, len(statuses))
	for _, status := range statuses {
		rows = append(rows, table.Row{status.State, status.Path})
	}
	styles := table.DefaultStyles()
	model := table.New(
		table.WithColumns([]table.Column{{Title: "Estado", Width: 10}, {Title: "Arquivo", Width: 64}}),
		table.WithRows(rows),
		table.WithWidth(80),
		table.WithHeight(max(2, min(len(statuses)+1, 12))),
		table.WithFocused(true),
		table.WithStyles(styles),
	)
	ok, attention, pending := 0, 0, 0
	for _, status := range statuses {
		switch status.State {
		case "ok":
			ok++
		case "pendente", "sem-historico":
			pending++
		default:
			attention++
		}
	}

	m := doctorModel{
		statuses: statuses, table: model, width: 80, height: 24,
		summary:        fmt.Sprintf("%d íntegros · %d atenção · %d pendentes", ok, attention, pending),
		compactSummary: fmt.Sprintf("%d OK · %d ! · %d ?", ok, attention, pending),
		attention:      attention, pending: pending,
	}
	m.setColors(true)
	m.wrapDetail()
	return m
}

func (m *doctorModel) setColors(isDark bool) {
	m.dark = isDark
	choose := lipgloss.LightDark(isDark)
	accent := choose(lipgloss.Color("#1a5fc0"), lipgloss.Color("#7fc3ef"))
	styles := table.DefaultStyles()
	styles.Header = styles.Header.Foreground(accent).Bold(true)
	styles.Selected = lipgloss.NewStyle().Bold(true).
		Foreground(choose(lipgloss.Color("#0b1526"), lipgloss.Color("#f4f4f5"))).
		Background(choose(lipgloss.Color("#e8f0fc"), lipgloss.Color("#14202e")))
	m.table.SetStyles(styles)
	m.summaryColor = accent
	if m.attention > 0 {
		m.summaryColor = choose(lipgloss.Color("#b91c1c"), lipgloss.Color("#fb7185"))
	} else if m.pending > 0 {
		m.summaryColor = choose(lipgloss.Color("#92400e"), lipgloss.Color("#fcd34d"))
	}
}

func (m doctorModel) Init() tea.Cmd { return tea.RequestBackgroundColor }

func (m doctorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		m.setColors(msg.IsDark())
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "left":
			m.detailOffset = max(0, m.detailOffset-max(1, m.detailHeight()-1))
			return m, nil
		case "right":
			m.detailOffset = min(max(0, len(m.detailLines)-m.detailHeight()), m.detailOffset+max(1, m.detailHeight()-1))
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.width = max(1, msg.Width)
		m.height = max(1, msg.Height)
		stateWidth := min(10, max(1, (m.width-4)/3))
		m.table.SetColumns([]table.Column{{Title: "Estado", Width: stateWidth}, {Title: "Arquivo", Width: max(1, m.width-stateWidth-4)}})
		m.table.SetWidth(m.width)
		m.table.SetHeight(max(2, min(len(m.statuses)+1, m.height/2, m.height-6)))
		m.wrapDetail()
	}
	before := m.table.Cursor()
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	if m.table.Cursor() != before {
		m.detailOffset = 0
		m.wrapDetail()
	}
	return m, cmd
}

func (m *doctorModel) wrapDetail() {
	detail := "Nenhuma instalação registrada."
	if selected := m.table.Cursor(); selected >= 0 && selected < len(m.statuses) {
		status := m.statuses[selected]
		detail = fmt.Sprintf("%s: %s", status.Path, status.Detail)
	}
	m.detailLines = strings.Split(ansi.Hardwrap(detail, m.width, false), "\n")
	m.detailOffset = min(m.detailOffset, max(0, len(m.detailLines)-m.detailHeight()))
}

func (m doctorModel) detailHeight() int {
	if m.width < 32 || m.height < 8 {
		return max(1, m.height-3)
	}
	return max(1, m.height-m.table.Height()-5)
}

func (m doctorModel) View() tea.View {
	start := min(m.detailOffset, max(0, len(m.detailLines)-m.detailHeight()))
	end := min(start+m.detailHeight(), len(m.detailLines))
	if m.width < 32 || m.height < 8 {
		blue := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.LightDark(m.dark)(lipgloss.Color("#1a5fc0"), lipgloss.Color("#7fc3ef")))
		lines := []string{
			ansi.Truncate(blue.Render("GEX / DOCTOR"), m.width, "…"),
			ansi.Truncate(m.compactSummary, m.width, "…"),
		}
		if m.height > 3 {
			lines = append(lines, m.detailLines[start:end]...)
			lines = append(lines, ansi.Truncate("←/→ detalhe · q", m.width, "…"))
		}
		return tea.NewView(strings.Join(lines, "\n"))
	}
	footer := "↑/↓ itens · ←/→ detalhe · q sair"
	if m.width < lipgloss.Width(footer) {
		footer = "↑↓ · ←→ · q"
	}
	summary := m.summary
	if m.width < lipgloss.Width(summary) {
		summary = m.compactSummary
	}
	content := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.LightDark(m.dark)(lipgloss.Color("#1a5fc0"), lipgloss.Color("#7fc3ef"))).Render("GEX / DOCTOR") + "\n" +
		lipgloss.Wrap(lipgloss.NewStyle().Foreground(m.summaryColor).Render(summary), m.width, "") + "\n" +
		m.table.View() + "\n" +
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.LightDark(m.dark)(lipgloss.Color("#1a5fc0"), lipgloss.Color("#7fc3ef"))).Render(fmt.Sprintf("DETALHE · %d–%d/%d", start+1, end, len(m.detailLines))) + "\n" +
		strings.Join(m.detailLines[start:end], "\n") + "\n" +
		lipgloss.Wrap(footer, m.width, "")
	return tea.NewView(content)
}
