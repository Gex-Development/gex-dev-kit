package components

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

const (
	buttonsGap     = 2
	buttonsPadding = 2
)

type Button struct {
	ID    string
	Label string
}

type ButtonPressMsg struct {
	Index  int
	Button Button
}

type Buttons struct {
	items   []Button
	focus   int
	hover   int
	originX int
	originY int
	keys    Keys
}

func NewButtons(items ...Button) Buttons {
	return Buttons{items: items, hover: -1, keys: DefaultKeys()}
}

func (b Buttons) Place(x, y int) Buttons {
	b.originX, b.originY = x, y
	return b
}

func (b Buttons) Focus() int {
	return b.focus
}

func (b Buttons) SetFocus(i int) Buttons {
	if i >= 0 && i < len(b.items) {
		b.focus = i
	}
	return b
}

func (b Buttons) Width() int {
	total := 0
	for i, item := range b.items {
		if i > 0 {
			total += buttonsGap
		}
		total += buttonsLabelWidth(item)
	}
	return total
}

func (b Buttons) Update(msg tea.Msg) (Buttons, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return b.updateKey(msg)
	case tea.MouseClickMsg:
		if msg.Button != tea.MouseLeft {
			return b, nil
		}
		if i := b.at(msg.X, msg.Y); i >= 0 {
			b.focus = i
			return b, b.press()
		}
	case tea.MouseMotionMsg:
		b.hover = b.at(msg.X, msg.Y)
	}
	return b, nil
}

func (b Buttons) updateKey(msg tea.KeyPressMsg) (Buttons, tea.Cmd) {
	n := len(b.items)
	if n == 0 {
		return b, nil
	}
	switch {
	case key.Matches(msg, b.keys.Left), msg.String() == "shift+tab":
		b.focus = (b.focus - 1 + n) % n
	case key.Matches(msg, b.keys.Right, b.keys.Tab):
		b.focus = (b.focus + 1) % n
	case key.Matches(msg, b.keys.Enter):
		return b, b.press()
	}
	return b, nil
}

func (b Buttons) press() tea.Cmd {
	press := ButtonPressMsg{Index: b.focus, Button: b.items[b.focus]}
	return func() tea.Msg { return press }
}

func (b Buttons) at(x, y int) int {
	if y != b.originY || x < b.originX {
		return -1
	}
	pos := b.originX
	for i, item := range b.items {
		width := buttonsLabelWidth(item)
		if x >= pos && x < pos+width {
			return i
		}
		pos += width + buttonsGap
	}
	return -1
}

func (b Buttons) View(t theme.Theme) string {
	parts := make([]string, len(b.items))
	for i, item := range b.items {
		parts[i] = buttonsRender(t, item, i == b.focus, i == b.hover)
	}
	return strings.Join(parts, strings.Repeat(" ", buttonsGap))
}

func buttonsLabelWidth(item Button) int {
	return ansi.StringWidth(item.Label) + 2*buttonsPadding
}

func buttonsRender(t theme.Theme, item Button, focused, hovered bool) string {
	if t.Plain {
		if focused {
			return "> " + item.Label + " <"
		}
		return "[ " + item.Label + " ]"
	}
	switch {
	case focused:
		return t.Pill.Padding(0, buttonsPadding).Render(item.Label)
	case hovered:
		return t.BarFocus.Padding(0, buttonsPadding).Render(item.Label)
	}
	return t.Bar.Padding(0, buttonsPadding).Render(item.Label)
}
