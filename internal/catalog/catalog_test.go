package catalog

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func lookup(t *testing.T, files []File, path string) File {
	t.Helper()
	for _, file := range files {
		if file.Path == path {
			return file
		}
	}
	t.Fatalf("arquivo ausente: %s", path)
	return File{}
}

func TestCorePreservesGlobalInstructionsAndDoesNotInstallOptional(t *testing.T) {
	home := t.TempDir()
	original := []byte("# Preferências pessoais\n\nNão altere meu modelo.\n")
	if err := os.Mkdir(filepath.Join(home, ".claude"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".claude", "CLAUDE.md"), original, 0600); err != nil {
		t.Fatal(err)
	}
	files, err := Build(home, "", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	claude := lookup(t, files, ".claude/CLAUDE.md")
	codex := lookup(t, files, ".codex/AGENTS.md")
	if !strings.HasPrefix(string(claude.Data), string(original)) || !strings.Contains(string(claude.Data), "# Instruções globais de desenvolvimento GEX") || !strings.Contains(string(codex.Data), "# Instruções globais de desenvolvimento GEX") {
		t.Fatal("instruções globais incompletas ou preferências substituídas")
	}
	for _, file := range files {
		if strings.Contains(file.Path, "/tdd/") || strings.Contains(file.Path, "hooks.json") || strings.Contains(file.Path, "settings.json") {
			t.Fatalf("opcional instalado sem seleção: %s", file.Path)
		}
	}
}

func TestCoreAlwaysIncludesSevenSkillsInBothAgents(t *testing.T) {
	want := []string{
		"gex-abrir-issue", "gex-auto-revisao", "gex-fluxo-desenvolvimento",
		"gex-preparar-pr", "gex-regras-de-negocio", "gex-revisar-pr", "gex-verificar-mudanca",
	}
	if !slices.Equal(CoreSkills(), want) {
		t.Fatalf("lista Core inesperada: %v", CoreSkills())
	}
	files, err := Build(t.TempDir(), "", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range want {
		for _, prefix := range []string{".claude/skills/", ".agents/skills/"} {
			file := lookup(t, files, prefix+name+"/SKILL.md")
			if !strings.Contains(string(file.Data), "name: "+name) {
				t.Fatalf("skill Core incompleta: %s", file.Path)
			}
		}
	}
	for _, file := range files {
		if strings.Contains(file.Path, "/tdd/") || strings.Contains(file.Path, "hooks.json") || strings.Contains(file.Path, "settings.json") || slices.ContainsFunc(SpecialtySkills(), func(name string) bool { return strings.Contains(file.Path, "/"+name+"/") }) {
			t.Fatalf("opcional instalado sem seleção: %s", file.Path)
		}
	}
}

func TestSelectedSpecialtyCopiesCompletePackOnlyWhenSelected(t *testing.T) {
	packs := map[string]string{
		"gex-buygoods-api":             "scripts/metadata.py",
		"gex-canonical-prod-readiness": "",
		"gex-clickbank-api":            "scripts/quickstats.py",
		"gex-design-system":            "",
		"gex-gateway-audit":            "",
		"gex-hub-design-system":        "",
		"gex-hub-gold-refresh-prod":    "",
		"gex-local-spark-tests":        "",
		"gex-pagamerican-api":          "scripts/metadata.py",
		"gex-pr-review":                "",
	}
	if !slices.Equal(SpecialtySkills(), []string{"gex-buygoods-api", "gex-canonical-prod-readiness", "gex-clickbank-api", "gex-design-system", "gex-gateway-audit", "gex-hub-design-system", "gex-hub-gold-refresh-prod", "gex-local-spark-tests", "gex-pagamerican-api", "gex-pr-review"}) {
		t.Fatalf("especializações inesperadas: %v", SpecialtySkills())
	}
	for name, script := range packs {
		if SpecialtyDescription(name) == "" || SkillDescription(name) == "" {
			t.Fatalf("especialização sem descrição: %s", name)
		}
		specialty, _ := SpecialtyByName(name)
		files, err := Build(t.TempDir(), "", append([]string{name}, specialty.Requires...), nil)
		if err != nil {
			t.Fatal(err)
		}
		for _, prefix := range []string{".claude/skills/", ".agents/skills/"} {
			skill := lookup(t, files, prefix+name+"/SKILL.md")
			if !strings.Contains(string(skill.Data), "name: "+name) {
				t.Fatalf("especialização incompleta: %s", skill.Path)
			}
			if script != "" && len(lookup(t, files, prefix+name+"/"+script).Data) == 0 {
				t.Fatalf("script obrigatório ausente: %s", name)
			}
		}
		for _, file := range files {
			for other := range packs {
				if other != name && !slices.Contains(specialty.Requires, other) && strings.Contains(file.Path, "/"+other+"/") {
					t.Fatalf("especialização não escolhida: %s", file.Path)
				}
			}
		}
	}
}

func TestUnauditedHookCandidatesStayUnavailable(t *testing.T) {
	for _, name := range []string{"smart-commit", "tdd-gate", "claude/smart-commit", "codex/tdd-gate"} {
		if err := ValidateSelection(nil, []string{name}); err == nil {
			t.Fatalf("hook sem instalador validado aceito: %s", name)
		}
	}
}

func TestReferenceCatalogDoesNotPromiseInstallation(t *testing.T) {
	for _, entry := range append(ReferencedSkills(), HookCandidates()...) {
		if entry.Name == "" || entry.Status == "" || entry.Reason == "" || entry.Source == "" {
			t.Fatalf("referência incompleta: %+v", entry)
		}
		if entry.Status != "por função" && entry.Status != "somente referência" {
			t.Fatalf("status não instalável inesperado: %+v", entry)
		}
		if err := ValidateSelection([]string{entry.Name}, nil); err == nil {
			t.Fatalf("referência aceita como skill instalável: %s", entry.Name)
		}
		if err := ValidateSelection(nil, []string{entry.Name}); err == nil {
			t.Fatalf("referência aceita como hook instalável: %s", entry.Name)
		}
	}
}

func TestAllThirdPartyHookTemplatesListedWithoutInstallability(t *testing.T) {
	count := map[string]int{}
	for _, entry := range HookCandidates() {
		for _, prefix := range []string{"claude/", "codex/"} {
			if strings.HasPrefix(entry.Name, prefix) {
				count[prefix]++
				if !strings.Contains(entry.Source, "/tree/") || entry.Status != "somente referência" {
					t.Fatalf("template sem fonte fixada ou status correto: %+v", entry)
				}
			}
		}
	}
	if count["claude/"] != 13 || count["codex/"] != 12 {
		t.Fatalf("inventário incompleto: %v", count)
	}
}

func TestSelectedSkillIncludesOfflineReferencesAndLicense(t *testing.T) {
	files, err := Build(t.TempDir(), "", []string{"tdd"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{".claude/skills/tdd/SKILL.md", ".claude/skills/tdd/tests.md", ".claude/skills/tdd/mocking.md", ".claude/skills/tdd/LICENSE", ".agents/skills/tdd/SKILL.md"} {
		if len(lookup(t, files, path).Data) == 0 {
			t.Fatal("arquivo sem conteúdo:", path)
		}
	}
	for _, file := range files {
		if strings.Contains(file.Path, "/diagnosing-bugs/") {
			t.Fatal("skill não selecionada:", file.Path)
		}
	}
}

func TestHookPreservesOtherHandlersAndPreferences(t *testing.T) {
	home := t.TempDir()
	for _, dir := range []string{".claude", ".codex"} {
		if err := os.Mkdir(filepath.Join(home, dir), 0700); err != nil {
			t.Fatal(err)
		}
	}
	claude := []byte(`{"model":"custom","hooks":{"PreToolUse":[{"matcher":"Bash","hooks":[{"type":"command","command":"echo existing"}]}],"Stop":[{"hooks":[{"type":"command","command":"echo stop"}]}]}}`)
	codex := []byte(`{"description":"personal","hooks":{"Stop":[{"hooks":[{"type":"command","command":"echo stop"}]}]}}`)
	for path, data := range map[string][]byte{".claude/settings.json": claude, ".codex/hooks.json": codex} {
		if err := os.WriteFile(filepath.Join(home, path), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	executable := filepath.Join(home, "gex-dev-kit")
	if err := os.WriteFile(executable, []byte("binary"), 0700); err != nil {
		t.Fatal(err)
	}
	files, err := Build(home, executable, nil, []string{"gex-conventional-commits"})
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{".claude/settings.json", ".codex/hooks.json"} {
		var config struct {
			Model       string                       `json:"model"`
			Description string                       `json:"description"`
			Hooks       map[string][]json.RawMessage `json:"hooks"`
		}
		if err := json.Unmarshal(lookup(t, files, path).Data, &config); err != nil {
			t.Fatal(err)
		}
		if len(config.Hooks["Stop"]) != 1 || len(config.Hooks["PreToolUse"]) == 0 {
			t.Fatalf("hooks alheios removidos: %s", path)
		}
		if path == ".claude/settings.json" && config.Model != "custom" || path == ".codex/hooks.json" && config.Description != "personal" {
			t.Fatal("preferência removida:", path)
		}
		if path == ".claude/settings.json" && !strings.Contains(string(lookup(t, files, path).Data), `"onFailure": "block"`) {
			t.Fatal("hook Claude deve bloquear falha do handler")
		}
	}
	if _, err := Build(home, "", nil, []string{"hook-inexistente"}); err == nil {
		t.Fatal("hook desconhecido aceito")
	}
}

func TestDependentSkillsNeedExplicitSelection(t *testing.T) {
	for _, chosen := range [][]string{{"grill-me"}, {"grill-with-docs", "grilling"}, {"improve-codebase-architecture", "codebase-design", "grilling"}} {
		if _, err := Build(t.TempDir(), "", chosen, nil); err == nil {
			t.Fatalf("dependência ausente aceita: %v", chosen)
		}
	}
	files, err := Build(t.TempDir(), "", []string{"grill-with-docs", "grilling", "domain-modeling"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(lookup(t, files, ".claude/skills/grill-with-docs/SKILL.md").Data) == 0 {
		t.Fatal("skill com dependências selecionadas indisponível")
	}
}

func TestSelectionValidationRejectsCrossGroupMissingDependency(t *testing.T) {
	if err := ValidateSelection([]string{"grill-with-docs", "domain-modeling"}, nil); err == nil || !strings.Contains(err.Error(), "grilling") {
		t.Fatalf("skill dependente sem produtividade aceita: %v", err)
	}
	if err := ValidateSelection([]string{"grill-with-docs", "domain-modeling", "grilling"}, nil); err != nil {
		t.Fatalf("dependências presentes rejeitadas: %v", err)
	}
}

func TestEngineeringStepChecksOnlyDependenciesAvailableInThatStep(t *testing.T) {
	if err := ValidateEngineeringSelection([]string{"grill-with-docs", "domain-modeling"}); err != nil {
		t.Fatalf("grilling é escolhida na próxima etapa: %v", err)
	}
	if err := ValidateEngineeringSelection([]string{"improve-codebase-architecture", "domain-modeling"}); err == nil || !strings.Contains(err.Error(), "codebase-design") {
		t.Fatalf("dependência disponível na etapa atual foi ignorada: %v", err)
	}
}

func TestHookKeepsOtherHandlersInSameMatcher(t *testing.T) {
	executable := filepath.Join(t.TempDir(), "gex-dev-kit")
	original := []byte(`{"hooks":{"PreToolUse":[{"matcher":"Bash","hooks":[{"type":"command","command":"OWN"},{"type":"command","command":"echo own"},{"type":"command","command":"/elsewhere/command"}]}]}}`)
	original = bytes.Replace(original, []byte("OWN"), []byte(executable), 1)
	selected := selectedHookSpecs([]string{"gex-conventional-commits"})
	merged, err := mergeHooks(original, executable, AgentClaude, selected)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(merged), "echo own") || !strings.Contains(string(merged), "/elsewhere/command") {
		t.Fatal("handler alheio perdido")
	}
	again, err := mergeHooks(merged, executable, AgentClaude, selected)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(again), `"onFailure": "block"`) != 1 {
		t.Fatal("hook duplicado")
	}
}

type catalogHookHandler struct {
	Type           string   `json:"type"`
	Command        string   `json:"command"`
	Args           []string `json:"args"`
	OnFailure      string   `json:"onFailure"`
	Timeout        int      `json:"timeout"`
	CommandWindows string   `json:"commandWindows"`
}

type catalogHookGroup struct {
	Matcher string               `json:"matcher"`
	Hooks   []catalogHookHandler `json:"hooks"`
}

func catalogHookGroups(t *testing.T, files []File, path string) map[string][]catalogHookGroup {
	t.Helper()
	var config struct {
		Hooks map[string][]catalogHookGroup `json:"hooks"`
	}
	if err := json.Unmarshal(lookup(t, files, path).Data, &config); err != nil {
		t.Fatal(err)
	}
	return config.Hooks
}

func TestAgentHooksAreFourAndAvailable(t *testing.T) {
	want := []string{"gex-conventional-commits", "gex-sensitive-action-warning", "gex-pre-pr", "gex-session-start"}
	if !slices.Equal(Hooks(), want) {
		t.Fatalf("hooks de agente inesperados: %v", Hooks())
	}
	if err := ValidateSelection(nil, want); err != nil {
		t.Fatalf("hooks de agente rejeitados: %v", err)
	}
	for _, name := range want {
		if HookDescription(name) == "" {
			t.Fatalf("hook sem descrição: %s", name)
		}
	}
	if len(HookCandidates()) != 25 {
		t.Fatalf("candidatos de terceiros inesperados: %d", len(HookCandidates()))
	}
	for _, entry := range HookCandidates() {
		if strings.HasPrefix(entry.Name, "gex-") {
			t.Fatalf("hook do GEX ainda listado como candidato: %s", entry.Name)
		}
	}
}

func TestIsKitHandlerRecognizesClaudeAndCodexForms(t *testing.T) {
	executable := filepath.Join(t.TempDir(), "gex-dev-kit")
	cases := []struct {
		name    string
		command string
		args    []string
		sub     string
		ok      bool
	}{
		{"claude com subcomando", executable, []string{"hook", "sensitive-action"}, "sensitive-action", true},
		{"claude legado sem args", executable, nil, "commit-msg", true},
		{"codex sem aspas", executable + " hook pre-pr", nil, "pre-pr", true},
		{"codex com aspas e flag", `"` + executable + `" hook session-start --agent codex`, nil, "session-start", true},
		{"claude sem subcomando", executable, []string{"hook"}, "", false},
		{"claude de outro executável", executable + "-outro", []string{"hook", "commit-msg"}, "", false},
		{"comando alheio", "echo own", nil, "", false},
	}
	for _, tc := range cases {
		sub, ok := IsKitHandler(tc.command, tc.args, executable)
		if sub != tc.sub || ok != tc.ok {
			t.Fatalf("%s: obtido (%q, %v), esperado (%q, %v)", tc.name, sub, ok, tc.sub, tc.ok)
		}
	}
}

func TestMergeHooksIsIdempotentForBothAgents(t *testing.T) {
	executable := filepath.Join(t.TempDir(), "gex-dev-kit")
	selected := selectedHookSpecs(Hooks())
	for _, agent := range []Agent{AgentClaude, AgentCodex} {
		first, err := mergeHooks(nil, executable, agent, selected)
		if err != nil {
			t.Fatal(err)
		}
		second, err := mergeHooks(first, executable, agent, selected)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(first, second) {
			t.Fatalf("merge não idempotente no agente %d:\n%s\n---\n%s", agent, first, second)
		}
	}
}

func TestMergeHooksKeepsForeignAndOtherKitHandlers(t *testing.T) {
	executable := filepath.Join(t.TempDir(), "gex-dev-kit")
	encoded, err := json.Marshal(executable)
	if err != nil {
		t.Fatal(err)
	}
	previous := []byte(`{"hooks":{"PreToolUse":[{"matcher":"Bash","hooks":[{"type":"command","command":"echo own && echo more"},{"type":"command","command":` + string(encoded) + `,"args":["hook","pre-pr"]}]}]}}`)
	merged, err := mergeHooks(previous, executable, AgentClaude, selectedHookSpecs([]string{"gex-conventional-commits"}))
	if err != nil {
		t.Fatal(err)
	}
	text := string(merged)
	if !strings.Contains(text, "echo own && echo more") || strings.Contains(text, `\u0026`) {
		t.Fatalf("handler alheio alterado:\n%s", text)
	}
	if strings.Count(text, `"pre-pr"`) != 1 {
		t.Fatalf("handler de outro hook do kit removido:\n%s", text)
	}
	if strings.Count(text, `"onFailure": "block"`) != 1 {
		t.Fatalf("handler commit-msg ausente ou duplicado:\n%s", text)
	}
}

func TestMergeHooksReplacesLegacyCommitMsgWithoutArgs(t *testing.T) {
	executable := filepath.Join(t.TempDir(), "gex-dev-kit")
	encoded, err := json.Marshal(executable)
	if err != nil {
		t.Fatal(err)
	}
	previous := []byte(`{"hooks":{"PreToolUse":[{"matcher":"Bash","hooks":[{"type":"command","command":` + string(encoded) + `}]}]}}`)
	merged, err := mergeHooks(previous, executable, AgentClaude, selectedHookSpecs([]string{"gex-conventional-commits"}))
	if err != nil {
		t.Fatal(err)
	}
	text := string(merged)
	if strings.Count(text, `"args"`) != 1 || strings.Count(text, `"commit-msg"`) != 1 {
		t.Fatalf("handler legado não substituído:\n%s", text)
	}
}

func TestMergeHooksDropsEmptyGroupAfterReplacing(t *testing.T) {
	executable := filepath.Join(t.TempDir(), "gex-dev-kit")
	encoded, err := json.Marshal(executable)
	if err != nil {
		t.Fatal(err)
	}
	previous := []byte(`{"hooks":{"PreToolUse":[{"matcher":"Bash|PowerShell","hooks":[{"type":"command","command":` + string(encoded) + `,"args":["hook","pre-pr"]}]}]}}`)
	merged, err := mergeHooks(previous, executable, AgentClaude, selectedHookSpecs([]string{"gex-pre-pr"}))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(merged), `"matcher"`) != 1 {
		t.Fatalf("grupo vazio não foi removido:\n%s", merged)
	}
}

func TestMergeHooksRejectsMalformedJSONWithAction(t *testing.T) {
	executable := filepath.Join(t.TempDir(), "gex-dev-kit")
	inputs := []string{`{`, `[]`, `{"hooks":[]}`, `{"hooks":{"PreToolUse":{}}}`, `{"hooks":{"PreToolUse":[{"hooks":"x"}]}}`}
	for _, input := range inputs {
		_, err := mergeHooks([]byte(input), executable, AgentClaude, selectedHookSpecs([]string{"gex-conventional-commits"}))
		if err == nil || !strings.Contains(err.Error(), "corrija o JSON; nada foi alterado") {
			t.Fatalf("entrada inválida aceita ou sem ação (%s): %v", input, err)
		}
	}
}

func TestBuildReportsInvalidSettingsWithPath(t *testing.T) {
	home := t.TempDir()
	if err := os.Mkdir(filepath.Join(home, ".claude"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".claude", "settings.json"), []byte("{"), 0600); err != nil {
		t.Fatal(err)
	}
	executable := filepath.Join(t.TempDir(), "gex-dev-kit")
	if err := os.WriteFile(executable, []byte("binary"), 0700); err != nil {
		t.Fatal(err)
	}
	_, err := Build(home, executable, nil, []string{"gex-conventional-commits"})
	if err == nil || !strings.Contains(err.Error(), ".claude/settings.json") || !strings.Contains(err.Error(), "nada foi alterado") {
		t.Fatalf("JSON inválido sem caminho e ação: %v", err)
	}
}

func TestWindowsCommandPathQuotesOnlyAsLastResort(t *testing.T) {
	shortened := func(string) (string, error) { return `C:\PROGRA~1\gex\gex-dev-kit.exe`, nil }
	failing := func(string) (string, error) { return "", errors.New("falha") }
	stillSpaced := func(string) (string, error) { return `C:\Users\Dev Team\gex-dev-kit.exe`, nil }
	cases := []struct {
		name       string
		executable string
		short      func(string) (string, error)
		want       string
	}{
		{"sem espaço não cita", `C:\gex\gex-dev-kit.exe`, failing, `C:\gex\gex-dev-kit.exe`},
		{"com espaço usa caminho curto", `C:\Users\Dev Team\gex-dev-kit.exe`, shortened, `C:\PROGRA~1\gex\gex-dev-kit.exe`},
		{"com espaço sem caminho curto cita", `C:\Users\Dev Team\gex-dev-kit.exe`, failing, `"C:\Users\Dev Team\gex-dev-kit.exe"`},
		{"caminho curto ainda com espaço cita", `C:\Users\Dev Team\gex-dev-kit.exe`, stillSpaced, `"C:\Users\Dev Team\gex-dev-kit.exe"`},
		{"sem função de caminho curto cita", `C:\Users\Dev Team\gex-dev-kit.exe`, nil, `"C:\Users\Dev Team\gex-dev-kit.exe"`},
	}
	for _, tc := range cases {
		if got := windowsCommandPath(tc.executable, tc.short); got != tc.want {
			t.Fatalf("%s: obtido %s, esperado %s", tc.name, got, tc.want)
		}
	}
}

func TestBuildAddsAllAgentHandlers(t *testing.T) {
	home := t.TempDir()
	executable := filepath.Join(t.TempDir(), "gex-dev-kit")
	if err := os.WriteFile(executable, []byte("binary"), 0700); err != nil {
		t.Fatal(err)
	}
	files, err := Build(home, executable, nil, Hooks())
	if err != nil {
		t.Fatal(err)
	}
	stable := StableBinary(home)

	claude := catalogHookGroups(t, files, ".claude/settings.json")
	if len(claude["PreToolUse"]) != 3 || len(claude["SessionStart"]) != 1 {
		t.Fatalf("grupos Claude inesperados: %+v", claude)
	}
	commit := claude["PreToolUse"][0]
	if commit.Matcher != "Bash" || commit.Hooks[0].Command != stable || !slices.Equal(commit.Hooks[0].Args, []string{"hook", "commit-msg"}) || commit.Hooks[0].OnFailure != "block" {
		t.Fatalf("handler commit-msg Claude inesperado: %+v", commit)
	}
	sensitive := claude["PreToolUse"][1]
	if sensitive.Matcher != "Bash|PowerShell|Read" || !slices.Equal(sensitive.Hooks[0].Args, []string{"hook", "sensitive-action"}) || sensitive.Hooks[0].OnFailure != "" {
		t.Fatalf("handler sensitive-action Claude inesperado: %+v", sensitive)
	}
	session := claude["SessionStart"][0]
	if session.Matcher != "startup|resume|clear|compact|fork" || !slices.Equal(session.Hooks[0].Args, []string{"hook", "session-start", "--agent", "claude"}) {
		t.Fatalf("handler session-start Claude inesperado: %+v", session)
	}

	codex := catalogHookGroups(t, files, ".codex/hooks.json")
	if len(codex["PreToolUse"]) != 1 || len(codex["SessionStart"]) != 1 || len(codex["PreToolUse"][0].Hooks) != 1 {
		t.Fatalf("grupos Codex inesperados: %+v", codex)
	}
	if got := codex["PreToolUse"][0].Hooks[0].Command; got != quoteCommand(stable)+" hook commit-msg" {
		t.Fatalf("comando Codex commit-msg inesperado: %s", got)
	}
	if got := codex["SessionStart"][0].Hooks[0].Command; got != quoteCommand(stable)+" hook session-start --agent codex" {
		t.Fatalf("comando Codex session-start inesperado: %s", got)
	}
	if strings.Contains(string(lookup(t, files, ".codex/hooks.json").Data), "sensitive-action") {
		t.Fatal("Codex recebeu hook sem suporte")
	}
}

func TestSensitiveOnlyLeavesCodexConfigOut(t *testing.T) {
	home := t.TempDir()
	executable := filepath.Join(t.TempDir(), "gex-dev-kit")
	if err := os.WriteFile(executable, []byte("binary"), 0700); err != nil {
		t.Fatal(err)
	}
	files, err := Build(home, executable, nil, []string{"gex-sensitive-action-warning"})
	if err != nil {
		t.Fatal(err)
	}
	lookup(t, files, ".claude/settings.json")
	for _, file := range files {
		if file.Path == ".codex/hooks.json" {
			t.Fatal("arquivo do Codex no plano sem hook suportado")
		}
	}
}

func TestBinaryRefreshedOnlyWithKitHandlerOrChangedBinary(t *testing.T) {
	binaryPath := ".local/share/gex-dev-kit/bin/" + catalogBinaryName()
	hasBinary := func(files []File) bool {
		for _, file := range files {
			if file.Path == binaryPath {
				return true
			}
		}
		return false
	}
	home := t.TempDir()
	executable := filepath.Join(t.TempDir(), "gex-dev-kit")
	if err := os.WriteFile(executable, []byte("binary-v2"), 0700); err != nil {
		t.Fatal(err)
	}
	files, err := Build(home, executable, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if hasBinary(files) {
		t.Fatal("binário incluído sem hook, handler do kit ou diferença")
	}
	if err := os.MkdirAll(filepath.Dir(StableBinary(home)), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(StableBinary(home), []byte("binary-v1"), 0700); err != nil {
		t.Fatal(err)
	}
	files, err = Build(home, executable, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !hasBinary(files) {
		t.Fatal("binário estável diferente não atualizado")
	}
	if err := os.WriteFile(StableBinary(home), []byte("binary-v2"), 0700); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(StableBinary(home))
	if err != nil {
		t.Fatal(err)
	}
	settings := []byte(`{"hooks":{"PreToolUse":[{"matcher":"Bash","hooks":[{"type":"command","command":` + string(encoded) + `,"args":["hook","pre-pr"]}]}]}}`)
	if err := os.MkdirAll(filepath.Join(home, ".claude"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".claude", "settings.json"), settings, 0600); err != nil {
		t.Fatal(err)
	}
	files, err = Build(home, executable, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !hasBinary(files) {
		t.Fatal("handler do kit instalado sem atualização do executável")
	}
}

func TestBuildRejectsSymlinkedSettings(t *testing.T) {
	home := t.TempDir()
	if err := os.Mkdir(filepath.Join(home, ".claude"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(home, "elsewhere"), filepath.Join(home, ".claude", "CLAUDE.md")); err != nil {
		t.Fatal(err)
	}
	if _, err := Build(home, "", nil, nil); err == nil {
		t.Fatal("link simbólico aceito")
	}
}

func TestRetroIsInstallableOnlyWithWritingForAgents(t *testing.T) {
	if len(Skills()) != 18 || !slices.Contains(Skills(), "retro") {
		t.Fatalf("catálogo Matt inesperado: %d skills", len(Skills()))
	}
	if !strings.Contains(SkillDescription("retro"), "invocação explícita") {
		t.Fatalf("descrição de retro sem aviso de logs: %s", SkillDescription("retro"))
	}
	if err := ValidateSelection([]string{"retro"}, nil); err == nil || !strings.Contains(err.Error(), "writing-for-agents") {
		t.Fatalf("retro aceita sem writing-for-agents: %v", err)
	}
	if err := ValidateEngineeringSelection([]string{"retro"}); err != nil {
		t.Fatalf("retro rejeitada na etapa de engenharia: %v", err)
	}
	files, err := Build(t.TempDir(), "", []string{"retro", "writing-for-agents"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	lookup(t, files, ".claude/skills/retro/SKILL.md")
	lookup(t, files, ".agents/skills/retro/agents/openai.yaml")
}

func TestReferencedSkillsListsElevenEntriesAtPinnedRevision(t *testing.T) {
	refs := ReferencedSkills()
	if len(refs) != 11 {
		t.Fatalf("referências inesperadas: %d", len(refs))
	}
	for _, name := range []string{"wizard", "to-spec", "to-tickets", "triage", "wayfinder", "implement"} {
		var found bool
		for _, ref := range refs {
			if ref.Name == name {
				found = true
				if !strings.HasPrefix(ref.Source, "https://github.com/mattpocock/skills/tree/"+MattRevision+"/skills/engineering/") {
					t.Fatalf("referência sem commit fixado: %s", ref.Source)
				}
			}
		}
		if !found {
			t.Fatalf("referência ausente: %s", name)
		}
	}
}

func TestStableBinaryAndBinaryFileUseStablePath(t *testing.T) {
	home := t.TempDir()
	executable := filepath.Join(t.TempDir(), "gex-dev-kit")
	if err := os.WriteFile(executable, []byte("binary"), 0700); err != nil {
		t.Fatal(err)
	}
	file, err := BinaryFile(executable)
	if err != nil {
		t.Fatal(err)
	}
	name := catalogBinaryName()
	if file.Path != ".local/share/gex-dev-kit/bin/"+name || file.Mode != 0755 || string(file.Data) != "binary" {
		t.Fatalf("binário estável inesperado: %+v", file)
	}
	if StableBinary(home) != filepath.Join(home, ".local", "share", "gex-dev-kit", "bin", name) {
		t.Fatalf("caminho estável inesperado: %s", StableBinary(home))
	}
	if _, err := BinaryFile(""); err == nil {
		t.Fatal("executável vazio aceito")
	}
	if _, err := BinaryFile(filepath.Join(t.TempDir(), "ausente")); err == nil {
		t.Fatal("executável inexistente aceito")
	}
}

func TestCoreRepeatedBuildDoesNotDuplicate(t *testing.T) {
	home := t.TempDir()
	first, err := Build(home, "", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range first {
		path := file.Path
		if err := os.MkdirAll(filepath.Dir(filepath.Join(home, path)), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(home, path), file.Data, file.Mode); err != nil {
			t.Fatal(err)
		}
	}
	second, err := Build(home, "", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(second) != len(first) {
		t.Fatalf("manifest mudou sem seleção: %d != %d", len(second), len(first))
	}
	for _, file := range first {
		if !bytes.Equal(lookup(t, second, file.Path).Data, file.Data) {
			t.Fatalf("duplicação em %s", file.Path)
		}
	}
}
