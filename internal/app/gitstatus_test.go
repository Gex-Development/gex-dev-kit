package app

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGitStatusRepositoryAndChanges(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git indisponível")
	}
	dir := t.TempDir()
	if _, ok := GitStatus(dir); ok {
		t.Fatal("pasta fora de repo não deve ter git")
	}
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1", "GIT_AUTHOR_NAME=Test", "GIT_AUTHOR_EMAIL=test@example.test", "GIT_COMMITTER_NAME=Test", "GIT_COMMITTER_EMAIL=test@example.test")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %s: %v", args, out, err)
		}
	}
	run("init", "-q")
	if got, ok := GitStatus(dir); !ok || got.Modified != 0 || got.Untracked != 0 || got.Branch == "" {
		t.Fatalf("repo limpo: %+v, %v", got, ok)
	}
	tracked := filepath.Join(dir, "tracked.txt")
	if err := os.WriteFile(tracked, []byte("before"), 0o600); err != nil {
		t.Fatal(err)
	}
	run("add", "tracked.txt")
	run("commit", "-qm", "initial")
	if err := os.WriteFile(tracked, []byte("after"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "new.txt"), []byte("new"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, ok := GitStatus(dir)
	if !ok || got.Modified != 1 || got.Untracked != 1 {
		t.Fatalf("mudanças contadas errado: %+v, %v", got, ok)
	}
	run("checkout", "--detach", "-q")
	if got, ok := GitStatus(dir); !ok || got.Branch != "HEAD" {
		t.Fatalf("HEAD destacado: %+v, %v", got, ok)
	}
}
