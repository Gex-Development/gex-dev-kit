package components

import (
	"strings"
	"testing"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
)

func TestHarnessIconRendersThreeRowsAndPlainFallback(t *testing.T) {
	rich := theme.New(theme.Options{Dark: true, Profile: colorprofile.TrueColor})
	plain := theme.New(theme.Options{Dark: false, Profile: colorprofile.ASCII, Plain: true})
	for name, fallback := range map[string]string{"claude": "[CC]", "codex": "[>_]", "omp": "[PI]"} {
		if got := HarnessIcon(plain, name); got != fallback {
			t.Fatalf("ícone ASCII %s: %q", name, got)
		}
		if got := ansi.Strip(HarnessIcon(rich, name)); len(strings.Split(got, "\n")) != 3 {
			t.Fatalf("ícone %s: %q", name, got)
		}
	}
}

func TestHarnessArtAndBannerWidth(t *testing.T) {
	rich := theme.New(theme.Options{Dark: true, Profile: colorprofile.TrueColor})
	plain := theme.New(theme.Options{Plain: true, Profile: colorprofile.ASCII})
	for _, id := range []string{"claude", "codex", "omp"} {
		for _, th := range []theme.Theme{rich, plain} {
			art := HarnessArt(th, id)
			if len(art) != 3 {
				t.Fatalf("%s art has %d rows", id, len(art))
			}
			for _, row := range art {
				if ansi.StringWidth(row) != 9 || th.Plain && !isASCII(ansi.Strip(row)) {
					t.Fatalf("%s art row: %q", id, row)
				}
			}
			if got := HarnessBanner(th, 29, id, []string{"title"}); got != strings.Join(art, "\n") {
				t.Fatalf("%s narrow banner: %q", id, got)
			}
			checkChatWidth(t, id+" banner", func(th theme.Theme, width int) string {
				return HarnessBanner(th, width, id, []string{"Claude Code v2.1.308", "Haiku 5.5 model", "/very/long/path/to/project/name"})
			})
			rows := strings.Split(ansi.Strip(HarnessBanner(th, 30, id, []string{"title", "model", "path"})), "\n")
			if len(rows) != 3 || !strings.Contains(rows[0], "  title") || !strings.Contains(rows[1], "  model") || !strings.Contains(rows[2], "  path") {
				t.Fatalf("%s banner alignment: %q", id, rows)
			}
		}
	}
	if got := ansi.Strip(HarnessArt(rich, "claude")[0]); got != " ▐▛███▜▌ " {
		t.Fatalf("claude mascot: %q", got)
	}
}

func TestHarnessCodexArtIsCloudWithPromptOnMiddleRow(t *testing.T) {
	rich := theme.New(theme.Options{Dark: true, Profile: colorprofile.TrueColor})
	plain := theme.New(theme.Options{Plain: true, Profile: colorprofile.ASCII})
	for _, th := range []theme.Theme{rich, plain} {
		rows := strings.Split(ansi.Strip(strings.Join(HarnessArt(th, "codex"), "\n")), "\n")
		if !strings.Contains(rows[1], ">_") || strings.Contains(rows[0], ">_") || strings.Contains(rows[2], ">_") {
			t.Fatalf("prompt deve ficar só na linha do meio: %q", rows)
		}
		if strings.TrimSpace(rows[0]) == "" || rows[0] == rows[1] || rows[2] == rows[1] || rows[0] == rows[2] {
			t.Fatalf("topo, meio e base devem formar uma silhueta, não um bloco: %q", rows)
		}
	}
	if got := ansi.Strip(HarnessArt(rich, "codex")[0]); strings.Trim(got, " ") == "" || strings.Contains(got, " ") {
		t.Fatalf("o topo da nuvem deve ser preenchido sem buracos de espaço: %q", got)
	}
}

func TestHarnessBadgeUsesBrandColorAndFixedWidth(t *testing.T) {
	rich := theme.New(theme.Options{Dark: true, Profile: colorprofile.TrueColor})
	plain := theme.New(theme.Options{Plain: true, Profile: colorprofile.ASCII})
	for _, tc := range []struct{ name, glyph, fallback string }{
		{"claude", "▐▛███▜▌", "[CC]   "},
		{"codex", ">_", "[>_]   "},
		{"omp", "π", "[PI]   "},
	} {
		if got := HarnessBadge(plain, tc.name); got != tc.fallback {
			t.Errorf("selo ASCII %s: %q", tc.name, got)
		}
		got := HarnessBadge(rich, tc.name)
		if ansi.StringWidth(got) != 7 || !strings.Contains(ansi.Strip(got), tc.glyph) || strings.Contains(got, "\n") || !strings.Contains(got, "\x1b[") {
			t.Errorf("selo %s: %q", tc.name, got)
		}
	}
	if got := HarnessBadge(rich, "unknown"); got != "" {
		t.Errorf("selo desconhecido: %q", got)
	}
}
