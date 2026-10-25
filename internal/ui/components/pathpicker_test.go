package components

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
	"github.com/charmbracelet/x/ansi"
)

func TestPathSuggestionsDirectoriesAndMarkers(t *testing.T) {
	home := t.TempDir()
	for _, dir := range []string{"alpha", "alpine", "Beta", ".hidden", "alpha/.git"} {
		if err := os.MkdirAll(filepath.Join(home, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(home, "alfile"), []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(home, "alpha"), filepath.Join(home, "alias")); err != nil {
		t.Fatal(err)
	}
	got := PathSuggestions("~/al", home, []string{".git"}, 10)
	want := []PathSuggestion{{"alias", "~/alias/", true}, {"alpha", "~/alpha/", true}, {"alpine", "~/alpine/", false}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("suggestions = %#v, want %#v", got, want)
	}
	got = PathSuggestions("~/b", home, nil, 10)
	if len(got) != 1 || got[0].Name != "Beta" {
		t.Fatalf("case-insensitive prefix = %#v", got)
	}
	got = PathSuggestions("~/.", home, nil, 10)
	if len(got) != 1 || got[0].Name != ".hidden" {
		t.Fatalf("hidden prefix = %#v", got)
	}
	got = PathSuggestions("~/", home, nil, 2)
	if len(got) != 2 || got[0].Name != "Beta" || got[1].Name != "alias" {
		t.Fatalf("directory listing or limit = %#v", got)
	}
	if got := PathSuggestions("~", home, nil, 10); len(got) != 4 || got[0].Path != "~/Beta/" {
		t.Fatalf("bare tilde should list home: %#v", got)
	}
	if got := PathSuggestions(filepath.Join(home, "missing")+"/", home, nil, 5); len(got) != 0 {
		t.Fatalf("missing directory = %#v", got)
	}
	if got := PathSuggestions(filepath.Join(home, "alfile")+"/", home, nil, 5); len(got) != 0 {
		t.Fatalf("unreadable directory = %#v", got)
	}
	if got := PathSuggestions("~/", home, nil, 0); len(got) != 0 {
		t.Fatalf("zero limit = %#v", got)
	}
}

func TestExpandHomeAndPathCompletion(t *testing.T) {
	home := t.TempDir()
	for _, name := range []string{"project", "progress", "other"} {
		if err := os.Mkdir(filepath.Join(home, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if ExpandHome("~", home) != home || ExpandHome("~/project", home) != filepath.Join(home, "project") || ExpandHome("~other", home) != "~other" {
		t.Fatal("home expansion changed non-home path or lost home")
	}
	p := NewPathPicker(home, nil).Sync("~/p")
	if p.Len() != 2 || p.Complete("~/p") != "~/pro" {
		t.Fatalf("common prefix should extend name without slash: %#v, %q", p, p.Complete("~/p"))
	}
	p = p.Sync("~/proj")
	if p.Complete("~/proj") != "~/project/" {
		t.Fatalf("unique completion = %q", p.Complete("~/proj"))
	}
	p = p.Sync("~/missing")
	if p.Complete("~/missing") != "~/missing" {
		t.Fatal("empty completion should preserve input")
	}
	p = p.Sync("~/")
	p, _ = p.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if highlighted, ok := p.Highlighted(); !ok || p.Complete("~/") != highlighted.Path {
		t.Fatal("highlighted item should take precedence over common prefix")
	}
}

func TestPathPickerNavigationMouseAndView(t *testing.T) {
	home := t.TempDir()
	for i := range 30 {
		if err := os.Mkdir(filepath.Join(home, fmt.Sprintf("folder%02d", i)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	p := NewPathPicker(home, nil).Sync("~/").MaxRows(5).Place(4, 9)
	view := ansi.Strip(p.View(theme.New(theme.Options{Plain: true}), 40))
	if p.Height(40) != 6 || !strings.Contains(view, "folder04/") || strings.Contains(view, "folder05/") || !strings.Contains(view, "... mais 25") {
		t.Fatalf("visible suggestions: %q", view)
	}
	p, _ = p.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	if selected, _ := p.Highlighted(); selected.Name != "folder04" {
		t.Fatalf("up from none: %q", selected.Name)
	}
	p, _ = p.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if selected, _ := p.Highlighted(); selected.Name != "folder00" {
		t.Fatalf("down wrap: %q", selected.Name)
	}
	for range 17 {
		p, _ = p.Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	}
	if selected, _ := p.Highlighted(); selected.Name != "folder02" {
		t.Fatalf("wheel crossed visible suggestions: %q", selected.Name)
	}
	_, cmd := p.Update(tea.MouseClickMsg{X: 5, Y: 10, Button: tea.MouseLeft})
	if cmd == nil || cmd().(PathPickMsg).Path != "~/folder01/" {
		t.Fatal("click must pick visible second folder")
	}
	_, cmd = p.Update(tea.MouseClickMsg{X: 5, Y: 14, Button: tea.MouseLeft})
	if cmd != nil {
		t.Fatal("overflow footer should not be clickable")
	}
	p = p.Sync("~/folder29")
	if p.Len() != 1 || p.Complete("~/folder29") != "~/folder29/" {
		t.Fatal("prefix should access folder outside initial suggestions")
	}
	p = p.Sync("~/none")
	if p.Height(40) != 1 || !strings.Contains(ansi.Strip(p.View(theme.New(theme.Options{Plain: true}), 40)), "(nenhuma pasta aqui)") {
		t.Fatal("empty list should show placeholder")
	}
	checkChatWidth(t, "path picker", func(th theme.Theme, w int) string {
		return NewPathPicker(home, nil).Sync("~/").MaxRows(5).View(th, w)
	})
}
