package components

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

const (
	confirmMinWidth  = 24
	confirmMinHeight = 11
	confirmMaxWidth  = 60
)

type Decision int

const (
	DecisionNone Decision = iota
	DecisionYes
	DecisionNo
)

type Confirm struct {
	title, body, yes, no string
	keys                 Keys
	buttons              Buttons
}

func NewConfirm(title, body, yes, no string) Confirm {
	return Confirm{
		title: title, body: body, yes: yes, no: no, keys: DefaultKeys(),
		buttons: NewButtons(Button{ID: "yes", Label: yes}, Button{ID: "no", Label: no}),
	}
}

func (c Confirm) Place(t theme.Theme, width, height int) Confirm {
	modal := c.box(t, min(width-1, confirmMaxWidth))
	x, y := max((width-lipgloss.Width(modal))/2, 0), max((height-lipgloss.Height(modal))/2, 0)
	buttons := ansi.Strip(c.buttons.View(t))
	for row, line := range strings.Split(ansi.Strip(modal), "\n") {
		if col := strings.Index(line, buttons); col >= 0 {
			c.buttons = c.buttons.Place(x+ansi.StringWidth(line[:col]), y+row)
			break
		}
	}
	return c
}

func (c Confirm) Visible(width, height int) bool {
	return width >= confirmMinWidth && height >= confirmMinHeight
}

func (c Confirm) Update(msg tea.Msg, visible bool) (Confirm, Decision) {
	if !visible {
		return c, DecisionNone
	}
	if press, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(press, c.keys.Yes):
			return c, DecisionYes
		case key.Matches(press, c.keys.No, c.keys.Back):
			return c, DecisionNo
		case press.String() == "enter":
			return c, DecisionNone
		}
	}
	var cmd tea.Cmd
	c.buttons, cmd = c.buttons.Update(msg)
	if cmd != nil {
		if pressed, ok := cmd().(ButtonPressMsg); ok {
			if pressed.Button.ID == "yes" {
				return c, DecisionYes
			}
			return c, DecisionNo
		}
	}
	return c, DecisionNone
}

func (c Confirm) View(t theme.Theme, width int) string {
	if width <= 0 {
		return ""
	}
	return confirmFit(c.box(t, min(width, confirmMaxWidth)), width, 100, confirmTail(t))
}

func (c Confirm) Overlay(t theme.Theme, base string, width, height int) string {
	if !c.Visible(width, height) {
		return c.View(t, width)
	}
	modal := c.box(t, min(width-1, confirmMaxWidth))
	mw, mh := lipgloss.Width(modal), lipgloss.Height(modal)
	x, y := max((width-mw)/2, 0), max((height-mh)/2, 0)
	layers := []*lipgloss.Layer{lipgloss.NewLayer(base)}
	if sw, sh := min(mw, width-x-1), min(mh, height-y-1); sw > 0 && sh > 0 {
		layers = append(layers, lipgloss.NewLayer(confirmShadow(t, sw, sh)).X(x+1).Y(y+1).Z(1))
	}
	layers = append(layers, lipgloss.NewLayer(modal).X(x).Y(y).Z(2))
	return confirmFit(lipgloss.NewCompositor(layers...).Render(), width, height, confirmTail(t))
}

func confirmFit(s string, width, height int, tail string) string {
	lines := strings.Split(s, "\n")
	if len(lines) > height {
		lines = lines[:height]
	}
	for i, line := range lines {
		lines[i] = ansi.Truncate(line, width, tail)
	}
	return strings.Join(lines, "\n")
}

func confirmTail(t theme.Theme) string {
	if t.Plain {
		return "..."
	}
	return "…"
}

func (c Confirm) box(t theme.Theme, w int) string {
	box := t.Modal
	if t.Plain {
		box = box.Border(lipgloss.ASCIIBorder())
	} else {
		box = box.Border(lipgloss.DoubleBorder())
		if t.Rich {
			box = box.BorderForegroundBlend(t.Palette.Gradient...)
		}
	}
	return box.Width(w).Render(c.content(t))
}

func (c Confirm) content(t theme.Theme) string {
	hints := t.KeyHintKey.Render("s/y") + " " + t.KeyHintDesc.Render(c.yes) +
		"    " + t.KeyHintKey.Render("n/esc") + " " + t.KeyHintDesc.Render(c.no)
	return strings.Join([]string{t.Title.Render(c.title), "", c.body, "", c.buttons.View(t), hints}, "\n")
}

func confirmShadow(t theme.Theme, w, h int) string {
	fill := "░"
	if t.Plain {
		fill = "."
	}
	line := strings.Repeat(fill, w)
	rows := make([]string, h)
	for i := range rows {
		rows[i] = line
	}
	return t.Divider.Render(strings.Join(rows, "\n"))
}
