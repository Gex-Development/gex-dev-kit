package components

import (
	"image/color"
	"math"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

const wordmarkBreakpoint = 60

var wordmarkFont = map[rune][]string{
	'G': {"####", "# ##", "####"},
	'E': {"####", "##  ", "####"},
	'X': {"#  #", " ## ", "#  #"},
}

func Wordmark(t theme.Theme, width int, reveal float64) string {
	if width <= 0 {
		return ""
	}
	reveal = wordmarkClamp(reveal)
	if width < wordmarkBreakpoint {
		return wordmarkCompact(t, width, reveal)
	}
	rows := wordmarkRows(t.Plain)
	lines := make([]string, 0, len(rows)+1)
	for _, row := range rows {
		lines = append(lines, wordmarkPaint(t, row, reveal))
	}
	label := wordmarkPad("DEV KIT", len([]rune(rows[0])))
	lines = append(lines, wordmarkPaint(t, label, reveal))
	return strings.Join(lines, "\n")
}

func wordmarkClamp(v float64) float64 {
	if !(v >= 0) {
		return 0
	}
	return min(v, 1)
}

func wordmarkCompact(t theme.Theme, width int, reveal float64) string {
	text := []rune(wordmarkTagline(t.Plain))
	if len(text) > width {
		text = text[:width]
	}
	return wordmarkPaint(t, string(text), reveal)
}

func wordmarkTagline(plain bool) string {
	if plain {
		return "GEX - DEV KIT"
	}
	return "GEX · DEV KIT"
}

func wordmarkRows(plain bool) []string {
	fill := "██"
	if plain {
		fill = "##"
	}
	rows := make([]string, 3)
	for r := range rows {
		var b strings.Builder
		for i, letter := range "GEX" {
			if i > 0 {
				b.WriteString("  ")
			}
			for _, cell := range wordmarkFont[letter][r] {
				if cell == '#' {
					b.WriteString(fill)
				} else {
					b.WriteString("  ")
				}
			}
		}
		rows[r] = b.String()
	}
	return rows
}

func wordmarkPad(s string, n int) string {
	if pad := n - len([]rune(s)); pad > 0 {
		return s + strings.Repeat(" ", pad)
	}
	return s
}

func wordmarkPaint(t theme.Theme, line string, reveal float64) string {
	cells := []rune(line)
	visible := int(math.Round(reveal * float64(len(cells))))
	colors := wordmarkColors(t, len(cells))
	var b strings.Builder
	for i, r := range cells {
		if i >= visible || r == ' ' {
			b.WriteRune(' ')
			continue
		}
		b.WriteString(wordmarkStyle(t, colors, i).Render(string(r)))
	}
	return b.String()
}

func wordmarkColors(t theme.Theme, n int) []color.Color {
	if !t.Rich {
		return nil
	}
	return t.Blend(n)
}

func wordmarkStyle(t theme.Theme, colors []color.Color, i int) lipgloss.Style {
	if i < len(colors) && colors[i] != nil {
		return t.Title.Foreground(colors[i])
	}
	return t.Title
}
