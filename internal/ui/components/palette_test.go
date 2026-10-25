package components

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
	"github.com/charmbracelet/x/ansi"
)

func TestPaletteFilterOrderAndSelection(t *testing.T) {
	p := NewPalette([]PaletteItem{
		{Name: "undo", Description: "doctor logs"},
		{Name: "doctor", Description: "health"},
		{Name: "docs", Description: "manual"},
		{Name: "redo", Description: "fix doctor"},
	})
	if p.Len() != 4 || p.Filter("/").Len() != 4 {
		t.Fatal("slash should list all commands")
	}
	p = p.Filter("/do")
	if got := ansi.Strip(p.View(theme.New(theme.Options{Plain: true}), 80)); strings.Index(got, "/doctor") > strings.Index(got, "/undo") || strings.Index(got, "/docs") > strings.Index(got, "/redo") {
		t.Fatalf("prefix matches should precede contained matches: %q", got)
	}
	p, _ = p.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	if selected, _ := p.Selected(); selected.Name != "docs" {
		t.Fatalf("selected %q, want docs", selected.Name)
	}
	p = p.Filter("/DOC")
	if selected, _ := p.Selected(); selected.Name != "docs" {
		t.Fatalf("filter lost selected item: %q", selected.Name)
	}
	p = p.Filter("/und")
	if selected, _ := p.Selected(); selected.Name != "undo" {
		t.Fatalf("missing selection should reset to first: %q", selected.Name)
	}
	if p.Filter("/not-found").Len() != 0 {
		t.Fatal("filter returned unmatched commands")
	}
}

func TestPaletteNavigationAndClick(t *testing.T) {
	items := make([]PaletteItem, 12)
	for i := range items {
		items[i] = PaletteItem{Name: fmt.Sprintf("item%02d", i)}
	}
	p := NewPalette(items).MaxRows(5).Place(3, 7)
	view := ansi.Strip(p.View(theme.New(theme.Options{Plain: true}), 30))
	if p.Height(30) != 6 || !strings.Contains(view, "/item04") || strings.Contains(view, "/item05") || !strings.Contains(view, "... mais 7") {
		t.Fatalf("visible items: %q", view)
	}
	p, _ = p.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	if selected, _ := p.Selected(); selected.Name != "item04" {
		t.Fatalf("up crossed visible items: %q", selected.Name)
	}
	for range 9 {
		p, _ = p.Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	}
	if selected, _ := p.Selected(); selected.Name != "item03" {
		t.Fatalf("wheel crossed visible items: %q", selected.Name)
	}
	p, _ = p.Update(tea.MouseMotionMsg{X: 4, Y: 8})
	if p.hover != 1 {
		t.Fatalf("hover row = %d", p.hover)
	}
	_, cmd := p.Update(tea.MouseClickMsg{X: 4, Y: 8, Button: tea.MouseLeft})
	if cmd == nil || cmd().(PaletteChooseMsg).Name != "item01" {
		t.Fatal("click should choose visible row")
	}
	_, cmd = p.Update(tea.MouseClickMsg{X: 4, Y: 12, Button: tea.MouseLeft})
	if cmd != nil {
		t.Fatal("overflow footer should not choose")
	}
	p = p.Filter("/item11")
	if p.Len() != 1 || p.Height(30) != 1 {
		t.Fatal("typing should find hidden item")
	}
	if selected, _ := p.Selected(); selected.Name != "item11" {
		t.Fatalf("filtered cursor = %q", selected.Name)
	}
	p = p.Filter("/missing")
	if p.Height(30) != 0 {
		t.Fatal("empty palette should have no rows")
	}
	_, cmd = p.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd != nil {
		t.Fatal("enter on empty palette should do nothing")
	}
}

func TestPaletteViewWidthsAndDescriptions(t *testing.T) {
	p := NewPalette([]PaletteItem{{Name: "a", Description: "some extended description"}, {Name: "longer", Description: "help"}})
	out := ansi.Strip(p.View(theme.New(theme.Options{Plain: true}), 30))
	if !strings.Contains(out, "/a       some extended") || !strings.Contains(out, "/longer  help") {
		t.Fatalf("names not aligned with descriptions: %q", out)
	}
	checkChatWidth(t, "palette", func(th theme.Theme, w int) string { return p.View(th, w) })
	if out := ansi.Strip(p.View(theme.New(theme.Options{Plain: true}), 10)); !strings.Contains(out, "...") {
		t.Fatalf("missing truncation indicator: %q", out)
	}
}
