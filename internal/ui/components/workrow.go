package components

import (
	"math"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
	"github.com/charmbracelet/x/ansi"
)

func WorkRow(t theme.Theme, width int, spinner, label string, elapsed time.Duration, cancelHint string) string {
	if width <= 0 {
		return ""
	}
	if t.Plain {
		spinner = "*"
	}
	separator := " · "
	if t.Plain {
		separator = " - "
	}
	rendered := t.Accent.Render(label)
	if !t.Plain {
		runes := []rune(label)
		pos := int(elapsed/(70*time.Millisecond))%(len(runes)+12) - 6
		colors := lipgloss.Blend1D(4, t.Accent.GetForeground(), t.Glow.GetForeground())
		var glow strings.Builder
		for i, r := range runes {
			level := max(0, 3-int(math.Abs(float64(i-pos))))
			glow.WriteString(t.Accent.Foreground(colors[level]).Render(string(r)))
		}
		rendered = glow.String()
	}
	left := spinner + " " + rendered + t.Faint.Render(separator+FormatElapsed(elapsed))
	left = ansi.Truncate(left, width, "")
	if cancelHint == "" {
		return left
	}
	available := width - ansi.StringWidth(left) - ansi.StringWidth(cancelHint)
	if available < 1 {
		return left
	}
	return left + strings.Repeat(" ", available) + t.Faint.Render(cancelHint)
}
