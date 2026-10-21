package screens

import (
	"fmt"
	"os"
	"time"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/components"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/layout"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/motion"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

const (
	showcaseFrameID     = 1
	showcaseListMax     = 22
	showcaseTitle       = "gex-dev-kit ds"
	showcaseDelayFirst  = 300 * time.Millisecond
	showcaseDelaySecond = 100 * time.Millisecond
	showcaseDelayThird  = 200 * time.Millisecond
)

type showcaseItem string

func (i showcaseItem) Title() string       { return string(i) }
func (i showcaseItem) Description() string { return "" }
func (i showcaseItem) FilterValue() string { return string(i) }

type showcaseTickMsg struct{}

type showcaseOrderMsg struct {
	Batch bool
	ID    int
}

type showcaseModel struct {
	version     string
	dark        bool
	profile     colorprofile.Profile
	motion      bool
	t           theme.Theme
	keys        components.Keys
	pages       []showcasePage
	list        list.Model
	vp          viewport.Model
	live        showcaseLive
	size        layout.Size
	demoFocus   bool
	help        bool
	interrupted bool
}

func newShowcaseModel(c Common) showcaseModel {
	pages := showcasePages()
	items := make([]list.Item, 0, len(pages))
	for _, p := range pages {
		items = append(items, showcaseItem(p.Name))
	}
	m := showcaseModel{
		version: c.Version,
		dark:    c.Dark,
		profile: c.Profile,
		motion:  c.Motion,
		keys:    components.DefaultKeys(),
		pages:   pages,
		vp:      viewport.New(),
	}
	m.t = theme.New(theme.Options{Dark: m.dark, Profile: m.profile})
	m.list = list.New(items, showcaseDelegate(m.t), 0, 0)
	m.list.SetFilteringEnabled(false)
	m.list.SetShowTitle(false)
	m.list.SetShowFilter(false)
	m.list.SetShowStatusBar(false)
	m.list.SetShowHelp(false)
	m.list.SetShowPagination(false)
	m.live = showcaseLive{Dark: m.dark, Motion: m.motion, Profile: m.profile, Last: 0}
	return m
}

func showcaseDelegate(t theme.Theme) list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.ShowDescription = false
	d.SetSpacing(0)
	d.Styles = t.ListItem()
	return d
}

func showcaseColumns(bs layout.Size) (int, int) {
	listW := min(showcaseListMax, bs.Width/3)
	return listW, max(bs.Width-listW-1, 0)
}

func showcaseTick() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg { return showcaseTickMsg{} })
}

func showcaseOrderTick(batch bool, id int, d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg { return showcaseOrderMsg{Batch: batch, ID: id} })
}

func (m showcaseModel) Init() tea.Cmd {
	cmds := []tea.Cmd{showcaseTick()}
	if m.motion {
		cmds = append(cmds, motion.Frame(showcaseFrameID))
	}
	return tea.Batch(cmds...)
}

func (m showcaseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	next, cmd := m.handle(msg)
	next.refresh()
	return next, cmd
}

func (m showcaseModel) handle(msg tea.Msg) (showcaseModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.size = layout.Size{Width: msg.Width, Height: msg.Height}
		m.live.Width, m.live.Height = msg.Width, msg.Height
		m.resize()
	case tea.KeyPressMsg:
		return m.key(msg)
	case tea.MouseClickMsg:
		m.live.Clicks++
		m.live.LastMouse = msg.Mouse().String()
	case tea.MouseWheelMsg:
		m.vp, _ = m.vp.Update(msg)
	case tea.FocusMsg:
		m.live.Focused = true
	case tea.BlurMsg:
		m.live.Focused = false
	case tea.ColorProfileMsg:
		m.profile = msg.Profile
		m.live.Profile = msg.Profile
		m.t = theme.New(theme.Options{Dark: m.dark, Profile: m.profile})
	case showcaseTickMsg:
		m.live.Ticks++
		return m, showcaseTick()
	case motion.FrameMsg:
		if !m.motion {
			return m, nil
		}
		m.live.Frames++
		return m, motion.Frame(msg.ID)
	case showcaseOrderMsg:
		if msg.Batch {
			m.live.Batch = append(m.live.Batch, msg.ID)
		} else {
			m.live.Sequence = append(m.live.Sequence, msg.ID)
		}
	}
	return m, nil
}

func (m showcaseModel) key(msg tea.KeyPressMsg) (showcaseModel, tea.Cmd) {
	switch {
	case msg.String() == "ctrl+c":
		m.interrupted = true
		return m, tea.Quit
	case showcaseKey(msg, "q"):
		return m, tea.Quit
	case showcaseKey(msg, "esc"):
		if m.help {
			m.help = false
			return m, nil
		}
		return m, tea.Quit
	case showcaseKey(msg, "?"):
		m.help = !m.help
		return m, nil
	case showcaseKey(msg, "tab"):
		m.demoFocus = !m.demoFocus
		return m, nil
	case showcaseKey(msg, "t"):
		m.dark = !m.dark
		m.applyTheme()
		return m, nil
	case showcaseKey(msg, "m"):
		m.motion = !m.motion
		m.live.Motion = m.motion
		if m.motion {
			return m, motion.Frame(showcaseFrameID)
		}
		return m, nil
	case m.keys.Suspend.Enabled() && showcaseKey(msg, "ctrl+z"):
		return m, tea.Suspend
	}
	if m.demoFocus {
		return m.demoKey(msg)
	}
	before := m.list.Index()
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	if m.list.Index() != before {
		m.vp.GotoTop()
	}
	return m, cmd
}

func showcaseKey(msg tea.KeyPressMsg, name string) bool {
	return msg.String() == name
}

func (m showcaseModel) demoKey(msg tea.KeyPressMsg) (showcaseModel, tea.Cmd) {
	var cmd tea.Cmd
	if m.currentPage().Name == showcasePagePrograma {
		cmd = m.programKey(msg)
	}
	if cmd == nil {
		m.vp, cmd = m.vp.Update(msg)
	}
	return m, cmd
}

func (m *showcaseModel) programKey(msg tea.KeyPressMsg) tea.Cmd {
	switch msg.String() {
	case "a":
		m.live.AltScreen = !m.live.AltScreen
	case "p":
		if m.live.AltScreen {
			return nil
		}
		m.live.Printed++
		return tea.Println(fmt.Sprintf("vitrine: linha %d impressa com tea.Println", m.live.Printed))
	case "1":
		m.live.Batch = nil
		return tea.Batch(
			showcaseOrderTick(true, 1, showcaseDelayFirst),
			showcaseOrderTick(true, 2, showcaseDelaySecond),
			showcaseOrderTick(true, 3, showcaseDelayThird),
		)
	case "2":
		m.live.Sequence = nil
		return tea.Sequence(
			showcaseOrderTick(false, 1, showcaseDelayFirst),
			showcaseOrderTick(false, 2, showcaseDelaySecond),
			showcaseOrderTick(false, 3, showcaseDelayThird),
		)
	case "e":
		m.live.Fault = showcaseFault{
			Cause:  "exemplo de erro tipado da vitrine",
			Action: "volte à lista e escolha outra página",
		}
	}
	return nil
}

func (m *showcaseModel) applyTheme() {
	m.t = theme.New(theme.Options{Dark: m.dark, Profile: m.profile})
	m.live.Dark = m.dark
	m.list.SetDelegate(showcaseDelegate(m.t))
}

func (m *showcaseModel) resize() {
	bs := layout.BodySize(m.size)
	listW, demoW := showcaseColumns(bs)
	contentH := max(bs.Height-1, 1)
	m.list.SetSize(listW, contentH)
	m.vp.SetWidth(demoW)
	m.vp.SetHeight(contentH)
}

func (m *showcaseModel) currentPage() showcasePage {
	return m.pages[m.list.Index()]
}

func (m *showcaseModel) refresh() {
	idx := m.list.Index()
	page := m.pages[idx]
	if page.Name != showcasePageAcessivel {
		m.live.Last = idx
	}
	_, demoW := showcaseColumns(layout.BodySize(m.size))
	body := showcaseBody(page, m.live, m.t, demoW)
	if m.help {
		body = components.KeyHints(m.t, demoW, true, m.keys)
	}
	m.vp.SetContent(body)
}

func (m showcaseModel) label(text string, focused bool) string {
	if focused {
		return m.t.Title.Render("▸ " + text)
	}
	return m.t.Faint.Render("  " + text)
}

func (m showcaseModel) compose() string {
	bs := layout.BodySize(m.size)
	listW, demoW := showcaseColumns(bs)
	contentH := max(bs.Height-1, 1)
	labels := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.PlaceHorizontal(listW, lipgloss.Left, layout.Fit(m.label("Páginas", !m.demoFocus), listW, 1)),
		" ",
		lipgloss.PlaceHorizontal(demoW, lipgloss.Left, layout.Fit(m.label("Demo", m.demoFocus), demoW, 1)),
	)
	left := lipgloss.Place(listW, contentH, lipgloss.Left, lipgloss.Top, layout.Fit(m.list.View(), listW, contentH))
	right := lipgloss.Place(demoW, contentH, lipgloss.Left, lipgloss.Top, m.vp.View())
	return lipgloss.JoinVertical(lipgloss.Left, labels, lipgloss.JoinHorizontal(lipgloss.Top, left, " ", right))
}

func (m showcaseModel) View() tea.View {
	frame := layout.Frame{
		Title:   showcaseTitle,
		Context: m.currentPage().Name,
		Version: m.version,
		Body:    m.compose(),
		Footer:  components.KeyHints(m.t, m.size.Width, false, m.keys),
	}
	v := tea.NewView(layout.Render(m.t, m.size, frame))
	v.AltScreen = m.live.AltScreen
	v.MouseMode = tea.MouseModeCellMotion
	v.ReportFocus = true
	return v
}

func RunShowcase(r Runner, c Common) error {
	final, err := r.Run(newShowcaseModel(c))
	if err != nil {
		return err
	}
	if m, ok := final.(showcaseModel); ok && m.interrupted {
		return SignalError{Signal: os.Interrupt}
	}
	return nil
}
