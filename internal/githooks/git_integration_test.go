package githooks

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Gex-Development/gex-dev-kit/internal/hooks"
	"github.com/Gex-Development/gex-dev-kit/internal/kit"
)

func integracaoRodarGit(dir string, args ...string) (string, error) {
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).Output()
	return string(out), err
}

func integracaoGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %s falhou: %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}

func integracaoTentarGit(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
	return string(out), err
}

func integracaoEscrever(t *testing.T, caminho, conteudo string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(caminho), 0755); err != nil {
		t.Fatalf("MkdirAll() erro: %v", err)
	}
	if err := os.WriteFile(caminho, []byte(conteudo), 0644); err != nil {
		t.Fatalf("WriteFile() erro: %v", err)
	}
}

func integracaoCompilarHookbin(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "hookbin")
	out, err := exec.Command("go", "build", "-o", bin, "./testdata/hookbin").CombinedOutput()
	if err != nil {
		t.Fatalf("go build do hookbin falhou: %v\n%s", err, out)
	}
	return bin
}

func integracaoIsolar(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(home, "gitconfig"))
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_EDITOR", "true")
	t.Setenv("GIT_AUTHOR_NAME", "teste")
	t.Setenv("GIT_AUTHOR_EMAIL", "teste")
	t.Setenv("GIT_COMMITTER_NAME", "teste")
	t.Setenv("GIT_COMMITTER_EMAIL", "teste")
}

func integracaoPreparar(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git ausente")
	}
	if runtime.GOOS == "windows" {
		t.Skip("hooks POSIX: integração roda só em sistemas POSIX")
	}
	bin := integracaoCompilarHookbin(t)
	integracaoIsolar(t)
	return bin
}

func integracaoNovoRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	integracaoGit(t, dir, "init", "-q")
	return dir
}

func integracaoRequisicao(bin string, preCommit, prePush []string) Request {
	return Request{
		Hooks:     []string{"commit-msg", "pre-commit", "pre-push"},
		PreCommit: preCommit,
		PrePush:   prePush,
		Binary:    bin,
	}
}

func integracaoInstalar(t *testing.T, dir string, r Request) string {
	t.Helper()
	repo, err := Resolve(dir, integracaoRodarGit)
	if err != nil {
		t.Fatalf("Resolve() erro: %v", err)
	}
	files, err := Files(r)
	if err != nil {
		t.Fatalf("Files() erro: %v", err)
	}
	plan, err := kit.Preview(repo.CommonDir, files)
	if err != nil {
		t.Fatalf("kit.Preview() erro: %v", err)
	}
	if _, err := kit.Apply(plan); err != nil {
		t.Fatalf("kit.Apply() erro: %v", err)
	}
	return repo.CommonDir
}

func integracaoCommitar(t *testing.T, dir, nome, conteudo, mensagem string) (string, error) {
	t.Helper()
	integracaoEscrever(t, filepath.Join(dir, nome), conteudo)
	integracaoGit(t, dir, "add", nome)
	return integracaoTentarGit(t, dir, "commit", "-q", "-m", mensagem)
}

func TestGitRealResolve(t *testing.T) {
	integracaoPreparar(t)
	dir := integracaoNovoRepo(t)
	repo, err := Resolve(dir, integracaoRodarGit)
	if err != nil {
		t.Fatalf("Resolve() em repositório comum erro: %v", err)
	}
	raiz, _ := filepath.EvalSymlinks(dir)
	if repo.Root != raiz || repo.CommonDir != filepath.Join(raiz, ".git") {
		t.Fatalf("Resolve() = %+v, queria raiz %q e .git", repo, raiz)
	}

	bare := filepath.Join(t.TempDir(), "remoto.git")
	integracaoGit(t, t.TempDir(), "init", "-q", "--bare", bare)
	if _, err := Resolve(bare, integracaoRodarGit); err != ErrBare {
		t.Fatalf("Resolve(bare) erro = %v, queria ErrBare", err)
	}

	integracaoGit(t, dir, "config", "core.hooksPath", "/exemplo/hooks")
	if _, err := Resolve(dir, integracaoRodarGit); err != ErrHooksPath {
		t.Fatalf("Resolve(hooksPath) erro = %v, queria ErrHooksPath", err)
	}

	if _, err := Resolve(t.TempDir(), integracaoRodarGit); err != ErrNotRepo {
		t.Fatalf("Resolve(fora de repositório) erro = %v, queria ErrNotRepo", err)
	}
}

func TestGitRealTemplateComentadoAceito(t *testing.T) {
	bin := integracaoPreparar(t)
	dir := integracaoNovoRepo(t)
	integracaoInstalar(t, dir, integracaoRequisicao(bin, nil, nil))
	integracaoEscrever(t, filepath.Join(dir, "a.txt"), "um\n")
	integracaoGit(t, dir, "add", "a.txt")
	msg := filepath.Join(t.TempDir(), "msg.txt")
	integracaoEscrever(t, msg, "feat: adiciona arquivo a\n\n# Linhas de comentário\n# On branch main\n")
	if out, err := integracaoTentarGit(t, dir, "commit", "-q", "-F", msg); err != nil {
		t.Fatalf("commit com template comentado foi bloqueado: %v\n%s", err, out)
	}
	if got := strings.TrimSpace(integracaoGit(t, dir, "log", "-1", "--format=%s")); got != "feat: adiciona arquivo a" {
		t.Fatalf("assunto do commit = %q", got)
	}
}

func TestGitRealFixupESquashAceitos(t *testing.T) {
	bin := integracaoPreparar(t)
	if hooks.CheckGitMessage(hooks.CleanGitMessage("fixup! feat: base\n", "#")) != nil {
		t.Skip("aceite de fixup! depende de internal/hooks (W1-T3), ausente nesta worktree")
	}
	dir := integracaoNovoRepo(t)
	integracaoInstalar(t, dir, integracaoRequisicao(bin, nil, nil))
	if out, err := integracaoCommitar(t, dir, "a.txt", "um\n", "feat: base"); err != nil {
		t.Fatalf("commit base bloqueado: %v\n%s", err, out)
	}
	integracaoEscrever(t, filepath.Join(dir, "a.txt"), "dois\n")
	integracaoGit(t, dir, "add", "a.txt")
	if out, err := integracaoTentarGit(t, dir, "commit", "-q", "--fixup=HEAD"); err != nil {
		t.Fatalf("--fixup bloqueado: %v\n%s", err, out)
	}
	integracaoEscrever(t, filepath.Join(dir, "a.txt"), "tres\n")
	integracaoGit(t, dir, "add", "a.txt")
	if out, err := integracaoTentarGit(t, dir, "commit", "-q", "--squash=HEAD"); err != nil {
		t.Fatalf("--squash bloqueado: %v\n%s", err, out)
	}
}

func TestGitRealMensagemLivreBloqueada(t *testing.T) {
	bin := integracaoPreparar(t)
	dir := integracaoNovoRepo(t)
	integracaoInstalar(t, dir, integracaoRequisicao(bin, nil, nil))
	if out, err := integracaoCommitar(t, dir, "a.txt", "um\n", "feat: base"); err != nil {
		t.Fatalf("commit base bloqueado: %v\n%s", err, out)
	}
	out, err := integracaoCommitar(t, dir, "b.txt", "dois\n", "mensagem livre")
	if err == nil {
		t.Fatalf("mensagem livre foi aceita\n%s", out)
	}
	if n := strings.TrimSpace(integracaoGit(t, dir, "rev-list", "--count", "HEAD")); n != "1" {
		t.Fatalf("commits = %s, queria 1", n)
	}
}

func TestGitRealCredencialBloqueadaSemEcoELiberadaPelaExcecao(t *testing.T) {
	bin := integracaoPreparar(t)
	dir := integracaoNovoRepo(t)
	integracaoInstalar(t, dir, integracaoRequisicao(bin, nil, nil))
	if out, err := integracaoCommitar(t, dir, "a.txt", "um\n", "feat: base"); err != nil {
		t.Fatalf("commit base bloqueado: %v\n%s", err, out)
	}
	chave := "AKIA" + "ABCDEFGHIJKLMNOP"
	out, err := integracaoCommitar(t, dir, "config.txt", "aws = "+chave+"\n", "feat: config")
	if err == nil {
		t.Fatalf("credencial foi aceita\n%s", out)
	}
	if !strings.Contains(out, "config.txt:1") || !strings.Contains(out, "aws-access-key") {
		t.Fatalf("saída sem o achado: %s", out)
	}
	if strings.Contains(out, chave) {
		t.Fatalf("saída ecoou o valor da credencial: %s", out)
	}
	commonDir := filepath.Join(dir, ".git")
	integracaoEscrever(t, hooks.AllowPath(commonDir), "config.txt\taws-access-key\n")
	if out, err := integracaoTentarGit(t, dir, "commit", "-q", "-m", "feat: config"); err != nil {
		t.Fatalf("exceção não liberou o commit: %v\n%s", err, out)
	}
}

func TestGitRealWorktreeLigadoUsaExcecaoDoDiretorioComum(t *testing.T) {
	bin := integracaoPreparar(t)
	dir := integracaoNovoRepo(t)
	commonDir := integracaoInstalar(t, dir, integracaoRequisicao(bin, nil, nil))
	if out, err := integracaoCommitar(t, dir, "a.txt", "um\n", "feat: base"); err != nil {
		t.Fatalf("commit base bloqueado: %v\n%s", err, out)
	}
	wt := filepath.Join(t.TempDir(), "ligado")
	integracaoGit(t, dir, "worktree", "add", "-q", "-b", "ligado", wt)
	repo, err := Resolve(wt, integracaoRodarGit)
	if err != nil {
		t.Fatalf("Resolve(worktree) erro: %v", err)
	}
	if want, _ := filepath.EvalSymlinks(commonDir); repo.CommonDir != want {
		t.Fatalf("CommonDir do worktree = %q, queria %q", repo.CommonDir, want)
	}
	if out, err := integracaoCommitar(t, wt, "config.txt", "aws = "+"AKIA"+"ABCDEFGHIJKLMNOP"+"\n", "feat: config"); err == nil {
		t.Fatalf("credencial no worktree foi aceita\n%s", out)
	}
	integracaoEscrever(t, hooks.AllowPath(commonDir), "config.txt\taws-access-key\n")
	if out, err := integracaoTentarGit(t, wt, "commit", "-q", "-m", "feat: config"); err != nil {
		t.Fatalf("exceção não liberou o worktree: %v\n%s", err, out)
	}
}

func TestGitRealPrePushComandoDecideOPush(t *testing.T) {
	bin := integracaoPreparar(t)
	casos := []struct {
		comando string
		passa   bool
	}{
		{"true", true},
		{"false", false},
	}
	for _, c := range casos {
		t.Run(c.comando, func(t *testing.T) {
			dir := integracaoNovoRepo(t)
			integracaoInstalar(t, dir, integracaoRequisicao(bin, nil, []string{c.comando}))
			if out, err := integracaoCommitar(t, dir, "a.txt", "um\n", "feat: base"); err != nil {
				t.Fatalf("commit base bloqueado: %v\n%s", err, out)
			}
			remoto := filepath.Join(t.TempDir(), "remoto.git")
			integracaoGit(t, t.TempDir(), "init", "-q", "--bare", remoto)
			out, err := integracaoTentarGit(t, dir, "push", "-q", remoto, "HEAD:refs/heads/main")
			if c.passa && err != nil {
				t.Fatalf("push com %q falhou: %v\n%s", c.comando, err, out)
			}
			if !c.passa && err == nil {
				t.Fatalf("push com %q passou\n%s", c.comando, out)
			}
		})
	}
}

func TestGitRealRollbackRemoveHooks(t *testing.T) {
	bin := integracaoPreparar(t)
	dir := integracaoNovoRepo(t)
	commonDir := integracaoInstalar(t, dir, integracaoRequisicao(bin, nil, nil))
	if _, err := kit.Rollback(commonDir); err != nil {
		t.Fatalf("kit.Rollback() erro: %v", err)
	}
	for _, nome := range []string{"commit-msg", "pre-commit", "pre-push"} {
		if _, err := os.Stat(filepath.Join(commonDir, "hooks", nome)); !os.IsNotExist(err) {
			t.Fatalf("hook %s continuou existindo após rollback (erro=%v)", nome, err)
		}
	}
}
