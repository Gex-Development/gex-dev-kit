package components

import (
	"fmt"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
	"github.com/charmbracelet/x/ansi"
)

type (
	PaletteItem      struct{ Name, Description string }
	PaletteChooseMsg struct{ Name string }
)

type Palette struct {
	items, matches                           []PaletteItem
	cursor, maxRows, originX, originY, hover int
}

func NewPalette(items []PaletteItem) Palette {
	return Palette{items: slices.Clone(items), matches: slices.Clone(items), maxRows: len(items), hover: -1}
}

func (p Palette) Filter(query string) Palette {
	selected, ok := p.Selected()
	query = strings.ToLower(strings.TrimPrefix(query, "/"))
	p.matches = nil
	for _, item := range p.items {
		if strings.HasPrefix(strings.ToLower(item.Name), query) {
			p.matches = append(p.matches, item)
		}
	}
	if query != "" {
		for _, item := range p.items {
			if !strings.HasPrefix(strings.ToLower(item.Name), query) && (strings.Contains(strings.ToLower(item.Name), query) || strings.Contains(strings.ToLower(item.Description), query)) {
				p.matches = append(p.matches, item)
			}
		}
	}
	p.cursor, p.hover = 0, -1
	if ok {
		for i, item := range p.matches {
			if item == selected {
				if i < p.visible() {
					p.cursor = i
				}
				break
			}
		}
	}
	return p
}

func (p Palette) Len() int { return len(p.matches) }

func (p Palette) Selected() (PaletteItem, bool) {
	if p.cursor < 0 || p.cursor >= p.visible() {
		return PaletteItem{}, false
	}
	return p.matches[p.cursor], true
}

func (p Palette) MaxRows(n int) Palette {
	p.maxRows = max(1, n)
	if p.cursor >= p.visible() {
		p.cursor = 0
	}
	return p
}
func (p Palette) Place(x, y int) Palette { p.originX, p.originY = x, y; return p }
func (p Palette) visible() int           { return min(len(p.matches), max(1, p.maxRows)) }
func (p Palette) Height(width int) int {
	if width <= 0 {
		return 0
	}
	if len(p.matches) > p.visible() {
		return p.visible() + 1
	}
	return p.visible()
}

func (p Palette) View(t theme.Theme, width int) string {
	if width <= 0 || len(p.matches) == 0 {
		return ""
	}
	longest := 0
	for _, item := range p.matches {
		longest = max(longest, ansi.StringWidth(item.Name))
	}
	ellipsis := "…"
	if t.Plain {
		ellipsis = "..."
	}
	rows := make([]string, 0, p.Height(width))
	for i := 0; i < p.visible(); i++ {
		item := p.matches[i]
		prefix := "  "
		style := t.Text
		if i == p.cursor {
			prefix = "┃ "
			if t.Plain {
				prefix = "> "
			}
			prefix = t.Accent.Render(prefix)
			style = t.Accent.Bold(true)
		} else if i == p.hover {
			style = t.Strong
		}
		name, description := "/"+item.Name, item.Description
		if t.Plain {
			name, description = plainASCII(name), plainASCII(description)
		}
		row := prefix + style.Render(name)
		if description != "" {
			row += strings.Repeat(" ", max(0, longest-ansi.StringWidth(item.Name))) + "  " + t.Secondary.Render(description)
		}
		rows = append(rows, ansi.Truncate(row, width, ellipsis))
	}
	if len(p.matches) > p.visible() {
		rows = append(rows, t.Faint.Render(ansi.Truncate(fmt.Sprintf("%s mais %d", ellipsis, len(p.matches)-p.visible()), width, ellipsis)))
	}
	return strings.Join(rows, "\n")
}

func plainASCII(s string) string {
	return strings.Map(func(r rune) rune {
		if r > 127 {
			return '?'
		}
		return r
	}, s)
}

func (p Palette) move(delta int) Palette {
	if count := p.visible(); count > 0 {
		p.cursor = (p.cursor + delta + count) % count
	}
	return p
}

func (p Palette) Update(msg tea.Msg) (Palette, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.Code {
		case tea.KeyUp:
			return p.move(-1), nil
		case tea.KeyDown:
			return p.move(1), nil
		case tea.KeyEnter:
			if item, ok := p.Selected(); ok {
				return p, func() tea.Msg { return PaletteChooseMsg{Name: item.Name} }
			}
		}
	case tea.MouseWheelMsg:
		if msg.Button == tea.MouseWheelUp {
			return p.move(-1), nil
		}
		if msg.Button == tea.MouseWheelDown {
			return p.move(1), nil
		}
	case tea.MouseMotionMsg:
		p.hover = p.itemAt(msg.X, msg.Y)
	case tea.MouseClickMsg:
		if msg.Button == tea.MouseLeft {
			if i := p.itemAt(msg.X, msg.Y); i >= 0 {
				p.cursor = i
				item := p.matches[i]
				return p, func() tea.Msg { return PaletteChooseMsg{Name: item.Name} }
			}
		}
	}
	return p, nil
}

func (p Palette) itemAt(x, y int) int {
	if x < p.originX || y < p.originY || y >= p.originY+p.visible() {
		return -1
	}
	return y - p.originY
}
