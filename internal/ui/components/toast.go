package components

import (
	"time"

	"charm.land/bubbles/v2/timer"
	tea "charm.land/bubbletea/v2"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

const toastTick = 100 * time.Millisecond

type Toast struct {
	clock   timer.Model
	text    string
	kind    theme.State
	visible bool
}

func NewToast() Toast {
	return Toast{}
}

func (x Toast) Show(text string, kind theme.State, d time.Duration) (Toast, tea.Cmd) {
	x.clock = timer.New(max(d, toastTick), timer.WithInterval(toastTick))
	x.text = text
	x.kind = kind
	x.visible = true
	return x, x.clock.Init()
}

func (x Toast) Update(msg tea.Msg) (Toast, tea.Cmd) {
	var cmd tea.Cmd
	x.clock, cmd = x.clock.Update(msg)
	if timeout, ok := msg.(timer.TimeoutMsg); ok && timeout.ID == x.clock.ID() {
		x.visible = false
	}
	return x, cmd
}

func (x Toast) Visible() bool {
	return x.visible
}

func (x Toast) View(t theme.Theme, width int) string {
	if !x.visible {
		return ""
	}
	return t.StateStyle(x.kind).MaxWidth(width).Render(x.text)
}
