package components

import (
	"strings"
	"testing"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
	"github.com/charmbracelet/x/ansi"
)

func checkChatWidth(t *testing.T, name string, render func(theme.Theme, int) string) {
	t.Helper()
	for _, plain := range []bool{false, true} {
		th := theme.New(theme.Options{Plain: plain})
		for w := 10; w <= 100; w++ {
			for _, line := range strings.Split(render(th, w), "\n") {
				if ansi.StringWidth(line) > w {
					t.Fatalf("%s width %d: %q", name, w, line)
				}
				if plain && !isASCII(ansi.Strip(line)) {
					t.Fatalf("%s plain non-ASCII: %q", name, line)
				}
			}
		}
	}
}

func isASCII(s string) bool {
	for _, c := range s {
		if c > 127 {
			return false
		}
	}
	return true
}

func TestChatMessage(t *testing.T) {
	th := theme.New(theme.Options{Plain: true})
	for _, tc := range []struct {
		role   ChatRole
		prefix string
	}{{ChatAssistant, "* "}, {ChatUser, "| "}, {ChatNote, "  "}, {ChatError, "x "}} {
		out := ChatMessage(th, 15, tc.role, "first line second line with wrapping")
		if !strings.HasPrefix(ansi.Strip(out), tc.prefix) {
			t.Fatalf("role %d: %q", tc.role, out)
		}
		checkChatWidth(t, "chat message", func(th theme.Theme, w int) string {
			return ChatMessage(th, w, tc.role, "first line second line with wrapping")
		})
	}
}
