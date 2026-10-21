package kit

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

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

func TestRollbackConflictReturnsErrConflictWithPaths(t *testing.T) {
	home := t.TempDir()
	install(t, home, file(".claude/CLAUDE.md", "installed", 0600), file(".codex/AGENTS.md", "installed", 0600))
	if err := os.WriteFile(filepath.Join(home, ".codex", "AGENTS.md"), []byte("user edit"), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := Rollback(home)
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("conflito sem sentinela ErrConflict: %v", err)
	}
	if !strings.Contains(err.Error(), ".codex/AGENTS.md") {
		t.Fatalf("conflito sem lista de caminhos: %v", err)
	}
}

func kitTestWrite(t *testing.T, p, text string, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(text), mode); err != nil {
		t.Fatal(err)
	}
}

func kitTestState(t *testing.T, home string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(home, filepath.FromSlash(historyState)))
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	return data
}

func kitTestMarkRecord(t *testing.T, home string, edit func(*record)) {
	t.Helper()
	root, err := openHome(home)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	h, err := readHistory(root)
	if err != nil {
		t.Fatal(err)
	}
	edit(&h.Records[len(h.Records)-1])
	if err := saveHistory(root, h); err != nil {
		t.Fatal(err)
	}
}

func kitTestWriteLock(t *testing.T, home string, pid int) {
	t.Helper()
	data, err := json.Marshal(lockFile{PID: pid, Started: time.Now().UTC().Format(time.RFC3339)})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, filepath.FromSlash(lockName)), data, 0600); err != nil {
		t.Fatal(err)
	}
}

func kitTestDeadPID(t *testing.T) int {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^$")
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	return cmd.Process.Pid
}

func kitTestStatus(statuses []Status, p string) (Status, bool) {
	for _, s := range statuses {
		if s.Path == p {
			return s, true
		}
	}
	return Status{}, false
}

func kitTestCheckEntry(t *testing.T, home string, e RollbackEntry) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(home, filepath.FromSlash(e.Path)))
	exists := err == nil
	switch e.Action {
	case "remover":
		if exists {
			t.Fatalf("%s: prévia disse remover e o arquivo continua", e.Path)
		}
	case "restaurar":
		if !exists || !bytes.Equal(data, e.Restore) {
			t.Fatalf("%s: prévia disse restaurar; conteúdo %q", e.Path, data)
		}
	case "inalterado":
		if !bytes.Equal(data, e.Current) {
			t.Fatalf("%s: prévia disse inalterado; conteúdo %q", e.Path, data)
		}
	default:
		t.Fatalf("ação desconhecida na prévia: %q", e.Action)
	}
}

func TestApplyStaleReturnsErrChanged(t *testing.T) {
	home := t.TempDir()
	kitTestWrite(t, filepath.Join(home, "existing.txt"), "before", 0600)
	plan, err := Preview(home, []catalog.File{file("existing.txt", "kit", 0600)})
	if err != nil {
		t.Fatal(err)
	}
	kitTestWrite(t, filepath.Join(home, "existing.txt"), "edited", 0600)
	if _, err := Apply(plan); !errors.Is(err, ErrChanged) {
		t.Fatalf("prévia desatualizada sem ErrChanged: %v", err)
	}
}

func TestSymlinkReturnsErrSymlink(t *testing.T) {
	home, outside := t.TempDir(), t.TempDir()
	if err := os.Symlink(outside, filepath.Join(home, ".claude")); err != nil {
		t.Fatal(err)
	}
	if _, err := Preview(home, []catalog.File{file(".claude/CLAUDE.md", "core", 0600)}); !errors.Is(err, ErrSymlink) {
		t.Fatalf("link simbólico sem ErrSymlink: %v", err)
	}
}

func TestPendingRecordReturnsErrIncomplete(t *testing.T) {
	home := t.TempDir()
	install(t, home, file("a.txt", "1", 0600))
	kitTestMarkRecord(t, home, func(r *record) { r.Pending = true })
	plan, err := Preview(home, []catalog.File{file("b.txt", "2", 0600)})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Apply(plan); !errors.Is(err, ErrIncomplete) {
		t.Fatalf("instalação anterior incompleta sem ErrIncomplete: %v", err)
	}
}

func TestBusyLockReturnsErrBusy(t *testing.T) {
	home := t.TempDir()
	install(t, home, file("a.txt", "1", 0600))
	kitTestWriteLock(t, home, os.Getpid())
	plan, err := Preview(home, []catalog.File{file("a.txt", "2", 0600)})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Apply(plan); !errors.Is(err, ErrBusy) {
		t.Fatalf("histórico com lock sem ErrBusy: %v", err)
	}
}

func TestRollbackPreviewPredictsRollback(t *testing.T) {
	cases := []struct {
		name  string
		setup func(t *testing.T, home string)
	}{
		{"sem histórico", nil},
		{"instalação simples", func(t *testing.T, home string) {
			kitTestWrite(t, filepath.Join(home, ".claude", "settings.json"), "old private config\n", 0600)
			install(t, home, file(".claude/settings.json", "new config\n", 0644))
		}},
		{"incremental", func(t *testing.T, home string) {
			install(t, home, file(".claude/CLAUDE.md", "core", 0600))
			install(t, home, file(".claude/CLAUDE.md", "core plus skill", 0600), file(".agents/skills/tdd/SKILL.md", "test", 0600))
		}},
		{"conflito", func(t *testing.T, home string) {
			install(t, home, file(".claude/CLAUDE.md", "installed", 0600), file(".codex/AGENTS.md", "installed", 0600))
			kitTestWrite(t, filepath.Join(home, ".codex", "AGENTS.md"), "user edit", 0600)
		}},
		{"modo alterado", func(t *testing.T, home string) {
			install(t, home, file("managed.txt", "installed", 0600))
			if err := os.Chmod(filepath.Join(home, "managed.txt"), 0644); err != nil {
				t.Fatal(err)
			}
		}},
		{"pendente", func(t *testing.T, home string) {
			install(t, home, file(".claude/CLAUDE.md", "installed", 0600))
			kitTestMarkRecord(t, home, func(r *record) { r.Pending = true })
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			if tc.setup != nil {
				tc.setup(t, home)
			}
			before := kitTestState(t, home)
			preview, err := RollbackPreview(home)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(before, kitTestState(t, home)) {
				t.Fatal("prévia alterou o histórico")
			}
			id, err := Rollback(home)
			if len(preview.Conflicts) > 0 {
				if !errors.Is(err, ErrConflict) {
					t.Fatalf("prévia previu conflito e rollback retornou %q, %v", id, err)
				}
				return
			}
			if err != nil || id != preview.ID {
				t.Fatalf("prévia previu %q e rollback retornou %q, %v", preview.ID, id, err)
			}
			for _, e := range preview.Entries {
				kitTestCheckEntry(t, home, e)
			}
		})
	}
}

func TestInterruptedRollbackCompletesOnSecondCall(t *testing.T) {
	home := t.TempDir()
	kitTestWrite(t, filepath.Join(home, "a.txt"), "old a", 0600)
	kitTestWrite(t, filepath.Join(home, "b.txt"), "old b", 0600)
	install(t, home, file("a.txt", "new a", 0600), file("b.txt", "new b", 0600))
	kitTestMarkRecord(t, home, func(r *record) { r.RollingBack = true })
	kitTestWrite(t, filepath.Join(home, "a.txt"), "old a", 0600)
	id, err := Rollback(home)
	if err != nil || id == "" {
		t.Fatalf("segunda chamada do rollback não concluiu: %q, %v", id, err)
	}
	for name, want := range map[string]string{"a.txt": "old a", "b.txt": "old b"} {
		got, err := os.ReadFile(filepath.Join(home, name))
		if err != nil || string(got) != want {
			t.Fatalf("%s após rollback: %q, %v", name, got, err)
		}
	}
	if again, err := Rollback(home); err != nil || again != "" {
		t.Fatalf("rollback concluído deve ser no-op: %q, %v", again, err)
	}
}

func TestLockReportsOwnerLiveness(t *testing.T) {
	home := t.TempDir()
	install(t, home, file("a.txt", "1", 0600))
	if info, present, err := Lock(home); err != nil || present {
		t.Fatalf("sem lock deve ser ausente: %+v %v %v", info, present, err)
	}
	kitTestWriteLock(t, home, os.Getpid())
	info, present, err := Lock(home)
	if err != nil || !present || !info.Alive || info.PID != os.Getpid() {
		t.Fatalf("lock do processo atual deve estar vivo: %+v %v %v", info, present, err)
	}
	kitTestWriteLock(t, home, kitTestDeadPID(t))
	info, present, err = Lock(home)
	if err != nil || !present || info.Alive {
		t.Fatalf("lock de processo morto deve ser órfão: %+v %v %v", info, present, err)
	}
}

func TestDoctorBlocksLockWithOwnerStateAndAction(t *testing.T) {
	home := t.TempDir()
	install(t, home, file("a.txt", "1", 0600))
	kitTestWriteLock(t, home, os.Getpid())
	statuses, err := Doctor(home)
	s, ok := kitTestStatus(statuses, lockName)
	if err != nil || !ok || s.State != "bloqueado" || !strings.Contains(s.Detail, "está ativa") {
		t.Fatalf("lock vivo no doctor: %+v %v", statuses, err)
	}
	kitTestWriteLock(t, home, kitTestDeadPID(t))
	statuses, err = Doctor(home)
	s, ok = kitTestStatus(statuses, lockName)
	if err != nil || !ok || s.State != "bloqueado" || !strings.Contains(s.Detail, "rode de novo") {
		t.Fatalf("lock órfão sem ação no doctor: %+v %v", statuses, err)
	}
}

func TestDoctorBlocksSymlinkItemAndKeepsOtherItems(t *testing.T) {
	home, outside := t.TempDir(), t.TempDir()
	install(t, home, file("linked.txt", "a", 0600), file("plain.txt", "b", 0600))
	if err := os.Remove(filepath.Join(home, "linked.txt")); err != nil {
		t.Fatal(err)
	}
	kitTestWrite(t, filepath.Join(outside, "target.txt"), "x", 0600)
	if err := os.Symlink(filepath.Join(outside, "target.txt"), filepath.Join(home, "linked.txt")); err != nil {
		t.Fatal(err)
	}
	statuses, err := Doctor(home)
	if err != nil {
		t.Fatalf("doctor abortou com link simbólico: %v", err)
	}
	if s, _ := kitTestStatus(statuses, "linked.txt"); s.State != "bloqueado" {
		t.Fatalf("link simbólico sem item bloqueado: %+v", statuses)
	}
	if s, _ := kitTestStatus(statuses, "plain.txt"); s.State != "ok" {
		t.Fatalf("item íntegro perdido após link simbólico: %+v", statuses)
	}
}

func TestDoctorCodexSessionStartHandlerIsPending(t *testing.T) {
	home := t.TempDir()
	install(t, home, file(".codex/hooks.json", `{"hooks":{"SessionStart":[{"command":"gex-dev-kit hook session-start --agent codex"}]}}`, 0600))
	statuses, err := Doctor(home)
	if s, ok := kitTestStatus(statuses, ".codex/hooks.json"); err != nil || !ok || s.State != "pendente" {
		t.Fatalf("handler session-start do kit deve ser pendente: %+v %v", statuses, err)
	}
	other := t.TempDir()
	install(t, other, file(".codex/hooks.json", `{"hooks":{"PreToolUse":[{"command":"outra-ferramenta run"}]}}`, 0600))
	statuses, err = Doctor(other)
	if s, ok := kitTestStatus(statuses, ".codex/hooks.json"); err != nil || !ok || s.State != "ok" {
		t.Fatalf("handler de terceiros não deve ser pendente: %+v %v", statuses, err)
	}
}

func TestPlanContentsMarksBinaryFiles(t *testing.T) {
	plan, err := Preview(t.TempDir(), []catalog.File{
		{Path: ".local/share/gex-dev-kit/bin/gex-dev-kit", Data: []byte("elf"), Mode: 0755},
		{Path: "blob.dat", Data: []byte{'a', 0, 'b'}, Mode: 0600},
		{Path: "notes.md", Data: []byte("texto"), Mode: 0600},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{".local/share/gex-dev-kit/bin/gex-dev-kit": true, "blob.dat": true, "notes.md": false}
	contents := plan.Contents()
	if len(contents) != len(want) {
		t.Fatalf("conteúdos da prévia: %+v", contents)
	}
	for _, c := range contents {
		if c.Binary != want[c.Path] {
			t.Fatalf("%s: binário=%v, esperado %v", c.Path, c.Binary, want[c.Path])
		}
	}
}

const kitTestBinPath = ".local/share/gex-dev-kit/bin/gex-dev-kit"

func kitTestBinary(seed byte) []byte {
	data := make([]byte, 2<<20)
	for i := range data {
		data[i] = seed + byte(i*31)
	}
	return data
}

func kitTestBlobNames(t *testing.T, home string) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(home, filepath.FromSlash(blobDir)))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

func kitTestHistory(t *testing.T, home string) history {
	t.Helper()
	root, err := openHome(home)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	h, err := readHistory(root)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func kitTestWriteLegacy(t *testing.T, home, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(home, filepath.FromSlash(historyDir)), 0700); err != nil {
		t.Fatal(err)
	}
	sealed, err := protectHistory([]byte(text))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, filepath.FromSlash(historyState)), sealed, 0600); err != nil {
		t.Fatal(err)
	}
}

func TestRepeatedBinaryInstallKeepsStateSmallAndDeduplicatesBlobs(t *testing.T) {
	home := t.TempDir()
	a, b := kitTestBinary(1), kitTestBinary(2)
	for _, bin := range [][]byte{a, b, a, b} {
		install(t, home, catalog.File{Path: kitTestBinPath, Data: bin, Mode: 0755})
	}
	state, err := os.ReadFile(filepath.Join(home, filepath.FromSlash(historyState)))
	if err != nil || len(state) >= 64<<10 {
		t.Fatalf("state.json com %d bytes após trocas do executável: %v", len(state), err)
	}
	if names := kitTestBlobNames(t, home); len(names) != 2 {
		t.Fatalf("blobs esperados 2 (deduplicados por hash), obtidos %d: %v", len(names), names)
	}
}

func TestSixInstallsPruneOldestBlobsAndKeepDoctorVisibility(t *testing.T) {
	home := t.TempDir()
	kitTestWrite(t, filepath.Join(home, "a.txt"), "original", 0600)
	install(t, home, file("a.txt", "v1", 0600), file("only1.txt", "x", 0600))
	for i := 2; i <= 6; i++ {
		install(t, home, file("a.txt", "v"+string(rune('0'+i)), 0600))
	}
	names := kitTestBlobNames(t, home)
	if len(names) != 5 {
		t.Fatalf("esperados 5 blobs das gravações 2 a 6, obtidos %d: %v", len(names), names)
	}
	for _, name := range names {
		if name == digest([]byte("original")) {
			t.Fatal("blob da 1ª gravação não foi podado")
		}
	}
	h := kitTestHistory(t, home)
	if len(h.Records) != 6 || !h.Records[0].Pruned || h.Records[0].Entries[0].BeforeHash != digest([]byte("original")) {
		t.Fatalf("hash da 1ª gravação não preservado: %+v", h.Records[0])
	}
	statuses, err := Doctor(home)
	if s, ok := kitTestStatus(statuses, "only1.txt"); err != nil || !ok || s.State != "ok" {
		t.Fatalf("doctor não enxerga arquivo da 1ª gravação: %+v %v", statuses, err)
	}
}

func TestRollbackRefusesPrunedRecordWithoutTouchingFiles(t *testing.T) {
	home := t.TempDir()
	kitTestWrite(t, filepath.Join(home, "a.txt"), "original", 0600)
	install(t, home, file("a.txt", "v1", 0600), file("only1.txt", "x", 0600))
	for i := 2; i <= 6; i++ {
		install(t, home, file("a.txt", "v"+string(rune('0'+i)), 0600))
	}
	for i := 0; i < 5; i++ {
		if _, err := Rollback(home); err != nil {
			t.Fatalf("rollback %d das 5 gravações mais recentes: %v", i+1, err)
		}
	}
	if _, err := RollbackPreview(home); !errors.Is(err, ErrPruned) {
		t.Fatalf("prévia da gravação podada sem ErrPruned: %v", err)
	}
	if _, err := Rollback(home); !errors.Is(err, ErrPruned) {
		t.Fatalf("rollback da gravação podada sem ErrPruned: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(home, "a.txt"))
	if err != nil || string(got) != "v1" {
		t.Fatalf("arquivo alterado pela recusa: %q %v", got, err)
	}
}

func TestLegacyV02HistoryRollsBackAfterMigration(t *testing.T) {
	home := t.TempDir()
	kitTestWrite(t, filepath.Join(home, "a.txt"), "new", 0600)
	kitTestWriteLegacy(t, home, `{"records":[{"id":"legado","entries":[{"path":"a.txt","before":"b2xk","before_mode":384,"existed":true,"after":"bmV3","after_mode":384}]}]}`)
	install(t, home, file("b.txt", "novo", 0600))
	h := kitTestHistory(t, home)
	if len(h.Records) != 2 || h.Records[0].Entries[0].BeforeHash != digest([]byte("old")) {
		t.Fatalf("migração não converteu o registro v0.2.0: %+v", h.Records)
	}
	if _, err := os.Stat(filepath.Join(home, filepath.FromSlash(blobPath(digest([]byte("old")))))); err != nil {
		t.Fatalf("blob do conteúdo legado não foi gravado: %v", err)
	}
	if id, err := Rollback(home); err != nil || id == "legado" {
		t.Fatalf("primeiro rollback deveria reverter b.txt: %q %v", id, err)
	}
	if id, err := Rollback(home); err != nil || id != "legado" {
		t.Fatalf("segundo rollback deveria reverter o registro legado: %q %v", id, err)
	}
	got, err := os.ReadFile(filepath.Join(home, "a.txt"))
	if err != nil || string(got) != "old" {
		t.Fatalf("conteúdo legado não restaurado: %q %v", got, err)
	}
}
