package screens

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/cursor"
	"charm.land/bubbles/v2/help"
	blist "charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/paginator"
	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/stopwatch"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/timer"
	"charm.land/bubbles/v2/viewport"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	lglist "charm.land/lipgloss/v2/list"
	lgtree "charm.land/lipgloss/v2/tree"
	"github.com/charmbracelet/harmonica"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/term"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/components"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

type demoItem struct {
	titulo    string
	descricao string
}

func (i demoItem) Title() string       { return i.titulo }
func (i demoItem) Description() string { return i.descricao }
func (i demoItem) FilterValue() string { return i.titulo }

func demoFit(s string, width int) string {
	if width <= 0 {
		return ""
	}
	linhas := strings.Split(s, "\n")
	if len(linhas) > 12 {
		linhas = linhas[:12]
	}
	for i, linha := range linhas {
		linhas[i] = ansi.Truncate(linha, width, "")
	}
	return strings.Join(linhas, "\n")
}

func demoBarra(cheio, total int) string {
	cheio = max(0, min(cheio, total))
	return strings.Repeat("█", cheio) + strings.Repeat("░", total-cheio)
}

func demoGrade(celulas []string, largura, width int) string {
	colunas := max(1, width/largura)
	var linhas []string
	for i := 0; i < len(celulas); i += colunas {
		fim := min(i+colunas, len(celulas))
		linhas = append(linhas, lipgloss.JoinHorizontal(lipgloss.Top, celulas[i:fim]...))
	}
	return lipgloss.JoinVertical(lipgloss.Left, linhas...)
}

func demoSimNao(valor bool) string {
	if valor {
		return "sim"
	}
	return "não"
}

func demoCaminhoAbsoluto(s string) error {
	if filepath.IsAbs(s) {
		return nil
	}
	return errors.New("use caminho absoluto")
}

func demoWordmark(t theme.Theme, width int) string {
	return demoFit(components.Wordmark(t, width, 1), width)
}

func demoBanner(t theme.Theme, width int) string {
	return demoFit(components.Banner(t, width, "gex-dev-kit", "vitrine do Design System"), width)
}

func demoPanel(t theme.Theme, width int) string {
	o := components.PanelOptions{Title: "Painel", Width: min(width, 40), Border: components.BorderRounded, Focused: true}
	return demoFit(components.Panel(t, o, "conteúdo do painel"), width)
}

func demoBadge(t theme.Theme, width int) string {
	estados := []theme.State{theme.StateOK, theme.StateWarn, theme.StateError, theme.StateInfo, theme.StatePending, theme.StateRunning, theme.StateSkipped}
	var selos []string
	for _, s := range estados {
		selos = append(selos, components.StatusBadge(t, s, s.Label()))
	}
	return demoFit(strings.Join(selos, " "), width)
}

func demoChecklist(t theme.Theme, width int) string {
	itens := []components.ChecklistItem{
		{Label: "Core", State: theme.StateOK, Detail: "12 skills"},
		{Label: "Hooks", State: theme.StateRunning, Detail: "gravando"},
		{Label: "Addons", State: theme.StatePending},
	}
	return demoFit(components.NewChecklist(itens, false).View(t, width), width)
}

func demoSteps(t theme.Theme, width int) string {
	passos := components.NewStepProgress(t, []string{"Prévia", "Gravação", "Resumo"}, false)
	passos, _ = passos.Advance(theme.StateOK)
	return demoFit(passos.View(t, width), width)
}

func demoToast(t theme.Theme, width int) string {
	aviso, _ := components.NewToast().Show("Core gravado", theme.StateOK, 3*time.Second)
	return demoFit(aviso.View(t, width), width)
}

func demoKeyHints(t theme.Theme, width int) string {
	return demoFit(components.KeyHints(t, width, false, components.DefaultKeys()), width)
}

func demoDiffView(t theme.Theme, width int) string {
	grupos := []components.ChangeGroup{{
		Title: "Arquivos do HOME",
		Changes: []components.Change{
			{Action: components.ChangeCreate, Path: "CLAUDE.md", Detail: "novo", Lines: []components.DiffLine{
				{Kind: components.DiffAdded, Text: "regra nova"},
				{Kind: components.DiffRemoved, Text: "regra antiga"},
			}},
			{Action: components.ChangeUpdate, Path: "AGENTS.md", Detail: "ajustado"},
		},
	}}
	return demoFit(components.DiffView(t, width, grupos, map[string]bool{"CLAUDE.md": true}), width)
}

func demoFileTree(t theme.Theme, width int) string {
	caminhos := []string{"README.md", "cmd/gex-dev-kit/main.go", "internal/ui/home.go", "internal/ui/screens/coverage.go"}
	return demoFit(components.FileTree(t, "projeto", caminhos), width)
}

func demoDSTable(t theme.Theme, width int) string {
	linhas := [][]string{{"Title", "títulos de tela"}, {"Success", "estado ok"}, {"Error", "estado com erro"}}
	return demoFit(components.StaticTable(t, width, []string{"Token", "Uso"}, linhas), width)
}

func demoEmptyState(t theme.Theme, width int) string {
	return demoFit(components.EmptyState(t, width, "Nada instalado ainda", "rode gex-dev-kit install"), width)
}

func demoErrorBox(t theme.Theme, width int) string {
	return demoFit(components.ErrorBox(t, width, "Falha ao gravar", "o HOME não tem permissão de escrita", "rode com permissão ou troque o HOME"), width)
}

func demoConfirm(t theme.Theme, width int) string {
	const altura = 11
	fundo := make([]string, altura)
	for i := range fundo {
		fundo[i] = strings.Repeat("·", max(width, 0))
	}
	c := components.NewConfirm("Gravar mudanças?", "A instalação altera o HOME.", "sim", "não")
	return demoFit(c.Overlay(t, strings.Join(fundo, "\n"), width, altura), width)
}

func demoMenu(t theme.Theme, width int) string {
	itens := []components.MenuItem{
		{ID: "install", Title: "Instalar", Description: "grava o Core"},
		{ID: "doctor", Title: "Diagnóstico", Description: "confere o HOME"},
		{ID: "ds", Title: "Design System", Description: "vitrine"},
	}
	return demoFit(components.NewMenu(itens, false).View(t, width, true), width)
}

func demoLoading(t theme.Theme, width int) string {
	return demoFit(components.NewLoading(false).View(t, width, 3), width)
}

func demoSection(t theme.Theme, width int) string {
	return demoFit(components.Section(t, width, "Core")+"\n"+components.Divider(t, width), width)
}

func demoSpinner(t theme.Theme, width int) string {
	quadros := []spinner.Spinner{spinner.Line, spinner.Dot, spinner.MiniDot, spinner.Jump, spinner.Pulse, spinner.Points, spinner.Globe, spinner.Moon, spinner.Monkey, spinner.Meter, spinner.Hamburger, spinner.Ellipsis}
	nomes := []string{"Line", "Dot", "MiniDot", "Jump", "Pulse", "Points", "Globe", "Moon", "Monkey", "Meter", "Hamburger", "Ellipsis"}
	var celulas []string
	for i, s := range quadros {
		celulas = append(celulas, t.Text.Width(16).Render(spinner.New(spinner.WithSpinner(s)).View()+" "+nomes[i]))
	}
	return demoFit(demoGrade(celulas, 16, width), width)
}

func demoProgress(t theme.Theme, width int) string {
	larguraBarra := max(4, min(40, width-8))
	barra := progress.New(progress.WithWidth(larguraBarra), progress.WithColors(t.ProgressColors()...), progress.WithFillCharacters(progress.DefaultFullCharFullBlock, '░'), progress.WithSpringOptions(18, 1))
	return demoFit("progresso com mola, 60%:\n"+barra.ViewAs(0.6), width)
}

func demoBubblesTable(t theme.Theme, width int) string {
	colunas := []components.Column{{Title: "Estado", Width: 8}, {Title: "Arquivo", Width: 24}}
	linhas := [][]string{{"ok", "CLAUDE.md"}, {"aviso", "AGENTS.md"}}
	return demoFit(components.NewTable(t, colunas, linhas, 4).View(), width)
}

func demoList(t theme.Theme, width int) string {
	itens := []blist.Item{
		demoItem{"gex-abrir-issue", "Core · abre issue no padrão do time"},
		demoItem{"tdd", "Matt · ciclo vermelho e verde"},
		demoItem{"gex-pr-review", "Especialização · revisa PR"},
	}
	delegate := blist.NewDefaultDelegate()
	delegate.Styles = t.ListItem()
	m := blist.New(itens, delegate, width, 6)
	m.SetShowHelp(false)
	m.SetShowStatusBar(false)
	m.SetShowTitle(false)
	return demoFit(m.View(), width)
}

func demoTextInput(t theme.Theme, width int) string {
	campo := textinput.New()
	campo.Placeholder = "Digite o nome da skill"
	campo.SetWidth(max(8, min(40, width-4)))
	return demoFit("textinput com placeholder:\n"+campo.View()+"\n"+t.Error.Render("✗ erro: use letras minúsculas"), width)
}

func demoTextArea(t theme.Theme, width int) string {
	area := textarea.New()
	area.ShowLineNumbers = true
	area.SetWidth(max(12, min(60, width-4)))
	area.SetHeight(3)
	area.SetValue("primeira linha\nsegunda linha\nterceira linha")
	return demoFit("textarea com numeração:\n"+area.View(), width)
}

func demoViewport(t theme.Theme, width int) string {
	var conteudo []string
	for i := 1; i <= 12; i++ {
		conteudo = append(conteudo, "linha "+strconv.Itoa(i))
	}
	vp := viewport.New(viewport.WithWidth(max(8, min(40, width))), viewport.WithHeight(4))
	vp.SetContent(strings.Join(conteudo, "\n"))
	vp.SetYOffset(3)
	return demoFit(vp.View()+fmt.Sprintf("\nposição: %d%%", int(vp.ScrollPercent()*100)), width)
}

func demoPaginator(t theme.Theme, width int) string {
	paginas := paginator.New(paginator.WithPerPage(1), paginator.WithTotalPages(4))
	paginas.Type = paginator.Dots
	paginas.ActiveDot = "●"
	paginas.InactiveDot = "○"
	return demoFit("paginator em pontos:\n"+paginas.View()+"\npágina 1 de 4", width)
}

func demoHelp(t theme.Theme, width int) string {
	ajuda := help.New()
	ajuda.Styles = t.Help()
	teclas := components.DefaultKeys()
	return demoFit("ajuda curta: "+ajuda.ShortHelpView(teclas.ShortHelp())+"\najuda completa:\n"+ajuda.FullHelpView(teclas.FullHelp()), width)
}

func demoFilePicker(t theme.Theme, width int) string {
	return demoFit("filepicker lista pastas e arquivos:\npasta atual: ./internal/ui/\n  ▸ screens/\n  ▸ theme/", width)
}

func demoTimer(t theme.Theme, width int) string {
	contagem := timer.New(90 * time.Second)
	return demoFit("timer com contagem regressiva:\n"+contagem.View(), width)
}

func demoStopwatch(t theme.Theme, width int) string {
	relogio := stopwatch.New()
	return demoFit("stopwatch conta para cima:\n"+relogio.View()+"\ncrescendo após 1,2 s de execução", width)
}

func demoCursor(t theme.Theme, width int) string {
	cursorTexto := cursor.New()
	cursorTexto.SetChar("▏")
	return demoFit("Texto digitado"+cursorTexto.View(), width)
}

func demoHuhInput(t theme.Theme, width int) string {
	campo := huh.NewInput().Title("Nome do projeto").Placeholder("gex-dev-kit").WithTheme(t.Huh())
	return demoFit(campo.View(), width)
}

func demoHuhText(t theme.Theme, width int) string {
	var comando string
	return demoFit(huh.NewText().Title("Comando do pre-commit").Placeholder("go test ./...").Value(&comando).WithTheme(t.Huh()).View(), width)
}

func demoHuhSelect(t theme.Theme, width int) string {
	var destino string
	campo := huh.NewSelect[string]().Title("Escolha o destino").Options(huh.NewOptions("Global (HOME)", "Repositório")...).Value(&destino)
	return demoFit(campo.WithTheme(t.Huh()).View(), width)
}

func demoHuhMultiSelect(t theme.Theme, width int) string {
	var escolhidas []string
	campo := huh.NewMultiSelect[string]().Title("Especializações").Options(huh.NewOptions("gex-pr-review", "gex-local-spark-tests")...).Value(&escolhidas)
	return demoFit(campo.WithTheme(t.Huh()).View(), width)
}

func demoHuhConfirm(t theme.Theme, width int) string {
	var atualizar bool
	campo := huh.NewConfirm().Title("Atualizar os agentes agora?").Affirmative("Atualizar").Negative("Agora não").Value(&atualizar)
	return demoFit(campo.WithTheme(t.Huh()).View(), width)
}

func demoHuhNote(t theme.Theme, width int) string {
	nota := huh.NewNote().Title("Core obrigatório").Description("O Core entra em toda instalação.")
	return demoFit(nota.WithTheme(t.Huh()).View(), width)
}

func demoHuhFilePicker(t theme.Theme, width int) string {
	var pasta string
	campo := huh.NewFilePicker().Title("Escolha a pasta").Value(&pasta)
	return demoFit(campo.WithTheme(t.Huh()).View(), width)
}

func demoHuhGroups(t theme.Theme, width int) string {
	var destino, pasta string
	primeiro := huh.NewGroup(huh.NewSelect[string]().Title("Escolha o destino").Options(huh.NewOptions("Global (HOME)", "Repositório")...).Value(&destino))
	segundo := huh.NewGroup(huh.NewInput().Title("Nome da pasta").Value(&pasta))
	formulario := huh.NewForm(primeiro, segundo).WithTheme(t.Huh()).WithShowHelp(false)
	formulario.Init()
	return demoFit("página 1/2 do formulário\n"+formulario.View(), width)
}

func demoHuhValidation(t theme.Theme, width int) string {
	var caminho string
	campo := huh.NewInput().Title("Caminho do repositório").Description("Informe um caminho absoluto.").Validate(demoCaminhoAbsoluto).Value(&caminho)
	erro := demoCaminhoAbsoluto("relativo/pasta")
	return demoFit(campo.WithTheme(t.Huh()).View()+"\n"+t.Error.Render("erro: "+erro.Error()), width)
}

func demoHuhTheme(t theme.Theme, width int) string {
	nota := huh.NewNote().Title("Tema GEX").Description("As cores seguem o tema claro ou escuro do terminal.")
	return demoFit(nota.WithTheme(t.Huh()).View(), width)
}

func demoHuhAccessible(t theme.Theme, width int) string {
	var nome string
	var saida bytes.Buffer
	formulario := huh.NewForm(huh.NewGroup(huh.NewInput().Title("Nome do projeto").Value(&nome))).
		WithAccessible(true).
		WithInput(strings.NewReader("gex-dev-kit\n")).
		WithOutput(&saida)
	if err := formulario.Run(); err != nil {
		return demoFit(err.Error(), width)
	}
	return demoFit(saida.String()+"Resposta: "+nome, width)
}

func demoJoin(t theme.Theme, width int) string {
	meia := max(8, (width-3)/2)
	esquerda := components.Panel(t, components.PanelOptions{Title: "A", Width: meia, Border: components.BorderRounded}, "coluna um")
	direita := components.Panel(t, components.PanelOptions{Title: "B", Width: meia, Border: components.BorderRounded}, "coluna dois")
	separador := strings.TrimSuffix(strings.Repeat(" │ \n", lipgloss.Height(esquerda)), "\n")
	lado := lipgloss.JoinHorizontal(lipgloss.Top, esquerda, separador, direita)
	return demoFit("JoinHorizontal e JoinVertical:\n"+lado+"\n"+lipgloss.JoinVertical(lipgloss.Left, "linha 1", "linha 2"), width)
}

func demoPlace(t theme.Theme, width int) string {
	caixa := components.Panel(t, components.PanelOptions{Title: "centro", Width: 20, Border: components.BorderNormal}, "caixa")
	return demoFit("cantos: ┌ ┐ └ ┘ (normal)\n"+lipgloss.Place(min(width, 40), 5, lipgloss.Center, lipgloss.Center, caixa), width)
}

func demoBorders(t theme.Theme, width int) string {
	tipos := []components.BorderKind{components.BorderRounded, components.BorderNormal, components.BorderThick, components.BorderDouble, components.BorderBlock, components.BorderASCII}
	nomes := []string{"arredondada", "normal", "grossa", "dupla", "bloco", "ASCII"}
	var celulas []string
	for i, tipo := range tipos {
		celulas = append(celulas, components.Panel(t, components.PanelOptions{Title: nomes[i], Width: 18, Border: tipo}, "canto"))
	}
	return demoFit("cantos: ╭ arredondada, ┏ grossa, ╔ dupla, █ bloco, + ASCII\n"+demoGrade(celulas, 18, width), width)
}

func demoGradient(t theme.Theme, width int) string {
	var faixa strings.Builder
	for _, cor := range t.Blend(40) {
		faixa.WriteString(t.Text.Foreground(cor).Render("█"))
	}
	return demoFit("gradiente com lipgloss.Blend1D:\n"+faixa.String(), width)
}

func demoLipglossTable(t theme.Theme, width int) string {
	linhas := [][]string{{"pre-commit", "ok"}, {"pre-push", "aviso"}}
	return demoFit(components.StaticTable(t, width, []string{"Hook", "Estado"}, linhas), width)
}

func demoLipglossList(t theme.Theme, width int) string {
	passos := lglist.New("primeiro passo", "segundo passo").Enumerator(lglist.Arabic)
	return demoFit(passos.String(), width)
}

func demoTree(t theme.Theme, width int) string {
	arvore := lgtree.Root("internal").Child(lgtree.Root("ui").Child("theme", "components"), "app")
	return demoFit(arvore.String(), width)
}

func demoAdaptive(t theme.Theme, width int) string {
	ajuste := lipgloss.LightDark(t.Dark)
	cor := ajuste(t.Palette.Navy, t.Palette.Text)
	return demoFit("lipgloss.LightDark escolhe a cor pelo fundo:\nmodo claro: texto escuro\nmodo escuro: texto claro\n"+t.Text.Foreground(cor).Render("amostra"), width)
}

func demoMeasures(t theme.Theme, width int) string {
	amostras := []string{"gex", "é", "e" + string(rune(0x0301)), "漢字", "🚀 deploy"}
	var linhas [][]string
	for _, s := range amostras {
		linhas = append(linhas, []string{s, strconv.Itoa(len(s)), strconv.Itoa(ansi.StringWidth(s)), strconv.Itoa(lipgloss.Width(s)), ansi.Truncate(s, 6, "…")})
	}
	cabecalho := []string{"texto", "len", "ansi.StringWidth", "lipgloss.Width", "corta 6"}
	return demoFit(components.StaticTable(t, width, cabecalho, linhas), width)
}

func demoSpring(t theme.Theme, width int) string {
	molas := []struct {
		nome          string
		frequencia    float64
		amortecimento float64
	}{{"Snappy", 8, 0.4}, {"Gentle", 3, 0.9}}
	var linhas []string
	for _, m := range molas {
		mola := harmonica.NewSpring(harmonica.FPS(30), m.frequencia, m.amortecimento)
		pos, vel := 0.0, 0.0
		for quadro := 1; quadro <= 15; quadro++ {
			pos, vel = mola.Update(pos, vel, 1)
			if quadro == 3 || quadro == 15 {
				linhas = append(linhas, fmt.Sprintf("%-8s quadro %2d %s", m.nome, quadro, demoBarra(int(pos*20+0.5), 20)))
			}
		}
	}
	linhas = append(linhas, fmt.Sprintf("%-8s quadro  1 %s", "desligada", demoBarra(20, 20)))
	return demoFit(strings.Join(linhas, "\n"), width)
}

func demoProjectile(t theme.Theme, width int) string {
	const colunas, linhas = 30, 6
	grade := make([][]rune, linhas)
	for i := range grade {
		grade[i] = []rune(strings.Repeat(" ", colunas))
	}
	gerador := rand.New(rand.NewSource(7))
	for i := 0; i < 5; i++ {
		velocidade := harmonica.Vector{X: 4 + gerador.Float64()*4, Y: 6 + gerador.Float64()*4}
		projetil := harmonica.NewProjectile(harmonica.FPS(30), harmonica.Point{X: 0, Y: 0}, velocidade, harmonica.Vector{X: 0, Y: -9.8})
		var ponto harmonica.Point
		for passo := 0; passo < 12; passo++ {
			ponto = projetil.Update()
		}
		x := max(0, min(colunas-1, int(ponto.X*4)))
		y := max(0, min(linhas-1, linhas-1-int(ponto.Y*2)))
		grade[y][x] = '*'
	}
	var desenho []string
	for _, linha := range grade {
		desenho = append(desenho, string(linha))
	}
	return demoFit("harmonica.NewProjectile, 5 partículas, semente 7:\n"+strings.Join(desenho, "\n"), width)
}

func demoAnsi(t theme.Theme, width int) string {
	bruto := t.Accent.Render("gex-dev-kit")
	return demoFit("ansi.Strip remove as cores: "+ansi.Strip(bruto)+"\nansi.StringWidth: "+strconv.Itoa(ansi.StringWidth(bruto))+"\nansi.Truncate(…, 6): "+ansi.Truncate(bruto, 6, "…"), width)
}

func demoTerm(t theme.Theme, width int) string {
	tamanho := "desconhecido"
	if colunas, linhas, err := term.GetSize(os.Stdout.Fd()); err == nil {
		tamanho = fmt.Sprintf("%d×%d", colunas, linhas)
	}
	return demoFit("stdin TTY: "+demoSimNao(term.IsTerminal(os.Stdin.Fd()))+"\nstdout TTY: "+demoSimNao(term.IsTerminal(os.Stdout.Fd()))+"\ntamanho: "+tamanho, width)
}

func demoTick(t theme.Theme, width int) string {
	return demoFit("tea.Tick agenda o próximo quadro da mola.\nquadro 1: "+demoBarra(3, 12)+"\nquadro 2: "+demoBarra(7, 12)+"\nCada quadro chega como motion.FrameMsg.", width)
}

func demoWindowSize(t theme.Theme, width int) string {
	return demoFit("tea.WindowSizeMsg recebido na vitrine:\ntamanho atual: 80×24\nlargura usada agora: "+strconv.Itoa(width), width)
}

func demoAltScreen(t theme.Theme, width int) string {
	return demoFit("tea.NewView liga AltScreen na prévia.\nmodo: alt screen (a alterna para inline)\nNa gravação, a tela volta ao modo inline.", width)
}

func demoMouse(t theme.Theme, width int) string {
	return demoFit("tea.MouseModeCellMotion captura cliques e movimento.\núltimo evento de mouse: clique esquerdo em 12, 4\ncontador de cliques: 3", width)
}

func demoFocus(t theme.Theme, width int) string {
	return demoFit("tea.ReportFocus avisa quando a janela ganha ou perde foco.\njanela com foco: sim\ntea.BlurMsg ao sair: janela com foco: não", width)
}

func demoBatchSequence(t theme.Theme, width int) string {
	return demoFit("tea.Batch: 3 ticks com atrasos de 300, 100 e 200 ms\nBatch: chegada na ordem 2, 3, 1\ntea.Sequence: ordem fixa 1, 2, 3", width)
}

func demoPrintln(t theme.Theme, width int) string {
	return demoFit("tea.Println imprime fora do alt screen.\nlinhas impressas: 1 acima da área viva", width)
}

func demoSuspend(t theme.Theme, width int) string {
	estado := "ativo"
	if !components.DefaultKeys().Suspend.Enabled() {
		estado = "desligado"
	}
	return demoFit("ctrl+z suspende (indisponível no Windows)\nbinding: "+estado, width)
}

func demoTypedErrors(t theme.Theme, width int) string {
	return demoFit("errMsg carrega um erro tipado do app.Problem:\n"+components.ErrorBox(t, width, "Erro tipado", "causa do erro", "troque o caminho e tente de novo"), width)
}
