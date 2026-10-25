package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Gex-Development/gex-dev-kit/internal/catalog"
	"github.com/Gex-Development/gex-dev-kit/internal/kit"
)

func coreStatuses(names ...string) []kit.Status {
	var statuses []kit.Status
	for _, name := range names {
		statuses = append(statuses,
			kit.Status{Path: ".claude/skills/" + name + "/SKILL.md", State: "ok"},
			kit.Status{Path: ".agents/skills/" + name + "/SKILL.md", State: "ok"},
		)
	}
	return statuses
}

func TestCoreSummaryStates(t *testing.T) {
	all := catalog.CoreSkills()
	total := len(all)
	for _, tc := range []struct {
		name      string
		statuses  []kit.Status
		state     string
		active    int
		attention int
	}{
		{"sem histórico", []kit.Status{{Path: "history/state.json", State: "sem-historico"}}, "não instalado", 0, 0},
		{"íntegro", coreStatuses(all...), "íntegro", total, 0},
		{"atualização disponível", coreStatuses(all[0]), "atualização disponível", 1, 0},
		{"pendente", append(coreStatuses(all...), kit.Status{Path: ".codex/hooks.json", State: "pendente"}), "pendente", total, 0},
		{"alterado", append(coreStatuses(all...), kit.Status{Path: ".claude/CLAUDE.md", State: "alterado"}), "alterado", total, 1},
		{"alterado após pendente", append(coreStatuses(all...), kit.Status{Path: ".codex/hooks.json", State: "pendente"}, kit.Status{Path: ".claude/CLAUDE.md", State: "alterado"}), "alterado", total, 1},
	} {
		got := coreSummary(tc.statuses)
		if got.State != tc.state || got.ActiveCore != tc.active || got.Attention != tc.attention || got.TotalCore != total {
			t.Fatalf("%s: %+v, esperado estado %q, ativas %d, atenção %d", tc.name, got, tc.state, tc.active, tc.attention)
		}
	}
}

func TestCoreSummaryListsHooksWhenHookFileIsManaged(t *testing.T) {
	got := coreSummary(append(coreStatuses(catalog.CoreSkills()...), kit.Status{Path: ".claude/settings.json", State: "ok"}))
	if len(got.Hooks) == 0 {
		t.Fatal("hook gerenciado não listado")
	}
	if none := coreSummary(coreStatuses(catalog.CoreSkills()...)); len(none.Hooks) != 0 {
		t.Fatalf("sem arquivo de hook não deveria listar hooks: %v", none.Hooks)
	}
}

func TestCoreWithoutHistoryIsNotInstalled(t *testing.T) {
	got, err := Core(t.TempDir())
	if err != nil {
		t.Fatalf("leitura falhou: %v", err)
	}
	if got.State != "não instalado" || got.ActiveCore != 0 || got.TotalCore != len(catalog.CoreSkills()) {
		t.Fatalf("estado sem histórico: %+v", got)
	}
}

func TestCoreAfterInstallIsIntactAndDetectsEdit(t *testing.T) {
	r := installRequest(t)
	if _, err := ApplyInstall(installPlan(t, r), nil); err != nil {
		t.Fatalf("aplicação falhou: %v", err)
	}
	got, err := Core(r.Home)
	if err != nil {
		t.Fatalf("leitura falhou: %v", err)
	}
	if got.State != "íntegro" || got.ActiveCore != got.TotalCore || got.Attention != 0 {
		t.Fatalf("instalação recém-feita: %+v", got)
	}
	skill := filepath.Join(r.Home, ".claude", "skills", catalog.CoreSkills()[0], "SKILL.md")
	if err := os.WriteFile(skill, []byte("editado"), 0644); err != nil {
		t.Fatal(err)
	}
	got, err = Core(r.Home)
	if err != nil {
		t.Fatalf("leitura falhou: %v", err)
	}
	if got.State != "alterado" || got.Attention == 0 {
		t.Fatalf("edição manual não detectada: %+v", got)
	}
}
