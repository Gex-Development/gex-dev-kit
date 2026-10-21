package components

import (
	"strings"

	"charm.land/bubbles/v2/help"
	"github.com/charmbracelet/x/ansi"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

func KeyHints(t theme.Theme, width int, showAll bool, keys help.KeyMap) string {
	if width <= 0 {
		return ""
	}
	m := help.New()
	m.Styles = t.Help()
	m.SetWidth(width)
	m.ShowAll = showAll
	lines := strings.Split(m.View(keys), "\n")
	for i, line := range lines {
		lines[i] = ansi.Truncate(line, width, keyhintsTail(t))
	}
	return strings.Join(lines, "\n")
}

func keyhintsTail(t theme.Theme) string {
	if t.Plain {
		return "..."
	}
	return "…"
}
