package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
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

func TestGuidedPreviewScrollsAndNeverAppliesOnCancel(t *testing.T) {
	home := t.TempDir()
	model := newInstallModel(home, "", nil, nil, true)
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 40, Height: 14})
	changes := make([]kit.Change, 30)
	for i := range changes {
		changes[i] = kit.Change{Action: "criar", Path: fmt.Sprintf(".claude/skills/exemplo-%02d/SKILL.md", i)}
	}
	updated, _ = updated.Update(installPrepared{plan: &kit.Plan{Changes: changes}})
	screen := updated.(installModel)
	if width, height := lipgloss.Width(screen.View().Content), lipgloss.Height(screen.View().Content); width > 40 || height > 14 {
		t.Fatalf("prévia excede 40x14: %dx%d", width, height)
	}
	updated, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyPgDown})
	scrolled := updated.(installModel)
	if scrolled.viewport.YOffset() == 0 {
		t.Fatal("prévia longa não rolou")
	}
	updated, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if updated.(installModel).phase != installCancelled {
		t.Fatal("Esc deveria cancelar sem aplicação")
	}
	if _, err := os.Stat(filepath.Join(home, ".claude", "CLAUDE.md")); !os.IsNotExist(err) {
		t.Fatalf("cancelamento alterou HOME: %v", err)
	}
}

func TestGuidedPlanningErrorStopsBeforeConfirmation(t *testing.T) {
	model := newInstallModel(t.TempDir(), "", nil, nil, true)
	updated, _ := model.Update(installPrepared{err: os.ErrPermission})
	screen := updated.(installModel)
	if screen.phase != installFailed || !strings.Contains(screen.View().Content, "permission denied") {
		t.Fatalf("falha da prévia ocultada: %q", screen.View().Content)
	}
}

func TestGuidedConfirmationRequiresExplicitYesAndSurfacesApplyError(t *testing.T) {
	model := newInstallModel(t.TempDir(), "", nil, nil, true)
	updated, _ := model.Update(installPrepared{plan: &kit.Plan{Changes: []kit.Change{{Path: ".claude/CLAUDE.md", Action: "criar"}}}})
	updated, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if screen := updated.(installModel); screen.phase != installConfirm || !strings.Contains(screen.View().Content, "APLICAR ALTERAÇÕES") {
		t.Fatalf("confirmação não está focada: %q", screen.View().Content)
	}
	updated, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if updated.(installModel).phase != installConfirm {
		t.Fatal("Enter sozinho não pode aplicar mudanças")
	}
	updated, _ = updated.Update(tea.KeyPressMsg{Code: 's', Text: "s"})
	result := updated.(installModel).apply().(installApplied)
	if result.err == nil {
		t.Fatal("falha da aplicação não foi propagada")
	}
	updated, _ = updated.Update(result)
	if updated.(installModel).phase != installFailed || !strings.Contains(updated.View().Content, "Falha:") {
		t.Fatalf("falha da aplicação não ficou visível: %q", updated.View().Content)
	}
}
