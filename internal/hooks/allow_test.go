package hooks

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestAllowPathFicaNoDiretorioComum(t *testing.T) {
	commonDir := t.TempDir()
	want := filepath.Join(commonDir, "gex-dev-kit", "pre-commit-allow")
	if got := AllowPath(commonDir); got != want {
		t.Fatalf("AllowPath = %q; esperado %q", got, want)
	}
}

func TestLoadAllowLeSoLinhasValidas(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pre-commit-allow")
	conteudo := "a.txt\taws-access-key\n\nlinha-sem-tabulacao\nb.txt\t\n\tregra-sem-arquivo\nc.txt\tprivate-key\r\n"
	if err := os.WriteFile(path, []byte(conteudo), 0o600); err != nil {
		t.Fatalf("gravar arquivo de exceção: %v", err)
	}
	allow, err := LoadAllow(path)
	if err != nil {
		t.Fatalf("LoadAllow devolveu erro: %v", err)
	}
	if len(allow) != 2 || !allow["a.txt\taws-access-key"] || !allow["c.txt\tprivate-key"] {
		t.Fatalf("liberados = %v; esperado só as duas linhas válidas", allow)
	}
}

func TestLoadAllowArquivoAusenteNaoEErro(t *testing.T) {
	allow, err := LoadAllow(filepath.Join(t.TempDir(), "pre-commit-allow"))
	if err != nil {
		t.Fatalf("arquivo ausente não deveria ser erro: %v", err)
	}
	if len(allow) != 0 {
		t.Fatalf("arquivo ausente deveria liberar nada; liberados = %v", allow)
	}
}

func TestAllowFilterRemoveSomenteOsLiberados(t *testing.T) {
	allow := Allow{"a.txt\taws-access-key": true}
	findings := []Finding{
		{Path: "a.txt", Line: 1, Rule: "aws-access-key"},
		{Path: "b.txt", Line: 2, Rule: "aws-access-key"},
		{Path: "a.txt", Line: 3, Rule: "private-key"},
	}
	got := allow.Filter(findings)
	want := []Finding{findings[1], findings[2]}
	if !slices.Equal(got, want) {
		t.Fatalf("Filter = %+v; esperado %+v", got, want)
	}
}
