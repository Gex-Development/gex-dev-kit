package screens

import (
	"bytes"
	"image/color"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/teatest/v2"

	"github.com/Gex-Development/gex-dev-kit/internal/app"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/motion"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

func homeTestCommon() Common {
	return Common{
		Version:  "0.3.0",
		Home:     "/home/demo",
		UserHome: "/home/demo",
		Dark:     true,
		Profile:  colorprofile.ANSI,
	}
}

func homeFakeDeps(core app.CoreStatus, agents []app.AgentChange) HomeDeps {
	return HomeDeps{
		Core:   func() (app.CoreStatus, error) { return core, nil },
		Agents: func() ([]app.AgentChange, error) { return agents, nil },
	}
}

func homeSend(m tea.Model, msgs ...tea.Msg) tea.Model {
	for _, msg := range msgs {
		m, _ = m.Update(msg)
	}
	return m
}

func homeEnter(t *testing.T, m tea.Model) tea.Model {
	t.Helper()
	next, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatalf("Enter no menu deveria produzir uma escolha")
	}
	return homeSend(next, cmd())
}

func homeCollect(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}
	msg := cmd()
	batch, ok := msg.(tea.BatchMsg)
	if !ok {
		return []tea.Msg{msg}
	}
	var out []tea.Msg
	for _, inner := range batch {
		out = append(out, homeCollect(inner)...)
	}
	return out
}

func TestHomeSelectsDoctorWithoutInstalling(t *testing.T) {
	c := homeTestCommon()
	c.Home = t.TempDir()
	model := homeSend(newHomeModel(c, homeFakeDeps(app.CoreStatus{}, nil)),
		tea.WindowSizeMsg{Width: 80, Height: 24},
		tea.KeyPressMsg{Code: tea.KeyDown},
		tea.KeyPressMsg{Code: tea.KeyDown},
	)
	final := homeEnter(t, model)
	if choice := final.(homeModel).choice; choice != ChoiceDoctor {
		t.Fatalf("item Doctor deveria abrir doctor, recebeu %q", choice)
	}
	if _, err := os.Stat(c.Home + "/.claude/CLAUDE.md"); !os.IsNotExist(err) {
		t.Fatalf("navegar não pode instalar Core: %v", err)
	}
}

func TestHomeMenuAjudaOpensHelpWithoutLeaving(t *testing.T) {
	model := homeSized(newHomeModel(homeTestCommon(), homeFakeDeps(app.CoreStatus{}, nil)), 80, 24)
	for range 8 {
		model = homeSend(model, tea.KeyPressMsg{Code: tea.KeyDown})
	}
	model = homeEnter(t, model)
	final := model.(homeModel)
	if final.choice != "" || !final.help {
		t.Fatalf("Ajuda deveria abrir a ajuda sem sair: escolha %q, ajuda %v", final.choice, final.help)
	}
}

func homeSized(m tea.Model, width, height int) tea.Model {
	return homeSend(m, tea.WindowSizeMsg{Width: width, Height: height})
}

func TestHomeRendersWithinNarrowTerminal(t *testing.T) {
	model := homeSized(newHomeModel(homeTestCommon(), homeFakeDeps(app.CoreStatus{}, nil)), 40, 18)
	screen := model.View().Content
	if width := lipgloss.Width(screen); width > 40 {
		t.Fatalf("painel excedeu terminal de 40 colunas: %d", width)
	}
	if height := lipgloss.Height(screen); height > 18 {
		t.Fatalf("painel excedeu terminal de 18 linhas: %d", height)
	}
}

func TestHomeKeepsActionsWithinCompactTerminal(t *testing.T) {
	cases := []struct {
		size   tea.WindowSizeMsg
		expect string
	}{
		{tea.WindowSizeMsg{Width: 40, Height: 12}, "Doctor"},
		{tea.WindowSizeMsg{Width: 40, Height: 8}, "Terminal"},
		{tea.WindowSizeMsg{Width: 14, Height: 8}, "Terminal"},
	}
	for _, tc := range cases {
		model := homeSized(newHomeModel(homeTestCommon(), homeFakeDeps(app.CoreStatus{}, nil)), tc.size.Width, tc.size.Height)
		screen := model.View().Content
		if lipgloss.Width(screen) > tc.size.Width || lipgloss.Height(screen) > tc.size.Height {
			t.Fatalf("painel ultrapassou %dx%d: %dx%d", tc.size.Width, tc.size.Height, lipgloss.Width(screen), lipgloss.Height(screen))
		}
		if !strings.Contains(screen, tc.expect) {
			t.Fatalf("%q deveria aparecer no terminal %dx%d: %s", tc.expect, tc.size.Width, tc.size.Height, screen)
		}
	}
}

func TestHomeReportsChangedCoreFromDisk(t *testing.T) {
	model := homeSend(newHomeModel(homeTestCommon(), homeFakeDeps(app.CoreStatus{}, nil)),
		tea.WindowSizeMsg{Width: 80, Height: 24},
		homeCoreMsg{status: app.CoreStatus{ActiveCore: 6, TotalCore: 7, Attention: 1, State: "alterado"}},
	)
	if view := model.View().Content; !strings.Contains(view, "alterado") {
		t.Fatalf("painel ocultou drift no Core: %s", view)
	}
}

func TestHomeHelpAndScanErrorAreVisibleAndDismissible(t *testing.T) {
	model := homeSized(newHomeModel(homeTestCommon(), homeFakeDeps(app.CoreStatus{}, nil)), 40, 14)
	model = homeSend(model, tea.KeyPressMsg{Code: '?', Text: "?"})
	if view := model.View().Content; !strings.Contains(view, "ATALHOS") || lipgloss.Width(view) > 40 || lipgloss.Height(view) > 14 {
		t.Fatalf("ajuda deve caber no terminal: %q", view)
	}
	model = homeSend(model, tea.KeyPressMsg{Code: tea.KeyEscape}, homeCoreMsg{err: os.ErrPermission})
	if view := model.View().Content; !strings.Contains(view, "permission denied") {
		t.Fatalf("erro da leitura precisa ser visível: %q", view)
	}
	model = homeSend(model, tea.KeyPressMsg{Code: tea.KeyEscape})
	if view := model.View().Content; strings.Contains(view, "permission denied") {
		t.Fatalf("esc deveria dispensar o erro: %q", view)
	}
}

func TestHomeReportsSevenCoreSkillsWithoutClaimingInstalled(t *testing.T) {
	model := homeSend(newHomeModel(homeTestCommon(), homeFakeDeps(app.CoreStatus{}, nil)),
		tea.WindowSizeMsg{Width: 80, Height: 24},
		homeCoreMsg{status: app.CoreStatus{TotalCore: 7, State: "não instalado"}},
	)
	if view := model.View().Content; !strings.Contains(view, "7 skills") || !strings.Contains(view, "não instalado") {
		t.Fatalf("contagem/estado incorretos: %q", view)
	}
}

func TestHomeDoesNotCallLegacyInstallCompleteWithoutSevenSkills(t *testing.T) {
	model := homeSend(newHomeModel(homeTestCommon(), homeFakeDeps(app.CoreStatus{}, nil)),
		tea.WindowSizeMsg{Width: 80, Height: 24},
		homeCoreMsg{status: app.CoreStatus{Checked: 1, TotalCore: 7, State: "atualização disponível"}},
	)
	if view := model.View().Content; strings.Contains(view, "íntegro") || !strings.Contains(view, "0/7") {
		t.Fatalf("instalação antiga não contém sete skills: %q", view)
	}
}

func TestHomeFlowFinishesChecklistAndChoosesInstall(t *testing.T) {
	c := homeTestCommon()
	c.Motion = true
	agents := []app.AgentChange{{Name: "claude", Current: "2.1.306", Target: "2.1.306"}, {Name: "codex", Target: "0.163.0"}}
	tm := teatest.NewTestModel(t, newHomeModel(c, homeFakeDeps(app.CoreStatus{TotalCore: 7, State: "íntegro"}, agents)),
		teatest.WithInitialTermSize(80, 24),
		teatest.WithProgramOptions(tea.WithColorProfile(colorprofile.ASCII), tea.WithEnvironment([]string{"TERM=xterm-256color"})),
	)
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("2.1.306"))
	}, teatest.WithDuration(5*time.Second))
	tm.Send(tea.KeyPressMsg{Code: tea.KeyEnter})
	final := tm.FinalModel(t, teatest.WithFinalTimeout(5*time.Second))
	if choice := final.(homeModel).choice; choice != ChoiceInstall {
		t.Fatalf("Enter no primeiro item deveria escolher install, recebeu %q", choice)
	}
}

func TestHomeMotionOffSchedulesNoFrames(t *testing.T) {
	c := homeTestCommon()
	c.Motion = false
	m := newHomeModel(c, homeFakeDeps(app.CoreStatus{TotalCore: 7, State: "íntegro"}, nil))
	cmds := []tea.Cmd{m.Init()}
	steps := []tea.Msg{
		tea.WindowSizeMsg{Width: 80, Height: 24},
		homeCoreMsg{status: app.CoreStatus{TotalCore: 7, State: "íntegro"}},
		homeAgentsMsg{},
		tea.KeyPressMsg{Code: tea.KeyDown},
		tea.BlurMsg{},
		tea.FocusMsg{},
	}
	var model tea.Model = m
	for _, msg := range steps {
		next, cmd := model.Update(msg)
		model = next
		cmds = append(cmds, cmd)
	}
	for _, cmd := range cmds {
		for _, msg := range homeCollect(cmd) {
			if _, frame := msg.(motion.FrameMsg); frame {
				t.Fatalf("com movimento desligado nenhum quadro pode ser agendado")
			}
		}
	}
}

func TestHomeMotionOnSchedulesFrames(t *testing.T) {
	c := homeTestCommon()
	c.Motion = true
	_, cmd := newHomeModel(c, homeFakeDeps(app.CoreStatus{}, nil)).Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	found := false
	for _, msg := range homeCollect(cmd) {
		if frame, ok := msg.(motion.FrameMsg); ok && frame.ID == homeFrameLoop {
			found = true
		}
	}
	if !found {
		t.Fatalf("com movimento ligado a entrada deveria agendar o quadro da home")
	}
}

func TestHomeFocusBlurPausesAndResumesFrames(t *testing.T) {
	c := homeTestCommon()
	c.Motion = true
	model := homeSized(newHomeModel(c, homeFakeDeps(app.CoreStatus{}, nil)), 80, 24)
	model = homeSend(model, tea.BlurMsg{})
	next, cmd := model.Update(motion.FrameMsg{ID: homeFrameLoop})
	if cmd != nil {
		t.Fatalf("sem foco não pode haver novo quadro: %v", homeCollect(cmd))
	}
	_, cmd = next.Update(tea.FocusMsg{})
	if cmd == nil {
		t.Fatalf("ao voltar o foco a animação deveria retomar")
	}
}

func TestHomeTinyTerminalIgnoresEnter(t *testing.T) {
	model := homeSized(newHomeModel(homeTestCommon(), homeFakeDeps(app.CoreStatus{}, nil)), 20, 5)
	next, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd != nil || next.(homeModel).choice != "" {
		t.Fatalf("terminal pequeno não pode decidir com Enter")
	}
	_, cmd = next.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	if cmd == nil {
		t.Fatalf("q deve continuar saindo no terminal pequeno")
	}
}

func TestHomeCtrlCMarksInterruptedAndQuits(t *testing.T) {
	model := homeSized(newHomeModel(homeTestCommon(), homeFakeDeps(app.CoreStatus{}, nil)), 80, 24)
	next, cmd := model.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	if cmd == nil {
		t.Fatalf("ctrl+c deveria pedir a saída")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("ctrl+c deve devolver tea.Quit, não tea.Interrupt")
	}
	if !next.(homeModel).interrupted {
		t.Fatalf("ctrl+c deveria marcar a sessão como interrompida")
	}
}

func TestHomeCtrlZSuspendsOutsideWindows(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("ctrl+z não existe no Windows")
	}
	model := homeSized(newHomeModel(homeTestCommon(), homeFakeDeps(app.CoreStatus{}, nil)), 80, 24)
	_, cmd := model.Update(tea.KeyPressMsg{Code: 'z', Mod: tea.ModCtrl})
	if cmd == nil {
		t.Fatalf("ctrl+z deveria suspender")
	}
	if _, ok := cmd().(tea.SuspendMsg); !ok {
		t.Fatalf("ctrl+z deve devolver tea.Suspend")
	}
}

func TestHomeBackgroundColorRebuildsTheme(t *testing.T) {
	model := newHomeModel(homeTestCommon(), homeFakeDeps(app.CoreStatus{}, nil))
	if !model.theme.Dark {
		t.Fatalf("tema inicial deveria ser escuro")
	}
	next, _ := model.Update(tea.BackgroundColorMsg{Color: color.White})
	if next.(homeModel).theme.Dark {
		t.Fatalf("fundo claro deveria recriar o tema claro")
	}
}

func TestHomeMouseClickSelectsMenuRow(t *testing.T) {
	c := homeTestCommon()
	c.Motion = false
	model := homeSized(newHomeModel(c, homeFakeDeps(app.CoreStatus{}, nil)), 80, 24)
	clickRow := homeMenuTopRow(t, model) + 2
	model = homeSend(model, tea.MouseClickMsg{X: 0, Y: clickRow, Button: tea.MouseLeft})
	if id := model.(homeModel).menu.Selected().ID; id != string(ChoiceDoctor) {
		t.Fatalf("clique na linha do Doctor selecionou %q", id)
	}
}

func homeMenuTopRow(t *testing.T, m tea.Model) int {
	t.Helper()
	lines := strings.Split(m.View().Content, "\n")
	for i, line := range lines {
		if strings.Contains(ansi.Strip(line), "Instalar") {
			return i
		}
	}
	t.Fatalf("linha do menu não encontrada: %q", m.View().Content)
	return 0
}

func TestHomePlainProfileIsReadableWithoutColor(t *testing.T) {
	c := homeTestCommon()
	c.Profile = colorprofile.ASCII
	c.Motion = false
	plain := ansi.Strip(homeSized(newHomeModel(c, homeFakeDeps(app.CoreStatus{}, nil)), 80, 24).View().Content)
	for _, want := range []string{"Instalar", "0.3.0", "· ~"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("sem cor a tela deveria mostrar %q: %s", want, plain)
		}
	}
}

func TestHomeHasColorWithTrueColor(t *testing.T) {
	c := homeTestCommon()
	c.Profile = colorprofile.TrueColor
	view := homeSized(newHomeModel(c, homeFakeDeps(app.CoreStatus{}, nil)), 80, 24).View().Content
	if !strings.Contains(view, string(rune(27))+"[") {
		t.Fatalf("perfil TrueColor deveria produzir cor")
	}
}

func TestHomeChecklistShowsAgentVersionsAndMissingInstall(t *testing.T) {
	agents := []app.AgentChange{{Name: "claude", Current: "2.1.306", Target: "2.1.306"}, {Name: "codex", Target: "0.163.0"}}
	model := homeSend(newHomeModel(homeTestCommon(), homeFakeDeps(app.CoreStatus{}, nil)),
		tea.WindowSizeMsg{Width: 80, Height: 24},
		homeAgentsMsg{changes: agents},
	)
	view := model.View().Content
	if !strings.Contains(view, "2.1.306") || !strings.Contains(view, "não instalado pelo kit") {
		t.Fatalf("checklist deveria mostrar versão e instalação ausente: %q", view)
	}
}

func TestHomeAgentTextAndStateCoverEveryCase(t *testing.T) {
	cases := []struct {
		name   string
		change app.AgentChange
		text   string
		state  theme.State
	}{
		{"instalado na versão alvo", app.AgentChange{Current: "2.1.306", Target: "2.1.306"}, "2.1.306", theme.StateOK},
		{"versão antiga", app.AgentChange{Current: "2.0.0", Target: "2.1.306"}, "2.0.0 → 2.1.306", theme.StateWarn},
		{"sem registro", app.AgentChange{Target: "2.1.306"}, "não instalado pelo kit", theme.StateWarn},
		{"bloqueado", app.AgentChange{Current: "2.1.306", Target: "2.1.306", Blocked: true, Reason: "PATH ambíguo"}, "PATH ambíguo", theme.StateWarn},
	}
	for _, tc := range cases {
		if text := homeAgentText(tc.change); text != tc.text {
			t.Fatalf("%s: texto %q, esperado %q", tc.name, text, tc.text)
		}
		if state := homeAgentState(tc.change); state != tc.state {
			t.Fatalf("%s: estado %v, esperado %v", tc.name, state, tc.state)
		}
	}
}

func TestHomeCoreStateMapsDriftAndNotInstalled(t *testing.T) {
	cases := map[string]theme.State{
		"íntegro":                theme.StateOK,
		"não instalado":          theme.StateInfo,
		"alterado":               theme.StateWarn,
		"atualização disponível": theme.StateWarn,
	}
	for state, want := range cases {
		if got := homeCoreState(state); got != want {
			t.Fatalf("estado %q: %v, esperado %v", state, got, want)
		}
	}
}

func TestHomeMarkPlacesParticleInsideAndBeyondLine(t *testing.T) {
	plain := theme.New(theme.Options{Dark: true, Profile: colorprofile.ASCII, Plain: true})
	if got := ansi.Strip(homeMark(plain, "abcdef", 2)); got != "ab*def" {
		t.Fatalf("partícula dentro da linha: %q", got)
	}
	if got := ansi.Strip(homeMark(plain, "ab", 4)); got != "ab  *" {
		t.Fatalf("partícula além do fim da linha: %q", got)
	}
	rich := theme.New(theme.Options{Dark: true, Profile: colorprofile.TrueColor})
	if got := ansi.Strip(homeMark(rich, "abc", 1)); got != "a·c" {
		t.Fatalf("glifo em tema rico: %q", got)
	}
}

func TestHomeOverlaySkipsParticlesOutsideBody(t *testing.T) {
	plain := theme.New(theme.Options{Dark: true, Profile: colorprofile.ASCII, Plain: true})
	lines := homeOverlay(plain, []string{"abc"}, []motion.Particle{{X: 1, Y: 0}, {X: 0, Y: 5}})
	if len(lines) != 1 || ansi.Strip(lines[0]) != "a*c" {
		t.Fatalf("sobreposição deve tocar só linhas do corpo: %q", lines)
	}
}

func TestHomeMenuIDsMatchEveryChoice(t *testing.T) {
	want := map[string]bool{
		string(ChoiceInstall): true, string(ChoiceCatalog): true, string(ChoiceDoctor): true,
		string(ChoiceRollback): true, string(ChoiceRollbackAgents): true, string(ChoiceAddons): true,
		string(ChoiceGitHooks): true, string(ChoiceShowcase): true, string(ChoiceHelp): true, string(ChoiceExit): true,
	}
	items := homeMenuItems()
	if len(items) != len(want) {
		t.Fatalf("menu tem %d itens, esperado %d", len(items), len(want))
	}
	for _, item := range items {
		if !want[item.ID] {
			t.Fatalf("item de menu sem escolha correspondente: %q", item.ID)
		}
	}
}

func TestHomeSplitKeepsEmptyLines(t *testing.T) {
	if got := homeSplit("a\n\nb"); len(got) != 3 || got[1] != "" {
		t.Fatalf("quebra de linhas: %q", got)
	}
}
