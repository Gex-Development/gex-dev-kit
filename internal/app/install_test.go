package app

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func installExecutable(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "gex-dev-kit")
	if err := os.WriteFile(path, []byte("binario de teste"), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func installRequest(t *testing.T) Request {
	t.Helper()
	return Request{Home: t.TempDir(), Executable: installExecutable(t), SkipAgentUpdate: true}
}

func installPlan(t *testing.T, r Request) Plan {
	t.Helper()
	plan, err := PlanInstall(r, 0)
	if err != nil {
		t.Fatalf("prévia falhou: %v", err)
	}
	return plan
}

func installNoticeTexts(notices []Notice) string {
	var texts []string
	for _, notice := range notices {
		texts = append(texts, notice.Text)
	}
	return strings.Join(texts, "\n")
}

func TestPlanInstallRejectsRelativeHome(t *testing.T) {
	r := installRequest(t)
	r.Home = "relativo"
	if _, err := PlanInstall(r, 0); err == nil || !strings.Contains(err.Error(), "caminho absoluto") {
		t.Fatalf("HOME relativo aceito: %v", err)
	}
}

func TestPlanInstallRejectsRemovedSkillBeforeWriting(t *testing.T) {
	r := installRequest(t)
	r.Skills = []string{"gex-obsoleta-api"}
	_, err := PlanInstall(r, 0)
	if err == nil || !strings.Contains(err.Error(), "gex-obsoleta-api") {
		t.Fatalf("skill desconhecida aceita: %v", err)
	}
	if _, err := os.Stat(filepath.Join(r.Home, ".claude", "CLAUDE.md")); !os.IsNotExist(err) {
		t.Fatalf("prévia alterou HOME com skill desconhecida: %v", err)
	}
}

func TestPlanInstallIsEmptyAfterApply(t *testing.T) {
	r := installRequest(t)
	plan := installPlan(t, r)
	if plan.Empty() {
		t.Fatal("primeira prévia sem alterações")
	}
	if _, err := ApplyInstall(plan, nil); err != nil {
		t.Fatalf("aplicação falhou: %v", err)
	}
	if again := installPlan(t, r); !again.Empty() {
		t.Fatalf("segunda prévia deveria estar vazia: %+v", again.Files)
	}
}

func TestApplyInstallReportsStepsInOrder(t *testing.T) {
	r := installRequest(t)
	plan := installPlan(t, r)
	var steps []Step
	result, err := ApplyInstall(plan, func(step Step) { steps = append(steps, step) })
	if err != nil {
		t.Fatalf("aplicação falhou: %v", err)
	}
	if want := []Step{StepAgents, StepFiles, StepDone}; !reflect.DeepEqual(steps, want) {
		t.Fatalf("etapas %v, esperado %v", steps, want)
	}
	if result.HistoryID == "" {
		t.Fatal("histórico não informado")
	}
	if result.HooksNeedTrust {
		t.Fatal("sem hook selecionado não deve exigir confiança")
	}
}

func TestApplyInstallSignalsCodexTrustForSelectedHook(t *testing.T) {
	r := installRequest(t)
	r.Hooks = []string{"gex-conventional-commits"}
	plan := installPlan(t, r)
	if !strings.Contains(installNoticeTexts(plan.Notices), "aprove em /hooks") {
		t.Fatalf("aviso de aprovação ausente: %+v", plan.Notices)
	}
	result, err := ApplyInstall(plan, nil)
	if err != nil {
		t.Fatalf("aplicação falhou: %v", err)
	}
	if !result.HooksNeedTrust {
		t.Fatal("hook com suporte Codex deve exigir aprovação em /hooks")
	}
}

func TestApplyInstallFailsWhenFileChangedAfterPreview(t *testing.T) {
	r := installRequest(t)
	plan := installPlan(t, r)
	if err := os.MkdirAll(filepath.Join(r.Home, ".claude"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(r.Home, ".claude", "CLAUDE.md"), []byte("alterado depois da prévia"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyInstall(plan, nil); err == nil || !strings.Contains(err.Error(), "Instalação do Core") {
		t.Fatalf("prévia desatualizada aplicada: %v", err)
	}
}

func TestInstallNoticesDescribeClaudeAndCodexSupport(t *testing.T) {
	notices := installNoticeTexts(installNotices([]string{"gex-sensitive-action-warning"}))
	for _, want := range []string{"pergunta antes", "claude -p", "Codex: não suportado"} {
		if !strings.Contains(notices, want) {
			t.Fatalf("aviso sem %q: %s", want, notices)
		}
	}
}

func TestPlanInstallReportsAgentsWithoutPath(t *testing.T) {
	t.Setenv("PATH", "")
	r := installRequest(t)
	t.Setenv("HOME", r.Home)
	t.Setenv("USERPROFILE", r.Home)
	r.SkipAgentUpdate = false
	plan := installPlan(t, r)
	if len(plan.Agents) != 2 {
		t.Fatalf("agentes %d, esperado 2: %+v", len(plan.Agents), plan.Agents)
	}
}
