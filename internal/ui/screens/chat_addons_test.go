package screens

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/Gex-Development/gex-dev-kit/internal/app"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

func chatFakeAddons() AddonDeps {
	return AddonDeps{
		Secrets: app.AddonSecretsNeeded,
		PlanRemoval: func(names, harnesses []string) (app.AddonPlan, error) {
			return app.PlanAddonRemoval(names, harnesses, func(string) (string, error) { return "/bin/fake", nil })
		},
		Plan: func(names, harnesses []string, secrets map[string]string) (app.AddonPlan, error) {
			return app.PlanAddonsWithSecrets(names, harnesses, func(string) (string, error) { return "/bin/fake", nil }, secrets)
		},
		Install: func(plan app.AddonPlan, home string, progress func(app.AddonStepResult)) app.AddonResult {
			return app.InstallAddons(plan, home, func(context.Context, string, []string, []string) (string, error) {
				return "", nil
			}, progress)
		},
	}
}

func TestChatAddonSecretsNeverReachTranscriptOrConfirmation(t *testing.T) {
	m := chatTestModel(false)
	m.deps.Addons = chatFakeAddons()
	m = chatSetupQuestion(t, m.flowAddons(), "Qual addon")
	m = chatSetupChoose(t, m, "brave-search")
	m = chatSetupQuestion(t, m, "O que você quer fazer")
	m = chatSetupChoose(t, m, "both")
	m = chatSetupQuestion(t, m, "Cole a")
	if len([]rune("? "+m.ask.question)) > 80 {
		t.Fatal("pergunta secreta não cabe em 80 colunas")
	}
	transcript := chatTestTranscript(m)
	first := "A Chave Brave Search fica em texto puro em ~/.claude.json e ~/.codex/config.toml; só você os lê."
	second := "Para criar: https://api-dashboard.search.brave.com/app/keys. Deixe vazio e tecle enter para cancelar."
	if i, j := strings.Index(transcript, first), strings.Index(transcript, second); i < 0 || j <= i {
		t.Fatal("instruções completas não aparecem antes do campo")
	}
	if m.input.EchoMode != textinput.EchoPassword {
		t.Fatal("entrada da chave não está mascarada")
	}
	m.input.SetValue("fake-key-123")
	if view := m.View().Content; strings.Contains(view, "fake-key-123") || !strings.Contains(view, "••••") {
		t.Fatal("chave visível ou máscara ausente durante digitação")
	}
	m = chatSetupText(t, m, "fake-key-123")
	m = chatSetupQuestion(t, m, "Rodo estes comandos?")
	if m.ask.options[0].ID != "cancel" || !strings.Contains(chatTestTranscript(m), "BRAVE_API_KEY=••••••••") {
		t.Fatal("confirmação sem cancelamento ou máscara")
	}
	assertChatSecretAbsent(t, m)
	m = chatSetupChoose(t, m, "install")
	m = chatSetupQuestion(t, m, "Qual addon")
	assertChatSecretAbsent(t, m)
	if m.input.EchoMode != textinput.EchoNormal {
		t.Fatal("modo da entrada não restaurado")
	}
}

func assertChatSecretAbsent(t *testing.T, m chatModel) {
	t.Helper()
	for _, text := range []string{chatTestTranscript(m), m.View().Content, fmt.Sprintf("%+v", m.request), fmt.Sprintf("%+v", m.history), fmt.Sprintf("%+v", m.ask), fmt.Sprintf("%+v", m.queue)} {
		if strings.Contains(text, "fake-key-123") {
			t.Fatal("chave exposta no chat")
		}
	}
	for _, entry := range m.entries {
		if strings.Contains(entry.text, "fake-key-123") {
			t.Fatal("chave exposta em entrada")
		}
	}
	for _, ask := range m.history {
		if strings.Contains(ask.question, "fake-key-123") {
			t.Fatal("chave exposta em histórico")
		}
	}
}

func TestChatAddonSecretCancelReturnsToList(t *testing.T) {
	for _, cancel := range []string{"empty", "esc"} {
		t.Run(cancel, func(t *testing.T) {
			m := chatTestModel(false)
			m.deps.Addons = chatFakeAddons()
			m = chatSetupQuestion(t, m.flowAddons(), "Qual addon")
			m = chatSetupChoose(t, m, "brave-search")
			m = chatSetupQuestion(t, m, "O que você quer fazer")
			m = chatSetupChoose(t, m, "claude")
			m = chatSetupQuestion(t, m, "Cole a")
			if cancel == "esc" {
				m.input.SetValue("fake-key-123")
				m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyEscape})
			} else {
				m = chatSetupText(t, m, "")
			}
			m = chatSetupQuestion(t, m, "Qual addon")
			if m.addonSecrets != nil || m.input.EchoMode != textinput.EchoNormal {
				t.Fatal("cancelar reteve chave ou máscara")
			}
			assertChatSecretAbsent(t, m)
		})
	}
}

func TestChatAddonMapboxCodexAndTerraformWithoutSecret(t *testing.T) {
	for _, tc := range []struct{ name, question string }{{"mapbox-devkit", "Cole a"}, {"terraform", "Rodo estes comandos?"}} {
		t.Run(tc.name, func(t *testing.T) {
			m := chatTestModel(false)
			m.deps.Addons = chatFakeAddons()
			m = chatSetupQuestion(t, m.flowAddons(), "Qual addon")
			m = chatSetupChoose(t, m, tc.name)
			m = chatSetupQuestion(t, m, "O que você quer fazer")
			m = chatSetupChoose(t, m, "codex")
			m = chatSetupQuestion(t, m, tc.question)
			if tc.name == "mapbox-devkit" {
				if strings.Contains(chatTestTranscript(m), "~/.claude.json; só você o lê.") || !strings.Contains(chatTestTranscript(m), "~/.codex/config.toml; só você o lê.") {
					t.Fatal("instruções incluem cliente não selecionado")
				}
				m = chatSetupText(t, m, "fake-key-123")
				m = chatSetupQuestion(t, m, "Rodo estes comandos?")
				if !strings.Contains(chatTestTranscript(m), "MAPBOX_ACCESS_TOKEN=••••••••") {
					t.Fatal("prévia Mapbox sem máscara")
				}
			}
			assertChatSecretAbsent(t, m)
		})
	}
}

func TestChatAddonPlanningErrorHidesSecret(t *testing.T) {
	m := chatTestModel(false)
	m.deps.Addons = chatFakeAddons()
	m.deps.Addons.Plan = func(_, _ []string, _ map[string]string) (app.AddonPlan, error) {
		return app.AddonPlan{}, errors.New("erro fake-key-123")
	}
	m = chatSetupQuestion(t, m.flowAddons(), "Qual addon")
	m = chatSetupChoose(t, m, "brave-search")
	m = chatSetupQuestion(t, m, "O que você quer fazer")
	m = chatSetupChoose(t, m, "codex")
	m = chatSetupQuestion(t, m, "Cole a")
	m = chatSetupText(t, m, "fake-key-123")
	m = chatSetupQuestion(t, m, "")
	assertChatSecretAbsent(t, m)
	if !strings.Contains(chatTestTranscript(m), "erro [oculto]") || m.addonSecrets != nil {
		t.Fatal("erro de planejamento expôs chave ou reteve segredo")
	}
}

func TestChatAddonFailureHidesSecretInToolRows(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	m := chatTestModel(false)
	m.common.Home = home
	m.deps.Addons = chatFakeAddons()
	m.deps.Addons.Install = func(plan app.AddonPlan, home string, progress func(app.AddonStepResult)) app.AddonResult {
		return app.InstallAddons(plan, home, func(context.Context, string, []string, []string) (string, error) {
			return "primeira linha\nfalha fake-key-123", errors.New("erro fake-key-123")
		}, progress)
	}
	m = chatSetupQuestion(t, m.flowAddons(), "Qual addon")
	m = chatSetupChoose(t, m, "brave-search")
	m = chatSetupQuestion(t, m, "O que você quer fazer")
	m = chatSetupChoose(t, m, "codex")
	m = chatSetupQuestion(t, m, "Cole a")
	m = chatSetupText(t, m, "fake-key-123")
	m = chatSetupQuestion(t, m, "Rodo estes comandos?")
	m = chatSetupChoose(t, m, "install")
	m = chatSetupQuestion(t, m, "Qual addon")
	if !strings.Contains(chatTestTranscript(m), "Falhou: brave-search (Codex): falha [oculto]") {
		t.Fatalf("faltou última linha redigida: %s", chatTestTranscript(m))
	}
	if strings.Contains(chatTestTranscript(m), "Falhou: brave-search (Codex): primeira linha") {
		t.Fatal("mensagem deve mostrar somente última linha")
	}
	assertChatSecretAbsent(t, m)
	if m.exitCode != 1 {
		t.Fatal("falha de instalação sem código de erro")
	}
	found := false
	for _, entry := range m.entries {
		if entry.isTool {
			tool := fmt.Sprintf("%+v", entry.tool)
			if strings.Contains(tool, "fake-key-123") {
				t.Fatal("bloco de ferramenta contém chave")
			}
			found = found || strings.Contains(tool, "[oculto]")
		}
	}
	if !found {
		t.Fatal("erro da ferramenta não foi ocultado")
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
	if m.ask.options[0].ID != "cancel" || m.ask.options[1].ID != "install" {
		t.Fatalf("confirmação não priorizou cancelar: %+v", m.ask.options)
	}
	if !strings.Contains(chatTestTranscript(m), "release v4.0.1 · commit conferido bad48a0") {
		t.Fatal("card confunde tag e revisão conferida")
	}
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

func TestChatAddonsCodexHookApprovalOnlyForPlugins(t *testing.T) {
	for _, tc := range []struct {
		addon     string
		wantHooks bool
	}{{"context7", false}, {"caveman", true}} {
		m := chatTestModel(false)
		m.deps.Addons = chatFakeAddons()
		m = chatSetupQuestion(t, m.flowAddons(), "Qual addon")
		m = chatSetupChoose(t, m, tc.addon)
		m = chatSetupQuestion(t, m, "O que você quer fazer")
		m = chatSetupChoose(t, m, "codex")
		m = chatSetupQuestion(t, m, "Rodo estes comandos?")
		m = chatSetupChoose(t, m, "install")
		m = chatSetupQuestion(t, m, "Qual addon")
		if got := strings.Contains(chatTestTranscript(m), "abra /hooks"); got != tc.wantHooks {
			t.Fatalf("%s: aprovação de hooks=%v, esperado %v: %s", tc.addon, got, tc.wantHooks, chatTestTranscript(m))
		}
	}
}

func TestChatMCPShowsKindCommandsAndRisksBeforeConfirmation(t *testing.T) {
	m := chatTestModel(false)
	m.deps.Addons = chatFakeAddons()
	m = chatSetupQuestion(t, m.flowAddons(), "Qual addon")
	if addonsKindLabel("mcp") != "MCP" {
		t.Fatal("tipo MCP não exibido")
	}
	m = chatSetupChoose(t, m, "context7")
	m = chatSetupQuestion(t, m, "O que você quer fazer")
	for _, text := range []string{"MIT", "Upstash", "limite anônimo"} {
		if !strings.Contains(chatTestTranscript(m), text) {
			t.Fatalf("card sem informação %q", text)
		}
	}
	if !strings.Contains(chatTestTranscript(m), "MCPs de terceiros: revise fonte e licença.") {
		t.Fatal("card MCP deve nomear MCPs")
	}
	m = chatSetupChoose(t, m, "both")
	m = chatSetupQuestion(t, m, "Rodo estes comandos?")
	transcript := chatTestTranscript(m)
	for _, text := range []string{"claude mcp add --scope user context7", "codex mcp add context7", "Envia perguntas e nomes de bibliotecas"} {
		if !strings.Contains(transcript, text) {
			t.Fatalf("prévia sem %q", text)
		}
	}
}

func TestChatMCPMissingRuntimeShowsManualInstruction(t *testing.T) {
	m := chatTestModel(false)
	m.deps.Addons = chatFakeAddons()
	m.deps.Addons.Plan = func(names, harnesses []string, secrets map[string]string) (app.AddonPlan, error) {
		return app.PlanAddonsWithSecrets(names, harnesses, func(name string) (string, error) {
			if name == "npx" {
				return "", errors.New("not found")
			}
			return "/bin/fake", nil
		}, secrets)
	}
	m = chatSetupQuestion(t, m.flowAddons(), "Qual addon")
	m = chatSetupChoose(t, m, "context7")
	m = chatSetupQuestion(t, m, "O que você quer fazer")
	m = chatSetupChoose(t, m, "both")
	m = chatSetupQuestion(t, m, "Rodo estes comandos?")
	for _, text := range []string{"Não encontrei no PATH: npx", "Instale Node.js (npx) e rode de novo."} {
		if !strings.Contains(chatTestTranscript(m), text) {
			t.Fatalf("prévia não explica requisito %q", text)
		}
	}
}

func TestChatMCPBlockedStepsNameTheRealMissingBinary(t *testing.T) {
	m := chatTestModel(false)
	m.deps.Addons = chatFakeAddons()
	m.deps.Addons.Plan = func(names, harnesses []string, secrets map[string]string) (app.AddonPlan, error) {
		return app.PlanAddonsWithSecrets(names, harnesses, func(name string) (string, error) {
			if name == "npx" {
				return "", errors.New("not found")
			}
			return "/bin/fake", nil
		}, secrets)
	}
	m = chatSetupQuestion(t, m.flowAddons(), "Qual addon")
	m = chatSetupChoose(t, m, "context7")
	m = chatSetupQuestion(t, m, "O que você quer fazer")
	m = chatSetupChoose(t, m, "both")
	m = chatSetupQuestion(t, m, "Rodo estes comandos?")
	m = chatSetupChoose(t, m, "install")
	m = chatSetupQuestion(t, m, "Qual addon")
	var skipped int
	for _, entry := range m.entries {
		if entry.isTool && entry.tool.Title == "Instalar plugins e MCPs" {
			for _, row := range entry.tool.Rows {
				if row.State == theme.StatePending && strings.Contains(row.Detail, "faltou npx") {
					skipped++
				}
			}
		}
	}
	if skipped != 2 {
		t.Fatalf("fake deve pular ambos os passos sem npx: %d", skipped)
	}
	transcript := chatTestTranscript(m)
	if strings.Contains(transcript, "instale claude") || strings.Contains(transcript, "instale codex") {
		t.Fatalf("culpou o cliente em vez do npx ausente: %s", transcript)
	}
	if !strings.Contains(transcript, "Instale Node.js (npx) e rode de novo.") {
		t.Fatalf("o resultado deve explicar que falta o Node: %s", transcript)
	}
}

func TestChatAddonsManualAndMissingReview(t *testing.T) {
	m := chatTestModel(false)
	m.deps.Addons = chatFakeAddons()
	m.deps.Addons.Plan = func(names, harnesses []string, secrets map[string]string) (app.AddonPlan, error) {
		return app.PlanAddonsWithSecrets(names, harnesses, func(string) (string, error) { return "", errors.New("not found") }, secrets)
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
	plan, _ := m.deps.Addons.Plan([]string{"superpowers"}, []string{"claude", "codex"}, nil)
	m = m.installAddons(plan, m.common.Home, func(m chatModel) chatModel { return m.flowIdle() })
	m = chatSetupQuestion(t, m, "")
	var rows []string
	for _, entry := range m.entries {
		if entry.isTool && entry.tool.Title == "Instalar plugins e MCPs" {
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
	if !strings.Contains(chatTestTranscript(m), "Falhou: superpowers (Claude Code): boom") || !strings.Contains(chatTestTranscript(m), "Alguns itens não foram instalados; veja os passos que falharam.") {
		t.Fatal("mensagem de falha ou resumo ausente")
	}
}

func TestChatAddonsSummarizesInstalledKindsSeparately(t *testing.T) {
	m := chatTestModel(false)
	m.deps.Addons = chatFakeAddons()
	plan, err := m.deps.Addons.Plan([]string{"caveman", "context7"}, []string{"claude"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	m = m.installAddons(plan, m.common.Home, func(m chatModel) chatModel { return m.flowIdle() })
	m = chatSetupQuestion(t, m, "")
	transcript := chatTestTranscript(m)
	if !strings.Contains(transcript, "Plugins instalados: caveman.") || !strings.Contains(transcript, "MCPs instalados: context7.") || strings.Contains(transcript, "Plugins instalados: caveman, context7.") {
		t.Fatal(transcript)
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

func TestChatAddonRemovalRunsRemovalCommandsAndStopsClaimingRollbackCannotUndo(t *testing.T) {
	var ran []string
	m := chatTestModel(false)
	m.deps.Addons = chatFakeAddons()
	m.deps.Addons.Install = func(plan app.AddonPlan, home string, progress func(app.AddonStepResult)) app.AddonResult {
		return app.InstallAddons(plan, home, func(_ context.Context, _ string, argv, _ []string) (string, error) {
			ran = append(ran, strings.Join(argv, " "))
			return "", nil
		}, progress)
	}
	m = chatSetupQuestion(t, m.flowAddons(), "Qual addon")
	m = chatSetupChoose(t, m, "brave-search")
	m = chatSetupQuestion(t, m, "O que você quer fazer")
	m = chatSetupChoose(t, m, "remove-both")
	m = chatSetupQuestion(t, m, "Removo brave-search?")
	if len(ran) != 0 || strings.Contains(chatTestTranscript(m), "Cole a") {
		t.Fatalf("removeu sem confirmar ou pediu chave: %v", ran)
	}
	m = chatSetupChoose(t, m, "remove")
	m = chatSetupQuestion(t, m, "Qual addon")
	transcript := chatTestTranscript(m)
	if !slices.Equal(ran, []string{"claude mcp remove --scope user brave-search", "codex mcp remove brave-search"}) {
		t.Fatalf("comandos = %v", ran)
	}
	for _, want := range []string{"MCPs removidos: brave-search.", "Revogue a chave de brave-search"} {
		if !strings.Contains(transcript, want) {
			t.Fatalf("falta %q em %s", want, transcript)
		}
	}
	if strings.Contains(transcript, "O rollback do kit não desfaz") || strings.Contains(transcript, "instalados") {
		t.Fatalf("texto de instalação na remoção: %s", transcript)
	}
}

func TestChatAddonRemovalCancelRunsNothing(t *testing.T) {
	ran := 0
	m := chatTestModel(false)
	m.deps.Addons = chatFakeAddons()
	m.deps.Addons.Install = func(app.AddonPlan, string, func(app.AddonStepResult)) app.AddonResult {
		ran++
		return app.AddonResult{}
	}
	m = chatSetupQuestion(t, m.flowAddons(), "Qual addon")
	m = chatSetupChoose(t, m, "context7")
	m = chatSetupQuestion(t, m, "O que você quer fazer")
	m = chatSetupChoose(t, m, "remove-claude")
	m = chatSetupQuestion(t, m, "Removo context7?")
	m = chatSetupChoose(t, m, "cancel")
	m = chatSetupQuestion(t, m, "Qual addon")
	if ran != 0 {
		t.Fatal("cancelar executou a remoção")
	}
}
