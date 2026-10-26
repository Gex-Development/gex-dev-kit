package screens

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/Gex-Development/gex-dev-kit/internal/app"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/components"
)

func chatSetupModel(deps ChatDeps) chatModel {
	m := chatTestModel(false)
	m.deps = deps
	m.request = deps.Preset
	if m.request.Home == "" {
		m.request.Home = m.common.Home
	}
	if deps.Install.Options != nil {
		m.groups = deps.Install.Options()
	}
	m = m.flowSetup()
	return m
}

func chatSetupQuestion(t *testing.T, m chatModel, question string) chatModel {
	t.Helper()
	for range 100 {
		if question == "" && m.ask == nil && m.running == nil && len(m.queue) == 0 {
			return m
		}
		if m.ask != nil && strings.Contains(m.ask.question, question) {
			return m
		}
		if m.running != nil {
			select {
			case msg := <-m.channel:
				m, _ = chatTestUpdate(m, msg)
			case <-time.After(2 * time.Second):
				t.Fatalf("ferramenta não terminou antes de %q", question)
			}
			continue
		}
		if m.streaming() {
			t.Fatalf("teste sem animação ficou em stream antes de %q", question)
		}
		if len(m.queue) == 0 {
			t.Fatalf("pergunta %q não apareceu; ask=%v, transcrição=%s", question, m.ask, chatTestTranscript(m))
		}
		m, _ = m.pump()
	}
	t.Fatalf("fila não chegou à pergunta %q", question)
	return m
}

func chatSetupChoose(t *testing.T, m chatModel, ids ...string) chatModel {
	t.Helper()
	if m.ask == nil || !m.ask.hasChoice {
		t.Fatalf("resposta %v sem escolha ativa", ids)
	}
	m, _ = chatTestUpdate(m, components.ChoiceDoneMsg{IDs: ids})
	return m
}

func chatSetupText(t *testing.T, m chatModel, text string) chatModel {
	t.Helper()
	if m.ask == nil || !m.ask.input {
		t.Fatalf("texto %q sem pergunta de texto ativa", text)
	}
	m.input.SetValue(text)
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyEnter})
	return m
}

func chatSetupPlan() app.Plan {
	return app.Plan{Files: []app.Change{{Path: ".claude/CLAUDE.md", Action: "criar"}}, Diffs: []app.FileDiff{{Path: ".claude/CLAUDE.md", Action: "criar", Summary: "novo"}}}
}

func chatInstallFakeDeps() ChatDeps {
	d := chatTestDeps()
	d.Install.Options = func() []app.OptionGroup { return nil }
	d.Install.Plan = func(req app.Request) (app.Plan, error) { p := chatSetupPlan(); p.Request = req; return p, nil }
	d.Install.Apply = func(_ app.Plan, _ func(app.Step)) (app.Result, error) { return app.Result{HistoryID: "hist-1"}, nil }
	return d
}

func TestChatSetupOMPDoesNotAskForAgentUpdate(t *testing.T) {
	m := chatSetupQuestion(t, chatSetupModel(chatInstallFakeDeps()), "Qual harness")
	m = chatSetupChoose(t, m, "omp")
	m = chatSetupQuestion(t, m, "Onde eu instalo?")
	if !m.request.SkipAgentUpdate || m.step != 2 || strings.Contains(chatTestTranscript(m), "Quer que eu atualize") {
		t.Fatalf("só OMP deve concluir etapa Agentes sem pergunta: %+v", m.request)
	}
}

func TestChatSetupClaudeCanKeepInstalledAgentVersion(t *testing.T) {
	d := chatInstallFakeDeps()
	d.Agents = func() ([]app.AgentChange, error) {
		return []app.AgentChange{{Name: "claude", Current: "2.0.0", Target: "2.1.0"}}, nil
	}
	m := chatSetupModel(d)
	m.agents = []app.AgentChange{{Name: "claude", Current: "2.0.0", Target: "2.1.0"}}
	m = chatSetupQuestion(t, m, "Qual harness")
	m = chatSetupChoose(t, m, "claude")
	m = chatSetupQuestion(t, m, "Quer que eu atualize os agentes")
	m = chatSetupChoose(t, m, "no")
	m = chatSetupQuestion(t, m, "Onde eu instalo?")
	if !m.request.SkipAgentUpdate {
		t.Fatal("Agora não deveria preservar versão atual do Claude")
	}
}

func TestChatSetupIsolatedHomeRejectsRelativePathAndAcceptsAbsolute(t *testing.T) {
	home := t.TempDir()
	m := chatSetupQuestion(t, chatSetupModel(chatInstallFakeDeps()), "Qual harness")
	m = chatSetupChoose(t, m, "omp")
	m = chatSetupQuestion(t, m, "Onde eu instalo?")
	m = chatSetupChoose(t, m, "isolated")
	m = chatSetupQuestion(t, m, "Caminho absoluto")
	m = chatSetupText(t, m, "x")
	m = chatSetupQuestion(t, m, "Caminho absoluto")
	if !strings.Contains(chatTestTranscript(m), "Não dá para usar isso:") || m.request.Home == "x" {
		t.Fatal("HOME relativo deveria emitir erro e repetir mesma pergunta")
	}
	m = chatSetupText(t, m, filepath.Join(home, "nao-existe"))
	m = chatSetupQuestion(t, m, "Caminho absoluto")
	if !strings.Contains(chatTestTranscript(m), "a pasta não existe") {
		t.Fatalf("pasta inexistente deveria ser recusada na pergunta: %s", chatTestTranscript(m))
	}
	m = chatSetupText(t, m, home+string(filepath.Separator))
	m = chatSetupQuestion(t, m, "Além do Core")
	m = chatSetupChoose(t, m)
	m = chatSetupQuestion(t, m, "Aplico essas alterações?")
	if m.request.Home != home {
		t.Fatalf("HOME isolado deve chegar normalizado, sem barra final: %q", m.request.Home)
	}
}

func TestChatSetupPlanErrorOffersChangingChoices(t *testing.T) {
	d := chatInstallFakeDeps()
	d.Install.Plan = func(app.Request) (app.Plan, error) { return app.Plan{}, errors.New("falha de planejamento") }
	m := chatSetupQuestion(t, chatSetupModel(d), "Qual harness")
	m = chatSetupChoose(t, m, "omp")
	m = chatSetupQuestion(t, m, "Onde eu instalo?")
	m = chatSetupChoose(t, m, "current")
	m = chatSetupQuestion(t, m, "Além do Core")
	m = chatSetupChoose(t, m)
	m = chatSetupQuestion(t, m, "O que eu faço?")
	m = chatSetupChoose(t, m, "change")
	m = chatSetupQuestion(t, m, "Qual harness")
}

func TestChatSetupGroupsSeparateSkillsFromHooks(t *testing.T) {
	d := chatInstallFakeDeps()
	d.Install.Options = func() []app.OptionGroup {
		return []app.OptionGroup{
			{ID: "skills", Title: "Skills", Options: []app.Option{{Value: "skill-a", Label: "A"}, {Value: "skill-b", Label: "B"}}},
			{ID: "hooks", Title: "Hooks", Options: []app.Option{{Value: "hook-a", Label: "H"}}},
		}
	}
	m := chatSetupQuestion(t, chatSetupModel(d), "Qual harness")
	m = chatSetupChoose(t, m, "omp")
	m = chatSetupQuestion(t, m, "Onde eu instalo?")
	m = chatSetupChoose(t, m, "current")
	m = chatSetupQuestion(t, m, "Além do Core")
	m = chatSetupChoose(t, m, "skills", "hooks")
	m = chatSetupQuestion(t, m, "Skills")
	m = chatSetupChoose(t, m, "skill-a", "skill-b")
	m = chatSetupQuestion(t, m, "Hooks")
	m = chatSetupChoose(t, m, "hook-a")
	m = chatSetupQuestion(t, m, "Aplico essas alterações?")
	if strings.Join(m.request.Skills, ",") != "skill-a,skill-b" || strings.Join(m.request.Hooks, ",") != "hook-a" {
		t.Fatalf("seleções devem seguir ordem das opções e grupos corretos: %+v", m.request)
	}
}

func TestChatSetupDoesNotWriteBeforeConfirmationAndWritesOnce(t *testing.T) {
	calls := 0
	d := chatInstallFakeDeps()
	d.Install.Apply = func(_ app.Plan, progress func(app.Step)) (app.Result, error) {
		calls++
		progress(app.StepFiles)
		return app.Result{HistoryID: "hist-1"}, nil
	}
	m := chatSetupQuestion(t, chatSetupModel(d), "Qual harness")
	m = chatSetupChoose(t, m, "omp")
	m = chatSetupQuestion(t, m, "Onde eu instalo?")
	m = chatSetupChoose(t, m, "current")
	m = chatSetupQuestion(t, m, "Além do Core")
	m = chatSetupChoose(t, m)
	m = chatSetupQuestion(t, m, "Aplico essas alterações?")
	if calls != 0 {
		t.Fatalf("Apply chamado antes de confirmação: %d", calls)
	}
	m = chatSetupChoose(t, m, "apply")
	m = chatSetupQuestion(t, m, "")
	if calls != 1 || !strings.Contains(chatTestTranscript(m), "Pronto! Instalei") || strings.Contains(chatTestTranscript(m), "Comando equivalente") {
		t.Fatalf("Apply=%d, resultado incompleto: %s", calls, chatTestTranscript(m))
	}
}

func TestChatSetupRefreshesCoreAfterApply(t *testing.T) {
	installed := false
	d := chatInstallFakeDeps()
	d.Core = func() (app.CoreStatus, error) {
		if installed {
			return app.CoreStatus{State: "instalado", ActiveCore: 7}, nil
		}
		return app.CoreStatus{State: "não instalado"}, nil
	}
	d.Install.Apply = func(app.Plan, func(app.Step)) (app.Result, error) {
		installed = true
		return app.Result{HistoryID: "hist-1"}, nil
	}
	m := chatSetupQuestion(t, chatSetupModel(d), "Qual harness")
	m = chatSetupChoose(t, m, "omp")
	m = chatSetupQuestion(t, m, "Onde eu instalo?")
	m = chatSetupChoose(t, m, "current")
	m = chatSetupQuestion(t, m, "Além do Core")
	m = chatSetupChoose(t, m)
	m = chatSetupQuestion(t, m, "Aplico essas alterações?")
	m = chatSetupChoose(t, m, "apply")
	m = chatSetupQuestion(t, m, "")
	if m.core.ActiveCore != 7 || m.core.State != "instalado" {
		t.Fatalf("Core não foi relido depois da gravação: %+v", m.core)
	}
}

func TestChatSetupPlanErrorOffersRetryAndReplans(t *testing.T) {
	calls := 0
	d := chatInstallFakeDeps()
	d.Install.Plan = func(req app.Request) (app.Plan, error) {
		calls++
		if calls == 1 {
			return app.Plan{}, errors.New("falha de planejamento")
		}
		p := chatSetupPlan()
		p.Request = req
		return p, nil
	}
	m := chatSetupQuestion(t, chatSetupModel(d), "Qual harness")
	m = chatSetupChoose(t, m, "omp")
	m = chatSetupQuestion(t, m, "Onde eu instalo?")
	m = chatSetupChoose(t, m, "current")
	m = chatSetupQuestion(t, m, "Além do Core")
	m = chatSetupChoose(t, m)
	m = chatSetupQuestion(t, m, "O que eu faço?")
	if calls != 1 || !strings.Contains(ansi.Strip(m.View().Content), "Tentar de novo") {
		t.Fatal("falha do plano deveria oferecer tentativa de novo")
	}
	m = chatSetupChoose(t, m, "retry")
	m = chatSetupQuestion(t, m, "Aplico essas alterações?")
	if calls != 2 {
		t.Fatalf("retry deveria executar Plan outra vez, chamadas=%d", calls)
	}
}

func TestChatSetupBlockedAgentSetsExitCodeAndExplainsResult(t *testing.T) {
	d := chatInstallFakeDeps()
	d.Install.Apply = func(_ app.Plan, _ func(app.Step)) (app.Result, error) {
		return app.Result{HistoryID: "hist-1", Blocked: true, HooksNeedTrust: true}, nil
	}
	m := chatSetupQuestion(t, chatSetupModel(d), "Qual harness")
	m = chatSetupChoose(t, m, "omp")
	m = chatSetupQuestion(t, m, "Onde eu instalo?")
	m = chatSetupChoose(t, m, "current")
	m = chatSetupQuestion(t, m, "Além do Core")
	m = chatSetupChoose(t, m)
	m = chatSetupQuestion(t, m, "Aplico essas alterações?")
	m = chatSetupChoose(t, m, "apply")
	m = chatSetupQuestion(t, m, "")
	text := chatTestTranscript(m)
	if m.exitCode != 1 || !strings.Contains(text, "Agentes bloqueados não foram atualizados.") || !strings.Contains(text, "No Codex, abra /hooks") {
		t.Fatalf("bloqueio e confiança dos hooks não reportados: exit=%d, transcrição=%s", m.exitCode, text)
	}
}

func TestChatSetupGroupValidationRepeatsQuestion(t *testing.T) {
	d := chatInstallFakeDeps()
	d.Install.Options = func() []app.OptionGroup {
		return []app.OptionGroup{{ID: "skills", Title: "Skills", Options: []app.Option{{Value: "skill-a", Label: "A"}}, Validate: func(ids []string) error {
			if len(ids) == 0 {
				return errors.New("selecione uma skill")
			}
			return nil
		}}}
	}
	m := chatSetupQuestion(t, chatSetupModel(d), "Qual harness")
	m = chatSetupChoose(t, m, "omp")
	m = chatSetupQuestion(t, m, "Onde eu instalo?")
	m = chatSetupChoose(t, m, "current")
	m = chatSetupQuestion(t, m, "Além do Core")
	m = chatSetupChoose(t, m, "skills")
	m = chatSetupQuestion(t, m, "Skills")
	m = chatSetupChoose(t, m)
	m = chatSetupQuestion(t, m, "Skills")
	if !strings.Contains(chatTestTranscript(m), "Não dá para seguir: selecione uma skill") {
		t.Fatal("grupo inválido deve explicar erro e repetir escolha")
	}
	m = chatSetupChoose(t, m, "skill-a")
	m = chatSetupQuestion(t, m, "Aplico essas alterações?")
	if len(m.request.Skills) != 1 || m.request.Skills[0] != "skill-a" {
		t.Fatalf("seleção válida não entrou no plano: %+v", m.request)
	}
}

func TestChatSetupDiffBackAndCancelNeverApply(t *testing.T) {
	calls := 0
	d := chatInstallFakeDeps()
	d.Install.Apply = func(_ app.Plan, _ func(app.Step)) (app.Result, error) {
		calls++
		return app.Result{}, nil
	}
	m := chatSetupQuestion(t, chatSetupModel(d), "Qual harness")
	m = chatSetupChoose(t, m, "omp")
	m = chatSetupQuestion(t, m, "Onde eu instalo?")
	m = chatSetupChoose(t, m, "current")
	m = chatSetupQuestion(t, m, "Além do Core")
	m = chatSetupChoose(t, m)
	m = chatSetupQuestion(t, m, "Aplico essas alterações?")
	m = chatSetupChoose(t, m, "diff")
	m = chatSetupQuestion(t, m, "Aplico essas alterações?")
	foundExpanded := false
	for _, e := range m.entries {
		if e.isTool && e.tool.Title == "Planejar instalação" && e.tool.Expanded {
			foundExpanded = true
		}
	}
	if !foundExpanded || calls != 0 {
		t.Fatal("Ver diff completo deve expandir plano sem gravar")
	}
	m = chatSetupChoose(t, m, "back")
	m = chatSetupQuestion(t, m, "Qual harness")
	if !strings.Contains(strings.Join(m.ask.choice.Selected(), ","), "omp") {
		t.Fatal("Mudar escolhas deveria preservar pré-seleção de harness")
	}
	m = chatSetupChoose(t, m, "omp")
	m = chatSetupQuestion(t, m, "Onde eu instalo?")
	m = chatSetupChoose(t, m, "current")
	m = chatSetupQuestion(t, m, "Além do Core")
	m = chatSetupChoose(t, m)
	m = chatSetupQuestion(t, m, "Aplico essas alterações?")
	m = chatSetupChoose(t, m, "cancel")
	m = chatSetupQuestion(t, m, "")
	if calls != 0 || !strings.Contains(chatTestTranscript(m), "Nada foi gravado.") || len(m.steps) != 0 || m.context != "pronto" {
		t.Fatalf("cancelar deveria voltar menu sem gravação: calls=%d, contexto=%s", calls, m.context)
	}
}

func TestChatSetupExtrasSkipAndSteps(t *testing.T) {
	d := chatInstallFakeDeps()
	d.Install.Options = func() []app.OptionGroup {
		return []app.OptionGroup{{ID: "design", Title: "Produto e Interface", Description: "Identidade visual", Options: []app.Option{{Value: "gex-design-system", Label: "Design System"}}}}
	}
	m := chatSetupQuestion(t, chatSetupModel(d), "Qual harness")
	m = chatSetupChoose(t, m, "omp")
	m = chatSetupQuestion(t, m, "Onde eu instalo?")
	m = chatSetupChoose(t, m, "current")
	m = chatSetupQuestion(t, m, "Além do Core")
	if strings.Join(m.steps, ",") != "Harness,Agentes,Destino,Extras,Plugins,Plano,Instalação" || m.step != 3 || m.context != "setup · Extras" {
		t.Fatalf("etapas incorretas: %v %d %s", m.steps, m.step, m.context)
	}
	m = chatTestSize(m, 130, 34)
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyEnter})
	m = chatSetupQuestion(t, m, "Aplico essas alterações?")
	if len(m.request.Skills) != 0 || len(m.request.Hooks) != 0 || strings.Contains(chatTestTranscript(m), "Produto e Interface —") || !strings.Contains(chatTestTranscript(m), "Só o Core") {
		t.Fatalf("pular extras deve planejar só Core: %+v", m.request)
	}
	if !strings.Contains(chatTestTranscript(m), "Resumo da instalação:\n  • Harness: OMP") || !strings.Contains(m.View().Content, "ETAPAS") {
		t.Fatal("resumo e etapas ausentes")
	}
}

func TestChatSetupExtrasBackDropsUnselectedGroup(t *testing.T) {
	d := chatInstallFakeDeps()
	d.Preset.Home = t.TempDir()
	d.Install.Options = func() []app.OptionGroup {
		return []app.OptionGroup{{ID: "design", Title: "Produto e Interface", Options: []app.Option{{Value: "gex-design-system", Label: "Design System"}}}}
	}
	m := chatSetupQuestion(t, chatSetupModel(d), "Qual harness")
	m = chatSetupChoose(t, m, "omp")
	m = chatSetupQuestion(t, m, "Onde eu instalo?")
	m = chatSetupChoose(t, m, "current")
	m = chatSetupQuestion(t, m, "Além do Core")
	m = chatSetupChoose(t, m, "design")
	m = chatSetupQuestion(t, m, "Produto e Interface —")
	if m.context != "setup · Extras · Produto e Interface" {
		t.Fatalf("contexto do grupo: %s", m.context)
	}
	m = chatSetupChoose(t, m, "gex-design-system")
	m = chatSetupQuestion(t, m, "Aplico essas alterações?")
	if !strings.Contains(chatTestTranscript(m), "Core GEX + 1 skill(s)") {
		t.Fatal("resumo sem seleção de skill")
	}
	for range 2 {
		m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyEscape})
	}
	if m.ask == nil || !strings.Contains(m.ask.question, "Além do Core") || strings.Join(m.ask.choice.Selected(), ",") != "design" {
		t.Fatal("esc deve restaurar marcas do funil")
	}
	m = chatSetupChoose(t, m)
	m = chatSetupQuestion(t, m, "Aplico essas alterações?")
	if len(m.request.Skills) != 0 {
		t.Fatal("grupo desmarcado deve descartar skills")
	}
}

func TestChatSetupHarnessBannersWithoutCheckTools(t *testing.T) {
	d := chatInstallFakeDeps()
	m := chatSetupModel(d)
	m.agents = []app.AgentChange{{Name: "claude", Current: "2.0.0", Target: "2.1.0", Blocked: true, Reason: "sem permissão"}, {Name: "codex", Current: "1.0.0", Target: "1.0.0"}}
	m = chatSetupQuestion(t, m, "Qual harness")
	m = chatSetupChoose(t, m, "claude", "codex")
	m = chatSetupQuestion(t, m, "Onde eu instalo?")
	var banners int
	for _, entry := range m.entries {
		if entry.banner != "" {
			banners++
			if len(entry.info) != 3 || !strings.Contains(entry.info[1], "versão mais recente:") {
				t.Fatalf("banner incompleto: %v", entry.info)
			}
		}
		if entry.isTool && strings.HasPrefix(entry.tool.Title, "Verificar ") {
			t.Fatalf("verificação redundante: %s", entry.tool.Title)
		}
	}
	text := chatTestTranscript(m)
	if banners != 2 || strings.Count(text, "Beleza, entendi") != 1 || !strings.Contains(text, "A atualização de Claude Code está bloqueada; sigo só com os arquivos do Core. Motivo: sem permissão.") {
		t.Fatalf("banners ou aviso incorretos: %d %s", banners, text)
	}
}

func TestChatSetupBannerShowsUnconsultedLatest(t *testing.T) {
	m := chatSetupModel(chatInstallFakeDeps())
	m.agents = []app.AgentChange{{Name: "claude", Blocked: true, Reason: "sem rede"}}
	m = chatSetupQuestion(t, m, "Qual harness")
	m = chatSetupChoose(t, m, "claude")
	for _, entry := range m.entries {
		if entry.banner == "claude" {
			if entry.info[1] != "versão mais recente: não consultada" {
				t.Fatalf("versão não consultada: %v", entry.info)
			}
			return
		}
	}
	t.Fatal("banner Claude ausente")
}
