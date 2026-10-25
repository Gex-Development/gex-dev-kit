package components

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
	"github.com/charmbracelet/x/ansi"
)

const diffviewBadgeWidth = 13

type ChangeAction int

const (
	ChangeCreate ChangeAction = iota
	ChangeUpdate
	ChangeRemove
	ChangeRestore
	ChangeBlocked
	ChangeUnchanged
)

type DiffLineKind int

const (
	DiffContext DiffLineKind = iota
	DiffAdded
	DiffRemoved
	DiffHunk
)

type DiffLine struct {
	Kind DiffLineKind
	Text string
}

type Change struct {
	Action  ChangeAction
	Path    string
	Detail  string
	Lines   []DiffLine
	Omitted int
	Masked  int
}

type ChangeGroup struct {
	Title   string
	Changes []Change
}

func DiffView(t theme.Theme, width int, groups []ChangeGroup, expanded map[string]bool) string {
	var blocks []string
	for _, group := range groups {
		title := diffviewText(group.Title)
		count := fmt.Sprintf("%d %s", len(group.Changes), diffviewPlural(len(group.Changes), "arquivo", "arquivos"))
		blocks = append(blocks, t.Title.Render(ansi.Truncate(title, max(1, width), "…"))+"  "+t.Secondary.Render(count))
		for _, change := range group.Changes {
			blocks = append(blocks, diffviewSummary(t, width, change))
			if body := DiffBody(t, width, change); expanded[change.Path] && body != "" {
				blocks = append(blocks, body)
			}
		}
		blocks = append(blocks, "")
	}
	return strings.TrimSuffix(strings.Join(blocks, "\n"), "\n")
}

func DiffBody(t theme.Theme, width int, c Change) string {
	var lines []string
	for _, line := range c.Lines {
		lines = append(lines, diffviewLine(t, width, line))
	}
	if c.Omitted > 0 {
		note := fmt.Sprintf("… %d %s (abra na TUI ou use a prévia completa)", c.Omitted, diffviewPlural(c.Omitted, "linha omitida", "linhas omitidas"))
		lines = append(lines, t.Secondary.Render(diffviewWrap(note, width)))
	}
	if c.Masked > 0 {
		note := fmt.Sprintf("%d %s por parecerem segredo", c.Masked, diffviewPlural(c.Masked, "valor mascarado", "valores mascarados"))
		lines = append(lines, t.Secondary.Render(diffviewWrap(note, width)))
	}
	return strings.Join(lines, "\n")
}

func diffviewSummary(t theme.Theme, width int, c Change) string {
	badge := diffviewBadge(t, c.Action)
	room := max(1, width-diffviewBadgeWidth-1)
	text := diffviewText(c.Path)
	if detail := diffviewText(c.Detail); detail != "" {
		text += " · " + detail
	}
	lines := strings.Split(ansi.Wrap(text, room, " "), "\n")
	indent := strings.Repeat(" ", diffviewBadgeWidth+1)
	for i, line := range lines {
		prefix := indent
		if i == 0 {
			prefix = badge + " "
		}
		lines[i] = prefix + t.Text.Render(line)
	}
	return strings.Join(lines, "\n")
}

func diffviewBadge(t theme.Theme, action ChangeAction) string {
	style := diffviewActionStyle(t, action).Width(diffviewBadgeWidth).Align(lipgloss.Center)
	if t.Rich {
		style = style.Reverse(true)
	}
	return style.Render(ChangeLabel(action))
}

func diffviewLine(t theme.Theme, width int, line DiffLine) string {
	text := diffviewText(line.Text)
	prefix, style := diffviewKind(t, line.Kind)
	if line.Kind == DiffHunk {
		if strings.HasPrefix(text, "@@") {
			prefix = ""
		} else {
			prefix = "@@ "
		}
	}
	return style.Render(ansi.Hardwrap(prefix+text, max(1, width), true))
}

func diffviewKind(t theme.Theme, kind DiffLineKind) (string, lipgloss.Style) {
	switch kind {
	case DiffAdded:
		return "+", t.Success
	case DiffRemoved:
		return "-", t.Error
	case DiffHunk:
		return "", t.Info
	default:
		return " ", t.Secondary
	}
}

func diffviewActionStyle(t theme.Theme, action ChangeAction) lipgloss.Style {
	switch action {
	case ChangeCreate:
		return t.Success
	case ChangeUpdate, ChangeRestore:
		return t.Info
	case ChangeRemove:
		return t.Warning
	case ChangeBlocked:
		return t.Error
	default:
		return t.Secondary
	}
}

func ChangeLabel(action ChangeAction) string {
	switch action {
	case ChangeCreate:
		return "criar"
	case ChangeUpdate:
		return "atualizar"
	case ChangeRemove:
		return "remover"
	case ChangeRestore:
		return "restaurar"
	case ChangeBlocked:
		return "bloqueado"
	default:
		return "sem mudança"
	}
}

func diffviewText(s string) string {
	s = ansi.Strip(s)
	s = strings.ReplaceAll(s, "\r", "")
	return strings.ReplaceAll(s, "\t", "    ")
}

func diffviewWrap(s string, width int) string {
	return ansi.Wrap(s, max(1, width), " ")
}

func diffviewPlural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}
