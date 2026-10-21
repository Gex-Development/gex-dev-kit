package main

import (
	"bytes"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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

func TestAccessibleHomeOffersActionsAndCanExit(t *testing.T) {
	for _, choice := range []string{"q", "0", "cancel"} {
		var out bytes.Buffer
		selected, err := showAccessibleHome(strings.NewReader(choice+"\n"), &out)
		if err != nil || selected != "exit" || !strings.Contains(out.String(), "Doctor") || !strings.Contains(out.String(), "Instalar Core") || strings.Contains(out.String(), "\x1b") {
			t.Fatalf("escolha %q: ação=%q err=%v menu=%q", choice, selected, err, out.String())
		}
	}
	var out bytes.Buffer
	selected, err := showAccessibleHome(strings.NewReader("2\n"), &out)
	if err != nil || selected != "doctor" {
		t.Fatalf("ação Doctor: ação=%q err=%v menu=%q", selected, err, out.String())
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

func TestDoctorWrapsLongDetailInNarrowTerminal(t *testing.T) {
	model := newDoctorModel([]kit.Status{{Path: ".claude/CLAUDE.md", State: "alterado", Detail: strings.Repeat("texto longo ", 12)}})
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 40, Height: 16})
	if width := lipgloss.Width(updated.View().Content); width > 40 {
		t.Fatalf("detalhe do doctor excedeu 40 colunas: %d", width)
	}
}

func TestDoctorShortListDoesNotFillScreenWithEmptyRows(t *testing.T) {
	model := newDoctorModel([]kit.Status{{Path: ".claude/CLAUDE.md", State: "ok", Detail: "Íntegro."}})
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	if height := lipgloss.Height(updated.View().Content); height > 10 {
		t.Fatalf("uma linha ocupou tela com espaço vazio: %d linhas", height)
	}
}

func TestDoctorFitsTinyTerminal(t *testing.T) {
	for _, size := range []tea.WindowSizeMsg{{Width: 14, Height: 6}, {Width: 40, Height: 8}} {
		model := newDoctorModel([]kit.Status{{Path: ".claude/CLAUDE.md", State: "alterado", Detail: strings.Repeat("informação importante ", 6)}})
		updated, _ := model.Update(size)
		screen := updated.View().Content
		if lipgloss.Width(screen) > size.Width || lipgloss.Height(screen) > size.Height {
			t.Fatalf("doctor ultrapassou %dx%d: %dx%d", size.Width, size.Height, lipgloss.Width(screen), lipgloss.Height(screen))
		}
	}
}

func TestTerminalAdaptsPaletteToLightBackground(t *testing.T) {
	light := tea.BackgroundColorMsg{Color: color.White}
	home := newHomeModel(t.TempDir())
	homeLight, _ := home.Update(light)
	if home.View().Content == homeLight.View().Content {
		t.Fatal("painel manteve cores de fundo escuro no terminal claro")
	}
	doctor := newDoctorModel([]kit.Status{{Path: "Core", State: "ok", Detail: "Íntegro."}})
	doctorLight, _ := doctor.Update(light)
	if doctor.View().Content == doctorLight.View().Content {
		t.Fatal("doctor manteve cores de fundo escuro no terminal claro")
	}
	if gexFormTheme(true).Focused.Title.Render("GEX") == gexFormTheme(false).Focused.Title.Render("GEX") {
		t.Fatal("formulário manteve mesmo destaque em fundos claro e escuro")
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

func TestGatewayCredentialGateDoesNotPrintSecretOrRequireNetwork(t *testing.T) {
	t.Setenv("BUYGOODS_ACCOUNT_ID", "secret-account")
	t.Setenv("BUYGOODS_API_TOKEN", "")
	err := gatewayCredentialError([]string{"gex-buygoods-api"})
	if err == nil || !strings.Contains(err.Error(), "BUYGOODS_API_TOKEN") || strings.Contains(err.Error(), "secret-account") {
		t.Fatalf("gate expôs valor ou omitiu variável faltante: %v", err)
	}
	t.Setenv("BUYGOODS_API_TOKEN", "secret-token")
	if err := gatewayCredentialError([]string{"gex-buygoods-api", "gex-local-spark-tests"}); err != nil {
		t.Fatalf("presença basta na instalação; não consultar rede: %v", err)
	}
}
