package screens

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/Gex-Development/gex-dev-kit/internal/app"
	"github.com/Gex-Development/gex-dev-kit/internal/catalog"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/components"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

type GitHooksDeps struct {
	Plan  func(app.GitHooksRequest) (app.GitHooksPlan, error)
	Apply func(app.GitHooksPlan) (app.Result, error)
}

func gitHooksGuard(d GitHooksDeps, group *sync.WaitGroup) GitHooksDeps {
	apply := d.Apply
	if apply != nil {
		d.Apply = func(p app.GitHooksPlan) (app.Result, error) { group.Add(1); defer group.Done(); return apply(p) }
	}
	return d
}

func (m chatModel) flowGitHooks() chatModel {
	m.context = "git hooks"
	req := m.gitHooksRequest
	if req.Home == "" {
		req.Home = m.common.Home
	}
	if req.Repo != "" {
		m.gitHooksRequest = req
		return m.gitHooksChoices()
	}
	return m.askFor(chatAsk{input: true, path: true, picker: components.NewPathPicker(m.common.UserHome, []string{".git"}), question: "Caminho absoluto do repositório Git:", validate: func(value string) error {
		if err := installAbsolutePath(value); err != nil {
			return err
		}
		path := filepath.Clean(strings.TrimSpace(value))
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return errors.New("repositório precisa ser uma pasta")
		}
		return nil
	}, onAnswer: func(m chatModel, a chatAnswer) chatModel {
		req.Repo = filepath.Clean(strings.TrimSpace(a.Text))
		m.gitHooksRequest = req
		return m.gitHooksChoices()
	}})
}

func (m chatModel) gitHooksChoices() chatModel {
	req := m.gitHooksRequest
	opts := make([]components.ChoiceOption, 0)
	for _, name := range catalog.GitHooks() {
		spec, ok := catalog.HookSpecByName(name)
		if ok {
			opts = append(opts, components.ChoiceOption{ID: name, Label: spec.GitHook, Hint: spec.Description})
		}
	}
	return m.askFor(chatAsk{hasChoice: true, question: "Quais git hooks eu configuro?", choice: components.NewChoice("Quais git hooks eu configuro?", opts, true, true).Preselect(req.Hooks), options: opts, onAnswer: func(m chatModel, a chatAnswer) chatModel {
		req.Hooks = slices.Clone(a.IDs)
		if !slices.Contains(req.Hooks, "gex-pre-commit") {
			req.PreCommit = nil
		}
		if !slices.Contains(req.Hooks, "gex-pre-push") {
			req.PrePush = nil
		}
		m.gitHooksRequest = req
		return m.gitHooksCommands(0)
	}})
}

func (m chatModel) gitHooksCommands(index int) chatModel {
	req := m.gitHooksRequest
	names := []struct{ id, label string }{{"gex-pre-commit", "pre-commit"}, {"gex-pre-push", "pre-push"}}
	for index < len(names) && !slices.Contains(req.Hooks, names[index].id) {
		index++
	}
	if index >= len(names) {
		return m.gitHooksPlan()
	}
	item := names[index]
	question := "Comandos do " + item.label + " (um por linha, até 8; vazio pula):"
	return m.askFor(chatAsk{textarea: true, question: question, validate: func(text string) error {
		lines := chatGitHooksLines(text)
		if problem := chatGitHooksCheck(item.label, lines); problem != "" {
			return errors.New(problem)
		}
		return nil
	}, onAnswer: func(m chatModel, a chatAnswer) chatModel {
		lines := chatGitHooksLines(a.Text)
		if item.id == "gex-pre-commit" {
			req.PreCommit = lines
		} else {
			req.PrePush = lines
		}
		m.gitHooksRequest = req
		return m.gitHooksCommands(index + 1)
	}})
}

func chatGitHooksLines(value string) []string {
	var lines []string
	for _, line := range strings.Split(value, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func chatGitHooksCheck(label string, lines []string) string {
	if len(lines) > 8 {
		return fmt.Sprintf("%s: use no máximo 8 comandos.", label)
	}
	for _, line := range lines {
		if len(line) > 300 {
			return fmt.Sprintf("%s: comando maior que 300 caracteres.", label)
		}
	}
	return ""
}

func (m chatModel) gitHooksPlan() chatModel {
	req := m.gitHooksRequest
	return m.tool(chatToolSpec{title: "Planejar git hooks", params: homeLabel(req.Repo, m.common.UserHome), phases: []string{"Lendo o repositório…", "Montando os hooks…"}, run: func(_ func(string)) chatToolResult {
		if m.deps.GitHooks.Plan == nil {
			return chatToolResult{state: theme.StateError, err: errors.New("planejador de git hooks indisponível")}
		}
		plan, err := m.deps.GitHooks.Plan(req)
		if err != nil {
			return chatToolResult{state: theme.StateError, err: err}
		}
		lines := make([]string, 0, len(plan.Diffs)+len(plan.Sources))
		for _, diff := range plan.Diffs {
			lines = append(lines, components.ChangeLabel(installAction(diff.Action))+" "+maskHome(diff.Path, m.common.UserHome)+" · "+maskHome(diff.Summary, m.common.UserHome))
		}
		for _, source := range plan.Sources {
			lines = append(lines, "documentado · "+maskHome(source.File, m.common.UserHome)+" · "+source.Command)
		}
		return chatToolResult{state: theme.StateOK, lines: lines, value: plan}
	}, then: func(m chatModel, r chatToolResult) chatModel {
		if r.err != nil {
			p := app.Explain(r.err)
			if len(m.entries) > 0 && m.entries[len(m.entries)-1].isTool {
				m.entries[len(m.entries)-1].tool.Lines = append(m.entries[len(m.entries)-1].tool.Lines, p.Title, p.Cause, p.Action)
			}
			m = m.say(p.Title + ": " + p.Cause + ". " + p.Action)
			return m.gitHooksFailure(nil)
		}
		plan := r.value.(app.GitHooksPlan)
		opts := []components.ChoiceOption{{ID: "yes", Label: "Sim, gravar"}, {ID: "no", Label: "Cancelar"}}
		return m.askFor(chatAsk{hasChoice: true, question: "Gravo esses hooks?", choice: components.NewChoice("Gravo esses hooks?", opts, false, true), options: opts, onAnswer: func(m chatModel, a chatAnswer) chatModel {
			if a.IDs[0] == "no" {
				return m.say("Nada foi gravado.").flowIdle()
			}
			return m.gitHooksApply(plan)
		}})
	}})
}

func (m chatModel) gitHooksFailure(plan *app.GitHooksPlan) chatModel {
	opts := []components.ChoiceOption{{ID: "retry", Label: "Tentar de novo"}, {ID: "menu", Label: "Concluir"}}
	return m.askFor(chatAsk{hasChoice: true, question: "O que eu faço?", choice: components.NewChoice("O que eu faço?", opts, false, true), options: opts, onAnswer: func(m chatModel, a chatAnswer) chatModel {
		if a.IDs[0] == "retry" {
			if plan != nil {
				return m.gitHooksApply(*plan)
			}
			return m.gitHooksPlan()
		}
		return m.flowIdle()
	}})
}

func (m chatModel) gitHooksApply(plan app.GitHooksPlan) chatModel {
	return m.tool(chatToolSpec{title: "Gravar git hooks", write: true, phases: []string{"Gravando os hooks…", "Salvando o histórico…"}, run: func(_ func(string)) chatToolResult {
		if m.deps.GitHooks.Apply == nil {
			return chatToolResult{state: theme.StateError, err: errors.New("gravador de git hooks indisponível")}
		}
		result, err := m.deps.GitHooks.Apply(plan)
		if err != nil {
			return chatToolResult{state: theme.StateError, err: err}
		}
		return chatToolResult{state: theme.StateOK, value: result}
	}, then: func(m chatModel, r chatToolResult) chatModel {
		if r.err != nil {
			p := app.Explain(r.err)
			if len(m.entries) > 0 && m.entries[len(m.entries)-1].isTool {
				m.entries[len(m.entries)-1].tool.Lines = append(m.entries[len(m.entries)-1].tool.Lines, p.Title, p.Cause, p.Action)
			}
			return m.say(p.Title + ": " + p.Cause + ". " + p.Action).gitHooksFailure(&plan)
		}
		result := r.value.(app.Result)
		return m.say(fmt.Sprintf("Hooks gravados. Histórico: %s.", result.HistoryID)).flowIdle()
	}})
}
