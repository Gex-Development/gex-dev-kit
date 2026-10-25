package ui

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

const (
	archModule   = "github.com/Gex-Development/gex-dev-kit"
	archRepoRoot = "../.."
	archLipgloss = "charm.land/lipgloss/v2"
	archCompat   = "charm.land/lipgloss/v2/compat"
	archHuh      = "charm.land/huh/v2"
	archCmdDir   = "cmd/gex-dev-kit/"
)

var archRoots = []string{"internal/ui", "cmd", "internal/app"}

var archLegacy = []string{}

var archHexLiteral = regexp.MustCompile(`#([0-9a-fA-F]{6}|[0-9a-fA-F]{3})([^0-9a-zA-Z]|$)`)

func TestArquiteturaDaUI(t *testing.T) {
	var falhas []string
	arquivos := 0
	for _, root := range archRoots {
		err := filepath.WalkDir(filepath.Join(archRepoRoot, root), func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || filepath.Ext(p) != ".go" {
				return nil
			}
			arquivos++
			rel, err := filepath.Rel(archRepoRoot, p)
			if err != nil {
				return err
			}
			fset, file, err := archParse(p, nil)
			if err != nil {
				return err
			}
			falhas = append(falhas, archCheckFile(fset, filepath.ToSlash(rel), file, archIsLegacy(rel))...)
			return nil
		})
		if err != nil {
			t.Fatalf("percorrer %s: %v", root, err)
		}
	}
	if arquivos == 0 {
		t.Fatalf("nenhum arquivo .go percorrido: confira o diretório de execução do teste")
	}
	falhas = append(falhas, archLegacyFalhas()...)
	if len(falhas) > 0 {
		t.Fatalf("violações da arquitetura da UI:\n%s", strings.Join(falhas, "\n"))
	}
}

func TestArquiteturaDetectaInfracoes(t *testing.T) {
	hex := string(rune(35)) + "1e6fd9"
	casos := []struct {
		nome   string
		rel    string
		src    string
		legado bool
		falha  bool
	}{
		{
			nome:  "NewStyle fora do tema",
			rel:   "internal/ui/layout/x.go",
			src:   "package layout\n\nimport \"charm.land/lipgloss/v2\"\n\nvar estilo = lipgloss.NewStyle()\n",
			falha: true,
		},
		{
			nome: "NewStyle dentro do tema",
			rel:  "internal/ui/theme/x.go",
			src:  "package theme\n\nimport \"charm.land/lipgloss/v2\"\n\nvar estilo = lipgloss.NewStyle()\n",
		},
		{
			nome:  "cor hexadecimal literal fora do tema",
			rel:   "internal/ui/screens/x.go",
			src:   "package screens\n\nvar cor = \"" + hex + "\"\n",
			falha: true,
		},
		{
			nome: "cor hexadecimal dentro do tema",
			rel:  "internal/ui/theme/x.go",
			src:  "package theme\n\nvar cor = \"" + hex + "\"\n",
		},
		{
			nome:  "import de lipgloss compat",
			rel:   "internal/ui/theme/x.go",
			src:   "package theme\n\nimport _ \"charm.land/lipgloss/v2/compat\"\n",
			falha: true,
		},
		{
			nome:  "saída de lipgloss ligada ao stdout",
			rel:   "internal/ui/layout/x.go",
			src:   "package layout\n\nimport \"charm.land/lipgloss/v2\"\n\nfunc f() { lipgloss.Println(\"x\") }\n",
			falha: true,
		},
		{
			nome:  "comentário de linha",
			rel:   "internal/ui/layout/x.go",
			src:   "package layout\n\n// texto\nvar x = 1\n",
			falha: true,
		},
		{
			nome:  "comentário de bloco",
			rel:   "cmd/gex-dev-kit/novo.go",
			src:   "package main\n\n/* texto */\nvar x = 1\n",
			falha: true,
		},
		{
			nome: "diretiva go:build permitida",
			rel:  "internal/ui/layout/x.go",
			src:  "//go:build linux\n\npackage layout\n",
		},
		{
			nome:  "huh importado em cmd fora de accessible.go",
			rel:   "cmd/gex-dev-kit/novo.go",
			src:   "package main\n\nimport \"charm.land/huh/v2\"\n\nvar _ = huh.NewForm\n",
			falha: true,
		},
		{
			nome: "huh importado em accessible.go",
			rel:  "cmd/gex-dev-kit/accessible.go",
			src:  "package main\n\nimport \"charm.land/huh/v2\"\n\nvar _ = huh.NewForm\n",
		},
		{
			nome:   "estilo fora do tema em arquivo da lista de exceções",
			rel:    "cmd/gex-dev-kit/home.go",
			src:    "package main\n\nimport \"charm.land/lipgloss/v2\"\n\nvar estilo = lipgloss.NewStyle()\n",
			legado: true,
		},
		{
			nome:  "estilo fora do tema em cmd sem exceção",
			rel:   "cmd/gex-dev-kit/novo.go",
			src:   "package main\n\nimport \"charm.land/lipgloss/v2\"\n\nvar estilo = lipgloss.NewStyle()\n",
			falha: true,
		},
		{
			nome:  "ui importando kit",
			rel:   "internal/ui/screens/x.go",
			src:   "package screens\n\nimport _ \"" + archModule + "/internal/kit\"\n",
			falha: true,
		},
		{
			nome:  "ui importando agents",
			rel:   "internal/ui/components/x.go",
			src:   "package components\n\nimport _ \"" + archModule + "/internal/agents\"\n",
			falha: true,
		},
		{
			nome:  "app importando ui",
			rel:   "internal/app/x.go",
			src:   "package app\n\nimport _ \"" + archModule + "/internal/ui/theme\"\n",
			falha: true,
		},
	}
	for _, c := range casos {
		got := archCheckSource(t, c.rel, c.src, c.legado)
		if (len(got) > 0) != c.falha {
			t.Fatalf("%s: falhas = %v, esperado falha=%v", c.nome, got, c.falha)
		}
	}
}

func archParse(name string, src any) (*token.FileSet, *ast.File, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, name, src, parser.ParseComments)
	return fset, file, err
}

func archCheckSource(t *testing.T, rel, src string, legado bool) []string {
	t.Helper()
	fset, file, err := archParse(rel, src)
	if err != nil {
		t.Fatalf("parse de %s: %v", rel, err)
	}
	return archCheckFile(fset, rel, file, legado)
}

func archIsLegacy(rel string) bool {
	for _, item := range archLegacy {
		if item == filepath.ToSlash(rel) {
			return true
		}
	}
	return false
}

func archLegacyFalhas() []string {
	var falhas []string
	for _, rel := range archLegacy {
		p := filepath.Join(archRepoRoot, rel)
		if _, err := os.Stat(p); err != nil {
			falhas = append(falhas, rel+": exceção aponta para arquivo inexistente")
			continue
		}
		fset, file, err := archParse(p, nil)
		if err != nil {
			falhas = append(falhas, rel+": "+err.Error())
			continue
		}
		if len(archCheckFile(fset, rel, file, false)) == 0 {
			falhas = append(falhas, rel+": sem violação; remova o arquivo de archLegacy")
		}
	}
	return falhas
}

func archCheckFile(fset *token.FileSet, rel string, file *ast.File, legado bool) []string {
	var falhas []string
	add := func(pos token.Pos, msg string) {
		falhas = append(falhas, fmt.Sprintf("%s:%d: %s", rel, fset.Position(pos).Line, msg))
	}
	inUI := strings.HasPrefix(rel, "internal/ui/")
	inCmd := strings.HasPrefix(rel, "cmd/")
	estilo := (inUI && !strings.HasPrefix(rel, "internal/ui/theme/") || inCmd) && !legado
	nomes := archLipglossNames(file)

	for _, imp := range file.Imports {
		caminho := archImportPath(imp)
		if caminho == archCompat {
			add(imp.Pos(), "importa charm.land/lipgloss/v2/compat")
		}
		if inUI && (archImportIs(caminho, archModule+"/internal/kit") || archImportIs(caminho, archModule+"/internal/agents")) {
			add(imp.Pos(), "internal/ui importa "+caminho)
		}
		if strings.HasPrefix(rel, "internal/app/") && archImportIs(caminho, archModule+"/internal/ui") {
			add(imp.Pos(), "internal/app importa "+caminho)
		}
		if strings.HasPrefix(rel, archCmdDir) && !strings.HasSuffix(rel, "_test.go") && filepath.Base(rel) != "accessible.go" && !legado && (caminho == archLipgloss || caminho == archHuh) {
			add(imp.Pos(), "cmd/gex-dev-kit importa "+caminho+" fora de accessible.go")
		}
	}

	if inUI || inCmd {
		for _, grupo := range file.Comments {
			for _, c := range grupo.List {
				if !strings.HasPrefix(c.Text, "//go:build") && !strings.HasPrefix(c.Text, "//go:embed") {
					add(c.Pos(), "comentário no código")
				}
			}
		}
	}

	ast.Inspect(file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			sel, ok := x.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			ident, ok := sel.X.(*ast.Ident)
			if !ok || !nomes[ident.Name] {
				return true
			}
			nome := sel.Sel.Name
			if estilo && (nome == "NewStyle" || nome == "Color") {
				add(x.Pos(), "lipgloss."+nome+"( fora de internal/ui/theme")
			}
			if inUI && (strings.HasPrefix(nome, "Print") || strings.HasPrefix(nome, "Fprint") || strings.HasPrefix(nome, "Sprint")) {
				add(x.Pos(), "lipgloss."+nome+" liga a saída ao os.Stdout; use terminal.Env.Writer")
			}
		case *ast.BasicLit:
			if estilo && x.Kind == token.STRING && archHexLiteral.MatchString(x.Value) {
				add(x.Pos(), "cor hexadecimal literal fora de internal/ui/theme")
			}
		}
		return true
	})
	return falhas
}

func archLipglossNames(file *ast.File) map[string]bool {
	nomes := map[string]bool{}
	for _, imp := range file.Imports {
		if archImportPath(imp) != archLipgloss {
			continue
		}
		nome := "lipgloss"
		if imp.Name != nil {
			nome = imp.Name.Name
		}
		nomes[nome] = true
	}
	return nomes
}

func archImportPath(imp *ast.ImportSpec) string {
	caminho, _ := strconv.Unquote(imp.Path.Value)
	return caminho
}

func archImportIs(caminho, pacote string) bool {
	return caminho == pacote || strings.HasPrefix(caminho, pacote+"/")
}
