package githooks

import (
	"os"
	"path/filepath"
	"testing"
)

func documentedRoot(t *testing.T, arquivos map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for nome, conteudo := range arquivos {
		if err := os.WriteFile(filepath.Join(root, nome), []byte(conteudo), 0644); err != nil {
			t.Fatalf("WriteFile(%s) erro: %v", nome, err)
		}
	}
	return root
}

func TestDocumentedEncontraOrigemNoTexto(t *testing.T) {
	root := documentedRoot(t, map[string]string{
		"AGENTS.md":       "# Agentes\n\nRode antes do commit:\n- `make test`\n",
		"README.md":       "Use `go test ./...`.\n",
		"CONTRIBUTING.md": "Rode `yarn run check` antes do PR.\n",
		"notas.txt":       "make secreto\n",
	})
	cases := []struct {
		comando string
		want    Source
	}{
		{"make test", Source{Command: "make test", File: "AGENTS.md", Line: 4}},
		{"go test ./...", Source{Command: "go test ./...", File: "README.md", Line: 1}},
		{"yarn run check", Source{Command: "yarn run check", File: "CONTRIBUTING.md", Line: 1}},
	}
	for _, c := range cases {
		got, ok := Documented(root, c.comando)
		if !ok {
			t.Fatalf("Documented(%q) não encontrou", c.comando)
		}
		if got != c.want {
			t.Fatalf("Documented(%q) = %+v, queria %+v", c.comando, got, c.want)
		}
	}
}

func TestDocumentedExigeLimitePorPalavra(t *testing.T) {
	root := documentedRoot(t, map[string]string{
		"AGENTS.md": "- `make test-all`\n- `go test ./...x`\n",
	})
	for _, comando := range []string{"make test", "go test ./..."} {
		if got, ok := Documented(root, comando); ok {
			t.Fatalf("Documented(%q) casou com texto maior: %+v", comando, got)
		}
	}
}

func TestDocumentedAceitaAlvoDoMakefile(t *testing.T) {
	root := documentedRoot(t, map[string]string{
		"Makefile": "CC := gcc\nbuild:\n\tgo build\n",
	})
	got, ok := Documented(root, "make build")
	if !ok {
		t.Fatalf("Documented(make build) não encontrou alvo no Makefile")
	}
	if want := (Source{Command: "make build", File: "Makefile", Line: 2}); got != want {
		t.Fatalf("Documented(make build) = %+v, queria %+v", got, want)
	}
	if _, ok := Documented(root, "make CC"); ok {
		t.Fatalf("Documented(make CC) aceitou variável como alvo")
	}
}

func TestDocumentedAceitaScriptDoPackageJson(t *testing.T) {
	root := documentedRoot(t, map[string]string{
		"package.json": "{\n  \"scripts\": {\n    \"lint:fix\": \"eslint --fix\"\n  }\n}\n",
	})
	cases := []struct {
		comando string
		want    Source
	}{
		{"npm run lint:fix", Source{Command: "npm run lint:fix", File: "package.json", Line: 3}},
		{"pnpm lint:fix", Source{Command: "pnpm lint:fix", File: "package.json", Line: 3}},
	}
	for _, c := range cases {
		got, ok := Documented(root, c.comando)
		if !ok || got != c.want {
			t.Fatalf("Documented(%q) = %+v, %v; queria %+v", c.comando, got, ok, c.want)
		}
	}
	if _, ok := Documented(root, "npm run lint"); ok {
		t.Fatalf("Documented(npm run lint) aceitou script inexistente")
	}
}

func TestDocumentedRecusaComandoNaoDocumentado(t *testing.T) {
	root := documentedRoot(t, map[string]string{"AGENTS.md": "Nada aqui.\n"})
	if got, ok := Documented(root, "rm -rf /"); ok {
		t.Fatalf("Documented aceitou comando não documentado: %+v", got)
	}
	if _, ok := Documented("", "make test"); ok {
		t.Fatalf("Documented aceitou raiz vazia")
	}
}
