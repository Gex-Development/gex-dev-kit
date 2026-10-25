package screens

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/app"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/components"
	"github.com/charmbracelet/x/ansi"
)

func TestChatHarnessSpaceEnterAndContinue(t *testing.T) {
	for _, key := range []tea.KeyPressMsg{{Code: ' '}, {Code: ' ', Text: " "}} {
		m := chatTestSize(chatSetupQuestion(t, chatSetupModel(chatInstallFakeDeps()), "Qual harness"), 60, 20)
		for i := range 3 {
			m, _ = chatTestUpdate(m, key)
			if i < 2 {
				m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyRight})
			}
		}
		if m.input.Value() != "" || strings.Join(m.ask.choice.Selected(), ",") != "claude,codex,omp" {
			t.Fatalf("espaço deve marcar três harness sem contaminar input: %q %v", m.input.Value(), m.ask.choice.Selected())
		}
		m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyEnter})
		if m.ask != nil && strings.Contains(m.ask.question, "Qual harness") {
			t.Fatal("enter deve confirmar três harness")
		}
		m = chatSetupQuestion(t, m, "Onde eu instalo?")
		text := chatTestTranscript(m)
		if !strings.Contains(text, "Claude Code, Codex, OMP") || strings.Contains(text, "Verificar Claude Code") || strings.Contains(text, "Verificar Codex") {
			t.Fatalf("verificação deve ser enxuta: %s", text)
		}
		if !strings.Contains(m.View().Content, "versão fixada pelo kit") {
			t.Fatal("banner dos harnesses não apareceu antes do destino")
		}
	}
}

func TestChatHarnessArtFitsTranscriptAndClicksAcrossCodexRows(t *testing.T) {
	m := chatTestSize(chatSetupQuestion(t, chatSetupModel(chatInstallFakeDeps()), "Qual harness"), 110, 38)
	view := m.View().Content
	for _, fragment := range []string{"▐▛███▜▌", "▝▜█████▛▘", "▘▘ ▝▝"} {
		if !strings.Contains(view, fragment) {
			t.Fatalf("art row %q missing", fragment)
		}
	}
	lines := strings.Split(view, "\n")
	codex := -1
	for i, line := range lines {
		if strings.Contains(line, "2. ") && strings.Contains(line, "Codex") {
			codex = i
			break
		}
	}
	if codex < 1 || codex+1 >= len(lines) {
		t.Fatal("Codex art block missing")
	}
	for _, y := range []int{codex - 1, codex, codex + 1} {
		m.ask.choice = m.ask.choice.Preselect(nil)
		clicked, _ := chatTestUpdate(m, tea.MouseClickMsg{X: 1, Y: y, Button: tea.MouseLeft})
		if strings.Join(clicked.ask.choice.Selected(), ",") != "codex" {
			t.Fatalf("click row %d missed Codex: selected=%v line=%q", y, clicked.ask.choice.Selected(), lines[y])
		}
	}
	m = chatTestSize(m, 60, 20)
	view = m.View().Content
	if !strings.Contains(view, "página 1/") || strings.Contains(view, "Codex") {
		t.Fatalf("layout compacto deveria paginar arte: %q", view)
	}
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyRight})
	view = m.View().Content
	if !strings.Contains(view, "Codex") {
		t.Fatalf("página de Codex ausente: page=%d pages=%d view=%q", m.ask.choice.Page(), m.ask.choice.Pages(), ansi.Strip(view))
	}
	for _, line := range strings.Split(view, "\n") {
		if ansi.StringWidth(line) > 60 {
			t.Fatalf("compact line too wide: %q", line)
		}
	}
}

func TestChatIdlePaletteAndBack(t *testing.T) {
	m := chatTestModel(false).flowIdle()
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: '/', Text: "/"})
	if !m.paletteOpen || !strings.Contains(m.View().Content, "configurar skills") {
		t.Fatal("barra deve abrir paleta")
	}
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: 'd', Text: "d"})
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyTab})
	if m.input.Value() != "/doctor " {
		t.Fatalf("tab deve completar comando: %q", m.input.Value())
	}
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyEscape})
	if m.paletteOpen || m.input.Value() != "" {
		t.Fatal("esc deve fechar e limpar paleta")
	}
	m.input.SetValue("/setup")
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyEnter})
	m = chatSetupQuestion(t, m, "Qual harness")
	m.input.SetValue("/voltar")
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.ask != nil || m.context != "pronto" || !strings.Contains(chatTestTranscript(m), "Cancelado. Digite /") {
		t.Fatal("/voltar primeira pergunta deve cancelar sem menu")
	}
}

func TestChatPathPasteNavigateAndLiteralKeywords(t *testing.T) {
	home := t.TempDir()
	if err := os.Mkdir(filepath.Join(home, "setup-git"), 0o700); err != nil {
		t.Fatal(err)
	}
	m := chatTestModel(false)
	m.common.UserHome = home
	m = m.askFor(chatAsk{input: true, path: true, picker: components.NewPathPicker(home, nil), question: "Caminho absoluto do HOME isolado:", validate: chatValidateHome, onAnswer: func(m chatModel, a chatAnswer) chatModel { m.request.Home = a.Text; return m.flowIdle() }})
	m = chatSetupQuestion(t, m, "Caminho absoluto")
	if !strings.Contains(m.View().Content, "setup-git/") {
		t.Fatal("campo vazio deve sugerir pastas do HOME")
	}
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyDown})
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.input.Value() != "~/setup-git/" || m.ask == nil {
		t.Fatalf("enter destacado deve descer pasta: %q", m.input.Value())
	}
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.request.Home != filepath.Join(home, "setup-git") || m.ask != nil {
		t.Fatalf("enter sem destaque envia caminho literal: %q", m.request.Home)
	}
	m = chatTestModel(false).askFor(chatAsk{input: true, question: "Texto:", onAnswer: func(m chatModel, a chatAnswer) chatModel { return m.say(a.Text) }})
	m = chatSetupQuestion(t, m, "Texto:")
	m, _ = chatTestUpdate(m, tea.PasteMsg{Content: "git setup\ninstalar"})
	if m.input.Value() != "git setup instalar" || m.paletteOpen {
		t.Fatalf("paste uma linha: %q", m.input.Value())
	}
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyEnter})
	if len(m.steps) != 0 || !strings.Contains(chatTestTranscript(m), "git setup instalar") {
		t.Fatal("palavra-chave em campo de texto não deve trocar fluxo")
	}
}

func TestChatBackFromPlanRestoresGroupsAndTruncatesTranscript(t *testing.T) {
	d := chatInstallFakeDeps()
	d.Install.Options = func() []app.OptionGroup {
		return []app.OptionGroup{{ID: "skills", Title: "Skills", Options: []app.Option{{Value: "skill-a", Label: "A"}}}, {ID: "hooks", Title: "Hooks", Options: []app.Option{{Value: "hook-a", Label: "H"}}}}
	}
	m := chatSetupQuestion(t, chatSetupModel(d), "Qual harness")
	m = chatSetupChoose(t, m, "omp")
	m = chatSetupQuestion(t, m, "Onde eu instalo?")
	m = chatSetupChoose(t, m, "current")
	m = chatSetupQuestion(t, m, "Além do Core")
	m = chatSetupChoose(t, m, "skills", "hooks")
	m = chatSetupQuestion(t, m, "Skills")
	m = chatSetupChoose(t, m, "skill-a")
	m = chatSetupQuestion(t, m, "Hooks")
	m = chatSetupChoose(t, m, "hook-a")
	m = chatSetupQuestion(t, m, "Aplico essas alterações?")
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyEscape})
	if m.ask == nil || !strings.Contains(m.ask.question, "Hooks") || strings.Join(m.ask.choice.Selected(), ",") != "hook-a" || strings.Contains(chatTestTranscript(m), "Planejar instalação") {
		t.Fatal("esc após ferramenta deve restaurar hooks e remover resumo/plano")
	}
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyEscape})
	if m.ask == nil || !strings.Contains(m.ask.question, "Skills") || strings.Join(m.ask.choice.Selected(), ",") != "skill-a" || m.step != 3 {
		t.Fatal("esc subsequente deve restaurar pré-seleção de skills")
	}
}

func TestChatPathTabAndEscapeClearBeforeBack(t *testing.T) {
	home := t.TempDir()
	if err := os.Mkdir(filepath.Join(home, "repo"), 0o700); err != nil {
		t.Fatal(err)
	}
	m := chatTestModel(false)
	m.common.UserHome = home
	m = m.askFor(chatAsk{input: true, path: true, question: "Caminho absoluto do repositório Git:", picker: components.NewPathPicker(home, []string{".git"})})
	m = chatSetupQuestion(t, m, "repositório Git")
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyTab})
	if m.input.Value() != "~/repo/" {
		t.Fatalf("tab deve completar pasta única: %q", m.input.Value())
	}
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyEscape})
	if m.ask == nil || m.input.Value() != "" {
		t.Fatal("primeiro esc limpa caminho sem voltar")
	}
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyEscape})
	if m.ask != nil || !strings.Contains(chatTestTranscript(m), "Cancelado.") {
		t.Fatal("segundo esc cancela pergunta inicial")
	}
}

func TestChatPathTypingAfterTabAppendsAtEnd(t *testing.T) {
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, "Desktop", "test-harness"), 0o700); err != nil {
		t.Fatal(err)
	}
	m := chatTestModel(false)
	m.common.UserHome = home
	m = m.askFor(chatAsk{input: true, path: true, question: "Caminho absoluto do HOME isolado:", picker: components.NewPathPicker(home, nil)})
	m = chatSetupQuestion(t, m, "HOME isolado")
	for _, r := range "~/Desk" {
		m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyTab})
	for _, r := range "test-h" {
		m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyTab})
	if m.input.Value() != "~/Desktop/test-harness/" {
		t.Fatalf("texto digitado depois do tab deve ir para o fim do campo: %q", m.input.Value())
	}
}

func TestChatTextareaBackAndLiteralFlowWords(t *testing.T) {
	m := chatTestModel(false).askFor(chatAsk{textarea: true, question: "Comandos:", onAnswer: func(m chatModel, a chatAnswer) chatModel { return m.say(a.Text).flowIdle() }})
	m = chatSetupQuestion(t, m, "Comandos:")
	m.area.SetValue("git setup")
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyEscape})
	if m.area.Value() != "" || m.ask == nil {
		t.Fatal("esc deve limpar textarea primeiro")
	}
	m.area.SetValue("git setup")
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyEnter})
	m = chatSetupQuestion(t, m, "")
	if len(m.steps) != 0 || !strings.Contains(chatTestTranscript(m), "git setup") {
		t.Fatal("palavras-chave em textarea devem ser conteúdo literal")
	}
	m = chatTestModel(false).askFor(chatAsk{textarea: true, question: "Comandos:"})
	m = chatSetupQuestion(t, m, "Comandos:")
	m.area.SetValue("/voltar")
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.ask != nil || !strings.Contains(chatTestTranscript(m), "Cancelado.") {
		t.Fatal("/voltar em textarea deve cancelar fluxo")
	}
}

func TestChatPaletteClickExecutesSelectedCommand(t *testing.T) {
	m := chatTestSize(chatTestModel(false).flowIdle(), 80, 24)
	m.input.SetValue("/")
	m.syncPalette()
	lines := strings.Split(m.View().Content, "\n")
	y := -1
	for i, line := range lines {
		if strings.Contains(line, "/setup") && strings.Contains(line, "configurar skills") {
			y = i
			break
		}
	}
	if y < 0 {
		t.Fatal("paleta não apareceu no layout")
	}
	m, _ = chatTestUpdate(m, tea.MouseClickMsg{X: 4, Y: y, Button: tea.MouseLeft})
	if m.paletteOpen || !strings.Contains(chatTestTranscript(m), "/setup") {
		t.Fatal("clique na paleta deve executar setup")
	}
	m = chatSetupQuestion(t, m, "Qual harness")
	if m.ask == nil {
		t.Fatal("setup não abriu pergunta")
	}
}

func TestChatPathPickerClickAndWheelNavigation(t *testing.T) {
	home := t.TempDir()
	for _, name := range []string{"repo-a", "repo-b"} {
		if err := os.Mkdir(filepath.Join(home, name), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	m := chatTestSize(chatTestModel(false), 80, 24)
	m.common.UserHome = home
	m = m.askFor(chatAsk{input: true, path: true, question: "Caminho absoluto do repositório Git:", picker: components.NewPathPicker(home, []string{".git"})})
	m = chatSetupQuestion(t, m, "repositório Git")
	lines := strings.Split(m.View().Content, "\n")
	y := -1
	for i, line := range lines {
		if strings.Contains(line, "repo-b/") {
			y = i
		}
	}
	if y < 0 {
		t.Fatal("sugestão não apareceu no layout")
	}
	m, _ = chatTestUpdate(m, tea.MouseWheelMsg{Button: tea.MouseWheelDown, X: 4, Y: y})
	m, _ = chatTestUpdate(m, tea.MouseClickMsg{Button: tea.MouseLeft, X: 4, Y: y})
	if m.input.Value() != "~/repo-b/" {
		t.Fatalf("clique deve preencher caminho sugerido: %q", m.input.Value())
	}
}

func TestChatPagedChoiceFitsAndRoutesInput(t *testing.T) {
	for _, size := range [][2]int{{110, 34}, {60, 20}} {
		m := chatTestSize(chatTestModel(false), size[0], size[1])
		options := make([]components.ChoiceOption, 11)
		for i := range options {
			options[i] = components.ChoiceOption{ID: fmt.Sprintf("skill-%d", i+1), Label: fmt.Sprintf("Skill %d", i+1), Hint: "Explicação da skill"}
		}
		a := chatAsk{hasChoice: true, question: "Engenharia", options: options, choice: components.NewChoice("Engenharia", options, true, false)}
		m.ask = &a
		text := strings.TrimSuffix(strings.Repeat("transcrição\n", 50), "\n")
		m.entries = []chatEntry{{role: components.ChatAssistant, text: text, shown: len([]rune(text))}}
		rendered := ansi.Strip(m.View().Content)
		lines := strings.Split(rendered, "\n")
		if len(lines) > size[1] || m.layoutList(size[0]) > size[1]-11 {
			t.Fatalf("size=%v linhas=%d lista ultrapassa entrada", size, len(lines))
		}
		if m.ask.choice.Pages() > 1 && !strings.Contains(rendered, "página 1/") {
			t.Fatalf("size=%v página ausente: %s", size, rendered)
		}
		if strings.Count(rendered, "transcrição") < 4 {
			t.Fatalf("size=%v transcrição cortada", size)
		}
		for _, line := range lines {
			if ansi.StringWidth(line) > size[0] {
				t.Fatalf("size=%v linha larga: %q", size, line)
			}
		}
		if m.ask.choice.Pages() == 1 {
			continue
		}
		m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: '1'})
		m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyRight})
		m.placeChoice()
		if m.ask.choice.Page() != 1 || len(m.ask.choice.Selected()) != 1 {
			t.Fatalf("size=%v paginação perdeu seleção", size)
		}
		m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: '2'})
		if len(m.ask.choice.Selected()) != 2 {
			t.Fatalf("size=%v dígito não marcou item da página 2", size)
		}
		m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyLeft})
		m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyPgDown})
		if m.ask.choice.Page() != 1 || m.scroll != 0 {
			t.Fatalf("size=%v PageDown não mudou página: page=%d pages=%d scroll=%d input=%q", size, m.ask.choice.Page(), m.ask.choice.Pages(), m.scroll, m.input.Value())
		}
		m.input.SetValue("abc")
		m.input.CursorEnd()
		page := m.ask.choice.Page()
		m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyLeft})
		m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: 'X', Text: "X"})
		if m.ask.choice.Page() != page || m.input.Value() != "abXc" {
			t.Fatalf("size=%v seta desviou cursor do campo: %q", size, m.input.Value())
		}
	}
}

func TestChatPageClicksUseRenderedCoordinates(t *testing.T) {
	m := chatTestSize(chatTestModel(false), 60, 20)
	options := make([]components.ChoiceOption, 11)
	for i := range options {
		options[i] = components.ChoiceOption{ID: fmt.Sprint(i), Label: fmt.Sprintf("Skill %d", i+1)}
	}
	a := chatAsk{hasChoice: true, question: "Engenharia", options: options, choice: components.NewChoice("Engenharia", options, true, false)}
	m.ask = &a
	for page := 1; page <= 1; page++ {
		lines := strings.Split(ansi.Strip(m.View().Content), "\n")
		y := -1
		for row, line := range lines {
			if strings.Contains(line, "página ") {
				y = row
				x := ansi.StringWidth(line[:strings.Index(line, "›")])
				if x < 0 {
					t.Fatal("controle de página ausente")
				}
				m, _ = chatTestUpdate(m, tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft})
				if m.ask.choice.Page() != page {
					t.Fatalf("clique página %d falhou: atual=%d pages=%d linha=%q y=%d x=%d", page, m.ask.choice.Page(), m.ask.choice.Pages(), line, y, x)
				}
				break
			}
		}
		if y < 0 {
			t.Fatal("linha de página ausente")
		}
	}
	lines := strings.Split(ansi.Strip(m.View().Content), "\n")
	for y, line := range lines {
		if strings.Contains(line, "2. ☐") {
			m, _ = chatTestUpdate(m, tea.MouseClickMsg{X: 4, Y: y, Button: tea.MouseLeft})
			if len(m.ask.choice.Selected()) != 1 {
				t.Fatal("clique no segundo item da página não marcou")
			}
			return
		}
	}
	t.Fatal("segundo item da página não apareceu")
}

func TestChatPaletteShowsAllCommandsWhenRoom(t *testing.T) {
	for _, size := range [][2]int{{110, 34}, {60, 20}} {
		m := chatTestSize(chatTestModel(false).flowIdle(), size[0], size[1])
		m.input.SetValue("/")
		m.syncPalette()
		view := ansi.Strip(m.View().Content)
		if size[0] == 110 && (strings.Contains(view, "mais ") || strings.Count(view, "/") < 12) {
			t.Fatal("paleta larga deveria mostrar 12 comandos")
		}
		if size[0] == 60 && !strings.Contains(view, "… mais ") {
			t.Fatal("paleta compacta deveria mostrar contador")
		}
	}
}

func TestChatPageUpWithoutPagedChoiceScrollsTranscript(t *testing.T) {
	m := chatTestSize(chatTestModel(false), 60, 20)
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyPgUp})
	if m.scroll == 0 {
		t.Fatal("PageUp sem escolha paginada deve rolar a conversa")
	}
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyPgDown})
	if m.scroll != 0 {
		t.Fatal("PageDown sem escolha paginada deve voltar ao fim")
	}
}

func TestChatActualEngineeringChoiceFitsBothSizes(t *testing.T) {
	for _, size := range [][2]int{{110, 34}, {60, 20}} {
		deps := chatInstallFakeDeps()
		deps.Install.Options = app.InstallOptions
		m := chatTestSize(chatSetupModel(deps), size[0], size[1])
		m = chatSetupQuestion(t, m, "Qual harness")
		m = chatSetupChoose(t, m, "omp")
		m = chatSetupQuestion(t, m, "Onde eu instalo?")
		m = chatSetupChoose(t, m, "current")
		m = chatSetupQuestion(t, m, "Além do Core")
		m = chatSetupChoose(t, m, "engenharia")
		m = chatSetupQuestion(t, m, "Engenharia")
		m.placeChoice()
		view := ansi.Strip(m.View().Content)
		if len(m.ask.options) != 11 || !strings.Contains(view, "Confirmar") && !strings.Contains(view, "Pular") {
			t.Fatalf("size=%v pergunta de Engenharia incompleta", size)
		}
		if m.ask.choice.Pages() > 1 && !strings.Contains(view, "página 1/") {
			t.Fatalf("size=%v página da Engenharia ausente", size)
		}
		if len(strings.Split(view, "\n")) > size[1] {
			t.Fatalf("size=%v pergunta passou da altura", size)
		}
	}
}
