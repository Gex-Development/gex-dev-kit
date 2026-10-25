package layout

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
	"github.com/charmbracelet/x/ansi"
)

type Size struct{ Width, Height int }

type Class int

const (
	ClassTiny Class = iota
	ClassCompact
	ClassRegular
	ClassWide
)

const (
	MinWidth      = 80
	MinHeight     = 24
	CompactWidth  = 60
	CompactHeight = 20
	TinyWidth     = 40
	TinyHeight    = 12
)

const (
	layoutWideWidth     = 120
	layoutWideHeight    = 30
	layoutChromeRows    = 4
	layoutDefaultFooter = "? ajuda · q sai"
	layoutTooSmallText  = "Terminal pequeno: amplie para 60×20 ou use --ui=cli · q sai"
	layoutEllipsis      = "…"
)

func Classify(s Size) Class {
	switch {
	case s.Width >= layoutWideWidth && s.Height >= layoutWideHeight:
		return ClassWide
	case s.Width >= MinWidth && s.Height >= MinHeight:
		return ClassRegular
	case s.Width >= TinyWidth && s.Height >= TinyHeight:
		return ClassCompact
	default:
		return ClassTiny
	}
}

type Frame struct {
	Title   string
	Context string
	Version string
	Body    string
	Footer  string
	Notice  string
}

func Render(t theme.Theme, s Size, f Frame) string {
	if Classify(s) == ClassTiny {
		return TooSmall(t, s)
	}
	body := BodySize(s)
	parts := []string{
		layoutHeader(t, s.Width, f),
		layoutDivider(t, s.Width),
		lipgloss.Place(body.Width, body.Height, lipgloss.Left, lipgloss.Top, Fit(f.Body, body.Width, body.Height)),
		lipgloss.PlaceHorizontal(s.Width, lipgloss.Left, Fit(f.Notice, s.Width, 1)),
		layoutFooter(t, s.Width, f.Footer),
	}
	return Fit(strings.Join(parts, "\n"), s.Width, s.Height)
}

func Fit(content string, width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}
	lines := strings.Split(content, "\n")
	if len(lines) > height {
		lines = lines[:height]
	}
	for i, line := range lines {
		if ansi.StringWidth(line) > width {
			lines[i] = ansi.Truncate(line, width, layoutEllipsis)
		}
	}
	return strings.Join(lines, "\n")
}

func Columns(t theme.Theme, width int, left, right string) string {
	if width <= 0 {
		return ""
	}
	rightWidth := lipgloss.Width(right)
	if rightWidth >= width {
		return Fit(right, width, 1)
	}
	left = Fit(left, width-rightWidth-1, lipgloss.Height(left))
	gap := width - lipgloss.Width(left) - rightWidth
	return lipgloss.JoinHorizontal(lipgloss.Top, left, strings.Repeat(" ", gap), right)
}

func Center(s Size, content string) string {
	if s.Width <= 0 || s.Height <= 0 {
		return ""
	}
	return lipgloss.Place(s.Width, s.Height, lipgloss.Center, lipgloss.Center, Fit(content, s.Width, s.Height))
}

func TooSmall(t theme.Theme, s Size) string {
	if s.Width <= 0 || s.Height <= 0 {
		return ""
	}
	return Center(s, t.Warning.Render(ansi.Hardwrap(layoutTooSmallText, s.Width, true)))
}

func BodySize(s Size) Size {
	return Size{Width: max(s.Width, 0), Height: max(s.Height-layoutChromeRows, 0)}
}

func Header(t theme.Theme, width int, f Frame) string { return layoutHeader(t, width, f) }

func Divider(t theme.Theme, width int) string { return layoutDivider(t, width) }

func layoutHeader(t theme.Theme, width int, f Frame) string {
	mark := "◆ " + f.Title
	if t.Plain {
		mark = "[" + f.Title + "]"
	}
	left := t.Pill.Render(mark)
	if t.Plain {
		left = mark
	}
	if f.Context != "" {
		left += " " + t.Secondary.Render("· "+f.Context)
	}
	return Columns(t, width, left, t.Faint.Render(f.Version))
}

func layoutDivider(t theme.Theme, width int) string {
	width = max(width, 0)
	if t.Plain {
		return t.Divider.Render(strings.Repeat("-", width))
	}
	if !t.Rich || width == 0 {
		return t.Divider.Render(strings.Repeat("─", width))
	}
	colors := t.Blend(width)
	var b strings.Builder
	for i := range width {
		b.WriteString(t.Divider.Foreground(colors[i]).Render("━"))
	}
	return b.String()
}

func layoutFooter(t theme.Theme, width int, keys string) string {
	if keys == "" {
		keys = layoutDefaultFooter
	}
	return lipgloss.PlaceHorizontal(width, lipgloss.Left, Fit(t.Secondary.Render(keys), width, 1))
}
