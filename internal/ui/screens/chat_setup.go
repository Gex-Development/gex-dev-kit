package screens

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Gex-Development/gex-dev-kit/internal/app"
	"github.com/Gex-Development/gex-dev-kit/internal/catalog"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/components"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

func (m chatModel) setupStep(index int) chatModel {
	m.step = index
	m.context = "setup · " + m.steps[index]
	return m
}

func (m chatModel) flowSetup() chatModel {
	m.history = nil
	m.steps = []string{"Harness", "Agentes", "Destino", "Extras", "Plugins", "Plano", "Instalação"}
	m = m.say("Vamos configurar seu ambiente. São poucos passos: harness, destino, extras, plugins, revisão e instalação. Você pode voltar com esc a qualquer momento.")
	m = m.setupStep(0)
	opts := []components.ChoiceOption{
		{ID: "claude", Label: "Claude Code", Icon: components.HarnessBadge(m.theme, "claude"), Art: func(t theme.Theme) []string { return components.HarnessArt(t, "claude") }},
		{ID: "codex", Label: "Codex", Icon: components.HarnessBadge(m.theme, "codex"), Art: func(t theme.Theme) []string { return components.HarnessArt(t, "codex") }},
		{ID: "omp", Label: "OMP (oh-my-pi)", Icon: components.HarnessBadge(m.theme, "omp"), Art: func(t theme.Theme) []string { return components.HarnessArt(t, "omp") }},
	}
	pre := m.harnessIDs
	if len(pre) == 0 && m.request.SkipAgentUpdate {
		pre = []string{"omp"}
	}
	return m.askFor(chatAsk{hasChoice: true, question: "Qual harness você usa?", choice: components.NewChoice("Qual harness você usa?", opts, true, true).Spaced().Preselect(pre), options: opts, onAnswer: func(m chatModel, a chatAnswer) chatModel { return m.setupHarness(a.IDs) }})
}

func (m chatModel) setupHarness(ids []string) chatModel {
	m.harnessIDs = slices.Clone(ids)
	labels := make([]string, 0, len(ids))
	for _, id := range ids {
		switch id {
		case "claude":
			labels = append(labels, "Claude Code")
		case "codex":
			labels = append(labels, "Codex")
		case "omp":
			labels = append(labels, "OMP (oh-my-pi)")
		}
	}
	if len(labels) == 1 {
		m = m.say("Beleza, entendi que o seu harness é o " + labels[0] + ".")
	} else {
		m = m.say("Beleza, entendi que você usa " + strings.Join(labels[:len(labels)-1], ", ") + " e " + labels[len(labels)-1] + ".")
	}
	needsUpdate := false
	var blocked []string
	var reasons []string
	for _, item := range []struct{ id, label string }{{"claude", "Claude Code"}, {"codex", "Codex"}} {
		if !slices.Contains(ids, item.id) {
			continue
		}
		change, ok := homeAgentChange(m.agents, item.id)
		if !ok {
			change = app.AgentChange{Name: item.id}
		}
		current := agentBannerTitle(item.label, change)
		status := homeLabel(m.common.Home, m.common.UserHome)
		if change.Blocked {
			status = "bloqueado: " + change.Reason
			blocked = append(blocked, item.label)
			reasons = append(reasons, change.Reason)
		} else if ok && change.Current != change.Target {
			needsUpdate = true
		}
		latest := change.Target
		if latest == "" {
			latest = "não consultada"
		}
		m.queue = append(m.queue, chatAction{banner: item.id, info: []string{current, "versão mais recente: " + latest, status}})
	}
	switch len(blocked) {
	case 1:
		m = m.say(fmt.Sprintf("A atualização de %s está bloqueada; sigo só com os arquivos do Core. Motivo: %s.", blocked[0], reasons[0]))
	case 2:
		m = m.say(fmt.Sprintf("A atualização de %s e %s está bloqueada; sigo só com os arquivos do Core. Motivos: %s: %s. %s: %s.", blocked[0], blocked[1], blocked[0], reasons[0], blocked[1], reasons[1]))
	}
	if slices.Contains(ids, "omp") {
		m = m.say("O OMP lê as skills de ~/.agents/skills; o kit já instala lá.")
	}
	m = m.setupStep(1)
	if needsUpdate {
		opts := []components.ChoiceOption{{ID: "yes", Label: "Sim, atualizar"}, {ID: "no", Label: "Agora não"}}
		return m.askFor(chatAsk{hasChoice: true, question: "Quer que eu atualize os agentes para as versões mais recentes?", choice: components.NewChoice("Quer que eu atualize os agentes para as versões mais recentes?", opts, false, true), options: opts, onAnswer: func(m chatModel, a chatAnswer) chatModel {
			m.request.SkipAgentUpdate = a.IDs[0] == "no"
			return m.setupDestination()
		}})
	}
	m.request.SkipAgentUpdate = true
	return m.setupDestination()
}

func (m chatModel) setupDestination() chatModel {
	m = m.setupStep(2)
	opts := []components.ChoiceOption{{ID: "current", Label: "Neste HOME (" + homeLabel(m.common.Home, m.common.UserHome) + ")"}, {ID: "isolated", Label: "Em um HOME isolado (pasta de teste)"}}
	return m.askFor(chatAsk{hasChoice: true, question: "Onde eu instalo?", choice: components.NewChoice("Onde eu instalo?", opts, false, true), options: opts, onAnswer: func(m chatModel, a chatAnswer) chatModel {
		if a.IDs[0] == "isolated" {
			return m.askFor(chatAsk{input: true, path: true, picker: components.NewPathPicker(m.common.UserHome, nil), question: "Caminho absoluto do HOME isolado:", validate: chatValidateHome, onAnswer: func(m chatModel, a chatAnswer) chatModel {
				m.request.Home = strings.TrimSpace(a.Text)
				return m.setupExtras()
			}})
		}
		m.request.Home = m.common.Home
		return m.setupExtras()
	}})
}

func chatValidateHome(value string) error {
	if err := installAbsolutePath(value); err != nil {
		return err
	}
	info, err := os.Stat(value)
	if errors.Is(err, fs.ErrNotExist) {
		return errors.New("a pasta não existe; crie-a antes ou escolha outra")
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return errors.New("o HOME isolado precisa ser uma pasta")
	}
	return nil
}

func installAbsolutePath(value string) error {
	if strings.TrimSpace(value) == "" {
		return errors.New("informe uma pasta para o HOME isolado")
	}
	if !filepath.IsAbs(value) {
		return errors.New("use caminho absoluto, começando na raiz do sistema ou na unidade (Windows)")
	}
	return nil
}

func (m chatModel) setupExtras() chatModel {
	m = m.setupStep(3)
	opts := make([]components.ChoiceOption, 0, len(m.groups))
	pre := slices.Clone(m.selectedGroups)
	if pre == nil {
		for _, g := range m.groups {
			items := m.request.Skills
			if g.ID == "hooks" {
				items = m.request.Hooks
			}
			for _, o := range g.Options {
				if slices.Contains(items, o.Value) {
					pre = append(pre, g.ID)
					break
				}
			}
		}
	}
	for _, g := range m.groups {
		word := "opções"
		if len(g.Options) == 1 {
			word = "opção"
		}
		opts = append(opts, components.ChoiceOption{ID: g.ID, Label: fmt.Sprintf("%s (%d %s)", g.Title, len(g.Options), word), Hint: g.Description})
	}
	question := "Além do Core (obrigatório), o que mais você quer instalar?"
	return m.askFor(chatAsk{hasChoice: true, question: question, choice: components.NewChoice(question, opts, true, false).Preselect(pre), options: opts, onAnswer: func(m chatModel, a chatAnswer) chatModel {
		m.selectedGroups = slices.Clone(a.IDs)
		for _, g := range m.groups {
			if slices.Contains(a.IDs, g.ID) {
				continue
			}
			for _, o := range g.Options {
				m.request.Skills = slices.DeleteFunc(m.request.Skills, func(s string) bool { return s == o.Value })
				m.request.Hooks = slices.DeleteFunc(m.request.Hooks, func(s string) bool { return s == o.Value })
			}
		}
		return m.setupGroup(0)
	}})
}

func (m chatModel) setupGroup(index int) chatModel {
	for index < len(m.groups) && !slices.Contains(m.selectedGroups, m.groups[index].ID) {
		index++
	}
	if index == len(m.groups) {
		return m.setupPlugins()
	}
	m = m.setupStep(3)
	g := m.groups[index]
	m.context = "setup · Extras · " + g.Title
	opts := make([]components.ChoiceOption, 0, len(g.Options))
	pre := m.request.Skills
	if g.ID == "hooks" {
		pre = m.request.Hooks
	}
	selected := make([]string, 0)
	for _, o := range g.Options {
		opts = append(opts, components.ChoiceOption{ID: o.Value, Label: o.Label, Hint: o.Hint})
		if slices.Contains(pre, o.Value) {
			selected = append(selected, o.Value)
		}
	}
	question := g.Title + " — " + g.Description
	var onAnswer func(chatModel, chatAnswer) chatModel
	onAnswer = func(m chatModel, a chatAnswer) chatModel {
		if g.Validate != nil {
			if err := g.Validate(a.IDs); err != nil {
				m = m.say("Não dá para seguir: " + err.Error())
				return m.setupGroup(index)
			}
		}
		if g.ID == "hooks" {
			m.request.Hooks = append([]string(nil), a.IDs...)
		} else {
			for _, o := range g.Options {
				m.request.Skills = slices.DeleteFunc(m.request.Skills, func(s string) bool { return s == o.Value })
			}
			m.request.Skills = append(m.request.Skills, a.IDs...)
		}
		return m.setupGroup(index + 1)
	}
	return m.askFor(chatAsk{hasChoice: true, question: question, choice: components.NewChoice(question, opts, true, false).Preselect(selected), options: opts, onAnswer: onAnswer})
}

func (m chatModel) setupPlugins() chatModel {
	m = m.setupStep(4)
	if !slices.Contains(m.harnessIDs, "claude") && !slices.Contains(m.harnessIDs, "codex") {
		m.addonNames = nil
		m.addonPlan = app.AddonPlan{}
		return m.say("Plugins (caveman, ponytail…) são para Claude Code e Codex; como você escolheu só o OMP, pulo esta etapa.").setupPlan()
	}
	addons := catalog.Addons()
	opts := make([]components.ChoiceOption, 0, len(addons))
	for _, addon := range addons {
		opts = append(opts, components.ChoiceOption{ID: addon.Name, Label: addon.Name, Hint: addonsKindLabel(addon.Kind) + " · " + addon.Summary})
	}
	question := "Quer instalar algum plugin? São instaladores oficiais de terceiros."
	return m.askFor(chatAsk{hasChoice: true, question: question, choice: components.NewChoice(question, opts, true, false).Preselect(m.addonNames), options: opts, onAnswer: func(m chatModel, a chatAnswer) chatModel {
		m.addonNames = slices.Clone(a.IDs)
		m.addonPlan = app.AddonPlan{}
		clear(m.addonSecrets)
		m.addonSecrets = nil
		if len(a.IDs) == 0 {
			return m.setupPlan()
		}
		var harnesses []string
		for _, id := range m.harnessIDs {
			if id == "claude" || id == "codex" {
				harnesses = append(harnesses, id)
			}
		}
		return m.reviewAddons(a.IDs, harnesses, func(m chatModel, plan app.AddonPlan) chatModel {
			m.addonPlan = plan
			return m.setupPlan()
		})
	}})
}

func (m chatModel) setupPlan() chatModel {
	m = m.setupStep(5)
	agents := "atualizar"
	if m.request.SkipAgentUpdate {
		agents = "manter versões"
	}
	labels := make([]string, 0, len(m.harnessIDs))
	for _, id := range m.harnessIDs {
		switch id {
		case "claude":
			labels = append(labels, "Claude Code")
		case "codex":
			labels = append(labels, "Codex")
		case "omp":
			labels = append(labels, "OMP (oh-my-pi)")
		}
	}
	summary := fmt.Sprintf("Resumo da instalação:\n  • Harness: %s\n  • Destino: %s\n  • Core GEX + %d skill(s) e %d hook(s) opcionais\n  • Agentes: %s", strings.Join(labels, ", "), homeLabel(m.request.Home, m.common.UserHome), len(m.request.Skills), len(m.request.Hooks), agents)
	if len(m.addonNames) > 0 {
		var targets []string
		for _, id := range m.harnessIDs {
			if id == "claude" || id == "codex" {
				targets = append(targets, addonHarness(id))
			}
		}
		summary += "\n  • Plugins: " + strings.Join(m.addonNames, ", ") + " (" + strings.Join(targets, ", ") + ")"
	}
	m = m.say(summary)
	m = m.say("Agora vou planejar a instalação.")
	req := m.request
	phases := []string{"Montando as regras do Core…", "Selecionando as skills…", "Comparando com o que já existe…"}
	return m.tool(chatToolSpec{title: "Planejar instalação", params: homeLabel(req.Home, m.common.UserHome), phases: phases, run: func(_ func(string)) chatToolResult {
		if m.deps.Install.Plan == nil {
			return chatToolResult{state: theme.StateError, err: errors.New("planejador de instalação indisponível")}
		}
		plan, err := m.deps.Install.Plan(req)
		if err != nil {
			return chatToolResult{state: theme.StateError, err: err}
		}
		lines := []string{fmt.Sprintf("%d arquivo(s) · %d agente(s) a atualizar", len(plan.Files), len(plan.Updates))}
		for _, diff := range plan.Diffs {
			lines = append(lines, components.ChangeLabel(installAction(diff.Action))+" "+maskHome(diff.Path, m.common.UserHome)+" · "+maskHome(diff.Summary, m.common.UserHome))
		}
		for _, notice := range plan.Notices {
			lines = append(lines, maskHome(notice.Text, m.common.UserHome))
		}
		return chatToolResult{state: theme.StateOK, lines: lines, value: plan}
	}, then: func(m chatModel, r chatToolResult) chatModel {
		if r.err != nil {
			return m.setupFailure(r.err, false)
		}
		m.plan = r.value.(app.Plan)
		if m.plan.Empty() && !m.plan.Blocked {
			if len(m.addonNames) == 0 {
				return m.say("Nada a alterar: está tudo em dia.").flowIdle()
			}
			m = m.say("O Core já está em dia; só os plugins serão instalados.")
		}
		return m.setupConfirm()
	}})
}

func (m chatModel) setupFailure(err error, apply bool) chatModel {
	p := app.Explain(err)
	if len(m.entries) > 0 && m.entries[len(m.entries)-1].isTool {
		m.entries[len(m.entries)-1].tool.Lines = append(m.entries[len(m.entries)-1].tool.Lines, p.Title, p.Cause, p.Action)
	}
	opts := []components.ChoiceOption{{ID: "retry", Label: "Tentar de novo"}, {ID: "menu", Label: "Concluir"}}
	if !apply {
		opts = []components.ChoiceOption{{ID: "retry", Label: "Tentar de novo"}, {ID: "change", Label: "Mudar escolhas"}, {ID: "menu", Label: "Concluir"}}
	}
	return m.askFor(chatAsk{hasChoice: true, question: "O que eu faço?", choice: components.NewChoice("O que eu faço?", opts, false, true), options: opts, onAnswer: func(m chatModel, a chatAnswer) chatModel {
		switch a.IDs[0] {
		case "menu":
			return m.flowIdle()
		case "change":
			return m.flowSetup()
		}
		if apply {
			return m.setupApply()
		}
		return m.setupPlan()
	}})
}

func (m chatModel) setupConfirm() chatModel {
	m = m.setupStep(len(m.steps) - 2)
	opts := []components.ChoiceOption{{ID: "apply", Label: "Sim, aplicar"}, {ID: "diff", Label: "Ver diff completo"}, {ID: "back", Label: "Mudar escolhas"}, {ID: "cancel", Label: "Cancelar"}}
	if len(m.addonNames) > 0 {
		m = m.say(fmt.Sprintf("Depois do Core vou rodar %d comando(s) de terceiros (%s); os comandos exatos estão na revisão da etapa Plugins.", len(m.addonPlan.Steps), strings.Join(m.addonNames, ", ")))
		opts = []components.ChoiceOption{{ID: "cancel", Label: "Cancelar"}, {ID: "apply", Label: "Sim, aplicar"}, {ID: "diff", Label: "Ver diff completo"}, {ID: "back", Label: "Mudar escolhas"}}
	}
	return m.askFor(chatAsk{hasChoice: true, question: "Aplico essas alterações?", choice: components.NewChoice("Aplico essas alterações?", opts, false, true), options: opts, onAnswer: func(m chatModel, a chatAnswer) chatModel {
		switch a.IDs[0] {
		case "apply":
			return m.setupApply()
		case "diff":
			for i := len(m.entries) - 1; i >= 0; i-- {
				if m.entries[i].isTool {
					m.entries[i].tool.Expanded = true
					break
				}
			}
			return m.setupConfirm()
		case "back":
			return m.flowSetup()
		default:
			return m.say("Nada foi gravado.").flowIdle()
		}
	}})
}

func (m chatModel) setupApply() chatModel {
	m = m.setupStep(len(m.steps) - 1)
	plan, req := m.plan, m.request
	if plan.Empty() && !plan.Blocked {
		return m.setupApplyAddons()
	}
	m = m.say("Agora vou instalar os componentes do Core.")
	phases := []string{"Criando regras…", "Instalando skills Core…"}
	if len(req.Skills) > 0 {
		phases = append(phases, "Instalando skills opcionais…")
	}
	if len(req.Hooks) > 0 {
		phases = append(phases, "Gravando hooks…")
	}
	phases = append(phases, "Salvando o histórico para rollback…")
	return m.tool(chatToolSpec{title: "Instalar Core", write: true, phases: phases, run: func(progress func(string)) chatToolResult {
		if m.deps.Install.Apply == nil {
			return chatToolResult{state: theme.StateError, err: errors.New("instalador indisponível")}
		}
		res, err := m.deps.Install.Apply(plan, func(step app.Step) { progress("✓ " + installStepLine(step)) })
		if err != nil {
			return chatToolResult{state: theme.StateError, err: err}
		}
		return chatToolResult{state: theme.StateOK, value: res}
	}, then: func(m chatModel, r chatToolResult) chatModel {
		if r.err != nil {
			return m.setupFailure(r.err, true)
		}
		res := r.value.(app.Result)
		m = m.say(fmt.Sprintf("Pronto! Instalei %d arquivo(s). Histórico: %s.", len(plan.Files), res.HistoryID))
		if res.HooksNeedTrust {
			m = m.say("No Codex, abra /hooks e aprove o hook novo.")
		}
		if res.Blocked {
			m.exitCode = 1
			m = m.say("Agentes bloqueados não foram atualizados.")
		}
		return m.setupApplyAddons()
	}})
}

func (m chatModel) setupApplyAddons() chatModel {
	command := "Comando equivalente: " + app.EquivalentCommand(m.request)
	if len(m.addonNames) > 0 {
		command += " (plugins não têm comando CLI equivalente)"
		plan := m.addonPlan
		m.addonPlan = app.AddonPlan{}
		return m.installAddons(plan, m.request.Home, func(m chatModel) chatModel {
			return m.say(command).flowIdle()
		})
	}
	return m.say(command).flowIdle()
}

func agentBannerTitle(label string, change app.AgentChange) string {
	switch {
	case change.Current != "":
		return label + " v" + strings.TrimPrefix(change.Current, "v")
	case change.Blocked:
		return label + " (não instalado pelo kit)"
	}
	return label + " (não instalado)"
}
