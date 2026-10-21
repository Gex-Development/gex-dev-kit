package components

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/stopwatch"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/motion"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

const (
	stepsTick       = 100 * time.Millisecond
	stepsClockCells = 8
)

type StepProgress struct {
	steps    []string
	states   []theme.State
	next     int
	started  bool
	animated bool
	bar      progress.Model
	clock    stopwatch.Model
}

func NewStepProgress(t theme.Theme, steps []string, animated bool) StepProgress {
	return StepProgress{
		steps:    slices.Clone(steps),
		states:   make([]theme.State, len(steps)),
		animated: animated,
		bar: progress.New(
			progress.WithColors(t.ProgressColors()...),
			progress.WithSpringOptions(motion.Gentle.Frequency, motion.Gentle.Damping),
			progress.WithoutPercentage(),
			stepsFillOption(t),
		),
		clock: stopwatch.New(stopwatch.WithInterval(stepsTick)),
	}
}

func (s StepProgress) WithTheme(t theme.Theme) StepProgress {
	progress.WithColors(t.ProgressColors()...)(&s.bar)
	stepsFillOption(t)(&s.bar)
	return s
}

func (s StepProgress) Start() (StepProgress, tea.Cmd) {
	if s.started {
		return s, nil
	}
	s.started = true
	if !s.animated {
		return s, nil
	}
	return s, s.clock.Start()
}

func (s StepProgress) Advance(state theme.State) (StepProgress, tea.Cmd) {
	if s.next >= len(s.steps) {
		return s, nil
	}
	s.states = slices.Clone(s.states)
	s.states[s.next] = state
	s.next++
	frame := s.bar.SetPercent(float64(s.next) / float64(len(s.steps)))
	if !s.animated {
		return s, nil
	}
	if s.next == len(s.steps) && s.started {
		return s, tea.Batch(frame, s.clock.Stop())
	}
	return s, frame
}

func (s StepProgress) Update(msg tea.Msg) (StepProgress, tea.Cmd) {
	switch msg.(type) {
	case progress.FrameMsg:
		var cmd tea.Cmd
		s.bar, cmd = s.bar.Update(msg)
		return s, cmd
	case stopwatch.StartStopMsg, stopwatch.TickMsg:
		var cmd tea.Cmd
		s.clock, cmd = s.clock.Update(msg)
		return s, cmd
	}
	return s, nil
}

func (s StepProgress) Elapsed() time.Duration {
	return s.clock.Elapsed()
}

func (s StepProgress) View(t theme.Theme, width int) string {
	clock := stepsPad(stepsClockText(s.clock.Elapsed()), stepsClockCells)
	bar := s.bar
	bar.SetWidth(max(1, width-stepsClockCells-1))
	head := bar.ViewAs(s.bar.Percent())
	if s.animated {
		head = bar.View()
	}
	lines := []string{head + " " + clock}
	for i, name := range s.steps {
		lines = append(lines, t.StateStyle(s.stateAt(i)).Render(name))
	}
	return t.Text.MaxWidth(width).Render(strings.Join(lines, "\n"))
}

func (s StepProgress) stateAt(i int) theme.State {
	switch {
	case i < s.next:
		return s.states[i]
	case i == s.next && s.started:
		return theme.StateRunning
	default:
		return theme.StatePending
	}
}

func stepsFillOption(t theme.Theme) progress.Option {
	if t.Plain {
		return progress.WithFillCharacters('#', '-')
	}
	return progress.WithFillCharacters(progress.DefaultFullCharHalfBlock, progress.DefaultEmptyCharBlock)
}

func stepsClockText(d time.Duration) string {
	return strings.Replace(fmt.Sprintf("%.1f s", d.Seconds()), ".", ",", 1)
}

func stepsPad(s string, cells int) string {
	return strings.Repeat(" ", max(0, cells-lipgloss.Width(s))) + s
}
