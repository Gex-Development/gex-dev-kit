package kit

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Gex-Development/gex-dev-kit/internal/catalog"
)

func file(path, text string, mode os.FileMode) catalog.File {
	return catalog.File{Path: path, Data: []byte(text), Mode: mode}
}

func install(t *testing.T, home string, files ...catalog.File) string {
	t.Helper()
	plan, err := Preview(home, files)
	if err != nil {
		t.Fatal(err)
	}
	id, err := Apply(plan)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func TestApplyRollbackRestoresOriginalBytesAndMode(t *testing.T) {
	home := t.TempDir()
	path := filepath.Join(home, ".claude", "settings.json")
	if err := os.Mkdir(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("old private config\n"), 0600); err != nil {
		t.Fatal(err)
	}
	id := install(t, home, file(".claude/settings.json", "new config\n", 0644))
	if id == "" {
		t.Fatal("missing history ID")
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != "new config\n" {
		t.Fatalf("installed: %q, %v", got, err)
	}
	restored, err := Rollback(home)
	if err != nil || restored != id {
		t.Fatalf("rollback: %q, %v", restored, err)
	}
	got, err = os.ReadFile(path)
	if err != nil || string(got) != "old private config\n" {
		t.Fatalf("restored: %q, %v", got, err)
	}
	stat, err := os.Stat(path)
	if err != nil || !sameMode(stat.Mode(), 0600) {
		t.Fatalf("mode: %v, %v", stat, err)
	}
	again, err := Rollback(home)
	if err != nil || again != "" {
		t.Fatalf("rollback sem registro ativo deve ser no-op: %q, %v", again, err)
	}
	got, err = os.ReadFile(path)
	if err != nil || string(got) != "old private config\n" {
		t.Fatalf("segundo rollback alterou arquivo: %q, %v", got, err)
	}
}

func TestIncrementalInstallRollbackOneStepAtATime(t *testing.T) {
	home := t.TempDir()
	first := install(t, home, file(".claude/CLAUDE.md", "core", 0600))
	second := install(t, home, file(".claude/CLAUDE.md", "core plus skill", 0600), file(".agents/skills/tdd/SKILL.md", "test", 0600))
	if first == second {
		t.Fatal("history ID reused")
	}
	id, err := Rollback(home)
	if err != nil || id != second {
		t.Fatalf("latest rollback: %s %v", id, err)
	}
	content, err := os.ReadFile(filepath.Join(home, ".claude/CLAUDE.md"))
	if err != nil || string(content) != "core" {
		t.Fatalf("first install not restored: %q %v", content, err)
	}
	if _, err := os.Stat(filepath.Join(home, ".agents/skills/tdd/SKILL.md")); !os.IsNotExist(err) {
		t.Fatalf("new file remains: %v", err)
	}
	id, err = Rollback(home)
	if err != nil || id != first {
		t.Fatalf("earlier rollback: %s %v", id, err)
	}
	if _, err := os.Stat(filepath.Join(home, ".claude/CLAUDE.md")); !os.IsNotExist(err) {
		t.Fatalf("originally absent file remains: %v", err)
	}
}

func TestRollbackConflictTouchesNothing(t *testing.T) {
	home := t.TempDir()
	install(t, home, file(".claude/CLAUDE.md", "installed", 0600), file(".codex/AGENTS.md", "installed", 0600))
	changed := filepath.Join(home, ".codex/AGENTS.md")
	if err := os.WriteFile(changed, []byte("user edit"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Rollback(home); err == nil {
		t.Fatal("rollback accepted modified target")
	}
	untouched, err := os.ReadFile(filepath.Join(home, ".claude/CLAUDE.md"))
	if err != nil || string(untouched) != "installed" {
		t.Fatalf("partial rollback: %q %v", untouched, err)
	}
	content, err := os.ReadFile(changed)
	if err != nil || string(content) != "user edit" {
		t.Fatalf("user edit lost: %q %v", content, err)
	}
}

func TestApplyRejectsChangedPreviewWithoutTouchingOtherFiles(t *testing.T) {
	home := t.TempDir()
	path := filepath.Join(home, "existing.txt")
	if err := os.WriteFile(path, []byte("before"), 0600); err != nil {
		t.Fatal(err)
	}
	plan, err := Preview(home, []catalog.File{file("created.txt", "new", 0600), file("existing.txt", "kit", 0600)})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("edited"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Apply(plan); err == nil {
		t.Fatal("accepted stale plan")
	}
	if _, err := os.Stat(filepath.Join(home, "created.txt")); !os.IsNotExist(err) {
		t.Fatalf("partial apply: %v", err)
	}
}

func TestSymlinkAndTraversalNeverEscapeHome(t *testing.T) {
	home, outside := t.TempDir(), t.TempDir()
	if err := os.Symlink(outside, filepath.Join(home, ".claude")); err != nil {
		t.Fatal(err)
	}
	if _, err := Preview(home, []catalog.File{file(".claude/CLAUDE.md", "secret", 0600)}); err == nil {
		t.Fatal("accepted linked parent")
	}
	if _, err := Preview(home, []catalog.File{file("../escape", "secret", 0600)}); err == nil {
		t.Fatal("accepted traversal")
	}
	if _, err := os.Stat(filepath.Join(outside, "CLAUDE.md")); !os.IsNotExist(err) {
		t.Fatalf("wrote outside HOME: %v", err)
	}
}

func TestDoctorReportsHistoryAndCodexTrustWithoutGuessingApproval(t *testing.T) {
	home := t.TempDir()
	statuses, err := Doctor(home)
	if err != nil {
		t.Fatal(err)
	}
	if len(statuses) != 1 || statuses[0].State != "sem-historico" {
		t.Fatalf("empty history: %+v", statuses)
	}
	install(t, home, file(".codex/hooks.json", `{"hooks":{"PreToolUse":[{"command":"gex-dev-kit hook commit-msg"}]}}`, 0600))
	statuses, err = Doctor(home)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, status := range statuses {
		if status.Path == ".codex/hooks.json" {
			found = status.State == "pendente" && strings.Contains(status.Detail, "/hooks")
		}
	}
	if !found {
		t.Fatalf("Codex trust instruction missing: %+v", statuses)
	}
}

func TestHistoryPrivateAndNoOpDoesNotOverwriteOtherFiles(t *testing.T) {
	home := t.TempDir()
	other := filepath.Join(home, "unrelated.txt")
	if err := os.WriteFile(other, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	managed := file(".claude/settings.json", `{"key":"private"}`, 0644)
	install(t, home, managed)
	dir := filepath.Join(home, ".local/share/gex-dev-kit/history")
	info, err := os.Stat(dir)
	if err != nil || !privateMode(info.Mode(), 0700) {
		t.Fatalf("history directory permissions: %v %v", info, err)
	}
	info, err = os.Stat(filepath.Join(dir, "state.json"))
	if err != nil || !privateMode(info.Mode(), 0600) {
		t.Fatalf("history file permissions: %v %v", info, err)
	}
	if runtime.GOOS == "windows" {
		backup, err := os.ReadFile(filepath.Join(dir, "state.json"))
		if err != nil || strings.Contains(string(backup), "private") {
			t.Fatalf("Windows history exposed configuration: %v", err)
		}
	}
	info, err = os.Stat(filepath.Join(home, managed.Path))
	if err != nil || !sameMode(info.Mode(), 0600) {
		t.Fatalf("private configuration mode: %v %v", info, err)
	}
	plan, err := Preview(home, []catalog.File{managed})
	if err != nil || len(plan.Changes) != 0 {
		t.Fatalf("idempotent preview: %+v %v", plan.Changes, err)
	}
	if _, err := Apply(plan); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(other)
	if err != nil || string(content) != "keep" {
		t.Fatalf("unmanaged file altered: %q %v", content, err)
	}
}

func TestRollbackDoesNotOverwriteFileAfterModeOnlyEdit(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows não representa permissões POSIX no modo")
	}
	home := t.TempDir()
	install(t, home, file("managed.txt", "installed", 0600))
	target := filepath.Join(home, "managed.txt")
	if err := os.Chmod(target, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Rollback(home); err == nil {
		t.Fatal("rollback accepted changed mode")
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("rollback removed edited file: %v", err)
	}
}

func TestRollbackRestoresModeChangedByInstall(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows não representa permissões POSIX no modo")
	}
	home := t.TempDir()
	target := filepath.Join(home, "script")
	if err := os.WriteFile(target, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(target, 0644); err != nil {
		t.Fatal(err)
	}
	install(t, home, file("script", "new", 0755))
	if _, err := Rollback(home); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(target)
	if err != nil || info.Mode().Perm() != 0644 {
		t.Fatalf("original mode not restored: %v %v", info, err)
	}
	content, err := os.ReadFile(target)
	if err != nil || string(content) != "old" {
		t.Fatalf("original bytes not restored: %q %v", content, err)
	}
}

func TestHomeAliasRetargetAfterPreviewAbortsWithoutWrites(t *testing.T) {
	area, first, second := t.TempDir(), t.TempDir(), t.TempDir()
	alias := filepath.Join(area, "home")
	if err := os.Symlink(first, alias); err != nil {
		t.Fatal(err)
	}
	plan, err := Preview(alias, []catalog.File{file(".claude/CLAUDE.md", "core", 0600)})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(alias); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(second, alias); err != nil {
		t.Fatal(err)
	}
	if _, err := Apply(plan); err == nil {
		t.Fatal("accepted redirected HOME")
	}
	for _, home := range []string{first, second} {
		if _, err := os.Stat(filepath.Join(home, ".claude/CLAUDE.md")); !os.IsNotExist(err) {
			t.Fatalf("wrote to redirected HOME: %s: %v", home, err)
		}
	}
}

func TestHomeAliasAllowsInstallInResolvedDirectory(t *testing.T) {
	area, home := t.TempDir(), t.TempDir()
	alias := filepath.Join(area, "home")
	if err := os.Symlink(home, alias); err != nil {
		t.Fatal(err)
	}
	install(t, alias, file(".claude/CLAUDE.md", "core", 0600))
	content, err := os.ReadFile(filepath.Join(home, ".claude/CLAUDE.md"))
	if err != nil || string(content) != "core" {
		t.Fatalf("resolved HOME not installed: %q %v", content, err)
	}
}
