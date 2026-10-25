package screens

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
)

type runnerTestModel struct {
	busy     bool
	received os.Signal
}

func (m *runnerTestModel) Init() tea.Cmd { return nil }

func (m *runnerTestModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if deferred, ok := msg.(InterruptDeferredMsg); ok {
		m.received = deferred.Signal
		m.busy = false
		return m, tea.Quit
	}
	return m, nil
}

func (m *runnerTestModel) View() tea.View { return tea.NewView("") }

func (m *runnerTestModel) Busy() bool { return m.busy }

func runnerTestIO(t *testing.T) (*os.File, *os.File) {
	t.Helper()
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		read.Close()
		write.Close()
	})
	out, err := os.Create(filepath.Join(t.TempDir(), "saida"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { out.Close() })
	return read, out
}

func runnerForTest(r Runner, m tea.Model) (tea.Model, error) {
	return r.start(m, tea.WithoutRenderer())
}

func runnerTestRunner(t *testing.T, signals chan os.Signal) Runner {
	t.Helper()
	in, out := runnerTestIO(t)
	return Runner{In: in, Out: out, Profile: colorprofile.NoTTY, Signals: signals}
}

func runnerWithin(t *testing.T, fn func() (tea.Model, error)) (tea.Model, error) {
	t.Helper()
	type result struct {
		m   tea.Model
		err error
	}
	done := make(chan result, 1)
	go func() {
		m, err := fn()
		done <- result{m, err}
	}()
	select {
	case r := <-done:
		return r.m, r.err
	case <-time.After(5 * time.Second):
		t.Fatal("programa não terminou no prazo")
		return nil, nil
	}
}

func TestRunnerRejectsNonTerminal(t *testing.T) {
	in, out := runnerTestIO(t)
	_, err := Runner{In: in, Out: out}.Run(&runnerTestModel{})
	if !errors.Is(err, ErrNotTerminal) {
		t.Fatalf("pipe deveria ser recusado: %v", err)
	}
	if _, err := (Runner{}).Run(&runnerTestModel{}); !errors.Is(err, ErrNotTerminal) {
		t.Fatalf("In/Out nulos deveriam ser recusados: %v", err)
	}
}

func TestRunnerBusyDefersInterrupt(t *testing.T) {
	signals := make(chan os.Signal, 1)
	model := &runnerTestModel{busy: true}
	r := runnerTestRunner(t, signals)
	signals <- os.Interrupt
	final, err := runnerWithin(t, func() (tea.Model, error) { return runnerForTest(r, model) })
	if err != nil {
		t.Fatalf("modelo ocupado não deveria devolver erro de sinal: %v", err)
	}
	if final.(*runnerTestModel).received != os.Interrupt {
		t.Fatalf("sinal adiado não chegou ao modelo: %v", final.(*runnerTestModel).received)
	}
}

func TestRunnerIdleSignalExitsWithCode(t *testing.T) {
	cases := []struct {
		signal os.Signal
		code   int
	}{
		{os.Interrupt, 130},
		{syscall.SIGTERM, 143},
		{syscall.SIGHUP, 129},
	}
	for _, tc := range cases {
		signals := make(chan os.Signal, 1)
		r := runnerTestRunner(t, signals)
		signals <- tc.signal
		_, err := runnerWithin(t, func() (tea.Model, error) { return runnerForTest(r, &runnerTestModel{}) })
		code, ok := ExitCode(err)
		if !ok || code != tc.code {
			t.Fatalf("sinal %v: código=%d ok=%v err=%v; quer %d", tc.signal, code, ok, err, tc.code)
		}
	}
}

func TestSignalErrorMessage(t *testing.T) {
	if msg := (SignalError{Signal: os.Interrupt}).Error(); msg == "" {
		t.Fatal("mensagem de sinal vazia")
	}
}
