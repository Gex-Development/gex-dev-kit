package screens

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"

	"github.com/Gex-Development/gex-dev-kit/internal/app"
	"github.com/Gex-Development/gex-dev-kit/internal/catalog"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/components"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/layout"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/motion"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

type HomeChoice string

const (
	ChoiceInstall        HomeChoice = "install"
	ChoiceCatalog        HomeChoice = "catalog"
	ChoiceDoctor         HomeChoice = "doctor"
	ChoiceRollback       HomeChoice = "rollback"
	ChoiceRollbackAgents HomeChoice = "rollback-agents"
	ChoiceAddons         HomeChoice = "addons"
	ChoiceGitHooks       HomeChoice = "githooks"
	ChoiceShowcase       HomeChoice = "ds"
	ChoiceHelp           HomeChoice = "help"
	ChoiceExit           HomeChoice = "exit"
)

type HomeDeps struct {
	Core   func() (app.CoreStatus, error)
	Agents func() ([]app.AgentChange, error)
}

func RunHome(r Runner, c Common, d HomeDeps) (HomeChoice, error) {
	final, err := r.Run(newHomeModel(c, d))
	if err != nil {
		return ChoiceExit, err
	}
	m, ok := final.(homeModel)
	if !ok {
		return ChoiceExit, errors.New("resultado inesperado da tela inicial")
	}
	if m.interrupted {
		return ChoiceExit, SignalError{Signal: os.Interrupt}
	}
	if m.choice == "" {
		return ChoiceExit, nil
	}
	return m.choice, nil
}

const (
	homeFrameLoop  = -1
	homeBurstCount = 14
	homeBurstSeed  = 7
	homeBurstRow   = 1
	homeHeaderRows = 2
	homeTitle      = "GEX DEV KIT"
)

type homeCoreMsg struct {
	status app.CoreStatus
	err    error
}

type homeAgentsMsg struct {
	changes []app.AgentChange
	err     error
}

type homeSlot struct{ top, offset, rows int }

type homeModel struct {
	common      Common
	deps        HomeDeps
	theme       theme.Theme
	dark        bool
	profile     colorprofile.Profile
	animated    bool
	focused     bool
	sized       bool
	help        bool
	interrupted bool
	choice      HomeChoice
	width       int
	height      int
	keys        components.Keys
	menu        components.Menu
	checklist   components.Checklist
	reveal      motion.Spring
	burst       motion.Burst
	coreDone    bool
	core        app.CoreStatus
	coreErr     error
	agentsDone  bool
	agents      []app.AgentChange
	agentsErr   error
	errOpen     bool
	errTitle    string
	errCause    string
}

func newHomeModel(c Common, d HomeDeps) homeModel {
	m := homeModel{
		common:   c,
		deps:     d,
		dark:     c.Dark,
		profile:  c.Profile,
		animated: c.Motion,
		focused:  true,
		keys:     components.DefaultKeys(),
		menu:     components.NewMenu(homeMenuItems(), c.Motion),
		reveal:   motion.NewSpring(c.Motion, motion.Gentle).Target(1),
	}
	m.theme = m.buildTheme()
	m.checklist = components.NewChecklist(m.checklistItems(), c.Motion)
	return m
}

func homeMenuItems() []components.MenuItem {
	return []components.MenuItem{
		{ID: string(ChoiceInstall), Title: "Instalar", Description: "prévia e instalação guiada do Core"},
		{ID: string(ChoiceCatalog), Title: "Catálogo", Description: "skills, hooks e referências"},
		{ID: string(ChoiceDoctor), Title: "Doctor", Description: "conferir integridade e agentes"},
		{ID: string(ChoiceRollback), Title: "Reverter Core", Description: "restaurar arquivos do kit"},
		{ID: string(ChoiceRollbackAgents), Title: "Reverter agentes", Description: "restaurar versões anteriores"},
		{ID: string(ChoiceAddons), Title: "Addons", Description: "ferramentas opcionais, sem instalação automática"},
		{ID: string(ChoiceGitHooks), Title: "Git hooks", Description: "hooks por repositório, com prévia"},
		{ID: string(ChoiceShowcase), Title: "Vitrine DS", Description: "componentes e padrões visuais"},
		{ID: string(ChoiceHelp), Title: "Ajuda", Description: "atalhos de teclado"},
		{ID: string(ChoiceExit), Title: "Sair", Description: "fechar o Kit GEX"},
	}
}

func homeCoreCmd(read func() (app.CoreStatus, error)) tea.Cmd {
	return func() tea.Msg {
		if read == nil {
			return homeCoreMsg{}
		}
		status, err := read()
		return homeCoreMsg{status: status, err: err}
	}
}

func homeAgentsCmd(read func() ([]app.AgentChange, error)) tea.Cmd {
	return func() tea.Msg {
		if read == nil {
			return homeAgentsMsg{}
		}
		changes, err := read()
		return homeAgentsMsg{changes: changes, err: err}
	}
}

func (m homeModel) Init() tea.Cmd {
	return tea.Batch(
		homeCoreCmd(m.deps.Core),
		homeAgentsCmd(m.deps.Agents),
		tea.RequestBackgroundColor,
		tea.RequestWindowSize,
		m.checklist.Tick(),
	)
}

func (m homeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	next, cmd := m.react(msg)
	next.menu = next.menu.Place(0, next.menuOriginY())
	return next, cmd
}

func (m homeModel) react(msg tea.Msg) (homeModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.resize(msg.Width, msg.Height)
	case tea.BackgroundColorMsg:
		m.dark = msg.IsDark()
		m.theme = m.buildTheme()
		return m, nil
	case tea.ColorProfileMsg:
		m.profile = msg.Profile
		m.theme = m.buildTheme()
		return m, nil
	case tea.KeyPressMsg:
		return m.key(msg)
	case tea.MouseClickMsg, tea.MouseWheelMsg:
		if m.inMenu() {
			return m.forwardMenu(msg)
		}
		return m, nil
	case tea.FocusMsg:
		m.focused = true
		return m, tea.Batch(m.frame(), m.spin())
	case tea.BlurMsg:
		m.focused = false
		return m, nil
	case motion.FrameMsg:
		return m.frameTick(msg)
	case spinner.TickMsg:
		if !m.loading() || !m.focused {
			return m, nil
		}
		var cmd tea.Cmd
		m.checklist, cmd = m.checklist.Update(msg)
		return m, cmd
	case homeCoreMsg:
		return m.coreLoaded(msg)
	case homeAgentsMsg:
		return m.agentsLoaded(msg)
	case components.MenuChoiceMsg:
		return m.choose(HomeChoice(msg.Item.ID))
	}
	return m, nil
}

func (m homeModel) resize(width, height int) (homeModel, tea.Cmd) {
	m.width, m.height = width, height
	if m.sized {
		return m, nil
	}
	m.sized = true
	m.burst = motion.NewBurst(m.animated, float64(width)/2, homeBurstRow, homeBurstCount, homeBurstSeed)
	return m, m.frame()
}

func (m homeModel) frame() tea.Cmd {
	if !m.focused || !m.animating() {
		return nil
	}
	return motion.Frame(homeFrameLoop)
}

func (m homeModel) animating() bool {
	return !m.reveal.Settled() || m.burst.Alive()
}

func (m homeModel) frameTick(msg motion.FrameMsg) (homeModel, tea.Cmd) {
	if msg.ID != homeFrameLoop {
		return m.forwardMenu(msg)
	}
	if !m.focused {
		return m, nil
	}
	m.reveal = m.reveal.Step()
	m.burst = m.burst.Step()
	return m, m.frame()
}

func (m homeModel) spin() tea.Cmd {
	if !m.loading() || !m.focused {
		return nil
	}
	return m.checklist.Tick()
}

func (m homeModel) key(msg tea.KeyPressMsg) (homeModel, tea.Cmd) {
	if msg.String() == "ctrl+c" {
		m.interrupted = true
		m.choice = ChoiceExit
		return m, tea.Quit
	}
	if key.Matches(msg, m.keys.Quit) {
		m.choice = ChoiceExit
		return m, tea.Quit
	}
	if !m.inMenu() {
		if m.help && key.Matches(msg, m.keys.Help, m.keys.Back, m.keys.Enter) {
			m.help = false
		}
		return m, nil
	}
	switch {
	case key.Matches(msg, m.keys.Back):
		if m.errOpen {
			m.errOpen = false
			return m, nil
		}
		m.choice = ChoiceExit
		return m, tea.Quit
	case key.Matches(msg, m.keys.Help):
		m.help = true
		return m, nil
	case key.Matches(msg, m.keys.Suspend):
		return m, tea.Suspend
	}
	return m.forwardMenu(msg)
}

func (m homeModel) forwardMenu(msg tea.Msg) (homeModel, tea.Cmd) {
	var cmd tea.Cmd
	m.menu, cmd = m.menu.Update(msg)
	return m, cmd
}

func (m homeModel) choose(choice HomeChoice) (homeModel, tea.Cmd) {
	if choice == ChoiceHelp {
		m.help = true
		return m, nil
	}
	m.choice = choice
	return m, tea.Quit
}

func (m homeModel) coreLoaded(msg homeCoreMsg) (homeModel, tea.Cmd) {
	m.coreDone, m.core, m.coreErr = true, msg.status, msg.err
	if msg.err != nil {
		m = m.raise("Não foi possível ler o Core", msg.err)
	}
	m = m.refreshChecklist()
	return m, m.spin()
}

func (m homeModel) agentsLoaded(msg homeAgentsMsg) (homeModel, tea.Cmd) {
	m.agentsDone, m.agents, m.agentsErr = true, msg.changes, msg.err
	if msg.err != nil {
		m = m.raise("Não foi possível ler os agentes", msg.err)
	}
	m = m.refreshChecklist()
	return m, m.spin()
}

func (m homeModel) raise(title string, err error) homeModel {
	m.errOpen, m.errTitle, m.errCause = true, title, err.Error()
	return m
}

func (m homeModel) refreshChecklist() homeModel {
	m.checklist = components.NewChecklist(m.checklistItems(), m.animated)
	return m
}

func (m homeModel) buildTheme() theme.Theme {
	return theme.New(theme.Options{Dark: m.dark, Profile: m.profile, Plain: m.profile < colorprofile.ANSI})
}

func (m homeModel) size() layout.Size {
	return layout.Size{Width: m.width, Height: m.height}
}

func (m homeModel) tiny() bool {
	return layout.Classify(m.size()) == layout.ClassTiny
}

func (m homeModel) regular() bool {
	class := layout.Classify(m.size())
	return class == layout.ClassRegular || class == layout.ClassWide
}

func (m homeModel) inMenu() bool {
	return !m.tiny() && !m.help
}

func (m homeModel) loading() bool {
	return !m.coreDone || !m.agentsDone
}

func (m homeModel) menuOriginY() int {
	_, slot := m.bodyLines()
	return homeHeaderRows + slot.top - slot.offset
}

func (m homeModel) bodyLines() ([]string, homeSlot) {
	s := m.size()
	if m.help {
		return m.helpLines(s.Width), homeSlot{}
	}
	var lines []string
	if m.regular() {
		lines = append(lines, homeSplit(components.Wordmark(m.theme, s.Width, m.reveal.Value()))...)
		lines = append(lines, "")
		lines = append(lines, homeSplit(m.checklist.View(m.theme, s.Width))...)
	} else {
		lines = append(lines, homeSplit(components.Wordmark(m.theme, min(s.Width, 59), m.reveal.Value()))...)
	}
	if m.errOpen {
		lines = append(lines, homeSplit(components.ErrorBox(m.theme, s.Width, m.errTitle, m.errCause, ""))...)
	}
	bodyHeight := layout.BodySize(s).Height
	slot := homeSlot{top: len(lines), rows: max(bodyHeight-len(lines), 1)}
	slot.offset = max(0, m.menu.Index()-slot.rows+1)
	rows := homeSplit(m.menu.View(m.theme, s.Width, m.regular()))
	end := min(len(rows), slot.offset+slot.rows)
	lines = append(lines, rows[slot.offset:end]...)
	if m.burst.Alive() {
		lines = homeOverlay(m.theme, lines, m.burst.Particles(s.Width, bodyHeight))
	}
	return lines, slot
}

func (m homeModel) helpLines(width int) []string {
	return []string{
		m.theme.Title.Render("ATALHOS · GEX DEV KIT"),
		"",
		strings.Join(homeSplit(components.KeyHints(m.theme, width, true, m.keys)), "\n"),
		"",
		m.theme.Text.Render("Instalar: Core obrigatório; opções opt-in."),
		m.theme.Text.Render("Doctor: conferir integridade e agentes."),
		"",
		m.theme.Secondary.Render("esc ou ? fecha esta ajuda"),
	}
}

func (m homeModel) View() tea.View {
	s := m.size()
	body, _ := m.bodyLines()
	frame := layout.Frame{
		Title:   homeTitle,
		Context: homeLabel(m.common.Home, m.common.UserHome),
		Version: m.common.Version,
		Body:    strings.Join(body, "\n"),
		Footer:  components.KeyHints(m.theme, s.Width, false, m.keys),
	}
	view := tea.NewView(layout.Render(m.theme, s, frame))
	view.AltScreen = true
	view.MouseMode = tea.MouseModeCellMotion
	view.ReportFocus = true
	return view
}

func (m homeModel) checklistItems() []components.ChecklistItem {
	return []components.ChecklistItem{
		m.agentItem("claude", "Claude Code"),
		m.agentItem("codex", "Codex"),
		m.coreItem(),
		m.hooksItem(),
		{Label: "Versão · " + m.common.Version, State: theme.StateInfo},
	}
}

func (m homeModel) agentItem(name, title string) components.ChecklistItem {
	switch {
	case !m.agentsDone:
		return components.ChecklistItem{Label: title + " · lendo…", State: theme.StateRunning}
	case m.agentsErr != nil:
		return components.ChecklistItem{Label: title + " · " + m.agentsErr.Error(), State: theme.StateError}
	}
	change, ok := homeAgentChange(m.agents, name)
	if !ok {
		return components.ChecklistItem{Label: title + " · sem dados", State: theme.StateInfo}
	}
	return components.ChecklistItem{Label: title + " · " + homeAgentText(change), State: homeAgentState(change)}
}

func homeAgentChange(changes []app.AgentChange, name string) (app.AgentChange, bool) {
	for _, change := range changes {
		if change.Name == name {
			return change, true
		}
	}
	return app.AgentChange{}, false
}

func homeAgentText(change app.AgentChange) string {
	switch {
	case change.Blocked:
		return change.Reason
	case change.Current == "":
		return "não instalado pelo kit"
	case change.Current != change.Target:
		return change.Current + " → " + change.Target
	}
	return change.Current
}

func homeAgentState(change app.AgentChange) theme.State {
	if change.Blocked || change.Current == "" || change.Current != change.Target {
		return theme.StateWarn
	}
	return theme.StateOK
}

func (m homeModel) coreItem() components.ChecklistItem {
	total := len(catalog.CoreSkills())
	switch {
	case !m.coreDone:
		return components.ChecklistItem{Label: fmt.Sprintf("Core · ?/%d skills · lendo…", total), State: theme.StateRunning}
	case m.coreErr != nil:
		return components.ChecklistItem{Label: fmt.Sprintf("Core · ?/%d skills · %s", total, m.coreErr), State: theme.StateError}
	}
	label := fmt.Sprintf("Core · %d/%d skills · %s", m.core.ActiveCore, total, m.core.State)
	return components.ChecklistItem{Label: label, State: homeCoreState(m.core.State)}
}

func homeCoreState(state string) theme.State {
	switch state {
	case "íntegro":
		return theme.StateOK
	case "não instalado":
		return theme.StateInfo
	}
	return theme.StateWarn
}

func (m homeModel) hooksItem() components.ChecklistItem {
	switch {
	case !m.coreDone:
		return components.ChecklistItem{Label: "Hooks · lendo…", State: theme.StateRunning}
	case m.coreErr != nil:
		return components.ChecklistItem{Label: "Hooks · indisponível", State: theme.StateInfo}
	case len(m.core.Hooks) > 0:
		return components.ChecklistItem{Label: "Hooks · " + strings.Join(m.core.Hooks, ", "), State: theme.StateOK}
	}
	return components.ChecklistItem{Label: "Hooks · nenhum global (opcional)", State: theme.StateInfo}
}

func homeSplit(text string) []string {
	return strings.Split(text, "\n")
}

func homeOverlay(t theme.Theme, lines []string, parts []motion.Particle) []string {
	for _, p := range parts {
		if p.Y < len(lines) {
			lines[p.Y] = homeMark(t, lines[p.Y], p.X)
		}
	}
	return lines
}

func homeMark(t theme.Theme, line string, x int) string {
	glyph := t.Accent.Render(homeGlyph(t))
	width := ansi.StringWidth(line)
	if x >= width {
		return line + strings.Repeat(" ", x-width) + glyph
	}
	return ansi.Truncate(line, x, "") + glyph + ansi.TruncateLeft(line, x+1, "")
}

func homeGlyph(t theme.Theme) string {
	if t.Plain {
		return "*"
	}
	return "·"
}
