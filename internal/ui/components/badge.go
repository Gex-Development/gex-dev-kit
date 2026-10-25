package components

import "github.com/Gex-Development/gex-dev-kit/internal/ui/theme"

func StatusBadge(t theme.Theme, s theme.State, label string) string {
	style := t.StateStyle(s)
	if t.Rich {
		style = style.Reverse(true).Padding(0, 1)
	}
	return style.Render(label)
}
