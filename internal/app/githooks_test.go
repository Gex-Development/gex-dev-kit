package app

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/Gex-Development/gex-dev-kit/internal/catalog"
	"github.com/Gex-Development/gex-dev-kit/internal/githooks"
	"github.com/Gex-Development/gex-dev-kit/internal/kit"
)

func gitHooksIsolate(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git indisponível")
	}
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(home, "gitconfig"))
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
}

func gitHooksRepo(t *testing.T) string {
	t.Helper()
	gitHooksIsolate(t)
	dir := t.TempDir()
	if out, err := exec.Command("git", "init", "-q", dir).CombinedOutput(); err != nil {
		t.Fatalf("git init falhou: %v\n%s", err, out)
	}
	writeAppTestFile(t, filepath.Join(dir, "README.md"), "Rode `make test` antes de enviar.\n")
	writeAppTestFile(t, filepath.Join(dir, "Makefile"), "test:\n\techo ok\n")
	return dir
}

func writeAppTestFile(t *testing.T, path, data string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatalf("arquivo de teste não gravado: %v", err)
	}
}

func gitHooksRequest(t *testing.T, repo string) GitHooksRequest {
	t.Helper()
	return GitHooksRequest{
		Home:       t.TempDir(),
		Repo:       repo,
		Executable: installExecutable(t),
		Hooks:      []string{"pre-push"},
		PrePush:    []string{"make test"},
	}
}

func TestPlanGitHooksRecordsDocumentedSource(t *testing.T) {
	repo := gitHooksRepo(t)
	plan, err := PlanGitHooks(gitHooksRequest(t, repo), 0)
	if err != nil {
		t.Fatalf("prévia falhou: %v", err)
	}
	if len(plan.Sources) != 1 || plan.Sources[0].File != "README.md" || plan.Sources[0].Line != 1 {
		t.Fatalf("origem do comando errada: %+v", plan.Sources)
	}
}

func TestPlanGitHooksRejectsUndocumentedCommand(t *testing.T) {
	repo := gitHooksRepo(t)
	r := gitHooksRequest(t, repo)
	r.PrePush = []string{"make deploy"}
	_, err := PlanGitHooks(r, 0)
	if !errors.Is(err, githooks.ErrUndocumented) {
		t.Fatalf("comando não documentado deveria recusar: %v", err)
	}
}

func TestPlanGitHooksRejectsHooksPath(t *testing.T) {
	repo := gitHooksRepo(t)
	if out, err := exec.Command("git", "-C", repo, "config", "core.hooksPath", ".githooks").CombinedOutput(); err != nil {
		t.Fatalf("config de teste falhou: %v\n%s", err, out)
	}
	_, err := PlanGitHooks(gitHooksRequest(t, repo), 0)
	if !errors.Is(err, githooks.ErrHooksPath) {
		t.Fatalf("core.hooksPath deveria recusar: %v", err)
	}
}

func TestPlanGitHooksWithoutRepoOnlyPreviewsScripts(t *testing.T) {
	gitHooksIsolate(t)
	r := gitHooksRequest(t, "")
	plan, err := PlanGitHooks(r, 0)
	if err != nil {
		t.Fatalf("prévia sem repositório falhou: %v", err)
	}
	if plan.Scripts["pre-push"] == "" || len(plan.RepoChanges) != 0 || len(plan.HomeChanges) != 0 {
		t.Fatalf("prévia sem repositório não deveria planejar gravação: %+v", plan)
	}
}

func TestApplyGitHooksWritesAndDoctorSeesOkThenRollbackRemoves(t *testing.T) {
	repo := gitHooksRepo(t)
	r := gitHooksRequest(t, repo)
	plan, err := PlanGitHooks(r, 0)
	if err != nil {
		t.Fatalf("prévia falhou: %v", err)
	}
	result, err := ApplyGitHooks(plan)
	if err != nil {
		t.Fatalf("gravação falhou: %v", err)
	}
	if result.HistoryID == "" {
		t.Fatal("histórico do repositório não informado")
	}
	hookFile := filepath.Join(plan.Repo.CommonDir, "hooks", "pre-push")
	info, err := os.Stat(hookFile)
	if err != nil {
		t.Fatalf("hook não gravado: %v", err)
	}
	if info.Mode().Perm()&0100 == 0 {
		t.Fatalf("hook sem permissão de execução: %v", info.Mode())
	}
	exe, err := os.ReadFile(r.Executable)
	if err != nil {
		t.Fatal(err)
	}
	stable, err := os.ReadFile(catalog.StableBinary(r.Home))
	if err != nil || string(stable) != string(exe) {
		t.Fatalf("executável do HOME não copiado: %v", err)
	}
	report, err := DoctorRepo(r.Home, repo, r.Executable)
	if err != nil {
		t.Fatalf("doctor do repositório falhou: %v", err)
	}
	if got := doctorFind(t, report.Items, "hooks/pre-push").State; got != "ok" {
		t.Fatalf("estado do hook depois da gravação: %s", got)
	}
	if _, err := kit.Rollback(plan.Repo.CommonDir); err != nil {
		t.Fatalf("rollback do repositório falhou: %v", err)
	}
	if _, err := os.Stat(hookFile); !os.IsNotExist(err) {
		t.Fatalf("rollback deveria remover o hook: %v", err)
	}
}

func TestApplyGitHooksWithoutRepoRefuses(t *testing.T) {
	gitHooksIsolate(t)
	plan, err := PlanGitHooks(gitHooksRequest(t, ""), 0)
	if err != nil {
		t.Fatalf("prévia falhou: %v", err)
	}
	if _, err := ApplyGitHooks(plan); err == nil {
		t.Fatal("prévia sem repositório não pode gravar")
	}
}

func TestPlanGitHooksAtualizaHookDoKitComNovoComando(t *testing.T) {
	repo := gitHooksRepo(t)
	writeAppTestFile(t, filepath.Join(repo, "README.md"), "Rode `true` e `false`.\n")
	r := gitHooksRequest(t, repo)
	r.Hooks = []string{"pre-commit"}
	r.PrePush = nil
	r.PreCommit = []string{"true"}
	plan, err := PlanGitHooks(r, 0)
	if err != nil {
		t.Fatalf("primeira prévia falhou: %v", err)
	}
	if _, err := ApplyGitHooks(plan); err != nil {
		t.Fatalf("primeira gravação falhou: %v", err)
	}
	r.PreCommit = []string{"true", "false"}
	plan, err = PlanGitHooks(r, 0)
	if err != nil {
		t.Fatalf("replanejar hook do kit falhou: %v", err)
	}
	if len(plan.RepoChanges) != 1 || plan.RepoChanges[0].Path != "hooks/pre-commit" {
		t.Fatalf("atualização do hook do kit não planejada: %+v", plan.RepoChanges)
	}
}

func TestApplyGitHooksDesfazHomeQuandoRepositorioFalha(t *testing.T) {
	repo := gitHooksRepo(t)
	r := gitHooksRequest(t, repo)
	plan, err := PlanGitHooks(r, 0)
	if err != nil {
		t.Fatalf("prévia falhou: %v", err)
	}
	writeAppTestFile(t, filepath.Join(plan.Repo.CommonDir, "hooks", "pre-push"), "alterado depois da prévia\n")
	_, err = ApplyGitHooks(plan)
	if !errors.Is(err, kit.ErrChanged) {
		t.Fatalf("gravação do repositório deveria falhar com ErrChanged: %v", err)
	}
	if _, err := os.Stat(catalog.StableBinary(r.Home)); !os.IsNotExist(err) {
		t.Fatalf("executável do HOME deveria voltar ao estado anterior: %v", err)
	}
}
