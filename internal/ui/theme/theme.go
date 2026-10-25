package theme

import (
	"image/color"
	"slices"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
)

type Options struct {
	Dark        bool
	Profile     colorprofile.Profile
	Plain       bool
	NoPowerline bool
}

type Theme struct {
	Dark      bool
	Rich      bool
	Plain     bool
	Powerline bool
	Palette   Palette

	Text, Secondary, Faint, Strong, Accent                                   lipgloss.Style
	Title, Eyebrow, Code                                                     lipgloss.Style
	Success, Warning, Error, Info                                            lipgloss.Style
	Selected, Unselected                                                     lipgloss.Style
	Panel, PanelFocused, Card, Modal                                         lipgloss.Style
	KeyHintKey, KeyHintDesc, Divider                                         lipgloss.Style
	Pill, Glow, Keycap, Bar, BarFocus                                        lipgloss.Style
	UserBar, Badge                                                           lipgloss.Style
	StatusBase, StatusPink, StatusTeal, StatusAmber, StatusLilac, PromptEdge lipgloss.Style
}

func New(o Options) Theme {
	p := Light()
	if o.Dark {
		p = Dark()
	}
	text := lipgloss.NewStyle().Foreground(p.Text)
	secondary := lipgloss.NewStyle().Foreground(p.TextSecondary)
	accent := lipgloss.NewStyle().Foreground(p.Accent)
	border := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(p.BorderCard)
	glow := p.Glow
	if !o.Dark {
		glow = p.Navy
	}
	return Theme{
		Dark:         o.Dark,
		Rich:         o.Profile >= colorprofile.ANSI256,
		Plain:        o.Plain,
		Powerline:    o.Profile >= colorprofile.ANSI256 && !o.NoPowerline && !o.Plain,
		Palette:      p,
		Text:         text,
		Secondary:    secondary,
		Faint:        lipgloss.NewStyle().Foreground(p.TextFaint),
		Strong:       text.Bold(true),
		Accent:       accent,
		Title:        accent.Bold(true),
		Eyebrow:      secondary,
		Code:         text.Background(p.SurfaceActive),
		Success:      lipgloss.NewStyle().Foreground(p.Positive).Bold(true),
		Warning:      lipgloss.NewStyle().Foreground(p.Warning).Bold(true),
		Error:        lipgloss.NewStyle().Foreground(p.Negative).Bold(true),
		Info:         accent,
		Selected:     lipgloss.NewStyle().Foreground(glow).Background(p.SurfaceActive).Bold(true),
		Unselected:   text,
		Panel:        border.Padding(0, 1),
		PanelFocused: border.BorderForeground(p.Accent).Padding(0, 1),
		Card:         border.Padding(0, 2),
		Modal:        border.BorderForeground(p.Accent).Padding(1, 2),
		KeyHintKey:   accent.Bold(true),
		KeyHintDesc:  secondary,
		Divider:      lipgloss.NewStyle().Foreground(p.BorderCard),
		Pill:         lipgloss.NewStyle().Foreground(p.Navy).Background(p.BrandCyan).Bold(true).Padding(0, 1),
		Glow:         lipgloss.NewStyle().Foreground(glow).Bold(true),
		Keycap:       lipgloss.NewStyle().Foreground(p.Accent).Background(p.SurfaceRaised).Bold(true).Padding(0, 1),
		Bar:          lipgloss.NewStyle().Foreground(p.Text).Background(p.SurfaceRaised),
		BarFocus:     lipgloss.NewStyle().Foreground(glow).Background(p.SurfaceActive).Bold(true),
		UserBar:      accent,
		Badge:        lipgloss.NewStyle().Foreground(p.Navy).Background(p.Accent).Bold(true).Padding(0, 1),
		StatusBase:   lipgloss.NewStyle().Background(p.SurfaceDeep),
		StatusPink:   lipgloss.NewStyle().Foreground(p.Pink).Background(p.SurfaceDeep).Bold(true),
		StatusTeal:   lipgloss.NewStyle().Foreground(p.Teal).Background(p.SurfaceDeep),
		StatusAmber:  lipgloss.NewStyle().Foreground(p.Amber).Background(p.SurfaceDeep),
		StatusLilac:  lipgloss.NewStyle().Foreground(p.Lilac),
		PromptEdge:   lipgloss.NewStyle().Foreground(p.Pink),
	}
}

func (t Theme) StateStyle(s State) lipgloss.Style {
	return themeStateBase(t, s).Transform(func(text string) string {
		if text == "" {
			text = s.Label()
		}
		return s.Symbol(t.Plain) + " " + text
	})
}

func (t Theme) Blend(n int) []color.Color {
	return lipgloss.Blend1D(n, slices.Clone(t.Palette.Gradient)...)
}

func (t Theme) ProgressColors() []color.Color {
	if !t.Rich {
		return []color.Color{t.Palette.Accent}
	}
	return slices.Clone(t.Palette.Gradient)
}

func (t Theme) Spinner() lipgloss.Style {
	return t.Accent
}

func themeStateBase(t Theme, s State) lipgloss.Style {
	switch s {
	case StateRunning:
		return t.Accent
	case StateOK:
		return t.Success
	case StateWarn:
		return t.Warning
	case StateError:
		return t.Error
	case StateInfo:
		return t.Info
	default:
		return t.Secondary
	}
}
