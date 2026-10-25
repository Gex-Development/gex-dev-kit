package theme

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

type Palette struct {
	BrandBlue, BrandBlueDeep, BrandCyan, BrandCyanSoft color.Color
	BrandClaude, BrandCodex, BrandOMP                  color.Color
	Accent, Navy                                       color.Color
	Text, TextSecondary, TextFaint                     color.Color
	SurfaceActive, BorderCard                          color.Color
	Surface, SurfaceRaised, BorderStrong, Glow         color.Color
	Positive, Negative, Warning                        color.Color
	SurfaceDeep, Pink, Teal, Amber, Lilac              color.Color
	Gradient                                           []color.Color
}

func Dark() Palette {
	return Palette{
		BrandBlue:     lipgloss.Color("#2b7fff"),
		BrandBlueDeep: lipgloss.Color("#0b3d86"),
		BrandCyan:     lipgloss.Color("#22d3ee"),
		BrandCyanSoft: lipgloss.Color("#8bd5fa"),
		Accent:        lipgloss.Color("#5ec8ff"),
		Navy:          lipgloss.Color("#0b1526"),
		Text:          lipgloss.Color("#e6e9f2"),
		TextSecondary: lipgloss.Color("#9aa3b8"),
		TextFaint:     lipgloss.Color("#5b6480"),
		SurfaceActive: lipgloss.Color("#1b2c4a"),
		BorderCard:    lipgloss.Color("#3a4468"),
		Surface:       lipgloss.Color("#1a1d2b"),
		SurfaceRaised: lipgloss.Color("#242940"),
		BorderStrong:  lipgloss.Color("#4f7fd6"),
		Glow:          lipgloss.Color("#d8f3ff"),
		Positive:      lipgloss.Color("#4ade80"),
		Negative:      lipgloss.Color("#ff6b6b"),
		Warning:       lipgloss.Color("#ffd24a"),
		BrandClaude:   lipgloss.Color("#e89a6b"),
		BrandCodex:    lipgloss.Color("#70c9a5"),
		BrandOMP:      lipgloss.Color("#d1b0ed"),
		SurfaceDeep:   lipgloss.Color("#10111c"),
		Pink:          lipgloss.Color("#e060c8"),
		Teal:          lipgloss.Color("#4fd1c0"),
		Amber:         lipgloss.Color("#f2a65a"),
		Lilac:         lipgloss.Color("#9a8fc0"),
		Gradient: []color.Color{
			lipgloss.Color("#2b7fff"),
			lipgloss.Color("#22d3ee"),
			lipgloss.Color("#8bd5fa"),
		},
	}
}

func Light() Palette {
	return Palette{
		BrandBlue:     lipgloss.Color("#1e6fd9"),
		BrandBlueDeep: lipgloss.Color("#0b3d86"),
		BrandCyan:     lipgloss.Color("#0891b2"),
		BrandCyanSoft: lipgloss.Color("#7fc3ef"),
		Accent:        lipgloss.Color("#1a5fc0"),
		Navy:          lipgloss.Color("#0b1526"),
		Text:          lipgloss.Color("#0b1526"),
		TextSecondary: lipgloss.Color("#4b5669"),
		TextFaint:     lipgloss.Color("#8a95a8"),
		SurfaceActive: lipgloss.Color("#dbe8fb"),
		BorderCard:    lipgloss.Color("#b4bfd0"),
		Surface:       lipgloss.Color("#f3f6fb"),
		SurfaceRaised: lipgloss.Color("#e8eef8"),
		BorderStrong:  lipgloss.Color("#1e6fd9"),
		Glow:          lipgloss.Color("#ffffff"),
		Positive:      lipgloss.Color("#047857"),
		Negative:      lipgloss.Color("#b91c1c"),
		Warning:       lipgloss.Color("#92400e"),
		BrandClaude:   lipgloss.Color("#a94d28"),
		BrandCodex:    lipgloss.Color("#137252"),
		BrandOMP:      lipgloss.Color("#76509d"),
		SurfaceDeep:   lipgloss.Color("#dee3ed"),
		Pink:          lipgloss.Color("#a33286"),
		Teal:          lipgloss.Color("#087f74"),
		Amber:         lipgloss.Color("#985015"),
		Lilac:         lipgloss.Color("#665286"),
		Gradient: []color.Color{
			lipgloss.Color("#0b3d86"),
			lipgloss.Color("#1e6fd9"),
			lipgloss.Color("#0891b2"),
		},
	}
}

type State int

const (
	StatePending State = iota
	StateRunning
	StateOK
	StateWarn
	StateError
	StateSkipped
	StateInfo
)

var (
	tokensPlainSymbols = map[State]string{
		StatePending: "[ ]",
		StateRunning: ">",
		StateOK:      "[ok]",
		StateWarn:    "[!]",
		StateError:   "[x]",
		StateSkipped: "[-]",
		StateInfo:    "[i]",
	}
	tokensUnicodeSymbols = map[State]string{
		StatePending: "○",
		StateRunning: "▸",
		StateOK:      "✓",
		StateWarn:    "!",
		StateError:   "✗",
		StateSkipped: "–",
		StateInfo:    "i",
	}
	tokensStateLabels = map[State]string{
		StatePending: "pendente",
		StateRunning: "rodando",
		StateOK:      "ok",
		StateWarn:    "atenção",
		StateError:   "erro",
		StateSkipped: "ignorado",
		StateInfo:    "info",
	}
)

func (s State) Symbol(plain bool) string {
	if plain {
		return tokensPlainSymbols[s]
	}
	return tokensUnicodeSymbols[s]
}

func (s State) Label() string {
	return tokensStateLabels[s]
}
