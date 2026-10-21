package screens

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/app"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/exp/teatest/v2"
)

func installTestCommon() Common {
	return Common{Version: "teste", Home: "/home/demo", UserHome: "/home/demo", Dark: true, Profile: colorprofile.ASCII}
}

func installTestRequest() app.Request {
	return app.Request{
		Home:            "/home/demo",
		Executable:      "/home/demo/.local/share/gex-dev-kit/bin/gex-dev-kit",
		Skills:          []string{"gex-x"},
		SkipAgentUpdate: true,
	}
}

func installTestPlan() app.Plan {
	return app.Plan{
		Request: installTestRequest(),
		Files: []app.Change{
			{Path: ".claude/CLAUDE.md", Action: "criar"},
			{Path: ".claude/skills/gex-x/SKILL.md", Action: "criar"},
		},
		Diffs: []app.FileDiff{
			{Path: ".claude/CLAUDE.md", Action: "criar", Summary: "novo, 2 linhas", Lines: []app.DiffLine{{Kind: '+', Text: "# Regras"}, {Kind: '+', Text: "linha"}}},
			{Path: ".claude/skills/gex-x/SKILL.md", Action: "criar", Summary: "novo, 1 linha"},
		},
	}
}

func installTestDeps(applied *atomic.Int32, release <-chan struct{}) InstallDeps {
	return InstallDeps{
		Plan: func(app.Request) (app.Plan, error) { return installTestPlan(), nil },
		Apply: func(p app.Plan, progress func(app.Step)) (app.Result, error) {
			applied.Add(1)
			progress(app.StepAgents)
			if release != nil {
				<-release
			}
			progress(app.StepFiles)
			progress(app.StepDone)
			return app.Result{HistoryID: "hist-1"}, nil
		},
	}
}

func TestInstallPreviewScrollsAndNeverAppliesOnCancel(t *testing.T) {
	var applied atomic.Int32
	var model tea.Model = newInstallModel(installTestCommon(), installTestDeps(&applied, nil), installTestRequest(), nil, nil)
	model, _ = model.Update(tea.WindowSizeMsg{Width: 40, Height: 14})
	changes := make([]app.Change, 30)
	for i := range changes {
		changes[i] = app.Change{Path: fmt.Sprintf(".claude/skills/exemplo-%02d/SKILL.md", i), Action: "criar"}
	}
	model, _ = model.Update(installPlanned{plan: app.Plan{Files: changes}})
	if width, height := lipgloss.Width(model.View().Content), lipgloss.Height(model.View().Content); width > 40 || height > 14 {
		t.Fatalf("prévia excede 40x14: %dx%d", width, height)
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyPgDown})
	scrolled := model.(installModel)
	if scrolled.view.YOffset() == 0 {
		t.Fatal("prévia longa não rolou")
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if model.(installModel).phase != installCancelled {
		t.Fatal("Esc deveria cancelar sem aplicação")
	}
	if applied.Load() != 0 {
		t.Fatal("cancelamento chamou a gravação")
	}
}

func TestInstallPlanningErrorStopsBeforeConfirmation(t *testing.T) {
	var applied atomic.Int32
	var model tea.Model = newInstallModel(installTestCommon(), installTestDeps(&applied, nil), installTestRequest(), nil, nil)
	model, _ = model.Update(installPlanned{err: os.ErrPermission})
	screen := model.(installModel)
	if screen.phase != installFailed || !strings.Contains(screen.View().Content, "permission denied") {
		t.Fatalf("falha da prévia ocultada: %q", screen.View().Content)
	}
	if applied.Load() != 0 {
		t.Fatal("falha de planejamento chamou a gravação")
	}
}

func TestInstallConfirmationRequiresExplicitYesAndSurfacesApplyError(t *testing.T) {
	var model tea.Model = newInstallModel(installTestCommon(), InstallDeps{
		Apply: func(app.Plan, func(app.Step)) (app.Result, error) {
			return app.Result{}, errors.New("gravação recusada")
		},
	}, installTestRequest(), nil, nil)
	model, _ = model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	model, _ = model.Update(installPlanned{plan: installTestPlan()})
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if screen := model.(installModel); screen.phase != installConfirm || !strings.Contains(screen.View().Content, "APLICAR ALTERAÇÕES") {
		t.Fatalf("confirmação não está visível: %q", screen.View().Content)
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if model.(installModel).phase != installConfirm {
		t.Fatal("Enter sozinho não pode aplicar mudanças")
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: 's', Text: "s"})
	for msg := range model.(installModel).ch {
		if applied, ok := msg.(installApplied); ok {
			model, _ = model.Update(applied)
		}
	}
	if screen := model.(installModel); screen.phase != installFailed || !strings.Contains(screen.View().Content, "gravação recusada") {
		t.Fatalf("falha da gravação não ficou visível: %q", screen.View().Content)
	}
}

func TestInstallSPreviewWithoutConfirmDoesNotApply(t *testing.T) {
	var applied atomic.Int32
	var model tea.Model = newInstallModel(installTestCommon(), installTestDeps(&applied, nil), installTestRequest(), nil, nil)
	model, _ = model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	model, _ = model.Update(installPlanned{plan: installTestPlan()})
	model, _ = model.Update(tea.KeyPressMsg{Code: 's', Text: "s"})
	if model.(installModel).phase != installPreview || applied.Load() != 0 {
		t.Fatal("s sem confirmação aberta não pode aplicar")
	}
}

func TestInstallClassTinyNeverDecides(t *testing.T) {
	var applied atomic.Int32
	var model tea.Model = newInstallModel(installTestCommon(), installTestDeps(&applied, nil), installTestRequest(), nil, nil)
	model, _ = model.Update(tea.WindowSizeMsg{Width: 30, Height: 10})
	model, _ = model.Update(installPlanned{plan: installTestPlan()})
	for _, key := range []tea.KeyPressMsg{{Code: tea.KeyEnter}, {Code: 's', Text: "s"}} {
		model, _ = model.Update(key)
	}
	if screen := model.(installModel); screen.phase != installPreview || applied.Load() != 0 {
		t.Fatalf("terminal pequeno decidiu: fase %d, gravações %d", screen.phase, applied.Load())
	}
}

func TestInstallBusyDuringApplyAndCtrlCIgnored(t *testing.T) {
	var applied atomic.Int32
	release := make(chan struct{})
	m := newInstallModel(installTestCommon(), installTestDeps(&applied, release), installTestRequest(), nil, nil)
	m.plan = installTestPlan()
	updated, _ := m.startApply()
	defer close(release)
	if !updated.Busy() {
		t.Fatal("Busy deveria ser verdadeiro durante a gravação")
	}
	next, cmd := updated.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	screen := next.(installModel)
	if screen.phase != installApplying || !screen.toast.Visible() {
		t.Fatal("Ctrl+C durante a gravação deveria mostrar aviso e continuar")
	}
	if cmd != nil {
		if _, quit := cmd().(tea.QuitMsg); quit {
			t.Fatal("Ctrl+C durante a gravação não pode encerrar")
		}
	}
	next, _ = screen.Update(InterruptDeferredMsg{Signal: os.Interrupt})
	if next.(installModel).phase != installApplying {
		t.Fatal("sinal adiado não pode interromper a gravação")
	}
}

func TestInstallCtrlCBeforeApplyQuitsAsInterrupted(t *testing.T) {
	var applied atomic.Int32
	m := newInstallModel(installTestCommon(), installTestDeps(&applied, nil), installTestRequest(), nil, nil)
	next, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	if _, quit := cmd().(tea.QuitMsg); !quit || !next.(installModel).interrupted {
		t.Fatal("Ctrl+C antes da gravação deveria sair marcado como interrompido")
	}
}

func TestInstallMotionOffSchedulesNoFrames(t *testing.T) {
	var applied atomic.Int32
	c := installTestCommon()
	m := newInstallModel(c, installTestDeps(&applied, nil), installTestRequest(), nil, nil)
	if _, cmd := m.finish(app.Result{}); cmd != nil {
		t.Fatal("Motion desligado não pode agendar quadros")
	}
	c.Motion = true
	on := newInstallModel(c, installTestDeps(&applied, nil), installTestRequest(), nil, nil)
	if _, cmd := on.finish(app.Result{}); cmd == nil {
		t.Fatal("Motion ligado deveria agendar quadros da conclusão")
	}
}

func TestInstallGuardWaitsForApplyAfterRunReturns(t *testing.T) {
	started := make(chan struct{})
	var finished atomic.Bool
	deps := InstallDeps{Apply: func(app.Plan, func(app.Step)) (app.Result, error) {
		close(started)
		time.Sleep(300 * time.Millisecond)
		finished.Store(true)
		return app.Result{}, nil
	}}
	guarded, wait := installGuard(deps)
	go func() {
		_, _ = guarded.Apply(app.Plan{}, nil)
	}()
	<-started
	wait()
	if !finished.Load() {
		t.Fatal("espera devolveu antes da gravação terminar")
	}
}

func TestInstallRequestFromSelectionKeepsFinalChoices(t *testing.T) {
	sel := &Selection{Home: "/home/outro", Skills: []string{"gex-x", "gex-y"}, Hooks: []string{"gex-conventional-commits"}, SkipAgentUpdate: false}
	got := installRequestFrom(installTestRequest(), sel)
	if got.Home != "/home/outro" || len(got.Skills) != 2 || len(got.Hooks) != 1 || got.SkipAgentUpdate {
		t.Fatalf("pedido final não reflete a seleção: %+v", got)
	}
	if got.Executable != installTestRequest().Executable {
		t.Fatalf("executável do pedido foi alterado: %q", got.Executable)
	}
	if cmd := app.EquivalentCommand(got); !strings.Contains(cmd, "--hook gex-conventional-commits") {
		t.Fatalf("comando equivalente sem o hook: %q", cmd)
	}
}

func TestInstallWithoutFormStartsAtPlanning(t *testing.T) {
	var applied atomic.Int32
	m := newInstallModel(installTestCommon(), installTestDeps(&applied, nil), installTestRequest(), &Selection{}, nil)
	if m.phase != installPlanning {
		t.Fatalf("sem formulário a fase inicial deveria ser planejamento, veio %d", m.phase)
	}
}

func TestInstallFlowWithTeatest(t *testing.T) {
	var applied atomic.Int32
	release := make(chan struct{})
	tm := teatest.NewTestModel(t, newInstallModel(installTestCommon(), installTestDeps(&applied, release), installTestRequest(), nil, nil),
		teatest.WithInitialTermSize(100, 30))
	var seen bytes.Buffer
	out := io.TeeReader(tm.Output(), &seen)
	installWait(t, out, "arquivo")
	tm.Send(tea.KeyPressMsg{Code: 's', Text: "s"})
	tm.Send(tea.KeyPressMsg{Code: tea.KeyEnter})
	installWait(t, out, "APLICAR ALTERAÇÕES")
	if applied.Load() != 0 {
		t.Fatal("s sem confirmação aberta chamou a gravação")
	}
	tm.Send(tea.KeyPressMsg{Code: 's', Text: "s"})
	installWait(t, out, "Etapa: agentes")
	tm.Send(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	installWait(t, out, "Ctrl+C adiado")
	close(release)
	installWait(t, out, "Comando equivalente")
	tm.Send(tea.KeyPressMsg{Code: tea.KeyEnter})
	final := tm.FinalModel(t, teatest.WithFinalTimeout(5*time.Second)).(installModel)
	rest, err := io.ReadAll(tm.FinalOutput(t, teatest.WithFinalTimeout(5*time.Second)))
	text := append(seen.Bytes(), rest...)
	if err != nil {
		t.Fatalf("ler saída final: %v", err)
	}
	if final.result.HistoryID != "hist-1" || applied.Load() != 1 {
		t.Fatalf("gravação incorreta: histórico %q, chamadas %d", final.result.HistoryID, applied.Load())
	}
	for _, want := range []string{"Etapa: agentes", "Etapa: arquivos do Core", "Etapa: concluída", "gex-dev-kit install"} {
		if !bytes.Contains(text, []byte(want)) {
			t.Fatalf("saída final sem %q", want)
		}
	}
}

func installWait(t *testing.T, out io.Reader, text string) {
	t.Helper()
	teatest.WaitFor(t, out, func(b []byte) bool {
		return bytes.Contains(b, []byte(text))
	}, teatest.WithDuration(3*time.Second))
}
