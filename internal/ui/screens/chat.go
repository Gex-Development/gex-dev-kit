package screens

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"

	"github.com/Gex-Development/gex-dev-kit/internal/app"
	"github.com/Gex-Development/gex-dev-kit/internal/catalog"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/components"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/motion"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

type ChatIntent string

const (
	IntentHome           ChatIntent = ""
	IntentSetup          ChatIntent = "setup"
	IntentCatalog        ChatIntent = "catalog"
	IntentDoctor         ChatIntent = "doctor"
	IntentRollback       ChatIntent = "rollback"
	IntentRollbackAgents ChatIntent = "rollback-agents"
	IntentGitHooks       ChatIntent = "githooks"
	IntentAddons         ChatIntent = "addons"
)

type (
	ChatResult  struct{ ExitCode int }
	InstallDeps struct {
		Options func() []app.OptionGroup
		Plan    func(app.Request) (app.Plan, error)
		Apply   func(app.Plan, func(app.Step)) (app.Result, error)
	}
)

type ChatDeps struct {
	Core           func() (app.CoreStatus, error)
	Agents         func() ([]app.AgentChange, error)
	Git            func() (app.GitInfo, bool)
	Dir            string
	Doctor         func() (app.Report, error)
	Install        InstallDeps
	Addons         AddonDeps
	Rollback       RollbackDeps
	RollbackAgents RollbackDeps
	GitHooks       GitHooksDeps
	GitHooksPreset app.GitHooksRequest
	Preset         app.Request
}
type chatEntry struct {
	role   components.ChatRole
	text   string
	runes  []rune
	at     []time.Duration
	shown  int
	tool   *components.ToolBlock
	isTool bool
	banner string
	info   []string
}
type chatAction struct {
	say    string
	tool   *chatToolSpec
	ask    *chatAsk
	banner string
	info   []string
	quit   bool
}
type chatRun struct {
	entry      int
	started    time.Time
	write      bool
	wasWrite   bool
	phases     []string
	completion *chatFinishedMsg
	displayed  int
}
type chatAnswer struct {
	IDs  []string
	Text string
}
type chatAsk struct {
	choice                           components.Choice
	picker                           components.PathPicker
	options                          []components.ChoiceOption
	secret                           bool
	hasChoice, input, textarea, path bool
	question                         string
	context                          string
	validate                         func(string) error
	onAnswer                         func(chatModel, chatAnswer) chatModel
	mark                             int
	step                             int
}
type chatToolSpec struct {
	title, params string
	phases        []string
	write         bool
	run           func(func(string)) chatToolResult
	fill          func(chatModel) chatToolResult
	then          func(chatModel, chatToolResult) chatModel
}
type chatToolResult struct {
	state theme.State
	rows  []components.ToolRow
	lines []string
	err   error
	value any
}
type chatProgressMsg struct {
	entry int
	line  string
}
type chatFinishedMsg struct {
	entry  int
	result chatToolResult
	fill   func(chatModel) chatToolResult
	then   func(chatModel, chatToolResult) chatModel
}
type chatCoreMsg struct {
	status app.CoreStatus
	err    error
}
type chatAgentsMsg struct {
	changes []app.AgentChange
	err     error
}
type chatGitMsg struct {
	info app.GitInfo
	ok   bool
}
type (
	chatBeginMsg struct{}
	chatPhase    uint8
)

const (
	chatWelcome chatPhase = iota
	chatTalking
)

const (
	chatFrame               = 913
	chatStreamFrame         = 914
	chatStreamCPS           = 55
	chatStreamBudgetSeconds = 9
	chatBeatDuration        = 450 * time.Millisecond
	chatPhaseDuration       = time.Second
	chatMaxRetention        = 6 * time.Second
)

type chatModel struct {
	common          Common
	deps            ChatDeps
	start           ChatIntent
	theme           theme.Theme
	width, height   int
	entries         []chatEntry
	queue           []chatAction
	running         *chatRun
	channel         chan tea.Msg
	turnStart       time.Time
	ask             *chatAsk
	history         []chatAsk
	palette         components.Palette
	paletteOpen     bool
	steps           []string
	step            int
	context         string
	scroll          int
	exitCode        int
	interrupted     bool
	signal          os.Signal
	phase           chatPhase
	input           textinput.Model
	area            textarea.Model
	spin            spinner.Model
	reveal          motion.Spring
	streamStart     time.Time
	streamLast      time.Time
	beat            time.Duration
	beatStart       time.Time
	pauseAt         time.Time
	focused         bool
	core            app.CoreStatus
	coreErr         error
	coreDone        bool
	agents          []app.AgentChange
	agentsErr       error
	agentsDone      bool
	git             app.GitInfo
	gitOK           bool
	menus           int
	groups          []app.OptionGroup
	request         app.Request
	harnessIDs      []string
	addonNames      []string
	addonPlan       app.AddonPlan
	addonSecrets    map[string]string
	selectedGroups  []string
	plan            app.Plan
	gitHooksRequest app.GitHooksRequest
}

func RunChat(r Runner, c Common, d ChatDeps, start ChatIntent) (ChatResult, error) {
	guarded, wait := installGuard(d.Install)
	d.Install = guarded
	var group sync.WaitGroup
	d.Rollback = rollbackGuard(d.Rollback, &group)
	d.RollbackAgents = rollbackGuard(d.RollbackAgents, &group)
	d.GitHooks = gitHooksGuard(d.GitHooks, &group)
	d.Addons = addonsGuard(d.Addons, &group)
	final, err := r.Run(newChatModel(c, d, start))
	wait()
	group.Wait()
	if err != nil {
		return ChatResult{}, err
	}
	m, ok := final.(chatModel)
	if !ok {
		return ChatResult{}, errors.New("resultado inesperado do chat")
	}
	if m.interrupted {
		return ChatResult{}, SignalError{Signal: os.Interrupt}
	}
	if m.signal != nil {
		return ChatResult{}, SignalError{Signal: m.signal}
	}
	return ChatResult{ExitCode: m.exitCode}, nil
}

func newChatModel(c Common, d ChatDeps, start ChatIntent) chatModel {
	input := textinput.New()
	input.Prompt = ""
	input.Placeholder = "Digite / para ver os comandos"
	input.Focus()
	area := textarea.New()
	area.SetHeight(3)
	area.ShowLineNumbers = false
	input.SetWidth(74)
	area.SetWidth(74)
	spin := spinner.New()
	spin.Spinner = spinner.MiniDot
	m := chatModel{common: c, deps: d, start: start, theme: theme.New(theme.Options{Dark: c.Dark, Profile: c.Profile, Plain: c.Profile == colorprofile.ASCII, NoPowerline: !c.Powerline}), width: 80, height: 24, input: input, area: area, spin: spin, reveal: motion.NewSpring(c.Motion && !c.NoIntro, motion.Intro).Target(1), focused: true, context: "pronto", request: d.Preset, gitHooksRequest: d.GitHooksPreset, palette: components.NewPalette(chatCommands)}
	if m.request.Home == "" {
		m.request.Home = c.Home
	}
	if d.Install.Options != nil {
		m.groups = d.Install.Options()
	}
	if start == IntentHome {
		m.phase = chatWelcome
	} else {
		m.phase = chatTalking
		m = m.flowGreeting(true)
	}
	return m
}
func (m chatModel) Busy() bool { return m.running != nil && m.running.write }
func (m chatModel) Init() tea.Cmd {
	cmds := []tea.Cmd{tea.RequestWindowSize, tea.RequestBackgroundColor, chatCoreCmd(m.deps.Core), chatAgentsCmd(m.deps.Agents), chatGitCmd(m.deps.Git)}
	if m.phase == chatTalking {
		cmds = append(cmds, func() tea.Msg { return chatBeginMsg{} })
	}
	if m.common.Motion && !m.reveal.Settled() {
		cmds = append(cmds, motion.Frame(chatFrame))
	}
	if m.locked() && m.focused {
		cmds = append(cmds, m.spin.Tick)
	}
	return tea.Batch(cmds...)
}

func chatCoreCmd(read func() (app.CoreStatus, error)) tea.Cmd {
	return func() tea.Msg {
		if read == nil {
			return chatCoreMsg{}
		}
		s, e := read()
		return chatCoreMsg{s, e}
	}
}

func chatAgentsCmd(read func() ([]app.AgentChange, error)) tea.Cmd {
	return func() tea.Msg {
		if read == nil {
			return chatAgentsMsg{}
		}
		s, e := read()
		return chatAgentsMsg{s, e}
	}
}

func chatGitCmd(read func() (app.GitInfo, bool)) tea.Cmd {
	if read == nil {
		return nil
	}
	return func() tea.Msg {
		info, ok := read()
		return chatGitMsg{info: info, ok: ok}
	}
}

func (m chatModel) say(text string) chatModel {
	m.queue = append(m.queue, chatAction{say: text})
	return m
}

func (m chatModel) tool(spec chatToolSpec) chatModel {
	m.queue = append(m.queue, chatAction{tool: &spec})
	return m
}

func (m chatModel) askFor(a chatAsk) chatModel {
	m.queue = append(m.queue, chatAction{ask: &a})
	return m
}
func (m chatModel) bye() chatModel { m.queue = append(m.queue, chatAction{quit: true}); return m }
func (m chatModel) streaming() bool {
	if len(m.entries) == 0 || m.entries[len(m.entries)-1].isTool {
		return false
	}
	e := m.entries[len(m.entries)-1]
	if e.runes != nil {
		return e.shown < len(e.runes)
	}
	return e.shown < len([]rune(e.text))
}

func (m chatModel) locked() bool {
	return m.streaming() || m.beat > 0 || m.running != nil || (m.ask == nil && len(m.queue) > 0)
}

func chatWorkLabel(phases []string, elapsed time.Duration) string {
	if len(phases) == 0 {
		return "Trabalhando…"
	}
	return phases[min(len(phases)-1, max(0, int(elapsed/chatPhaseDuration)))]
}

func chatStreamSchedule(runes []rune) []time.Duration {
	cps := max(chatStreamCPS, (len(runes)+chatStreamBudgetSeconds-1)/chatStreamBudgetSeconds)
	scale := max(0.3, min(1, float64(chatStreamCPS)/float64(cps)))
	at := make([]time.Duration, len(runes))
	var elapsed time.Duration
	start, word, last, pendingPause, index := 0, 0, rune(0), 0, 0
	for i := 0; i < len(runes); {
		if runes[i] == '\x1b' && i+1 < len(runes) && runes[i+1] == '[' {
			i += 2
			for i < len(runes) && (runes[i] < '@' || runes[i] > '~') {
				i++
			}
			if i < len(runes) {
				i++
			}
			continue
		}
		if runes[i] == ' ' || runes[i] == '\n' {
			if last != 0 {
				pause := 0
				switch last {
				case '.', '!', '?', '…':
					pause = 280
				case ':':
					pause = 170
				case ',', ';':
					pause = 120
				}
				pendingPause = max(pendingPause, pause)
				hash := uint32(index+1)*2654435761 ^ uint32(last)*2246822519
				hash ^= hash >> 16
				jitter := 0.6 + float64(hash%801)/1000
				elapsed += time.Duration(float64(time.Second) / float64(cps) * float64(word) * jitter)
				for j := start; j < i; j++ {
					at[j] = elapsed
				}
				word, last = 0, 0
				index++
			}
			if runes[i] == '\n' {
				pendingPause = max(pendingPause, 180)
			}
			i++
			continue
		}
		if word == 0 {
			elapsed += time.Duration(float64(pendingPause) * scale * float64(time.Millisecond))
			pendingPause, start = 0, i
		}
		word++
		last = runes[i]
		i++
	}
	if word > 0 {
		hash := uint32(index+1)*2654435761 ^ uint32(last)*2246822519
		hash ^= hash >> 16
		jitter := 0.6 + float64(hash%801)/1000
		elapsed += time.Duration(float64(time.Second) / float64(cps) * float64(word) * jitter)
		for j := start; j < len(runes); j++ {
			at[j] = elapsed
		}
	}
	next := elapsed
	for i := len(at) - 1; i >= 0; i-- {
		if at[i] == 0 {
			at[i] = next
		} else {
			next = at[i]
		}
	}
	if len(at) > 0 && at[len(at)-1] > chatStreamBudgetSeconds*time.Second {
		total := at[len(at)-1]
		for i := range at {
			at[i] = time.Duration(float64(at[i]) * float64(chatStreamBudgetSeconds*time.Second) / float64(total))
		}
	}
	return at
}

func chatStreamTick() tea.Cmd {
	return tea.Tick(time.Second/60, func(t time.Time) tea.Msg {
		return motion.FrameMsg{ID: chatStreamFrame, Time: t}
	})
}

func (m chatModel) pump() (chatModel, tea.Cmd) {
	if m.running != nil || m.streaming() || m.beat > 0 || len(m.queue) == 0 {
		return m, nil
	}
	if m.ask == nil && m.turnStart.IsZero() {
		m.turnStart = time.Now()
	}
	action := m.queue[0]
	m.queue = m.queue[1:]
	switch {
	case action.banner != "":
		m.entries = append(m.entries, chatEntry{banner: action.banner, info: action.info})
		return m.pump()
	case action.tool != nil:
		spec := action.tool
		i := len(m.entries)
		m.entries = append(m.entries, chatEntry{isTool: true, tool: &components.ToolBlock{Title: spec.title, Params: spec.params, State: theme.StateRunning}})
		m.running = &chatRun{entry: i, started: time.Now(), write: spec.write, wasWrite: spec.write, phases: spec.phases}
		ch := make(chan tea.Msg, 8)
		m.channel = ch
		go func() {
			res := spec.run(func(line string) { ch <- chatProgressMsg{entry: i, line: line} })
			ch <- chatFinishedMsg{entry: i, result: res, then: spec.then, fill: spec.fill}
			close(ch)
		}()
		if m.common.Motion {
			return m, tea.Batch(installListen(ch), m.spin.Tick, motion.Frame(chatFrame))
		}
		return m, tea.Batch(installListen(ch), m.spin.Tick)
	case action.ask != nil:
		a := *action.ask
		a.mark = len(m.entries)
		a.step = m.step
		a.context = m.context
		m.ask = &a
		m.history = append(m.history, a)
		m.input.Placeholder = "Escreva e tecle enter"
		m.input.EchoMode = textinput.EchoNormal
		if a.secret {
			m.input.EchoMode = textinput.EchoPassword
			m.input.EchoCharacter = '•'
		}
		if a.hasChoice {
			m.input.Placeholder = "ou digite um comando (/ajuda)"
		}
		m.input.Reset()
		m.input.Focus()
		if a.path {
			m.ask.picker = a.picker.Sync("~/")
		}
		if a.textarea {
			m.area.SetValue("")
			return m, m.area.Focus()
		}
		return m, nil
	case action.quit:
		return m, tea.Quit
	default:
		shown := len([]rune(action.say))
		var runes []rune
		var at []time.Duration
		if m.common.Motion {
			runes = []rune(action.say)
			at = chatStreamSchedule(runes)
			shown = 0
		}
		m.entries = append(m.entries, chatEntry{role: components.ChatAssistant, text: action.say, runes: runes, at: at, shown: shown})
		if m.common.Motion && len(runes) > 0 {
			m.streamStart, m.streamLast = time.Time{}, time.Time{}
			return m, tea.Batch(chatStreamTick(), m.spin.Tick, motion.Frame(chatFrame))
		}
		return m.pump()
	}
}

func (m chatModel) finishTool() (chatModel, tea.Cmd) {
	run := m.running
	m = m.fillTool(time.Now())
	if run.completion.fill != nil {
		return m, motion.Frame(chatFrame)
	}
	msg := run.completion
	e := &m.entries[run.entry]
	e.tool.Elapsed = time.Since(run.started)
	e.tool.State = msg.result.state
	e.tool.Rows = append(e.tool.Rows, msg.result.rows[min(run.displayed, len(msg.result.rows)):]...)
	e.tool.Lines = append(e.tool.Lines, msg.result.lines[max(0, run.displayed-len(msg.result.rows)):]...)
	m.running = nil
	if e.tool.Title == "Instalar plugins" {
		e.tool.Params = ""
	}
	m.channel = nil
	if m.signal != nil {
		return m, tea.Quit
	}
	var refresh tea.Cmd
	if run.wasWrite && msg.result.err == nil {
		if m.deps.Core != nil {
			m.core, m.coreErr = m.deps.Core()
		}
		refresh = tea.Batch(chatAgentsCmd(m.deps.Agents), chatGitCmd(m.deps.Git))
	}
	if msg.then != nil {
		m = msg.then(m, msg.result)
	}
	var cmd tea.Cmd
	m, cmd = m.pump()
	return m, tea.Batch(refresh, cmd)
}

func (m chatModel) fillTool(now time.Time) chatModel {
	if m.running == nil || m.running.completion == nil || m.running.completion.fill == nil {
		return m
	}
	if (!m.coreDone || !m.agentsDone) && now.Sub(m.running.started) < 8*time.Second {
		return m
	}
	msg := m.running.completion
	filled := msg.fill(m)
	msg.result.rows = append(msg.result.rows, filled.rows...)
	msg.result.lines = append(msg.result.lines, filled.lines...)
	msg.fill = nil
	return m
}

func (m chatModel) finishSay() (chatModel, tea.Cmd) {
	e := &m.entries[len(m.entries)-1]
	e.at = nil
	m.streamStart, m.streamLast = time.Time{}, time.Time{}
	if m.common.Motion && len(m.queue) > 0 {
		m.beat = chatBeatDuration
		if len(e.runes) > 300 {
			m.beat += 250 * time.Millisecond
		}
		m.beatStart = time.Time{}
		return m, motion.Frame(chatFrame)
	}
	return m.pump()
}

func (m chatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case chatBeginMsg:
		return m.pump()
	case tea.WindowSizeMsg:
		m.width, m.height = max(1, msg.Width), max(1, msg.Height)
		mainW := m.width
		if m.width >= 120 && m.height >= 30 {
			mainW -= 33
		}
		m.input.SetWidth(max(1, mainW-6))
		m.area.SetWidth(max(1, mainW-6))
	case tea.BackgroundColorMsg:
		m.theme = theme.New(theme.Options{Dark: msg.IsDark(), Profile: m.common.Profile, Plain: m.common.Profile == colorprofile.ASCII, NoPowerline: !m.common.Powerline})
	case tea.FocusMsg:
		m.focused = true
		if !m.pauseAt.IsZero() {
			if !m.streamStart.IsZero() {
				m.streamStart = m.streamStart.Add(time.Since(m.pauseAt))
				m.streamLast = time.Time{}
			}
			if !m.beatStart.IsZero() {
				m.beatStart = m.beatStart.Add(time.Since(m.pauseAt))
			}
			m.pauseAt = time.Time{}
		}
		if m.locked() {
			if m.common.Motion {
				if m.streaming() {
					return m, tea.Batch(m.spin.Tick, chatStreamTick(), motion.Frame(chatFrame))
				}
				return m, tea.Batch(m.spin.Tick, motion.Frame(chatFrame))
			}
			return m, m.spin.Tick
		}
	case tea.BlurMsg:
		m.focused = false
		m.pauseAt = time.Now()
	case chatCoreMsg:
		m.coreDone, m.core, m.coreErr = true, msg.status, msg.err
		m = m.fillTool(time.Now())
		if m.running != nil && m.running.completion != nil && m.running.completion.fill == nil && !m.common.Motion {
			return m.finishTool()
		}
	case chatAgentsMsg:
		m.agentsDone, m.agents, m.agentsErr = true, msg.changes, msg.err
		m = m.fillTool(time.Now())
		if m.running != nil && m.running.completion != nil && m.running.completion.fill == nil && !m.common.Motion {
			return m.finishTool()
		}
	case chatGitMsg:
		m.git, m.gitOK = msg.info, msg.ok
	case InterruptDeferredMsg:
		if m.Busy() {
			m.signal = msg.Signal
			m.entries = append(m.entries, chatEntry{role: components.ChatNote, text: "Gravação em andamento: Ctrl+C adiado até concluir", shown: 1000})
			return m, nil
		}
		m.interrupted = true
		return m, tea.Quit
	case chatProgressMsg:
		if m.running != nil && m.running.entry == msg.entry {
			e := &m.entries[msg.entry]
			if e.tool.Title == "Instalar plugins" && strings.HasPrefix(msg.line, "Instalando ") {
				e.tool.Params = msg.line
			} else {
				e.tool.Lines = append(e.tool.Lines, msg.line)
			}
			return m, installListen(m.channel)
		}
	case chatFinishedMsg:
		if m.running == nil || m.running.entry != msg.entry {
			return m, nil
		}
		m.running.write = false
		m.running.completion = &msg
		m = m.fillTool(time.Now())
		if m.signal != nil {
			return m, tea.Quit
		}
		if !m.common.Motion {
			if m.running.completion.fill != nil {
				return m, motion.Frame(chatFrame)
			}
			return m.finishTool()
		}
		return m, nil
	case motion.FrameMsg:
		if !m.focused {
			return m, nil
		}
		if msg.ID == chatStreamFrame {
			if !m.streaming() {
				return m, nil
			}
			e := &m.entries[len(m.entries)-1]
			if m.streamStart.IsZero() {
				m.streamStart = msg.Time
			}
			if !m.streamLast.IsZero() && msg.Time.Sub(m.streamLast) > 250*time.Millisecond {
				m.streamStart = m.streamStart.Add(msg.Time.Sub(m.streamLast) - 250*time.Millisecond)
			}
			m.streamLast = msg.Time
			elapsed := msg.Time.Sub(m.streamStart)
			for e.shown < len(e.at) && e.at[e.shown] <= elapsed {
				e.shown++
			}
			if m.streaming() {
				return m, chatStreamTick()
			}
			return m.finishSay()
		}
		if msg.ID != chatFrame {
			return m, nil
		}
		m.reveal = m.reveal.Step()
		if m.beat > 0 {
			if m.beatStart.IsZero() {
				m.beatStart = msg.Time
			}
			if msg.Time.Sub(m.beatStart) < m.beat {
				return m, motion.Frame(chatFrame)
			}
			m.beat = 0
			return m.pump()
		}
		if m.running != nil && (m.common.Motion || m.running.completion != nil && m.running.completion.fill != nil) {
			m = m.fillTool(msg.Time)
			run := m.running
			if run.completion != nil {
				if run.completion.fill != nil {
					return m, motion.Frame(chatFrame)
				}
				if !m.common.Motion {
					return m.finishTool()
				}
				if m.entries[run.entry].tool.Title == "Instalar plugins" && len(run.phases) == 0 {
					return m.finishTool()
				}
				retention := min(time.Duration(max(len(run.phases), 1))*chatPhaseDuration, chatMaxRetention)
				elapsed := msg.Time.Sub(run.started)
				total := len(run.completion.result.rows) + len(run.completion.result.lines)
				next := min(total, max(0, int(int64(total)*int64(elapsed)/int64(retention))))
				if elapsed >= retention {
					next = total
				}
				if next > run.displayed {
					rows := run.completion.result.rows
					from, to := min(run.displayed, len(rows)), min(next, len(rows))
					m.entries[run.entry].tool.Rows = append(m.entries[run.entry].tool.Rows, rows[from:to]...)
					lines := run.completion.result.lines
					from, to = max(0, run.displayed-len(rows)), max(0, next-len(rows))
					m.entries[run.entry].tool.Lines = append(m.entries[run.entry].tool.Lines, lines[from:to]...)
					run.displayed = next
				}
				if elapsed >= retention {
					return m.finishTool()
				}
			}
			return m, motion.Frame(chatFrame)
		}
		if !m.reveal.Settled() {
			return m, motion.Frame(chatFrame)
		}
	case spinner.TickMsg:
		if !m.focused || !m.locked() {
			return m, nil
		}
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)
		return m, cmd
	case tea.MouseWheelMsg:
		if !m.locked() && m.paletteOpen {
			m.placeChoice()
			m.palette, _ = m.palette.Update(msg)
			return m, nil
		}
		if m.ask != nil && !m.locked() && (m.ask.hasChoice || m.ask.path) {
			m.placeChoice()
			if m.ask.path {
				m.ask.picker, _ = m.ask.picker.Update(msg)
			} else {
				m.ask.choice, _ = m.ask.choice.Update(msg)
			}
			return m, nil
		}
		if msg.Button == tea.MouseWheelUp {
			m.scroll += 3
		} else if msg.Button == tea.MouseWheelDown {
			m.scroll = max(0, m.scroll-3)
		}
	case tea.MouseClickMsg:
		if m.phase == chatWelcome {
			if m.coreDone && m.agentsDone {
				m.phase = chatTalking
				m = m.flowGreeting(false)
				return m.pump()
			}
			return m, nil
		}
		_, hits := chatLayout(m)
		for _, hit := range hits {
			if hit.y == msg.Y {
				m.entries[hit.entry].tool.Expanded = !m.entries[hit.entry].tool.Expanded
				return m, nil
			}
		}
		if !m.locked() {
			m.placeChoice()
			if m.paletteOpen {
				var cmd tea.Cmd
				m.palette, cmd = m.palette.Update(msg)
				return m.dispatchChoice(cmd)
			}
			if m.ask != nil && m.ask.path {
				var cmd tea.Cmd
				m.ask.picker, cmd = m.ask.picker.Update(msg)
				return m.dispatchChoice(cmd)
			}
			if m.ask != nil && m.ask.hasChoice {
				var cmd tea.Cmd
				m.ask.choice, cmd = m.ask.choice.Update(msg)
				return m.dispatchChoice(cmd)
			}
		}
	case tea.MouseMotionMsg:
		if !m.locked() {
			m.placeChoice()
			if m.paletteOpen {
				m.palette, _ = m.palette.Update(msg)
			} else if m.ask != nil && m.ask.hasChoice {
				m.ask.choice, _ = m.ask.choice.Update(msg)
			}
		}
	case tea.PasteMsg:
		if !m.locked() {
			if m.ask != nil && m.ask.textarea {
				m.area, _ = m.area.Update(msg)
			} else {
				msg.Content = strings.NewReplacer("\r\n", " ", "\n", " ", "\r", " ").Replace(msg.Content)
				var cmd tea.Cmd
				m.input, cmd = m.input.Update(msg)
				if m.ask != nil && m.ask.path {
					m.ask.picker = m.ask.picker.Sync(m.input.Value())
				}
				m.syncPalette()
				return m, cmd
			}
		}
		return m, nil
	case tea.KeyPressMsg:
		return m.key(msg)
	case components.ChoiceDoneMsg:
		return m.answer(chatAnswer{IDs: msg.IDs})
	case components.ChoiceBackMsg:
		return m.back()
	case components.PathPickMsg:
		if m.ask != nil && m.ask.path {
			m.input.SetValue(msg.Path)
			m.input.CursorEnd()
			m.ask.picker = m.ask.picker.Sync(msg.Path)
		}
		return m, nil
	case components.PaletteChooseMsg:
		intent, special, _ := chatIntentFor("/" + msg.Name)
		return m.command("/"+msg.Name, intent, special)
	}
	return m, nil
}

func (m chatModel) dispatchChoice(cmd tea.Cmd) (tea.Model, tea.Cmd) {
	if cmd == nil {
		return m, nil
	}
	return m.Update(cmd())
}

func (m chatModel) key(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	k := msg.String()
	if k == "ctrl+c" {
		if m.Busy() {
			m.signal = os.Interrupt
			m.entries = append(m.entries, chatEntry{role: components.ChatNote, text: "Gravação em andamento: Ctrl+C adiado até concluir", shown: 1000})
			return m, nil
		}
		m.interrupted = true
		return m, tea.Quit
	}
	if k == "ctrl+o" {
		for i := len(m.entries) - 1; i >= 0; i-- {
			if m.entries[i].isTool {
				m.entries[i].tool.Expanded = !m.entries[i].tool.Expanded
				break
			}
		}
		return m, nil
	}
	if k == "ctrl+z" {
		return m, tea.Suspend
	}
	if m.phase == chatWelcome {
		if k == "enter" && m.coreDone && m.agentsDone {
			m.phase = chatTalking
			m = m.flowGreeting(false)
			return m.pump()
		}
		return m, nil
	}
	space := msg.Code == ' ' || msg.Text == " " || k == "space" || k == " "
	if m.locked() {
		if k == "enter" || space {
			if m.streaming() {
				e := &m.entries[len(m.entries)-1]
				e.shown = len(e.runes)
				e.at = nil
				m.streamStart, m.streamLast = time.Time{}, time.Time{}
				return m.pump()
			}
			if m.beat > 0 {
				m.beat = 0
				m.beatStart = time.Time{}
				return m.pump()
			}
			if m.running != nil && m.running.completion != nil {
				return m.finishTool()
			}
		}
		return m, nil
	}
	if m.ask != nil && m.ask.secret && k == "esc" {
		m.input.Reset()
		m.input.EchoMode = textinput.EchoNormal
		clear(m.addonSecrets)
		m.addonSecrets = nil
		m.addonPlan = app.AddonPlan{}
		if m.context == "addons" {
			m = m.flowAddons()
		} else {
			m.addonNames = nil
			m = m.setupPlugins()
		}
		return m.pump()
	}
	if k == "esc" {
		if m.paletteOpen {
			m.paletteOpen = false
			m.input.Reset()
			return m, nil
		}
		if m.ask != nil && m.ask.textarea {
			if m.area.Value() != "" {
				m.area.SetValue("")
				return m, nil
			}
			return m.back()
		}
		if m.input.Value() != "" {
			m.input.Reset()
			if m.ask != nil && m.ask.path {
				m.ask.picker = m.ask.picker.Sync("~/")
			}
			return m, nil
		}
		return m.back()
	}
	if m.ask != nil && m.ask.textarea {
		if k == "enter" {
			text := strings.TrimSpace(m.area.Value())
			if strings.HasPrefix(text, "/") {
				if intent, special, ok := chatIntentFor(text); ok && (special == "voltar" || special == "sair") {
					m.area.SetValue("")
					return m.command(text, intent, special)
				}
			}
			if m.ask.validate != nil {
				if err := m.ask.validate(text); err != nil {
					m = m.say("Não dá para usar isso: " + err.Error())
					return m.pump()
				}
			}
			m.area.SetValue("")
			return m.answer(chatAnswer{Text: text})
		}
		if k == "shift+enter" {
			msg = tea.KeyPressMsg{Code: tea.KeyEnter}
		}
		var cmd tea.Cmd
		m.area, cmd = m.area.Update(msg)
		return m, cmd
	}
	if m.ask != nil && m.ask.path {
		if k == "up" || k == "down" {
			m.ask.picker, _ = m.ask.picker.Update(msg)
			return m, nil
		}
		if k == "tab" {
			value := m.ask.picker.Complete(m.input.Value())
			m.input.SetValue(value)
			m.input.CursorEnd()
			m.ask.picker = m.ask.picker.Sync(value)
			return m, nil
		}
		if k == "enter" {
			if item, ok := m.ask.picker.Highlighted(); ok {
				m.input.SetValue(item.Path)
				m.input.CursorEnd()
				m.ask.picker = m.ask.picker.Sync(item.Path)
				return m, nil
			}
		}
	}
	if m.ask != nil && m.ask.hasChoice && m.input.Value() == "" {
		m.placeChoice()
	}
	pageKey := k == "left" || k == "right" || k == "pgup" || k == "pgdown"
	if m.ask != nil && m.ask.hasChoice && m.input.Value() == "" && (k == "up" || k == "down" || k == "j" || k == "k" || pageKey && m.ask.choice.Pages() > 1 || space || k == "enter" || len(k) == 1 && k[0] >= '1' && k[0] <= '9') {
		m.input.Reset()
		var cmd tea.Cmd
		m.ask.choice, cmd = m.ask.choice.Update(msg)
		return m.dispatchChoice(cmd)
	}
	if k == "pgup" {
		m.scroll += max(1, m.height-2)
		return m, nil
	}
	if k == "pgdown" {
		m.scroll = max(0, m.scroll-max(1, m.height-2))
		return m, nil
	}
	if m.paletteOpen && (k == "up" || k == "down") {
		m.palette, _ = m.palette.Update(msg)
		return m, nil
	}
	if m.paletteOpen && k == "tab" {
		if item, ok := m.palette.Selected(); ok {
			m.input.SetValue("/" + item.Name + " ")
			m.input.CursorEnd()
			m.syncPalette()
		}
		return m, nil
	}
	if k == "enter" {
		if m.ask != nil && m.ask.secret {
			text := m.input.Value()
			m.input.Reset()
			m.input.EchoMode = textinput.EchoNormal
			return m.answer(chatAnswer{Text: text})
		}
		text := strings.TrimSpace(m.input.Value())
		if m.paletteOpen {
			if _, _, ok := chatIntentFor(text); !ok {
				if item, selected := m.palette.Selected(); selected {
					intent, special, _ := chatIntentFor("/" + item.Name)
					return m.command("/"+item.Name, intent, special)
				}
			}
		}
		m.input.Reset()
		m.paletteOpen = false
		if text == "" {
			return m, nil
		}
		if m.ask != nil && m.ask.input {
			if strings.HasPrefix(text, "/") {
				if intent, special, ok := chatIntentFor(text); ok && (special == "voltar" || special == "sair") {
					return m.command(text, intent, special)
				}
			}
			if m.ask.path {
				text = components.ExpandHome(text, m.common.UserHome)
				if filepath.IsAbs(text) {
					text = filepath.Clean(text)
				}
			}
			if m.ask.validate != nil {
				if err := m.ask.validate(text); err != nil {
					if m.ask.path {
						m.ask.picker = m.ask.picker.Sync("~/")
					}
					m = m.say("Não dá para usar isso: " + err.Error())
					return m.pump()
				}
			}
			return m.answer(chatAnswer{Text: text})
		}
		if intent, special, ok := chatIntentFor(text); ok {
			return m.command(text, intent, special)
		}
		m = m.say("Não entendi. Digite / para ver os comandos.")
		return m.pump()
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	if m.ask != nil && m.ask.path {
		m.ask.picker = m.ask.picker.Sync(m.input.Value())
	}
	m.syncPalette()
	return m, cmd
}

func (m chatModel) answer(ans chatAnswer) (tea.Model, tea.Cmd) {
	if m.ask == nil {
		return m, nil
	}
	a := *m.ask
	label := ans.Text
	if a.secret {
		label = "••••••••"
	}
	if len(ans.IDs) > 0 {
		var names []string
		for _, id := range ans.IDs {
			name := id
			for _, o := range a.options {
				if o.ID == id {
					name = o.Label
					break
				}
			}
			names = append(names, name)
		}
		label = strings.Join(names, ", ")
	}
	if label == "" && a.hasChoice {
		label = "Nenhum item adicional"
		if a.question == "Além do Core (obrigatório), o que mais você quer instalar?" {
			label = "Só o Core"
		}
	}
	m.entries = append(m.entries, chatEntry{role: components.ChatUser, text: label, shown: len([]rune(label))})
	if a.hasChoice && len(m.history) > 0 {
		m.history[len(m.history)-1].choice = a.choice.Preselect(ans.IDs)
	}
	m.ask = nil
	m.scroll = 0
	m.turnStart = time.Now()
	if a.onAnswer != nil {
		m = a.onAnswer(m, ans)
	}
	return m.pump()
}

func (m chatModel) back() (tea.Model, tea.Cmd) {
	if len(m.history) > 1 {
		m.history = m.history[:len(m.history)-1]
		a := m.history[len(m.history)-1]
		m.ask = &a
		m.step = a.step
		m.context = a.context
		m.entries = m.entries[:min(a.mark, len(m.entries))]
		m.queue = nil
		clear(m.addonSecrets)
		m.addonSecrets = nil
		m.addonPlan = app.AddonPlan{}
		m.input.Reset()
		m.input.Placeholder = "Escreva e tecle enter"
		if a.hasChoice {
			m.input.Placeholder = "ou digite um comando (/ajuda)"
		}
		m.input.EchoMode = textinput.EchoNormal
		if a.secret {
			m.input.EchoMode = textinput.EchoPassword
			m.input.EchoCharacter = '•'
		}
		if a.path {
			m.ask.picker = a.picker.Sync("~/")
		}
		m.area.SetValue("")
		return m, nil
	}
	if len(m.steps) > 0 || m.ask != nil {
		m.queue = nil
		m = m.flowIdle()
		m = m.say("Cancelado. Digite / para ver os comandos.")
		return m.pump()
	}
	m = m.say("Use /sair ou ctrl+c para sair.")
	return m.pump()
}

func (m chatModel) flowGreeting(short bool) chatModel {
	if !short {
		m = m.say("Olá! Eu sou o assistente do Kit GEX. Vou configurar seu ambiente de desenvolvimento com você.")
	}
	m = m.tool(chatToolSpec{title: "Verificar ambiente", params: homeLabel(m.common.Home, m.common.UserHome), phases: []string{"Checando sua pasta home…", "Procurando o Claude Code…", "Procurando o Codex…", "Lendo o Core instalado…"}, run: func(_ func(string)) chatToolResult { return chatToolResult{state: theme.StateOK} }, fill: func(m chatModel) chatToolResult {
		rows := []components.ToolRow{{State: theme.StateOK, Label: "HOME", Detail: homeLabel(m.common.Home, m.common.UserHome)}}
		for _, name := range []string{"claude", "codex"} {
			label := "Claude Code"
			if name == "codex" {
				label = "Codex"
			}
			row := components.ToolRow{State: theme.StateInfo, Label: label, Detail: "não detectado"}
			change, found := homeAgentChange(m.agents, name)
			switch {
			case m.agentsErr != nil:
				row.State, row.Detail = theme.StateError, "erro ao ler: "+maskHome(m.agentsErr.Error(), m.common.UserHome)
			case !m.agentsDone:
				row.State, row.Detail = theme.StateError, "erro ao ler: tempo esgotado"
			case found:
				row.State = homeAgentState(change)
				switch {
				case change.Blocked:
					row.Detail = "bloqueado"
					if change.Reason != "" {
						row.Detail += " · " + maskHome(change.Reason, m.common.UserHome)
					}
				case change.Current == "":
					row.Detail = "não instalado"
					if change.Target != "" {
						row.Detail += " · mais recente: " + maskHome(change.Target, m.common.UserHome)
					}
				case change.Current != change.Target:
					row.Detail = maskHome(change.Current, m.common.UserHome)
					if change.Target != "" {
						row.Detail += " · mais recente: " + maskHome(change.Target, m.common.UserHome)
					}
				default:
					row.Detail = maskHome(change.Current, m.common.UserHome) + " · em dia"
				}
			}
			rows = append(rows, row)
		}
		core := components.ToolRow{State: homeCoreState(m.core.State), Label: "Core", Detail: fmt.Sprintf("%d/%d skills · %s", m.core.ActiveCore, len(catalog.CoreSkills()), m.core.State)}
		if m.coreErr != nil {
			core.State, core.Detail = theme.StateError, "erro ao ler: "+maskHome(m.coreErr.Error(), m.common.UserHome)
		} else if !m.coreDone {
			core.State, core.Detail = theme.StateError, "erro ao ler: tempo esgotado"
		} else if m.core.State == "" {
			core.State, core.Detail = theme.StateInfo, fmt.Sprintf("%d/%d skills · não detectado", m.core.ActiveCore, len(catalog.CoreSkills()))
		}
		return chatToolResult{rows: append(rows, core)}
	}, then: func(m chatModel, _ chatToolResult) chatModel {
		if m.coreErr != nil {
			m = m.say("Não consegui ler o Core: " + m.coreErr.Error())
		} else {
			m = m.say(coreGreeting(m.core))
		}
		if short {
			return m.startIntent(m.start)
		}
		return m.flowIdle().say("Digite / para ver o que posso fazer. Para configurar o ambiente, use /setup.")
	}})
	return m
}

func coreGreeting(core app.CoreStatus) string {
	total := len(catalog.CoreSkills())
	switch {
	case core.Attention > 0:
		return fmt.Sprintf("Encontrei o Core com %d arquivo(s) que precisam de atenção (%d/%d skills em dia). Rode /doctor para ver quais.", core.Attention, core.ActiveCore, total)
	case core.ActiveCore > 0:
		return fmt.Sprintf("Encontrei o Core instalado (%d/%d skills).", core.ActiveCore, total)
	}
	return "Ainda não há Core instalado neste HOME."
}

func (m chatModel) flowIdle() chatModel {
	m.ask = nil
	m.history = nil
	m.steps = nil
	m.step = 0
	m.context = "pronto"
	m.paletteOpen = false
	m.input.Reset()
	m.input.EchoMode = textinput.EchoNormal
	m.input.Placeholder = "Digite / para ver os comandos"
	clear(m.addonSecrets)
	m.addonSecrets = nil
	m.addonPlan = app.AddonPlan{}
	m.plan = app.Plan{}
	return m
}

func (m *chatModel) syncPalette() {
	if m.ask == nil || !m.ask.input {
		m.paletteOpen = strings.HasPrefix(m.input.Value(), "/")
		if m.paletteOpen {
			m.palette = m.palette.Filter(m.input.Value())
		}
	}
}

func (m chatModel) flowMenu() chatModel {
	m = m.flowIdle()
	question := "O que você quer fazer?"
	label := "Setup do zero"
	if m.core.ActiveCore > 0 {
		label = "Atualizar setup"
	}
	opts := []components.ChoiceOption{{ID: "setup", Label: label}, {ID: "catalog", Label: "Ver catálogo de skills e hooks"}, {ID: "doctor", Label: "Verificar integridade (doctor)"}, {ID: "rollback", Label: "Desfazer a última instalação"}, {ID: "rollback-agents", Label: "Restaurar versões anteriores dos agentes"}, {ID: "githooks", Label: "Configurar git hooks de um repositório"}, {ID: "addons", Label: "Ver addons opcionais"}, {ID: "exit", Label: "Sair"}}
	return m.askFor(chatAsk{hasChoice: true, choice: components.NewChoice(question, opts, false, true), options: opts, question: question, onAnswer: func(m chatModel, a chatAnswer) chatModel {
		m.history = nil
		switch a.IDs[0] {
		case "setup":
			return m.flowSetup()
		case "catalog":
			return m.flowCatalog()
		case "doctor":
			return m.flowDoctor()
		case "rollback":
			return m.flowRollback(false)
		case "rollback-agents":
			return m.flowRollback(true)
		case "githooks":
			return m.flowGitHooks()
		case "addons":
			return m.flowAddons()
		default:
			return m.say("Até mais!").bye()
		}
	}})
}

func (m chatModel) startIntent(intent ChatIntent) chatModel {
	switch intent {
	case IntentSetup:
		return m.flowSetup()
	case IntentCatalog:
		return m.flowCatalog()
	case IntentAddons:
		return m.flowAddons()
	case IntentDoctor:
		return m.flowDoctor()
	case IntentRollback:
		return m.flowRollback(false)
	case IntentRollbackAgents:
		return m.flowRollback(true)
	case IntentGitHooks:
		return m.flowGitHooks()
	}
	return m
}

func (m chatModel) command(text string, intent ChatIntent, special string) (tea.Model, tea.Cmd) {
	m.paletteOpen = false
	m.input.Reset()
	if special == "voltar" {
		return m.back()
	}
	m.entries = append(m.entries, chatEntry{role: components.ChatUser, text: text, shown: len([]rune(text))})
	m.turnStart = time.Now()
	switch special {
	case "sair":
		m.ask = nil
		m.queue = nil
		m = m.say("Até mais!").bye()
	case "limpar":
		m.entries = nil
	case "menu":
		m = m.flowMenu()
	case "ajuda":
		var lines []string
		for _, item := range chatCommands {
			lines = append(lines, "/"+item.Name+" — "+item.Description)
		}
		m = m.say(strings.Join(lines, "\n") + "\nesc voltar · ctrl+o expandir · ctrl+c sair")
	default:
		if intent != IntentHome {
			m.ask = nil
			m.history = nil
			m.steps = nil
			m.queue = nil
			m = m.startIntent(intent)
		} else {
			m = m.say("Não entendi. Digite / para ver os comandos.")
		}
	}
	return m.pump()
}
