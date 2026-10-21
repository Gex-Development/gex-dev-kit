package components

import "github.com/Gex-Development/gex-dev-kit/internal/ui/theme"

func StatusBadge(t theme.Theme, s theme.State, label string) string {
	return t.StateStyle(s).Render(label)
}
