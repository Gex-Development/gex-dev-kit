package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func TestHomeSelectsDoctorWithoutInstalling(t *testing.T) {
	home := t.TempDir()
	model := newHomeModel(home)
	updated, _ := model.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	updated, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if updated.(homeModel).choice != "doctor" {
		t.Fatalf("segunda ação deveria abrir doctor, recebeu %q", updated.(homeModel).choice)
	}
	if _, err := os.Stat(filepath.Join(home, ".claude", "CLAUDE.md")); !os.IsNotExist(err) {
		t.Fatalf("navegar não pode instalar Core: %v", err)
	}
}

func TestHomeRendersWithinNarrowTerminal(t *testing.T) {
	model := newHomeModel(t.TempDir())
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 40, Height: 18})
	if width := lipgloss.Width(updated.View().Content); width > 40 {
		t.Fatalf("painel excedeu terminal de 40 colunas: %d", width)
	}
}

func TestHomeKeepsActionsWithinCompactTerminal(t *testing.T) {
	for _, size := range []tea.WindowSizeMsg{{Width: 40, Height: 8}, {Width: 14, Height: 8}} {
		model := newHomeModel(t.TempDir())
		updated, _ := model.Update(size)
		screen := updated.View().Content
		if lipgloss.Width(screen) > size.Width || lipgloss.Height(screen) > size.Height {
			t.Fatalf("painel ultrapassou %dx%d: %dx%d", size.Width, size.Height, lipgloss.Width(screen), lipgloss.Height(screen))
		}
		if !strings.Contains(screen, "Doctor") {
			t.Fatalf("ação de diagnóstico desapareceu do terminal %dx%d", size.Width, size.Height)
		}
	}
}

func TestHomeReportsChangedCoreFromDisk(t *testing.T) {
	home := t.TempDir()
	if code, _, output := invoke(t, "install", "--home", home, "--skip-agent-update", "--yes"); code != 0 {
		t.Fatalf("instalação isolada: %s", output)
	}
	if err := os.WriteFile(filepath.Join(home, ".claude", "CLAUDE.md"), []byte("mudança local"), 0600); err != nil {
		t.Fatal(err)
	}
	status := scanHome(home)()
	updated, _ := newHomeModel(home).Update(status)
	if !strings.Contains(updated.View().Content, "alterado") {
		t.Fatalf("painel ocultou drift no Core: %s", updated.View().Content)
	}
}
