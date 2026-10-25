package motion

import (
	"bytes"
	"reflect"
	"strconv"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/exp/teatest/v2"
)

type motionDoneMsg struct{}

const motionQuietWindow = 10 * time.Second / FPS

type motionLoopModel struct {
	spring  Spring
	frames  int
	extra   int
	settled bool
}

func (m motionLoopModel) Init() tea.Cmd {
	return Frame(1)
}

func (m motionLoopModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case FrameMsg:
		if m.settled {
			m.extra++
			return m, nil
		}
		m.frames++
		m.spring = m.spring.Step()
		if m.spring.Settled() {
			m.settled = true
			return m, tea.Tick(motionQuietWindow, func(time.Time) tea.Msg { return motionDoneMsg{} })
		}
		return m, Frame(msg.ID + 1)
	case motionDoneMsg:
		return m, tea.Quit
	}
	return m, nil
}

func (m motionLoopModel) View() tea.View {
	if m.settled {
		return tea.NewView("assentou em " + strconv.Itoa(m.frames) + " quadros")
	}
	return tea.NewView("movendo")
}

func TestFrameDeliversID(t *testing.T) {
	msg, ok := Frame(7)().(FrameMsg)
	if !ok || msg.ID != 7 {
		t.Fatalf("Frame(7) entregou %#v; quero FrameMsg com ID 7", msg)
	}
}

func TestSpringDisabledJumps(t *testing.T) {
	s := NewSpring(false, Gentle).Target(42)
	if s.Value() != 42 {
		t.Fatalf("mola desligada não saltou ao alvo: valor %v", s.Value())
	}
	if !s.Settled() {
		t.Fatalf("mola desligada não está assentada")
	}
	if got := s.Step().Value(); got != 42 {
		t.Fatalf("Step da mola desligada mudou o valor para %v", got)
	}
	if NewSpring(true, Snappy).Snap(5).Value() != 5 || !NewSpring(true, Snappy).Snap(5).Settled() {
		t.Fatalf("Snap deveria saltar e assentar a mola")
	}
}

func TestSpringEnabledMovesThenSnapsToTarget(t *testing.T) {
	s := NewSpring(true, Gentle).Target(10)
	if s.Settled() {
		t.Fatalf("mola ligada já assentada logo após o alvo")
	}
	if first := s.Step().Value(); first <= 0 || first >= 10 {
		t.Fatalf("primeiro passo deveria andar até o alvo, valor %v", first)
	}
	for !s.Settled() {
		s = s.Step()
	}
	if s.Value() != 10 {
		t.Fatalf("mola assentou em %v; quero 10 exatos", s.Value())
	}
}

func TestGentleSettlesWithin300Steps(t *testing.T) {
	s := NewSpring(true, Gentle).Target(100)
	passos := 0
	for !s.Settled() {
		if passos == 300 {
			t.Fatalf("mola Gentle não assentou em 300 passos; valor %v", s.Value())
		}
		s = s.Step()
		passos++
	}
}

func TestBurstDisabledNeverAlive(t *testing.T) {
	b := NewBurst(false, 1, 1, 10, 1)
	if b.Alive() {
		t.Fatalf("explosão desligada está viva")
	}
	if got := b.Particles(80, 24); len(got) != 0 {
		t.Fatalf("explosão desligada tem %d partículas", len(got))
	}
	if b.Step().Alive() {
		t.Fatalf("Step de explosão desligada ficou viva")
	}
}

func TestBurstLifeIsEightTenthsOfSecond(t *testing.T) {
	b := NewBurst(true, 10, 5, 6, 3)
	passos := 0
	for b.Alive() {
		b = b.Step()
		passos++
		if passos > 100 {
			t.Fatalf("explosão continua viva após 100 quadros")
		}
	}
	if want := int(0.8 * FPS); passos != want {
		t.Fatalf("explosão viveu %d quadros; quero %d (0,8 s)", passos, want)
	}
}

func TestBurstSameSeedSameParticles(t *testing.T) {
	a := NewBurst(true, 40, 12, 8, 99).Step().Step().Step()
	b := NewBurst(true, 40, 12, 8, 99).Step().Step().Step()
	if !reflect.DeepEqual(a.Particles(80, 24), b.Particles(80, 24)) {
		t.Fatalf("mesma seed deu partículas diferentes")
	}
	c := NewBurst(true, 40, 12, 8, 100).Step().Step().Step()
	if reflect.DeepEqual(a.Particles(80, 24), c.Particles(80, 24)) {
		t.Fatalf("seeds diferentes deram partículas iguais")
	}
}

func TestBurstOldCopyUnchanged(t *testing.T) {
	antiga := NewBurst(true, 20, 10, 5, 7)
	antes := append([]burstParticle(nil), antiga.parts...)
	nova := antiga.Step().Step().Step()
	if !reflect.DeepEqual(antiga.parts, antes) {
		t.Fatalf("cópia antiga mudou quando a nova avançou")
	}
	if reflect.DeepEqual(nova.parts, antes) {
		t.Fatalf("a cópia nova não avançou")
	}
}

func TestBurstParticlesOnlyInsideScreen(t *testing.T) {
	if got := NewBurst(true, 5, 5, 4, 3).Particles(0, 0); len(got) != 0 {
		t.Fatalf("partículas fora da tela: %v", got)
	}
}

func TestFramesStopAfterSpringSettles(t *testing.T) {
	m := motionLoopModel{spring: NewSpring(true, Gentle).Target(1)}
	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(80, 24),
		teatest.WithProgramOptions(
			tea.WithColorProfile(colorprofile.ASCII),
			tea.WithEnvironment([]string{"TERM=xterm-256color"}),
		),
	)
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("assentou"))
	}, teatest.WithDuration(5*time.Second))
	tm.WaitFinished(t, teatest.WithFinalTimeout(5*time.Second))
	final, ok := tm.FinalModel(t, teatest.WithFinalTimeout(5*time.Second)).(motionLoopModel)
	if !ok {
		t.Fatalf("modelo final de tipo inesperado")
	}
	if !final.spring.Settled() || final.spring.Value() != 1 {
		t.Fatalf("mola não assentou no alvo: valor %v", final.spring.Value())
	}
	if final.frames == 0 || final.frames > 300 {
		t.Fatalf("quadros usados = %d; quero entre 1 e 300", final.frames)
	}
	if final.extra != 0 {
		t.Fatalf("%d quadros chegaram depois de assentar", final.extra)
	}
}
