package components

import (
	"strings"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
	"github.com/charmbracelet/x/ansi"
)

func PromptBox(t theme.Theme, width int, segments []StatusSegment, metrics []StatusMetric, view string, locked bool, placeholder string) []string {
	if width < 2 {
		return nil
	}
	edge := t.PromptEdge
	if locked {
		edge = t.Faint
	}
	h, v, topLeft, topRight, bottomLeft, bottomRight := "─", "│", "╭", "╮", "╰", "╯"
	if t.Plain {
		h, v, topLeft, topRight, bottomLeft, bottomRight = "-", "|", "+", "+", "+", "+"
	}
	barWidth := max(0, width-4)
	rows := []string{edge.Render(topLeft+h) + StatusLine(t, barWidth, segments, metrics) + edge.Render(h+topRight)}
	if locked {
		view = placeholder
	}
	lines := strings.Split(view, "\n")
	for i, line := range lines {
		prefix := "  "
		if i == 0 {
			prefix = "> "
		}
		available := max(0, width-5)
		text := ansi.Truncate(line, available, "")
		pad := strings.Repeat(" ", max(0, available-ansi.StringWidth(text)))
		rows = append(rows, edge.Render(v)+" "+t.Success.Render(prefix)+text+pad+edge.Render(v))
	}
	rows = append(rows, edge.Render(bottomLeft+strings.Repeat(h, width-2)+bottomRight))
	return rows
}
