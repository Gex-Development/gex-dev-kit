package components

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
	"github.com/charmbracelet/x/ansi"
)

func choiceTestOptions(n int) []ChoiceOption {
	o := make([]ChoiceOption, n)
	for i := range o {
		o[i] = ChoiceOption{ID: fmt.Sprint(i), Label: fmt.Sprintf("option%d", i)}
	}
	return o
}

func choiceDone(t *testing.T, cmd tea.Cmd, want string) {
	t.Helper()
	if cmd == nil {
		t.Fatal("missing command")
	}
	got, ok := cmd().(ChoiceDoneMsg)
	if !ok || strings.Join(got.IDs, ",") != want {
		t.Fatalf("done: %#v", cmd())
	}
}

func TestChoiceDigitAndBack(t *testing.T) {
	c := NewChoice("Question", choiceTestOptions(3), false, true)
	c, cmd := c.Update(tea.KeyPressMsg{Code: '2', Text: "2"})
	choiceDone(t, cmd, "1")
	_, cmd = c.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if _, ok := cmd().(ChoiceBackMsg); !ok {
		t.Fatal("escape not back")
	}
}

func TestChoiceRequiredMultiOrder(t *testing.T) {
	c := NewChoice("Question", choiceTestOptions(3), true, true)
	var cmd tea.Cmd
	c, cmd = c.Update(tea.KeyPressMsg{Code: '3', Text: "3"})
	if cmd != nil {
		t.Fatal("digit must only toggle")
	}
	c, cmd = c.Update(tea.KeyPressMsg{Code: '1', Text: "1"})
	if cmd != nil {
		t.Fatal("digit must only toggle")
	}
	_, cmd = c.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	choiceDone(t, cmd, "0,2")
}

func TestChoiceMultiEmptyEnter(t *testing.T) {
	c := NewChoice("Question", choiceTestOptions(2), true, true)
	_, cmd := c.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	choiceDone(t, cmd, "0")
	c = c.move(2)
	c, cmd = c.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd != nil || !strings.Contains(c.View(theme.New(theme.Options{Plain: true}), 60), "[ Marque uma opção ]") {
		t.Fatal("required button stays disabled when empty")
	}
	c = NewChoice("Question", choiceTestOptions(2), true, false)
	_, cmd = c.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	choiceDone(t, cmd, "")
}

func TestChoicePagesAndMultiConfirm(t *testing.T) {
	plain := theme.New(theme.Options{Plain: true})
	c := NewChoice("Question", choiceTestOptions(11), true, true).MaxRows(4).Place(3, 11, 40)
	if c.Pages() != 3 || c.Page() != 0 {
		t.Fatalf("pages = %d, page = %d", c.Pages(), c.Page())
	}
	view := ansi.Strip(c.View(plain, 40))
	if !strings.Contains(view, "1. [ ] option0") || !strings.Contains(view, "4. [ ] option3") || strings.Contains(view, "option4") || !strings.Contains(view, "< pagina 1/3 >") {
		t.Fatalf("first page: %q", view)
	}
	c, _ = c.Update(tea.KeyPressMsg{Code: '1'})
	c, _ = c.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	if c.Page() != 1 || c.cursor != 4 {
		t.Fatalf("next page and cursor: %d %d", c.Page(), c.cursor)
	}
	c, _ = c.Update(tea.KeyPressMsg{Code: '2'})
	if strings.Join(c.Selected(), ",") != "0,5" {
		t.Fatalf("relative digit: %v", c.Selected())
	}
	for range 3 {
		c, _ = c.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	}
	c, _ = c.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if c.Page() != 1 || c.cursor != 4 {
		t.Fatalf("down crossed page: %d %d", c.Page(), c.cursor)
	}
	c, _ = c.Update(tea.KeyPressMsg{Code: tea.KeyPgDown})
	if c.Page() != 2 || c.cursor != 8 || !strings.Contains(c.View(plain, 40), "[ Confirmar (2) ]") {
		t.Fatalf("last page lost marks: %d %d %s", c.Page(), c.cursor, c.View(plain, 40))
	}
	_, cmd := c.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	choiceDone(t, cmd, "0,5")
}

func TestChoicePageClickAndSingleDigit(t *testing.T) {
	c := NewChoice("Question", choiceTestOptions(11), false, true).MaxRows(4).Place(3, 11, 40)
	footer := ansi.Strip(c.View(theme.New(theme.Options{Plain: true}), 40))
	lines := strings.Split(footer, "\n")
	y := c.originY + len(lines) - 3
	x := c.originX + ansi.StringWidth(strings.Split(lines[len(lines)-2], ">")[0])
	c, _ = c.Update(tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft})
	if c.Page() != 1 || c.cursor != 4 {
		t.Fatalf("next arrow click: %d %d", c.Page(), c.cursor)
	}
	c, _ = c.Update(tea.MouseClickMsg{X: 4, Y: choiceRowY(t, c, "option5"), Button: tea.MouseLeft})
	if c.cursor != 5 {
		t.Fatalf("second visible item cursor = %d", c.cursor)
	}
	_, cmd := c.Update(tea.KeyPressMsg{Code: '2'})
	choiceDone(t, cmd, "5")
	one := NewChoice("Question", choiceTestOptions(3), false, true).MaxRows(4)
	if one.Pages() != 1 || strings.Contains(one.View(theme.New(theme.Options{Plain: true}), 40), "pagina") {
		t.Fatal("single page gained footer")
	}
}

func TestChoiceMultiPlainHintAndSpace(t *testing.T) {
	c := NewChoice("Question", choiceTestOptions(2), true, false)
	c, cmd := c.Update(tea.KeyPressMsg{Code: ' '})
	if cmd != nil || strings.Join(c.Selected(), ",") != "0" {
		t.Fatal("space toggles without confirming")
	}
	view := c.View(theme.New(theme.Options{Plain: true}), 100)
	if !strings.Contains(view, "espaco marca | enter confirma") || strings.ContainsAny(view, "↑↓·☐☑") {
		t.Fatalf("plain hint not ASCII: %q", view)
	}
}

func TestChoiceMultiValueCopiesSelection(t *testing.T) {
	original := NewChoice("Question", choiceTestOptions(2), true, true)
	next, _ := original.Update(tea.KeyPressMsg{Code: '1', Text: "1"})
	if len(original.Selected()) != 0 || strings.Join(next.Selected(), ",") != "0" {
		t.Fatal("selection mutated previous value")
	}
	again := next.Preselect([]string{"1"})
	if strings.Join(next.Selected(), ",") != "0" || strings.Join(again.Selected(), ",") != "1" {
		t.Fatal("preselect mutated previous value")
	}
}

func choiceRowY(t *testing.T, c Choice, label string) int {
	t.Helper()
	rendered := strings.Split(ansi.Strip(c.View(theme.New(theme.Options{Plain: true}), 40)), "\n")
	for i, line := range rendered {
		if strings.Contains(line, label) {
			return c.originY + i - 1
		}
	}
	t.Fatalf("option %q not rendered", label)
	return -1
}

func TestChoiceClickRenderedLine(t *testing.T) {
	c := NewChoice("Question", []ChoiceOption{{ID: "a", Label: "first", Hint: "detail"}, {ID: "b", Label: "second"}}, false, true).Place(3, 11)
	c, cmd := c.Update(tea.MouseClickMsg{X: 4, Y: choiceRowY(t, c, "second"), Button: tea.MouseLeft})
	if cmd != nil || c.cursor != 1 {
		t.Fatalf("first click: %d %v", c.cursor, cmd)
	}
	_, cmd = c.Update(tea.MouseClickMsg{X: 4, Y: choiceRowY(t, c, "second"), Button: tea.MouseLeft})
	choiceDone(t, cmd, "b")
}

func TestChoicePageWrap(t *testing.T) {
	c := NewChoice("Question", choiceTestOptions(6), false, true).MaxRows(3)
	c, _ = c.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	if c.Page() != 0 || c.cursor != 2 {
		t.Fatalf("up crossed page: %d %d", c.Page(), c.cursor)
	}
	c, _ = c.Update(tea.KeyPressMsg{Code: tea.KeyPgDown})
	if c.Page() != 1 || c.cursor != 3 {
		t.Fatalf("page down: %d %d", c.Page(), c.cursor)
	}
	c, _ = c.Update(tea.KeyPressMsg{Code: tea.KeyPgUp})
	if c.Page() != 0 || c.cursor != 0 {
		t.Fatalf("page up: %d %d", c.Page(), c.cursor)
	}
	c, _ = c.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	c = c.MaxRows(4)
	if c.cursor < c.pageStart() || c.cursor >= c.pageEnd() {
		t.Fatalf("resizing hid cursor: page %d cursor %d", c.Page(), c.cursor)
	}
	checkChatWidth(t, "choice", func(th theme.Theme, w int) string { return c.View(th, w) })
}

func TestChoiceSpacedAddsBlankRowAndKeepsClicksAligned(t *testing.T) {
	plain := theme.New(theme.Options{Plain: true})
	c := NewChoice("Question", choiceTestOptions(3), false, true).Spaced().Place(0, 5)
	lines := strings.Split(ansi.Strip(c.View(plain, 40)), "\n")
	if len(lines) != c.Height(40) || lines[2] != "" || lines[4] != "" {
		t.Fatalf("height %d, rows %q", c.Height(40), lines)
	}
	if got := c.optionAt(0, choiceRowY(t, c, "option2")); got != 2 {
		t.Fatalf("click on third option resolved to %d", got)
	}
	if got := c.optionAt(0, 5+1); got != -1 {
		t.Fatalf("blank row is clickable: %d", got)
	}
	checkChatWidth(t, "choice spaced", func(th theme.Theme, w int) string { return c.View(th, w) })
}

func TestChoiceArtRowsAndClicks(t *testing.T) {
	plain := theme.New(theme.Options{Plain: true})
	opts := []ChoiceOption{
		{ID: "a", Label: "first", Icon: "[A]", Art: func(theme.Theme) []string { return []string{" /A\\ ", "<AAA>", " \\A/ "} }},
		{ID: "b", Label: "second", Icon: "[B]", Hint: "detail", Art: func(theme.Theme) []string { return []string{" /B\\ ", "<BBB>", " \\B/ "} }},
	}
	c := NewChoice("Question", opts, true, true).WithArt(true).Spaced().Place(0, 5, 40)
	lines := strings.Split(ansi.Strip(c.View(plain, 40)), "\n")
	if len(lines) != c.Height(40) || len(lines) != 11 {
		t.Fatalf("art height %d, rows %q", c.Height(40), lines)
	}
	if !strings.Contains(lines[2], "1. [ ] <AAA> first") || !strings.Contains(lines[6], "2. [ ] <BBB> second") {
		t.Fatalf("labels outside middle row: %q", lines)
	}
	for _, line := range lines[1:4] {
		if !strings.HasPrefix(line, "> ") {
			t.Fatalf("cursor missing from art row: %q", line)
		}
	}
	if lines[4] != "" || c.optionAt(0, 5+3) != -1 {
		t.Fatalf("blank row clickable: %q", lines)
	}
	for row := range 3 {
		next, cmd := c.Update(tea.MouseClickMsg{X: 0, Y: 5 + 4 + row, Button: tea.MouseLeft})
		if cmd != nil || next.cursor != 1 || strings.Join(next.Selected(), ",") != "b" {
			t.Fatalf("art row %d click: %d %v %v", row, next.cursor, next.Selected(), cmd)
		}
	}
	hover, _ := c.Update(tea.MouseMotionMsg{X: 0, Y: 5 + 4, Button: tea.MouseNone})
	if hover.hover != 1 {
		t.Fatalf("art first row hover resolved to %d", hover.hover)
	}
	if !strings.Contains(ansi.Strip(c.WithArt(false).View(plain, 40)), "[A] first") || strings.Contains(c.WithArt(false).View(plain, 40), "<AAA>") {
		t.Fatalf("compact art fallback did not use icon: %q", c.WithArt(false).View(plain, 40))
	}
}

func TestChoiceSingleLineArtRendersInsteadOfIcon(t *testing.T) {
	c := NewChoice("Question", []ChoiceOption{{ID: "a", Label: "name", Icon: "old", Art: func(theme.Theme) []string { return []string{"new"} }}}, false, true).WithArt(true)
	view := ansi.Strip(c.View(theme.New(theme.Options{Plain: true}), 40))
	if !strings.Contains(view, "new name") || strings.Contains(view, "old") || c.Height(40) != len(strings.Split(view, "\n")) {
		t.Fatalf("single-line art not rendered: %q", view)
	}
}

func TestChoiceArtPagesWidthAndPlain(t *testing.T) {
	plain := theme.New(theme.Options{Plain: true})
	opts := []ChoiceOption{
		{ID: "a", Label: strings.Repeat("long label ", 12), Icon: "A", Art: func(t theme.Theme) []string { return HarnessArt(t, "claude") }},
		{ID: "b", Label: "second", Icon: "B", Art: func(t theme.Theme) []string { return HarnessArt(t, "codex") }},
		{ID: "c", Label: "third", Icon: "C", Art: func(t theme.Theme) []string { return HarnessArt(t, "omp") }},
	}
	c := NewChoice("Question", opts, false, true).WithArt(true).MaxRows(2)
	c, _ = c.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	if c.Pages() != 2 || c.Page() != 1 || !strings.Contains(c.View(plain, 40), "third") || strings.Contains(c.View(plain, 40), "long label") {
		t.Fatalf("art page lost cursor: %s", c.View(plain, 40))
	}
	if !strings.Contains(ansi.Strip(c.View(plain, 10)), "C") || strings.Contains(c.View(plain, 10), "---------") {
		t.Fatal("narrow art should fall back to icon")
	}
	for offset := range 91 {
		width := offset + 10
		view := c.View(plain, width)
		if len(strings.Split(view, "\n")) != c.Height(width) {
			t.Fatalf("width %d: height mismatch", width)
		}
		for _, line := range strings.Split(ansi.Strip(view), "\n") {
			if ansi.StringWidth(line) > width || !isASCII(line) {
				t.Fatalf("width %d: invalid plain row %q", width, line)
			}
		}
	}
	checkChatWidth(t, "choice art", func(th theme.Theme, width int) string { return c.View(th, width) })
}
