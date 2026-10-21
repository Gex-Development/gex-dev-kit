package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/catalog"
	"github.com/Gex-Development/gex-dev-kit/internal/kit"
	"github.com/charmbracelet/x/ansi"
)

type homeAction struct {
	name, title, description string
}

func (a homeAction) Title() string       { return a.title }
func (a homeAction) Description() string { return a.description }
func (a homeAction) FilterValue() string { return a.title }

var homeActions = []homeAction{
	{"install", "Instalar Core", "Prévia e instalação guiada"},
	{"doctor", "Doctor", "Inspecionar estado do kit"},
	{"rollback", "Reverter Core", "Restaurar arquivos do kit"},
	{"rollback-agents", "Reverter agentes", "Restaurar versões anteriores"},
	{"exit", "Sair", "Fechar o Kit GEX"},
}

type homeStatus struct {
	statuses []kit.Status
	err      error
}

type homeModel struct {
	home       string
	choice     string
	list       list.Model
	delegate   list.DefaultDelegate
	spinner    spinner.Model
	loading    bool
	status     string
	width      int
	height     int
	dark       bool
	help       bool
	checked    int
	attention  int
	activeCore int
}

func newHomeModel(home string) homeModel {
	items := make([]list.Item, len(homeActions))
	for i, action := range homeActions {
		items[i] = action
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
		delegate.ShowDescription = m.height >= 25 && m.width >= 40
		m.list.SetDelegate(delegate)
		if !delegate.ShowDescription {
			m.list.SetSize(m.width, max(1, m.height-5))
		} else {
			m.list.SetSize(max(1, min(m.width, 76)-4), min(10, m.height-15))
		}
	case tea.BackgroundColorMsg:
		m.dark = msg.IsDark()
		m.delegate = homeDelegate(m.dark)
		m.delegate.ShowDescription = m.height >= 25 && m.width >= 40
		m.list.SetDelegate(m.delegate)
		m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.LightDark(m.dark)(lipgloss.Color("#1a5fc0"), lipgloss.Color("#7fc3ef")))
	case homeStatus:
		m.loading = false
		if msg.err != nil {
			m.status = "Core: erro na leitura: " + msg.err.Error()
			break
		}
		m.status = "Core: íntegro"
		m.checked, m.attention = 0, 0
		stateByPath := make(map[string]string, len(msg.statuses))
		for _, status := range msg.statuses {
			stateByPath[status.Path] = status.State
		}
		m.activeCore = 0
		for _, name := range catalog.CoreSkills() {
			if stateByPath[".claude/skills/"+name+"/SKILL.md"] == "ok" &&
				stateByPath[".agents/skills/"+name+"/SKILL.md"] == "ok" {
				m.activeCore++
			}
		}
		for _, status := range msg.statuses {
			switch status.State {
			case "ok", "pendente":
				m.checked++
				if status.State == "pendente" && m.attention == 0 {
					m.status = "Core: pendente · confirme hook em /hooks"
				}
			case "sem-historico":
				m.status = "Core: não instalado"
			default:
				m.attention++
				m.status = "Core: " + status.State + " · execute doctor"
			}
		}
		if m.status == "Core: íntegro" && m.activeCore != len(catalog.CoreSkills()) {
			m.status = "Core: atualização disponível · execute install"
		}
	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	case tea.KeyPressMsg:
		switch msg.String() {
		case "?":
			m.help = !m.help
			return m, nil
		case "esc":
			if m.help {
				m.help = false
				return m, nil
			}
			m.choice = "exit"
			return m, tea.Quit
		case "q", "ctrl+c":
			m.choice = "exit"
			return m, tea.Quit
		case "enter":
			if m.help {
				m.help = false
				return m, nil
			}
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
	width := max(1, min(m.width, 76))
	choose := lipgloss.LightDark(m.dark)
	blue := lipgloss.NewStyle().Foreground(choose(lipgloss.Color("#1a5fc0"), lipgloss.Color("#7fc3ef"))).Bold(true)
	muted := lipgloss.NewStyle().Foreground(choose(lipgloss.Color("#4b5669"), lipgloss.Color("#a1a1aa")))
	heading := blue.Render("GEX  /  DEV KIT") + "   " + muted.Render(version)
	label := m.status
	if m.loading {
		label = m.spinner.View() + " " + label
	}
	if m.height < 4 {
		title := heading
		if selected, ok := m.list.SelectedItem().(homeAction); ok {
			title = selected.title
		}
		view := tea.NewView(ansi.Truncate(title, width, "…"))
		view.AltScreen = true
		return view
	}
	if m.help {
		lines := []string{
			blue.Render("ATALHOS  /  GEX DEV KIT"),
			"",
			"↑ / ↓   Navegar pelas ações",
			"Enter   Abrir ação selecionada",
			"?       Mostrar ou fechar ajuda",
			"Esc / q Sair sem alterar arquivos",
			"",
			"Instalar: Core obrigatório; opções opt-in.",
			"Doctor: conferir integridade e agentes.",
			"",
			muted.Render("Esc ou Enter para voltar"),
		}
		view := tea.NewView(homeFrame(lines, width, m.height))
		view.AltScreen = true
		return view
	}
	if m.height < 25 || m.width < 40 {
		lines := []string{
			heading,
			label,
			fmt.Sprintf("Core %d/%d skills · %d verificados · %d atenção", m.activeCore, len(catalog.CoreSkills()), m.checked, m.attention),
			strings.TrimSuffix(m.list.View(), "\n"),
			muted.Render("↑↓ navegar · Enter abrir · ? ajuda · q sair"),
		}
		view := tea.NewView(homeFrame(lines, width, m.height))
		view.AltScreen = true
		return view
	}
	core := lipgloss.NewStyle().Width(max(1, width-4)).Padding(0, 1).
		Border(lipgloss.RoundedBorder()).BorderForeground(choose(lipgloss.Color("#1e6fd9"), lipgloss.Color("#38bdf8"))).
		Render(strings.Join([]string{
			blue.Render("CORE GEX  /  OBRIGATÓRIO"),
			fmt.Sprintf("%d skills para Claude Code + Codex", len(catalog.CoreSkills())),
			fmt.Sprintf("%d/%d skills íntegras", m.activeCore, len(catalog.CoreSkills())),
			label,
			muted.Render(fmt.Sprintf("%d arquivos verificados · %d atenção", m.checked, m.attention)),
		}, "\n"))
	menu := lipgloss.NewStyle().Width(max(1, width-4)).Padding(0, 1).
		Border(lipgloss.RoundedBorder()).BorderForeground(choose(lipgloss.Color("#d0d7e1"), lipgloss.Color("#26272d"))).
		Render(m.list.View())
	lines := []string{
		heading,
		muted.Render("INSTALAR  /  VERIFICAR  /  REVERTER"),
		"",
		core,
		muted.Render(fmt.Sprintf("OPCIONAIS  %d Matt · %d especializações · %d hook global", len(catalog.Skills()), len(catalog.SpecialtySkills()), len(catalog.Hooks()))),
		"",
		menu,
		muted.Render("↑↓ navegar  ·  Enter abrir  ·  ? atalhos  ·  q sair"),
	}
	view := tea.NewView(homeFrame(lines, width, m.height))
	view.AltScreen = true
	return view
}

func homeFrame(lines []string, width, height int) string {
	content := strings.Join(lines, "\n")
	rows := strings.Split(content, "\n")
	if len(rows) > height {
		rows = rows[:height]
	}
	for i, line := range rows {
		rows[i] = ansi.Truncate(line, width, "…")
	}
	return strings.Join(rows, "\n")
}

func showHome(in io.Reader, out io.Writer, home string) (string, error) {
	model, err := tea.NewProgram(newHomeModel(home), tea.WithInput(in), tea.WithOutput(out)).Run()
	if err != nil {
		return "", err
	}
	return model.(homeModel).choice, nil
}

func showAccessibleHome(in io.Reader, out io.Writer) (string, error) {
	scanner := bufio.NewScanner(in)
	for {
		fmt.Fprintf(out, "GEX / DEV KIT\nCore obrigatório: %d skills. Matt e especializações opcionais; hook global só por escolha.\nEscolha uma ação:\n", len(catalog.CoreSkills()))
		for i, action := range homeActions {
			fmt.Fprintf(out, "%d. %s — %s\n", i+1, action.title, action.description)
		}
		fmt.Fprint(out, "Ação (0/q/cancel para sair): ")
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return "", err
			}
			return "exit", nil
		}
		choice := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if choice == "0" || choice == "q" || choice == "cancel" {
			return "exit", nil
		}
		if index, err := strconv.Atoi(choice); err == nil && index > 0 && index <= len(homeActions) {
			return homeActions[index-1].name, nil
		}
		fmt.Fprintln(out, "Ação inválida.")
	}
}
