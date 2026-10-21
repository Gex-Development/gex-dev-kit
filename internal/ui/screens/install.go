package screens

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/app"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/components"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/layout"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/motion"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
)

type InstallDeps struct {
	Options func() []app.OptionGroup
	Plan    func(app.Request) (app.Plan, error)
	Apply   func(app.Plan, func(app.Step)) (app.Result, error)
}

type installPhase uint8

const (
	installForm installPhase = iota
	installPlanning
	installPreview
	installConfirm
	installApplying
	installDone
	installFailed
	installCancelled
)

const (
	installBurstID     = 1
	installBurstCount  = 24
	installBurstRows   = 6
	installToastTime   = 3 * time.Second
	installPlanningMsg = "Lendo catálogo e estado dos arquivos…"
	installSignalMsg   = "Sinal recebido; concluindo a gravação para não deixar arquivos pela metade"
	installCtrlCMsg    = "Gravação em andamento: Ctrl+C adiado até concluir"
)

type installPlanned struct {
	plan app.Plan
	err  error
}

type installStepMsg struct{ step app.Step }

type installApplied struct {
	result app.Result
	err    error
}

type installModel struct {
	c           Common
	d           InstallDeps
	req         app.Request
	sel         *Selection
	form        *huh.Form
	phase       installPhase
	plan        app.Plan
	result      app.Result
	err         error
	ch          chan tea.Msg
	keys        components.Keys
	helpOn      bool
	theme       theme.Theme
	width       int
	height      int
	loading     components.Loading
	progress    components.StepProgress
	confirm     components.Confirm
	toast       components.Toast
	view        viewport.Model
	burst       motion.Burst
	interrupted bool
}

func newInstallModel(c Common, d InstallDeps, req app.Request, sel *Selection, form *huh.Form) installModel {
	t := installTheme(c, c.Dark)
	phase := installPlanning
	if form != nil {
		phase = installForm
	}
	return installModel{
		c:        c,
		d:        d,
		req:      req,
		sel:      sel,
		form:     form,
		phase:    phase,
		keys:     components.DefaultKeys(),
		theme:    t,
		width:    layout.MinWidth,
		height:   layout.MinHeight,
		loading:  components.NewLoading(c.Motion),
		progress: components.NewStepProgress(t, []string{"Agentes", "Arquivos do Core", "Concluído"}, c.Motion),
		confirm:  components.NewConfirm("APLICAR ALTERAÇÕES?", "Nada é gravado antes da sua confirmação.", "aplicar", "voltar"),
		toast:    components.NewToast(),
		view:     viewport.New(viewport.WithWidth(72), viewport.WithHeight(12)),
	}
}

func installTheme(c Common, dark bool) theme.Theme {
	return theme.New(theme.Options{Dark: dark, Profile: c.Profile, Plain: c.Profile == colorprofile.ASCII})
}

func installGuard(d InstallDeps) (InstallDeps, func()) {
	var wg sync.WaitGroup
	apply := d.Apply
	if apply != nil {
		d.Apply = func(p app.Plan, progress func(app.Step)) (app.Result, error) {
			wg.Add(1)
			defer wg.Done()
			return apply(p, progress)
		}
	}
	return d, wg.Wait
}

func RunInstall(r Runner, c Common, d InstallDeps, preset app.Request, wizard bool) (app.Result, error) {
	deps, wait := installGuard(d)
	sel := Selection{Home: preset.Home, Skills: preset.Skills, Hooks: preset.Hooks, SkipAgentUpdate: preset.SkipAgentUpdate}
	var form *huh.Form
	if wizard {
		var groups []app.OptionGroup
		if d.Options != nil {
			groups = d.Options()
		}
		form = NewInstallForm(installTheme(c, c.Dark), &sel, groups)
	}
	final, err := r.Run(newInstallModel(c, deps, preset, &sel, form))
	wait()
	if err != nil {
		return app.Result{}, err
	}
	model, ok := final.(installModel)
	if !ok {
		return app.Result{}, errors.New("estado final da instalação inválido")
	}
	if model.interrupted {
		return model.result, SignalError{Signal: os.Interrupt}
	}
	return model.result, model.err
}

func (m installModel) Busy() bool {
	return m.phase == installApplying
}

func (m installModel) Init() tea.Cmd {
	if m.phase == installForm {
		return m.form.Init()
	}
	return tea.Batch(m.planCmd(), m.loading.Tick(), tea.RequestBackgroundColor)
}

func (m installModel) planCmd() tea.Cmd {
	req, plan := m.req, m.d.Plan
	return func() tea.Msg {
		p, err := plan(req)
		return installPlanned{plan: p, err: err}
	}
}

func (m installModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.phase == installForm {
		return m.updateForm(msg)
	}
	var cmds []tea.Cmd
	var cmd tea.Cmd
	m.toast, cmd = m.toast.Update(msg)
	cmds = append(cmds, cmd)
	m.loading, cmd = m.loading.Update(msg)
	cmds = append(cmds, cmd)
	m.progress, cmd = m.progress.Update(msg)
	cmds = append(cmds, cmd)
	m, cmd = m.handle(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m installModel) updateForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	next, cmd := m.form.Update(msg)
	if f, ok := next.(*huh.Form); ok {
		m.form = f
	}
	switch m.form.State {
	case huh.StateCompleted:
		m.req = installRequestFrom(m.req, m.sel)
		m.phase = installPlanning
		return m, tea.Batch(cmd, m.planCmd(), m.loading.Tick())
	case huh.StateAborted:
		m.phase = installCancelled
		return m, tea.Batch(cmd, tea.Quit)
	}
	return m, cmd
}

func installRequestFrom(req app.Request, sel *Selection) app.Request {
	if sel == nil {
		return req
	}
	if sel.Home != "" {
		req.Home = sel.Home
	}
	req.Skills = sel.Skills
	req.Hooks = sel.Hooks
	req.SkipAgentUpdate = sel.SkipAgentUpdate
	return req
}

func (m installModel) handle(msg tea.Msg) (installModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = max(1, msg.Width), max(1, msg.Height)
		m.setPreviewSize()
	case tea.BackgroundColorMsg:
		m.theme = installTheme(m.c, msg.IsDark())
		m.progress = m.progress.WithTheme(m.theme)
		m.setPreviewSize()
	case installPlanned:
		return m.planned(msg)
	case installStepMsg:
		return m.step(msg.step)
	case installApplied:
		return m.applied(msg)
	case motion.FrameMsg:
		m.burst = m.burst.Step()
		if m.burst.Alive() {
			return m, motion.Frame(installBurstID)
		}
	case InterruptDeferredMsg:
		return m.interrupt()
	case tea.KeyPressMsg:
		return m.key(msg)
	}
	return m, nil
}

func (m installModel) planned(msg installPlanned) (installModel, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
		m.phase = installFailed
		return m, nil
	}
	m.plan = msg.plan
	if m.plan.Empty() && !m.plan.Blocked {
		return m.finish(app.Result{})
	}
	m.phase = installPreview
	m.setPreviewSize()
	return m, nil
}

func (m installModel) step(s app.Step) (installModel, tea.Cmd) {
	var cmd tea.Cmd
	m.progress, cmd = m.progress.Advance(theme.StateOK)
	return m, tea.Batch(tea.Println(installStepLine(s)), cmd, installListen(m.ch))
}

func (m installModel) applied(msg installApplied) (installModel, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
		m.phase = installFailed
		return m, nil
	}
	return m.finish(msg.result)
}

func (m installModel) finish(r app.Result) (installModel, tea.Cmd) {
	m.result = r
	m.phase = installDone
	m.burst = motion.NewBurst(m.c.Motion, float64(m.width)/2, 2, installBurstCount, time.Now().UnixNano())
	if m.burst.Alive() {
		return m, motion.Frame(installBurstID)
	}
	return m, nil
}

func (m installModel) interrupt() (installModel, tea.Cmd) {
	if m.phase == installApplying {
		return m.notify(installSignalMsg)
	}
	m.interrupted = true
	m.phase = installCancelled
	return m, tea.Quit
}

func (m installModel) notify(text string) (installModel, tea.Cmd) {
	var cmd tea.Cmd
	m.toast, cmd = m.toast.Show(text, theme.StateWarn, installToastTime)
	return m, cmd
}

func (m installModel) key(msg tea.KeyPressMsg) (installModel, tea.Cmd) {
	key := msg.String()
	if m.phase == installApplying {
		if key == "ctrl+c" {
			return m.notify(installCtrlCMsg)
		}
		return m, nil
	}
	if key == "ctrl+c" {
		if m.phase == installPlanning || m.phase == installPreview || m.phase == installConfirm {
			m.interrupted = true
			m.phase = installCancelled
		}
		return m, tea.Quit
	}
	if layout.Classify(layout.Size{Width: m.width, Height: m.height}) == layout.ClassTiny && key != "q" {
		return m, nil
	}
	switch m.phase {
	case installPlanning:
		if key == "esc" || key == "q" {
			m.phase = installCancelled
			return m, tea.Quit
		}
	case installPreview:
		return m.keyPreview(msg, key)
	case installConfirm:
		return m.keyConfirm(msg)
	case installDone:
		if key == "enter" || key == "q" {
			return m, tea.Sequence(tea.Println(m.summary()), tea.Quit)
		}
	case installFailed:
		if key == "enter" || key == "q" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m installModel) keyPreview(msg tea.KeyPressMsg, key string) (installModel, tea.Cmd) {
	switch key {
	case "esc", "q":
		m.phase = installCancelled
		return m, tea.Quit
	case "enter":
		m.phase = installConfirm
		return m, nil
	case "?":
		m.helpOn = !m.helpOn
		return m, nil
	}
	var cmd tea.Cmd
	m.view, cmd = m.view.Update(msg)
	return m, cmd
}

func (m installModel) keyConfirm(msg tea.KeyPressMsg) (installModel, tea.Cmd) {
	var decision components.Decision
	m.confirm, decision = m.confirm.Update(msg, m.confirm.Visible(m.width, m.height))
	switch decision {
	case components.DecisionYes:
		return m.startApply()
	case components.DecisionNo:
		m.phase = installPreview
	}
	return m, nil
}

func (m installModel) startApply() (installModel, tea.Cmd) {
	m.phase = installApplying
	ch := make(chan tea.Msg, 8)
	m.ch = ch
	plan, apply := m.plan, m.d.Apply
	go func() {
		result, err := apply(plan, func(s app.Step) {
			ch <- installStepMsg{step: s}
		})
		ch <- installApplied{result: result, err: err}
		close(ch)
	}()
	var cmd tea.Cmd
	m.progress, cmd = m.progress.Start()
	return m, tea.Batch(installListen(ch), cmd)
}

func installListen(ch chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return msg
	}
}

func installStepLine(s app.Step) string {
	switch s {
	case app.StepAgents:
		return "Etapa: agentes"
	case app.StepFiles:
		return "Etapa: arquivos do Core"
	}
	return "Etapa: concluída"
}

func (m installModel) summary() string {
	lines := installSummaryLines(m.result)
	lines = append(lines, "Comando equivalente: "+app.EquivalentCommand(m.req))
	return strings.Join(lines, "\n")
}

func installSummaryLines(r app.Result) []string {
	var lines []string
	if r.AgentsUpdated {
		lines = append(lines, "Agentes atualizados.")
	}
	if r.HistoryID != "" {
		lines = append(lines, "Core instalado. Histórico: "+r.HistoryID)
	}
	if !r.AgentsUpdated && r.HistoryID == "" {
		lines = append(lines, "Nada a alterar.")
	}
	if r.HooksNeedTrust {
		lines = append(lines, "Codex: aprove o hook novo em /hooks para ativá-lo.")
	}
	if r.Blocked {
		lines = append(lines, "Agentes bloqueados não foram atualizados.")
	}
	return lines
}

func (m installModel) View() tea.View {
	if m.phase == installForm {
		return tea.NewView(m.form.View())
	}
	return tea.NewView(m.render())
}

func (m installModel) render() string {
	size := layout.Size{Width: m.width, Height: m.height}
	base := layout.Render(m.theme, size, layout.Frame{
		Title:   "GEX / DEV KIT",
		Context: "Instalação",
		Version: m.c.Version,
		Body:    m.body(),
		Footer:  components.KeyHints(m.theme, m.width, m.helpOn, m.keys),
		Notice:  m.toast.View(m.theme, m.width),
	})
	if m.phase == installConfirm {
		return m.confirm.Overlay(m.theme, base, m.width, m.height)
	}
	return base
}

func (m installModel) body() string {
	switch m.phase {
	case installPlanning:
		return m.loading.View(m.theme, m.width, 3) + "\n\n" + installPlanningMsg + "\n\nNenhuma alteração aplicada."
	case installPreview, installConfirm:
		return m.previewBody()
	case installApplying:
		return m.progress.View(m.theme, m.width) + "\n\nNão feche o terminal durante a operação."
	case installDone:
		return m.particles() + "\n" + m.summary() + "\n\nEnter ou q para sair"
	case installFailed:
		p := app.Explain(m.err)
		return components.ErrorBox(m.theme, m.width, p.Title, p.Cause, p.Action) + "\n\nEnter ou q para sair"
	case installCancelled:
		return "Instalação cancelada; nenhum arquivo aplicado."
	}
	return ""
}

func (m installModel) previewBody() string {
	head := fmt.Sprintf("%d arquivo(s) · %d agente(s) a atualizar", len(m.plan.Files), len(m.plan.Updates))
	if m.plan.Blocked {
		head += "\nAgentes bloqueados exigem ação manual; o Core pode prosseguir."
	}
	return m.theme.Strong.Render(head) + "\n" + m.view.View()
}

func (m installModel) particles() string {
	if !m.burst.Alive() {
		return ""
	}
	rows := make([][]rune, installBurstRows)
	for i := range rows {
		rows[i] = []rune(strings.Repeat(" ", max(1, m.width)))
	}
	for _, p := range m.burst.Particles(max(1, m.width), installBurstRows) {
		rows[p.Y][p.X] = '·'
	}
	lines := make([]string, len(rows))
	for i, row := range rows {
		lines[i] = string(row)
	}
	return m.theme.Faint.Render(strings.Join(lines, "\n"))
}

func (m *installModel) setPreviewSize() {
	body := layout.BodySize(layout.Size{Width: m.width, Height: m.height})
	m.view.SetWidth(max(1, body.Width))
	m.view.SetHeight(max(1, body.Height-3))
	m.view.SetContent(ansi.Hardwrap(m.previewText(max(1, body.Width)), max(1, body.Width), false))
}

func (m installModel) previewText(width int) string {
	var parts []string
	if groups := installGroups(m.plan.Diffs); len(groups) > 0 {
		parts = append(parts, components.DiffView(m.theme, width, groups, installExpanded(m.plan.Diffs)))
	}
	if len(m.plan.Files) > 0 {
		paths := make([]string, 0, len(m.plan.Files))
		for _, f := range m.plan.Files {
			paths = append(paths, f.Path)
		}
		parts = append(parts, components.FileTree(m.theme, homeLabel(m.req.Home, m.c.UserHome), paths))
	}
	if agents := m.agentLines(); agents != "" {
		parts = append(parts, agents)
	}
	if notices := m.noticeLines(); notices != "" {
		parts = append(parts, notices)
	}
	return strings.Join(parts, "\n\n")
}

func (m installModel) agentLines() string {
	var lines []string
	for _, a := range m.plan.Agents {
		current := a.Current
		if current == "" {
			current = "não gerenciado"
		}
		switch {
		case a.Blocked:
			lines = append(lines, m.theme.Error.Render(fmt.Sprintf("bloqueado · %s (%s → %s): %s", a.Name, current, a.Target, a.Reason)))
			lines = append(lines, m.theme.Secondary.Render("Ação manual: "+a.Manual))
		case a.Current != a.Target:
			lines = append(lines, m.theme.Info.Render(fmt.Sprintf("atualizar · %s %s → %s", a.Name, current, a.Target)))
		}
	}
	if len(lines) == 0 {
		return ""
	}
	return m.theme.Strong.Render("AGENTES") + "\n" + strings.Join(lines, "\n")
}

func (m installModel) noticeLines() string {
	var lines []string
	for _, n := range m.plan.Notices {
		style := m.theme.Secondary
		if n.Level == "aviso" {
			style = m.theme.Warning
		}
		lines = append(lines, style.Render(n.Text))
	}
	return strings.Join(lines, "\n")
}

func installGroups(diffs []app.FileDiff) []components.ChangeGroup {
	if len(diffs) == 0 {
		return nil
	}
	group := components.ChangeGroup{Title: "Alterações"}
	for _, d := range diffs {
		group.Changes = append(group.Changes, components.Change{
			Action:  installAction(d.Action),
			Path:    d.Path,
			Detail:  d.Summary,
			Lines:   installLines(d.Lines),
			Omitted: d.Omitted,
			Masked:  d.Masked,
		})
	}
	return []components.ChangeGroup{group}
}

func installExpanded(diffs []app.FileDiff) map[string]bool {
	expanded := make(map[string]bool, len(diffs))
	for _, d := range diffs {
		expanded[d.Path] = true
	}
	return expanded
}

func installAction(action string) components.ChangeAction {
	switch action {
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

func installLines(lines []app.DiffLine) []components.DiffLine {
	out := make([]components.DiffLine, 0, len(lines))
	for _, line := range lines {
		kind := components.DiffContext
		switch line.Kind {
		case '+':
			kind = components.DiffAdded
		case '-':
			kind = components.DiffRemoved
		case '@':
			kind = components.DiffHunk
		}
		out = append(out, components.DiffLine{Kind: kind, Text: line.Text})
	}
	return out
}
