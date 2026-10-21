package components

import (
	"strings"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

func Banner(t theme.Theme, width int, title, subtitle string) string {
	head := t.Title.Render(title)
	if subtitle != "" {
		head += "\n" + t.Secondary.Render(subtitle)
	}
	if width <= 0 {
		return head
	}
	return t.Text.Width(width).MaxWidth(width).Render(head) + "\n" + bannerRule(t, width)
}

func bannerRule(t theme.Theme, width int) string {
	glyph := "─"
	if t.Plain {
		glyph = "-"
	}
	if !t.Rich {
		return t.Divider.Render(strings.Repeat(glyph, width))
	}
	colors := t.Blend(width)
	var b strings.Builder
	for i := range width {
		b.WriteString(t.Divider.Foreground(colors[i]).Render(glyph))
	}
	return b.String()
}
