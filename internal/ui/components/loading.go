package components

import (
	"strings"
	"time"

	"charm.land/bubbles/v2/stopwatch"
	tea "charm.land/bubbletea/v2"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

const (
	loadingTick  = 100 * time.Millisecond
	loadingShine = 6
	loadingStep  = 2
)

type Loading struct {
	animated bool
	clock    stopwatch.Model
}

func NewLoading(animated bool) Loading {
	return Loading{
		animated: animated,
		clock:    stopwatch.New(stopwatch.WithInterval(loadingTick)),
	}
}

func (l Loading) Tick() tea.Cmd {
	if !l.animated {
		return nil
	}
	return l.clock.Start()
}

func (l Loading) Update(msg tea.Msg) (Loading, tea.Cmd) {
	switch msg.(type) {
	case stopwatch.StartStopMsg, stopwatch.TickMsg:
		if !l.animated {
			return l, nil
		}
		var cmd tea.Cmd
		l.clock, cmd = l.clock.Update(msg)
		return l, cmd
	}
	return l, nil
}

func (l Loading) View(t theme.Theme, width, lines int) string {
	if width <= 0 || lines <= 0 {
		return ""
	}
	shift := int(l.clock.Elapsed()/loadingTick) * loadingStep
	rows := make([]string, lines)
	for i := range rows {
		rows[i] = loadingRow(t, max(1, width*(lines-i)/lines), shift, l.animated)
	}
	return strings.Join(rows, "\n")
}

func loadingRow(t theme.Theme, width, shift int, animated bool) string {
	glyph := loadingGlyph(t)
	if !animated {
		return t.Faint.Render(strings.Repeat(glyph, width))
	}
	shine := min(loadingShine, width)
	start := shift%(width+shine) - shine
	colors := t.Blend(shine)
	var b strings.Builder
	for c := range width {
		style := t.Faint
		if k := c - start; k >= 0 && k < shine {
			style = style.Foreground(colors[k])
		}
		b.WriteString(style.Render(glyph))
	}
	return b.String()
}

func loadingGlyph(t theme.Theme) string {
	if t.Plain {
		return "#"
	}
	return "█"
}
