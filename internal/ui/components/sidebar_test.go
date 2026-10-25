package components

import (
	"strings"
	"testing"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

func TestSidebar(t *testing.T) {
	sections := []SidebarSection{{Title: "Environment", Items: []ChecklistItem{{Label: "Core", State: theme.StateOK, Detail: "ready"}}}, {Title: "Steps", Items: []ChecklistItem{{Label: "Install", State: theme.StatePending}}}}
	th := theme.New(theme.Options{Plain: true})
	got := Sidebar(th, 30, 4, sections)
	if len(strings.Split(got, "\n")) != 4 {
		t.Fatalf("wanted 4 rows: %q", got)
	}
	checkChatWidth(t, "sidebar", func(th theme.Theme, w int) string { return Sidebar(th, w, 8, sections) })
}

func TestSidebarDrawsStateSymbolOnce(t *testing.T) {
	th := theme.New(theme.Options{Plain: true})
	got := Sidebar(th, 40, 12, []SidebarSection{{Title: "Env", Items: []ChecklistItem{{Label: "Core", State: theme.StateOK}}}})
	if strings.Count(got, "[ok]") != 1 {
		t.Fatalf("state symbol drawn %d times: %q", strings.Count(got, "[ok]"), got)
	}
}
