package components

import (
	"strings"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

func Section(t theme.Theme, width int, title string) string {
	if width <= 0 {
		return t.Title.Render(title)
	}
	return t.Title.Width(width).MaxWidth(width).Render(title) + "\n" + Divider(t, width)
}

func Divider(t theme.Theme, width int) string {
	if width <= 0 {
		return ""
	}
	return t.Divider.Render(strings.Repeat(sectionRule(t.Plain), width))
}

func sectionRule(plain bool) string {
	if plain {
		return "-"
	}
	return "─"
}
