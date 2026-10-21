package components

import (
	"math"
	"slices"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/motion"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

type MenuItem struct {
	ID, Title, Description string
	Disabled               bool
}

type MenuChoiceMsg struct {
	Index int
	Item  MenuItem
}

type Menu struct {
	items   []MenuItem
	cursor  int
	spring  motion.Spring
	keys    Keys
	frame   int
	originX int
	originY int
}

func NewMenu(items []MenuItem, animated bool) Menu {
	m := Menu{
		items:  slices.Clone(items),
		keys:   DefaultKeys(),
		spring: motion.NewSpring(animated, motion.Snappy),
	}
	m.cursor = menuFirstEnabled(m.items)
	m.spring = m.spring.Snap(float64(m.cursor))
	return m
}

func (m Menu) Update(msg tea.Msg) (Menu, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return m.updateKey(msg)
	case tea.MouseClickMsg:
		return m.updateClick(msg)
	case tea.MouseWheelMsg:
		return m.updateWheel(msg)
	case motion.FrameMsg:
		return m.updateFrame(msg)
	}
	return m, nil
}

func (m Menu) Selected() MenuItem {
	if m.cursor < 0 || m.cursor >= len(m.items) {
		return MenuItem{}
	}
	return m.items[m.cursor]
}

func (m Menu) Index() int {
	return m.cursor
}

func (m Menu) Animating() bool {
	return !m.spring.Settled()
}

func (m Menu) Place(x, y int) Menu {
	m.originX, m.originY = x, y
	return m
}

func (m Menu) View(t theme.Theme, width int, descriptions bool) string {
	if width <= 0 {
		return ""
	}
	rows := make([]string, 0, len(m.items))
	for i, item := range m.items {
		rows = append(rows, m.row(t, width, i, item, descriptions))
	}
	return strings.Join(rows, "\n")
}

func (m Menu) updateKey(msg tea.KeyPressMsg) (Menu, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Up):
		return m.move(-1)
	case key.Matches(msg, m.keys.Down):
		return m.move(1)
	case key.Matches(msg, m.keys.Enter):
		return m.choose()
	}
	return m, nil
}

func (m Menu) updateClick(msg tea.MouseClickMsg) (Menu, tea.Cmd) {
	if msg.Button != tea.MouseLeft || msg.X < m.originX {
		return m, nil
	}
	row := msg.Y - m.originY
	if row < 0 || row >= len(m.items) || m.items[row].Disabled || row == m.cursor {
		return m, nil
	}
	return m.target(row)
}

func (m Menu) updateWheel(msg tea.MouseWheelMsg) (Menu, tea.Cmd) {
	switch msg.Button {
	case tea.MouseWheelUp:
		return m.move(-1)
	case tea.MouseWheelDown:
		return m.move(1)
	}
	return m, nil
}

func (m Menu) updateFrame(msg motion.FrameMsg) (Menu, tea.Cmd) {
	if msg.ID != m.frame || m.spring.Settled() {
		return m, nil
	}
	m.spring = m.spring.Step()
	if m.spring.Settled() {
		return m, nil
	}
	return m, motion.Frame(m.frame)
}

func (m Menu) move(delta int) (Menu, tea.Cmd) {
	next := menuStep(m.items, m.cursor, delta)
	if next == m.cursor {
		return m, nil
	}
	return m.target(next)
}

func (m Menu) target(i int) (Menu, tea.Cmd) {
	wasSettled := m.spring.Settled()
	m.cursor = i
	m.spring = m.spring.Target(float64(i))
	if wasSettled && !m.spring.Settled() {
		m.frame++
		return m, motion.Frame(m.frame)
	}
	return m, nil
}

func (m Menu) choose() (Menu, tea.Cmd) {
	if m.cursor < 0 || m.cursor >= len(m.items) || m.items[m.cursor].Disabled {
		return m, nil
	}
	choice := MenuChoiceMsg{Index: m.cursor, Item: m.items[m.cursor]}
	return m, func() tea.Msg { return choice }
}

func (m Menu) row(t theme.Theme, width, i int, item MenuItem, descriptions bool) string {
	pointer := theme.StateRunning.Symbol(t.Plain)
	active := i == int(math.Round(m.spring.Value()))
	lead := strings.Repeat(" ", ansi.StringWidth(pointer))
	if active {
		lead = pointer
	}
	text := lead + " " + item.Title
	if descriptions && item.Description != "" {
		text += "  " + item.Description
	}
	text = ansi.Truncate(text, width, menuTail(t))
	switch {
	case item.Disabled:
		return t.Faint.Render(text)
	case active:
		return t.Selected.Render(text)
	default:
		return t.Unselected.Render(text)
	}
}

func menuTail(t theme.Theme) string {
	if t.Plain {
		return "..."
	}
	return "…"
}

func menuFirstEnabled(items []MenuItem) int {
	for i, item := range items {
		if !item.Disabled {
			return i
		}
	}
	return 0
}

func menuStep(items []MenuItem, from, delta int) int {
	for i := from + delta; i >= 0 && i < len(items); i += delta {
		if !items[i].Disabled {
			return i
		}
	}
	return from
}
