package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Gex-Development/gex-dev-kit/internal/app"
	"github.com/Gex-Development/gex-dev-kit/internal/githooks"
	"github.com/Gex-Development/gex-dev-kit/internal/kit"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/components"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/terminal"
)

func writeCall(t *testing.T, o options, fn func(session) int) (int, string, string) {
	t.Helper()
	return writeCallWith(t, o, realDeps(), fn)
}

func writeCallWith(t *testing.T, o options, d deps, fn func(session) int) (int, string, string) {
	t.Helper()
	var out, errOut bytes.Buffer
	env := terminal.Detect(strings.NewReader(""), &out, os.Environ())
	s := session{in: strings.NewReader(""), out: &out, errOut: &errOut, env: env, opts: o, deps: d, guard: newGuard(nil)}
	code := fn(s)
	return code, out.String(), errOut.String()
}

func TestInstallDoesNotApplyWithoutConfirmation(t *testing.T) {
	home := t.TempDir()
	code, out, errors := writeCall(t, options{Home: home, SkipAgentUpdate: true}, cmdInstall)
	if code != 1 || errors == "" || !strings.Contains(out, "Prévia de instalação:") {
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
			code, out, _ := writeCall(t, options{Home: home, SkipAgentUpdate: true}, cmdInstall)
			if code != 1 || !strings.Contains(out, "Prévia de instalação:") || strings.Contains(out, "\x1b") {
				t.Fatalf("prévia sem cor: exit=%d out=%q", code, out)
			}
		})
	}
}

func TestUnknownSkillCannotWrite(t *testing.T) {
	home := t.TempDir()
	code, _, errors := writeCall(t, options{Home: home, Skills: []string{"inexistente"}, SkipAgentUpdate: true, Yes: true}, cmdInstall)
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
	code, out, errors := writeCall(t, options{Home: home, Yes: true}, cmdInstall)
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
	code, _, errors := writeCall(t, options{Home: home}, cmdRollbackAgents)
	if code != 1 || !strings.Contains(errors, "--yes") {
		t.Fatalf("rollback sem confirmação: exit=%d err=%q", code, errors)
	}
}

func TestInstallUnchangedCoreNeedsNoConfirmation(t *testing.T) {
	home := t.TempDir()
	if code, _, errors := writeCall(t, options{Home: home, SkipAgentUpdate: true, Yes: true}, cmdInstall); code != 0 {
		t.Fatalf("instalação: %s", errors)
	}
	code, out, errors := writeCall(t, options{Home: home, SkipAgentUpdate: true}, cmdInstall)
	if code != 0 || errors != "" || !strings.Contains(out, "Nada a alterar.") {
		t.Fatalf("instalação idempotente: exit=%d out=%q err=%q", code, out, errors)
	}
}

func TestRollbackNoopDoesNotClaimRestoration(t *testing.T) {
	home := t.TempDir()
	code, out, errors := writeCall(t, options{Home: home}, cmdRollback)
	if code != 0 || errors != "" || strings.Contains(out, "Restaurado") {
		t.Fatalf("rollback vazio: exit=%d out=%q err=%q", code, out, errors)
	}
}

func TestRollbackRepoRestauraGitHooks(t *testing.T) {
	cliIsolate(t)
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git indisponível")
	}
	repo := t.TempDir()
	hookGit(t, repo, "init", "-q")
	commitMsg := filepath.Join(repo, ".git", "hooks", "commit-msg")
	if code, _, errOut := invoke(t, "githooks", "--repo", repo, "--hook", "gex-commit-msg", "--yes"); code != 0 {
		t.Fatalf("githooks falhou: exit=%d err=%q", code, errOut)
	}
	if _, err := os.Stat(commitMsg); err != nil {
		t.Fatalf("commit-msg não foi instalado: %v", err)
	}
	code, out, errOut := invoke(t, "rollback", "--repo", repo, "--yes")
	if code != 0 || errOut != "" || !strings.Contains(out, "Restaurado") {
		t.Fatalf("rollback --repo: exit=%d out=%q err=%q", code, out, errOut)
	}
	if _, err := os.Stat(commitMsg); !os.IsNotExist(err) {
		t.Fatalf("commit-msg deveria ter sido removido pelo rollback: %v", err)
	}
}

func TestGatewaySkillNeedsEnvironmentBeforeInstalling(t *testing.T) {
	t.Setenv("PAGAMERICAN_API_KEY", "")
	t.Setenv("BUYGOODS_ACCOUNT_ID", "")
	t.Setenv("BUYGOODS_API_TOKEN", "")
	t.Setenv("CLICKBANK_API_KEY", "")
	t.Setenv("CLICKBANK_ACCOUNT", "")
	for _, skill := range []string{"gex-pagamerican-api", "gex-buygoods-api", "gex-clickbank-api"} {
		home := t.TempDir()
		code, _, errors := writeCall(t, options{Home: home, Skills: []string{skill}, SkipAgentUpdate: true, Yes: true}, cmdInstall)
		if code == 0 || !strings.Contains(errors, skill) || !strings.Contains(errors, "canal corporativo") {
			t.Fatalf("%s sem credencial: exit=%d err=%q", skill, code, errors)
		}
		if _, err := os.Stat(filepath.Join(home, ".claude", "CLAUDE.md")); !os.IsNotExist(err) {
			t.Fatalf("%s alterou HOME sem credencial: %v", skill, err)
		}
	}
}

func TestInstallDoctorRollbackOnExplicitHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("PATH", "")
	code, out, errors := writeCall(t, options{Home: home, SkipAgentUpdate: true, Yes: true}, cmdInstall)
	if code != 0 || errors != "[..] Atualizando agentes\n[..] Gravando Core\n[ok] Core gravado\n" || !strings.Contains(out, "Instalado") {
		t.Fatalf("install: exit=%d out=%q err=%q", code, out, errors)
	}
	claude := filepath.Join(home, ".claude", "CLAUDE.md")
	if _, err := os.Stat(claude); err != nil {
		t.Fatalf("Core não instalado: %v", err)
	}
	code, out, errors = writeCall(t, options{Home: home, Yes: true}, cmdRollback)
	if code != 0 || errors != "" || !strings.Contains(out, "Restaurado") {
		t.Fatalf("rollback: exit=%d out=%q err=%q", code, out, errors)
	}
	if _, err := os.Stat(claude); !os.IsNotExist(err) {
		t.Fatalf("rollback deveria remover Core criado; stat=%v", err)
	}
}

func TestInstallYesNonTTYWritesHomeAndKeepsStdoutResult(t *testing.T) {
	home := t.TempDir()
	code, out, errors := writeCall(t, options{Home: home, SkipAgentUpdate: true, Yes: true}, cmdInstall)
	if code != 0 || !strings.Contains(out, "Prévia de instalação:") || !strings.Contains(out, "Instalado. Histórico:") {
		t.Fatalf("--yes: exit=%d out=%q err=%q", code, out, errors)
	}
	if strings.Contains(errors, "\x1b") || strings.Contains(out, "\x1b") {
		t.Fatalf("--yes sem TTY com ANSI: out=%q err=%q", out, errors)
	}
	if _, err := os.Stat(filepath.Join(home, ".claude", "CLAUDE.md")); err != nil {
		t.Fatalf("Core não gravado: %v", err)
	}
}

func TestInstallQuietKeepsPreviewAndResult(t *testing.T) {
	home := t.TempDir()
	code, out, _ := writeCall(t, options{Home: home, SkipAgentUpdate: true, Yes: true, Quiet: true}, cmdInstall)
	if code != 0 || !strings.Contains(out, "Prévia de instalação:") || !strings.Contains(out, "Instalado") {
		t.Fatalf("--quiet: exit=%d out=%q", code, out)
	}
}

func TestRollbackWithoutYesNonTTYKeepsHome(t *testing.T) {
	home := t.TempDir()
	if code, _, errors := writeCall(t, options{Home: home, SkipAgentUpdate: true, Yes: true}, cmdInstall); code != 0 {
		t.Fatalf("instalação: %s", errors)
	}
	code, _, errors := writeCall(t, options{Home: home}, cmdRollback)
	if code != 1 || !strings.Contains(errors, "--yes") {
		t.Fatalf("rollback sem --yes: exit=%d err=%q", code, errors)
	}
	if _, err := os.Stat(filepath.Join(home, ".claude", "CLAUDE.md")); err != nil {
		t.Fatalf("rollback sem --yes alterou HOME: %v", err)
	}
}

func TestRollbackConflictAbortsBeforeConfirmation(t *testing.T) {
	d := realDeps()
	d.PlanRollback = func(string, int) (app.RollbackPreview, error) {
		return app.RollbackPreview{Plan: kit.RollbackPlan{ID: "x", Conflicts: []string{"CLAUDE.md"}}}, nil
	}
	code, out, errors := writeCallWith(t, options{Home: t.TempDir()}, d, cmdRollback)
	if code != 1 || !strings.Contains(errors, "CLAUDE.md") || !strings.Contains(errors, "não alterou nada") || strings.Contains(out, "Restaurado") {
		t.Fatalf("conflito: exit=%d out=%q err=%q", code, out, errors)
	}
}

func TestGitHooksWithoutRepoPrintsScriptsWithoutWriting(t *testing.T) {
	home := t.TempDir()
	code, out, errors := writeCall(t, options{Home: home, Hooks: []string{"gex-commit-msg"}}, cmdGitHooks)
	if code != 0 || errors != "" || !strings.Contains(out, "hooks/commit-msg") {
		t.Fatalf("githooks sem repo: exit=%d out=%q err=%q", code, out, errors)
	}
	if _, err := os.Stat(filepath.Join(home, ".local")); !os.IsNotExist(err) {
		t.Fatalf("githooks sem repo gravou arquivos: %v", err)
	}
}

func TestGitHooksUndocumentedPlanFails(t *testing.T) {
	d := realDeps()
	d.PlanGitHooks = func(app.GitHooksRequest, int) (app.GitHooksPlan, error) {
		return app.GitHooksPlan{}, githooks.ErrUndocumented
	}
	code, _, errors := writeCallWith(t, options{Home: t.TempDir(), Repo: t.TempDir(), Hooks: []string{"gex-pre-push"}, PrePush: []string{"comando-inexistente"}}, d, cmdGitHooks)
	if code != 1 || !strings.Contains(errors, "documentado") {
		t.Fatalf("comando não documentado: exit=%d err=%q", code, errors)
	}
}

func TestWriteActionMapeiaAcoes(t *testing.T) {
	cases := map[string]components.ChangeAction{
		"criar":      components.ChangeCreate,
		"atualizar":  components.ChangeUpdate,
		"remover":    components.ChangeRemove,
		"restaurar":  components.ChangeRestore,
		"inalterado": components.ChangeUnchanged,
		"outra":      components.ChangeUpdate,
	}
	for in, want := range cases {
		if got := writeAction(in); got != want {
			t.Fatalf("ação %q: got=%d want=%d", in, got, want)
		}
	}
}

func TestWriteGroupsConvertsLinhas(t *testing.T) {
	groups := writeGroups([]app.FileDiff{{Path: "a.md", Action: "criar", Lines: []app.DiffLine{{Kind: '+', Text: "x"}, {Kind: '-', Text: "y"}, {Kind: '@', Text: "z"}, {Kind: ' ', Text: "w"}}}})
	if len(groups) != 1 || len(groups[0].Changes) != 1 || groups[0].Changes[0].Path != "a.md" {
		t.Fatalf("grupos: %+v", groups)
	}
	lines := groups[0].Changes[0].Lines
	if len(lines) != 4 || lines[0].Kind != components.DiffAdded || lines[1].Kind != components.DiffRemoved || lines[2].Kind != components.DiffHunk || lines[3].Kind != components.DiffContext {
		t.Fatalf("linhas convertidas: %+v", lines)
	}
	if writeGroups(nil) != nil {
		t.Fatalf("sem diffs deveria devolver nil")
	}
}

func TestWriteHomeAbsolutoOuPadrao(t *testing.T) {
	if home, err := writeHome(""); err != nil || home == "" {
		t.Fatalf("HOME padrão: home=%q err=%v", home, err)
	}
	if _, err := writeHome("relativo"); err == nil {
		t.Fatalf("caminho relativo deveria falhar")
	}
}

func TestWriteFailMostraCausaEAcao(t *testing.T) {
	var buf bytes.Buffer
	code := writeFail(&buf, errors.New("falha de teste"))
	if code != 1 || !strings.Contains(buf.String(), "falha de teste") {
		t.Fatalf("writeFail: exit=%d err=%q", code, buf.String())
	}
}

func TestCmdSemLipglossNemHuh(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") || file == "accessible.go" {
			continue
		}
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(data), "\"charm.land/lipgloss") || strings.Contains(string(data), "\"charm.land/huh") {
			t.Fatalf("%s importa lipgloss ou huh; a CLI usa só os pacotes de internal/ui", file)
		}
	}
}
