package components

import (
	"strings"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

func ErrorBox(t theme.Theme, width int, title, cause, action string) string {
	causeLabel, actionLabel := errorboxLabels(t.Plain)
	defaultCause, defaultAction := errorboxDefaults(t.Plain)
	if cause == "" {
		cause = defaultCause
	}
	if action == "" {
		action = defaultAction
	}
	content := strings.Join([]string{
		t.StateStyle(theme.StateError).Render(title),
		"",
		t.Strong.Render(causeLabel) + " " + cause,
		t.Strong.Render(actionLabel) + " " + action,
	}, "\n")
	box := t.Panel.Border(panelBorder(t, BorderThick)).BorderForeground(t.Palette.Negative)
	return panelFit(box, width).Render(content)
}

func errorboxLabels(plain bool) (string, string) {
	if plain {
		return "Causa:", "Acao:"
	}
	return "Causa:", "Ação:"
}

func errorboxDefaults(plain bool) (string, string) {
	if plain {
		return "nao informada", "Rode gex-dev-kit doctor e envie a saida ao responsavel"
	}
	return "não informada", "Rode gex-dev-kit doctor e envie a saída ao responsável"
}
