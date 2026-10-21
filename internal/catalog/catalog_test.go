package catalog

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
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

func TestHookKeepsOtherHandlersInSameMatcher(t *testing.T) {
	executable := filepath.Join(t.TempDir(), "gex-dev-kit")
	original := []byte(`{"hooks":{"PreToolUse":[{"matcher":"Bash","hooks":[{"type":"command","command":"OWN"},{"type":"command","command":"echo own"},{"type":"command","command":"/elsewhere/command"}]}]}}`)
	original = bytes.Replace(original, []byte("OWN"), []byte(executable), 1)
	merged, err := mergeHook(original, executable, true)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(merged), "echo own") || !strings.Contains(string(merged), "/elsewhere/command") {
		t.Fatal("handler alheio perdido")
	}
	again, err := mergeHook(merged, executable, true)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(again), `"onFailure": "block"`) != 1 {
		t.Fatal("hook duplicado")
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

func TestCoreRepeatedBuildDoesNotDuplicate(t *testing.T) {
	home := t.TempDir()
	first, err := Build(home, "", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{".claude/CLAUDE.md", ".codex/AGENTS.md"} {
		file := lookup(t, first, path)
		if err := os.MkdirAll(filepath.Dir(filepath.Join(home, path)), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(home, path), file.Data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	second, err := Build(home, "", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{".claude/CLAUDE.md", ".codex/AGENTS.md"} {
		if string(lookup(t, second, path).Data) != string(lookup(t, first, path).Data) {
			t.Fatalf("duplicação em %s", path)
		}
	}
}
