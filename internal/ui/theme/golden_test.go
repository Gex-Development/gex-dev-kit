package theme

import (
	"fmt"
	"image/color"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/exp/golden"
)

type goldenStyle struct {
	name  string
	style lipgloss.Style
}

func TestGoldenDark(t *testing.T) {
	golden.RequireEqual(t, goldenSample(New(Options{Dark: true, Profile: colorprofile.TrueColor})))
}

func TestGoldenLight(t *testing.T) {
	golden.RequireEqual(t, goldenSample(New(Options{Profile: colorprofile.TrueColor})))
}

func goldenSample(th Theme) string {
	var b strings.Builder
	b.WriteString("Tokens\n")
	for _, item := range tokensList(th.Palette) {
		fmt.Fprintf(&b, "%-14s %s %s\n", item.name, tokensHex(item.value), goldenSwatches([]color.Color{item.value}, "  "))
	}
	b.WriteString("\nEstilos\n")
	for _, item := range goldenStyles(th) {
		b.WriteString(item.name + "\n" + item.style.Render("Exemplo "+item.name) + "\n")
	}
	b.WriteString("\nEstados\n")
	for state := StatePending; state <= StateInfo; state++ {
		fmt.Fprintf(&b, "%d %s\n", state, th.StateStyle(state).Render(""))
	}
	b.WriteString("\nGradiente\n" + goldenSwatches(th.Blend(40), " ") + "\n")
	b.WriteString("\nBarra de progresso\n" + goldenSwatches(th.ProgressColors(), " ") + "\n")
	return b.String()
}

func goldenStyles(th Theme) []goldenStyle {
	return []goldenStyle{
		{"Text", th.Text},
		{"Secondary", th.Secondary},
		{"Faint", th.Faint},
		{"Strong", th.Strong},
		{"Accent", th.Accent},
		{"Title", th.Title},
		{"Eyebrow", th.Eyebrow},
		{"Code", th.Code},
		{"Success", th.Success},
		{"Warning", th.Warning},
		{"Error", th.Error},
		{"Info", th.Info},
		{"Selected", th.Selected},
		{"Unselected", th.Unselected},
		{"Panel", th.Panel},
		{"PanelFocused", th.PanelFocused},
		{"Card", th.Card},
		{"Modal", th.Modal},
		{"KeyHintKey", th.KeyHintKey},
		{"KeyHintDesc", th.KeyHintDesc},
		{"Divider", th.Divider},
	}
}

func goldenSwatches(cores []color.Color, mark string) string {
	var b strings.Builder
	for _, c := range cores {
		b.WriteString(lipgloss.NewStyle().Background(c).Render(mark))
	}
	return b.String()
}
