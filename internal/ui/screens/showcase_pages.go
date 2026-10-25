package screens

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/components"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/terminal"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

const (
	showcasePageIdentidade  = "Identidade"
	showcasePageBordas      = "Bordas e cor"
	showcasePageLayout      = "Layout"
	showcasePageMedidas     = "Medidas"
	showcasePageTerminal    = "Terminal"
	showcasePagePrograma    = "Programa"
	showcasePageMovimento   = "Movimento"
	showcasePageFeedback    = "Feedback"
	showcasePageNavegacao   = "Navegação"
	showcasePageDados       = "Dados"
	showcasePageFormularios = "Formulários"
	showcasePageAcessivel   = "Acessível"
	showcasePageBubbles     = "Bubbles"
	showcasePageTestes      = "Testes"
	showcasePageCobertura   = "Cobertura Charm"
)

type showcaseDemo func(t theme.Theme, width int) string

type showcaseLive struct {
	Dark      bool
	Motion    bool
	Profile   colorprofile.Profile
	Env       terminal.Env
	Width     int
	Height    int
	AltScreen bool
	Focused   bool
	Clicks    int
	LastMouse string
	Ticks     int
	Batch     []int
	Sequence  []int
	Fault     error
	Printed   int
	Frames    int
	Last      int
}

type showcasePage struct {
	Name  string
	Own   func(l showcaseLive, t theme.Theme, width int) string
	Demos []showcaseDemo
}

type showcaseFault struct {
	Cause  string
	Action string
}

func (f showcaseFault) Error() string {
	return f.Cause
}

type showcaseTestFile struct {
	Path  string
	Kind  string
	Tests string
}

var showcaseTestFiles = []showcaseTestFile{
	{Path: "internal/ui/theme/golden_test.go", Kind: "golden", Tests: "TestGoldenDark, TestGoldenLight"},
	{Path: "internal/ui/motion/motion_test.go", Kind: "teatest", Tests: "quadros param quando a mola assenta"},
	{Path: "internal/ui/components/steps_test.go", Kind: "teatest", Tests: "barra com mola e cronômetro"},
	{Path: "internal/ui/components/toast_test.go", Kind: "teatest", Tests: "toast some depois do prazo"},
	{Path: "internal/ui/components/loading_test.go", Kind: "teatest", Tests: "esqueleto de carregamento"},
	{Path: "internal/ui/screens/showcase_test.go", Kind: "teatest", Tests: "navegação, tema e movimento da vitrine"},
}

func showcasePages() []showcasePage {
	return []showcasePage{
		{Name: showcasePageIdentidade, Demos: []showcaseDemo{demoWordmark, demoBanner, demoCursor}},
		{Name: showcasePageBordas, Demos: []showcaseDemo{demoBorders, demoGradient, demoPanel, demoAdaptive}},
		{Name: showcasePageLayout, Demos: []showcaseDemo{demoJoin, demoPlace, demoSection, demoWindowSize}},
		{Name: showcasePageMedidas, Own: showcaseMeasuresOwn, Demos: []showcaseDemo{demoMeasures, demoAnsi}},
		{Name: showcasePageTerminal, Own: showcaseTerminalOwn, Demos: []showcaseDemo{demoTerm}},
		{Name: showcasePagePrograma, Own: showcaseProgramOwn, Demos: []showcaseDemo{demoAltScreen, demoMouse, demoFocus, demoBatchSequence, demoPrintln, demoSuspend, demoTypedErrors}},
		{Name: showcasePageMovimento, Demos: []showcaseDemo{demoTick, demoSpring, demoProjectile}},
		{Name: showcasePageFeedback, Demos: []showcaseDemo{demoChecklist, demoSteps, demoProgress, demoToast, demoTimer, demoStopwatch, demoErrorBox, demoEmptyState, demoLoading, demoBadge, demoLipglossList}},
		{Name: showcasePageNavegacao, Demos: []showcaseDemo{demoKeyHints, demoMenu, demoConfirm, demoHelp}},
		{Name: showcasePageDados, Demos: []showcaseDemo{demoBubblesTable, demoDSTable, demoLipglossTable, demoTree, demoFileTree, demoDiffView}},
		{Name: showcasePageFormularios, Demos: []showcaseDemo{demoHuhNote, demoHuhInput, demoHuhValidation, demoHuhSelect, demoHuhMultiSelect, demoHuhConfirm, demoHuhFilePicker, demoHuhGroups, demoHuhTheme}},
		{Name: showcasePageAcessivel, Own: showcaseAccessibleOwn, Demos: []showcaseDemo{demoHuhText, demoHuhAccessible}},
		{Name: showcasePageBubbles, Demos: []showcaseDemo{demoSpinner, demoList, demoTextInput, demoTextArea, demoViewport, demoPaginator, demoFilePicker}},
		{Name: showcasePageTestes, Own: showcaseTestsOwn},
		{Name: showcasePageCobertura, Own: showcaseCoverageOwn},
	}
}

func showcasePageAt(i int) showcasePage {
	pages := showcasePages()
	return pages[min(max(i, 0), len(pages)-1)]
}

func showcaseBody(p showcasePage, l showcaseLive, t theme.Theme, width int) string {
	var parts []string
	if p.Own != nil {
		parts = append(parts, p.Own(l, t, width))
	}
	for _, demo := range p.Demos {
		if out := demo(t, width); out != "" {
			parts = append(parts, out)
		}
	}
	if width > 0 {
		for i, part := range parts {
			parts[i] = ansi.Hardwrap(part, width, true)
		}
	}
	return strings.Join(parts, "\n\n")
}

func showcaseProgramOwn(l showcaseLive, t theme.Theme, width int) string {
	lines := []string{
		t.Title.Render("Programa: recursos de runtime do bubbletea"),
		"modo: " + showcaseModeLabel(l.AltScreen),
		"tamanho: " + showcaseSize(l),
		"janela com foco: " + showcaseYesNo(l.Focused),
		fmt.Sprintf("cliques: %d · último evento de mouse: %s", l.Clicks, showcaseOrDash(l.LastMouse)),
		fmt.Sprintf("ticks de 1 s (tea.Tick): %d", l.Ticks),
		"Batch: " + showcaseOrder(l.Batch),
		"Sequence: " + showcaseOrder(l.Sequence),
		fmt.Sprintf("linhas impressas com tea.Println: %d", l.Printed),
		"quadros de movimento: " + strconv.Itoa(l.Frames),
		showcaseSuspendLine(),
	}
	if fault := showcaseFaultBox(l.Fault, t, width); fault != "" {
		lines = append(lines, fault)
	} else {
		lines = append(lines, t.Secondary.Render("tecla e gera um erro tipado"))
	}
	lines = append(lines, t.Secondary.Render("teclas: a alt screen · p println (só inline) · 1 batch · 2 sequence · e erro"))
	return strings.Join(lines, "\n")
}

func showcaseFaultBox(err error, t theme.Theme, width int) string {
	var fault showcaseFault
	if !errors.As(err, &fault) {
		return ""
	}
	return components.ErrorBox(t, width, "Erro tipado", fault.Cause, fault.Action)
}

func showcaseSuspendLine() string {
	state := "ativo"
	if !components.DefaultKeys().Suspend.Enabled() {
		state = "indisponível no Windows"
	}
	return "ctrl+z suspende (" + state + ")"
}

func showcaseTerminalOwn(l showcaseLive, t theme.Theme, width int) string {
	env := l.Env
	mode, err := env.Resolve(terminal.ModeAuto)
	modeText := "—"
	if err == nil {
		modeText = showcaseTerminalMode(mode)
	}
	rows := [][]string{
		{"stdin TTY", showcaseYesNo(env.StdinTTY)},
		{"stdout TTY", showcaseYesNo(env.StdoutTTY)},
		{"tamanho", showcaseSize(l)},
		{"perfil de cor", l.Profile.String()},
		{"movimento ativo", showcaseYesNo(l.Motion)},
		{"modo automático", modeText},
		{"NO_COLOR", showcaseYesNo(env.NoColor)},
		{"TERM", showcaseOrDash(env.Term)},
		{"ACCESSIBLE", showcaseYesNo(env.Accessible)},
		{"CI", showcaseYesNo(env.CI)},
		{"GEX_NO_ANIM", showcaseYesNo(env.NoAnim)},
	}
	return components.StaticTable(t, width, []string{"Item", "Valor"}, rows)
}

func showcaseMeasuresOwn(_ showcaseLive, t theme.Theme, width int) string {
	samples := []struct {
		label string
		text  string
	}{
		{"ASCII", "cafe"},
		{"NFC", "café"},
		{"NFD", "cafe" + string(rune(0x0301))},
		{"CJK", "漢字"},
		{"emoji", "🙂 ok"},
	}
	rows := make([][]string, 0, len(samples))
	for _, s := range samples {
		rows = append(rows, []string{
			s.label,
			strconv.Itoa(len(s.text)),
			strconv.Itoa(ansi.StringWidth(s.text)),
			strconv.Itoa(lipgloss.Width(s.text)),
			ansi.Truncate(s.text, 6, "…"),
		})
	}
	table := components.StaticTable(t, width, []string{"Caso", "len", "ansi.StringWidth", "lipgloss.Width", "Truncate 6"}, rows)
	wrapped := ansi.Hardwrap("texto longo com acentos (café), CJK (漢字) e emoji 🙂 para quebra de linha", max(width/2, 8), true)
	return strings.Join([]string{
		table,
		fmt.Sprintf("ansi.Hardwrap: %d linhas", lipgloss.Height(wrapped)),
		wrapped,
		fmt.Sprintf("ansi.Strip de texto com estilo: %s", ansi.Strip(t.Strong.Render("negrito"))),
	}, "\n")
}

func showcaseAccessibleOwn(l showcaseLive, _ theme.Theme, width int) string {
	page := showcasePageAt(l.Last)
	plain := theme.New(theme.Options{Dark: l.Dark, Plain: true})
	if page.Name == showcasePageAcessivel {
		return plain.Title.Render("Versão Plain: escolha outra página para ver o conteúdo")
	}
	head := plain.Title.Render("Versão Plain de " + page.Name + ": sem cor e sem símbolos decorativos")
	return head + "\n\n" + showcaseBody(page, l, plain, width)
}

func showcaseTestsOwn(_ showcaseLive, t theme.Theme, _ int) string {
	lines := []string{t.Title.Render("Testes golden e teatest do repositório")}
	for _, f := range showcaseTestFiles {
		lines = append(lines, fmt.Sprintf("%s · %s · %s", f.Path, f.Kind, f.Tests))
	}
	lines = append(lines, t.Secondary.Render("x/exp/teatest e x/exp/golden são bibliotecas de teste: não têm tela própria."))
	return strings.Join(lines, "\n")
}

func showcaseCoverageOwn(_ showcaseLive, t theme.Theme, width int) string {
	rows := make([][]string, 0, len(CoverageMap))
	for _, entry := range CoverageMap {
		rows = append(rows, []string{entry.Lib, entry.Feature, entry.ScreenFile, entry.Page})
	}
	return strings.Join([]string{
		t.Title.Render(fmt.Sprintf("Mapa de cobertura: %d entradas", len(CoverageMap))),
		components.StaticTable(t, width, []string{"Lib", "Recurso", "Tela", "Página"}, rows),
	}, "\n")
}

func showcaseModeLabel(altScreen bool) string {
	if altScreen {
		return "alt screen"
	}
	return "inline"
}

func showcaseTerminalMode(mode terminal.Mode) string {
	switch mode {
	case terminal.ModeTUI:
		return "TUI"
	case terminal.ModeCLI:
		return "CLI"
	}
	return "automático"
}

func showcaseSize(l showcaseLive) string {
	if l.Width == 0 || l.Height == 0 {
		return "sem medida"
	}
	return fmt.Sprintf("%d×%d", l.Width, l.Height)
}

func showcaseOrder(values []int) string {
	if len(values) == 0 {
		return "—"
	}
	parts := make([]string, len(values))
	for i, v := range values {
		parts[i] = strconv.Itoa(v)
	}
	return strings.Join(parts, " → ")
}

func showcaseYesNo(v bool) string {
	if v {
		return "sim"
	}
	return "não"
}

func showcaseOrDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}
