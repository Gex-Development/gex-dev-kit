package screens

import (
	"bytes"
	"image/color"
	"os"
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/exp/teatest/v2"

	"github.com/Gex-Development/gex-dev-kit/internal/catalog"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/motion"
)

func catalogTestCommon(motionOn bool) Common {
	return Common{Version: "v0.3.0", Home: "/home/demo", UserHome: "/home/demo", Dark: true, Profile: colorprofile.ASCII, Motion: motionOn}
}

func catalogSized(motionOn bool, width, height int) catalogModel {
	next, _ := newCatalogModel(catalogTestCommon(motionOn)).Update(tea.WindowSizeMsg{Width: width, Height: height})
	return next.(catalogModel)
}

func catalogKey(s string) tea.KeyPressMsg {
	switch s {
	case "tab":
		return tea.KeyPressMsg{Code: tea.KeyTab}
	case "shift+tab":
		return tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEscape}
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "ctrl+c":
		return tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}
	}
	r := []rune(s)[0]
	return tea.KeyPressMsg{Code: r, Text: s}
}

func catalogSend(m catalogModel, s string) catalogModel {
	next, cmd := m.Update(catalogKey(s))
	return catalogDrain(next.(catalogModel), cmd)
}

func catalogDrain(m catalogModel, cmd tea.Cmd) catalogModel {
	if cmd == nil {
		return m
	}
	switch msg := cmd().(type) {
	case tea.BatchMsg:
		for _, next := range msg {
			m = catalogDrain(m, next)
		}
	case list.FilterMatchesMsg:
		next, _ := m.Update(msg)
		m = next.(catalogModel)
	}
	return m
}

func TestCatalogCabecalhoMostraContagemDeSkills(t *testing.T) {
	view := catalogSized(false, 80, 24).View().Content
	for _, want := range []string{"7 Core", "18 Matt", "7 especializações"} {
		if !strings.Contains(view, want) {
			t.Fatalf("cabeçalho sem %q:\n%s", want, view)
		}
	}
}

func TestCatalogSkillsTemUmaEntradaPorSkill(t *testing.T) {
	items := catalogSkillItems()
	if len(items) != 32 {
		t.Fatalf("lista de skills tem %d entradas, esperado 32", len(items))
	}
	counts := map[string]int{}
	for _, item := range items {
		counts[strings.SplitN(item.(catalogItem).badge, " ·", 2)[0]]++
	}
	if counts["Core"] != 7 || counts["Matt"] != 18 || counts["Especialização"] != 7 {
		t.Fatalf("contagem por grupo errada: %v", counts)
	}
}

func TestCatalogFiltroPorLetrasVaiParaOCampo(t *testing.T) {
	m := catalogSized(false, 80, 24)
	m = catalogSend(m, "/")
	for _, r := range "gex-pr" {
		m = catalogSend(m, string(r))
	}
	visible := m.skills.VisibleItems()
	if len(visible) != 1 {
		t.Fatalf("filtro gex-pr deixou %d entradas, esperado 1", len(visible))
	}
	if name := visible[0].(catalogItem).name; name != "gex-pr-review" {
		t.Fatalf("filtro gex-pr deixou %q, esperado gex-pr-review", name)
	}
	if m.tab != catalogSkillsTab {
		t.Fatalf("letras digitadas no filtro trocaram de aba")
	}
}

func catalogRankIndexes(ranks []list.Rank) []int {
	var got []int
	for _, rank := range ranks {
		got = append(got, rank.Index)
	}
	return got
}

func TestCatalogFiltroCasaInicioDePalavra(t *testing.T) {
	got := catalogRankIndexes(catalogFilter("PR", []string{"gex-pr-review", "gex-preparar-pr", "xpr", "pr Matt"}))
	if len(got) != 3 || got[0] != 0 || got[1] != 1 || got[2] != 3 {
		t.Fatalf("filtro pr casou %v, esperado [0 1 3]", got)
	}
}

func TestCatalogFiltroPrefereOPalavraInteira(t *testing.T) {
	targets := []string{"gex-pr-review", "gex-preparar-pr"}
	if got := catalogRankIndexes(catalogFilter("gex-pr", targets)); len(got) != 1 || got[0] != 0 {
		t.Fatalf("filtro gex-pr casou %v, esperado só a palavra inteira [0]", got)
	}
	if got := catalogRankIndexes(catalogFilter("gex-p", targets)); len(got) != 2 {
		t.Fatalf("sem palavra inteira, filtro gex-p deveria cair no prefixo: %v", got)
	}
}

func TestCatalogQNoFiltroNaoSai(t *testing.T) {
	m := catalogSized(false, 80, 24)
	m = catalogSend(m, "/")
	next, cmd := m.Update(catalogKey("q"))
	if cmd != nil {
		if _, ok := cmd().(tea.QuitMsg); ok {
			t.Fatalf("q dentro do filtro encerrou a tela")
		}
	}
	if !next.(catalogModel).skills.SettingFilter() {
		t.Fatalf("q dentro do filtro saiu do modo de filtro")
	}
}

func TestCatalogTabTrocaParaHooksERefs(t *testing.T) {
	m := catalogSized(false, 80, 24)
	m = catalogSend(m, "tab")
	if m.tab != catalogHooksTab {
		t.Fatalf("tab não foi para Hooks")
	}
	view := m.View().Content
	for _, want := range []string{"gex-sensitive-action-warning", "Codex: não suportado"} {
		if !strings.Contains(view, want) {
			t.Fatalf("aba Hooks sem %q:\n%s", want, view)
		}
	}
	m = catalogSend(m, "tab")
	m = catalogSend(m, "/")
	for _, r := range "smart-commit" {
		m = catalogSend(m, string(r))
	}
	m = catalogSend(m, "enter")
	if m.tab != catalogRefsTab {
		t.Fatalf("tab não foi para Referências")
	}
	visible := m.refs.VisibleItems()
	if len(visible) != 1 || visible[0].(catalogItem).name != "claude/smart-commit" {
		t.Fatalf("filtro smart-commit em Referências deixou %d entradas", len(visible))
	}
	m = catalogSend(m, "shift+tab")
	if m.tab != catalogHooksTab {
		t.Fatalf("shift+tab não voltou para Hooks")
	}
}

func TestCatalogHookInicialEOPrimeiroComLimitacao(t *testing.T) {
	hooks := catalog.HookSpecs()
	m := catalogSized(false, 80, 24)
	if got := hooks[m.hookCursor].Name; got != "gex-sensitive-action-warning" {
		t.Fatalf("cursor inicial em %q, esperado gex-sensitive-action-warning", got)
	}
}

func TestCatalogHooksSobeEDesceComCursor(t *testing.T) {
	m := catalogSized(false, 80, 24)
	m = catalogSend(m, "tab")
	m = catalogSend(m, "j")
	if got := m.hooks[m.hookCursor].Name; got != "gex-pre-pr" {
		t.Fatalf("j levou o cursor para %q, esperado gex-pre-pr", got)
	}
	m = catalogSend(m, "k")
	m = catalogSend(m, "k")
	if m.hookCursor != 0 {
		t.Fatalf("k acima do primeiro hook deixou o cursor em %d", m.hookCursor)
	}
}

func TestCatalogMolaSemMovimentoNaoAgendaQuadros(t *testing.T) {
	m := catalogSized(false, 80, 24)
	next, cmd := m.Update(catalogKey("tab"))
	if cmd != nil {
		t.Fatalf("tab com Motion=false agendou quadro")
	}
	if next.(catalogModel).tab != catalogHooksTab {
		t.Fatalf("tab com Motion=false não trocou de aba")
	}
}

func TestCatalogMolaComMovimentoAcomodaEmQuadros(t *testing.T) {
	next, cmd := catalogSized(true, 80, 24).Update(catalogKey("tab"))
	m := next.(catalogModel)
	if cmd == nil {
		t.Fatalf("tab com Motion=true não agendou quadro")
	}
	for i := 0; i < 200 && !m.highlight.Settled(); i++ {
		next, _ = m.Update(motion.FrameMsg{ID: m.frame})
		m = next.(catalogModel)
	}
	if !m.highlight.Settled() || m.highlight.Value() != 1 {
		t.Fatalf("mola não assentou em Hooks: valor %v", m.highlight.Value())
	}
}

func TestCatalogRespeitaLarguraEAlturaDoTerminal(t *testing.T) {
	sizes := []struct{ width, height int }{{80, 24}, {60, 20}, {40, 12}}
	steps := []string{"", "tab", "tab", "?", "?", "tab"}
	for _, size := range sizes {
		m := catalogSized(false, size.width, size.height)
		for _, step := range steps {
			if step != "" {
				next, _ := m.Update(catalogKey(step))
				m = next.(catalogModel)
			}
			view := m.View().Content
			if lipgloss.Width(view) > size.width || lipgloss.Height(view) > size.height {
				t.Fatalf("tela %dx%d passou do limite após %q: largura %d, altura %d",
					size.width, size.height, step, lipgloss.Width(view), lipgloss.Height(view))
			}
		}
	}
}

func TestCatalogClassTinyAceitaSoQ(t *testing.T) {
	m := catalogSized(false, 30, 10)
	next, cmd := m.Update(catalogKey("tab"))
	if cmd != nil || next.(catalogModel).tab != catalogSkillsTab {
		t.Fatalf("tela tiny reagiu a tab")
	}
	if !strings.Contains(m.View().Content, "Terminal pequeno") {
		t.Fatalf("tela tiny sem aviso de tamanho")
	}
	_, cmd = m.Update(catalogKey("q"))
	if cmd == nil {
		t.Fatalf("q na tela tiny não saiu")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("q na tela tiny não devolveu tea.Quit")
	}
}

func TestCatalogCtrlCSaiSemTeaInterrupt(t *testing.T) {
	next, cmd := catalogSized(false, 80, 24).Update(catalogKey("ctrl+c"))
	if cmd == nil {
		t.Fatalf("ctrl+c não devolveu comando de saída")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("ctrl+c não devolveu tea.Quit")
	}
	err := catalogExitError(next, nil)
	if code, ok := ExitCode(err); !ok || code != 130 {
		t.Fatalf("ctrl+c deveria virar saída 130, veio %v", err)
	}
}

func TestCatalogSemInterrupcaoDevolveNil(t *testing.T) {
	if err := catalogExitError(catalogSized(false, 80, 24), nil); err != nil {
		t.Fatalf("saída normal devolveu erro: %v", err)
	}
	if err := catalogExitError(nil, os.ErrClosed); err != os.ErrClosed {
		t.Fatalf("erro do runner não foi devolvido: %v", err)
	}
}

func TestCatalogEscSaiQuandoNaoHaFiltro(t *testing.T) {
	_, cmd := catalogSized(false, 80, 24).Update(catalogKey("esc"))
	if cmd == nil {
		t.Fatalf("esc sem filtro não saiu")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("esc sem filtro não devolveu tea.Quit")
	}
}

func TestCatalogAjudaAlternaComInterrogacao(t *testing.T) {
	m := catalogSized(false, 80, 24)
	m = catalogSend(m, "?")
	if !m.help || !strings.Contains(m.View().Content, "ajuda") {
		t.Fatalf("? não abriu a ajuda")
	}
	m = catalogSend(m, "?")
	if m.help {
		t.Fatalf("? não fechou a ajuda")
	}
}

func TestCatalogSublinhadoAcompanhaAMola(t *testing.T) {
	starts := []int{0, 10, 20}
	widths := []int{6, 8, 12}
	cases := []struct {
		pos         float64
		start, size int
	}{{0, 0, 6}, {1, 10, 8}, {0.5, 5, 7}, {2, 20, 12}}
	for _, c := range cases {
		start, size := catalogUnderline(c.pos, starts, widths)
		if start != c.start || size != c.size {
			t.Fatalf("posição %v: sublinhado em %d/%d, esperado %d/%d", c.pos, start, size, c.start, c.size)
		}
	}
}

func TestCatalogTemaSegueFundoDoTerminal(t *testing.T) {
	next, _ := catalogSized(false, 80, 24).Update(tea.BackgroundColorMsg{Color: color.White})
	if next.(catalogModel).c.Dark {
		t.Fatalf("fundo claro não trocou o tema para claro")
	}
}

func TestCatalogFluxoComTeatest(t *testing.T) {
	tm := teatest.NewTestModel(t, newCatalogModel(catalogTestCommon(false)),
		teatest.WithInitialTermSize(80, 24),
		teatest.WithProgramOptions(tea.WithColorProfile(colorprofile.ASCII), tea.WithEnvironment([]string{"TERM=xterm-256color"})),
	)
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("7 Core"))
	}, teatest.WithDuration(3*time.Second))
	tm.Send(catalogKey("tab"))
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Codex: não suportado"))
	}, teatest.WithDuration(3*time.Second))
	tm.Send(catalogKey("ctrl+c"))
	final := tm.FinalModel(t, teatest.WithFinalTimeout(5*time.Second))
	if !final.(catalogModel).interrupted {
		t.Fatalf("ctrl+c no teatest não marcou interrupção")
	}
}
