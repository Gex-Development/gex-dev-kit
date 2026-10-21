package main

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/agents"
	"github.com/Gex-Development/gex-dev-kit/internal/kit"
)

func TestPreviewKeepsBlockedAgentVisibleOnNarrowTerminal(t *testing.T) {
	view := previewPanel([]kit.Change{{Action: "criar", Path: ".claude/skills/very-long-skill-name/SKILL.md"}}, []agents.Change{{Name: "claude", Target: "2.1.306", Blocked: true, Reason: "outro gestor de pacotes", Manual: "confira versão atual"}}, 40)
	if width := lipgloss.Width(view); width > 40 {
		t.Fatalf("prévia excedeu 40 colunas: %d", width)
	}
	if !strings.Contains(view, "bloqueado") || !strings.Contains(view, "Ação manual") {
		t.Fatalf("prévia ocultou bloqueio e orientação: %s", view)
	}
}
