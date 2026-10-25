package components

import (
	"strings"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/layout"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
	"github.com/charmbracelet/x/ansi"
)

type SidebarSection struct {
	Title string
	Items []ChecklistItem
}

func Sidebar(t theme.Theme, width, height int, sections []SidebarSection) string {
	if width <= 0 || height <= 0 {
		return ""
	}
	rows := []string{Wordmark(t, width, 1), ""}
	ellipsis := "…"
	if t.Plain {
		ellipsis = "..."
	}
	for _, section := range sections {
		rows = append(rows, t.Eyebrow.Render(ansi.Truncate(strings.ToUpper(section.Title), width, ellipsis)))
		for _, item := range section.Items {
			row := t.StateStyle(item.State).Render(item.Label)
			rows = append(rows, ansi.Truncate(row, width, ellipsis))
			if item.Detail != "" {
				rows = append(rows, t.Secondary.Render(ansi.Truncate("  "+item.Detail, width, ellipsis)))
			}
		}
	}
	return layout.Fit(strings.Join(rows, "\n"), width, height)
}
