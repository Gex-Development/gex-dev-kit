package screens

import (
	"errors"
	"os"
	"sync/atomic"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/term"
)

var ErrNotTerminal = errors.New("TUI exige terminal interativo; use --ui=cli")

type Runner struct {
	In      *os.File
	Out     *os.File
	Environ []string
	Profile colorprofile.Profile
	Signals <-chan os.Signal
}

type runnerQuitMsg struct{ Signal os.Signal }

type runnerModel struct {
	inner  tea.Model
	busy   atomic.Bool
	signal os.Signal
}

func (m *runnerModel) Init() tea.Cmd {
	return m.inner.Init()
}

func (m *runnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if quit, ok := msg.(runnerQuitMsg); ok {
		m.signal = quit.Signal
		return m, tea.Quit
	}
	next, cmd := m.inner.Update(msg)
	m.inner = next
	m.busy.Store(isBusy(next))
	return m, cmd
}

func (m *runnerModel) View() tea.View {
	return m.inner.View()
}

func isBusy(model tea.Model) bool {
	busy, ok := model.(Busy)
	return ok && busy.Busy()
}

func (r Runner) Run(m tea.Model) (tea.Model, error) {
	if !runnerIsTerminal(r.In) || !runnerIsTerminal(r.Out) {
		return m, ErrNotTerminal
	}
	return r.start(m)
}

func runnerIsTerminal(f *os.File) bool {
	return f != nil && term.IsTerminal(f.Fd())
}

func (r Runner) start(m tea.Model, extra ...tea.ProgramOption) (tea.Model, error) {
	wrapped := &runnerModel{inner: m}
	wrapped.busy.Store(isBusy(m))
	opts := append([]tea.ProgramOption{
		tea.WithInput(r.In),
		tea.WithOutput(r.Out),
		tea.WithColorProfile(r.Profile),
		tea.WithEnvironment(r.Environ),
		tea.WithoutSignalHandler(),
	}, extra...)
	p := tea.NewProgram(wrapped, opts...)
	done := make(chan struct{})
	defer close(done)
	if r.Signals != nil {
		go relayRunnerSignals(p, wrapped, r.Signals, done)
	}
	final, err := p.Run()
	result := m
	if w, ok := final.(*runnerModel); ok {
		result = w.inner
	}
	if err == nil && wrapped.signal != nil {
		return result, SignalError{Signal: wrapped.signal}
	}
	return result, err
}

func relayRunnerSignals(p *tea.Program, m *runnerModel, signals <-chan os.Signal, done <-chan struct{}) {
	for {
		select {
		case <-done:
			return
		case sig, ok := <-signals:
			if !ok {
				return
			}
			if m.busy.Load() {
				p.Send(InterruptDeferredMsg{Signal: sig})
				continue
			}
			p.Send(runnerQuitMsg{Signal: sig})
		}
	}
}
