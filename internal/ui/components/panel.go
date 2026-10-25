package components

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

type BorderKind int

const (
	BorderRounded BorderKind = iota
	BorderNormal
	BorderThick
	BorderDouble
	BorderBlock
	BorderASCII
)

type PanelOptions struct {
	Title   string
	Width   int
	Border  BorderKind
	Focused bool
	Glow    bool
}

type CardOptions struct {
	Title, Meta, Footer string
	Width               int
	Selected            bool
}

func Panel(t theme.Theme, o PanelOptions, body string) string {
	content := body
	if o.Title != "" {
		content = panelTitleLine(t, o.Title, o.Focused) + "\n" + body
	}
	return panelFit(panelFrame(t, o), o.Width).Render(content)
}

func Card(t theme.Theme, o CardOptions, body string) string {
	parts := make([]string, 0, 4)
	if o.Title != "" {
		parts = append(parts, cardTitleLine(t, o.Title, o.Selected))
	}
	if o.Meta != "" {
		parts = append(parts, t.Secondary.Render(o.Meta))
	}
	if body != "" {
		parts = append(parts, body)
	}
	if o.Footer != "" {
		parts = append(parts, t.Faint.Render(o.Footer))
	}
	return panelFit(cardFrame(t, o.Selected), o.Width).Render(strings.Join(parts, "\n"))
}

func panelBorder(t theme.Theme, kind BorderKind) lipgloss.Border {
	if t.Plain {
		return lipgloss.ASCIIBorder()
	}
	switch kind {
	case BorderNormal:
		return lipgloss.NormalBorder()
	case BorderThick:
		return lipgloss.ThickBorder()
	case BorderDouble:
		return lipgloss.DoubleBorder()
	case BorderBlock:
		return lipgloss.BlockBorder()
	case BorderASCII:
		return lipgloss.ASCIIBorder()
	case BorderRounded:
		return lipgloss.RoundedBorder()
	default:
		return lipgloss.RoundedBorder()
	}
}

func panelFrame(t theme.Theme, o PanelOptions) lipgloss.Style {
	style := t.Panel.Border(panelBorder(t, o.Border))
	if o.Glow && t.Rich {
		return style.BorderForegroundBlend(t.Palette.Gradient...)
	}
	return style.BorderForeground(t.Palette.Accent)
}

func panelTitleLine(t theme.Theme, title string, focused bool) string {
	if focused {
		return t.Title.Render(title)
	}
	return t.Secondary.Render(title)
}

func cardFrame(t theme.Theme, selected bool) lipgloss.Style {
	base := t.Card.Border(panelBorder(t, BorderRounded))
	if !selected {
		return base
	}
	if t.Rich {
		return base.BorderForegroundBlend(t.Palette.Gradient...)
	}
	return base.BorderForeground(t.Palette.Accent)
}

func cardTitleLine(t theme.Theme, title string, selected bool) string {
	if selected {
		return t.Selected.Render(title)
	}
	return t.Strong.Render(title)
}

func panelFit(s lipgloss.Style, width int) lipgloss.Style {
	if width <= 0 {
		return s
	}
	return s.Width(width).MaxWidth(width)
}
