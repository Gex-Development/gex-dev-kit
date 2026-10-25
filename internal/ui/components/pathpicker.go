package components

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
	"github.com/charmbracelet/x/ansi"
)

func ExpandHome(value, home string) string {
	if value == "~" {
		return home
	}
	if strings.HasPrefix(value, "~/") {
		return filepath.Join(home, value[2:])
	}
	return value
}

type PathSuggestion struct {
	Name, Path string
	Marked     bool
}

func PathSuggestions(value, home string, markers []string, limit int) []PathSuggestion {
	if limit <= 0 {
		return nil
	}
	dir, prefix, displayDir := ExpandHome(value, home), "", value
	if value == "" {
		dir, displayDir = ".", ""
	} else if value == "~" {
		displayDir = "~/"
	} else if !strings.HasSuffix(value, "/") {
		if cut := strings.LastIndexByte(value, '/'); cut >= 0 {
			dir, prefix, displayDir = ExpandHome(value[:cut+1], home), value[cut+1:], value[:cut+1]
		} else {
			dir, prefix, displayDir = ".", value, ""
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	result := make([]PathSuggestion, 0, min(len(entries), limit))
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") && !strings.HasPrefix(prefix, ".") {
			continue
		}
		if !strings.HasPrefix(strings.ToLower(name), strings.ToLower(prefix)) {
			continue
		}
		full := filepath.Join(dir, name)
		info, err := os.Stat(full)
		if err != nil || !info.IsDir() {
			continue
		}
		marked := false
		for _, marker := range markers {
			if _, err := os.Stat(filepath.Join(full, marker)); err == nil {
				marked = true
				break
			}
		}
		result = append(result, PathSuggestion{Name: name, Path: displayDir + name + "/", Marked: marked})
		if len(result) == limit {
			break
		}
	}
	return result
}

type PathPickMsg struct{ Path string }

type PathPicker struct {
	home                              string
	markers                           []string
	suggestions                       []PathSuggestion
	cursor, maxRows, originX, originY int
}

func NewPathPicker(home string, markers []string) PathPicker {
	return PathPicker{home: home, markers: append([]string(nil), markers...), cursor: -1, maxRows: 5}
}

func (p PathPicker) Sync(value string) PathPicker {
	p.suggestions = PathSuggestions(value, p.home, p.markers, int(^uint(0)>>1))
	p.cursor = -1
	return p
}

func (p PathPicker) Len() int { return len(p.suggestions) }

func (p PathPicker) Highlighted() (PathSuggestion, bool) {
	if p.cursor < 0 || p.cursor >= p.visible() {
		return PathSuggestion{}, false
	}
	return p.suggestions[p.cursor], true
}

func (p PathPicker) Complete(value string) string {
	if item, ok := p.Highlighted(); ok {
		return item.Path
	}
	if len(p.suggestions) == 0 {
		return value
	}
	if len(p.suggestions) == 1 {
		return p.suggestions[0].Path
	}
	common := []rune(p.suggestions[0].Name)
	for _, item := range p.suggestions[1:] {
		name := []rune(item.Name)
		i := 0
		for i < len(common) && i < len(name) && common[i] == name[i] {
			i++
		}
		common = common[:i]
	}
	return strings.TrimSuffix(p.suggestions[0].Path, p.suggestions[0].Name+"/") + string(common)
}

func (p PathPicker) Place(x, y int) PathPicker { p.originX, p.originY = x, y; return p }
func (p PathPicker) MaxRows(n int) PathPicker {
	p.maxRows = max(1, n)
	if p.cursor >= p.visible() {
		p.cursor = -1
	}
	return p
}
func (p PathPicker) visible() int { return min(len(p.suggestions), max(1, p.maxRows)) }
func (p PathPicker) Height(width int) int {
	if width <= 0 {
		return 0
	}
	if len(p.suggestions) == 0 {
		return 1
	}
	if len(p.suggestions) > p.visible() {
		return p.visible() + 1
	}
	return p.visible()
}

func (p PathPicker) View(t theme.Theme, width int) string {
	if width <= 0 {
		return ""
	}
	ellipsis, more, none, mark := "…", "… mais ", "(nenhuma pasta aqui)", " ●"
	if t.Plain {
		ellipsis, more, mark = "...", "... mais ", " *"
	}
	if len(p.suggestions) == 0 {
		return t.Faint.Render(ansi.Truncate(none, width, ellipsis))
	}
	rows := make([]string, 0, p.Height(width))
	for i := 0; i < p.visible(); i++ {
		item := p.suggestions[i]
		prefix := "  "
		style := t.Text
		if i == p.cursor {
			prefix = "┃ "
			if t.Plain {
				prefix = "> "
			}
			prefix = t.Accent.Render(prefix)
			style = t.Accent.Bold(true)
		}
		name := item.Name + "/"
		if t.Plain {
			name = plainASCII(name)
		}
		row := prefix + style.Render(name)
		if item.Marked {
			row += t.Success.Render(mark)
		}
		rows = append(rows, ansi.Truncate(row, width, ellipsis))
	}
	if len(p.suggestions) > p.visible() {
		rows = append(rows, t.Faint.Render(ansi.Truncate(more+strconv.Itoa(len(p.suggestions)-p.visible()), width, ellipsis)))
	}
	return strings.Join(rows, "\n")
}

func (p PathPicker) move(delta int) PathPicker {
	count := p.visible()
	if count > 0 {
		if p.cursor < 0 {
			if delta < 0 {
				p.cursor = count - 1
			} else {
				p.cursor = 0
			}
		} else {
			p.cursor = (p.cursor + delta + count) % count
		}
	}
	return p
}

func (p PathPicker) Update(msg tea.Msg) (PathPicker, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.Code {
		case tea.KeyUp:
			return p.move(-1), nil
		case tea.KeyDown:
			return p.move(1), nil
		}
	case tea.MouseWheelMsg:
		if msg.Button == tea.MouseWheelUp {
			return p.move(-1), nil
		}
		if msg.Button == tea.MouseWheelDown {
			return p.move(1), nil
		}
	case tea.MouseClickMsg:
		if msg.Button == tea.MouseLeft && msg.X >= p.originX && msg.Y >= p.originY && msg.Y < p.originY+p.visible() {
			p.cursor = msg.Y - p.originY
			item := p.suggestions[p.cursor]
			return p, func() tea.Msg { return PathPickMsg{Path: item.Path} }
		}
	}
	return p, nil
}
