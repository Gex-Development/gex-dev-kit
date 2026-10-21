package screens

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/app"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/components"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/layout"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
	"github.com/charmbracelet/colorprofile"
)

type RollbackKind int

const (
	RollbackCore RollbackKind = iota
	RollbackRepo
	RollbackAgents
)

type RollbackDeps struct {
	Plan  func() (app.RollbackPreview, error)
	Apply func() (string, error)
}

type rollbackStage int

const (
	rollbackPlanning rollbackStage = iota
	rollbackPreview
	rollbackEmpty
	rollbackBlocked
	rollbackConfirm
	rollbackApplying
	rollbackDone
	rollbackFailed
)

const (
	rollbackStableBinDir = ".local/share/gex-dev-kit/bin/"
	rollbackStableWarn   = "Os hooks passam a ignorar a verificação até nova instalação."
	rollbackLoadingLines = 3
)

type rollbackPlannedMsg struct {
	preview app.RollbackPreview
	err     error
}

type rollbackAppliedMsg struct {
	id  string
	err error
}

type rollbackKeyMap struct{ keys components.Keys }

func (k rollbackKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.keys.Enter, k.keys.Up, k.keys.Down, k.keys.Back, k.keys.Quit, k.keys.Help}
}

func (k rollbackKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.keys.Up, k.keys.Down},
		{k.keys.Enter, k.keys.Back},
		{k.keys.Quit, k.keys.Help},
	}
}

type rollbackModel struct {
	c           Common
	kind        RollbackKind
	deps        RollbackDeps
	theme       theme.Theme
	profile     colorprofile.Profile
	size        layout.Size
	stage       rollbackStage
	preview     app.RollbackPreview
	stable      bool
	problem     app.Problem
	result      string
	diffs       viewport.Model
	confirm     components.Confirm
	loading     components.Loading
	steps       components.StepProgress
	help        bool
	interrupted bool
	signaled    bool
}

func newRollbackModel(c Common, kind RollbackKind, d RollbackDeps) rollbackModel {
	m := rollbackModel{
		c:       c,
		kind:    kind,
		deps:    d,
		profile: c.Profile,
		size:    layout.Size{Width: 80, Height: 24},
		stage:   rollbackPlanning,
		diffs:   viewport.New(viewport.WithWidth(80), viewport.WithHeight(1)),
	}
	m.theme = doctorTheme(c.Dark, c.Profile)
	m.loading = components.NewLoading(c.Motion)
	m.steps = components.NewStepProgress(m.theme, []string{"Restaurar arquivos"}, c.Motion)
	m.confirm = components.NewConfirm(rollbackTitle(kind), rollbackConfirmBody(kind, false), "confirmar", "cancelar")
	if kind == RollbackAgents {
		m.stage = rollbackConfirm
	}
	return m
}

func rollbackTitle(kind RollbackKind) string {
	switch kind {
	case RollbackRepo:
		return "Restaurar git hooks do repositório"
	case RollbackAgents:
		return "Reverter agentes"
	}
	return "Restaurar instalação do Core"
}

func rollbackConfirmBody(kind RollbackKind, stable bool) string {
	if kind == RollbackAgents {
		return "Claude e Codex voltam à versão anterior ao kit."
	}
	if stable {
		return rollbackStableWarn
	}
	return "Os arquivos listados voltam ao estado anterior à instalação."
}

func (m rollbackModel) Init() tea.Cmd {
	cmds := []tea.Cmd{tea.RequestBackgroundColor, m.loading.Tick()}
	if m.stage == rollbackPlanning {
		cmds = append(cmds, m.planCmd())
	}
	return tea.Batch(cmds...)
}

func (m rollbackModel) planCmd() tea.Cmd {
	plan := m.deps.Plan
	return func() tea.Msg {
		preview, err := plan()
		return rollbackPlannedMsg{preview: preview, err: err}
	}
}

func (m rollbackModel) applyCmd() tea.Cmd {
	apply := m.deps.Apply
	return func() tea.Msg {
		id, err := apply()
		return rollbackAppliedMsg{id: id, err: err}
	}
}

func (m rollbackModel) Busy() bool {
	return m.stage == rollbackApplying
}

func (m rollbackModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		m.theme = doctorTheme(msg.IsDark(), m.profile)
		m.steps = m.steps.WithTheme(m.theme)
	case tea.ColorProfileMsg:
		m.profile = msg.Profile
		m.theme = doctorTheme(m.theme.Dark, m.profile)
		m.steps = m.steps.WithTheme(m.theme)
	case tea.WindowSizeMsg:
		m.size = layout.Size{Width: max(1, msg.Width), Height: max(1, msg.Height)}
		m.refreshDiff()
	case rollbackPlannedMsg:
		m = m.planned(msg)
	case rollbackAppliedMsg:
		return m.applied(msg)
	case InterruptDeferredMsg:
		return m.interrupt()
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	default:
		return m.tick(msg)
	}
	return m, nil
}

func (m rollbackModel) tick(msg tea.Msg) (tea.Model, tea.Cmd) {
	var loadCmd, stepCmd tea.Cmd
	m.loading, loadCmd = m.loading.Update(msg)
	m.steps, stepCmd = m.steps.Update(msg)
	return m, tea.Batch(loadCmd, stepCmd)
}

func (m rollbackModel) planned(msg rollbackPlannedMsg) rollbackModel {
	switch {
	case msg.err != nil:
		m.stage = rollbackBlocked
		m.problem = app.Explain(msg.err)
	case len(msg.preview.Plan.Conflicts) > 0:
		m.stage = rollbackBlocked
		m.problem = rollbackConflictProblem(msg.preview.Plan.Conflicts)
	case len(msg.preview.Plan.Entries) == 0:
		m.stage = rollbackEmpty
	default:
		m.preview = msg.preview
		m.stable = rollbackRemovesStable(msg.preview)
		m.stage = rollbackPreview
		m.confirm = components.NewConfirm(rollbackTitle(m.kind), rollbackConfirmBody(m.kind, m.stable), "confirmar", "cancelar")
		m.refreshDiff()
	}
	return m
}

func (m rollbackModel) applied(msg rollbackAppliedMsg) (tea.Model, tea.Cmd) {
	var advance tea.Cmd
	if msg.err != nil {
		m.stage = rollbackFailed
		m.problem = app.Explain(msg.err)
		m.steps, advance = m.steps.Advance(theme.StateError)
	} else {
		m.stage = rollbackDone
		m.result = msg.id
		m.steps, advance = m.steps.Advance(theme.StateOK)
	}
	if m.signaled {
		return m, tea.Sequence(advance, tea.Quit)
	}
	return m, advance
}

func (m rollbackModel) interrupt() (tea.Model, tea.Cmd) {
	if m.stage != rollbackApplying {
		m.interrupted = true
		return m, tea.Quit
	}
	m.signaled = true
	return m, nil
}

func (m rollbackModel) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	keys := components.DefaultKeys()
	if key.Matches(msg, keys.Quit) {
		if m.stage == rollbackApplying {
			return m, nil
		}
		m.interrupted = msg.String() == "ctrl+c"
		return m, tea.Quit
	}
	if layout.Classify(m.size) == layout.ClassTiny || m.stage == rollbackApplying {
		return m, nil
	}
	if key.Matches(msg, keys.Help) {
		m.help = !m.help
		return m, nil
	}
	if m.stage == rollbackConfirm {
		return m.handleConfirm(msg)
	}
	return m.handleStage(msg, keys)
}

func (m rollbackModel) handleConfirm(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	var decision components.Decision
	m.confirm, decision = m.confirm.Update(msg, m.confirm.Visible(m.size.Width, m.size.Height))
	switch decision {
	case components.DecisionYes:
		return m.startApply()
	case components.DecisionNo:
		if m.kind == RollbackAgents {
			return m, tea.Quit
		}
		m.stage = rollbackPreview
	}
	return m, nil
}

func (m rollbackModel) startApply() (tea.Model, tea.Cmd) {
	m.stage = rollbackApplying
	var start tea.Cmd
	m.steps, start = m.steps.Start()
	return m, tea.Batch(start, m.applyCmd())
}

func (m rollbackModel) handleStage(msg tea.KeyPressMsg, keys components.Keys) (tea.Model, tea.Cmd) {
	switch m.stage {
	case rollbackPreview:
		if key.Matches(msg, keys.Enter) {
			m.stage = rollbackConfirm
			return m, nil
		}
		if key.Matches(msg, keys.Back) {
			return m, tea.Quit
		}
		var cmd tea.Cmd
		m.diffs, cmd = m.diffs.Update(msg)
		return m, cmd
	case rollbackPlanning:
		if key.Matches(msg, keys.Back) {
			return m, tea.Quit
		}
		return m, nil
	}
	if key.Matches(msg, keys.Enter, keys.Back) {
		return m, tea.Quit
	}
	return m, nil
}

func (m *rollbackModel) refreshDiff() {
	body := layout.BodySize(m.size)
	m.diffs.SetWidth(max(1, body.Width))
	m.diffs.SetHeight(max(1, body.Height-rollbackHeaderRows(m.stable)))
	m.diffs.SetContent(components.DiffView(m.theme, body.Width, rollbackGroups(m.preview), rollbackExpanded(m.preview)))
}

func rollbackHeaderRows(stable bool) int {
	if stable {
		return 2
	}
	return 1
}

func (m rollbackModel) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	v.MouseMode = tea.MouseModeNone
	return v
}

func (m rollbackModel) render() string {
	frame := layout.Frame{
		Title:   "GEX / ROLLBACK",
		Context: rollbackTitle(m.kind),
		Version: m.c.Version,
		Body:    m.body(),
		Footer:  components.KeyHints(m.theme, m.size.Width, m.help, rollbackKeyMap{keys: components.DefaultKeys()}),
	}
	base := layout.Render(m.theme, m.size, frame)
	if m.stage == rollbackConfirm && layout.Classify(m.size) != layout.ClassTiny {
		return m.confirm.Overlay(m.theme, base, m.size.Width, m.size.Height)
	}
	return base
}

func (m rollbackModel) body() string {
	body := layout.BodySize(m.size)
	switch m.stage {
	case rollbackPlanning:
		return m.theme.Secondary.Render("Conferindo o que a restauração vai alterar.") + "\n" + m.loading.View(m.theme, body.Width, rollbackLoadingLines)
	case rollbackEmpty:
		return components.EmptyState(m.theme, body.Width, "Nada a restaurar", "Nenhum arquivo mudou desde a instalação.")
	case rollbackBlocked, rollbackFailed:
		return components.ErrorBox(m.theme, body.Width, m.problem.Title, m.problem.Cause, m.problem.Action)
	case rollbackApplying:
		return m.theme.Secondary.Render("Gravando a restauração. Não feche o terminal até concluir.") + "\n" + m.steps.View(m.theme, body.Width)
	case rollbackDone:
		return m.doneBody(body.Width)
	}
	return m.previewBody()
}

func (m rollbackModel) previewBody() string {
	lines := []string{m.theme.Strong.Render(fmt.Sprintf("%d alterações previstas", len(m.preview.Plan.Entries)))}
	if m.stable {
		lines = append(lines, components.StatusBadge(m.theme, theme.StateWarn, rollbackStableWarn))
	}
	lines = append(lines, m.diffs.View())
	return strings.Join(lines, "\n")
}

func (m rollbackModel) doneBody(width int) string {
	if m.result == "" {
		if m.kind == RollbackAgents {
			return components.StatusBadge(m.theme, theme.StateOK, "Agentes revertidos.")
		}
		return components.EmptyState(m.theme, width, "Nada a restaurar", "Nenhum arquivo foi alterado.")
	}
	return components.StatusBadge(m.theme, theme.StateOK, "Restauração concluída.") + "\n" +
		m.theme.Secondary.Render("ID do histórico: ") + m.theme.Strong.Render(m.result)
}

func rollbackRemovesStable(preview app.RollbackPreview) bool {
	for _, e := range preview.Plan.Entries {
		if e.Action == "remover" && strings.HasPrefix(e.Path, rollbackStableBinDir) {
			return true
		}
	}
	return false
}

func rollbackGroups(preview app.RollbackPreview) []components.ChangeGroup {
	diffs := make(map[string]app.FileDiff, len(preview.Diffs))
	for _, d := range preview.Diffs {
		diffs[d.Path] = d
	}
	changes := make([]components.Change, 0, len(preview.Plan.Entries))
	for _, e := range preview.Plan.Entries {
		change := components.Change{Action: rollbackChangeAction(e.Action), Path: e.Path}
		if d, ok := diffs[e.Path]; ok {
			change.Detail = d.Summary
			change.Lines = rollbackDiffLines(d.Lines)
			change.Omitted = d.Omitted
			change.Masked = d.Masked
		}
		changes = append(changes, change)
	}
	return []components.ChangeGroup{{Title: "Arquivos", Changes: changes}}
}

func rollbackExpanded(preview app.RollbackPreview) map[string]bool {
	expanded := make(map[string]bool, len(preview.Plan.Entries))
	for _, e := range preview.Plan.Entries {
		expanded[e.Path] = true
	}
	return expanded
}

func rollbackChangeAction(action string) components.ChangeAction {
	switch action {
	case "restaurar":
		return components.ChangeRestore
	case "remover":
		return components.ChangeRemove
	}
	return components.ChangeUnchanged
}

func rollbackDiffLines(lines []app.DiffLine) []components.DiffLine {
	out := make([]components.DiffLine, 0, len(lines))
	for _, line := range lines {
		out = append(out, components.DiffLine{Kind: rollbackDiffKind(line.Kind), Text: line.Text})
	}
	return out
}

func rollbackDiffKind(kind byte) components.DiffLineKind {
	switch kind {
	case '+':
		return components.DiffAdded
	case '-':
		return components.DiffRemoved
	case '@':
		return components.DiffHunk
	}
	return components.DiffContext
}

func rollbackConflictProblem(paths []string) app.Problem {
	return app.Problem{
		Title:  "Rollback abortado",
		Cause:  "O rollback não alterou nada: estes arquivos mudaram depois da instalação: " + strings.Join(paths, ", ") + ".",
		Action: "Devolva-os à versão instalada (veja `doctor`) ou siga sem rollback; o kit nunca sobrescreve edição posterior.",
	}
}

func rollbackGuard(d RollbackDeps, group *sync.WaitGroup) RollbackDeps {
	apply := d.Apply
	if apply == nil {
		return d
	}
	d.Apply = func() (string, error) {
		group.Add(1)
		defer group.Done()
		return apply()
	}
	return d
}

func RunRollback(r Runner, c Common, kind RollbackKind, d RollbackDeps) (string, error) {
	var group sync.WaitGroup
	final, err := r.Run(newRollbackModel(c, kind, rollbackGuard(d, &group)))
	group.Wait()
	if err != nil {
		return "", err
	}
	m, ok := final.(rollbackModel)
	if !ok {
		return "", errors.New("tela de rollback terminou sem resultado")
	}
	if m.interrupted {
		return "", SignalError{Signal: os.Interrupt}
	}
	if m.stage == rollbackFailed {
		return "", errors.New(m.problem.Cause)
	}
	return m.result, nil
}
