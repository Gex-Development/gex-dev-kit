package components

import (
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
	"github.com/charmbracelet/x/ansi"
)

type StatusTone uint8

const (
	ToneBrand StatusTone = iota
	ToneHarness
	ToneDir
	ToneGit
	ToneWarning
	ToneMuted
	ToneFaint
	ToneSuccess
)

type StatusSegment struct {
	Icon, Text          string
	Tone                StatusTone
	Brand               string
	Modified, Untracked int
}

type StatusMetric struct {
	Text string
	Tone StatusTone
}

func StatusLine(t theme.Theme, width int, segments []StatusSegment, metrics []StatusMetric) string {
	if width <= 0 {
		return ""
	}
	visible := append([]StatusSegment(nil), segments...)
	for i := range visible {
		if visible[i].Tone == ToneDir {
			visible[i].Text = statusTruncateLeft(visible[i].Text, 24, t.Plain)
		}
	}
	remaining := metrics
	rule := "─"
	if t.Plain {
		rule = "-"
	}
	for {
		bar := statusBar(t, visible)
		ruler := statusRuler(t, remaining)
		gap := width - ansi.StringWidth(bar) - ansi.StringWidth(ruler)
		if gap >= 0 {
			return bar + t.Faint.Render(strings.Repeat(rule, gap)) + ruler
		}
		if len(remaining) > 0 {
			remaining = remaining[1:]
			continue
		}
		if len(visible) > 1 {
			last := visible[len(visible)-1]
			if last.Tone == ToneDir {
				base := statusBar(t, append(append([]StatusSegment(nil), visible[:len(visible)-1]...), StatusSegment{Icon: last.Icon, Tone: last.Tone}))
				available := width - ansi.StringWidth(base) - ansi.StringWidth(ruler)
				if available > 1 {
					last.Text = statusTruncateLeft(last.Text, available, t.Plain)
					candidate := append(append([]StatusSegment(nil), visible[:len(visible)-1]...), last)
					if ansi.StringWidth(statusBar(t, candidate))+ansi.StringWidth(ruler) <= width {
						visible = candidate
						continue
					}
				}
			}
			visible = visible[:len(visible)-1]
			continue
		}
		if len(visible) == 1 {
			visible = nil
			continue
		}
		return t.Faint.Render(strings.Repeat(rule, width))
	}
}

func statusTruncateLeft(text string, width int, plain bool) string {
	if width <= 0 {
		return ""
	}
	if ansi.StringWidth(text) <= width {
		return text
	}
	ellipsis := "…"
	if plain {
		ellipsis = "..."
	}
	if width <= ansi.StringWidth(ellipsis) {
		return ansi.Truncate(ellipsis, width, "")
	}
	budget := width - ansi.StringWidth(ellipsis)
	runes := []rune(text)
	i := len(runes)
	for i > 0 && ansi.StringWidth(string(runes[i-1])) <= budget {
		i--
		budget -= ansi.StringWidth(string(runes[i]))
	}
	return ellipsis + string(runes[i:])
}

func statusStyle(t theme.Theme, tone StatusTone, brand string) lipgloss.Style {
	switch tone {
	case ToneBrand:
		return t.Secondary.Background(t.Palette.SurfaceDeep)
	case ToneHarness:
		switch brand {
		case "claude":
			return t.StatusPink.Foreground(t.Palette.BrandClaude)
		case "codex":
			return t.StatusPink.Foreground(t.Palette.BrandCodex)
		case "omp":
			return t.StatusPink.Foreground(t.Palette.BrandOMP)
		}
		return t.StatusPink
	case ToneDir, ToneSuccess:
		return t.StatusTeal
	case ToneGit, ToneWarning:
		return t.StatusAmber
	case ToneMuted:
		return t.Secondary.Background(t.Palette.SurfaceDeep)
	case ToneFaint:
		return t.StatusLilac.Background(t.Palette.SurfaceDeep)
	default:
		return t.StatusPink
	}
}

func statusBar(t theme.Theme, segments []StatusSegment) string {
	if len(segments) == 0 {
		return ""
	}
	var out strings.Builder
	for i, s := range segments {
		if t.Plain {
			text := s.Text
			if s.Tone == ToneBrand {
				text = "GEX"
			}
			if s.Modified > 0 {
				text += " *" + strconv.Itoa(s.Modified)
			}
			if s.Untracked > 0 {
				text += " ?" + strconv.Itoa(s.Untracked)
			}
			out.WriteString(" [" + strings.Map(statusASCII, text) + "]")
			continue
		}
		if i > 0 {
			sep := "\ue0b1"
			if !t.Powerline {
				sep = "›"
			}
			out.WriteString(t.Faint.Background(t.Palette.SurfaceDeep).Render(sep))
		}
		icon := s.Icon
		if !t.Powerline && icon == "\ue0a0" {
			icon = "⎇"
		}
		if !t.Powerline {
			switch s.Tone {
			case ToneDir:
				icon = "▸"
			case ToneGit:
				icon = "⎇"
			case ToneWarning:
				icon = "⌂"
			}
		}
		style := statusStyle(t, s.Tone, s.Brand)
		if icon != "" {
			out.WriteString(style.Render(" " + icon + " " + s.Text + " "))
		} else {
			out.WriteString(style.Render(" " + s.Text + " "))
		}
		if s.Modified > 0 {
			out.WriteString(t.StatusAmber.Render("*" + strconv.Itoa(s.Modified) + " "))
		}
		if s.Untracked > 0 {
			out.WriteString(t.StatusTeal.Render("?" + strconv.Itoa(s.Untracked) + " "))
		}
	}
	if t.Powerline && !t.Plain {
		out.WriteString(t.StatusBase.Foreground(t.Palette.SurfaceDeep).Background(t.Palette.Surface).Render("\ue0b0"))
	}
	return out.String()
}

func statusRuler(t theme.Theme, metrics []StatusMetric) string {
	var out strings.Builder
	rule := "─"
	if t.Plain {
		rule = "-"
	}
	for _, m := range metrics {
		out.WriteString(t.Faint.Render(" " + rule + rule + " "))
		text := m.Text
		if t.Plain {
			text = strings.Map(statusASCII, text)
		}
		style := t.StatusLilac
		switch m.Tone {
		case ToneBrand, ToneHarness:
			style = t.PromptEdge.Bold(true)
		case ToneDir, ToneSuccess:
			style = t.Faint.Foreground(t.Palette.Teal)
		case ToneGit, ToneWarning:
			style = t.Faint.Foreground(t.Palette.Amber)
		}
		out.WriteString(style.Render(text))
	}
	out.WriteString(t.Faint.Render(rule))
	return out.String()
}

func statusASCII(r rune) rune {
	if r < 32 || r > 126 {
		return '?'
	}
	return r
}
