package screens

import (
	"fmt"
	"io"
	"math"
	"os"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/Gex-Development/gex-dev-kit/internal/catalog"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/components"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/layout"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/motion"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

const catalogCoreDescription = "Skill do bloco Core, instalada com o kit."

var (
	catalogTabLabels  = []string{"Skills", "Hooks", "Referências"}
	catalogHookHeader = []string{"Hook", "Claude", "Codex", "Git", "Evento", "Escopo"}
	catalogPrevTab    = key.NewBinding(key.WithKeys("shift+tab"))
)

type catalogTab int

const (
	catalogSkillsTab catalogTab = iota
	catalogHooksTab
	catalogRefsTab
)

type catalogItem struct {
	name, badge, desc, source string
}

func (i catalogItem) FilterValue() string { return i.name + " " + i.badge }

type catalogDelegate struct{ t theme.Theme }

func (d catalogDelegate) Height() int                         { return 2 }
func (d catalogDelegate) Spacing() int                        { return 0 }
func (d catalogDelegate) Update(tea.Msg, *list.Model) tea.Cmd { return nil }

func (d catalogDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	row, ok := item.(catalogItem)
	if !ok {
		return
	}
	styles := d.t.ListItem()
	title, desc := styles.NormalTitle, styles.NormalDesc
	if index == m.Index() {
		title, desc = styles.SelectedTitle, styles.SelectedDesc
	}
	head := title.Render(layout.Fit(row.name+"  "+row.badge, m.Width()-title.GetHorizontalFrameSize(), 1))
	body := desc.Render(layout.Fit(row.desc, m.Width()-desc.GetHorizontalFrameSize(), 1))
	fmt.Fprint(w, head+"\n"+body)
}

type catalogModel struct {
	c           Common
	t           theme.Theme
	keys        components.Keys
	size        layout.Size
	tab         catalogTab
	highlight   motion.Spring
	frame       int
	help        bool
	hooks       []catalog.HookSpec
	hookCursor  int
	skills      list.Model
	refs        list.Model
	interrupted bool
}

func RunCatalog(r Runner, c Common) error {
	m, err := r.Run(newCatalogModel(c))
	return catalogExitError(m, err)
}

func catalogExitError(m tea.Model, err error) error {
	if err != nil {
		return err
	}
	if model, ok := m.(catalogModel); ok && model.interrupted {
		return SignalError{Signal: os.Interrupt}
	}
	return nil
}

func newCatalogModel(c Common) catalogModel {
	t := catalogTheme(c)
	hooks := catalog.HookSpecs()
	m := catalogModel{
		c:          c,
		t:          t,
		keys:       components.DefaultKeys(),
		highlight:  motion.NewSpring(c.Motion, motion.Snappy).Snap(0),
		hooks:      hooks,
		hookCursor: catalogFirstLimited(hooks),
		skills:     catalogList(catalogSkillItems(), t),
		refs:       catalogList(catalogRefItems(), t),
	}
	return m
}

func catalogTheme(c Common) theme.Theme {
	return theme.New(theme.Options{Dark: c.Dark, Profile: c.Profile})
}

func catalogList(items []list.Item, t theme.Theme) list.Model {
	l := list.New(items, catalogDelegate{t: t}, 0, 0)
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.SetStatusBarItemName("item", "itens")
	l.Filter = catalogFilter
	l.FilterInput.Prompt = "Filtro: "
	l.DisableQuitKeybindings()
	return l
}

func catalogFilter(term string, targets []string) []list.Rank {
	needle := strings.ToLower(term)
	if exact := catalogRanks(targets, needle, true); len(exact) > 0 {
		return exact
	}
	return catalogRanks(targets, needle, false)
}

func catalogRanks(targets []string, needle string, wholeWord bool) []list.Rank {
	var ranks []list.Rank
	for i, target := range targets {
		if catalogMatchesWords(strings.ToLower(target), needle, wholeWord) {
			ranks = append(ranks, list.Rank{Index: i})
		}
	}
	return ranks
}

func catalogMatchesWords(target, needle string, wholeWord bool) bool {
	for i := range len(target) {
		if i > 0 && !strings.ContainsRune(" -/", rune(target[i-1])) {
			continue
		}
		end := i + len(needle)
		if !strings.HasPrefix(target[i:], needle) {
			continue
		}
		if !wholeWord || end == len(target) || strings.ContainsRune(" -/", rune(target[end])) {
			return true
		}
	}
	return false
}

func catalogSkillItems() []list.Item {
	var items []list.Item
	for _, name := range catalog.CoreSkills() {
		items = append(items, catalogItem{name: name, badge: "Core", desc: catalogCoreDescription})
	}
	for _, name := range catalog.Skills() {
		items = append(items, catalogItem{name: name, badge: "Matt", desc: catalog.SkillDescription(name), source: catalog.MattSource(name)})
	}
	for _, specialty := range catalog.Specialties() {
		items = append(items, catalogItem{name: specialty.Name, badge: "Especialização · " + specialty.Group, desc: specialty.Description})
	}
	return items
}

func catalogRefItems() []list.Item {
	var items []list.Item
	for _, ref := range append(catalog.ReferencedSkills(), catalog.HookCandidates()...) {
		items = append(items, catalogItem{name: ref.Name, badge: ref.Status, desc: ref.Reason, source: ref.Source})
	}
	return items
}

func catalogFirstLimited(hooks []catalog.HookSpec) int {
	for i, spec := range hooks {
		if spec.Scope == catalog.HookScopeAgent && !(spec.Claude.Supported && spec.Codex.Supported) {
			return i
		}
	}
	return 0
}

func (m catalogModel) Init() tea.Cmd { return nil }

func (m catalogModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.size = layout.Size{Width: msg.Width, Height: msg.Height}
		return m.resize(), nil
	case tea.BackgroundColorMsg:
		m.c.Dark = msg.IsDark()
		return m.withTheme(catalogTheme(m.c)), nil
	case tea.ColorProfileMsg:
		m.c.Profile = msg.Profile
		return m.withTheme(catalogTheme(m.c)), nil
	case motion.FrameMsg:
		return m.step(msg)
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}
	return m.forward(msg)
}

func (m catalogModel) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

func (m catalogModel) resize() catalogModel {
	body := layout.BodySize(m.size)
	listHeight := max(body.Height-3, 1)
	m.skills.SetSize(body.Width, listHeight)
	m.refs.SetSize(body.Width, listHeight)
	return m
}

func (m catalogModel) withTheme(t theme.Theme) catalogModel {
	m.t = t
	catalogApplyTheme(&m.skills, t)
	catalogApplyTheme(&m.refs, t)
	return m
}

func catalogApplyTheme(l *list.Model, t theme.Theme) {
	l.Styles = list.DefaultStyles(t.Dark)
	l.SetDelegate(catalogDelegate{t: t})
}

func (m catalogModel) step(msg motion.FrameMsg) (tea.Model, tea.Cmd) {
	if msg.ID != m.frame {
		return m, nil
	}
	m.highlight = m.highlight.Step()
	return m, m.frameCmd()
}

func (m catalogModel) frameCmd() tea.Cmd {
	if m.highlight.Settled() {
		return nil
	}
	return motion.Frame(m.frame)
}

func (m catalogModel) selectTab(next catalogTab) (tea.Model, tea.Cmd) {
	n := len(catalogTabLabels)
	m.tab = catalogTab((int(next)%n + n) % n)
	m.frame++
	m.highlight = m.highlight.Target(float64(m.tab))
	return m, m.frameCmd()
}

func (m catalogModel) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "ctrl+c" {
		m.interrupted = true
		return m, tea.Quit
	}
	if layout.Classify(m.size) == layout.ClassTiny {
		if key.Matches(msg, m.keys.Quit) {
			return m, tea.Quit
		}
		return m, nil
	}
	keys := m.keys
	if m.textMode() {
		keys = m.keys.TextInput()
	}
	switch {
	case key.Matches(msg, keys.Help):
		m.help = !m.help
		return m, nil
	case key.Matches(msg, keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, keys.Back):
		if m.filtered() {
			return m.forward(msg)
		}
		return m, tea.Quit
	case key.Matches(msg, keys.Tab):
		return m.selectTab(m.tab + 1)
	case key.Matches(msg, catalogPrevTab):
		return m.selectTab(m.tab - 1)
	case m.tab == catalogHooksTab && key.Matches(msg, keys.Up):
		m.hookCursor = max(m.hookCursor-1, 0)
		return m, nil
	case m.tab == catalogHooksTab && key.Matches(msg, keys.Down):
		m.hookCursor = min(m.hookCursor+1, len(m.hooks)-1)
		return m, nil
	}
	return m.forward(msg)
}

func (m catalogModel) forward(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.tab {
	case catalogSkillsTab:
		m.skills, cmd = m.skills.Update(msg)
	case catalogRefsTab:
		m.refs, cmd = m.refs.Update(msg)
	}
	return m, cmd
}

func (m catalogModel) activeList() list.Model {
	if m.tab == catalogRefsTab {
		return m.refs
	}
	return m.skills
}

func (m catalogModel) textMode() bool {
	return m.tab != catalogHooksTab && m.activeList().SettingFilter()
}

func (m catalogModel) filtered() bool {
	return m.tab != catalogHooksTab && m.activeList().FilterState() != list.Unfiltered
}

func (m catalogModel) render() string {
	return layout.Render(m.t, m.size, layout.Frame{
		Title:   "Catálogo",
		Context: m.context(),
		Version: m.c.Version,
		Body:    m.body(),
		Footer:  m.footer(),
	})
}

func (m catalogModel) context() string {
	switch m.tab {
	case catalogHooksTab:
		return fmt.Sprintf("Hooks: %d de agente, %d do Git", len(catalog.Hooks()), len(catalog.GitHooks()))
	case catalogRefsTab:
		return fmt.Sprintf("Referências: %d skills, %d candidatos a hook", len(catalog.ReferencedSkills()), len(catalog.HookCandidates()))
	}
	return fmt.Sprintf("Skills: %d Core · %d Matt · %d especializações", len(catalog.CoreSkills()), len(catalog.Skills()), len(catalog.Specialties()))
}

func (m catalogModel) footer() string {
	switch {
	case m.textMode():
		return "enter aplica · esc cancela o filtro"
	case m.help:
		return "? fecha a ajuda · esc sai"
	case m.tab == catalogHooksTab:
		return "tab aba · ↑/↓ escolher hook · ? ajuda · esc sai"
	}
	return "tab aba · ↑/↓ mover · / filtrar · ? ajuda · esc sai"
}

func (m catalogModel) body() string {
	if m.help {
		return components.KeyHints(m.t, m.size.Width, true, m.keys)
	}
	return m.tabBar() + "\n" + m.tabContent()
}

func (m catalogModel) tabContent() string {
	if m.tab == catalogHooksTab {
		return m.hooksContent()
	}
	return m.listContent(m.activeList())
}

func (m catalogModel) listContent(l list.Model) string {
	detail := ""
	if item, ok := l.SelectedItem().(catalogItem); ok && item.source != "" {
		detail = "Fonte: " + item.source
	}
	return l.View() + "\n" + layout.Fit(m.t.Secondary.Render(detail), m.size.Width, 1)
}

func (m catalogModel) tabBar() string {
	var labels []string
	for i, label := range catalogTabLabels {
		style := m.t.Secondary
		if catalogTab(i) == m.tab {
			style = m.t.Selected
		}
		labels = append(labels, style.Render(" "+label+" "))
	}
	return strings.Join(labels, " ") + "\n" + m.tabUnderline()
}

func (m catalogModel) tabUnderline() string {
	starts, widths := catalogTabSpans()
	start, width := catalogUnderline(m.highlight.Value(), starts, widths)
	return strings.Repeat(" ", max(start, 0)) + m.t.Accent.Render(strings.Repeat(catalogRule(m.t.Plain), max(width, 0)))
}

func catalogTabSpans() ([]int, []int) {
	var starts, widths []int
	pos := 0
	for _, label := range catalogTabLabels {
		width := lipgloss.Width(label) + 2
		starts = append(starts, pos)
		widths = append(widths, width)
		pos += width + 1
	}
	return starts, widths
}

func catalogUnderline(pos float64, starts, widths []int) (int, int) {
	last := len(starts) - 1
	i := min(max(int(math.Floor(pos)), 0), last)
	if i == last {
		return starts[i], widths[i]
	}
	frac := min(max(pos-float64(i), 0), 1)
	start := float64(starts[i]) + float64(starts[i+1]-starts[i])*frac
	width := float64(widths[i]) + float64(widths[i+1]-widths[i])*frac
	return int(math.Round(start)), int(math.Round(width))
}

func catalogRule(plain bool) string {
	if plain {
		return "-"
	}
	return "━"
}

func (m catalogModel) hooksContent() string {
	table := components.StaticTable(m.t, m.size.Width, catalogHookHeader, m.hookRows())
	return table + "\n" + m.hookDetail()
}

func (m catalogModel) hookRows() [][]string {
	rows := make([][]string, 0, len(m.hooks))
	for i, spec := range m.hooks {
		name := "  " + spec.Name
		if i == m.hookCursor {
			name = "› " + spec.Name
		}
		rows = append(rows, []string{
			name,
			catalogYesNo(spec.Claude.Supported),
			catalogYesNo(spec.Codex.Supported),
			catalogYesNo(spec.GitHook != ""),
			catalogEvent(spec),
			catalogPolicy(spec.Policy),
		})
	}
	return rows
}

func (m catalogModel) hookDetail() string {
	if len(m.hooks) == 0 {
		return ""
	}
	spec := m.hooks[m.hookCursor]
	lines := []string{
		spec.Name + " · escopo " + catalogPolicy(spec.Policy),
		spec.Description,
		"Claude: " + catalogSupportText(spec.Claude),
		"Codex: " + catalogSupportText(spec.Codex),
	}
	for i, line := range lines {
		lines[i] = layout.Fit(line, m.size.Width, 1)
	}
	return strings.Join(lines, "\n")
}

func catalogYesNo(ok bool) string {
	if ok {
		return "sim"
	}
	return "não"
}

func catalogEvent(spec catalog.HookSpec) string {
	switch {
	case spec.GitHook != "":
		return spec.GitHook
	case spec.Claude.Supported:
		return spec.Claude.Event
	case spec.Codex.Supported:
		return spec.Codex.Event
	}
	return "—"
}

func catalogPolicy(policy string) string {
	if policy == "gex" {
		return "só GEX"
	}
	return "todos"
}

func catalogSupportText(support catalog.HookSupport) string {
	if support.Supported {
		return "suportado · " + support.Event + " · decisão " + support.Decision
	}
	return "não suportado. " + support.Reason
}
