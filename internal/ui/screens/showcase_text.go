package screens

import (
	"strings"

	"github.com/charmbracelet/x/ansi"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

func ShowcaseText(t theme.Theme, width int) string {
	var b strings.Builder
	for i, p := range showcasePages() {
		live := showcaseLive{Dark: t.Dark, Width: width, Last: max(i-1, 0)}
		b.WriteString(p.Name)
		b.WriteString("\n")
		b.WriteString(showcaseBody(p, live, t, width))
		b.WriteString("\n\n")
	}
	return ansi.Strip(b.String())
}
