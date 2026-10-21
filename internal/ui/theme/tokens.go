package theme

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

type Palette struct {
	BrandBlue, BrandBlueDeep, BrandCyan, BrandCyanSoft color.Color
	Accent, Navy                                       color.Color
	Text, TextSecondary, TextFaint                     color.Color
	SurfaceActive, BorderCard                          color.Color
	Positive, Negative, Warning                        color.Color
	Gradient                                           []color.Color
}

func Dark() Palette {
	return Palette{
		BrandBlue:     lipgloss.Color("#1e6fd9"),
		BrandBlueDeep: lipgloss.Color("#0b3d86"),
		BrandCyan:     lipgloss.Color("#38bdf8"),
		BrandCyanSoft: lipgloss.Color("#7fc3ef"),
		Accent:        lipgloss.Color("#7fc3ef"),
		Navy:          lipgloss.Color("#0b1526"),
		Text:          lipgloss.Color("#f4f4f5"),
		TextSecondary: lipgloss.Color("#a1a1aa"),
		TextFaint:     lipgloss.Color("#3f3f46"),
		SurfaceActive: lipgloss.Color("#14202e"),
		BorderCard:    lipgloss.Color("#26272d"),
		Positive:      lipgloss.Color("#34d399"),
		Negative:      lipgloss.Color("#f87171"),
		Warning:       lipgloss.Color("#fcd34d"),
		Gradient: []color.Color{
			lipgloss.Color("#1e6fd9"),
			lipgloss.Color("#38bdf8"),
			lipgloss.Color("#7fc3ef"),
		},
	}
}

func Light() Palette {
	return Palette{
		BrandBlue:     lipgloss.Color("#1e6fd9"),
		BrandBlueDeep: lipgloss.Color("#0b3d86"),
		BrandCyan:     lipgloss.Color("#38bdf8"),
		BrandCyanSoft: lipgloss.Color("#7fc3ef"),
		Accent:        lipgloss.Color("#1a5fc0"),
		Navy:          lipgloss.Color("#0b1526"),
		Text:          lipgloss.Color("#0b1526"),
		TextSecondary: lipgloss.Color("#4b5669"),
		TextFaint:     lipgloss.Color("#b6bfcc"),
		SurfaceActive: lipgloss.Color("#e8f0fc"),
		BorderCard:    lipgloss.Color("#d0d7e1"),
		Positive:      lipgloss.Color("#047857"),
		Negative:      lipgloss.Color("#b91c1c"),
		Warning:       lipgloss.Color("#92400e"),
		Gradient: []color.Color{
			lipgloss.Color("#0b3d86"),
			lipgloss.Color("#1e6fd9"),
			lipgloss.Color("#38bdf8"),
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
