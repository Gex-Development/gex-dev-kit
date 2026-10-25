package components

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
	"github.com/charmbracelet/x/ansi"
)

type (
	ChoiceOption struct {
		ID, Label, Hint, Icon string
		Art                   func(theme.Theme) []string
	}
	ChoiceDoneMsg struct{ IDs []string }
	ChoiceBackMsg struct{}
)

type Choice struct {
	question                                      string
	options                                       []ChoiceOption
	multi, required                               bool
	cursor                                        int
	chosen                                        map[int]bool
	originX, originY, maxRows, hover, width, page int
	spaced                                        bool
	art                                           bool
}

func NewChoice(question string, options []ChoiceOption, multi, required bool) Choice {
	rows := len(options)
	if multi {
		rows++
	}
	return Choice{question: question, options: slices.Clone(options), multi: multi, required: required, chosen: make(map[int]bool), maxRows: rows, hover: -1}
}

func (c Choice) Preselect(ids []string) Choice {
	c.chosen = make(map[int]bool)
	for i, option := range c.options {
		for _, id := range ids {
			if id == option.ID {
				c.chosen[i] = true
			}
		}
	}
	return c
}

func (c Choice) Place(x, y int, width ...int) Choice {
	c.originX, c.originY = x, y
	if len(width) > 0 {
		c.width = width[0]
	}
	return c
}

func (c Choice) MaxRows(n int) Choice {
	c.maxRows = max(1, n)
	c.page = min(c.page, c.Pages()-1)
	if !c.multi || c.cursor != len(c.options) {
		if c.cursor < c.pageStart() || c.cursor >= c.pageEnd() {
			c.cursor = c.pageStart()
		}
	}
	return c
}
func (c Choice) Spaced() Choice         { c.spaced = true; return c }
func (c Choice) WithArt(on bool) Choice { c.art = on; return c }
func (c Choice) Multi() bool            { return c.multi }
func (c Choice) Pages() int             { return max(1, (len(c.options)+max(1, c.maxRows)-1)/max(1, c.maxRows)) }
func (c Choice) Page() int              { return c.page }
func (c Choice) pageStart() int         { return c.page * max(1, c.maxRows) }
func (c Choice) pageEnd() int           { return min(len(c.options), c.pageStart()+max(1, c.maxRows)) }
func (c Choice) rowsOf(i, width int) int {
	if i >= len(c.options) || !c.art || c.options[i].Art == nil {
		return 1
	}
	art := c.options[i].Art(theme.Theme{Plain: true})
	if len(art) == 0 {
		return 1
	}
	number := fmt.Sprintf("%d. ", i-c.pageStart()+1)
	if c.multi {
		number += "[ ] "
	}
	if width > 0 && width < 2+ansi.StringWidth(number)+ansi.StringWidth(art[0])+1 {
		return 1
	}
	return len(art)
}

func (c Choice) visible() int {
	count := c.pageEnd() - c.pageStart()
	if c.multi {
		count++
	}
	return count
}

func (c Choice) Height(width int) int {
	if width <= 0 {
		return 0
	}
	h := 2
	start, end := c.pageStart(), c.pageEnd()
	for i := start; i < end; i++ {
		h += c.rowsOf(i, width)
		if i == c.cursor && c.options[i].Hint != "" {
			h++
		}
		if c.spaced && (i < end-1 || c.multi) {
			h++
		}
	}
	if c.multi {
		h++
		if !c.spaced && end > start {
			h++
		}
	}
	if c.Pages() > 1 {
		h++
	}
	return h
}

func (c Choice) View(t theme.Theme, width int) string {
	if width <= 0 {
		return ""
	}
	ellipsis := "…"
	if t.Plain {
		ellipsis = "..."
	}
	badge := t.Badge.Render("?")
	if t.Plain {
		badge = "[?]"
	}
	rows := []string{ansi.Truncate(badge+" "+t.Strong.Render(c.question), width, ellipsis)}
	for row := c.pageStart(); row < c.pageEnd()+btoi(c.multi); row++ {
		i := row
		if row == c.pageEnd() {
			i = len(c.options)
		}
		prefix := "  "
		if i == c.cursor {
			prefix = "┃ "
			if t.Plain {
				prefix = "> "
			}
			prefix = t.Accent.Render(prefix)
		}
		if c.multi && i == len(c.options) {
			label := "Marque uma opção"
			n := len(c.Selected())
			if n > 0 {
				label = fmt.Sprintf("Confirmar (%d)", n)
			} else if !c.required {
				label = "Pular (nenhuma marcada)"
			}
			button := t.Badge.Render(" " + label + " ")
			if t.Plain {
				button = "[ " + label + " ]"
			}
			rows = append(rows, ansi.Truncate(prefix+button, width, ellipsis))
			continue
		}
		name := fmt.Sprintf("%d. ", i-c.pageStart()+1)
		if c.multi {
			mark := "☐ "
			if c.chosen[i] {
				mark = "☑ "
			}
			if t.Plain {
				mark = "[ ] "
				if c.chosen[i] {
					mark = "[x] "
				}
			}
			name += mark
		}
		labelStyle := t.Text
		if i == c.cursor {
			labelStyle = t.Accent.Bold(true)
		} else if i == c.hover {
			labelStyle = t.Strong
		}
		artRows := c.rowsOf(i, width)
		var art []string
		if c.art && c.options[i].Art != nil {
			art = c.options[i].Art(t)
			if len(art) == 0 || len(art) != artRows || width < 2+ansi.StringWidth(name)+ansi.StringWidth(art[0])+1 {
				art = nil
			}
		}
		if len(art) > 0 {
			indent := strings.Repeat(" ", ansi.StringWidth(name))
			for row, image := range art {
				lead := indent
				if row == len(art)/2 {
					lead = name
				}
				text := prefix + lead + image
				if row == len(art)/2 {
					text += " " + ansi.Truncate(labelStyle.Render(c.options[i].Label), max(0, width-ansi.StringWidth(text)-1), ellipsis)
				}
				rows = append(rows, text)
			}
		} else {
			if c.options[i].Icon != "" {
				name += c.options[i].Icon + " "
			}
			rows = append(rows, ansi.Truncate(prefix+name+labelStyle.Render(c.options[i].Label), width, ellipsis))
		}
		if i == c.cursor && c.options[i].Hint != "" {
			marker := "    ↳ "
			if t.Plain {
				marker = "    -> "
			}
			rows = append(rows, t.Secondary.Render(ansi.Truncate(marker+c.options[i].Hint, width, ellipsis)))
		}
		if (c.spaced && row < c.pageEnd()-1) || (c.multi && row == c.pageEnd()-1) {
			rows = append(rows, "")
		}
	}
	if c.Pages() > 1 {
		left, right := "‹", "›"
		label := fmt.Sprintf(" página %d/%d ", c.page+1, c.Pages())
		if t.Plain {
			left, right = "<", ">"
			label = fmt.Sprintf(" pagina %d/%d ", c.page+1, c.Pages())
		}
		if c.page > 0 {
			left = t.Accent.Render(left)
		}
		if c.page < c.Pages()-1 {
			right = t.Accent.Render(right)
		}
		rows = append(rows, t.Faint.Render(ansi.Truncate(left+label+right, width, ellipsis)))
	}
	hint := "↑/↓ mover · 1-9 escolher · enter confirmar · esc voltar"
	if c.multi {
		hint = "↑/↓ mover · espaço marca · enter confirma · esc voltar"
	}
	if c.Pages() > 1 {
		hint = "↑/↓ mover · ←/→ página · 1-9 escolher · enter confirmar · esc voltar"
		if c.multi {
			hint = "↑/↓ mover · ←/→ página · espaço marca · enter confirma · esc voltar"
		}
	}
	if t.Plain {
		hint = "up/down mover | 1-9 escolher | enter confirmar | esc voltar"
		if c.multi {
			hint = "up/down mover | espaco marca | enter confirma | esc voltar"
		}
		if c.Pages() > 1 {
			hint = "up/down mover | left/right pagina | 1-9 escolher | enter confirmar | esc voltar"
			if c.multi {
				hint = "up/down mover | left/right pagina | espaco marca | enter confirma | esc voltar"
			}
		}
	}
	rows = append(rows, t.Faint.Render(ansi.Truncate(hint, width, ellipsis)))
	return strings.Join(rows, "\n")
}

func btoi(value bool) int {
	if value {
		return 1
	}
	return 0
}

func (c Choice) changePage(delta int) Choice {
	next := c.page + delta
	if next >= 0 && next < c.Pages() {
		c.page = next
		c.cursor = c.pageStart()
		c.hover = -1
	}
	return c
}

func (c Choice) move(delta int) Choice {
	start, count := c.pageStart(), c.visible()
	if count > 0 {
		position := c.cursor - start
		if c.multi && c.cursor == len(c.options) {
			position = count - 1
		}
		position = (position + delta + count) % count
		c.cursor = start + position
		if c.multi && position == count-1 {
			c.cursor = len(c.options)
		}
	}
	return c
}

func (c Choice) Selected() []string {
	ids := make([]string, 0, len(c.chosen))
	for i, option := range c.options {
		if c.chosen[i] {
			ids = append(ids, option.ID)
		}
	}
	return ids
}

func (c Choice) confirm() (Choice, tea.Cmd) {
	if c.multi {
		ids := c.Selected()
		if c.required && len(ids) == 0 {
			if c.cursor >= len(c.options) {
				return c, nil
			}
			ids = []string{c.options[c.cursor].ID}
		}
		return c, func() tea.Msg { return ChoiceDoneMsg{IDs: ids} }
	}
	if len(c.options) == 0 {
		return c, nil
	}
	ids := []string{c.options[c.cursor].ID}
	return c, func() tea.Msg { return ChoiceDoneMsg{IDs: ids} }
}

func (c Choice) toggle(i int) Choice {
	c.chosen = maps.Clone(c.chosen)
	c.chosen[i] = !c.chosen[i]
	return c
}

func (c Choice) optionAt(x, y int) int {
	if x < c.originX || y < c.originY {
		return -1
	}
	row := 0
	start, end := c.pageStart(), c.pageEnd()
	for position := start; position < end+btoi(c.multi); position++ {
		i := position
		if position == end {
			i = len(c.options)
		}
		size := c.rowsOf(i, c.width)
		if y-c.originY >= row && y-c.originY < row+size {
			return i
		}
		row += size
		if i == c.cursor && i < len(c.options) && c.options[i].Hint != "" {
			row++
		}
		if (c.spaced && position < end-1) || (c.multi && position == end-1) {
			row++
		}
	}
	return -1
}

func (c Choice) Update(msg tea.Msg) (Choice, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.Code {
		case tea.KeyUp, 'k':
			return c.move(-1), nil
		case tea.KeyLeft, tea.KeyPgUp:
			return c.changePage(-1), nil
		case tea.KeyRight, tea.KeyPgDown:
			return c.changePage(1), nil
		case tea.KeyDown, 'j':
			return c.move(1), nil
		case tea.KeyEscape:
			return c, func() tea.Msg { return ChoiceBackMsg{} }
		case tea.KeyEnter:
			return c.confirm()
		case ' ':
			if c.multi && c.cursor < len(c.options) {
				return c.toggle(c.cursor), nil
			}
			return c, nil
		}
		if msg.Code >= '1' && msg.Code <= '9' {
			i := c.pageStart() + int(msg.Code-'1')
			if i < c.pageEnd() {
				c.cursor = i
				if c.multi {
					return c.toggle(i), nil
				}
				return c.confirm()
			}
		}
	case tea.MouseClickMsg:
		if msg.Button != tea.MouseLeft {
			return c, nil
		}
		if c.Pages() > 1 && msg.X >= c.originX {
			row := c.Height(max(1, c.width)) - 3
			if msg.Y == c.originY+row {
				label := fmt.Sprintf(" página %d/%d ", c.page+1, c.Pages())
				if msg.X == c.originX {
					return c.changePage(-1), nil
				}
				if msg.X == c.originX+1+ansi.StringWidth(label) {
					return c.changePage(1), nil
				}
			}
		}
		i := c.optionAt(msg.X, msg.Y)
		if i >= 0 {
			if c.multi {
				c.cursor = i
				if i == len(c.options) {
					return c.confirm()
				}
				return c.toggle(i), nil
			}
			if i != c.cursor {
				c.cursor = i
				return c, nil
			}
			return c.confirm()
		}
	case tea.MouseMotionMsg:
		c.hover = c.optionAt(msg.X, msg.Y)
	case tea.MouseWheelMsg:
		if msg.Button == tea.MouseWheelUp {
			return c.move(-1), nil
		}
		if msg.Button == tea.MouseWheelDown {
			return c.move(1), nil
		}
	}
	return c, nil
}
