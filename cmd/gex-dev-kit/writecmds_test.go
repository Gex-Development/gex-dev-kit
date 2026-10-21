package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

func TestInstallPreviewHasNoANSIWithoutColor(t *testing.T) {
	for _, env := range []string{"NO_COLOR", "TERM"} {
		t.Run(env, func(t *testing.T) {
			if env == "NO_COLOR" {
				t.Setenv(env, "1")
			} else {
				t.Setenv(env, "dumb")
			}
			home := t.TempDir()
			code, out, _ := invoke(t, "install", "--home", home, "--skip-agent-update")
			if code != 1 || !strings.Contains(out, "Prévia de instalação:") || strings.Contains(out, "\x1b") {
				t.Fatalf("prévia sem cor: exit=%d out=%q", code, out)
			}
		})
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

func TestRollbackNoopDoesNotClaimRestoration(t *testing.T) {
	home := t.TempDir()
	code, out, errors := invoke(t, "rollback", "--home", home)
	if code != 0 || errors != "" || strings.Contains(out, "Restaurado") {
		t.Fatalf("rollback vazio: exit=%d out=%q err=%q", code, out, errors)
	}
}

func TestGatewaySkillNeedsEnvironmentBeforeInstalling(t *testing.T) {
	t.Setenv("PAGAMERICAN_API_KEY", "")
	t.Setenv("BUYGOODS_ACCOUNT_ID", "")
	t.Setenv("BUYGOODS_API_TOKEN", "")
	t.Setenv("CLICKBANK_API_KEY", "")
	t.Setenv("CLICKBANK_ACCOUNT", "")
	for _, tc := range []struct{ skill, missing string }{
		{"gex-pagamerican-api", "PAGAMERICAN_API_KEY"},
		{"gex-buygoods-api", "BUYGOODS_ACCOUNT_ID"},
		{"gex-clickbank-api", "CLICKBANK_API_KEY"},
	} {
		home := t.TempDir()
		code, _, errors := invoke(t, "install", "--home", home, "--skip-agent-update", "--yes", "--skill", tc.skill)
		if code == 0 || !strings.Contains(errors, tc.missing) || !strings.Contains(errors, "canal corporativo") {
			t.Fatalf("%s sem credencial: exit=%d err=%q", tc.skill, code, errors)
		}
		if _, err := os.Stat(filepath.Join(home, ".claude", "CLAUDE.md")); !os.IsNotExist(err) {
			t.Fatalf("%s alterou HOME sem credencial: %v", tc.skill, err)
		}
	}
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
