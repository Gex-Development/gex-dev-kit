package components

import (
	"strings"
	"testing"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
)

func TestPromptBox(t *testing.T) {
	for _, plain := range []bool{false, true} {
		th := theme.New(theme.Options{Dark: true, Profile: colorprofile.TrueColor, Plain: plain})
		for width := 10; width <= 120; width++ {
			rows := PromptBox(th, width, []StatusSegment{{Text: "GEX", Tone: ToneBrand}}, nil, strings.Repeat("x", 150), false, "waiting")
			if len(rows) != 3 {
				t.Fatalf("linhas=%d", len(rows))
			}
			for _, row := range rows {
				if got := ansi.StringWidth(row); got != width {
					t.Fatalf("plain=%v width=%d got=%d: %q", plain, width, got, ansi.Strip(row))
				}
				if plain {
					for _, r := range ansi.Strip(row) {
						if r > 127 {
							t.Fatalf("plain Unicode: %q", ansi.Strip(row))
						}
					}
				}
			}
		}
		locked := PromptBox(th, 40, nil, nil, "secret view", true, "waiting")
		if strings.Contains(strings.Join(locked, ""), "secret view") || !strings.Contains(ansi.Strip(locked[1]), "waiting") {
			t.Fatal(locked)
		}
		area := PromptBox(th, 40, nil, nil, "first\nsecond\nthird", false, "waiting")
		if len(area) != 5 || !strings.Contains(ansi.Strip(area[2]), "second") || !strings.Contains(ansi.Strip(area[3]), "third") {
			t.Fatal(area)
		}
	}
}
