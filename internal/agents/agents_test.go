package agents

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func isolated(t *testing.T) (string, string) {
	t.Helper()
	home := t.TempDir()
	bin := filepath.Join(home, ".local", "bin")
	if err := os.MkdirAll(bin, 0700); err != nil {
		t.Fatal(err)
	}
	path := bin
	codexBin := filepath.Dir(binPath(home, "codex"))
	if codexBin != bin {
		if err := os.MkdirAll(codexBin, 0700); err != nil {
			t.Fatal(err)
		}
		path += string(os.PathListSeparator) + codexBin
	}
	t.Setenv("PATH", path)
	t.Setenv("HOME", home)
	original := windowsUserPathReady
	windowsUserPathReady = func(string, string) bool { return true }
	t.Cleanup(func() { windowsUserPathReady = original })
	t.Setenv("USERPROFILE", home)
	return home, bin
}
func TestWindowsCodexUsesOfficialExclusiveVisibleDirectory(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows installer layout")
	}
	home, _ := isolated(t)
	expected := filepath.Join(home, "AppData", "Local", "Programs", "OpenAI", "Codex", "bin", "codex.exe")
	if binPath(home, "codex") != expected {
		t.Fatalf("unexpected Codex launcher: %s", binPath(home, "codex"))
	}
	if filepath.Dir(binPath(home, "claude")) == filepath.Dir(expected) {
		t.Fatal("Codex junction would replace Claude directory")
	}
	for _, entry := range safeEnvironment(home) {
		if strings.HasPrefix(strings.ToUpper(entry), "CODEX_INSTALL_DIR=") {
			t.Fatal("Codex destination override would create junction in shared bin")
		}
	}
	changes, err := Preview(home)
	if err != nil {
		t.Fatal(err)
	}
	if changes[1].Blocked {
		t.Fatalf("official target on PATH blocked: %+v", changes[1])
	}
}

func TestPreviewBlocksUnownedAndManagedCommands(t *testing.T) {
	home, bin := isolated(t)
	path := filepath.Join(bin, executable("claude"))
	if err := os.WriteFile(path, []byte("old"), 0700); err != nil {
		t.Fatal(err)
	}
	changes, err := Preview(home)
	if err != nil {
		t.Fatal(err)
	}
	if !changes[0].Blocked || !strings.Contains(changes[0].Reason, "não gerenciado") {
		t.Fatalf("unexpected Claude change: %+v", changes[0])
	}
	if !strings.Contains(changes[0].Manual, "claude") {
		t.Fatalf("missing manual command: %+v", changes[0])
	}
	if err := Apply(home, changes); err == nil {
		t.Fatal("unowned executable accepted")
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(home, "node_modules", ".bin"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", filepath.Join(home, "node_modules", ".bin")+string(os.PathListSeparator)+bin)
	if err := os.WriteFile(filepath.Join(home, "node_modules", ".bin", executable("codex")), []byte("npm"), 0700); err != nil {
		t.Fatal(err)
	}
	changes, err = Preview(home)
	if err != nil {
		t.Fatal(err)
	}
	if !changes[1].Blocked || !strings.Contains(changes[1].Reason, "npm") {
		t.Fatalf("npm collision accepted: %+v", changes[1])
	}
}

func TestApplyRejectsBlockedBatchBeforeAnyInstall(t *testing.T) {
	home, _ := isolated(t)
	if err := os.WriteFile(binPath(home, "codex"), []byte("unowned"), 0700); err != nil {
		t.Fatal(err)
	}
	changes, err := Preview(home)
	if err != nil {
		t.Fatal(err)
	}
	if changes[0].Blocked || !changes[1].Blocked {
		t.Fatalf("unexpected changes: %+v", changes)
	}
	if err := Apply(home, changes); err == nil {
		t.Fatal("blocked batch accepted")
	}
	if _, err := os.Stat(binPath(home, "claude")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Claude changed: %v", err)
	}
	if _, err := os.Stat(statePath(home)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("history changed: %v", err)
	}
}
func TestRejectsTamperedRollbackVersionWithoutWriting(t *testing.T) {
	home, _ := isolated(t)
	if err := save(home, history{"claude": {Current: "2.1.306", Previous: "; rm -rf ~"}}); err != nil {
		t.Fatal(err)
	}
	_, err := Preview(home)
	if err == nil || !strings.Contains(err.Error(), "histórico agentes inválido") {
		t.Fatalf("invalid history accepted: %v", err)
	}
}

func TestApplyAndRollbackUseExactVersions(t *testing.T) {
	home, bin := isolated(t)
	previousFetch, previousRun := fetchScript, runCommand
	t.Cleanup(func() { fetchScript, runCommand = previousFetch, previousRun })
	fetchScript = func(url string) ([]byte, error) { return []byte("official installer"), nil }
	target := filepath.Join(bin, executable("claude"))
	runCommand = func(home, path string, input []byte, args ...string) (string, error) {
		if len(input) > 0 || strings.EqualFold(path, "powershell.exe") {
			version := args[len(args)-1]
			if err := os.WriteFile(target, []byte(version), 0700); err != nil {
				return "", err
			}
			return "", nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		return string(b) + " (Claude Code)", nil
	}
	changes, err := Preview(home)
	if err != nil {
		t.Fatal(err)
	}
	if changes[0].Blocked || changes[0].Target != "2.1.306" {
		t.Fatalf("unexpected preview %+v", changes[0])
	}
	if err := Apply(home, changes[:1]); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("2.1.306"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := Rollback(home); err == nil {
		t.Fatal("first install has no previous version")
	}
}

func TestRollbackRestoresPreviousKitRelease(t *testing.T) {
	home, bin := isolated(t)
	oldFetch, oldRun := fetchScript, runCommand
	t.Cleanup(func() { fetchScript, runCommand = oldFetch, oldRun })
	fetchScript = func(string) ([]byte, error) { return []byte("official installer"), nil }
	target := filepath.Join(bin, executable("claude"))
	if err := os.WriteFile(target, []byte("2.1.305"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := save(home, history{"claude": {Current: "2.1.305"}}); err != nil {
		t.Fatal(err)
	}
	runCommand = func(home, path string, input []byte, args ...string) (string, error) {
		if len(input) > 0 || strings.EqualFold(path, "powershell.exe") {
			if err := os.WriteFile(target, []byte(args[len(args)-1]), 0700); err != nil {
				return "", err
			}
			return "", nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		return string(b) + " (Claude Code)", nil
	}
	changes, err := Preview(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := Apply(home, changes[:1]); err != nil {
		t.Fatal(err)
	}
	if err := Rollback(home); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "2.1.305" {
		t.Fatalf("rollback installed %s", b)
	}
	h, err := load(home)
	if err != nil {
		t.Fatal(err)
	}
	if h["claude"].Current != "2.1.305" || h["claude"].Previous != "" {
		t.Fatalf("history after rollback: %+v", h)
	}
}

func TestCodexPinAndVerificationFailure(t *testing.T) {
	home, _ := isolated(t)
	oldFetch, oldRun := fetchScript, runCommand
	t.Cleanup(func() { fetchScript, runCommand = oldFetch, oldRun })
	fetchScript = func(string) ([]byte, error) { return []byte("official installer"), nil }
	target := binPath(home, "codex")
	runCommand = func(home, path string, input []byte, args ...string) (string, error) {
		if len(input) > 0 || strings.EqualFold(path, "powershell.exe") {
			if args[len(args)-1] != "0.163.0" {
				t.Fatalf("wrong release argument: %v", args)
			}
			if err := os.WriteFile(target, []byte("codex"), 0700); err != nil {
				return "", err
			}
			return "", nil
		}
		return "codex-cli 0.162.0", nil
	}
	changes, err := Preview(home)
	if err != nil {
		t.Fatal(err)
	}
	if changes[1].Target != "0.163.0" || changes[1].Blocked {
		t.Fatalf("unexpected preview: %+v", changes[1])
	}
	err = Apply(home, changes[1:])
	if err == nil || !strings.Contains(err.Error(), "esperado 0.163.0") {
		t.Fatalf("version mismatch not rejected: %v", err)
	}
	h, err := load(home)
	if err != nil {
		t.Fatal(err)
	}
	if !h["codex"].Pending || h["codex"].Current != "" {
		t.Fatalf("failed install marked committed: %+v", h)
	}
}

func TestApplyFailsClosedOnInstallerFailure(t *testing.T) {
	home, _ := isolated(t)
	oldFetch, oldRun := fetchScript, runCommand
	t.Cleanup(func() { fetchScript, runCommand = oldFetch, oldRun })
	fetchScript = func(string) ([]byte, error) { return []byte("installer"), nil }
	runCommand = func(string, string, []byte, ...string) (string, error) { return "", errors.New("network unavailable") }
	changes, err := Preview(home)
	if err != nil {
		t.Fatal(err)
	}
	err = Apply(home, changes[:1])
	if err == nil || !strings.Contains(err.Error(), "network unavailable") {
		t.Fatalf("unexpected error: %v", err)
	}
	if err = Rollback(home); err == nil {
		t.Fatal("incomplete install must not be rolled back")
	}
}

func TestStateSymlinkRejected(t *testing.T) {
	home, _ := isolated(t)
	path := statePath(home)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "private.json")
	if err := os.WriteFile(outside, []byte(`{}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, path); err != nil {
		t.Skipf("symlinks indisponíveis: %v", err)
	}
	if _, err := Preview(home); err == nil {
		t.Fatal("histórico por symlink aceito")
	}
}

func TestInstallerEnvironmentExcludesSecrets(t *testing.T) {
	home, _ := isolated(t)
	t.Setenv("ANTHROPIC_API_KEY", "synthetic-canary")
	t.Setenv("OPENAI_API_KEY", "synthetic-canary")
	t.Setenv("CODEX_HOME", "/other-user/codex")
	env := strings.Join(safeEnvironment(home), "\n")
	if strings.Contains(env, "synthetic-canary") || strings.Contains(env, "/other-user/") {
		t.Fatalf("credenciais ou diretório alheio passaram ao instalador")
	}
	if !strings.Contains(env, "CODEX_HOME="+filepath.Join(home, ".codex")) {
		t.Fatal("CODEX_HOME não isolado")
	}
}

func TestFailedCommandDoesNotRevealOutput(t *testing.T) {
	if filepath.Separator == '\\' {
		t.Skip("shell Unix")
	}
	home, _ := isolated(t)
	_, err := runCommand(home, "/bin/sh", nil, "-c", "echo synthetic-canary >&2; exit 7")
	if err == nil || strings.Contains(err.Error(), "synthetic-canary") {
		t.Fatalf("saída do comando exposta: %v", err)
	}
}
