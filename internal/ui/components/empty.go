package components

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

func EmptyState(t theme.Theme, width int, title, hint string) string {
	lines := []string{t.Strong.Render(title)}
	if hint != "" {
		lines = append(lines, t.Secondary.Render(hint))
	}
	style := t.Text.Align(lipgloss.Center)
	if width > 0 {
		style = style.Width(width).MaxWidth(width)
	}
	return style.Render(strings.Join(lines, "\n"))
}
