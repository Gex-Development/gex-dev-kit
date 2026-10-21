package components

import (
	"strings"
	"testing"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
)

func filetreeTestTheme() theme.Theme {
	return theme.New(theme.Options{Dark: true, Profile: colorprofile.ASCII})
}

func TestFileTreeGroupsSkillsByDirectory(t *testing.T) {
	paths := []string{".claude/skills/x/SKILL.md", ".agents/skills/x/SKILL.md"}
	got := ansi.Strip(FileTree(filetreeTestTheme(), "GEX", paths))
	want := strings.Join([]string{
		"GEX",
		"├── .agents",
		"│   ╰── skills",
		"│       ╰── x",
		"│           ╰── SKILL.md",
		"╰── .claude",
		"    ╰── skills",
		"        ╰── x",
		"            ╰── SKILL.md",
	}, "\n")
	if got != want {
		t.Fatalf("árvore agrupada errada:\n%s\nesperado:\n%s", got, want)
	}
}

func TestFileTreeSortsSiblingsAndKeepsFilesInTheirFolder(t *testing.T) {
	paths := []string{"b.txt", "a/z.md", "a/y.md"}
	got := ansi.Strip(FileTree(filetreeTestTheme(), "raiz", paths))
	want := strings.Join([]string{
		"raiz",
		"├── a",
		"│   ├── y.md",
		"│   ╰── z.md",
		"╰── b.txt",
	}, "\n")
	if got != want {
		t.Fatalf("ordem ou agrupamento errado:\n%s\nesperado:\n%s", got, want)
	}
}

func TestFileTreeWithoutPathsShowsOnlyRoot(t *testing.T) {
	if got := ansi.Strip(FileTree(filetreeTestTheme(), "vazio", nil)); got != "vazio" {
		t.Fatalf("árvore sem caminhos deveria mostrar só a raiz: %q", got)
	}
}

func TestFileTreeInsertIgnoresEmptySegments(t *testing.T) {
	root := &filetreeNode{}
	filetreeInsert(root, "/a//b/")
	filetreeInsert(root, "a/b")
	if len(root.children) != 1 || root.children[0].name != "a" {
		t.Fatalf("segmentos vazios criaram nós extras: %+v", root.children)
	}
	if b := root.children[0].children; len(b) != 1 || b[0].name != "b" || len(b[0].children) != 0 {
		t.Fatalf("caminho repetido duplicou o nó: %+v", b)
	}
}
