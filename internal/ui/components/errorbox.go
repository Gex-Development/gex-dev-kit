package components

import (
	"strings"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

func ErrorBox(t theme.Theme, width int, title, cause, action string) string {
	if cause == "" {
		cause = "não informada"
	}
	if action == "" {
		action = "Rode gex-dev-kit doctor e envie a saída ao responsável"
	}
	content := strings.Join([]string{
		t.StateStyle(theme.StateError).Render(title),
		"",
		t.Strong.Render("Causa:") + " " + cause,
		t.Strong.Render("Ação:") + " " + action,
	}, "\n")
	box := t.Panel.Border(panelBorder(t, BorderThick)).BorderForeground(t.Palette.Negative)
	return panelFit(box, width).Render(content)
}
