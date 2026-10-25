package githooks

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func pathsFakeGit(answers map[string]string) func(dir string, args ...string) (string, error) {
	return func(dir string, args ...string) (string, error) {
		out, ok := answers[strings.Join(args, " ")]
		if !ok {
			return "", errors.New("fatal: " + strings.Join(args, " "))
		}
		return out, nil
	}
}

func pathsRepoAnswers() map[string]string {
	return map[string]string{
		"rev-parse --is-bare-repository":                    "false\n",
		"rev-parse --show-toplevel":                         "C:/x\n",
		"rev-parse --path-format=absolute --git-common-dir": " C:/x/.git \n",
	}
}

func TestResolveDevolveRaizEDiretorioComum(t *testing.T) {
	repo, err := Resolve("C:/x", pathsFakeGit(pathsRepoAnswers()))
	if err != nil {
		t.Fatalf("Resolve() erro inesperado: %v", err)
	}
	if want := filepath.FromSlash("C:/x"); repo.Root != want {
		t.Fatalf("Root = %q, queria %q", repo.Root, want)
	}
	if want := filepath.FromSlash("C:/x/.git"); repo.CommonDir != want {
		t.Fatalf("CommonDir = %q, queria %q", repo.CommonDir, want)
	}
}

func TestResolveRecusaCasosSentinela(t *testing.T) {
	bare := pathsRepoAnswers()
	bare["rev-parse --is-bare-repository"] = "true\n"
	hooksPath := pathsRepoAnswers()
	hooksPath["config --get core.hooksPath"] = "/x/hooks\n"
	noToplevel := pathsRepoAnswers()
	delete(noToplevel, "rev-parse --show-toplevel")
	cases := []struct {
		nome    string
		answers map[string]string
		want    error
	}{
		{"fora de repositório", map[string]string{}, ErrNotRepo},
		{"bare", bare, ErrBare},
		{"core.hooksPath definido", hooksPath, ErrHooksPath},
		{"sem árvore de trabalho", noToplevel, ErrNotRepo},
	}
	for _, c := range cases {
		t.Run(c.nome, func(t *testing.T) {
			_, err := Resolve("C:/x", pathsFakeGit(c.answers))
			if !errors.Is(err, c.want) {
				t.Fatalf("erro = %v, queria %v", err, c.want)
			}
		})
	}
}

func TestResolveHooksPathVazioNaoBloqueia(t *testing.T) {
	answers := pathsRepoAnswers()
	answers["config --get core.hooksPath"] = "  \n"
	if _, err := Resolve("C:/x", pathsFakeGit(answers)); err != nil {
		t.Fatalf("Resolve() com hooksPath vazio falhou: %v", err)
	}
}

func TestNormalizeGitPath(t *testing.T) {
	cases := map[string]string{
		"C:/x/.git":                     filepath.FromSlash("C:/x/.git"),
		"C:/x/.git\r\n":                 filepath.FromSlash("C:/x/.git"),
		"  /exemplo/Minha Pasta/.git\n": filepath.FromSlash("/exemplo/Minha Pasta/.git"),
	}
	for entrada, want := range cases {
		if got := NormalizeGitPath(entrada); got != want {
			t.Fatalf("NormalizeGitPath(%q) = %q, queria %q", entrada, got, want)
		}
	}
}

func TestConflictsDetectaHookAlheioEAlterado(t *testing.T) {
	commonDir := t.TempDir()
	dir := filepath.Join(commonDir, "hooks")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("MkdirAll() erro: %v", err)
	}
	escreve := func(nome, conteudo string) {
		if err := os.WriteFile(filepath.Join(dir, nome), []byte(conteudo), 0755); err != nil {
			t.Fatalf("WriteFile(%s) erro: %v", nome, err)
		}
	}
	escreve("commit-msg", "do kit")
	escreve("pre-push", "alheio")
	escreve("pre-commit", "mudou")
	repo := Repo{Root: commonDir, CommonDir: commonDir}
	managed := map[string]string{"commit-msg": "do kit", "pre-commit": "original"}

	got, err := Conflicts(repo, []string{"commit-msg", "pre-commit", "pre-push", "post-merge"}, managed)
	if err != nil {
		t.Fatalf("Conflicts() erro inesperado: %v", err)
	}
	if want := "pre-commit,pre-push"; strings.Join(got, ",") != want {
		t.Fatalf("Conflicts() = %v, queria %s", got, want)
	}
}

func TestConflictsSemRaizFalha(t *testing.T) {
	if _, err := Conflicts(Repo{}, []string{"commit-msg"}, nil); !errors.Is(err, ErrNotRepo) {
		t.Fatalf("erro = %v, queria ErrNotRepo", err)
	}
}
