package app

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Gex-Development/gex-dev-kit/internal/catalog"
	"github.com/Gex-Development/gex-dev-kit/internal/kit"
)

func rollbackTestInstall(t *testing.T, home, exe string, hooks []string) {
	t.Helper()
	files, err := catalog.Build(home, exe, nil, hooks)
	if err != nil {
		t.Fatalf("catalog.Build falhou: %v", err)
	}
	plan, err := kit.Preview(home, files)
	if err != nil {
		t.Fatalf("kit.Preview falhou: %v", err)
	}
	if _, err := kit.Apply(plan); err != nil {
		t.Fatalf("kit.Apply falhou: %v", err)
	}
}

func rollbackTestDiff(t *testing.T, diffs []FileDiff, path string) FileDiff {
	t.Helper()
	for _, d := range diffs {
		if d.Path == path {
			return d
		}
	}
	t.Fatalf("prévia de rollback sem %s", path)
	return FileDiff{}
}

func TestPlanRollbackWithoutHistoryIsEmpty(t *testing.T) {
	plan, err := PlanRollback(t.TempDir(), DiffLinesTUI)
	if err != nil {
		t.Fatalf("PlanRollback sem histórico falhou: %v", err)
	}
	if len(plan.Plan.Entries) != 0 || len(plan.Diffs) != 0 {
		t.Fatalf("sem histórico a prévia deveria ser vazia: %+v", plan)
	}
}

func TestPlanRollbackShowsCoreBlockAfterInstall(t *testing.T) {
	home := t.TempDir()
	rollbackTestInstall(t, home, "", nil)
	plan, err := PlanRollback(home, DiffLinesTUI)
	if err != nil {
		t.Fatalf("PlanRollback falhou: %v", err)
	}
	d := rollbackTestDiff(t, plan.Diffs, ".claude/CLAUDE.md")
	if d.Action != "remover" || len(d.Lines) == 0 {
		t.Fatalf("CLAUDE.md na prévia: ação=%q linhas=%d", d.Action, len(d.Lines))
	}
}

func TestPlanRollbackWarnsWhenStableExecutableIsRemoved(t *testing.T) {
	home := t.TempDir()
	exe := filepath.Join(t.TempDir(), "gex-dev-kit")
	if err := os.WriteFile(exe, []byte{0, 1, 2}, 0755); err != nil {
		t.Fatalf("não criou executável de teste: %v", err)
	}
	rollbackTestInstall(t, home, exe, []string{"gex-conventional-commits"})
	plan, err := PlanRollback(home, DiffLinesTUI)
	if err != nil {
		t.Fatalf("PlanRollback falhou: %v", err)
	}
	d := rollbackTestDiff(t, plan.Diffs, ".local/share/gex-dev-kit/bin/gex-dev-kit")
	if !d.Binary || !strings.Contains(d.Summary, "os hooks passam a ignorar a verificação até nova instalação") {
		t.Fatalf("aviso do executável estável ausente: binário=%v resumo=%q", d.Binary, d.Summary)
	}
}

func TestRollbackRestoresInstalledState(t *testing.T) {
	home := t.TempDir()
	rollbackTestInstall(t, home, "", nil)
	id, err := Rollback(home)
	if err != nil {
		t.Fatalf("Rollback falhou: %v", err)
	}
	if id == "" {
		t.Fatalf("Rollback não devolveu o identificador da restauração")
	}
	if _, err := os.Stat(filepath.Join(home, ".claude", "CLAUDE.md")); !os.IsNotExist(err) {
		t.Fatalf("CLAUDE.md deveria ter sido removido pelo rollback: %v", err)
	}
}

func TestRollbackRefusesEditedFileWithExplanation(t *testing.T) {
	home := t.TempDir()
	rollbackTestInstall(t, home, "", nil)
	if err := os.WriteFile(filepath.Join(home, ".claude", "CLAUDE.md"), []byte("editado depois"), 0600); err != nil {
		t.Fatalf("não editou o arquivo de teste: %v", err)
	}
	_, err := Rollback(home)
	if !errors.Is(err, kit.ErrConflict) {
		t.Fatalf("rollback com arquivo editado deveria recusar com conflito, veio: %v", err)
	}
	if got := Explain(err); !strings.Contains(got.Cause, ".claude/CLAUDE.md") {
		t.Fatalf("explicação não cita o arquivo em conflito: %+v", got)
	}
}

func TestRollbackAgentsWithoutHistoryReportsNothingToRestore(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	err := RollbackAgents(t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "nenhuma instalação anterior") {
		t.Fatalf("rollback de agentes sem histórico deveria avisar que nada foi restaurado, veio: %v", err)
	}
}
