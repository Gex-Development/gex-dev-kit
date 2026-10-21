package screens

import (
	"bytes"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/app"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/layout"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/exp/teatest/v2"
)

const rollbackTestPath = ".claude/CLAUDE.md"

type rollbackTestCalls struct {
	mu      sync.Mutex
	applied int
}

func (c *rollbackTestCalls) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.applied
}

func rollbackTestCommon(t *testing.T) Common {
	return Common{Version: "teste", UserHome: t.TempDir(), Dark: true, Profile: colorprofile.TrueColor}
}

func rollbackTestPreview() app.RollbackPreview {
	preview := app.RollbackPreview{
		Diffs: []app.FileDiff{{
			Path:    rollbackTestPath,
			Action:  "restaurar",
			Summary: "2 → 3 bytes",
			Lines:   []app.DiffLine{{Kind: '-', Text: "antigo"}, {Kind: '+', Text: "novo"}},
		}},
	}
	rollbackTestSetEntries(&preview.Plan, [2]string{rollbackTestPath, "restaurar"})
	return preview
}

func rollbackTestSetEntries(plan *app.RollbackPlan, pairs ...[2]string) {
	field := reflect.ValueOf(plan).Elem().FieldByName("Entries")
	slice := reflect.MakeSlice(field.Type(), 0, len(pairs))
	for _, p := range pairs {
		entry := reflect.New(field.Type().Elem()).Elem()
		entry.FieldByName("Path").SetString(p[0])
		entry.FieldByName("Action").SetString(p[1])
		slice = reflect.Append(slice, entry)
	}
	field.Set(slice)
}

func rollbackTestDeps(preview app.RollbackPreview, planErr error, calls *rollbackTestCalls) RollbackDeps {
	return RollbackDeps{
		Plan: func() (app.RollbackPreview, error) { return preview, planErr },
		Apply: func() (string, error) {
			calls.mu.Lock()
			calls.applied++
			calls.mu.Unlock()
			return "h-1", nil
		},
	}
}

func rollbackTestKey(s string) tea.KeyPressMsg {
	if s == "enter" {
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	}
	return tea.KeyPressMsg{Code: rune(s[0]), Text: s}
}

func rollbackTestWaitFor(t *testing.T, tm *teatest.TestModel, text string) {
	t.Helper()
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte(text))
	}, teatest.WithDuration(3*time.Second))
}

func TestRollbackPreviewConfirmsWithSAndShowsID(t *testing.T) {
	calls := &rollbackTestCalls{}
	tm := teatest.NewTestModel(t, newRollbackModel(rollbackTestCommon(t), RollbackCore, rollbackTestDeps(rollbackTestPreview(), nil, calls)),
		teatest.WithInitialTermSize(100, 30))
	rollbackTestWaitFor(t, tm, rollbackTestPath)
	tm.Send(rollbackTestKey("enter"))
	tm.Send(rollbackTestKey("s"))
	rollbackTestWaitFor(t, tm, "ID do histórico")
	tm.Send(rollbackTestKey("q"))
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
	final := tm.FinalModel(t).(rollbackModel)
	if calls.count() != 1 || final.result != "h-1" {
		t.Fatalf("aplicação: chamadas=%d resultado=%q", calls.count(), final.result)
	}
	if view := final.View().Content; !bytes.Contains([]byte(view), []byte("h-1")) {
		t.Fatalf("resultado não mostra o ID do histórico:\n%s", view)
	}
}

func TestRollbackConflictShowsErrorBoxAndDoesNotApply(t *testing.T) {
	calls := &rollbackTestCalls{}
	preview := rollbackTestPreview()
	preview.Plan.Conflicts = []string{rollbackTestPath}
	tm := teatest.NewTestModel(t, newRollbackModel(rollbackTestCommon(t), RollbackCore, rollbackTestDeps(preview, nil, calls)),
		teatest.WithInitialTermSize(100, 30))
	rollbackTestWaitFor(t, tm, "Rollback abortado")
	tm.Send(rollbackTestKey("s"))
	tm.Send(rollbackTestKey("q"))
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
	if calls.count() != 0 {
		t.Fatalf("conflito não pode aplicar: chamadas=%d", calls.count())
	}
}

func TestRollbackEmptyPreviewQuitsWithoutApply(t *testing.T) {
	calls := &rollbackTestCalls{}
	empty := app.RollbackPreview{}
	tm := teatest.NewTestModel(t, newRollbackModel(rollbackTestCommon(t), RollbackCore, rollbackTestDeps(empty, nil, calls)),
		teatest.WithInitialTermSize(100, 30))
	rollbackTestWaitFor(t, tm, "Nada a restaurar")
	tm.Send(rollbackTestKey("enter"))
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
	if calls.count() != 0 {
		t.Fatalf("prévia vazia não pode aplicar: chamadas=%d", calls.count())
	}
}

func TestRollbackBusyWhileApplying(t *testing.T) {
	calls := &rollbackTestCalls{}
	m := newRollbackModel(rollbackTestCommon(t), RollbackCore, rollbackTestDeps(rollbackTestPreview(), nil, calls))
	m.stage = rollbackConfirm
	m.size = layout.Size{Width: 100, Height: 30}
	if m.Busy() {
		t.Fatal("ocupado antes de gravar")
	}
	next, _ := m.Update(rollbackTestKey("s"))
	if !next.(rollbackModel).Busy() {
		t.Fatal("tela não ficou ocupada durante a gravação")
	}
	ignored, _ := next.Update(rollbackTestKey("q"))
	if ignored.(rollbackModel).interrupted || ignored.(rollbackModel).stage != rollbackApplying {
		t.Fatal("q durante a gravação não pode sair")
	}
}

func TestRollbackAgentsStartsAtConfirmWithoutPlan(t *testing.T) {
	m := newRollbackModel(rollbackTestCommon(t), RollbackAgents, RollbackDeps{Apply: func() (string, error) { return "", nil }})
	if m.stage != rollbackConfirm {
		t.Fatalf("agentes devem começar na confirmação, estágio=%v", m.stage)
	}
	m.size = layout.Size{Width: 100, Height: 30}
	if m.View().Content == "" {
		t.Fatal("confirmação dos agentes sem conteúdo")
	}
}

func TestRollbackTinyTerminalIgnoresDecisions(t *testing.T) {
	calls := &rollbackTestCalls{}
	m := newRollbackModel(rollbackTestCommon(t), RollbackCore, rollbackTestDeps(rollbackTestPreview(), nil, calls))
	m.size = layout.Size{Width: 30, Height: 8}
	m.stage = rollbackPreview
	next, _ := m.Update(rollbackTestKey("enter"))
	if next.(rollbackModel).stage != rollbackPreview {
		t.Fatal("terminal pequeno não pode tomar decisões")
	}
}

func TestRollbackPlanErrorIsExplained(t *testing.T) {
	calls := &rollbackTestCalls{}
	m := newRollbackModel(rollbackTestCommon(t), RollbackCore, rollbackTestDeps(app.RollbackPreview{}, errors.New("falha"), calls))
	next, _ := m.Update(rollbackPlannedMsg{err: errors.New("falha")})
	if next.(rollbackModel).stage != rollbackBlocked {
		t.Fatal("erro de planejamento deve bloquear a confirmação")
	}
}

func TestRollbackStableRemovalWarns(t *testing.T) {
	stable := app.RollbackPreview{}
	rollbackTestSetEntries(&stable.Plan, [2]string{".local/share/gex-dev-kit/bin/gex-dev-kit", "remover"})
	if !rollbackRemovesStable(stable) {
		t.Fatal("remoção do executável estável não virou aviso")
	}
	if rollbackRemovesStable(rollbackTestPreview()) {
		t.Fatal("arquivo comum virou aviso de executável estável")
	}
}

func TestRollbackGuardWaitsForApply(t *testing.T) {
	var group sync.WaitGroup
	guarded := rollbackGuard(RollbackDeps{Apply: func() (string, error) { return "h-9", nil }}, &group)
	id, err := guarded.Apply()
	group.Wait()
	if id != "h-9" || err != nil {
		t.Fatalf("gravação protegida: id=%q err=%v", id, err)
	}
}

func TestRollbackConflictProblemListsPaths(t *testing.T) {
	problem := rollbackConflictProblem([]string{".claude/CLAUDE.md", ".codex/AGENTS.md"})
	if problem.Title != "Rollback abortado" || !bytes.Contains([]byte(problem.Cause), []byte(".codex/AGENTS.md")) {
		t.Fatalf("problema de conflito incompleto: %+v", problem)
	}
}

func TestRollbackEnterOpensConfirmation(t *testing.T) {
	m := newRollbackModel(rollbackTestCommon(t), RollbackCore, rollbackTestDeps(rollbackTestPreview(), nil, &rollbackTestCalls{}))
	m.size = layout.Size{Width: 100, Height: 30}
	m = m.planned(rollbackPlannedMsg{preview: rollbackTestPreview()})
	next, _ := m.Update(rollbackTestKey("enter"))
	if next.(rollbackModel).stage != rollbackConfirm {
		t.Fatal("Enter na prévia deve abrir a confirmação")
	}
}
