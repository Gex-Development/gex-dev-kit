package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Gex-Development/gex-dev-kit/internal/catalog"
	"github.com/Gex-Development/gex-dev-kit/internal/kit"
)

func doctorIsolate(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(t.TempDir(), "config"))
}

func doctorFind(t *testing.T, items []kit.Status, path string) kit.Status {
	t.Helper()
	for _, item := range items {
		if item.Path == path {
			return item
		}
	}
	t.Fatalf("item %q ausente em %v", path, items)
	return kit.Status{}
}

func doctorHasWarning(warnings []kit.Status, path, state string) bool {
	for _, w := range warnings {
		if w.Path == path && w.State == state {
			return true
		}
	}
	return false
}

func TestCountsSplitsStatesIntoOkAttentionAndPending(t *testing.T) {
	report := Report{Items: []kit.Status{
		{Path: "a", State: "ok"},
		{Path: "b", State: "alterado"},
		{Path: "c", State: "atualizar"},
		{Path: "d", State: "pendente"},
		{Path: "e", State: "sem-historico"},
		{Path: "f", State: "bloqueado"},
	}}
	ok, attention, pending := report.Counts()
	if ok != 1 || attention != 3 || pending != 2 {
		t.Fatalf("contagem errada: ok=%d atenção=%d pendente=%d", ok, attention, pending)
	}
}

func TestExitCodeOnlyFailsForKitBlockingStates(t *testing.T) {
	cases := map[string]int{"ok": 0, "pendente": 0, "atualizar": 0, "sem-historico": 0, "alterado": 1, "bloqueado": 1}
	for state, want := range cases {
		report := Report{Items: []kit.Status{{Path: "x", State: state}}}
		if got := report.ExitCode(); got != want {
			t.Fatalf("estado %s: código %d, esperado %d", state, got, want)
		}
	}
}

func TestDoctorAgentNotManagedWithoutPathIsWarning(t *testing.T) {
	doctorIsolate(t)
	t.Setenv("PATH", "")
	home := t.TempDir()
	report, err := Doctor(home, nil, installExecutable(t))
	if err != nil {
		t.Fatalf("doctor falhou: %v", err)
	}
	if !doctorHasWarning(report.Warnings, "agente claude", "aviso") {
		t.Fatalf("agente não gerenciado deveria ser aviso: %v", report.Warnings)
	}
	if report.ExitCode() != 0 {
		t.Fatalf("agente não gerenciado não pode mudar o código de saída: %d", report.ExitCode())
	}
}

func TestDoctorAlteredKitFileExitsOne(t *testing.T) {
	doctorIsolate(t)
	t.Setenv("PATH", "")
	r := installRequest(t)
	t.Setenv("HOME", r.Home)
	plan := installPlan(t, r)
	if _, err := ApplyInstall(plan, nil); err != nil {
		t.Fatalf("instalação de teste falhou: %v", err)
	}
	if err := os.WriteFile(filepath.Join(r.Home, ".claude", "CLAUDE.md"), []byte("editado"), 0600); err != nil {
		t.Fatalf("edição de teste falhou: %v", err)
	}
	report, err := Doctor(r.Home, nil, r.Executable)
	if err != nil {
		t.Fatalf("doctor falhou: %v", err)
	}
	if got := doctorFind(t, report.Items, ".claude/CLAUDE.md").State; got != "alterado" {
		t.Fatalf("estado do arquivo editado: %s", got)
	}
	if report.ExitCode() != 1 {
		t.Fatalf("arquivo alterado deveria sair 1, saiu %d", report.ExitCode())
	}
}

func TestDoctorStableExecutableDiffersIsUpdate(t *testing.T) {
	doctorIsolate(t)
	home := t.TempDir()
	stable := catalog.StableBinary(home)
	if err := os.MkdirAll(filepath.Dir(stable), 0755); err != nil {
		t.Fatalf("pasta de teste não criada: %v", err)
	}
	if err := os.WriteFile(stable, []byte("versao antiga"), 0755); err != nil {
		t.Fatalf("executável de teste não gravado: %v", err)
	}
	report, err := Doctor(home, nil, installExecutable(t))
	if err != nil {
		t.Fatalf("doctor falhou: %v", err)
	}
	item := doctorFind(t, report.Items, ".local/share/gex-dev-kit/bin/"+filepath.Base(stable))
	if item.State != "atualizar" || !strings.Contains(item.Detail, "gex-dev-kit install") {
		t.Fatalf("item de atualização errado: %+v", item)
	}
	if report.ExitCode() != 0 {
		t.Fatalf("atualização pendente não sai 1: %d", report.ExitCode())
	}
}

func TestDoctorWarnsConfigDirOutsideHome(t *testing.T) {
	doctorIsolate(t)
	t.Setenv("PATH", "")
	home := t.TempDir()
	outside := t.TempDir()
	report, err := Doctor(home, []string{"CLAUDE_CONFIG_DIR=" + outside}, installExecutable(t))
	if err != nil {
		t.Fatalf("doctor falhou: %v", err)
	}
	if !doctorHasWarning(report.Warnings, "CLAUDE_CONFIG_DIR", "aviso") {
		t.Fatalf("CLAUDE_CONFIG_DIR fora do HOME deveria avisar: %v", report.Warnings)
	}
}

func TestDoctorListsEffectiveOrgsAndInvalidLine(t *testing.T) {
	doctorIsolate(t)
	t.Setenv("PATH", "")
	configDir, err := os.UserConfigDir()
	if err != nil {
		t.Fatalf("pasta de configuração indisponível: %v", err)
	}
	dir := filepath.Join(configDir, "gex-dev-kit")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("pasta de teste não criada: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "orgs"), []byte("Minha-Org\nlinha invalida\n"), 0644); err != nil {
		t.Fatalf("arquivo orgs não gravado: %v", err)
	}
	report, err := Doctor(t.TempDir(), nil, installExecutable(t))
	if err != nil {
		t.Fatalf("doctor falhou: %v", err)
	}
	joined := ""
	for _, w := range report.Warnings {
		joined += w.Path + " " + w.Detail + "\n"
	}
	if !strings.Contains(joined, "Minha-Org") || !strings.Contains(joined, "Gex-Development") {
		t.Fatalf("organizações efetivas ausentes: %s", joined)
	}
	if !strings.Contains(joined, "linha 2") {
		t.Fatalf("linha inválida não avisada: %s", joined)
	}
}
