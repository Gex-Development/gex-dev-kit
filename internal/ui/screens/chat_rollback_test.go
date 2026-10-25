package screens

import (
	"strings"
	"testing"

	"github.com/Gex-Development/gex-dev-kit/internal/app"
)

func TestChatRollbackNoHistory(t *testing.T) {
	m := chatTestModel(false)
	m.deps.Rollback = RollbackDeps{Plan: func() (app.RollbackPreview, error) { return app.RollbackPreview{}, nil }}
	m = m.flowRollback(false)
	m = chatSetupQuestion(t, m, "")
	if !strings.Contains(chatTestTranscript(m), "Nada a restaurar.") {
		t.Fatal("rollback sem histórico deve informar ausência")
	}
}

func TestChatRollbackConflictDoesNotApply(t *testing.T) {
	applied := 0
	m := chatTestModel(false)
	m.deps.Rollback = RollbackDeps{Plan: func() (app.RollbackPreview, error) {
		p := app.RollbackPreview{}
		p.Plan.Conflicts = []string{"changed.txt"}
		return p, nil
	}, Apply: func() (string, error) { applied++; return "", nil }}
	m = m.flowRollback(false)
	m = chatSetupQuestion(t, m, "")
	if applied != 0 || m.exitCode != 1 || !strings.Contains(chatTestTranscript(m), "Não restauro") {
		t.Fatalf("conflito precisa abortar; aplicado=%d, código=%d", applied, m.exitCode)
	}
}

func TestChatRollbackAppliesAfterConfirmation(t *testing.T) {
	applied := 0
	m := chatTestModel(false)
	m.deps.Rollback = RollbackDeps{Plan: func() (app.RollbackPreview, error) { p := app.RollbackPreview{}; p.Plan.ID = "h-1"; return p, nil }, Apply: func() (string, error) { applied++; return "h-1", nil }}
	m = m.flowRollback(false)
	m = chatSetupQuestion(t, m, "Restauro esta instalação?")
	if applied != 0 {
		t.Fatal("gravação antes da confirmação")
	}
	m = chatSetupChoose(t, m, "yes")
	m = chatSetupQuestion(t, m, "")
	if applied != 1 || !strings.Contains(chatTestTranscript(m), "Restaurado. Histórico: h-1.") {
		t.Fatalf("rollback esperado uma vez; aplicado=%d", applied)
	}
}
