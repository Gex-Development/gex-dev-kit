package components

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
	"github.com/charmbracelet/x/ansi"
)

func TestFormatElapsed(t *testing.T) {
	for _, tc := range []struct {
		seconds int
		want    string
	}{{5, "5s"}, {83, "1m 23s"}, {3840, "1h 4m"}} {
		if got := FormatElapsed(time.Duration(tc.seconds) * time.Second); got != tc.want {
			t.Fatalf("%d: %q", tc.seconds, got)
		}
	}
}

func TestToolBlock(t *testing.T) {
	for _, plain := range []bool{false, true} {
		th := theme.New(theme.Options{Dark: true, Plain: plain})
		b := ToolBlock{Title: "Verificar ambiente", Params: "~", State: theme.StateOK, Elapsed: 4 * time.Second, Rows: []ToolRow{
			{State: theme.StateOK, Label: "HOME", Detail: "~"},
			{State: theme.StateWarn, Label: "Claude Code", Detail: "não instalado"},
			{State: theme.StateInfo, Label: "Core", Detail: "0/7 skills"},
		}}
		out := ansi.Strip(b.View(th, 72))
		lines := strings.Split(out, "\n")
		if len(lines) != 5 || !strings.Contains(lines[0], "Verificar ambiente ~") || !strings.Contains(lines[4], " 4s ") {
			t.Fatalf("moldura e metadados: %q", out)
		}
		if !plain && (!strings.HasPrefix(lines[0], "╭─") || !strings.HasSuffix(lines[0], "╮") || !strings.HasPrefix(lines[4], "╰") || !strings.HasSuffix(lines[4], "╯")) {
			t.Fatal(out)
		}
		if plain {
			if !strings.HasPrefix(lines[0], "+-") || !strings.HasPrefix(lines[4], "+-") {
				t.Fatal(out)
			}
			for _, r := range out {
				if r > 127 {
					t.Fatalf("Plain não ASCII: %q", out)
				}
			}
		}
		for _, line := range lines {
			if ansi.StringWidth(line) != 72 {
				t.Fatalf("largura irregular %d: %q", ansi.StringWidth(line), line)
			}
		}
		if got := ansi.StringWidth(strings.Split(ansi.Strip(b.View(th, 120)), "\n")[0]); got != 100 {
			t.Fatalf("cartão deve limitar largura a 100: %d", got)
		}
		detail := "não instalado"
		if plain {
			detail = plainASCII(detail)
		}
		starts := []int{ansi.StringWidth(lines[1][:strings.Index(lines[1], "~")]), ansi.StringWidth(lines[2][:strings.Index(lines[2], detail)]), ansi.StringWidth(lines[3][:strings.Index(lines[3], "0/7 skills")])}
		if starts[0] != starts[1] || starts[1] != starts[2] {
			t.Fatalf("detalhes desalinhados: %q", out)
		}
		for i := range 10 {
			b.Lines = append(b.Lines, fmt.Sprintf("line%d", i))
		}
		out = ansi.Strip(b.View(th, 45))
		if !strings.Contains(out, "(7 linhas ocultas)") || strings.Contains(out, "line9") {
			t.Fatal(out)
		}
		b.Expanded = true
		if !strings.Contains(ansi.Strip(b.View(th, 45)), "line9") {
			t.Fatal("expanded missing line")
		}
		checkChatWidth(t, "tool block mixed", func(th theme.Theme, w int) string { return b.View(th, w) })
		b.Rows = nil
		if text := ansi.Strip(b.View(th, 45)); !strings.Contains(text, "| line0") && !strings.Contains(text, "│ line0") {
			t.Fatal(text)
		}
		checkChatWidth(t, "tool block", func(th theme.Theme, w int) string { return b.View(th, w) })
	}
}

func TestToolBlockElapsedYieldsToTitle(t *testing.T) {
	th := theme.New(theme.Options{Plain: true})
	b := ToolBlock{Title: "Verificar ambiente", Params: "~", State: theme.StateOK, Elapsed: 4 * time.Second}
	if got := ansi.Strip(b.View(th, 22)); strings.Contains(got, "4s") || !strings.Contains(got, "Verificar") {
		t.Fatal(got)
	}
	if got := ansi.Strip(b.View(th, 32)); !strings.Contains(got, "4s") {
		t.Fatal(got)
	}
}

func TestToolBlockSkippedAndShortDuration(t *testing.T) {
	th := theme.New(theme.Options{Plain: true})
	b := ToolBlock{Title: "Plano", State: theme.StateOK, Elapsed: 900 * time.Millisecond}
	if got := ansi.Strip(b.View(th, 30)); strings.Contains(got, "0s") {
		t.Fatal(got)
	}
	b.State = theme.StateSkipped
	if got := ansi.Strip(b.View(th, 30)); !strings.Contains(got, "cancelado") {
		t.Fatal(got)
	}
}

func TestToolBlockNarrowAndChecklist(t *testing.T) {
	th := theme.New(theme.Options{Dark: true})
	b := ToolBlock{Title: "Verificar ambiente", State: theme.StateRunning, Elapsed: 1500 * time.Millisecond, Spinner: "⠹", Phases: []string{"Checando sua pasta home…", "Procurando o Claude Code…", "Procurando o Codex…", "Lendo o Core instalado…"}, PhaseRows: true, Rows: []ToolRow{{State: theme.StateOK, Label: "HOME", Detail: "~"}}}
	out := ansi.Strip(b.View(th, 90))
	if !strings.Contains(out, "✓ HOME") || !strings.Contains(out, "⠹ Procurando o Claude Code…") || !strings.Contains(out, "○ Procurando o Codex") || strings.Contains(out, "Checando sua pasta home") {
		t.Fatal(out)
	}
	b.PhaseRows = false
	out = ansi.Strip(b.View(th, 90))
	if !strings.Contains(out, "✓ Checando sua pasta home") || strings.Contains(out, "HOME") {
		t.Fatal(out)
	}
	if got := ansi.Strip(b.View(th, 19)); strings.ContainsAny(got, "╭╮╰╯") {
		t.Fatalf("coluna estreita sem moldura: %q", got)
	}
}
