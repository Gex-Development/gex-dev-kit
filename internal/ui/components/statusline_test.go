package components

import (
	"strings"
	"testing"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
)

func TestStatusLineWidthsAndASCII(t *testing.T) {
	segments := []StatusSegment{{Icon: "◆", Text: "GEX", Tone: ToneBrand}, {Icon: "●", Text: "Claude Code 2.1.306", Tone: ToneHarness, Brand: "claude"}, {Icon: "⌂", Text: "~/very/long/project/directory", Tone: ToneDir}, {Icon: "\ue0a0", Text: "feature/main *3 ?2", Tone: ToneGit}}
	metrics := []StatusMetric{{Text: "31%", Tone: ToneBrand}, {Text: "Core 0/7", Tone: ToneMuted}}
	for _, plain := range []bool{false, true} {
		th := theme.New(theme.Options{Dark: true, Profile: colorprofile.TrueColor, Plain: plain})
		for width := 10; width <= 120; width++ {
			out := StatusLine(th, width, segments, metrics)
			if got := ansi.StringWidth(out); got != width {
				t.Fatalf("plain=%v width=%d got=%d: %q", plain, width, got, ansi.Strip(out))
			}
			if strings.Contains(out, "\n") {
				t.Fatalf("linha quebrada: %q", out)
			}
			if plain {
				for _, r := range ansi.Strip(out) {
					if r >= 128 {
						t.Fatalf("Plain não ASCII: %q", ansi.Strip(out))
					}
				}
			}
		}
	}
}

func TestStatusLinePlainSanitizesUnicodeValues(t *testing.T) {
	th := theme.New(theme.Options{Dark: true, Profile: colorprofile.ASCII, Plain: true})
	out := ansi.Strip(StatusLine(th, 40, []StatusSegment{{Text: "pasta/café", Tone: ToneDir}}, []StatusMetric{{Text: "Etapa 1/6", Tone: ToneMuted}}))
	for _, r := range out {
		if r >= 128 {
			t.Fatalf("Plain contém Unicode: %q", out)
		}
	}
}

func TestStatusLinePlainGroupsGitCounts(t *testing.T) {
	th := theme.New(theme.Options{Dark: true, Profile: colorprofile.ASCII, Plain: true})
	out := ansi.Strip(StatusLine(th, 65, []StatusSegment{{Text: "GEX", Tone: ToneBrand}, {Text: "~", Tone: ToneDir}, {Text: "main", Tone: ToneGit, Modified: 3, Untracked: 2}}, nil))
	for _, part := range []string{"[GEX]", "[~]", "[main *3 ?2]"} {
		if !strings.Contains(out, part) {
			t.Fatalf("agrupamento plain %s: %q", part, out)
		}
	}
}

func TestStatusLinePriorityAndTruncation(t *testing.T) {
	th := theme.New(theme.Options{Dark: true, Profile: colorprofile.TrueColor})
	segments := []StatusSegment{{Text: "GEX", Tone: ToneBrand}, {Text: "~/very/long/project", Tone: ToneDir}, {Text: "branch", Tone: ToneGit}}
	out := ansi.Strip(StatusLine(th, 21, segments, nil))
	if !strings.Contains(out, "GEX") || !strings.Contains(out, "…") || !strings.Contains(out, "project") || strings.Contains(out, "branch") {
		t.Fatalf("prioridade ou corte esquerdo: %q", out)
	}
	withMetrics := ansi.Strip(StatusLine(th, 25, []StatusSegment{{Text: "GEX", Tone: ToneBrand}}, []StatusMetric{{Text: "31%", Tone: ToneBrand}, {Text: "Core 0/7", Tone: ToneMuted}}))
	if !strings.Contains(withMetrics, "Core 0/7") || strings.Contains(withMetrics, "31%") {
		t.Fatalf("descarte métricas da esquerda: %q", withMetrics)
	}
}

func TestStatusLinePowerlineSeparators(t *testing.T) {
	th := theme.New(theme.Options{Dark: true, Profile: colorprofile.TrueColor})
	parts := []StatusSegment{{Text: "GEX"}, {Text: "harness"}, {Icon: "\ue0a0", Text: "home"}}
	rich := ansi.Strip(StatusLine(th, 60, parts, nil))
	if got := strings.Count(rich, "\ue0b1"); got != len(parts)-1 {
		t.Fatalf("separadores=%d: %q", got, rich)
	}
	if !strings.Contains(rich, "\ue0b0") {
		t.Fatalf("fim ausente: %q", rich)
	}
	th.Powerline = false
	noPowerline := ansi.Strip(StatusLine(th, 60, parts, nil))
	if strings.ContainsAny(noPowerline, "\ue0a0\ue0b0\ue0b1") || !strings.Contains(noPowerline, "›") {
		t.Fatalf("glifos proibidos: %q", noPowerline)
	}
}

func TestStatusLineNoPowerlineIconsAndDirectoryLimit(t *testing.T) {
	th := theme.New(theme.Options{Dark: true, Profile: colorprofile.TrueColor, NoPowerline: true})
	segments := []StatusSegment{{Icon: "◆", Text: "GEX", Tone: ToneBrand}, {Icon: "\uf07c", Text: "a-very-long-directory-name-that-exceeds-limit", Tone: ToneDir}, {Icon: "\uf418", Text: "main", Tone: ToneGit}, {Icon: "\uf015", Text: "isolated", Tone: ToneWarning}}
	line := ansi.Strip(StatusLine(th, 100, segments, nil))
	if strings.ContainsAny(line, "\ue0b0\ue0b1\uf07c\uf418\uf015") || !strings.Contains(line, "▸ …") || !strings.Contains(line, "⎇ main") || !strings.Contains(line, "⌂ isolated") {
		t.Fatal(line)
	}
	if strings.Count(line, "›") != len(segments)-1 {
		t.Fatal(line)
	}
}
