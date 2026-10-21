package app

import (
	"strings"
	"testing"

	"github.com/Gex-Development/gex-dev-kit/internal/kit"
)

func TestMaskLineHidesSecretsAndCounts(t *testing.T) {
	ghp := "ghp_" + strings.Repeat("a1", 18)
	webhook := "https://hooks.slack.com/services/T" + strings.Repeat("0", 8) + "/B" + strings.Repeat("1", 8) + "/" + strings.Repeat("x", 24)
	long := "chave" + strings.Repeat("9", 3) + strings.Repeat("q", 17)
	cases := []struct {
		name, in, want string
		count          int
	}{
		{"rotulo com igual", "token=abc123", "token=•••", 1},
		{"rotulo com dois pontos", "API_KEY: xyz", "API_KEY:•••", 1},
		{"autorizacao com bearer", "Authorization: Bearer zzz", "Authorization:•••", 1},
		{"prefixo de token", "usar " + ghp, "usar •••", 1},
		{"url de webhook com caminho", webhook, "https://hooks.slack.com/•••", 1},
		{"url sem caminho fica", "https://example.com", "https://example.com", 0},
		{"sequencia longa com letra e digito", long, "•••", 1},
		{"sequencia longa so com letras fica", "abcdefghijklmnopqrstuvwxyz", "abcdefghijklmnopqrstuvwxyz", 0},
		{"texto comum fica", "docs/README.md de exemplo", "docs/README.md de exemplo", 0},
	}
	for _, c := range cases {
		got, n := maskLine(c.in)
		if got != c.want || n != c.count {
			t.Fatalf("%s: maskLine(%q) = (%q, %d), esperado (%q, %d)", c.name, c.in, got, n, c.want, c.count)
		}
	}
}

const diffTestStableExe = "/home/teste/.local/share/gex-dev-kit/bin/gex-dev-kit"

func diffTestTexts(lines []DiffLine) string {
	var b strings.Builder
	for _, l := range lines {
		b.WriteString(l.Text)
		b.WriteString("\n")
	}
	return b.String()
}

func diffTestFind(t *testing.T, diffs []FileDiff, path string) FileDiff {
	t.Helper()
	for _, d := range diffs {
		if d.Path == path {
			return d
		}
	}
	t.Fatalf("diff de %s não encontrado", path)
	return FileDiff{}
}

func TestBuildDiffsCoreBlockOnly(t *testing.T) {
	before := []byte("# Notas do usuário\n<!-- início gex-dev-kit core -->\nversão antiga\n<!-- fim gex-dev-kit core -->\n\n## Seção pessoal\nsegredo pessoal do usuário\n")
	after := []byte("# Notas do usuário\n<!-- início gex-dev-kit core -->\nversão nova\n<!-- fim gex-dev-kit core -->\n\n## Seção pessoal\nsegredo pessoal do usuário\n")
	d := diffTestFind(t, buildDiffs([]kit.FileContent{{Path: ".claude/CLAUDE.md", Before: before, After: after, Existed: true}}, DiffLinesTUI), ".claude/CLAUDE.md")
	text := diffTestTexts(d.Lines)
	if strings.Contains(text, "segredo pessoal") || strings.Contains(text, "Notas do usuário") {
		t.Fatalf("diff expôs conteúdo fora do bloco Core: %q", text)
	}
	if !strings.Contains(text, "versão nova") {
		t.Fatalf("diff não mostrou o bloco Core novo: %q", text)
	}
}

func TestBuildDiffsHidesThirdPartyHooks(t *testing.T) {
	before := []byte(`{"env":{"SENHA":"SENHA-NAO-REPETIR"},"hooks":{"PreToolUse":[{"matcher":"Bash","hooks":[{"type":"command","command":"/usr/bin/alheio","args":["x"]}]}]}}`)
	after := []byte(`{"env":{"SENHA":"SENHA-NAO-REPETIR"},"hooks":{"PreToolUse":[{"matcher":"Bash","hooks":[{"type":"command","command":"/usr/bin/alheio","args":["x"]},{"type":"command","command":"curl https://hooks.slack.com/services/T00000000/B00000000/xxxxxxxxxxxxxxxxxxxxxxxx"}]},{"matcher":"Bash","hooks":[{"type":"command","command":"` + diffTestStableExe + `","args":["hook","commit-msg"],"onFailure":"block"}]}]}}`)
	d := diffTestFind(t, buildDiffs([]kit.FileContent{{Path: ".claude/settings.json", Before: before, After: after, Existed: true}}, DiffLinesTUI), ".claude/settings.json")
	text := diffTestTexts(d.Lines)
	if strings.Contains(text, "SENHA-NAO-REPETIR") || strings.Contains(text, "hooks.slack.com") || strings.Contains(text, "/usr/bin/alheio") {
		t.Fatalf("diff expôs handler de terceiro ou valor de env: %q", text)
	}
	if !strings.Contains(text, "gex-dev-kit") {
		t.Fatalf("diff não mostrou o handler do kit: %q", text)
	}
	if d.Summary != "2 handlers de terceiros preservados, não exibidos" {
		t.Fatalf("resumo de terceiros = %q", d.Summary)
	}
}

func TestBuildDiffsMasksWebhookInSkillLine(t *testing.T) {
	before := []byte("linha inicial\n")
	after := []byte("linha inicial\nuse https://hooks.slack.com/services/T00000000/B00000000/xxxxxxxxxxxxxxxxxxxxxxxx agora\n")
	d := diffTestFind(t, buildDiffs([]kit.FileContent{{Path: ".claude/skills/gex-x/SKILL.md", Before: before, After: after, Existed: true}}, DiffLinesTUI), ".claude/skills/gex-x/SKILL.md")
	text := diffTestTexts(d.Lines)
	if strings.Contains(text, "xxxxxxxx") || !strings.Contains(text, "https://hooks.slack.com/•••") {
		t.Fatalf("webhook na skill não foi mascarado: %q", text)
	}
	if d.Masked != 1 {
		t.Fatalf("Masked = %d, esperado 1", d.Masked)
	}
}

func TestBuildDiffsBinaryHasNoLines(t *testing.T) {
	d := diffTestFind(t, buildDiffs([]kit.FileContent{{Path: ".local/share/gex-dev-kit/bin/gex-dev-kit", Before: []byte{0, 1}, After: []byte{0, 2, 3}, Existed: true, Binary: true}}, DiffLinesTUI), ".local/share/gex-dev-kit/bin/gex-dev-kit")
	if len(d.Lines) != 0 || !d.Binary {
		t.Fatalf("binário deveria vir sem linhas: lines=%d binary=%v", len(d.Lines), d.Binary)
	}
	if !strings.Contains(d.Summary, "2 → 3 bytes") || d.Action != "atualizar" {
		t.Fatalf("resumo do binário = %q, ação = %q", d.Summary, d.Action)
	}
}

func TestBuildDiffsNewFileSummary(t *testing.T) {
	d := diffTestFind(t, buildDiffs([]kit.FileContent{{Path: ".claude/skills/x/a.md", After: []byte("1\n2\n3\n"), Existed: false}}, DiffLinesTUI), ".claude/skills/x/a.md")
	if d.Action != "criar" || d.Summary != "novo, 3 linhas" || len(d.Lines) != 0 {
		t.Fatalf("arquivo novo: ação=%q resumo=%q linhas=%d", d.Action, d.Summary, len(d.Lines))
	}
}

func TestBuildDiffsLimitsLinesPerFile(t *testing.T) {
	var before, after strings.Builder
	for i := 0; i < 100; i++ {
		before.WriteString("linha antiga\n")
		after.WriteString("linha nova\n")
	}
	d := diffTestFind(t, buildDiffs([]kit.FileContent{{Path: ".claude/skills/x/b.md", Before: []byte(before.String()), After: []byte(after.String()), Existed: true}}, DiffLinesCLI), ".claude/skills/x/b.md")
	if len(d.Lines) != DiffLinesCLI || d.Omitted <= 0 {
		t.Fatalf("limite: linhas=%d omitidas=%d", len(d.Lines), d.Omitted)
	}
}

func TestDiffKitHandlerRecognizesEveryKitForm(t *testing.T) {
	exe := diffTestStableExe
	cases := []struct {
		name string
		h    map[string]any
		want bool
	}{
		{"claude com args hook", map[string]any{"command": exe, "args": []any{"hook", "commit-msg"}}, true},
		{"claude legado sem args", map[string]any{"command": exe}, true},
		{"codex com aspas", map[string]any{"command": `"` + exe + `" hook commit-msg`}, true},
		{"codex sem aspas", map[string]any{"command": exe + " hook commit-msg"}, true},
		{"claude de terceiro", map[string]any{"command": "/usr/bin/alheio", "args": []any{"hook"}}, false},
		{"claude de outro executavel do kit", map[string]any{"command": "/usr/bin/outra", "args": []any{"hook", "commit-msg"}}, false},
	}
	for _, c := range cases {
		if got := diffKitHandler(c.h); got != c.want {
			t.Fatalf("%s: diffKitHandler = %v, esperado %v", c.name, got, c.want)
		}
	}
}
