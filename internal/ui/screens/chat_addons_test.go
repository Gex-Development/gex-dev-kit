package screens

import (
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Gex-Development/gex-dev-kit/internal/app"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

func chatFakeAddons() AddonDeps {
	return AddonDeps{
		Plan: func(names, harnesses []string) (app.AddonPlan, error) {
			return app.PlanAddons(names, harnesses, func(string) (string, error) { return "/bin/fake", nil })
		},
		Install: func(plan app.AddonPlan, _ string, progress func(app.AddonStepResult)) app.AddonResult {
			var result app.AddonResult
			result.Manual = plan.Manual
			for _, step := range plan.Steps {
				progress(app.AddonStepResult{Step: step, Status: app.AddonPending})
				result.Steps = append(result.Steps, app.AddonStepResult{Step: step, Status: app.AddonOK})
			}
			return result
		},
	}
}

func TestChatAddonsConfirmAndCancel(t *testing.T) {
	calls := 0
	deps := chatTestDeps()
	deps.Addons = chatFakeAddons()
	install := deps.Addons.Install
	deps.Addons.Install = func(plan app.AddonPlan, home string, progress func(app.AddonStepResult)) app.AddonResult {
		calls++
		if home != "/tmp/home" || len(plan.Steps) != 2 || plan.Steps[0].Harness != "claude" {
			t.Errorf("plano ou HOME errado: %+v, %s", plan, home)
		}
		return install(plan, home, progress)
	}
	m := chatTestModel(false)
	m.common.Home = "/tmp/home"
	m.deps = deps
	m = chatSetupQuestion(t, m.flowAddons(), "Qual addon")
	m = chatSetupChoose(t, m, "caveman")
	m = chatSetupQuestion(t, m, "O que você quer fazer")
	if !strings.Contains(chatTestTranscript(m), "Plugins de terceiros: revise fonte e licença.") {
		t.Fatal("card deve mostrar aviso")
	}
	m = chatSetupChoose(t, m, "claude")
	m = chatSetupQuestion(t, m, "Rodo estes comandos?")
	if calls != 0 || !strings.Contains(chatTestTranscript(m), "claude plugin marketplace add JuliusBrussee/caveman") || !strings.Contains(chatTestTranscript(m), "Atenção:") {
		t.Fatalf("revisão deve vir antes da instalação: %s", chatTestTranscript(m))
	}
	m = chatSetupChoose(t, m, "cancel")
	m = chatSetupQuestion(t, m, "Qual addon")
	if calls != 0 {
		t.Fatal("cancelar não instala")
	}
	m = chatSetupChoose(t, m, "caveman")
	m = chatSetupQuestion(t, m, "O que você quer fazer")
	m = chatSetupChoose(t, m, "claude")
	m = chatSetupQuestion(t, m, "Rodo estes comandos?")
	m = chatSetupChoose(t, m, "install")
	m = chatSetupQuestion(t, m, "Qual addon")
	if calls != 1 || !strings.Contains(chatTestTranscript(m), "Plugins instalados: caveman.") || !strings.Contains(chatTestTranscript(m), "O rollback do kit não desfaz plugins") {
		t.Fatalf("instalação incompleta: %s", chatTestTranscript(m))
	}
}

func TestChatAddonsManualAndMissingReview(t *testing.T) {
	m := chatTestModel(false)
	m.deps.Addons = chatFakeAddons()
	m.deps.Addons.Plan = func(names, harnesses []string) (app.AddonPlan, error) {
		p, err := app.PlanAddons(names, harnesses, func(string) (string, error) { return "", errors.New("not found") })
		return p, err
	}
	m = chatSetupQuestion(t, m.flowAddons(), "Qual addon")
	m = chatSetupChoose(t, m, "superpowers")
	m = chatSetupQuestion(t, m, "O que você quer fazer")
	m = chatSetupChoose(t, m, "both")
	m = chatSetupQuestion(t, m, "Rodo estes comandos?")
	transcript := chatTestTranscript(m)
	if !strings.Contains(transcript, "Passo manual (superpowers, Codex):") || !strings.Contains(transcript, "Não encontrei no PATH: claude.") {
		t.Fatalf("revisão incompleta: %s", transcript)
	}
}

func TestChatAddonsRowsFailureAndManual(t *testing.T) {
	m := chatTestModel(false)
	m.deps.Addons = chatFakeAddons()
	m.deps.Addons.Install = func(plan app.AddonPlan, _ string, progress func(app.AddonStepResult)) app.AddonResult {
		for _, step := range plan.Steps {
			progress(app.AddonStepResult{Step: step, Status: app.AddonPending})
		}
		return app.AddonResult{Steps: []app.AddonStepResult{{Step: plan.Steps[0], Status: app.AddonFailed, Output: "first\nboom"}, {Step: app.AddonPlanStep{Addon: "ponytail", Harness: "claude", Argv: []string{"claude"}}, Status: app.AddonSkipped}}, Manual: plan.Manual}
	}
	plan, _ := m.deps.Addons.Plan([]string{"superpowers"}, []string{"claude", "codex"})
	m = m.installAddons(plan, m.common.Home, func(m chatModel) chatModel { return m.flowIdle() })
	m = chatSetupQuestion(t, m, "")
	var rows []string
	for _, entry := range m.entries {
		if entry.isTool && entry.tool.Title == "Instalar plugins" {
			if entry.tool.State != theme.StateError {
				t.Fatal("falha deve marcar ferramenta")
			}
			for _, row := range entry.tool.Rows {
				rows = append(rows, row.Detail)
			}
		}
	}
	if m.exitCode != 1 || !strings.Contains(strings.Join(rows, "|"), "falhou · boom") || !strings.Contains(strings.Join(rows, "|"), "ignorado") || !strings.Contains(strings.Join(rows, "|"), "manual") {
		t.Fatalf("status e linhas: exit=%d rows=%v", m.exitCode, rows)
	}
}

func TestChatAddonInstallDefersInterrupt(t *testing.T) {
	m := chatTestModel(false)
	m.deps.Addons = chatFakeAddons()
	m = m.installAddons(app.AddonPlan{}, m.common.Home, func(m chatModel) chatModel { return m.flowIdle() })
	m, _ = m.pump()
	if !m.Busy() {
		t.Fatal("ferramenta de plugin deve ser gravação")
	}
	model, _ := m.Update(InterruptDeferredMsg{})
	m = model.(chatModel)
	if !strings.Contains(chatTestTranscript(m), "Ctrl+C adiado") {
		t.Fatal("interrupção não adiada")
	}
}

func TestAddonsGuardWaitsForInstaller(t *testing.T) {
	started, release, waited := make(chan struct{}), make(chan struct{}), make(chan struct{})
	var group sync.WaitGroup
	guarded := addonsGuard(AddonDeps{Install: func(app.AddonPlan, string, func(app.AddonStepResult)) app.AddonResult {
		close(started)
		<-release
		return app.AddonResult{}
	}}, &group)
	go guarded.Install(app.AddonPlan{}, t.TempDir(), nil)
	<-started
	go func() { group.Wait(); close(waited) }()
	select {
	case <-waited:
		t.Fatal("espera terminou antes de instalação")
	case <-time.After(10 * time.Millisecond):
	}
	close(release)
	select {
	case <-waited:
	case <-time.After(time.Second):
		t.Fatal("espera não terminou após instalação")
	}
}
