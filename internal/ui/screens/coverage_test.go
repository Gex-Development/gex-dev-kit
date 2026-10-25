package screens

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

func TestCoverageMapIsComplete(t *testing.T) {
	paginas := map[string]bool{}
	for _, p := range showcasePages() {
		paginas[p.Name] = true
	}
	for i, c := range CoverageMap {
		if c.Lib == "" || c.Feature == "" || c.Page == "" {
			t.Fatalf("entrada %d sem Lib, Feature ou Page: %+v", i, c)
		}
		if !paginas[c.Page] {
			t.Fatalf("entrada %d (%s) aponta para página inexistente %q", i, c.Feature, c.Page)
		}
		if c.TestOnly {
			continue
		}
		if c.Demo == nil {
			t.Fatalf("entrada %d (%s) sem Demo", i, c.Feature)
		}
		if c.Marker == "" {
			t.Fatalf("entrada %d (%s) sem Marker", i, c.Feature)
		}
	}
}

func TestCoverageTestOnlyEhSoParaTesteDeTUI(t *testing.T) {
	for _, c := range CoverageMap {
		if !c.TestOnly {
			continue
		}
		if c.Lib != "x" || (c.Call != "teatest.NewTestModel" && c.Call != "golden.RequireEqual") {
			t.Fatalf("TestOnly indevido em %s (%s): só teatest e golden podem ser TestOnly", c.Feature, c.Call)
		}
	}
}

func TestCoverageCobreCadaLibDaTabela(t *testing.T) {
	libs := map[string]bool{}
	for _, c := range CoverageMap {
		libs[c.Lib] = true
	}
	for _, lib := range []string{"bubbletea", "bubbles", "huh", "lipgloss", "harmonica", "x"} {
		if !libs[lib] {
			t.Fatalf("lib %q da tabela §4 não aparece em CoverageMap", lib)
		}
	}
}

func TestShowcaseCoversCharmMap(t *testing.T) {
	tema := theme.New(theme.Options{Dark: true, Profile: colorprofile.ASCII})
	for i, c := range CoverageMap {
		if c.Page == "" {
			t.Fatalf("entrada %d (%s) sem página na vitrine", i, c.Feature)
		}
		if c.TestOnly {
			if !showcaseTestesListaArquivoExistente(t, tema) {
				t.Fatalf("entrada %d (%s): a página Testes não lista nenhum arquivo de teste existente", i, c.Feature)
			}
			continue
		}
		if c.Demo == nil {
			t.Fatalf("entrada %d (%s) sem Demo", i, c.Feature)
		}
		if !strings.Contains(ansi.Strip(c.Demo(tema, 80)), c.Marker) {
			t.Fatalf("entrada %d (%s): marcador %q ausente da demonstração", i, c.Feature, c.Marker)
		}
	}
}

func showcaseTestesListaArquivoExistente(t *testing.T, tema theme.Theme) bool {
	t.Helper()
	for _, p := range showcasePages() {
		if p.Name != showcasePageTestes {
			continue
		}
		pagina := ansi.Strip(p.Own(showcaseLive{}, tema, 80))
		for _, f := range showcaseTestFiles {
			if _, err := os.Stat(showcaseRepoPath(f.Path)); err == nil && strings.Contains(pagina, f.Path) {
				return true
			}
		}
	}
	return false
}

func TestRealScreensCallConstructors(t *testing.T) {
	raiz := coverageRepoRaiz(t)
	raizes := []string{"internal/ui/screens", "internal/ui/components", "internal/ui/layout", "internal/ui/theme", "internal/ui/terminal", "internal/ui/motion", "cmd/gex-dev-kit"}
	arquivos := coverageArquivosPorNome(t, raiz, raizes)
	var falhas []string
	for i, c := range CoverageMap {
		candidatos := arquivos[c.ScreenFile]
		if len(candidatos) == 0 {
			falhas = append(falhas, fmt.Sprintf("entrada %d (%s): arquivo %s não encontrado", i, c.Feature, c.ScreenFile))
			continue
		}
		if !coverageContemChamada(t, candidatos, c.Call) {
			falhas = append(falhas, fmt.Sprintf("entrada %d (%s): %s não aparece em %s", i, c.Feature, c.Call, c.ScreenFile))
		}
	}
	if len(falhas) > 0 {
		t.Fatalf("entradas de CoverageMap sem chamada no código:\n%s", strings.Join(falhas, "\n"))
	}
}

func coverageRepoRaiz(t *testing.T) string {
	t.Helper()
	_, arquivo, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("não foi possível localizar o diretório do teste")
	}
	return filepath.Join(filepath.Dir(arquivo), "..", "..", "..")
}

func coverageArquivosPorNome(t *testing.T, raiz string, pastas []string) map[string][]string {
	t.Helper()
	arquivos := map[string][]string{}
	for _, pasta := range pastas {
		err := filepath.WalkDir(filepath.Join(raiz, pasta), func(caminho string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && strings.HasSuffix(d.Name(), ".go") {
				arquivos[d.Name()] = append(arquivos[d.Name()], caminho)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("falha ao ler %s: %v", pasta, err)
		}
	}
	return arquivos
}

func coverageContemChamada(t *testing.T, caminhos []string, chamada string) bool {
	t.Helper()
	for _, caminho := range caminhos {
		arvore, err := parser.ParseFile(token.NewFileSet(), caminho, nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatalf("falha ao analisar %s: %v", caminho, err)
		}
		if coverageNomesDoArquivo(arvore)[chamada] {
			return true
		}
	}
	return false
}

func coverageNomesDoArquivo(arvore *ast.File) map[string]bool {
	nomes := map[string]bool{}
	ast.Inspect(arvore, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.Ident:
			nomes[x.Name] = true
		case *ast.SelectorExpr:
			nomes[x.Sel.Name] = true
			if base, ok := x.X.(*ast.Ident); ok {
				nomes[base.Name+"."+x.Sel.Name] = true
			}
		}
		return true
	})
	return nomes
}
