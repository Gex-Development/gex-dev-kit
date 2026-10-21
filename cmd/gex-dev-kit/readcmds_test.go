package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCatalogCommandsDistinguishInstalledFromReferences(t *testing.T) {
	code, skills, errors := invoke(t, "skills")
	if code != 0 || errors != "" || !strings.Contains(skills, "Core GEX (obrigatório):") || !strings.Contains(skills, "Referências (não instaláveis):") || strings.Contains(skills, "\x1b") {
		t.Fatalf("skills: exit=%d out=%q err=%q", code, skills, errors)
	}
	code, hooks, errors := invoke(t, "hooks")
	if code != 0 || errors != "" || !strings.Contains(hooks, "gex-conventional-commits") || !strings.Contains(hooks, "claude/smart-commit [somente referência]") || !strings.Contains(hooks, "codex/secret-scanner [somente referência]") || strings.Contains(hooks, "\x1b") {
		t.Fatalf("hooks: exit=%d out=%q err=%q", code, hooks, errors)
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
