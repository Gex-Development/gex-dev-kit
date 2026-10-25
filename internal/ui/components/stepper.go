package components

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/x/ansi"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

func Stepper(t theme.Theme, width int, labels []string, current int) string {
	if width <= 0 || len(labels) == 0 {
		return ""
	}
	sep := t.Faint.Render(" › ")
	if t.Plain {
		sep = " > "
	}
	parts := make([]string, len(labels))
	for i, label := range labels {
		text := strconv.Itoa(i+1) + " " + label
		switch {
		case i == current:
			parts[i] = t.BarFocus.Padding(0, 1).Render(text)
		case i < current:
			parts[i] = t.Success.Render("✓ " + label)
			if t.Plain {
				parts[i] = t.Success.Render("[x] " + label)
			}
		default:
			parts[i] = t.Faint.Render(text)
		}
	}
	return ansi.Truncate(strings.Join(parts, sep), width, "…")
}
