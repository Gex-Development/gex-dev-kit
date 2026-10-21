package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Gex-Development/gex-dev-kit/internal/hooks"
)

const hookSecret = "SENHA-NAO-REPETIR"

func hookIsolate(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(home, "gitconfig"))
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	return home
}

func hookGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v falhou: %v: %s", args, err, out)
	}
}

func hookRepo(t *testing.T, remote string) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git indisponível")
	}
	dir := t.TempDir()
	hookGit(t, dir, "init", "-q")
	if remote != "" {
		hookGit(t, dir, "remote", "add", "origin", remote)
	}
	t.Chdir(dir)
	return dir
}

func hookBashEvent(t *testing.T, command string) string {
	t.Helper()
	data, err := json.Marshal(map[string]any{
		"hook_event_name": "PreToolUse",
		"tool_name":       "Bash",
		"tool_input":      map[string]string{"command": command},
	})
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestHookBlocksInvalidCommitWithoutLeakingInput(t *testing.T) {
	hookIsolate(t)
	hookRepo(t, "https://github.com/Gex-Development/x.git")
	event := hookBashEvent(t, "git commit -m '"+hookSecret+" fora do padrão'")
	var out, errors bytes.Buffer
	code := run([]string{"hook", "commit-msg"}, strings.NewReader(event), &out, &errors)
	if code != 2 || strings.Contains(errors.String(), hookSecret) || out.Len() != 0 || !strings.Contains(errors.String(), "Commit bloqueado") {
		t.Fatalf("hook: exit=%d out=%q err=%q", code, out.String(), errors.String())
	}
}

func TestHookAllowsInvalidCommitOutsideScope(t *testing.T) {
	hookIsolate(t)
	hookRepo(t, "https://github.com/outra-org/x.git")
	event := hookBashEvent(t, "git commit -m 'texto livre'")
	var out, errors bytes.Buffer
	code := run([]string{"hook", "commit-msg"}, strings.NewReader(event), &out, &errors)
	if code != 0 || out.Len() != 0 || errors.Len() != 0 {
		t.Fatalf("fora do escopo: exit=%d out=%q err=%q", code, out.String(), errors.String())
	}
}

func TestHookAllowsConventionalCommitInScope(t *testing.T) {
	hookIsolate(t)
	hookRepo(t, "https://github.com/Gex-Development/x.git")
	event := hookBashEvent(t, "git commit -m 'feat(cli): liberar instalação'")
	var out, errors bytes.Buffer
	if code := run([]string{"hook", "commit-msg"}, strings.NewReader(event), &out, &errors); code != 0 {
		t.Fatalf("commit válido: exit=%d err=%q", code, errors.String())
	}
}

func TestHookReadsGitMessageFromPath(t *testing.T) {
	hookIsolate(t)
	file := filepath.Join(t.TempDir(), "COMMIT_EDITMSG")
	if err := os.WriteFile(file, []byte("feat(cli): liberar instalação\n"), 0600); err != nil {
		t.Fatal(err)
	}
	code, out, errors := invoke(t, "hook", "commit-msg", file)
	if code != 0 || out != "" || errors != "" {
		t.Fatalf("mensagem Git: exit=%d out=%q err=%q", code, out, errors)
	}
}

func TestHookRejectsBadGitMessageWithoutEcho(t *testing.T) {
	hookIsolate(t)
	file := filepath.Join(t.TempDir(), "COMMIT_EDITMSG")
	if err := os.WriteFile(file, []byte("# comentário\n"+hookSecret+" livre\n"), 0600); err != nil {
		t.Fatal(err)
	}
	code, out, errors := invoke(t, "hook", "commit-msg", file)
	if code != 2 || out != "" || strings.Contains(errors, hookSecret) {
		t.Fatalf("mensagem ruim: exit=%d out=%q err=%q", code, out, errors)
	}
}

func TestHookSensitiveActionAsksInScope(t *testing.T) {
	hookIsolate(t)
	hookRepo(t, "https://github.com/Gex-Development/x.git")
	event := hookBashEvent(t, "cat .env")
	var out, errors bytes.Buffer
	code := run([]string{"hook", "sensitive-action"}, strings.NewReader(event), &out, &errors)
	if code != 0 || !strings.Contains(out.String(), `"permissionDecision":"ask"`) || errors.Len() != 0 {
		t.Fatalf("ação sensível: exit=%d out=%q err=%q", code, out.String(), errors.String())
	}
}

func TestHookSensitiveActionSilentOutsideScope(t *testing.T) {
	hookIsolate(t)
	hookRepo(t, "https://github.com/outra-org/x.git")
	event := hookBashEvent(t, "cat .env")
	var out, errors bytes.Buffer
	code := run([]string{"hook", "sensitive-action"}, strings.NewReader(event), &out, &errors)
	if code != 0 || out.Len() != 0 {
		t.Fatalf("fora do escopo: exit=%d out=%q", code, out.String())
	}
}

func TestHookSensitiveActionDoesNotEchoSecret(t *testing.T) {
	hookIsolate(t)
	hookRepo(t, "https://github.com/Gex-Development/x.git")
	event := hookBashEvent(t, "printenv "+hookSecret)
	var out, errors bytes.Buffer
	run([]string{"hook", "sensitive-action"}, strings.NewReader(event), &out, &errors)
	if !strings.Contains(out.String(), `"ask"`) || strings.Contains(out.String(), hookSecret) || strings.Contains(errors.String(), hookSecret) {
		t.Fatalf("segredo ecoado: out=%q err=%q", out.String(), errors.String())
	}
}

func TestHookPrePRAsksOnlyOnCreate(t *testing.T) {
	hookIsolate(t)
	hookRepo(t, "https://github.com/Gex-Development/x.git")
	var out bytes.Buffer
	run([]string{"hook", "pre-pr"}, strings.NewReader(hookBashEvent(t, "gh pr create --fill")), &out, &bytes.Buffer{})
	if !strings.Contains(out.String(), `"permissionDecision":"ask"`) {
		t.Fatalf("pr create sem pergunta: %q", out.String())
	}
	out.Reset()
	run([]string{"hook", "pre-pr"}, strings.NewReader(hookBashEvent(t, "gh pr list")), &out, &bytes.Buffer{})
	if out.Len() != 0 {
		t.Fatalf("pr list não deveria perguntar: %q", out.String())
	}
}

func TestHookSessionStartWithoutGit(t *testing.T) {
	hookIsolate(t)
	t.Setenv("PATH", "")
	t.Chdir(t.TempDir())
	event := `{"hook_event_name":"SessionStart","source":"startup"}`
	var out, errors bytes.Buffer
	code := run([]string{"hook", "session-start", "--agent", "claude"}, strings.NewReader(event), &out, &errors)
	if code != 0 || errors.Len() != 0 || out.Len() > 1200 || !strings.Contains(out.String(), "GEX") {
		t.Fatalf("session-start sem git: exit=%d len=%d err=%q out=%q", code, out.Len(), errors.String(), out.String())
	}
}

func TestHookSessionStartRejectsUnknownAgent(t *testing.T) {
	hookIsolate(t)
	code, _, errors := invoke(t, "hook", "session-start", "--agent", "outro")
	if code != 2 || errors == "" {
		t.Fatalf("agente inválido: exit=%d err=%q", code, errors)
	}
}

func TestHookPreCommitBlocksCredentialWithoutEcho(t *testing.T) {
	hookIsolate(t)
	dir := hookRepo(t, "")
	key := "AKIA" + "ABCDEFGHIJKLMNOP"
	if err := os.WriteFile(filepath.Join(dir, "config.txt"), []byte("chave="+key+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	hookGit(t, dir, "add", "config.txt")
	var out, errors bytes.Buffer
	code := run([]string{"hook", "pre-commit"}, strings.NewReader(""), &out, &errors)
	if code != 1 || !strings.Contains(errors.String(), "possível credencial") || strings.Contains(errors.String(), key) {
		t.Fatalf("credencial: exit=%d err=%q", code, errors.String())
	}
}

func TestHookPreCommitHonorsAllowFile(t *testing.T) {
	hookIsolate(t)
	dir := hookRepo(t, "")
	key := "AKIA" + "ABCDEFGHIJKLMNOP"
	if err := os.WriteFile(filepath.Join(dir, "config.txt"), []byte("chave="+key+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	hookGit(t, dir, "add", "config.txt")
	allow := hooks.AllowPath(filepath.Join(dir, ".git"))
	if err := os.MkdirAll(filepath.Dir(allow), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(allow, []byte("config.txt\taws-access-key\n"), 0600); err != nil {
		t.Fatal(err)
	}
	var out, errors bytes.Buffer
	if code := run([]string{"hook", "pre-commit"}, strings.NewReader(""), &out, &errors); code != 0 {
		t.Fatalf("exceção ignorada: exit=%d err=%q", code, errors.String())
	}
}

func TestHookPreCommitPassesWithoutStagedFiles(t *testing.T) {
	hookIsolate(t)
	hookRepo(t, "")
	var out, errors bytes.Buffer
	if code := run([]string{"hook", "pre-commit"}, strings.NewReader(""), &out, &errors); code != 0 || errors.Len() != 0 {
		t.Fatalf("sem staged: exit=%d err=%q", code, errors.String())
	}
}

func TestHookRejectsUnknownSubcommand(t *testing.T) {
	code, _, errors := invoke(t, "hook", "inexistente")
	if code != 2 || errors == "" {
		t.Fatalf("subcomando inválido: exit=%d err=%q", code, errors)
	}
}
