package components

import (
	"image/color"
	"strings"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
	"github.com/charmbracelet/x/ansi"
)

const harnessIconWidth = 9

func HarnessIcon(t theme.Theme, name string) string {
	brand, glyph, fallback := harnessBrand(t, name)
	if glyph == "" {
		return ""
	}
	if t.Plain {
		return fallback
	}
	edge := t.Text.Foreground(brand)
	face := t.Text.Foreground(t.Palette.Navy).Background(brand).Bold(true)
	mid := harnessCenter(glyph, harnessIconWidth-2)
	return strings.Join([]string{
		edge.Render("▗" + strings.Repeat("▄", harnessIconWidth-2) + "▖"),
		edge.Render("█") + face.Render(mid) + edge.Render("█"),
		edge.Render("▝" + strings.Repeat("▀", harnessIconWidth-2) + "▘"),
	}, "\n")
}

func HarnessBadge(t theme.Theme, name string) string {
	brand, glyph, fallback := harnessBrand(t, name)
	if glyph == "" {
		return ""
	}
	if t.Plain {
		return fallback + strings.Repeat(" ", 7-ansi.StringWidth(fallback))
	}
	if name == "claude" {
		return t.Text.Foreground(brand).Render("▐▛███▜▌")
	}
	return t.Text.Foreground(t.Palette.Navy).Background(brand).Bold(true).Render(harnessCenter(glyph, 7))
}

func HarnessArt(t theme.Theme, id string) []string {
	brand, glyph, _ := harnessBrand(t, id)
	if glyph == "" {
		return nil
	}
	if t.Plain {
		switch id {
		case "claude":
			return []string{"  /###\\  ", " <#####> ", "  || ||  "}
		case "codex":
			return []string{"  .---.  ", " ( >_  ) ", "  '---'  "}
		}
		return []string{"---------", harnessCenter("[PI]", 9), "---------"}
	}
	switch id {
	case "claude":
		rows := []string{" ▐▛███▜▌ ", "▝▜█████▛▘", "  ▘▘ ▝▝  "}
		for i := range rows {
			rows[i] = t.Text.Foreground(brand).Render(rows[i])
		}
		return rows
	case "codex":
		edge := t.Text.Foreground(brand)
		face := t.Text.Foreground(t.Palette.Navy).Background(brand).Bold(true)
		return []string{edge.Render("▗▄██▄██▄▖"), face.Render(harnessCenter(glyph, 9)), edge.Render("▝▀█████▀▘")}
	}
	style := t.Text.Foreground(t.Palette.Navy).Background(brand).Bold(true)
	return []string{style.Render("         "), style.Render(harnessCenter(glyph, 9)), style.Render("         ")}
}

func HarnessBanner(t theme.Theme, width int, id string, info []string) string {
	art := HarnessArt(t, id)
	if len(art) == 0 {
		return ""
	}
	if width < 30 {
		return strings.Join(art, "\n")
	}
	for i := range art {
		if i >= len(info) {
			continue
		}
		style := t.Secondary
		if i == 0 {
			style = t.Strong
		}
		ellipsis := "…"
		if t.Plain {
			ellipsis = "..."
		}
		art[i] += "  " + style.Render(ansi.Truncate(info[i], width-11, ellipsis))
	}
	return strings.Join(art, "\n")
}

func harnessBrand(t theme.Theme, name string) (color.Color, string, string) {
	switch name {
	case "claude":
		return t.Palette.BrandClaude, "✻", "[CC]"
	case "codex":
		return t.Palette.BrandCodex, ">_", "[>_]"
	case "omp":
		return t.Palette.BrandOMP, "π", "[PI]"
	default:
		return nil, "", ""
	}
}

func harnessCenter(glyph string, width int) string {
	gap := max(width-ansi.StringWidth(glyph), 0)
	left := gap / 2
	return strings.Repeat(" ", left) + glyph + strings.Repeat(" ", gap-left)
}
