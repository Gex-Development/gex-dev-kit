package screens

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"charm.land/bubbles/v2/filepicker"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"

	"github.com/Gex-Development/gex-dev-kit/internal/app"
	"github.com/Gex-Development/gex-dev-kit/internal/catalog"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/components"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/layout"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

type GitHooksDeps struct {
	Plan  func(app.GitHooksRequest) (app.GitHooksPlan, error)
	Apply func(app.GitHooksPlan) (app.Result, error)
}

type githooksPhase int

const (
	ghEdit githooksPhase = iota
	ghPlanning
	ghPreview
	ghWriting
	ghDone
)

const (
	ghMaxCommands   = 8
	ghMaxCommandLen = 300
	ghToastTime     = 3 * time.Second
	ghPreCommit     = "gex-pre-commit"
	ghPrePush       = "gex-pre-push"
)

var ghStepNames = map[int]string{
	1: "Passo 1 de 4 · Repositório",
	2: "Passo 2 de 4 · Hooks",
	3: "Passo 3 de 4 · Comandos",
	4: "Passo 4 de 4 · Prévia",
}

type githooksRow struct {
	name        string
	label       string
	description string
	marked      bool
}

type githooksPlanned struct {
	plan app.GitHooksPlan
	err  error
}

type githooksWritten struct {
	result app.Result
	err    error
}

type githooksModel struct {
	c      Common
	deps   GitHooksDeps
	writes *sync.WaitGroup
	t      theme.Theme
	size   layout.Size
	keys   components.Keys
	toast  components.Toast
	help   bool

	step    int
	phase   githooksPhase
	repo    string
	repoErr string
	path    textinput.Model
	picker  filepicker.Model
	picking bool

	hooks   []githooksRow
	cursor  int
	hookErr string

	commit textarea.Model
	push   textarea.Model
	onPush bool
	cmdErr string

	request     app.GitHooksRequest
	planned     githooksPlanned
	preview     viewport.Model
	asking      bool
	confirm     components.Confirm
	progress    components.StepProgress
	written     githooksWritten
	signal      os.Signal
	interrupted bool
}

func RunGitHooks(r Runner, c Common, d GitHooksDeps, preset app.GitHooksRequest) (app.Result, error) {
	var writes sync.WaitGroup
	wrapped := d
	wrapped.Apply = func(p app.GitHooksPlan) (app.Result, error) {
		defer writes.Done()
		return d.Apply(p)
	}
	final, err := r.Run(newGitHooksModel(c, wrapped, &writes, preset))
	writes.Wait()
	if err != nil {
		return app.Result{}, err
	}
	m, ok := final.(githooksModel)
	if !ok {
		return app.Result{}, errors.New("estado da tela de git hooks inesperado")
	}
	if m.signal != nil {
		return m.written.result, SignalError{Signal: m.signal}
	}
	if m.interrupted {
		return app.Result{}, SignalError{Signal: os.Interrupt}
	}
	return m.written.result, m.written.err
}

func newGitHooksModel(c Common, d GitHooksDeps, writes *sync.WaitGroup, preset app.GitHooksRequest) githooksModel {
	m := githooksModel{
		c:       c,
		deps:    d,
		writes:  writes,
		keys:    components.DefaultKeys(),
		toast:   components.NewToast(),
		step:    1,
		hooks:   githooksRows(preset),
		commit:  githooksArea(preset.PreCommit),
		push:    githooksArea(preset.PrePush),
		preview: viewport.New(viewport.WithWidth(1), viewport.WithHeight(1)),
		confirm: components.NewConfirm(
			"Gravar git hooks?",
			"Confira a prévia antes de confirmar.",
			"gravar",
			"cancelar",
		),
	}
	m = m.withTheme()
	m.progress = components.NewStepProgress(m.t, []string{"Gravar scripts e hooks", "Registrar no histórico"}, c.Motion)
	m.path = textinput.New()
	m.path.Placeholder = "/caminho/absoluto/do/repositorio"
	m.path.SetValue(preset.Repo)
	m.path.Focus()
	m.picker = filepicker.New()
	m.picker.DirAllowed = true
	m.picker.FileAllowed = false
	m.picker.AutoHeight = false
	m.picker.CurrentDirectory = githooksStartDir(c)
	if preset.Repo != "" {
		m.repo = preset.Repo
		m.step = 2
	}
	return m
}

func githooksStartDir(c Common) string {
	if c.UserHome != "" {
		return c.UserHome
	}
	return "."
}

func githooksArea(lines []string) textarea.Model {
	ta := textarea.New()
	ta.Placeholder = "make test"
	ta.ShowLineNumbers = false
	ta.SetHeight(4)
	ta.SetValue(strings.Join(lines, "\n"))
	return ta
}

func githooksRows(preset app.GitHooksRequest) []githooksRow {
	var rows []githooksRow
	for _, name := range catalog.GitHooks() {
		spec, ok := catalog.HookSpecByName(name)
		if !ok {
			continue
		}
		rows = append(rows, githooksRow{
			name:        name,
			label:       spec.GitHook,
			description: spec.Description,
			marked:      slices.Contains(preset.Hooks, name),
		})
	}
	return rows
}

func (m githooksModel) withTheme() githooksModel {
	m.t = theme.New(theme.Options{Dark: m.c.Dark, Profile: m.c.Profile, Plain: m.c.Profile <= colorprofile.ASCII})
	return m
}

func (m githooksModel) Busy() bool {
	return m.phase == ghWriting
}

func (m githooksModel) Init() tea.Cmd {
	return tea.Batch(m.picker.Init(), m.path.Focus())
}

func (m githooksModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.size = layout.Size{Width: msg.Width, Height: msg.Height}
		return m.resize(), nil
	case tea.BackgroundColorMsg:
		m.c.Dark = msg.IsDark()
		m = m.withTheme()
		m.progress = m.progress.WithTheme(m.t)
		return m, nil
	case tea.ColorProfileMsg:
		m.c.Profile = msg.Profile
		return m.withTheme(), nil
	case tea.KeyPressMsg:
		return m.key(msg)
	case githooksPlanned:
		m.planned = msg
		m.phase = ghPreview
		if msg.err == nil {
			m.preview.SetContent(m.previewText())
		}
		return m, nil
	case githooksWritten:
		return m.finishWrite(msg)
	case InterruptDeferredMsg:
		return m.interrupt(msg)
	}
	var c1, c2, c3, c4 tea.Cmd
	m.toast, c1 = m.toast.Update(msg)
	m.progress, c2 = m.progress.Update(msg)
	m.path, c3 = m.path.Update(msg)
	m.commit, c4 = m.commit.Update(msg)
	return m, tea.Batch(c1, c2, c3, c4)
}

func (m githooksModel) resize() githooksModel {
	m.preview.SetWidth(max(1, m.size.Width))
	m.preview.SetHeight(max(3, m.size.Height-8))
	m.path.SetWidth(max(10, m.size.Width-4))
	m.commit.SetWidth(max(10, m.size.Width-4))
	m.push.SetWidth(max(10, m.size.Width-4))
	m.picker.SetHeight(max(3, m.size.Height-10))
	return m
}

func (m githooksModel) notice(text string, kind theme.State) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.toast, cmd = m.toast.Show(text, kind, ghToastTime)
	return m, cmd
}

func (m githooksModel) interrupt(msg InterruptDeferredMsg) (tea.Model, tea.Cmd) {
	if m.phase != ghWriting {
		m.interrupted = true
		return m, tea.Quit
	}
	m.signal = msg.Signal
	return m.notice("Sinal recebido; concluindo a gravação para não deixar arquivos pela metade.", theme.StateWarn)
}

func (m githooksModel) textFocused() bool {
	return (m.step == 1 && !m.picking) || m.step == 3
}

func (m githooksModel) key(p tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if m.phase == ghWriting {
		return m.notice("Gravação em andamento; aguarde a conclusão.", theme.StateWarn)
	}
	if p.String() == "ctrl+c" {
		m.interrupted = true
		return m, tea.Quit
	}
	keys := m.keys
	if m.textFocused() {
		keys = keys.TextInput()
	}
	if key.Matches(p, keys.Quit) {
		return m, tea.Quit
	}
	if layout.Classify(m.size) == layout.ClassTiny {
		return m, nil
	}
	if m.asking {
		next, dec := m.confirm.Update(p, m.confirm.Visible(m.size.Width, m.size.Height))
		m.confirm = next
		switch dec {
		case components.DecisionYes:
			m.asking = false
			return m.startWrite()
		case components.DecisionNo:
			m.asking = false
		}
		return m, nil
	}
	if key.Matches(p, keys.Help) {
		m.help = !m.help
		return m, nil
	}
	if key.Matches(p, keys.Back) {
		return m.back()
	}
	switch m.step {
	case 1:
		return m.keyRepo(p, keys)
	case 2:
		return m.keyHooks(p, keys)
	case 3:
		return m.keyCommands(p)
	}
	return m.keyPreview(p, keys)
}

func (m githooksModel) back() (tea.Model, tea.Cmd) {
	if m.phase == ghWriting || m.step <= 1 {
		return m, nil
	}
	m.asking = false
	m.step--
	m.phase = ghEdit
	if m.step == 3 {
		var cmd tea.Cmd
		m, cmd = m.focusCommands()
		return m, cmd
	}
	return m, nil
}

func (m githooksModel) focusCommands() (githooksModel, tea.Cmd) {
	if m.onPush {
		m.commit.Blur()
		return m, m.push.Focus()
	}
	m.push.Blur()
	return m, m.commit.Focus()
}

func (m githooksModel) keyRepo(p tea.KeyPressMsg, keys components.Keys) (tea.Model, tea.Cmd) {
	if p.String() == "tab" {
		m.picking = !m.picking
		if m.picking {
			m.path.Blur()
			return m, m.picker.Init()
		}
		return m, m.path.Focus()
	}
	if m.picking {
		var cmd tea.Cmd
		m.picker, cmd = m.picker.Update(p)
		if chosen := m.picker.Path; chosen != "" {
			m.repo = chosen
			m.path.SetValue(chosen)
			m.repoErr = ""
		}
		return m, cmd
	}
	if key.Matches(p, keys.Enter) {
		return m.acceptRepo()
	}
	m.repoErr = ""
	var cmd tea.Cmd
	m.path, cmd = m.path.Update(p)
	return m, cmd
}

func (m githooksModel) acceptRepo() (tea.Model, tea.Cmd) {
	raw := strings.TrimSpace(m.path.Value())
	if !filepath.IsAbs(raw) {
		m.repoErr = "Use caminho absoluto. Ex.: /Users/voce/projeto"
		return m, nil
	}
	info, err := os.Stat(raw)
	if err != nil || !info.IsDir() {
		m.repoErr = "Pasta não encontrada: confira o caminho."
		return m, nil
	}
	m.repo = filepath.Clean(raw)
	m.repoErr = ""
	m.step = 2
	return m, nil
}

func (m githooksModel) keyHooks(p tea.KeyPressMsg, keys components.Keys) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(p, keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(p, keys.Down):
		if m.cursor < len(m.hooks)-1 {
			m.cursor++
		}
	case p.String() == "space":
		m.hooks = slices.Clone(m.hooks)
		m.hooks[m.cursor].marked = !m.hooks[m.cursor].marked
		m.hookErr = ""
	case key.Matches(p, keys.Enter):
		if !m.anyMarked() {
			m.hookErr = "Marque ao menos um hook com espaço."
			return m, nil
		}
		m.step = 3
		return m.focusCommands()
	}
	return m, nil
}

func (m githooksModel) anyMarked() bool {
	for _, h := range m.hooks {
		if h.marked {
			return true
		}
	}
	return false
}

func (m githooksModel) marked(name string) bool {
	for _, h := range m.hooks {
		if h.name == name && h.marked {
			return true
		}
	}
	return false
}

func (m githooksModel) markedNames() []string {
	var names []string
	for _, h := range m.hooks {
		if h.marked {
			names = append(names, h.name)
		}
	}
	return names
}

func (m githooksModel) keyCommands(p tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch p.String() {
	case "tab":
		m.onPush = !m.onPush
		return m.focusCommands()
	case "ctrl+s":
		return m.acceptCommands()
	}
	var cmd tea.Cmd
	if m.onPush {
		m.push, cmd = m.push.Update(p)
	} else {
		m.commit, cmd = m.commit.Update(p)
	}
	m.cmdErr = ""
	return m, cmd
}

func (m githooksModel) acceptCommands() (tea.Model, tea.Cmd) {
	commit := githooksLines(m.commit.Value())
	push := githooksLines(m.push.Value())
	if msg := githooksCheck("pre-commit", commit, m.marked(ghPreCommit)); msg != "" {
		m.cmdErr = msg
		return m, nil
	}
	if msg := githooksCheck("pre-push", push, m.marked(ghPrePush)); msg != "" {
		m.cmdErr = msg
		return m, nil
	}
	exe, _ := os.Executable()
	m.request = app.GitHooksRequest{
		Home:       m.c.Home,
		Repo:       m.repo,
		Executable: exe,
		Hooks:      m.markedNames(),
		PreCommit:  commit,
		PrePush:    push,
	}
	m.step = 4
	m.phase = ghPlanning
	m.planned = githooksPlanned{}
	m.asking = false
	return m, m.planCmd(m.request)
}

func (m githooksModel) planCmd(req app.GitHooksRequest) tea.Cmd {
	plan := m.deps.Plan
	return func() tea.Msg {
		p, err := plan(req)
		return githooksPlanned{plan: p, err: err}
	}
}

func githooksLines(s string) []string {
	var lines []string
	for _, line := range strings.Split(s, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func githooksCheck(label string, lines []string, required bool) string {
	if len(lines) > ghMaxCommands {
		return fmt.Sprintf("%s: use no máximo %d comandos.", label, ghMaxCommands)
	}
	for _, line := range lines {
		if len(line) > ghMaxCommandLen {
			return fmt.Sprintf("%s: comando maior que %d caracteres.", label, ghMaxCommandLen)
		}
	}
	if required && len(lines) == 0 {
		return fmt.Sprintf("%s: escreva ao menos um comando, porque o hook está marcado.", label)
	}
	return ""
}

func (m githooksModel) keyPreview(p tea.KeyPressMsg, keys components.Keys) (tea.Model, tea.Cmd) {
	if m.phase != ghPreview || m.planned.err != nil {
		return m, nil
	}
	if key.Matches(p, keys.Enter) {
		m.asking = true
		return m, nil
	}
	var cmd tea.Cmd
	m.preview, cmd = m.preview.Update(p)
	return m, cmd
}

func (m githooksModel) startWrite() (tea.Model, tea.Cmd) {
	if m.phase != ghPreview || m.planned.err != nil {
		return m, nil
	}
	m.phase = ghWriting
	m.writes.Add(1)
	var start tea.Cmd
	m.progress, start = m.progress.Start()
	plan := m.planned.plan
	apply := m.deps.Apply
	return m, tea.Batch(start, func() tea.Msg {
		res, err := apply(plan)
		return githooksWritten{result: res, err: err}
	})
}

func (m githooksModel) finishWrite(msg githooksWritten) (tea.Model, tea.Cmd) {
	m.written = msg
	m.phase = ghDone
	state := theme.StateOK
	if msg.err != nil {
		state = theme.StateError
	}
	var c1, c2 tea.Cmd
	m.progress, c1 = m.progress.Advance(state)
	m.progress, c2 = m.progress.Advance(state)
	if m.signal != nil {
		return m, tea.Batch(c1, c2, tea.Quit)
	}
	return m, tea.Batch(c1, c2)
}

func (m githooksModel) previewText() string {
	p := m.planned.plan
	names := slices.Sorted(maps.Keys(p.Scripts))
	expanded := map[string]bool{}
	var scripts []components.Change
	for _, n := range names {
		expanded[n] = true
		scripts = append(scripts, components.Change{Action: components.ChangeCreate, Path: n, Lines: githooksScriptLines(p.Scripts[n])})
	}
	var changes []components.Change
	var paths []string
	for _, list := range [][]githooksFileChange{githooksFiles(p.HomeChanges), githooksFiles(p.RepoChanges)} {
		for _, c := range list {
			changes = append(changes, components.Change{Action: c.action, Path: c.path})
			paths = append(paths, c.path)
		}
	}
	groups := []components.ChangeGroup{{Title: "Scripts", Changes: scripts}, {Title: "Arquivos", Changes: changes}}
	parts := []string{
		components.DiffView(m.t, m.size.Width, groups, expanded),
		"",
		components.FileTree(m.t, "Arquivos", paths),
		"",
		m.t.Strong.Render("Origem dos comandos"),
	}
	for _, s := range p.Sources {
		parts = append(parts, fmt.Sprintf("%s  ←  %s:%d", s.Command, s.File, s.Line))
	}
	return strings.Join(parts, "\n")
}

type githooksFileChange struct {
	action components.ChangeAction
	path   string
}

func githooksFiles(changes []app.Change) []githooksFileChange {
	out := make([]githooksFileChange, 0, len(changes))
	for _, c := range changes {
		out = append(out, githooksFileChange{action: githooksAction(c.Action), path: c.Path})
	}
	return out
}

func githooksScriptLines(script string) []components.DiffLine {
	var lines []components.DiffLine
	for _, line := range strings.Split(script, "\n") {
		lines = append(lines, components.DiffLine{Kind: components.DiffAdded, Text: line})
	}
	return lines
}

func githooksAction(s string) components.ChangeAction {
	switch s {
	case "criar":
		return components.ChangeCreate
	case "atualizar":
		return components.ChangeUpdate
	case "remover":
		return components.ChangeRemove
	case "restaurar":
		return components.ChangeRestore
	case "bloqueado":
		return components.ChangeBlocked
	}
	return components.ChangeUnchanged
}

func (m githooksModel) View() tea.View {
	content := layout.Render(m.t, m.size, layout.Frame{
		Title:   "Git hooks",
		Context: ghStepNames[m.step],
		Version: m.c.Version,
		Body:    m.body(),
		Footer:  components.KeyHints(m.t, m.size.Width, m.help, m.keys),
		Notice:  m.toast.View(m.t, m.size.Width),
	})
	v := tea.NewView(content)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeNone
	return v
}

func (m githooksModel) body() string {
	switch m.step {
	case 1:
		return m.bodyRepo()
	case 2:
		return m.bodyHooks()
	case 3:
		return m.bodyCommands()
	}
	return m.bodyPreview()
}

func (m githooksModel) bodyRepo() string {
	lines := []string{
		m.t.Strong.Render("Caminho do repositório Git"),
		m.t.Secondary.Render("Digite o caminho absoluto. Tab alterna para a escolha de pastas."),
	}
	if m.picking {
		lines = append(lines, m.picker.View())
	} else {
		lines = append(lines, m.path.View())
	}
	if m.repoErr != "" {
		lines = append(lines, m.t.Error.Render(m.repoErr))
	}
	return strings.Join(lines, "\n")
}

func (m githooksModel) bodyHooks() string {
	items := make([]components.ChecklistItem, len(m.hooks))
	for i, h := range m.hooks {
		state := theme.StatePending
		if h.marked {
			state = theme.StateOK
		}
		label := h.label
		if i == m.cursor {
			label = "> " + label
		}
		items[i] = components.ChecklistItem{Label: label, State: state, Detail: h.description}
	}
	lines := []string{
		m.t.Secondary.Render("Espaço marca ou desmarca; enter avança."),
		components.NewChecklist(items, false).View(m.t, m.size.Width),
	}
	if m.hookErr != "" {
		lines = append(lines, m.t.Error.Render(m.hookErr))
	}
	return strings.Join(lines, "\n")
}

func (m githooksModel) bodyCommands() string {
	lines := []string{
		m.t.Strong.Render("pre-commit · um comando por linha (até 8)"),
		m.commit.View(),
		m.t.Strong.Render("pre-push · um comando por linha (até 8)"),
		m.push.View(),
	}
	if m.cmdErr != "" {
		lines = append(lines, m.t.Error.Render(m.cmdErr))
	}
	return strings.Join(lines, "\n")
}

func (m githooksModel) bodyPreview() string {
	if m.phase == ghPlanning {
		return m.t.Secondary.Render("Montando a prévia...")
	}
	if m.planned.err != nil {
		p := app.Explain(m.planned.err)
		return components.ErrorBox(m.t, m.size.Width, p.Title, p.Cause, p.Action)
	}
	if m.phase == ghWriting {
		return m.progress.View(m.t, m.size.Width) + "\n" + m.t.Secondary.Render("Gravando... não feche o terminal.")
	}
	if m.phase == ghDone {
		return m.bodyDone()
	}
	if m.asking {
		return m.confirm.View(m.t, m.size.Width)
	}
	return m.preview.View()
}

func (m githooksModel) bodyDone() string {
	if m.written.err != nil {
		p := app.Explain(m.written.err)
		return components.ErrorBox(m.t, m.size.Width, p.Title, p.Cause, p.Action)
	}
	return m.t.Success.Render("Hooks gravados. ID do histórico: " + m.written.result.HistoryID)
}
