package screens

import (
	"fmt"
	"os"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/app"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/components"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/layout"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
)

const (
	doctorChromeRows  = 3
	doctorStateColumn = 12
	doctorPathColumn  = 60
	doctorNoItems     = "Nenhuma instalação registrada."
	doctorEmptyDetail = "Sem detalhes."
	doctorTitle       = "GEX / DOCTOR"
	doctorSummaryGap  = "  "
)

type doctorKeyMap struct{ keys components.Keys }

func (k doctorKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.keys.Up, k.keys.Down, k.keys.Left, k.keys.Right, k.keys.Back, k.keys.Quit, k.keys.Help}
}

func (k doctorKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.keys.Up, k.keys.Down},
		{k.keys.Left, k.keys.Right},
		{k.keys.Back, k.keys.Quit, k.keys.Help},
	}
}

type doctorModel struct {
	c           Common
	theme       theme.Theme
	profile     colorprofile.Profile
	size        layout.Size
	report      app.Report
	table       components.Table
	detail      viewport.Model
	help        bool
	interrupted bool
}

func newDoctorModel(c Common, report app.Report) doctorModel {
	m := doctorModel{
		c:       c,
		profile: c.Profile,
		size:    layout.Size{Width: 80, Height: 24},
		report:  report,
		detail:  viewport.New(viewport.WithWidth(80), viewport.WithHeight(1)),
	}
	m.theme = doctorTheme(c.Dark, c.Profile)
	m.table = components.NewTable(m.theme, doctorColumns(), doctorRows(report.Items, c.UserHome), 2)
	m.table = m.table.Resize(m.size.Width, m.tableRows(m.size.Height-doctorChromeRows))
	m.refreshDetail()
	return m
}

func doctorTheme(dark bool, profile colorprofile.Profile) theme.Theme {
	return theme.New(theme.Options{Dark: dark, Profile: profile, Plain: profile <= colorprofile.Ascii})
}

func doctorColumns() []components.Column {
	return []components.Column{{Title: "Estado", Width: doctorStateColumn}, {Title: "Arquivo", Width: doctorPathColumn}}
}

func doctorRows(items []app.Status, userHome string) [][]string {
	rows := make([][]string, 0, len(items))
	for _, item := range items {
		rows = append(rows, []string{item.State, homeLabel(item.Path, userHome)})
	}
	return rows
}

func (m doctorModel) Init() tea.Cmd { return tea.RequestBackgroundColor }

func (m doctorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		m.theme = doctorTheme(msg.IsDark(), m.profile)
	case tea.ColorProfileMsg:
		m.profile = msg.Profile
		m.theme = doctorTheme(m.theme.Dark, m.profile)
	case tea.WindowSizeMsg:
		m.size = layout.Size{Width: max(1, msg.Width), Height: max(1, msg.Height)}
		m.table = m.table.Resize(m.size.Width, m.tableRows(m.size.Height-doctorChromeRows))
		m.refreshDetail()
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m doctorModel) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	keys := components.DefaultKeys()
	if key.Matches(msg, keys.Quit) {
		m.interrupted = msg.String() == "ctrl+c"
		return m, tea.Quit
	}
	if m.size.Width > 0 && layout.Classify(m.size) == layout.ClassTiny {
		return m, nil
	}
	switch {
	case key.Matches(msg, keys.Back):
		return m, tea.Quit
	case key.Matches(msg, keys.Help):
		m.help = !m.help
		return m, nil
	case key.Matches(msg, keys.Left):
		m.detail.PageUp()
		return m, nil
	case key.Matches(msg, keys.Right):
		m.detail.PageDown()
		return m, nil
	}
	before := m.table.Cursor()
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	if m.table.Cursor() != before {
		m.refreshDetail()
	}
	return m, cmd
}

func (m *doctorModel) refreshDetail() {
	m.detail.SetWidth(max(1, m.size.Width))
	m.detail.SetContent(m.detailText())
	m.detail.SetHeight(m.detailHeight(m.size.Height - doctorChromeRows))
}

func (m doctorModel) detailText() string {
	item, ok := m.selected()
	if !ok {
		return doctorEmptyDetail
	}
	detail := item.Detail
	if detail == "" {
		detail = doctorEmptyDetail
	}
	return ansi.Hardwrap(detail, max(1, m.size.Width), true)
}

func (m doctorModel) selected() (app.Status, bool) {
	i := m.table.Cursor()
	if i < 0 || i >= len(m.report.Items) {
		return app.Status{}, false
	}
	return m.report.Items[i], true
}

func (m doctorModel) tableRows(bodyHeight int) int {
	want := len(m.report.Items) + 1
	return min(want, max(2, bodyHeight/3))
}

func (m doctorModel) warningRows(bodyHeight int) int {
	return min(len(m.report.Warnings), max(0, bodyHeight/4))
}

func (m doctorModel) detailHeight(bodyHeight int) int {
	used := 1 + m.warningRows(bodyHeight) + m.tableRows(bodyHeight) + 1
	room := max(1, bodyHeight-used)
	lines := strings.Count(m.detailText(), "\n") + 1
	return min(room, lines)
}

func (m doctorModel) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	v.MouseMode = tea.MouseModeNone
	return v
}

func (m doctorModel) render() string {
	if layout.Classify(m.size) == layout.ClassTiny {
		return doctorTooSmall(m.theme, m.size)
	}
	w := m.size.Width
	header := layout.Columns(m.theme, w, m.theme.Title.Render(doctorTitle)+" "+m.theme.Secondary.Render("· "+homeLabel(m.report.Root, m.c.UserHome)), m.theme.Secondary.Render(m.c.Version))
	divider := m.theme.Divider.Render(strings.Repeat("─", w))
	if m.theme.Plain {
		divider = strings.Repeat("-", w)
	}
	body := m.body(max(0, m.size.Height-doctorChromeRows))
	footer := layout.Fit(components.KeyHints(m.theme, w, m.help, doctorKeyMap{keys: components.DefaultKeys()}), w, m.size.Height)
	return layout.Fit(strings.Join([]string{header, divider, body, footer}, "\n"), w, m.size.Height)
}

func (m doctorModel) body(bodyHeight int) string {
	w := m.size.Width
	lines := []string{layout.Fit(m.summary(), w, 1)}
	for _, warning := range m.report.Warnings[:m.warningRows(bodyHeight)] {
		lines = append(lines, layout.Fit(components.StatusBadge(m.theme, theme.StateInfo, homeLabel(warning.Path, m.c.UserHome)+" · "+maskHome(warning.Detail, m.c.UserHome)), w, 1))
	}
	lines = append(lines, m.table.View())
	lines = append(lines, layout.Fit(m.detailHeader(), w, 1))
	m.detail.SetHeight(m.detailHeight(bodyHeight))
	lines = append(lines, m.detail.View())
	return layout.Fit(strings.Join(lines, "\n"), w, bodyHeight)
}

func (m doctorModel) summary() string {
	ok, attention, pending := m.report.Counts()
	attentionState := theme.StateOK
	if attention > 0 {
		attentionState = theme.StateError
	}
	return strings.Join([]string{
		components.StatusBadge(m.theme, theme.StateOK, fmt.Sprintf("%d íntegros", ok)),
		components.StatusBadge(m.theme, attentionState, fmt.Sprintf("%d atenção", attention)),
		components.StatusBadge(m.theme, theme.StatePending, fmt.Sprintf("%d pendentes", pending)),
	}, doctorSummaryGap)
}

func (m doctorModel) detailHeader() string {
	item, ok := m.selected()
	if !ok {
		return m.theme.Secondary.Render(doctorNoItems)
	}
	return components.StatusBadge(m.theme, doctorStateKind(item.State), item.State) + " " + m.theme.Strong.Render(homeLabel(item.Path, m.c.UserHome))
}

func doctorStateKind(state string) theme.State {
	switch state {
	case "ok":
		return theme.StateOK
	case "pendente", "sem-historico":
		return theme.StatePending
	case "atualizar":
		return theme.StateWarn
	case "aviso":
		return theme.StateInfo
	case "alterado", "ausente", "incompleto", "bloqueado":
		return theme.StateError
	}
	return theme.StateWarn
}

func doctorTooSmall(t theme.Theme, s layout.Size) string {
	var lines []string
	for _, line := range strings.Split(layout.TooSmall(t, s), "\n") {
		if strings.TrimSpace(ansi.Strip(line)) != "" {
			lines = append(lines, line)
		}
	}
	return layout.Fit(strings.Join(lines, "\n"), s.Width, s.Height)
}

func RunDoctor(r Runner, c Common, report app.Report) error {
	final, err := r.Run(newDoctorModel(c, report))
	if err != nil {
		return err
	}
	if m, ok := final.(doctorModel); ok && m.interrupted {
		return SignalError{Signal: os.Interrupt}
	}
	return nil
}
