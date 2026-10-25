package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
	"github.com/charmbracelet/x/ansi"
)

const toolblockCollapsed = 6

type ToolRow struct {
	State  theme.State
	Label  string
	Detail string
}

type ToolBlock struct {
	Title, Params string
	State         theme.State
	Rows          []ToolRow
	Lines         []string
	Expanded      bool
	Elapsed       time.Duration
	Spinner       string
	Phases        []string
	PhaseRows     bool
}

func FormatElapsed(d time.Duration) string {
	seconds := int(d.Seconds())
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	if seconds < 3600 {
		return fmt.Sprintf("%dm %ds", seconds/60, seconds%60)
	}
	return fmt.Sprintf("%dh %dm", seconds/3600, (seconds%3600)/60)
}

func (b ToolBlock) View(t theme.Theme, width int) string {
	if width <= 0 {
		return ""
	}
	icon, iconStyle := "●", t.Faint
	switch b.State {
	case theme.StateRunning:
		icon, iconStyle = b.Spinner, t.Accent
		if icon == "" {
			icon = "●"
		}
	case theme.StateOK:
		icon, iconStyle = "✓", t.Success
	case theme.StateWarn:
		icon, iconStyle = "!", t.Warning
	case theme.StateError:
		icon, iconStyle = "×", t.Error
	}
	if t.Plain {
		switch b.State {
		case theme.StateRunning:
			icon = "*"
		case theme.StateOK:
			icon = "v"
		case theme.StateWarn:
			icon = "!"
		case theme.StateError:
			icon = "x"
		case theme.StateSkipped:
			icon = "-"
		default:
			icon = "o"
		}
	}
	ellipsis, horizontal, vertical, topLeft, topRight, bottomLeft, bottomRight := "…", "─", "│", "╭", "╮", "╰", "╯"
	if t.Plain {
		ellipsis, horizontal, vertical, topLeft, topRight, bottomLeft, bottomRight = "...", "-", "|", "+", "+", "+", "+"
	}
	title, params := b.Title, b.Params
	if t.Plain {
		title, params = plainASCII(title), plainASCII(params)
	}
	meta := ""
	if b.State == theme.StateSkipped {
		meta = "cancelado"
	} else if b.Elapsed >= time.Second {
		meta = FormatElapsed(b.Elapsed)
	}
	cardWidth := min(width, 100)
	framed := width >= 20
	contentWidth := width
	if framed {
		contentWidth = cardWidth - 4
	}
	border := t.Faint
	switch b.State {
	case theme.StateRunning:
		border = t.Accent
	case theme.StateError:
		border = t.Error
	}
	headText := iconStyle.Render(icon) + " " + t.Accent.Bold(true).Render(title)
	if params != "" {
		headText += " " + t.Secondary.Render(params)
	}
	if !framed {
		out := []string{ansi.Truncate(headText, width, ellipsis)}
		for _, line := range b.body(t, contentWidth, ellipsis) {
			out = append(out, ansi.Truncate("  "+line, width, ellipsis))
		}
		return strings.Join(out, "\n")
	}
	headFullWidth := ansi.StringWidth(headText)
	headText = ansi.Truncate(headText, cardWidth-6, ellipsis)
	top := border.Render(topLeft+horizontal+" ") + headText + border.Render(" "+strings.Repeat(horizontal, max(0, cardWidth-ansi.StringWidth(headText)-5))+topRight)
	out := []string{top}
	for _, line := range b.body(t, contentWidth, ellipsis) {
		line = ansi.Truncate(line, contentWidth, ellipsis)
		out = append(out, border.Render(vertical+" ")+line+strings.Repeat(" ", max(0, contentWidth-ansi.StringWidth(line)))+border.Render(" "+vertical))
	}
	if meta != "" && (headFullWidth+5 > cardWidth || ansi.StringWidth(meta)+7 > cardWidth) {
		meta = ""
	}
	footer := bottomLeft + strings.Repeat(horizontal, cardWidth-2)
	if meta != "" {
		footer = bottomLeft + strings.Repeat(horizontal, cardWidth-ansi.StringWidth(meta)-5) + " " + meta + " " + horizontal
	}
	if meta == "" {
		out = append(out, border.Render(footer+bottomRight))
	} else {
		prefix := bottomLeft + strings.Repeat(horizontal, cardWidth-ansi.StringWidth(meta)-5)
		out = append(out, border.Render(prefix+" ")+t.Faint.Render(meta)+border.Render(" "+horizontal+bottomRight))
	}
	return strings.Join(out, "\n")
}

func (b ToolBlock) body(t theme.Theme, width int, ellipsis string) []string {
	count := len(b.Rows) + len(b.Lines)
	if len(b.Phases) > 0 {
		count = len(b.Phases)
	}
	visible := count
	if !b.Expanded && count > toolblockCollapsed+1 {
		visible = toolblockCollapsed
	}
	labelWidth := 0
	for _, row := range b.Rows {
		label := row.Label
		if t.Plain {
			label = plainASCII(label)
		}
		labelWidth = max(labelWidth, ansi.StringWidth(row.State.Symbol(t.Plain)+" "+label))
	}
	var output []string
	for i := range visible {
		if len(b.Phases) > 0 && (!b.PhaseRows || i >= len(b.Rows)) {
			label := strings.TrimSuffix(b.Phases[i], "…")
			state := theme.StatePending
			if time.Duration(i+1)*time.Second <= b.Elapsed {
				state = theme.StateOK
			} else if time.Duration(i)*time.Second <= b.Elapsed {
				state = theme.StateRunning
				label = b.Phases[i]
			}
			if t.Plain {
				label = plainASCII(label)
			}
			icon := state.Symbol(t.Plain)
			style := t.Faint
			if state == theme.StateRunning {
				icon, style = b.Spinner, t.Accent
				if t.Plain || icon == "" {
					icon = "*"
				}
			}
			output = append(output, ansi.Truncate(style.Render(icon+" "+label), width, ellipsis))
			continue
		}
		if i < len(b.Rows) {
			row := b.Rows[i]
			label, detail := row.Label, row.Detail
			if t.Plain {
				label, detail = plainASCII(label), plainASCII(detail)
			}
			style := t.StateStyle(row.State)
			if row.State == theme.StateInfo || row.State == theme.StatePending {
				style = style.Foreground(t.Palette.TextFaint)
			}
			rendered := style.Render(label)
			rendered += strings.Repeat(" ", max(0, labelWidth-ansi.StringWidth(rendered)))
			output = append(output, ansi.Truncate(rendered+"  "+t.Secondary.Render(detail), width, ellipsis))
			continue
		}
		text := b.Lines[i-len(b.Rows)]
		if t.Plain {
			text = plainASCII(text)
		}
		output = append(output, ansi.Truncate(t.Secondary.Render(text), width, ellipsis))
	}
	if visible < count {
		notice := fmt.Sprintf("… (%d linhas ocultas) [clique ou ctrl+o para expandir]", count-visible)
		if t.Plain {
			notice = plainASCII(notice)
		}
		output = append(output, t.Faint.Render(ansi.Truncate(notice, width, ellipsis)))
	}
	return output
}
