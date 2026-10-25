package motion

import (
	"math"
	"math/rand"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/harmonica"
)

const FPS = 30

const (
	motionSettleTolerance = 0.001
	motionLifeFrames      = FPS * 8 / 10
)

type FrameMsg struct {
	ID   int
	Time time.Time
}

func Frame(id int) tea.Cmd {
	return tea.Tick(time.Second/FPS, func(t time.Time) tea.Msg {
		return FrameMsg{ID: id, Time: t}
	})
}

type Preset struct{ Frequency, Damping float64 }

var (
	Snappy = Preset{Frequency: 9, Damping: 1}
	Gentle = Preset{Frequency: 6, Damping: 0.85}
	Intro  = Preset{Frequency: 2.4, Damping: 1}
)

type Spring struct {
	enabled bool
	spring  harmonica.Spring
	pos     float64
	vel     float64
	target  float64
}

func NewSpring(enabled bool, p Preset) Spring {
	return Spring{
		enabled: enabled,
		spring:  harmonica.NewSpring(harmonica.FPS(FPS), p.Frequency, p.Damping),
	}
}

func (s Spring) Target(v float64) Spring {
	if !s.enabled {
		return s.Snap(v)
	}
	s.target = v
	return s
}

func (s Spring) Snap(v float64) Spring {
	s.pos, s.target, s.vel = v, v, 0
	return s
}

func (s Spring) Step() Spring {
	if s.Settled() {
		return s.Snap(s.target)
	}
	s.pos, s.vel = s.spring.Update(s.pos, s.vel, s.target)
	if s.Settled() {
		return s.Snap(s.target)
	}
	return s
}

func (s Spring) Value() float64 {
	return s.pos
}

func (s Spring) Settled() bool {
	if !s.enabled {
		return true
	}
	return math.Abs(s.pos-s.target) < motionSettleTolerance && math.Abs(s.vel) < motionSettleTolerance
}

type Particle struct{ X, Y int }

type burstParticle struct {
	pos harmonica.Point
	vel harmonica.Vector
}

type Burst struct {
	enabled bool
	parts   []burstParticle
	frames  int
}

func NewBurst(enabled bool, originX, originY float64, count int, seed int64) Burst {
	if !enabled {
		return Burst{}
	}
	rng := rand.New(rand.NewSource(seed))
	parts := make([]burstParticle, 0, max(count, 0))
	for i := 0; i < count; i++ {
		angle := rng.Float64() * 2 * math.Pi
		speed := 8 + rng.Float64()*12
		parts = append(parts, burstParticle{
			pos: harmonica.Point{X: originX, Y: originY},
			vel: harmonica.Vector{X: speed * math.Cos(angle), Y: speed * math.Sin(angle)},
		})
	}
	return Burst{enabled: true, parts: parts}
}

func (b Burst) Step() Burst {
	if !b.Alive() {
		return b
	}
	next := make([]burstParticle, len(b.parts))
	for i, p := range b.parts {
		proj := harmonica.NewProjectile(harmonica.FPS(FPS), p.pos, p.vel, harmonica.TerminalGravity)
		proj.Update()
		next[i] = burstParticle{pos: proj.Position(), vel: proj.Velocity()}
	}
	return Burst{enabled: true, parts: next, frames: b.frames + 1}
}

func (b Burst) Alive() bool {
	return b.enabled && len(b.parts) > 0 && b.frames < motionLifeFrames
}

func (b Burst) Particles(width, height int) []Particle {
	if !b.Alive() {
		return nil
	}
	var visible []Particle
	for _, p := range b.parts {
		x, y := int(math.Round(p.pos.X)), int(math.Round(p.pos.Y))
		if x < 0 || y < 0 || x >= width || y >= height {
			continue
		}
		visible = append(visible, Particle{X: x, Y: y})
	}
	return visible
}
