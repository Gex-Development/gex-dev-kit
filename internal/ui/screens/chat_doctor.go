package screens

import (
	"fmt"
	"slices"

	"github.com/Gex-Development/gex-dev-kit/internal/app"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/components"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

func doctorStateKind(state string) theme.State {
	switch state {
	case "ok":
		return theme.StateOK
	case "pendente", "sem-historico":
		return theme.StatePending
	case "atualizar":
		return theme.StateWarn
	case "aviso":
		return theme.StateInfo
	case "alterado", "ausente", "incompleto", "bloqueado":
		return theme.StateError
	}
	return theme.StateWarn
}

func doctorLines(report app.Report, plain bool, userHome string) []string {
	var attention, warnings, pending, ok []string
	for _, item := range report.Items {
		line := doctorStateKind(item.State).Symbol(plain) + " " + item.State + " " + maskHome(item.Path, userHome)
		switch app.StateLevel(item.State) {
		case "atencao":
			attention = append(attention, line)
		case "ok":
			ok = append(ok, line)
		default:
			pending = append(pending, line)
		}
	}
	for _, warning := range report.Warnings {
		warnings = append(warnings, "aviso · "+maskHome(warning.Path, userHome)+" · "+maskHome(warning.Detail, userHome))
	}
	return slices.Concat(attention, warnings, pending, ok)
}

func (m chatModel) flowDoctor() chatModel {
	m.context = "doctor"
	m = m.say("Vou conferir a integridade do que o kit instalou.")
	return m.tool(chatToolSpec{title: "Doctor", params: homeLabel(m.common.Home, m.common.UserHome), phases: []string{"Lendo os arquivos instalados…", "Comparando com o catálogo…", "Conferindo os hooks…"}, run: func(_ func(string)) chatToolResult {
		if m.deps.Doctor == nil {
			return chatToolResult{state: theme.StateError, err: fmt.Errorf("relatório indisponível")}
		}
		report, err := m.deps.Doctor()
		if err != nil {
			return chatToolResult{state: theme.StateError, err: err}
		}
		_, attention, _ := report.Counts()
		state := theme.StateOK
		if attention > 0 {
			state = theme.StateWarn
		}
		return chatToolResult{state: state, lines: doctorLines(report, m.theme.Plain, m.common.UserHome), value: report}
	}, then: func(m chatModel, r chatToolResult) chatModel {
		if r.err != nil {
			p := app.Explain(r.err)
			return m.say(p.Title + ": " + p.Cause).flowIdle()
		}
		report := r.value.(app.Report)
		m.exitCode = report.ExitCode()
		ok, attention, _ := report.Counts()
		if attention == 0 {
			return m.say(fmt.Sprintf("Tudo íntegro: %d item(ns) ok.", ok)).flowIdle()
		}
		m = m.say(fmt.Sprintf("Encontrei %d item(ns) com atenção.", attention))
		opts := []components.ChoiceOption{{ID: "setup", Label: "Reinstalar com o setup"}, {ID: "menu", Label: "Concluir"}}
		return m.askFor(chatAsk{hasChoice: true, question: "O que eu faço?", choice: components.NewChoice("O que eu faço?", opts, false, true), options: opts, onAnswer: func(m chatModel, a chatAnswer) chatModel {
			if a.IDs[0] == "setup" {
				return m.flowSetup()
			}
			return m.flowIdle()
		}})
	}})
}
