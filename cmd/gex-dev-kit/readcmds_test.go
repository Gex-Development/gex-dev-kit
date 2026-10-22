package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Gex-Development/gex-dev-kit/internal/app"
	"github.com/Gex-Development/gex-dev-kit/internal/githooks"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/terminal"
)

func readTestSession(opts options, d deps) (session, *bytes.Buffer, *bytes.Buffer) {
	var out, errOut bytes.Buffer
	return session{out: &out, errOut: &errOut, opts: opts, deps: d}, &out, &errOut
}

func readFakeDeps(report app.Report) (deps, *[]string) {
	var calls []string
	return deps{
		Doctor: func(home string, environ []string, executable string) (app.Report, error) {
			calls = append(calls, "home")
			return report, nil
		},
		DoctorRepo: func(home, repo, executable string) (app.Report, error) {
			calls = append(calls, "repo")
			return report, nil
		},
	}, &calls
}

func readDriftReport() app.Report {
	return app.Report{
		Root: "home",
		Items: []app.Status{
			{Path: ".claude/AGENTS.md", State: "ok", Detail: "Íntegro."},
			{Path: ".claude/CLAUDE.md", State: "alterado", Detail: "Arquivo mudou depois da instalação."},
		},
	}
}

func readDecode(t *testing.T, raw string, value any) {
	t.Helper()
	if err := json.Unmarshal([]byte(raw), value); err != nil {
		t.Fatalf("JSON inválido: %v\nsaída: %q", err, raw)
	}
}

func TestCatalogCommandsDistinguishInstalledFromReferences(t *testing.T) {
	s, out, errOut := readTestSession(options{}, deps{})
	if code := cmdSkills(s); code != 0 || errOut.String() != "" {
		t.Fatalf("skills: exit=%d err=%q", code, errOut.String())
	}
	skills := out.String()
	if !strings.Contains(skills, "Core GEX (obrigatório):") || !strings.Contains(skills, "Referências (não instaláveis):") || strings.Contains(skills, "\x1b") {
		t.Fatalf("skills sem seções esperadas ou com ANSI: %q", skills)
	}
	s, out, errOut = readTestSession(options{}, deps{})
	if code := cmdHooks(s); code != 0 || errOut.String() != "" {
		t.Fatalf("hooks: exit=%d err=%q", code, errOut.String())
	}
	hooks := out.String()
	if !strings.Contains(hooks, "gex-conventional-commits") || !strings.Contains(hooks, "claude/smart-commit [somente referência]") || !strings.Contains(hooks, "codex/secret-scanner [somente referência]") || strings.Contains(hooks, "\x1b") {
		t.Fatalf("hooks sem referências esperadas ou com ANSI: %q", hooks)
	}
}

func TestDoctorReportsDriftAsFailure(t *testing.T) {
	d, _ := readFakeDeps(readDriftReport())
	s, out, errOut := readTestSession(options{Home: t.TempDir()}, d)
	code := cmdDoctor(s)
	if code != 1 || errOut.String() != "" || !strings.Contains(out.String(), "alterado") {
		t.Fatalf("drift: exit=%d out=%q err=%q", code, out.String(), errOut.String())
	}
}

func TestDoctorTextPrintsOneItemPerLine(t *testing.T) {
	d, _ := readFakeDeps(readDriftReport())
	s, out, _ := readTestSession(options{Home: t.TempDir()}, d)
	cmdDoctor(s)
	want := "ok .claude/AGENTS.md (Íntegro.)\nalterado .claude/CLAUDE.md (Arquivo mudou depois da instalação.)\n"
	if !strings.Contains(out.String(), want) {
		t.Fatalf("linhas do doctor fora do formato: %q", out.String())
	}
}

func TestDoctorRepoUsesRepoDependency(t *testing.T) {
	d, calls := readFakeDeps(app.Report{Root: "repo"})
	s, _, _ := readTestSession(options{Repo: t.TempDir()}, d)
	cmdDoctor(s)
	if len(*calls) != 1 || (*calls)[0] != "repo" {
		t.Fatalf("doctor --repo chamou %v; esperado somente repo", *calls)
	}
}

func TestDoctorWarningsOnlyKeepExitZero(t *testing.T) {
	report := app.Report{
		Root:     "home",
		Warnings: []app.Status{{Path: "CLAUDE_CONFIG_DIR", State: "aviso", Detail: "Variável definida."}},
		Agents:   []app.AgentChange{{Name: "claude", Target: "2.1.306", Blocked: true, Reason: "não gerenciado"}},
	}
	d, _ := readFakeDeps(report)
	s, _, errOut := readTestSession(options{Home: t.TempDir()}, d)
	if code := cmdDoctor(s); code != 0 || errOut.String() != "" {
		t.Fatalf("avisos de agente não gerenciado: exit=%d err=%q", code, errOut.String())
	}
}

func TestDoctorJSONMatchesReturnedCode(t *testing.T) {
	d, _ := readFakeDeps(readDriftReport())
	s, out, _ := readTestSession(options{Home: t.TempDir(), JSON: true}, d)
	code := cmdDoctor(s)
	var doc struct {
		Schema   string `json:"schema"`
		OK       bool   `json:"ok"`
		ExitCode int    `json:"exit_code"`
	}
	readDecode(t, out.String(), &doc)
	if doc.Schema != "gex-dev-kit.doctor/v1" || doc.ExitCode != code || doc.OK != (code == 0) {
		t.Fatalf("doctor --json: code=%d doc=%+v", code, doc)
	}
}

func TestDoctorJSONEmptyListsAreArrays(t *testing.T) {
	d, _ := readFakeDeps(app.Report{Root: "home"})
	s, out, _ := readTestSession(options{Home: t.TempDir(), JSON: true}, d)
	cmdDoctor(s)
	for _, key := range []string{`"items": []`, `"agents": []`, `"warnings": []`} {
		if !strings.Contains(out.String(), key) {
			t.Fatalf("lista vazia não é [] (%s): %q", key, out.String())
		}
	}
}

func TestDoctorJSONReplacesHomeWithTilde(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	report := app.Report{
		Root:  "home",
		Items: []app.Status{{Path: ".claude/CLAUDE.md", State: "alterado", Detail: filepath.Join(home, ".claude", "CLAUDE.md") + " mudou"}},
	}
	d, _ := readFakeDeps(report)
	s, out, _ := readTestSession(options{Home: home, JSON: true}, d)
	cmdDoctor(s)
	if strings.Contains(out.String(), home) || !strings.Contains(out.String(), "~/.claude/CLAUDE.md mudou") {
		t.Fatalf("HOME não virou ~ em detail: %q", out.String())
	}
}

func TestDoctorJSONKeepsAmpersandLiteral(t *testing.T) {
	report := app.Report{
		Root:   "home",
		Agents: []app.AgentChange{{Name: "claude", Target: "2.1.306", Blocked: true, Reason: "PATH ambíguo", Manual: "npm i -g a@1 && npm i -g b@2"}},
	}
	d, _ := readFakeDeps(report)
	s, out, _ := readTestSession(options{Home: t.TempDir(), JSON: true}, d)
	cmdDoctor(s)
	if !strings.Contains(out.String(), "npm i -g a@1 && npm i -g b@2") {
		t.Fatalf("&& escapado no JSON: %q", out.String())
	}
}

func TestSkillsJSONCountsAndSchema(t *testing.T) {
	s, out, _ := readTestSession(options{JSON: true}, deps{})
	if code := cmdSkills(s); code != 0 {
		t.Fatalf("skills --json: exit=%d", code)
	}
	var doc struct {
		Schema      string `json:"schema"`
		Core        []any  `json:"core"`
		Specialties []any  `json:"specialties"`
		Matt        struct {
			Revision string `json:"revision"`
			Skills   []any  `json:"skills"`
		} `json:"matt"`
		References []any `json:"references"`
	}
	readDecode(t, out.String(), &doc)
	if doc.Schema != "gex-dev-kit.skills/v1" || len(doc.Core) != 7 || len(doc.Specialties) != 10 || len(doc.Matt.Skills) != 18 || len(doc.References) != 11 {
		t.Fatalf("contagens de skills: schema=%s core=%d esp=%d matt=%d refs=%d", doc.Schema, len(doc.Core), len(doc.Specialties), len(doc.Matt.Skills), len(doc.References))
	}
	if doc.Matt.Revision == "" {
		t.Fatalf("revisão Matt ausente no JSON")
	}
}

func TestHooksJSONCountsAndSchema(t *testing.T) {
	s, out, _ := readTestSession(options{JSON: true}, deps{})
	if code := cmdHooks(s); code != 0 {
		t.Fatalf("hooks --json: exit=%d", code)
	}
	var doc struct {
		Schema     string `json:"schema"`
		AgentHooks []any  `json:"agent_hooks"`
		GitHooks   []any  `json:"git_hooks"`
		References []any  `json:"references"`
	}
	readDecode(t, out.String(), &doc)
	if doc.Schema != "gex-dev-kit.hooks/v1" || len(doc.AgentHooks) != 4 || len(doc.GitHooks) != 3 || len(doc.References) != 25 {
		t.Fatalf("contagens de hooks: schema=%s agent=%d git=%d refs=%d", doc.Schema, len(doc.AgentHooks), len(doc.GitHooks), len(doc.References))
	}
}

func TestAddonsJSONListsFourAndSelectsOne(t *testing.T) {
	s, out, _ := readTestSession(options{JSON: true}, deps{})
	if code := cmdAddons(s); code != 0 {
		t.Fatalf("addons --json: exit=%d", code)
	}
	var doc struct {
		Schema string `json:"schema"`
		Addons []struct {
			Name          string `json:"name"`
			ExecutedByKit bool   `json:"executed_by_kit"`
		} `json:"addons"`
	}
	readDecode(t, out.String(), &doc)
	if doc.Schema != "gex-dev-kit.addons/v1" || len(doc.Addons) != 4 {
		t.Fatalf("addons: schema=%s total=%d", doc.Schema, len(doc.Addons))
	}
	for _, addon := range doc.Addons {
		if addon.ExecutedByKit {
			t.Fatalf("addon %s marcado como executado pelo kit", addon.Name)
		}
	}
	s, out, _ = readTestSession(options{JSON: true, Args: []string{"caveman"}}, deps{})
	cmdAddons(s)
	readDecode(t, out.String(), &doc)
	if len(doc.Addons) != 1 || doc.Addons[0].Name != "caveman" {
		t.Fatalf("addon único: %+v", doc.Addons)
	}
}

func TestAddonsUnknownNameIsUsageError(t *testing.T) {
	s, out, errOut := readTestSession(options{Args: []string{"nao-existe"}}, deps{})
	if code := cmdAddons(s); code != 2 || out.String() != "" || !strings.Contains(errOut.String(), "addon desconhecido") {
		t.Fatalf("addon inexistente: exit=%d out=%q err=%q", code, out.String(), errOut.String())
	}
}

func TestAddonsUnknownNameListsValidosESugere(t *testing.T) {
	s, _, errOut := readTestSession(options{Args: []string{"cavemn"}}, deps{})
	if code := cmdAddons(s); code != 2 {
		t.Fatalf("addon inexistente: exit=%d; quero 2", code)
	}
	want := "addon desconhecido: cavemn\nAddons: caveman, ponytail, superpowers, spec-kit. Você quis dizer caveman?\n"
	if errOut.String() != want {
		t.Fatalf("saída de addon desconhecido = %q; quero %q", errOut.String(), want)
	}
}

func TestAddonsDesconhecidoSemParecidoNaoSugere(t *testing.T) {
	s, _, errOut := readTestSession(options{Args: []string{"nao-existe"}}, deps{})
	cmdAddons(s)
	if strings.Contains(errOut.String(), "Você quis dizer") {
		t.Fatalf("sugeriu sem parecido: %q", errOut.String())
	}
}

func TestJSONStdoutContainsOnlyJSON(t *testing.T) {
	d, _ := readFakeDeps(readDriftReport())
	commands := map[string]func(session) int{
		"skills": cmdSkills,
		"hooks":  cmdHooks,
		"addons": cmdAddons,
		"doctor": cmdDoctor,
	}
	for name, command := range commands {
		s, out, _ := readTestSession(options{JSON: true, Home: t.TempDir()}, d)
		command(s)
		if !json.Valid(out.Bytes()) {
			t.Fatalf("%s --json: stdout não é só JSON: %q", name, out.String())
		}
	}
}

func TestJSONRejectsTUIMode(t *testing.T) {
	s, _, errOut := readTestSession(options{JSON: true, UI: terminal.ModeTUI}, deps{})
	if code := cmdDoctor(s); code != 2 || !strings.Contains(errOut.String(), "--json") {
		t.Fatalf("--json com --ui=tui: exit=%d err=%q", code, errOut.String())
	}
}

func readDoctorConfere(t *testing.T, home string) (int, readDoctorCheck) {
	t.Helper()
	s, out, _ := readTestSession(options{Home: home, JSON: true}, realDeps())
	code := cmdDoctor(s)
	var doc readDoctorCheck
	readDecode(t, out.String(), &doc)
	if doc.ExitCode != code || doc.OK != (code == 0) {
		t.Fatalf("ok e exit_code divergem: exit=%d doc=%+v", code, doc)
	}
	var ok, atencao, pendente int
	for _, item := range doc.Items {
		switch item.Level {
		case "ok":
			ok++
		case "atencao":
			atencao++
		case "pendente":
			pendente++
		}
	}
	if doc.Counts.OK != ok || doc.Counts.Attention != atencao || doc.Counts.Pending != pendente {
		t.Fatalf("counts %+v diverge dos níveis dos itens: ok=%d atencao=%d pendente=%d", doc.Counts, ok, atencao, pendente)
	}
	return code, doc
}

type readDoctorCheck struct {
	OK       bool             `json:"ok"`
	ExitCode int              `json:"exit_code"`
	Counts   jsonDoctorCounts `json:"counts"`
	Items    []jsonDoctorItem `json:"items"`
}

func TestDoctorJSONSemInstalacao(t *testing.T) {
	home := cliIsolate(t)
	code, doc := readDoctorConfere(t, home)
	if code != 0 || !doc.OK || doc.Counts.Pending == 0 {
		t.Fatalf("sem instalação: exit=%d doc=%+v", code, doc)
	}
}

func TestDoctorJSONInstalacaoLimpa(t *testing.T) {
	home := cliIsolate(t)
	cliInstalaCore(t, home)
	code, doc := readDoctorConfere(t, home)
	if code != 0 || !doc.OK || doc.Counts.Attention != 0 {
		t.Fatalf("instalação limpa: exit=%d doc=%+v", code, doc)
	}
}

func TestDoctorJSONArquivoAlterado(t *testing.T) {
	home := cliIsolate(t)
	cliInstalaCore(t, home)
	if err := os.WriteFile(filepath.Join(home, ".claude", "CLAUDE.md"), []byte("alterado pelo teste\n"), 0o644); err != nil {
		t.Fatalf("alterar arquivo do kit: %v", err)
	}
	code, doc := readDoctorConfere(t, home)
	if code != 1 || doc.OK || doc.Counts.Attention == 0 {
		t.Fatalf("arquivo alterado: exit=%d doc=%+v", code, doc)
	}
}

func TestReadFailNaoUsaPrefixoErro(t *testing.T) {
	var buf bytes.Buffer
	code := readFail(&buf, githooks.ErrNotRepo)
	if code != 2 || strings.Contains(buf.String(), "erro:") || !strings.Contains(buf.String(), "Não é repositório Git") {
		t.Fatalf("readFail: exit=%d saída=%q", code, buf.String())
	}
}
