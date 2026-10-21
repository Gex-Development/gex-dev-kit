package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/kit"
)

func invoke(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	var out, errors bytes.Buffer
	code := run(args, strings.NewReader(""), &out, &errors)
	return code, out.String(), errors.String()
}

func TestInstallDoctorRollbackOnExplicitHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("PATH", "")
	code, out, errors := invoke(t, "install", "--home", home, "--skip-agent-update", "--yes")
	if code != 0 || errors != "" || !strings.Contains(out, "Instalado") {
		t.Fatalf("install: exit=%d out=%q err=%q", code, out, errors)
	}
	claude := filepath.Join(home, ".claude", "CLAUDE.md")
	if _, err := os.Stat(claude); err != nil {
		t.Fatalf("Core não instalado: %v", err)
	}
	code, out, errors = invoke(t, "doctor", "--home", home)
	if code != 1 || errors != "" || !strings.Contains(out, "CLAUDE.md") || !strings.Contains(out, "bloqueado") {
		t.Fatalf("doctor: exit=%d out=%q err=%q", code, out, errors)
	}
	code, out, errors = invoke(t, "rollback", "--home", home)
	if code != 0 || errors != "" || !strings.Contains(out, "Restaurado") {
		t.Fatalf("rollback: exit=%d out=%q err=%q", code, out, errors)
	}
	if _, err := os.Stat(claude); !os.IsNotExist(err) {
		t.Fatalf("rollback deveria remover Core criado; stat=%v", err)
	}
}

func TestInstallDoesNotApplyWithoutConfirmation(t *testing.T) {
	home := t.TempDir()
	code, out, errors := invoke(t, "install", "--home", home, "--skip-agent-update")
	if code != 1 || errors == "" || !strings.Contains(out, "Prévia") {
		t.Fatalf("sem confirmação: exit=%d out=%q err=%q", code, out, errors)
	}
	if _, err := os.Stat(filepath.Join(home, ".claude", "CLAUDE.md")); !os.IsNotExist(err) {
		t.Fatalf("install sem confirmação alterou Core: %v", err)
	}
}

func TestNoArgsNonTTYShowsHelp(t *testing.T) {
	code, out, errors := invoke(t)
	if code != 0 || errors != "" || !strings.Contains(out, "install") {
		t.Fatalf("ajuda: exit=%d out=%q err=%q", code, out, errors)
	}
}

func TestHookBlocksInvalidCommitWithoutLeakingInput(t *testing.T) {
	message := "mensagem confidencial fora do padrão"
	var out, errors bytes.Buffer
	code := run([]string{"hook", "commit-msg"}, strings.NewReader(message), &out, &errors)
	if code != 2 || strings.Contains(errors.String(), message) || out.Len() != 0 {
		t.Fatalf("hook: exit=%d out=%q err=%q", code, out.String(), errors.String())
	}
}

func TestUnknownSkillCannotWrite(t *testing.T) {
	home := t.TempDir()
	code, _, errors := invoke(t, "install", "--home", home, "--skill", "inexistente", "--skip-agent-update", "--yes")
	if code == 0 || errors == "" {
		t.Fatalf("skill desconhecida: exit=%d err=%q", code, errors)
	}
	if _, err := os.Stat(filepath.Join(home, ".claude", "CLAUDE.md")); !os.IsNotExist(err) {
		t.Fatalf("skill desconhecida alterou Core: %v", err)
	}
}

func TestHookReadsGitMessageFromPath(t *testing.T) {
	file := filepath.Join(t.TempDir(), "COMMIT_EDITMSG")
	if err := os.WriteFile(file, []byte("feat(cli): liberar instalação\n"), 0600); err != nil {
		t.Fatal(err)
	}
	code, out, errors := invoke(t, "hook", "commit-msg", file)
	if code != 0 || out != "" || errors != "" {
		t.Fatalf("mensagem Git: exit=%d out=%q err=%q", code, out, errors)
	}
}

func TestDoctorReportsDriftAsFailure(t *testing.T) {
	home := t.TempDir()
	if code, _, errors := invoke(t, "install", "--home", home, "--skip-agent-update", "--yes"); code != 0 {
		t.Fatalf("instalação: %s", errors)
	}
	claude := filepath.Join(home, ".claude", "CLAUDE.md")
	if err := os.WriteFile(claude, []byte("alterado"), 0600); err != nil {
		t.Fatal(err)
	}
	code, out, errors := invoke(t, "doctor", "--home", home)
	if code != 1 || errors != "" || !strings.Contains(out, "alterado") {
		t.Fatalf("drift: exit=%d out=%q err=%q", code, out, errors)
	}
}

func TestDoctorTableFitsNarrowTerminal(t *testing.T) {
	model := newDoctorModel([]kit.Status{{Path: ".claude/CLAUDE.md", State: "alterado", Detail: "Arquivo foi alterado."}})
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 40, Height: 9})
	screen := updated.(doctorModel)
	columns := screen.table.Columns()
	if screen.table.Width() > 40 || columns[0].Width+columns[1].Width+4 > 40 || screen.table.Height() > 6 {
		t.Fatalf("tabela excedeu terminal: width=%d cols=%v height=%d", screen.table.Width(), columns, screen.table.Height())
	}
	if screen.table.Rows()[0][1] != ".claude/CLAUDE.md" {
		t.Fatalf("caminho desapareceu da tabela: %v", screen.table.Rows())
	}
}

func TestBlockedAgentUpdateStillInstallsCoreButReportsFailure(t *testing.T) {
	t.Setenv("PATH", "")
	home := t.TempDir()
	code, out, errors := invoke(t, "install", "--home", home, "--yes")
	if code != 1 || !strings.Contains(out, "bloqueado") || !strings.Contains(out, "Instalado") || errors == "" {
		t.Fatalf("bloqueio de agente: exit=%d out=%q err=%q", code, out, errors)
	}
	if _, err := os.Stat(filepath.Join(home, ".claude", "CLAUDE.md")); err != nil {
		t.Fatalf("Core deveria instalar apesar de bloqueio: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".local", "state", "gex-dev-kit", "agent-updates.json")); !os.IsNotExist(err) {
		t.Fatalf("histórico de atualização não deveria ser criado: %v", err)
	}
}

func TestRollbackAgentsRequiresConfirmationWhenNonTTY(t *testing.T) {
	home := t.TempDir()
	code, _, errors := invoke(t, "rollback-agents", "--home", home)
	if code != 1 || !strings.Contains(errors, "--yes") {
		t.Fatalf("rollback sem confirmação: exit=%d err=%q", code, errors)
	}
}

func TestInstallUnchangedCoreNeedsNoConfirmation(t *testing.T) {
	home := t.TempDir()
	if code, _, errors := invoke(t, "install", "--home", home, "--skip-agent-update", "--yes"); code != 0 {
		t.Fatalf("instalação: %s", errors)
	}
	code, out, errors := invoke(t, "install", "--home", home, "--skip-agent-update")
	if code != 0 || errors != "" || !strings.Contains(out, "Nada a alterar.") {
		t.Fatalf("instalação idempotente: exit=%d out=%q err=%q", code, out, errors)
	}
}

func TestSubcommandHelpWritesOnlyStdout(t *testing.T) {
	for _, args := range [][]string{
		{"install", "--help"}, {"doctor", "-h"}, {"rollback", "--help"},
		{"rollback-agents", "--help"}, {"skills", "--help"},
		{"hooks", "-h"}, {"hook", "commit-msg", "--help"},
	} {
		code, out, errors := invoke(t, args...)
		if code != 0 || errors != "" || !strings.Contains(out, args[0]) {
			t.Fatalf("%v: exit=%d out=%q err=%q", args, code, out, errors)
		}
	}
}

func TestVersionShortFlag(t *testing.T) {
	code, out, errors := invoke(t, "-v")
	if code != 0 || strings.TrimSpace(out) != version || errors != "" {
		t.Fatalf("-v: exit=%d out=%q err=%q", code, out, errors)
	}
}

func TestRollbackNoopDoesNotClaimRestoration(t *testing.T) {
	home := t.TempDir()
	code, out, errors := invoke(t, "rollback", "--home", home)
	if code != 0 || errors != "" || strings.Contains(out, "Restaurado") {
		t.Fatalf("rollback vazio: exit=%d out=%q err=%q", code, out, errors)
	}
}
