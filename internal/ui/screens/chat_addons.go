package screens

import (
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/Gex-Development/gex-dev-kit/internal/app"
	"github.com/Gex-Development/gex-dev-kit/internal/catalog"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/components"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

const (
	addonsDisclaimer = "Plugins de terceiros: revise fonte e licença. O kit só roda o comando oficial depois da sua confirmação."
	addonsShortSha   = 7
)

type AddonDeps struct {
	Plan    func(names, harnesses []string, secrets map[string]string) (app.AddonPlan, error)
	Secrets func(names []string) ([]catalog.AddonSecret, error)
	Install func(plan app.AddonPlan, home string, progress func(app.AddonStepResult)) app.AddonResult
}

func addonsGuard(d AddonDeps, group *sync.WaitGroup) AddonDeps {
	install := d.Install
	if install != nil {
		d.Install = func(plan app.AddonPlan, home string, progress func(app.AddonStepResult)) app.AddonResult {
			group.Add(1)
			defer group.Done()
			return install(plan, home, progress)
		}
	}
	return d
}

func (m chatModel) flowAddons() chatModel {
	m.history = nil
	clear(m.addonSecrets)
	m.addonSecrets = nil
	m.addonPlan = app.AddonPlan{}
	m.context = "addons"
	addons := catalog.Addons()
	opts := make([]components.ChoiceOption, 0, len(addons)+1)
	for _, a := range addons {
		opts = append(opts, components.ChoiceOption{ID: a.Name, Label: a.Name, Hint: addonsKindLabel(a.Kind)})
	}
	opts = append(opts, components.ChoiceOption{ID: "menu", Label: "Concluir"})
	return m.askFor(chatAsk{hasChoice: true, question: "Qual addon você quer ver?", choice: components.NewChoice("Qual addon você quer ver?", opts, false, true), options: opts, onAnswer: func(m chatModel, a chatAnswer) chatModel {
		if a.IDs[0] == "menu" {
			return m.flowIdle()
		}
		addon, ok := catalog.AddonByName(a.IDs[0])
		if !ok {
			return m.flowAddons()
		}
		m = m.say(addonsCard(m.theme, addon, max(1, m.width-2)))
		opts := []components.ChoiceOption{
			{ID: "claude", Label: "Instalar no Claude Code"},
			{ID: "codex", Label: "Instalar no Codex"},
			{ID: "both", Label: "Instalar nos dois"},
			{ID: "back", Label: "Voltar à lista"},
		}
		question := "O que você quer fazer com " + addon.Name + "?"
		return m.askFor(chatAsk{hasChoice: true, question: question, choice: components.NewChoice(question, opts, false, true), options: opts, onAnswer: func(m chatModel, a chatAnswer) chatModel {
			if a.IDs[0] == "back" {
				return m.flowAddons()
			}
			harnesses := []string{a.IDs[0]}
			if a.IDs[0] == "both" {
				harnesses = []string{"claude", "codex"}
			}
			return m.reviewAddons([]string{addon.Name}, harnesses, func(m chatModel, plan app.AddonPlan) chatModel {
				return m.confirmAddons(plan)
			})
		}})
	}})
}

func addonHarness(id string) string {
	if id == "claude" {
		return "Claude Code"
	}
	return "Codex"
}

func (m chatModel) reviewAddons(names, harnesses []string, next func(chatModel, app.AddonPlan) chatModel) chatModel {
	if m.deps.Addons.Plan == nil || m.deps.Addons.Secrets == nil {
		return m.say("Planejador de plugins indisponível.").flowIdle()
	}
	secrets, err := m.deps.Addons.Secrets(names)
	if err != nil {
		return m.say("Não consegui consultar as chaves dos plugins.").flowIdle()
	}
	if len(secrets) > 0 {
		return m.askAddonSecret(names, harnesses, secrets, 0, next)
	}
	return m.reviewAddonsPlan(names, harnesses, next)
}

func (m chatModel) askAddonSecret(names, harnesses []string, secrets []catalog.AddonSecret, index int, next func(chatModel, app.AddonPlan) chatModel) chatModel {
	secret := secrets[index]
	var files []string
	if slices.Contains(harnesses, "claude") {
		files = append(files, "~/.claude.json")
	}
	if slices.Contains(harnesses, "codex") {
		files = append(files, "~/.codex/config.toml")
	}
	reader := "os lê"
	if len(files) == 1 {
		reader = "o lê"
	}
	m = m.say("A " + secret.Label + " fica em texto puro em " + strings.Join(files, " e ") + "; só você " + reader + ".")
	m = m.say("Para criar: " + secret.URL + ". Deixe vazio e tecle enter para cancelar.")
	return m.askFor(chatAsk{input: true, secret: true, question: "Cole a " + secret.Label, onAnswer: func(m chatModel, a chatAnswer) chatModel {
		if strings.TrimSpace(a.Text) == "" {
			clear(m.addonSecrets)
			m.addonSecrets = nil
			m.addonNames = nil
			if m.context == "addons" {
				return m.flowAddons()
			}
			return m.setupPlugins()
		}
		if m.addonSecrets == nil {
			m.addonSecrets = make(map[string]string, len(secrets))
		}
		m.addonSecrets[secret.Env] = a.Text
		if index+1 < len(secrets) {
			return m.askAddonSecret(names, harnesses, secrets, index+1, next)
		}
		return m.reviewAddonsPlan(names, harnesses, next)
	}})
}

func (m chatModel) reviewAddonsPlan(names, harnesses []string, next func(chatModel, app.AddonPlan) chatModel) chatModel {
	plan, err := m.deps.Addons.Plan(names, harnesses, m.addonSecrets)
	for _, value := range m.addonSecrets {
		if err != nil && value != "" {
			err = fmt.Errorf("%s", strings.ReplaceAll(err.Error(), value, "[oculto]"))
		}
	}
	clear(m.addonSecrets)
	m.addonSecrets = nil
	if err != nil {
		return m.say("Não consegui planejar os plugins: " + err.Error()).flowIdle()
	}
	var lines []string
	lines = append(lines, "Vou rodar estes comandos (código de terceiros, confira fonte e licença):")
	for _, name := range names {
		addon, _ := catalog.AddonByName(name)
		for _, harness := range harnesses {
			if harness != "claude" && harness != "codex" {
				continue
			}
			for _, command := range addon.DisplayCommands(harness) {
				lines = append(lines, "  "+name+" · "+addonHarness(harness)+": "+strings.NewReplacer(catalog.SecretPlaceholder, "••••••••", "<sua chave>", "••••••••").Replace(command))
			}
		}
	}
	lines = append(lines, "Atenção:")
	for _, name := range names {
		addon, _ := catalog.AddonByName(name)
		for _, conflict := range addon.Conflicts {
			lines = append(lines, "  "+name+": "+conflict)
		}
		lines = append(lines, "  "+name+": "+addon.CoreCompat)
	}
	for _, manual := range plan.Manual {
		if manual.Harness == "" {
			lines = append(lines, fmt.Sprintf("Passo manual (%s): %s", manual.Addon, manual.Text))
		} else {
			lines = append(lines, fmt.Sprintf("Passo manual (%s, %s): %s", manual.Addon, addonHarness(manual.Harness), manual.Text))
		}
	}
	m = m.say(strings.Join(lines, "\n"))
	if len(plan.Missing) > 0 {
		m = m.say("Não encontrei no PATH: " + strings.Join(plan.Missing, ", ") + ". Esses passos serão pulados e viram instrução manual.")
	}
	return next(m, plan)
}

func (m chatModel) confirmAddons(plan app.AddonPlan) chatModel {
	opts := []components.ChoiceOption{{ID: "cancel", Label: "Cancelar"}, {ID: "install", Label: "Sim, instalar"}}
	return m.askFor(chatAsk{hasChoice: true, question: "Rodo estes comandos?", choice: components.NewChoice("Rodo estes comandos?", opts, false, false), options: opts, onAnswer: func(m chatModel, answer chatAnswer) chatModel {
		if answer.IDs[0] == "cancel" {
			return m.flowAddons()
		}
		m.history = nil
		return m.installAddons(plan, m.common.Home, func(m chatModel) chatModel { return m.flowAddons() })
	}})
}

func (m chatModel) installAddons(plan app.AddonPlan, home string, next func(chatModel) chatModel) chatModel {
	return m.tool(chatToolSpec{title: "Instalar plugins", write: true, run: func(progress func(string)) chatToolResult {
		if m.deps.Addons.Install == nil {
			return chatToolResult{state: theme.StateError, err: fmt.Errorf("instalador de plugins indisponível")}
		}
		result := m.deps.Addons.Install(plan, home, func(item app.AddonStepResult) {
			if item.Status == app.AddonPending {
				progress("Instalando " + item.Step.Addon + " no " + addonHarness(item.Step.Harness) + "…")
			}
		})
		rows := make([]components.ToolRow, 0, len(result.Steps)+len(result.Manual))
		index := make(map[string]int)
		state := theme.StateOK
		for _, item := range result.Steps {
			label := item.Step.Addon + " · " + addonHarness(item.Step.Harness)
			row := components.ToolRow{State: theme.StateOK, Label: label, Detail: "instalado"}
			switch item.Status {
			case app.AddonFailed:
				state, row.State, row.Detail = theme.StateError, theme.StateError, "falhou · "+lastAddonLine(item.Output)
			case app.AddonSkipped:
				row.State, row.Detail = theme.StatePending, "ignorado · falha anterior"
				if !item.Step.Available {
					row.Detail = "ignorado · faltou " + item.Step.BlockedBy
				}
			}
			if i, ok := index[label]; ok {
				if rows[i].State != theme.StateError && (rows[i].State != theme.StatePending || row.State == theme.StateError) {
					rows[i] = row
				}
				continue
			}
			index[label] = len(rows)
			rows = append(rows, row)
		}
		for _, item := range result.Manual {
			label := item.Addon
			if item.Harness != "" {
				label += " · " + addonHarness(item.Harness)
			}
			rows = append(rows, components.ToolRow{State: theme.StatePending, Label: label, Detail: "manual · " + item.Text})
		}
		return chatToolResult{state: state, rows: rows, value: result}
	}, then: func(m chatModel, r chatToolResult) chatModel {
		if r.err != nil {
			m.exitCode = 1
			return m.say("Não consegui instalar plugins: " + r.err.Error()).flowIdle()
		}
		result := r.value.(app.AddonResult)
		installed := make([]string, 0)
		codex := false
		incomplete := make(map[string]bool)
		for _, item := range result.Steps {
			if item.Status != app.AddonOK {
				incomplete[item.Step.Addon+"·"+item.Step.Harness] = true
			}
			if item.Status == app.AddonFailed {
				m.exitCode = 1
			}
			if item.Status == app.AddonSkipped && !item.Step.Available && item.Step.BlockedBy == item.Step.Argv[0] {
				m = m.say("Passo manual (" + item.Step.Addon + ", " + addonHarness(item.Step.Harness) + "): instale " + item.Step.Argv[0] + " e rode " + item.Step.Display)
			}
		}
		for _, item := range result.Steps {
			if item.Status == app.AddonOK && !incomplete[item.Step.Addon+"·"+item.Step.Harness] {
				if !slices.Contains(installed, item.Step.Addon) {
					installed = append(installed, item.Step.Addon)
				}
				codex = codex || (item.Step.Harness == "codex" && !addonIsMCP(item.Step.Addon))
			}
		}
		if m.exitCode == 1 {
			m = m.say("Alguns plugins não foram instalados; veja os passos que falharam.")
		}
		if len(installed) > 0 {
			m = m.say("Plugins instalados: " + strings.Join(installed, ", ") + ".")
		}
		if codex {
			m = m.say("No Codex, abra /hooks e aprove os hooks novos.")
		}
		m = m.say("O rollback do kit não desfaz plugins; para remover, use o próprio Claude Code ou Codex.")
		for _, item := range result.Manual {
			m = m.say(fmt.Sprintf("Passo manual (%s, %s): %s", item.Addon, addonHarness(item.Harness), item.Text))
		}
		return next(m)
	}})
}

func addonIsMCP(name string) bool {
	addon, ok := catalog.AddonByName(name)
	return ok && addon.Kind == "mcp"
}

func lastAddonLine(output string) string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	return lines[len(lines)-1]
}

func addonsKindLabel(kind string) string {
	switch kind {
	case "metodo":
		return "método"
	case "mcp":
		return "MCP"
	default:
		return kind
	}
}

func addonsShortRevision(revision string) string {
	if len(revision) < addonsShortSha {
		return revision
	}
	return revision[:addonsShortSha]
}

func addonsCard(t theme.Theme, a catalog.Addon, width int) string {
	lines := []string{
		addonsField(t, "Resumo", a.Summary, width),
		addonsField(t, "Fonte", a.Source, width),
		addonsField(t, "Licença", a.License, width),
		addonsField(t, "Versões conferidas em "+a.CheckedAt, "release "+a.Release+" · commit conferido "+addonsShortRevision(a.Revision), width),
		addonsCommand(t, "Comando oficial no Claude Code", a.InstallClaude, width),
		addonsCommand(t, "Comando oficial no Codex", a.InstallCodex, width),
		t.Warning.Width(width).Render(addonsDisclaimer),
		"",
		addonsList(t, "O que instala", a.Installs, width),
		"",
		addonsList(t, "Conflitos", a.Conflicts, width),
		"",
		addonsField(t, "Compatibilidade com o Core", a.CoreCompat, width),
	}
	return strings.Join(lines, "\n")
}

func addonsField(t theme.Theme, label, value string, width int) string {
	head := t.Strong.Render(label + ":")
	return t.Text.Width(width).Render(head + " " + value)
}

func addonsCommand(t theme.Theme, label, command string, width int) string {
	if strings.TrimSpace(command) == "" {
		return addonsField(t, label, "sem comando oficial publicado.", width)
	}
	return t.Strong.Render(label+":") + "\n" + t.Code.Width(width).Render(command)
}

func addonsList(t theme.Theme, label string, items []string, width int) string {
	lines := []string{t.Strong.Render(label + ":")}
	for _, item := range items {
		lines = append(lines, addonsBullet(t, item, width))
	}
	return strings.Join(lines, "\n")
}

func addonsBullet(t theme.Theme, text string, width int) string {
	bullet := "•"
	if t.Plain {
		bullet = "-"
	}
	wrapped := strings.Split(t.Text.Width(max(width-2, 1)).Render(strings.ReplaceAll(text, "\t", " ")), "\n")
	for i, line := range wrapped {
		prefix := "  "
		if i == 0 {
			prefix = bullet + " "
		}
		wrapped[i] = prefix + line
	}
	return strings.Join(wrapped, "\n")
}
