package screens

import "github.com/Gex-Development/gex-dev-kit/internal/ui/theme"

type Coverage struct {
	Lib        string
	Feature    string
	ScreenFile string
	Call       string
	Page       string
	Demo       func(t theme.Theme, width int) string
	Marker     string
	TestOnly   bool
}

var CoverageMap = []Coverage{
	{Lib: "bubbletea", Feature: "tea.Model, Update e View", ScreenFile: "home.go", Call: "tea.NewView", Page: showcasePageIdentidade, Demo: demoBanner, Marker: "gex-dev-kit"},
	{Lib: "bubbletea", Feature: "tea.Tick", ScreenFile: "home.go", Call: "motion.Frame", Page: showcasePageMovimento, Demo: demoTick, Marker: "█"},
	{Lib: "bubbletea", Feature: "WindowSizeMsg", ScreenFile: "home.go", Call: "tea.WindowSizeMsg", Page: showcasePageLayout, Demo: demoWindowSize, Marker: "80×24"},
	{Lib: "bubbletea", Feature: "alt screen", ScreenFile: "install.go", Call: "AltScreen", Page: showcasePagePrograma, Demo: demoAltScreen, Marker: "modo: alt screen"},
	{Lib: "bubbletea", Feature: "mouse", ScreenFile: "home.go", Call: "tea.MouseModeCellMotion", Page: showcasePagePrograma, Demo: demoMouse, Marker: "último evento de mouse"},
	{Lib: "bubbletea", Feature: "foco e blur", ScreenFile: "home.go", Call: "tea.BlurMsg", Page: showcasePagePrograma, Demo: demoFocus, Marker: "janela com foco"},
	{Lib: "bubbletea", Feature: "tea.Batch e tea.Sequence", ScreenFile: "home.go", Call: "tea.Batch", Page: showcasePagePrograma, Demo: demoBatchSequence, Marker: "Batch:"},
	{Lib: "bubbletea", Feature: "tea.Println fora do alt screen", ScreenFile: "install.go", Call: "tea.Println", Page: showcasePagePrograma, Demo: demoPrintln, Marker: "linhas impressas"},
	{Lib: "bubbletea", Feature: "suspensão (Ctrl+Z)", ScreenFile: "home.go", Call: "tea.Suspend", Page: showcasePagePrograma, Demo: demoSuspend, Marker: "ctrl+z suspende"},
	{Lib: "bubbletea", Feature: "mensagens de erro tipadas", ScreenFile: "install.go", Call: "errMsg", Page: showcasePagePrograma, Demo: demoTypedErrors, Marker: "Erro tipado"},

	{Lib: "bubbles", Feature: "spinner (vários estilos)", ScreenFile: "checklist.go", Call: "spinner.New", Page: showcasePageBubbles, Demo: demoSpinner, Marker: "⣾"},
	{Lib: "bubbles", Feature: "progress com mola", ScreenFile: "steps.go", Call: "progress.WithSpringOptions", Page: showcasePageFeedback, Demo: demoProgress, Marker: "█"},
	{Lib: "bubbles", Feature: "table", ScreenFile: "doctor.go", Call: "table.New", Page: showcasePageDados, Demo: demoBubblesTable, Marker: "Estado"},
	{Lib: "bubbles", Feature: "list com filtro, paginação e delegate", ScreenFile: "catalog.go", Call: "list.New", Page: showcasePageBubbles, Demo: demoList, Marker: "Core"},
	{Lib: "bubbles", Feature: "textinput", ScreenFile: "githooks.go", Call: "textinput.New", Page: showcasePageBubbles, Demo: demoTextInput, Marker: "Digite"},
	{Lib: "bubbles", Feature: "textarea", ScreenFile: "githooks.go", Call: "textarea.New", Page: showcasePageBubbles, Demo: demoTextArea, Marker: "1 "},
	{Lib: "bubbles", Feature: "viewport (prévia rolável)", ScreenFile: "install.go", Call: "viewport.New", Page: showcasePageBubbles, Demo: demoViewport, Marker: "%"},
	{Lib: "bubbles", Feature: "paginator", ScreenFile: "addons.go", Call: "paginator.New", Page: showcasePageBubbles, Demo: demoPaginator, Marker: "●"},
	{Lib: "bubbles", Feature: "help e key", ScreenFile: "keyhints.go", Call: "help.New", Page: showcasePageNavegacao, Demo: demoHelp, Marker: "ajuda"},
	{Lib: "bubbles", Feature: "filepicker", ScreenFile: "githooks.go", Call: "filepicker.New", Page: showcasePageBubbles, Demo: demoFilePicker, Marker: "/"},
	{Lib: "bubbles", Feature: "timer", ScreenFile: "toast.go", Call: "timer.New", Page: showcasePageFeedback, Demo: demoTimer, Marker: "s"},
	{Lib: "bubbles", Feature: "stopwatch", ScreenFile: "steps.go", Call: "stopwatch.New", Page: showcasePageFeedback, Demo: demoStopwatch, Marker: "s"},
	{Lib: "bubbles", Feature: "cursor", ScreenFile: "home.go", Call: "cursor.New", Page: showcasePageIdentidade, Demo: demoCursor, Marker: "▏"},

	{Lib: "huh", Feature: "Note", ScreenFile: "installform.go", Call: "huh.NewNote", Page: showcasePageFormularios, Demo: demoHuhNote, Marker: "Core obrigatório"},
	{Lib: "huh", Feature: "Select", ScreenFile: "installform.go", Call: "huh.NewSelect", Page: showcasePageFormularios, Demo: demoHuhSelect, Marker: "destino"},
	{Lib: "huh", Feature: "FilePicker", ScreenFile: "installform.go", Call: "huh.NewFilePicker", Page: showcasePageFormularios, Demo: demoHuhFilePicker, Marker: "pasta"},
	{Lib: "huh", Feature: "Input com validação", ScreenFile: "installform.go", Call: "huh.NewInput", Page: showcasePageFormularios, Demo: demoHuhValidation, Marker: "caminho absoluto"},
	{Lib: "huh", Feature: "MultiSelect", ScreenFile: "installform.go", Call: "huh.NewMultiSelect", Page: showcasePageFormularios, Demo: demoHuhMultiSelect, Marker: "gex-pr-review"},
	{Lib: "huh", Feature: "Confirm", ScreenFile: "installform.go", Call: "huh.NewConfirm", Page: showcasePageFormularios, Demo: demoHuhConfirm, Marker: "Agora não"},
	{Lib: "huh", Feature: "Text", ScreenFile: "accessible.go", Call: "huh.NewText", Page: showcasePageAcessivel, Demo: demoHuhText, Marker: "Comando"},
	{Lib: "huh", Feature: "grupos e páginas", ScreenFile: "installform.go", Call: "huh.NewGroup", Page: showcasePageFormularios, Demo: demoHuhGroups, Marker: "1/"},
	{Lib: "huh", Feature: "tema GEX", ScreenFile: "installform.go", Call: "WithTheme", Page: showcasePageFormularios, Demo: demoHuhTheme, Marker: "tema"},
	{Lib: "huh", Feature: "Accessible(true)", ScreenFile: "accessible.go", Call: "WithAccessible", Page: showcasePageAcessivel, Demo: demoHuhAccessible, Marker: "Resposta"},

	{Lib: "lipgloss", Feature: "JoinHorizontal e JoinVertical", ScreenFile: "layout.go", Call: "lipgloss.JoinHorizontal", Page: showcasePageLayout, Demo: demoJoin, Marker: "│"},
	{Lib: "lipgloss", Feature: "Place", ScreenFile: "layout.go", Call: "lipgloss.Place", Page: showcasePageLayout, Demo: demoPlace, Marker: "┌"},
	{Lib: "lipgloss", Feature: "bordas (normal, arredondada, grossa, dupla, custom)", ScreenFile: "panel.go", Call: "lipgloss.RoundedBorder", Page: showcasePageBordas, Demo: demoBorders, Marker: "╭"},
	{Lib: "lipgloss", Feature: "gradientes", ScreenFile: "wordmark.go", Call: "lipgloss.Blend1D", Page: showcasePageBordas, Demo: demoGradient, Marker: "█"},
	{Lib: "lipgloss", Feature: "table", ScreenFile: "table.go", Call: "lgtable.New", Page: showcasePageDados, Demo: demoLipglossTable, Marker: "│"},
	{Lib: "lipgloss", Feature: "list", ScreenFile: "checklist.go", Call: "list.New", Page: showcasePageFeedback, Demo: demoLipglossList, Marker: "1."},
	{Lib: "lipgloss", Feature: "tree", ScreenFile: "filetree.go", Call: "tree.Root", Page: showcasePageDados, Demo: demoTree, Marker: "├──"},
	{Lib: "lipgloss", Feature: "estilos adaptativos", ScreenFile: "theme.go", Call: "lipgloss.LightDark", Page: showcasePageBordas, Demo: demoAdaptive, Marker: "claro"},
	{Lib: "lipgloss", Feature: "medidas ANSI-safe", ScreenFile: "layout.go", Call: "lipgloss.Width", Page: showcasePageMedidas, Demo: demoMeasures, Marker: "ansi.StringWidth"},
	{Lib: "lipgloss", Feature: "Compositor e Layer", ScreenFile: "confirm.go", Call: "lipgloss.NewCompositor", Page: showcasePageNavegacao, Demo: demoConfirm, Marker: "sim"},

	{Lib: "harmonica", Feature: "NewSpring e Spring.Update", ScreenFile: "spring.go", Call: "harmonica.NewSpring", Page: showcasePageMovimento, Demo: demoSpring, Marker: "Snappy"},
	{Lib: "harmonica", Feature: "NewProjectile", ScreenFile: "burst.go", Call: "harmonica.NewProjectile", Page: showcasePageMovimento, Demo: demoProjectile, Marker: "*"},

	{Lib: "x", Feature: "x/ansi (truncar, largura, strip)", ScreenFile: "layout.go", Call: "ansi.Truncate", Page: showcasePageMedidas, Demo: demoAnsi, Marker: "ansi.Strip"},
	{Lib: "x", Feature: "x/term (TTY e tamanho)", ScreenFile: "terminal.go", Call: "term.IsTerminal", Page: showcasePageTerminal, Demo: demoTerm, Marker: "stdin TTY"},
	{Lib: "x", Feature: "x/exp/teatest/v2", ScreenFile: "showcase_test.go", Call: "teatest.NewTestModel", Page: showcasePageTestes, TestOnly: true},
	{Lib: "x", Feature: "x/exp/golden", ScreenFile: "golden_test.go", Call: "golden.RequireEqual", Page: showcasePageTestes, TestOnly: true},
}
