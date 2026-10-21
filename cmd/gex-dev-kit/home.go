package main

import (
	"io"
	"strings"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/kit"
	"github.com/charmbracelet/x/ansi"
)

type homeAction struct {
	name, title, description string
}

func (a homeAction) Title() string       { return a.title }
func (a homeAction) Description() string { return a.description }
func (a homeAction) FilterValue() string { return a.title }

type homeStatus struct {
	statuses []kit.Status
	err      error
}

type homeModel struct {
	home     string
	choice   string
	list     list.Model
	delegate list.DefaultDelegate
	spinner  spinner.Model
	loading  bool
	status   string
	width    int
	height   int
	dark     bool
}

func newHomeModel(home string) homeModel {
	items := []list.Item{
		homeAction{"install", "Instalar Core", "Prévia e instalação guiada"},
		homeAction{"doctor", "Doctor", "Inspecionar estado do kit"},
		homeAction{"rollback", "Reverter Core", "Restaurar arquivos do kit"},
		homeAction{"rollback-agents", "Reverter agentes", "Restaurar versões anteriores"},
		homeAction{"exit", "Sair", "Fechar o Kit GEX"},
	}
	delegate := homeDelegate(true)
	menu := list.New(items, delegate, 38, 10)
	menu.SetShowTitle(false)
	menu.SetShowStatusBar(false)
	menu.SetShowPagination(false)
	menu.SetShowHelp(false)
	menu.SetFilteringEnabled(false)
	menu.SetSize(38, 10)
	return homeModel{
		home: home, list: menu, delegate: delegate,
		spinner: spinner.New(spinner.WithSpinner(spinner.MiniDot), spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#7fc3ef")))),
		loading: true, status: "Lendo estado do Core...", width: 40, height: 18, dark: true,
	}
}

func homeDelegate(isDark bool) list.DefaultDelegate {
	choose := lipgloss.LightDark(isDark)
	delegate := list.NewDefaultDelegate()
	delegate.SetSpacing(0)
	delegate.Styles.NormalTitle = lipgloss.NewStyle().Foreground(choose(lipgloss.Color("#0b1526"), lipgloss.Color("#f4f4f5"))).PaddingLeft(2)
	delegate.Styles.NormalDesc = lipgloss.NewStyle().Foreground(choose(lipgloss.Color("#4b5669"), lipgloss.Color("#a1a1aa"))).PaddingLeft(2)
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(choose(lipgloss.Color("#1a5fc0"), lipgloss.Color("#7fc3ef"))).Bold(true).Border(lipgloss.Border{Left: "▸"}, false, false, false, true).PaddingLeft(1)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().Foreground(choose(lipgloss.Color("#1a5fc0"), lipgloss.Color("#7fc3ef"))).PaddingLeft(2)
	return delegate
}

func scanHome(home string) tea.Cmd {
	return func() tea.Msg {
		statuses, err := kit.Doctor(home)
		return homeStatus{statuses: statuses, err: err}
	}
}

func (m homeModel) Init() tea.Cmd {
	return tea.Batch(scanHome(m.home), m.spinner.Tick, tea.RequestBackgroundColor)
}

func (m homeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = max(1, msg.Width), max(1, msg.Height)
		delegate := m.delegate
		delegate.ShowDescription = m.height >= 16 && m.width >= 40
		m.list.SetDelegate(delegate)
		if !delegate.ShowDescription {
			m.list.SetSize(m.width, max(1, m.height-3))
		} else {
			m.list.SetSize(max(1, min(m.width, 72)-2), max(2, min(10, m.height-8)))
		}
	case tea.BackgroundColorMsg:
		m.dark = msg.IsDark()
		m.delegate = homeDelegate(m.dark)
		m.delegate.ShowDescription = m.height >= 16 && m.width >= 40
		m.list.SetDelegate(m.delegate)
		m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.LightDark(m.dark)(lipgloss.Color("#1a5fc0"), lipgloss.Color("#7fc3ef")))
	case homeStatus:
		m.loading = false
		if msg.err != nil {
			m.status = "Core: erro na leitura: " + msg.err.Error()
			break
		}
		m.status = "Core: íntegro"
		if len(msg.statuses) == 0 {
			m.status = "Core: nenhuma instalação registrada"
		}
		for _, status := range msg.statuses {
			switch status.State {
			case "alterado":
				m.status = "Core: alterado · execute doctor"
				return m, nil
			case "ausente", "incompleto":
				m.status = "Core: " + status.State + " · execute doctor"
			case "sem-historico":
				if m.status == "Core: íntegro" {
					m.status = "Core: nenhuma instalação registrada"
				}
			case "pendente":
				if m.status == "Core: íntegro" {
					m.status = "Core: pendente · confirme hook em /hooks"
				}
			}
		}
	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "q", "ctrl+c":
			m.choice = "exit"
			return m, tea.Quit
		case "enter":
			if selected, ok := m.list.SelectedItem().(homeAction); ok {
				m.choice = selected.name
				return m, tea.Quit
			}
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m homeModel) View() tea.View {
	width := max(1, min(m.width, 72))
	choose := lipgloss.LightDark(m.dark)
	blue := lipgloss.NewStyle().Foreground(choose(lipgloss.Color("#1a5fc0"), lipgloss.Color("#7fc3ef"))).Bold(true)
	muted := lipgloss.NewStyle().Foreground(choose(lipgloss.Color("#4b5669"), lipgloss.Color("#a1a1aa")))
	heading := blue.Render("GEX / DEV KIT") + "  " + muted.Render(version)
	label := m.status
	if m.loading {
		label = m.spinner.View() + " " + label
	}
	if m.height < 16 || m.width < 40 {
		if m.height < 4 {
			if selected, ok := m.list.SelectedItem().(homeAction); ok {
				return tea.NewView(ansi.Truncate(selected.title, width, "…"))
			}
			return tea.NewView(ansi.Truncate(heading, width, "…"))
		}
		return tea.NewView(strings.Join([]string{
			ansi.Truncate(heading, width, "…"),
			ansi.Truncate(label, width, "…"),
			strings.TrimSuffix(m.list.View(), "\n"),
			ansi.Truncate(muted.Render("↑↓ navegar · Enter · q sair"), width, "…"),
		}, "\n"))
	}
	panel := lipgloss.NewStyle().Width(max(1, width-2)).
		Border(lipgloss.RoundedBorder()).BorderForeground(choose(lipgloss.Color("#1e6fd9"), lipgloss.Color("#7fc3ef"))).
		Render(m.list.View())
	content := strings.Join([]string{
		ansi.Truncate(heading, width, "…"),
		blue.Render("│ CORE OBRIGATÓRIO"),
		ansi.Truncate("│ "+label, width, "…"),
		ansi.Truncate(muted.Render("  SKILLS opcionais  ·  HOOK opcional"), width, "…"),
		blue.Render("ESCOLHA UMA AÇÃO"),
		panel,
		ansi.Truncate(muted.Render("Enter selecionar · ↑↓ navegar · q sair"), width, "…"),
	}, "\n")
	return tea.NewView(content)
}

func showHome(in io.Reader, out io.Writer, home string) (string, error) {
	model, err := tea.NewProgram(newHomeModel(home), tea.WithInput(in), tea.WithOutput(out)).Run()
	if err != nil {
		return "", err
	}
	return model.(homeModel).choice, nil
}
