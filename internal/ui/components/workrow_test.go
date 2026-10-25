package components

import (
	"strings"
	"testing"
	"time"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
	"github.com/charmbracelet/x/ansi"
)

func TestWorkRow(t *testing.T) {
	out := ansi.Strip(WorkRow(theme.New(theme.Options{Plain: true}), 20, "*", "working", 5*time.Second, "cancel hint that does not fit"))
	if strings.Contains(out, "cancel hint") || ansi.StringWidth(out) > 20 {
		t.Fatal(out)
	}
	checkChatWidth(t, "work row", func(th theme.Theme, w int) string {
		return WorkRow(th, w, "*", "working progress", 83*time.Second, "ctrl+c sair")
	})
}

func TestWorkRowShimmerKeepsTextAndWidth(t *testing.T) {
	rich := theme.New(theme.Options{Dark: true})
	plain := theme.New(theme.Options{Plain: true})
	label := "Procurando o Claude Code…"
	first := WorkRow(rich, 65, "⠹", label, 0, "")
	second := WorkRow(rich, 65, "⠹", label, 420*time.Millisecond, "")
	if first == second || ansi.Strip(first) != ansi.Strip(second) || ansi.StringWidth(first) != ansi.StringWidth(second) || !strings.Contains(ansi.Strip(first), label) {
		t.Fatalf("brilho deve mudar só a cor: %q, %q", first, second)
	}
	if a, b := WorkRow(plain, 65, "⠹", label, 0, ""), WorkRow(plain, 65, "⠹", label, 420*time.Millisecond, ""); strings.Split(a, " - ")[0] != strings.Split(b, " - ")[0] {
		t.Fatal("Plain sem brilho")
	}
}
