package catalog

import (
	"slices"
	"strings"
	"testing"
)

func TestHookSpecContagemPorEscopo(t *testing.T) {
	specs := HookSpecs()
	if len(specs) != 7 {
		t.Fatalf("HookSpecs() = %d hooks, esperado 7", len(specs))
	}
	agente, git := 0, 0
	for _, spec := range specs {
		switch spec.Scope {
		case HookScopeAgent:
			agente++
		case HookScopeGit:
			git++
		}
	}
	if agente != 4 || git != 3 {
		t.Fatalf("escopos: agente=%d git=%d, esperado 4 e 3", agente, git)
	}
}

func TestHookSpecNomesEmOrdem(t *testing.T) {
	want := []string{
		"gex-conventional-commits",
		"gex-sensitive-action-warning",
		"gex-pre-pr",
		"gex-session-start",
		"gex-commit-msg",
		"gex-pre-commit",
		"gex-pre-push",
	}
	var got []string
	for _, spec := range HookSpecs() {
		got = append(got, spec.Name)
	}
	if !slices.Equal(got, want) {
		t.Fatalf("nomes = %v, esperado %v", got, want)
	}
}

func TestHookSpecGitHooksLista(t *testing.T) {
	want := []string{"gex-commit-msg", "gex-pre-commit", "gex-pre-push"}
	if got := GitHooks(); !slices.Equal(got, want) {
		t.Fatalf("GitHooks() = %v, esperado %v", got, want)
	}
}

func TestHookSpecByNameEncontraEAusenteRetornaFalso(t *testing.T) {
	spec, ok := HookSpecByName("gex-pre-pr")
	if !ok || spec.Name != "gex-pre-pr" {
		t.Fatalf("HookSpecByName(gex-pre-pr) = %q, %v", spec.Name, ok)
	}
	if _, ok := HookSpecByName("gex-inexistente"); ok {
		t.Fatal("hook inexistente foi encontrado")
	}
}

func TestHookSpecSuporteSemAgenteTemMotivo(t *testing.T) {
	for _, spec := range HookSpecs() {
		if !spec.Codex.Supported && spec.Codex.Reason == "" {
			t.Fatalf("%s: sem suporte Codex e sem motivo", spec.Name)
		}
		if !spec.Claude.Supported && spec.Claude.Reason == "" {
			t.Fatalf("%s: sem suporte Claude e sem motivo", spec.Name)
		}
	}
}

func TestHookSpecSuporteEsperadoPorAgente(t *testing.T) {
	want := map[string][2]bool{
		"gex-conventional-commits":     {true, true},
		"gex-sensitive-action-warning": {true, false},
		"gex-pre-pr":                   {true, false},
		"gex-session-start":            {true, true},
		"gex-commit-msg":               {false, false},
		"gex-pre-commit":               {false, false},
		"gex-pre-push":                 {false, false},
	}
	for _, spec := range HookSpecs() {
		esperado, ok := want[spec.Name]
		if !ok {
			t.Fatalf("hook sem expectativa: %s", spec.Name)
		}
		if spec.Claude.Supported != esperado[0] || spec.Codex.Supported != esperado[1] {
			t.Fatalf("%s: suporte Claude=%v Codex=%v, esperado Claude=%v Codex=%v", spec.Name, spec.Claude.Supported, spec.Codex.Supported, esperado[0], esperado[1])
		}
	}
}

func TestHookSpecEventoEMatcherPorAgente(t *testing.T) {
	want := map[string][4]string{
		"gex-conventional-commits":     {"PreToolUse", "Bash", "PreToolUse", "Bash"},
		"gex-sensitive-action-warning": {"PreToolUse", "Bash|PowerShell|Read", "", ""},
		"gex-pre-pr":                   {"PreToolUse", "Bash|PowerShell", "", ""},
		"gex-session-start":            {"SessionStart", "startup|resume|clear|compact|fork", "SessionStart", "startup|resume|clear|compact|fork"},
	}
	for name, esperado := range want {
		spec, ok := HookSpecByName(name)
		if !ok {
			t.Fatalf("hook ausente: %s", name)
		}
		got := [4]string{spec.Claude.Event, spec.Claude.Matcher, spec.Codex.Event, spec.Codex.Matcher}
		if got != esperado {
			t.Fatalf("%s: evento e matcher = %v, esperado %v", name, got, esperado)
		}
	}
}

func TestHookSpecSubcomandoDoAdaptador(t *testing.T) {
	want := map[string]string{
		"gex-conventional-commits":     "commit-msg",
		"gex-sensitive-action-warning": "sensitive-action",
		"gex-pre-pr":                   "pre-pr",
		"gex-session-start":            "session-start",
		"gex-commit-msg":               "commit-msg",
		"gex-pre-commit":               "pre-commit",
		"gex-pre-push":                 "",
	}
	for _, spec := range HookSpecs() {
		if spec.Command != want[spec.Name] {
			t.Fatalf("%s: Command = %q, esperado %q", spec.Name, spec.Command, want[spec.Name])
		}
	}
}

func TestHookSpecPolicyGEXouTodos(t *testing.T) {
	want := map[string]string{
		"gex-conventional-commits":     "gex",
		"gex-sensitive-action-warning": "gex",
		"gex-pre-pr":                   "gex",
		"gex-session-start":            "todos",
		"gex-commit-msg":               "todos",
		"gex-pre-commit":               "todos",
		"gex-pre-push":                 "todos",
	}
	for _, spec := range HookSpecs() {
		if spec.Policy != want[spec.Name] {
			t.Fatalf("%s: Policy = %q, esperado %q", spec.Name, spec.Policy, want[spec.Name])
		}
	}
}

func TestHookSpecDecisaoEValidaQuandoHaSuporte(t *testing.T) {
	for _, spec := range HookSpecs() {
		for _, suporte := range []HookSupport{spec.Claude, spec.Codex} {
			if !suporte.Supported {
				if suporte.Decision != "" {
					t.Fatalf("%s: decisão sem suporte: %q", spec.Name, suporte.Decision)
				}
				continue
			}
			if !slices.Contains([]string{"bloqueia", "pergunta", "contexto"}, suporte.Decision) {
				t.Fatalf("%s: decisão inválida: %q", spec.Name, suporte.Decision)
			}
		}
		if spec.Codex.Supported && !spec.Codex.Trust {
			t.Fatalf("%s: hook do Codex precisa de aprovação em /hooks", spec.Name)
		}
	}
}

func TestHookSpecDescricaoOrigemEEscopoGit(t *testing.T) {
	for _, spec := range HookSpecs() {
		if strings.TrimSpace(spec.Description) == "" || !strings.HasPrefix(spec.Source, "Autoria GEX") {
			t.Fatalf("%s: descrição ou fonte ausente", spec.Name)
		}
		if spec.Scope == HookScopeGit {
			if spec.GitHook == "" || spec.Name != "gex-"+spec.GitHook {
				t.Fatalf("%s: GitHook = %q, esperado o nome do hook Git", spec.Name, spec.GitHook)
			}
		} else if spec.GitHook != "" {
			t.Fatalf("%s: hook de agente com GitHook = %q", spec.Name, spec.GitHook)
		}
	}
}
