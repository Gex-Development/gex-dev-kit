package components

import (
	"slices"
	"strings"

	"charm.land/lipgloss/v2/tree"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

type filetreeNode struct {
	name     string
	children []*filetreeNode
}

func FileTree(t theme.Theme, root string, paths []string) string {
	top := &filetreeNode{}
	for _, path := range paths {
		filetreeInsert(top, path)
	}
	view := tree.Root(root).
		Enumerator(filetreeEnumerator).
		EnumeratorStyle(t.Faint).
		ItemStyle(t.Text).
		RootStyle(t.Strong)
	filetreeFill(view, top)
	return view.String()
}

func filetreeEnumerator(children tree.Children, index int) string {
	return tree.RoundedEnumerator(children, index) + " "
}

func filetreeInsert(node *filetreeNode, path string) {
	current := node
	for _, segment := range strings.Split(path, "/") {
		if segment == "" {
			continue
		}
		current = filetreeChild(current, segment)
	}
}

func filetreeChild(node *filetreeNode, name string) *filetreeNode {
	for _, child := range node.children {
		if child.name == name {
			return child
		}
	}
	child := &filetreeNode{name: name}
	node.children = append(node.children, child)
	return child
}

func filetreeFill(view *tree.Tree, node *filetreeNode) {
	children := slices.Clone(node.children)
	slices.SortFunc(children, func(a, b *filetreeNode) int {
		return strings.Compare(a.name, b.name)
	})
	for _, child := range children {
		if len(child.children) == 0 {
			view.Child(child.name)
			continue
		}
		sub := tree.Root(child.name)
		filetreeFill(sub, child)
		view.Child(sub)
	}
}
