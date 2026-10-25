package screens

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/Gex-Development/gex-dev-kit/internal/app"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/components"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

type RollbackDeps struct {
	Plan  func() (app.RollbackPreview, error)
	Apply func() (string, error)
}

func rollbackGuard(d RollbackDeps, group *sync.WaitGroup) RollbackDeps {
	apply := d.Apply
	if apply != nil {
		d.Apply = func() (string, error) {
			group.Add(1)
			defer group.Done()
			return apply()
		}
	}
	return d
}

func (m chatModel) flowRollback(agents bool) chatModel {
	title, done := "Ler histórico", "Restaurar"
	d := m.deps.Rollback
	if agents {
		title = "Ler versões anteriores"
		done = "Restaurar agentes"
		d = m.deps.RollbackAgents
		m.context = "rollback-agents"
	} else {
		m.context = "rollback"
		m = m.say("Vou ler o histórico da última instalação.")
	}
	return m.tool(chatToolSpec{title: title, phases: []string{"Lendo o histórico da última instalação…", "Conferindo o que mudou depois…"}, run: func(_ func(string)) chatToolResult {
		if d.Plan == nil {
			return chatToolResult{state: theme.StateError, err: errors.New("histórico indisponível")}
		}
		p, err := d.Plan()
		if err != nil {
			return chatToolResult{state: theme.StateError, err: err}
		}
		lines := make([]string, 0, len(p.Diffs))
		for _, diff := range p.Diffs {
			lines = append(lines, components.ChangeLabel(installAction(diff.Action))+" "+maskHome(diff.Path, m.common.UserHome)+" · "+maskHome(diff.Summary, m.common.UserHome))
		}
		for _, path := range p.Plan.Conflicts {
			lines = append(lines, "conflito · "+maskHome(path, m.common.UserHome))
		}
		return chatToolResult{state: theme.StateOK, lines: lines, value: p}
	}, then: func(m chatModel, r chatToolResult) chatModel {
		if r.err != nil {
			m.exitCode = 1
			p := app.Explain(r.err)
			return m.say(p.Title + ": " + p.Cause).flowIdle()
		}
		p := r.value.(app.RollbackPreview)
		if !agents && p.Plan.ID == "" && len(p.Plan.Conflicts) == 0 {
			return m.say("Nada a restaurar.").flowIdle()
		}
		if len(p.Plan.Conflicts) > 0 {
			m.exitCode = 1
			return m.say("Não restauro: estes arquivos mudaram depois da instalação.").flowIdle()
		}
		question := "Restauro esta instalação?"
		if agents {
			question = "Restauro Claude e Codex às versões anteriores ao kit?"
		}
		opts := []components.ChoiceOption{{ID: "yes", Label: "Sim, restaurar"}, {ID: "no", Label: "Cancelar"}}
		if !agents {
			for _, entry := range p.Plan.Entries {
				if entry.Action == "remover" && strings.HasPrefix(entry.Path, ".local/share/gex-dev-kit/bin/") {
					m = m.say("Os hooks passam a ignorar a verificação até nova instalação.")
					break
				}
			}
		}
		return m.askFor(chatAsk{hasChoice: true, question: question, choice: components.NewChoice(question, opts, false, true), options: opts, onAnswer: func(m chatModel, a chatAnswer) chatModel {
			if a.IDs[0] == "no" {
				return m.say("Nada foi restaurado.").flowIdle()
			}
			return m.tool(chatToolSpec{title: done, write: true, phases: []string{"Restaurando os arquivos…", "Conferindo o resultado…"}, run: func(_ func(string)) chatToolResult {
				if d.Apply == nil {
					return chatToolResult{state: theme.StateError, err: errors.New("restauração indisponível")}
				}
				id, err := d.Apply()
				if err != nil {
					return chatToolResult{state: theme.StateError, err: err}
				}
				return chatToolResult{state: theme.StateOK, value: id}
			}, then: func(m chatModel, r chatToolResult) chatModel {
				if r.err != nil {
					m.exitCode = 1
					p := app.Explain(r.err)
					return m.say(p.Title + ": " + p.Cause).flowIdle()
				}
				if agents {
					return m.say("Agentes restaurados.").flowIdle()
				}
				return m.say(fmt.Sprintf("Restaurado. Histórico: %s.", r.value.(string))).flowIdle()
			}})
		}})
	}})
}
