package screens

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/app"
)

func chatSetupPluginQuestion(t *testing.T, deps ChatDeps, harness string) chatModel {
	t.Helper()
	m := chatSetupQuestion(t, chatSetupModel(deps), "Qual harness")
	m = chatSetupChoose(t, m, harness)
	m = chatSetupQuestion(t, m, "Onde eu instalo?")
	m = chatSetupChoose(t, m, "current")
	m = chatSetupQuestion(t, m, "Além do Core")
	m = chatSetupChoose(t, m)
	return chatSetupQuestion(t, m, "Quer instalar algum plugin?")
}

func TestChatSetupBraveSecretsMaskedBeforeApply(t *testing.T) {
	deps := chatInstallFakeDeps()
	deps.Addons = chatFakeAddons()
	m := chatSetupPluginQuestion(t, deps, "claude")
	m = chatSetupChoose(t, m, "brave-search")
	m = chatSetupQuestion(t, m, "Cole a")
	m = chatSetupText(t, m, "fake-key-123")
	m = chatSetupQuestion(t, m, "Aplico essas alterações?")
	if m.ask.options[0].ID != "cancel" || m.addonSecrets != nil || !strings.Contains(chatTestTranscript(m), "BRAVE_API_KEY=••••••••") {
		t.Fatal("setup sem confirmação mascarada ou manteve chave")
	}
	assertChatSecretAbsent(t, m)
	m = chatSetupChoose(t, m, "apply")
	m = chatSetupQuestion(t, m, "")
	if m.addonPlan.Steps != nil {
		t.Fatal("plano com chave persistiu no modelo após instalar")
	}
	assertChatSecretAbsent(t, m)
}

func TestChatSetupEmptySecretCancelsPlugin(t *testing.T) {
	deps := chatInstallFakeDeps()
	deps.Addons = chatFakeAddons()
	m := chatSetupPluginQuestion(t, deps, "codex")
	m = chatSetupChoose(t, m, "mapbox-devkit")
	m = chatSetupQuestion(t, m, "Cole a")
	m = chatSetupText(t, m, "")
	m = chatSetupQuestion(t, m, "Quer instalar algum plugin?")
	if m.addonSecrets != nil || len(m.addonNames) != 0 {
		t.Fatal("chave vazia não cancelou seleção")
	}
}

func TestChatSetupSecretEscClearsInput(t *testing.T) {
	deps := chatInstallFakeDeps()
	deps.Addons = chatFakeAddons()
	m := chatSetupPluginQuestion(t, deps, "codex")
	m = chatSetupChoose(t, m, "mapbox-devkit")
	m = chatSetupQuestion(t, m, "Cole a")
	m.input.SetValue("fake-key-123")
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyEscape})
	m = chatSetupQuestion(t, m, "Quer instalar algum plugin?")
	if m.input.Value() != "" || m.input.EchoMode != textinput.EchoNormal || m.addonSecrets != nil || len(m.addonNames) != 0 {
		t.Fatal("Esc reteve token ou não voltou à lista")
	}
	assertChatSecretAbsent(t, m)
}

func TestChatSetupSecondSecretCancelClearsFirst(t *testing.T) {
	deps := chatInstallFakeDeps()
	deps.Addons = chatFakeAddons()
	m := chatSetupPluginQuestion(t, deps, "claude")
	m = chatSetupChoose(t, m, "brave-search", "mapbox-devkit")
	m = chatSetupQuestion(t, m, "Chave Brave Search")
	m = chatSetupText(t, m, "fake-key-123")
	m = chatSetupQuestion(t, m, "Token Mapbox")
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyEscape})
	m = chatSetupQuestion(t, m, "Quer instalar algum plugin?")
	if m.addonSecrets != nil || m.addonPlan.Steps != nil || len(m.addonNames) != 0 {
		t.Fatal("cancelar segunda chave reteve primeira")
	}
	assertChatSecretAbsent(t, m)
}

func TestChatSetupPluginsReviewAndCoreFirst(t *testing.T) {
	var calls []string
	deps := chatInstallFakeDeps()
	deps.Addons = chatFakeAddons()
	deps.Install.Apply = func(app.Plan, func(app.Step)) (app.Result, error) {
		calls = append(calls, "core")
		return app.Result{HistoryID: "hist-1"}, nil
	}
	install := deps.Addons.Install
	deps.Addons.Install = func(plan app.AddonPlan, home string, progress func(app.AddonStepResult)) app.AddonResult {
		calls = append(calls, "plugins")
		if home != "/home/demo" {
			t.Errorf("HOME errado: %s", home)
		}
		return install(plan, home, progress)
	}
	m := chatSetupPluginQuestion(t, deps, "claude")
	if m.step != 4 || len(m.steps) != 7 {
		t.Fatalf("plugins fora da etapa 5/7: %d %v", m.step, m.steps)
	}
	m = chatSetupChoose(t, m, "caveman")
	m = chatSetupQuestion(t, m, "Aplico essas alterações?")
	if m.ask.options[0].ID != "cancel" || m.ask.options[1].ID != "apply" {
		t.Fatalf("confirmação de terceiros não priorizou cancelar: %+v", m.ask.options)
	}
	if !strings.Contains(chatTestTranscript(m), "Depois do Core vou rodar 2 comando(s) de terceiros (caveman); os comandos exatos estão na revisão da etapa Plugins.") {
		t.Fatal("aviso de execução de terceiros ausente")
	}
	transcript := chatTestTranscript(m)
	for _, text := range []string{"claude plugin marketplace add JuliusBrussee/caveman", "Atenção:", "Muda o estilo de todas as respostas", "• Plugins: caveman (Claude Code)"} {
		if !strings.Contains(transcript, text) {
			t.Fatalf("revisão incompleta (%s): %s", text, transcript)
		}
	}
	if len(calls) != 0 {
		t.Fatalf("instalação prematura: %v", calls)
	}
	m = chatSetupChoose(t, m, "apply")
	m = chatSetupQuestion(t, m, "")
	if strings.Join(calls, ",") != "core,plugins" || !strings.Contains(chatTestTranscript(m), "plugins não têm comando CLI equivalente") {
		t.Fatalf("ordem ou comando incorreto: %v %s", calls, chatTestTranscript(m))
	}
}

func TestChatSetupOnlyPluginsWhenCoreCurrent(t *testing.T) {
	calls := 0
	deps := chatInstallFakeDeps()
	deps.Addons = chatFakeAddons()
	deps.Install.Plan = func(app.Request) (app.Plan, error) { return app.Plan{}, nil }
	deps.Install.Apply = func(app.Plan, func(app.Step)) (app.Result, error) {
		t.Fatal("Core não deve rodar")
		return app.Result{}, nil
	}
	install := deps.Addons.Install
	deps.Addons.Install = func(plan app.AddonPlan, home string, progress func(app.AddonStepResult)) app.AddonResult {
		calls++
		return install(plan, home, progress)
	}
	m := chatSetupPluginQuestion(t, deps, "codex")
	m = chatSetupChoose(t, m, "ponytail")
	m = chatSetupQuestion(t, m, "Aplico essas alterações?")
	if !strings.Contains(chatTestTranscript(m), "O Core já está em dia; só os plugins serão instalados.") || calls != 0 {
		t.Fatal("Core vazio ainda exige confirmação")
	}
	m = chatSetupChoose(t, m, "apply")
	m = chatSetupQuestion(t, m, "")
	if calls != 1 || !strings.Contains(chatTestTranscript(m), "No Codex, abra /hooks e aprove os hooks novos.") {
		t.Fatalf("plugin codex não instalado: calls=%d %s", calls, chatTestTranscript(m))
	}
}

func TestChatSetupPluginsEscKeepsSelection(t *testing.T) {
	deps := chatInstallFakeDeps()
	deps.Addons = chatFakeAddons()
	m := chatSetupPluginQuestion(t, deps, "claude")
	m = chatSetupChoose(t, m, "caveman")
	m = chatSetupQuestion(t, m, "Aplico essas alterações?")
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyEscape})
	if m.ask == nil || m.step != 4 || strings.Join(m.ask.choice.Selected(), ",") != "caveman" {
		t.Fatalf("esc perdeu seleção: step=%d selected=%v", m.step, m.ask)
	}
	m = chatSetupChoose(t, m, "caveman", "ponytail")
	m = chatSetupQuestion(t, m, "Aplico essas alterações?")
	m = chatSetupChoose(t, m, "back")
	m = chatSetupQuestion(t, m, "Qual harness")
	m = chatSetupChoose(t, m, "claude")
	m = chatSetupQuestion(t, m, "Onde eu instalo?")
	m = chatSetupChoose(t, m, "current")
	m = chatSetupQuestion(t, m, "Além do Core")
	m = chatSetupChoose(t, m)
	m = chatSetupQuestion(t, m, "Quer instalar algum plugin?")
	if strings.Join(m.ask.choice.Selected(), ",") != "caveman,ponytail" {
		t.Fatalf("Mudar escolhas perdeu seleção: %v", m.ask.choice.Selected())
	}
}

func TestChatSetupOMPBypassesPlugins(t *testing.T) {
	deps := chatInstallFakeDeps()
	deps.Addons = chatFakeAddons()
	m := chatSetupQuestion(t, chatSetupModel(deps), "Qual harness")
	m = chatSetupChoose(t, m, "omp")
	m = chatSetupQuestion(t, m, "Onde eu instalo?")
	m = chatSetupChoose(t, m, "current")
	m = chatSetupQuestion(t, m, "Além do Core")
	m = chatSetupChoose(t, m)
	m = chatSetupQuestion(t, m, "Aplico essas alterações?")
	if !strings.Contains(chatTestTranscript(m), "como você escolheu só o OMP, pulo esta etapa") || strings.Contains(chatTestTranscript(m), "Quer instalar algum plugin?") {
		t.Fatal("OMP deve pular plugins")
	}
}

func TestChatSetupPluginFailureSetsExitCode(t *testing.T) {
	deps := chatInstallFakeDeps()
	deps.Addons = chatFakeAddons()
	deps.Install.Plan = func(app.Request) (app.Plan, error) { return app.Plan{}, nil }
	deps.Addons.Install = func(plan app.AddonPlan, _ string, progress func(app.AddonStepResult)) app.AddonResult {
		step := plan.Steps[0]
		progress(app.AddonStepResult{Step: step, Status: app.AddonPending})
		return app.AddonResult{Steps: []app.AddonStepResult{{Step: step, Status: app.AddonFailed, Output: "boom"}}}
	}
	m := chatSetupPluginQuestion(t, deps, "claude")
	m = chatSetupChoose(t, m, "caveman")
	m = chatSetupQuestion(t, m, "Aplico essas alterações?")
	m = chatSetupChoose(t, m, "apply")
	m = chatSetupQuestion(t, m, "")
	if m.exitCode != 1 || strings.Contains(chatTestTranscript(m), "No Codex, abra /hooks") {
		t.Fatalf("falha deve sair 1 sem instrução de hooks do Codex: %s", chatTestTranscript(m))
	}
}
