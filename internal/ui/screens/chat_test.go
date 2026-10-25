package screens

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/colorprofile"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/Gex-Development/gex-dev-kit/internal/app"
	"github.com/Gex-Development/gex-dev-kit/internal/catalog"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/components"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/motion"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

func chatTestDeps() ChatDeps {
	return ChatDeps{
		Core:   func() (app.CoreStatus, error) { return app.CoreStatus{State: "não instalado", TotalCore: 7}, nil },
		Agents: func() ([]app.AgentChange, error) { return nil, nil },
		Install: InstallDeps{
			Options: func() []app.OptionGroup { return nil },
			Plan:    func(req app.Request) (app.Plan, error) { return app.Plan{Request: req}, nil },
		},
	}
}

func chatTestModel(motionEnabled bool) chatModel {
	c := Common{Version: "0.3.0", Home: "/home/demo", UserHome: "/home/demo", Dark: true, Profile: colorprofile.ANSI, Motion: motionEnabled}
	m := newChatModel(c, chatTestDeps(), IntentHome)
	m.phase = chatTalking
	return m
}

func chatTestUpdate(m chatModel, msg tea.Msg) (chatModel, tea.Cmd) {
	next, cmd := m.Update(msg)
	return next.(chatModel), cmd
}

func chatTestSize(m chatModel, width, height int) chatModel {
	m, _ = chatTestUpdate(m, tea.WindowSizeMsg{Width: width, Height: height})
	return m
}

func chatTestChoice(question string, options ...components.ChoiceOption) chatAsk {
	return chatAsk{hasChoice: true, choice: components.NewChoice(question, options, false, true), options: options}
}

func TestChatStatusLineShowsHarnessGitAndSetupMetrics(t *testing.T) {
	m := chatTestSize(chatTestModel(false), 100, 24)
	m.deps.Dir = "/tmp/project"
	m.agents = []app.AgentChange{{Name: "claude", Current: "2.1.306"}}
	m, _ = chatTestUpdate(m, chatGitMsg{info: app.GitInfo{Branch: "main", Modified: 3, Untracked: 2}, ok: true})
	segments, metrics := m.statusContent()
	line := ansi.Strip(components.StatusLine(m.theme, 100, segments, metrics))
	for _, want := range []string{"GEX", "Claude Code 2.1.306", "project", "main", "*3", "?2", "Core 0/7", "v0.3.0"} {
		if !strings.Contains(line, want) {
			t.Fatalf("status line não mostra %q: %q", want, line)
		}
	}
	m.steps, m.step = []string{"Harness", "Agentes", "Destino", "Extras", "Plugins", "Plano", "Instalação"}, 2
	segments, metrics = m.statusContent()
	line = ansi.Strip(components.StatusLine(m.theme, 100, segments, metrics))
	if !strings.Contains(line, "43%") || !strings.Contains(line, "Etapa 3/7") {
		t.Fatalf("setup não mostra progresso: %q", line)
	}
	m, _ = chatTestUpdate(m, chatGitMsg{ok: false})
	segments, metrics = m.statusContent()
	if line = ansi.Strip(components.StatusLine(m.theme, 100, segments, metrics)); strings.Contains(line, "main") {
		t.Fatalf("git ausente não deve aparecer: %q", line)
	}
}

func TestChatStatusContentHomeAndSelectedHarness(t *testing.T) {
	m := chatTestModel(false)
	m.deps.Dir = "/work/gex-dev-kit"
	m.harnessIDs = []string{"claude", "codex"}
	segments, _ := m.statusContent()
	line := ansi.Strip(components.StatusLine(m.theme, 100, segments, nil))
	if !strings.Contains(line, "Claude Code + Codex") || !strings.Contains(line, "gex-dev-kit") || strings.Contains(line, "/home/demo") {
		t.Fatal(line)
	}
	m.request.Home = "/tmp/isolation"
	segments, _ = m.statusContent()
	line = ansi.Strip(components.StatusLine(m.theme, 100, segments, nil))
	if !strings.Contains(line, "/tmp/isolation") {
		t.Fatal(line)
	}
}

func TestChatPromptBoxWidthAndTranscriptHeight(t *testing.T) {
	for _, size := range [][2]int{{80, 24}, {60, 20}, {120, 30}, {100, 34}} {
		m := chatTestSize(chatTestModel(false), size[0], size[1])
		text := strings.TrimSuffix(strings.Repeat("conteudo\n", 50), "\n")
		m.entries = []chatEntry{{role: components.ChatAssistant, text: text, shown: len([]rune(text))}}
		out, _ := chatLayout(m)
		rows := strings.Split(ansi.Strip(out), "\n")
		want := size[1] - 6
		if got := strings.Count(ansi.Strip(out), "conteudo"); got != want {
			t.Fatalf("size=%v transcrição=%d esperado=%d", size, got, want)
		}
		for _, row := range rows {
			if ansi.StringWidth(row) > size[0] {
				t.Fatalf("size=%v largura=%d", size, ansi.StringWidth(row))
			}
		}
		mainW := size[0]
		if size[0] >= 120 && size[1] >= 30 {
			mainW -= 33
		}
		if got := m.input.Width(); got != mainW-6 {
			t.Fatalf("size=%v entrada=%d esperado=%d", size, got, mainW-6)
		}
	}
}

func TestChatStatusLineGivesTranscriptOneMoreRow(t *testing.T) {
	m := chatTestSize(chatTestModel(false), 80, 24)
	text := strings.TrimSuffix(strings.Repeat("conteudo\n", 30), "\n")
	m.entries = []chatEntry{{role: components.ChatAssistant, text: text, shown: len([]rune(text))}}
	out, _ := chatLayout(m)
	rows := strings.Split(ansi.Strip(out), "\n")
	if len(rows) < 1 || !strings.Contains(rows[0], "conteudo") {
		t.Fatalf("transcrição precisa começar na primeira linha: %q", rows[0])
	}
	if got := strings.Count(ansi.Strip(out), "conteudo"); got != 18 {
		t.Fatalf("transcrição mudou de altura: visíveis=%d, esperado 18", got)
	}
}

func TestChatGitRefreshAfterSuccessfulWrite(t *testing.T) {
	m := chatTestModel(false)
	m.deps.Git = func() (app.GitInfo, bool) { return app.GitInfo{Branch: "updated"}, true }
	m = m.tool(chatToolSpec{title: "Gravar Core", write: true, run: func(func(string)) chatToolResult {
		return chatToolResult{state: theme.StateOK}
	}})
	m, _ = m.pump()
	_, cmd := chatTestUpdate(m, chatFinishedMsg{entry: m.running.entry, result: chatToolResult{state: theme.StateOK}})
	batch, ok := cmd().(tea.BatchMsg)
	if !ok {
		t.Fatalf("gravação deve agendar atualização git: %T", cmd())
	}
	for _, part := range batch {
		switch result := part().(type) {
		case chatGitMsg:
			if !result.ok || result.info.Branch != "updated" {
				t.Fatalf("estado git desatualizado: %+v", result)
			}
			return
		case tea.BatchMsg:
			for _, nested := range result {
				if git, ok := nested().(chatGitMsg); ok {
					if !git.ok || git.info.Branch != "updated" {
						t.Fatalf("estado git desatualizado: %+v", git)
					}
					return
				}
			}
		}
	}
	t.Fatal("gravação não agendou leitura git")
}

func TestChatStreamingLocksInputUntilSayFinishes(t *testing.T) {
	m := chatTestModel(true)
	m = m.say("Uma resposta animada bem longa, com mais texto do que cabe num frame.")
	m = m.say("Próxima resposta")
	m, _ = m.pump()
	if !m.locked() || len(m.entries) != 1 || m.entries[0].shown >= len([]rune(m.entries[0].text)) {
		t.Fatalf("primeiro say deveria começar animado e manter segundo na fila: %+v", m.entries)
	}
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: 'z', Text: "z"})
	if m.input.Value() != "" || len(m.entries) != 1 {
		t.Fatal("letra não pode entrar enquanto GEX escreve")
	}
	for range 2 {
		m, _ = chatTestUpdate(m, motion.FrameMsg{ID: chatFrame, Time: time.Now()})
	}
	if len(m.entries) != 1 || len(m.queue) != 1 {
		t.Fatal("fila avançou antes do fim do texto")
	}
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.entries[0].shown != len([]rune(m.entries[0].text)) || len(m.entries) < 2 {
		t.Fatal("enter revela texto atual e inicia próxima mensagem sem espera")
	}
	for _, entry := range m.entries {
		if entry.role == components.ChatUser {
			t.Fatal("enter enquanto GEX escreve não pode enviar resposta")
		}
	}
}

func TestChatStreamRespectsFrameClock(t *testing.T) {
	t0 := time.Now()
	m := chatTestModel(true).say("Olá. Tudo bem?")
	m, _ = m.pump()
	m, _ = chatTestUpdate(m, motion.FrameMsg{ID: chatStreamFrame, Time: t0})
	m, _ = chatTestUpdate(m, motion.FrameMsg{ID: chatStreamFrame, Time: t0.Add(100 * time.Millisecond)})
	if m.entries[0].shown == 0 || !m.streaming() {
		t.Fatal("escrita deve crescer parcialmente pelo relógio")
	}
	if m.entries[0].at[5]-m.entries[0].at[3] < 250*time.Millisecond {
		t.Fatal("ponto deve criar pausa visível")
	}
}

func TestChatStreamSkipsInvisibleAnsiWithoutSplittingSequences(t *testing.T) {
	at := chatStreamSchedule([]rune("\x1b[31mOlá\x1b[0m!"))
	if len(at) != len([]rune("\x1b[31mOlá\x1b[0m!")) || at[0] != at[4] || at[1] != at[4] {
		t.Fatalf("ANSI deve aparecer atomicamente sem consumir tempo: %v", at)
	}
	if at[len(at)-1] > 150*time.Millisecond {
		t.Fatalf("estilos não devem atrasar escrita: %s", at[len(at)-1])
	}
}

func TestChatStreamPausesWhenTerminalBlurs(t *testing.T) {
	t0 := time.Now()
	m := chatTestModel(true).say(strings.Repeat("abcd", 100))
	m, _ = m.pump()
	m, _ = chatTestUpdate(m, motion.FrameMsg{ID: chatStreamFrame, Time: t0})
	m, _ = chatTestUpdate(m, tea.BlurMsg{})
	m, _ = chatTestUpdate(m, motion.FrameMsg{ID: chatStreamFrame, Time: t0.Add(5 * time.Second)})
	if m.entries[0].shown != 0 {
		t.Fatal("stream não pode avançar enquanto terminal perde foco")
	}
	m, cmd := chatTestUpdate(m, tea.FocusMsg{})
	if cmd == nil || !m.streaming() {
		t.Fatal("foco deve retomar stream sem liberar entrada")
	}
}

func TestChatAddonCardSchedule(t *testing.T) {
	m := chatTestModel(true)
	card := addonsCard(m.theme, catalog.Addons()[0], 108)
	at := chatStreamSchedule([]rune(card))
	t.Logf("addon: %d runas brutas, %d visíveis, cronograma %s", len([]rune(card)), len([]rune(ansi.Strip(card))), at[len(at)-1])
	if at[len(at)-1] > 11*time.Second {
		t.Fatalf("card longo não deve parecer arrastado: %s", at[len(at)-1])
	}
}

func TestChatWorkLabelStages(t *testing.T) {
	phases := []string{"Primeira…", "Segunda…", "Última…"}
	for _, tc := range []struct {
		elapsed time.Duration
		want    string
	}{{0, "Primeira…"}, {1200 * time.Millisecond, "Segunda…"}, {time.Hour, "Última…"}} {
		if got := chatWorkLabel(phases, tc.elapsed); got != tc.want {
			t.Fatalf("%s: %q, esperado %q", tc.elapsed, got, tc.want)
		}
	}
	if got := chatWorkLabel(nil, time.Second); got != "Trabalhando…" {
		t.Fatalf("sem fases: %q", got)
	}
}

func TestChatStreamSchedule(t *testing.T) {
	letters := chatStreamSchedule([]rune(strings.Repeat("abcdefgh", 13)))
	if got := letters[len(letters)-1]; got < 1300*time.Millisecond || got > 3500*time.Millisecond {
		t.Fatalf("100 runas devem levar cerca de 2 s: %s", got)
	}
	if !slices.Equal(letters, chatStreamSchedule([]rune(strings.Repeat("abcdefgh", 13)))) {
		t.Fatal("cronograma deve ser determinístico")
	}
	words := chatStreamSchedule([]rune("Hoje temos quatro palavras diferentes."))
	if words[0] != words[3] || words[4] != words[9] || words[4] <= words[3] {
		t.Fatalf("palavra e pontuação aparecem juntas; espaço vem com próxima: %v", words)
	}
	intervals := []time.Duration{words[4] - words[3], words[10] - words[9], words[17] - words[16]}
	if intervals[0] == intervals[1] && intervals[1] == intervals[2] {
		t.Fatal("intervalos entre palavras devem variar")
	}
	punctuation := chatStreamSchedule([]rune("a. b, c d"))
	if punctuation[3]-punctuation[1] <= punctuation[6]-punctuation[4] || punctuation[6]-punctuation[4] <= punctuation[8]-punctuation[6] {
		t.Fatalf("pausa após ponto supera vírgula e palavras: %v", punctuation)
	}
	version := chatStreamSchedule([]rune("2.1.306 seguinte"))
	if version[0] != version[6] || version[8]-version[6] > 250*time.Millisecond {
		t.Fatalf("versão não cria pausa por ponto interno: %v", version)
	}
	long := chatStreamSchedule([]rune(strings.Repeat("abcdefgh", 188)))
	if got := long[len(long)-1]; got < 6*time.Second || got > 10*time.Second {
		t.Fatalf("texto longo não pode despejar nem arrastar: %s", got)
	}
	long = chatStreamSchedule([]rune(strings.Repeat("Palavra, outra. ", 100)))
	if got := long[len(long)-1]; got < 6*time.Second || got > 10*time.Second {
		t.Fatalf("pontuação em texto longo deve respeitar orçamento: %s", got)
	}
}

func TestChatStreamClockSkipAndBeat(t *testing.T) {
	t0 := time.Now()
	m := chatTestSize(chatTestModel(true), 100, 28).say(strings.Repeat("abcdefgh ", 50)).say("Segunda")
	m, _ = m.pump()
	m, _ = chatTestUpdate(m, motion.FrameMsg{ID: chatStreamFrame, Time: t0})
	at := m.entries[0].at
	m, _ = chatTestUpdate(m, motion.FrameMsg{ID: chatStreamFrame, Time: t0.Add(at[2])})
	if m.entries[0].shown != 8 {
		t.Fatalf("palavra inteira aparece no mesmo instante: %d", m.entries[0].shown)
	}
	m, _ = chatTestUpdate(m, motion.FrameMsg{ID: chatStreamFrame, Time: t0.Add(2 * time.Second)})
	if m.entries[0].shown >= len(m.entries[0].runes) {
		t.Fatal("travada do terminal não pode despejar texto")
	}
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.entries[0].shown != len(m.entries[0].runes) || m.entries[0].at != nil {
		t.Fatal("enter revela mensagem inteira")
	}
	m = chatTestSize(chatTestModel(true), 100, 28).say(strings.Repeat("abcdefgh ", 50)).say("Segunda")
	m, _ = m.pump()
	e := &m.entries[0]
	e.shown = len(e.at) - 1
	m.streamStart = t0.Add(-e.at[len(e.at)-1])
	m.streamLast = t0.Add(-time.Second / 60)
	m, _ = chatTestUpdate(m, motion.FrameMsg{ID: chatStreamFrame, Time: t0})
	if m.beat == 0 || !strings.Contains(ansi.Strip(m.View().Content), "GEX está pensando…") {
		t.Fatal("stream concluído inicia batida")
	}
	m, _ = chatTestUpdate(m, motion.FrameMsg{ID: chatFrame, Time: t0})
	m, _ = chatTestUpdate(m, motion.FrameMsg{ID: chatFrame, Time: t0.Add(699 * time.Millisecond)})
	if len(m.entries) != 1 {
		t.Fatal("mensagem longa acrescenta 250 ms à batida")
	}
	m, _ = chatTestUpdate(m, motion.FrameMsg{ID: chatFrame, Time: t0.Add(700 * time.Millisecond)})
	if len(m.entries) != 2 {
		t.Fatal("batida longa termina em 700 ms")
	}
}

func TestChatBeatBetweenMessagesAndSkip(t *testing.T) {
	m := chatTestSize(chatTestModel(true), 100, 28).say("Primeira").say("Segunda")
	m, _ = m.pump()
	m.entries[0].shown = len(m.entries[0].at) - 1
	t0 := time.Now()
	m.streamStart = t0.Add(-m.entries[0].at[len(m.entries[0].at)-1])
	m.streamLast = t0.Add(-time.Second / 60)
	m, _ = chatTestUpdate(m, motion.FrameMsg{ID: chatStreamFrame, Time: t0})
	if len(m.entries) != 1 || !strings.Contains(ansi.Strip(m.View().Content), "GEX está pensando…") {
		t.Fatal("batida deve separar mensagens")
	}
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: ' '})
	if len(m.entries) != 2 || !m.streaming() {
		t.Fatal("espaço deve pular batida e iniciar próxima mensagem")
	}
}

func TestChatToolRetentionStaggerAndInterrupt(t *testing.T) {
	t0 := time.Now()
	m := chatTestSize(chatTestModel(true), 100, 28)
	m = m.tool(chatToolSpec{title: "Verificar ambiente", write: true, phases: []string{"Home…", "Claude…", "Codex…", "Core…"}, run: func(func(string)) chatToolResult {
		return chatToolResult{state: theme.StateOK}
	}})
	m, _ = m.pump()
	entry := m.running.entry
	m.running.started = t0
	m, _ = chatTestUpdate(m, chatFinishedMsg{entry: entry, result: chatToolResult{state: theme.StateOK, lines: []string{"a", "b", "c", "d"}}})
	if m.running == nil || m.Busy() {
		t.Fatal("resultado real libera Busy sem encerrar retenção")
	}
	for _, tc := range []struct {
		elapsed time.Duration
		lines   int
		label   string
	}{
		{500 * time.Millisecond, 0, "Home…"},
		{1200 * time.Millisecond, 1, "Claude…"},
		{2500 * time.Millisecond, 2, "Codex…"},
		{3500 * time.Millisecond, 3, "Core…"},
		{4 * time.Second, 4, ""},
	} {
		m, _ = chatTestUpdate(m, motion.FrameMsg{ID: chatFrame, Time: t0.Add(tc.elapsed)})
		if got := len(m.entries[entry].tool.Lines); got != tc.lines {
			t.Fatalf("em %s, linhas=%d, esperado %d", tc.elapsed, got, tc.lines)
		}
		if tc.label != "" && chatWorkLabel(m.running.phases, tc.elapsed) != tc.label {
			t.Fatalf("fase em %s: %s", tc.elapsed, chatWorkLabel(m.running.phases, tc.elapsed))
		}
	}
	if m.running != nil || m.entries[entry].tool.State != theme.StateOK {
		t.Fatal("ferramenta deve terminar aos 4 s")
	}
	m = chatTestModel(true).tool(chatToolSpec{title: "Gravar", write: true, phases: []string{"Uma…", "Duas…"}, run: func(func(string)) chatToolResult { return chatToolResult{} }})
	m, _ = m.pump()
	m, _ = chatTestUpdate(m, chatFinishedMsg{entry: m.running.entry, result: chatToolResult{state: theme.StateOK}})
	m, cmd := chatTestUpdate(m, tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	if m.signal != nil || !m.interrupted || cmd == nil {
		t.Fatal("ctrl+c durante retenção não pode ser adiado")
	}
}

func TestChatToolChecklistReplacesRowsAndResults(t *testing.T) {
	m := chatTestSize(chatTestModel(true), 100, 28).tool(chatToolSpec{
		title: "Verificar ambiente", phases: []string{"Checando sua pasta home…", "Procurando o Claude Code…", "Procurando o Codex…", "Lendo o Core instalado…"},
		run: func(func(string)) chatToolResult { return chatToolResult{state: theme.StateOK} },
	})
	m, _ = m.pump()
	entry := m.running.entry
	t0 := time.Now()
	m.running.started = t0
	rows := []components.ToolRow{
		{State: theme.StateOK, Label: "HOME", Detail: "~"},
		{State: theme.StateWarn, Label: "Claude Code", Detail: "bloqueado"},
		{State: theme.StateWarn, Label: "Codex", Detail: "não instalado"},
		{State: theme.StateInfo, Label: "Core", Detail: "0/7 skills"},
	}
	m, _ = chatTestUpdate(m, chatFinishedMsg{entry: entry, result: chatToolResult{state: theme.StateOK, rows: rows}})
	for _, tc := range []struct {
		elapsed time.Duration
		want    []string
		absent  []string
	}{
		{500 * time.Millisecond, []string{"Checando sua pasta home…", "○ Procurando o Claude Code"}, []string{"✓ HOME"}},
		{1500 * time.Millisecond, []string{"✓ HOME", "Procurando o Claude Code…", "○ Procurando o Codex"}, []string{"Checando sua pasta home"}},
		{2500 * time.Millisecond, []string{"✓ HOME", "! Claude Code", "Procurando o Codex…"}, []string{"○ Procurando o Claude Code"}},
	} {
		m, _ = chatTestUpdate(m, motion.FrameMsg{ID: chatFrame, Time: t0.Add(tc.elapsed)})
		m.running.started = time.Now().Add(-tc.elapsed)
		view := ansi.Strip(m.View().Content)
		for _, want := range tc.want {
			if !strings.Contains(view, want) {
				t.Fatalf("%s sem %q: %s", tc.elapsed, want, view)
			}
		}
		for _, absent := range tc.absent {
			if strings.Contains(view, absent) {
				t.Fatalf("%s ainda contém %q", tc.elapsed, absent)
			}
		}
		m.running.started = t0
	}
	m, _ = chatTestUpdate(m, motion.FrameMsg{ID: chatFrame, Time: t0.Add(4 * time.Second)})
	if m.running != nil || !strings.Contains(ansi.Strip(m.View().Content), "0/7 skills") {
		t.Fatal("resultado final deve substituir checklist inteiro")
	}
}

func TestChatToolChecklistOtherResultsAndMotionOff(t *testing.T) {
	for _, motionEnabled := range []bool{true, false} {
		m := chatTestSize(chatTestModel(motionEnabled), 100, 28).tool(chatToolSpec{
			title: "Planejar instalação", phases: []string{"Montando as regras do Core…", "Conferindo arquivos…", "Comparando versões…"},
			run: func(func(string)) chatToolResult { return chatToolResult{state: theme.StateOK} },
		})
		m, _ = m.pump()
		entry := m.running.entry
		t0 := time.Now()
		m.running.started = t0
		m, _ = chatTestUpdate(m, chatFinishedMsg{entry: entry, result: chatToolResult{state: theme.StateOK, lines: []string{"linha do plano"}}})
		if motionEnabled {
			m, _ = chatTestUpdate(m, motion.FrameMsg{ID: chatFrame, Time: t0.Add(1500 * time.Millisecond)})
			m.running.started = time.Now().Add(-1500 * time.Millisecond)
			view := ansi.Strip(m.View().Content)
			if !strings.Contains(view, "✓ Montando as regras do Core") || !strings.Contains(view, "Conferindo arquivos…") || strings.Contains(view, "linha do plano") {
				t.Fatal(view)
			}
			m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyEnter})
		}
		view := ansi.Strip(m.View().Content)
		if m.running != nil || !strings.Contains(view, "linha do plano") || strings.Contains(view, "Montando as regras do Core") {
			t.Fatal(view)
		}
	}
}

func TestChatToolFillSeesLatestModel(t *testing.T) {
	m := chatTestModel(false).tool(chatToolSpec{
		title: "Ambiente",
		run: func(func(string)) chatToolResult {
			return chatToolResult{state: theme.StateOK, lines: []string{"inicial"}}
		},
		fill: func(m chatModel) chatToolResult { return chatToolResult{lines: []string{m.core.State}} },
	})
	m, _ = m.pump()
	m.coreDone, m.agentsDone = true, true
	m.core.State = "instalado"
	m, _ = chatTestUpdate(m, chatFinishedMsg{
		entry: m.running.entry, result: chatToolResult{state: theme.StateOK, lines: []string{"inicial"}},
		fill: func(m chatModel) chatToolResult { return chatToolResult{lines: []string{m.core.State}} },
	})
	if got := m.entries[0].tool.Lines; !slices.Equal(got, []string{"inicial", "instalado"}) {
		t.Fatalf("fill deve usar estado atual: %v", got)
	}
}

func TestChatWelcomeWaitsForEnvironment(t *testing.T) {
	m := newChatModel(Common{Home: "/home/demo", UserHome: "/home/demo"}, chatTestDeps(), IntentHome)
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyEnter})
	m, _ = chatTestUpdate(m, tea.MouseClickMsg{})
	if m.phase != chatWelcome {
		t.Fatal("não avançar antes de carregar")
	}
	m, _ = chatTestUpdate(m, chatCoreMsg{status: app.CoreStatus{State: "não instalado"}})
	m, _ = chatTestUpdate(m, chatAgentsMsg{})
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.phase != chatTalking {
		t.Fatal("avançar após carregar")
	}
}

func TestChatEnvironmentShowsEachCurrentStatus(t *testing.T) {
	m := chatTestModel(false).flowGreeting(true)
	spec := m.queue[0].tool
	m, _ = m.pump()
	m.coreDone, m.agentsDone = true, true
	m.core = app.CoreStatus{State: "não instalado", ActiveCore: 0}
	m.agents = []app.AgentChange{{Name: "claude", Current: "2.1.306", Target: "2.1.306"}}
	m, _ = chatTestUpdate(m, chatFinishedMsg{entry: m.running.entry, result: chatToolResult{state: theme.StateOK}, fill: spec.fill})
	rows := m.entries[0].tool.Rows
	if len(rows) != 4 || rows[0].Label != "HOME" || rows[1].Detail != "2.1.306 · em dia" || rows[2].Detail != "não detectado" || rows[3].Detail != "0/7 skills · não instalado" {
		t.Fatalf("resultado correto e ordenado: %+v", rows)
	}
	if strings.Contains(ansi.Strip(m.entries[0].tool.View(m.theme, 100)), "sem dados") {
		t.Fatal("resultado não deve exibir sem dados")
	}
}

func TestChatEnvironmentAgentDetails(t *testing.T) {
	for _, tc := range []struct {
		change app.AgentChange
		want   string
		state  theme.State
	}{
		{app.AgentChange{Current: "1.0", Target: "2.0"}, "1.0 · o kit fixa 2.0", theme.StateWarn},
		{app.AgentChange{Target: "2.0"}, "não instalado · o kit fixa 2.0", theme.StateWarn},
		{app.AgentChange{Blocked: true, Reason: "conflito"}, "bloqueado · conflito", theme.StateWarn},
	} {
		m := chatTestModel(false)
		m.agentsDone, m.coreDone = true, true
		tc.change.Name = "claude"
		m.agents = []app.AgentChange{tc.change}
		rows := m.flowGreeting(true).queue[0].tool.fill(m).rows
		if rows[1].Detail != tc.want || rows[1].State != tc.state {
			t.Fatalf("%+v: %+v", tc.change, rows[1])
		}
	}
	m := chatTestModel(false)
	m.agentsErr = fmt.Errorf("falha")
	m.coreErr = fmt.Errorf("falha")
	rows := m.flowGreeting(true).queue[0].tool.fill(m).rows
	if rows[1].State != theme.StateError || rows[1].Detail != "erro ao ler: falha" || rows[3].State != theme.StateError || rows[3].Detail != "erro ao ler: falha" {
		t.Fatal(rows)
	}
}

func TestChatEnvironmentWaitsForReadsAndTimeout(t *testing.T) {
	for _, done := range []bool{true, false} {
		m := chatTestSize(chatTestModel(true), 100, 28).flowGreeting(true)
		spec := m.queue[0].tool
		m, _ = m.pump()
		now := time.Now()
		m.running.started = now
		m, _ = chatTestUpdate(m, chatFinishedMsg{entry: m.running.entry, result: chatToolResult{state: theme.StateOK}, fill: spec.fill})
		m, _ = chatTestUpdate(m, motion.FrameMsg{ID: chatFrame, Time: now.Add(5 * time.Second)})
		if m.running == nil || len(m.entries[0].tool.Rows) != 0 || !strings.Contains(ansi.Strip(m.View().Content), "Lendo o ambiente") {
			t.Fatal("resultado espera leituras")
		}
		m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyEnter})
		if m.running == nil {
			t.Fatal("Enter não pode pular leitura")
		}
		if done {
			m, _ = chatTestUpdate(m, chatCoreMsg{status: app.CoreStatus{State: "não instalado"}})
			m, _ = chatTestUpdate(m, chatAgentsMsg{})
			m, _ = chatTestUpdate(m, motion.FrameMsg{ID: chatFrame, Time: now.Add(6 * time.Second)})
			if m.running != nil || m.entries[0].tool.Rows[3].Detail != "0/7 skills · não instalado" {
				t.Fatal("concluir depois das leituras")
			}
		} else {
			m, _ = chatTestUpdate(m, motion.FrameMsg{ID: chatFrame, Time: now.Add(8 * time.Second)})
			if m.running != nil || m.entries[0].tool.Rows[3].Detail != "erro ao ler: tempo esgotado" {
				t.Fatal("teto de oito segundos")
			}
		}
	}
}

func TestChatToolRevealsRowsPerPhase(t *testing.T) {
	now := time.Now()
	m := chatTestModel(true).tool(chatToolSpec{title: "Ambiente", phases: []string{"HOME", "Claude", "Codex", "Core"}, run: func(func(string)) chatToolResult { return chatToolResult{} }})
	m, _ = m.pump()
	m.running.started = now
	rows := []components.ToolRow{{Label: "HOME"}, {Label: "Claude"}, {Label: "Codex"}, {Label: "Core"}}
	m, _ = chatTestUpdate(m, chatFinishedMsg{entry: m.running.entry, result: chatToolResult{state: theme.StateOK, rows: rows}})
	for i := range 4 {
		m, _ = chatTestUpdate(m, motion.FrameMsg{ID: chatFrame, Time: now.Add(time.Duration(i+1) * time.Second)})
		if got := len(m.entries[0].tool.Rows); got != i+1 {
			t.Fatalf("fase %d exibiu %d linhas", i+1, got)
		}
	}
}

func TestChatPlanHasOnlyCoreSelectionAndComparisonPhases(t *testing.T) {
	m := chatTestModel(false)
	m.steps = []string{"Harness", "Agentes", "Destino", "Extras", "Plano", "Instalação"}
	m = m.setupPlan()
	phases := m.queue[len(m.queue)-1].tool.phases
	if !slices.Equal(phases, []string{"Montando as regras do Core…", "Selecionando as skills…", "Comparando com o que já existe…"}) {
		t.Fatalf("fases do plano: %v", phases)
	}
}

func TestChatDeferredSignalSkipsRetention(t *testing.T) {
	m := chatTestModel(true).tool(chatToolSpec{title: "Gravar", write: true, phases: []string{"Primeira…", "Segunda…"}, run: func(func(string)) chatToolResult { return chatToolResult{} }})
	m, _ = m.pump()
	m.signal = os.Interrupt
	m, cmd := chatTestUpdate(m, chatFinishedMsg{entry: m.running.entry, result: chatToolResult{state: theme.StateOK}})
	if cmd == nil {
		t.Fatal("resultado após sinal deve sair sem reter")
	}
	if _, quit := cmd().(tea.QuitMsg); !quit {
		t.Fatal("resultado após sinal deve emitir Quit")
	}
}

func TestChatToolRetentionSkipAndNoMotion(t *testing.T) {
	for _, enabled := range []bool{true, false} {
		m := chatTestModel(enabled).tool(chatToolSpec{title: "Plano", phases: []string{"Uma…", "Duas…"}, run: func(func(string)) chatToolResult { return chatToolResult{} }})
		m, _ = m.pump()
		m, _ = chatTestUpdate(m, chatFinishedMsg{entry: m.running.entry, result: chatToolResult{state: theme.StateOK, lines: []string{"resultado"}}})
		if enabled {
			if m.running == nil || m.entries[0].tool.State != theme.StateRunning {
				t.Fatal("Motion deve reter resultado")
			}
			m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyEnter})
		}
		if m.running != nil || m.entries[0].tool.State != theme.StateOK || len(m.entries[0].tool.Lines) != 1 {
			t.Fatalf("resultado deve aparecer inteiro após skip ou sem Motion: %+v", m.entries[0].tool)
		}
	}
}

func TestChatToolWaitsForPhaseDeadline(t *testing.T) {
	m := chatTestSize(chatTestModel(true), 100, 28)
	m = m.tool(chatToolSpec{title: "Plano", phases: []string{"Primeira…", "Segunda…"}, run: func(func(string)) chatToolResult { return chatToolResult{} }})
	m, _ = m.pump()
	m, _ = chatTestUpdate(m, chatFinishedMsg{entry: m.running.entry, result: chatToolResult{state: theme.StateOK}})
	t0 := time.Now()
	m.running.started = t0.Add(-1200 * time.Millisecond)
	m, _ = chatTestUpdate(m, motion.FrameMsg{ID: chatFrame, Time: t0})
	if m.running == nil || !strings.Contains(ansi.Strip(m.View().Content), "Segunda…") {
		t.Fatal("resultado retido até fase final terminar")
	}
	m, _ = chatTestUpdate(m, motion.FrameMsg{ID: chatFrame, Time: t0.Add(800 * time.Millisecond)})
	if m.running != nil {
		t.Fatal("retenção deve terminar depois de 2 s")
	}
	m.reveal = m.reveal.Snap(1)
	_, cmd := chatTestUpdate(m, motion.FrameMsg{ID: chatFrame, Time: time.Now()})
	if cmd != nil {
		t.Fatal("frame ocioso não deve agendar outro frame")
	}
}

func TestChatInstallProgressKeepsTimedLabels(t *testing.T) {
	m := chatTestSize(chatTestModel(true), 100, 28)
	m = m.tool(chatToolSpec{title: "Instalar Core", phases: []string{"Criando regras…", "Instalando skills Core…", "Salvando o histórico…"}, run: func(func(string)) chatToolResult { return chatToolResult{} }})
	m, _ = m.pump()
	m, _ = chatTestUpdate(m, chatProgressMsg{entry: m.running.entry, line: "✓ " + installStepLine(app.StepDone)})
	if !strings.Contains(ansi.Strip(m.View().Content), "Criando regras…") || m.entries[0].tool.Lines[0] != "✓ "+installStepLine(app.StepDone) {
		t.Fatal("progresso real aparece já no bloco; fase segue tempo de exibição")
	}
	m.running.started = time.Now().Add(-1200 * time.Millisecond)
	if !strings.Contains(ansi.Strip(m.View().Content), "Instalando skills Core…") {
		t.Fatal("fase avança pelo tempo, mesmo após StepDone")
	}
}

func TestChatChoiceEchoPrecedesAssistantReply(t *testing.T) {
	m := chatTestModel(false)
	a := chatTestChoice("Qual opção?", components.ChoiceOption{ID: "a", Label: "Primeira"})
	a.onAnswer = func(m chatModel, _ chatAnswer) chatModel { return m.say("Resposta do GEX") }
	m.ask = &a
	m, _ = chatTestUpdate(m, components.ChoiceDoneMsg{IDs: []string{"a"}})
	if len(m.entries) == 0 || m.entries[0].role != components.ChatUser || !strings.Contains(m.entries[0].text, "Primeira") {
		t.Fatalf("escolha precisa entrar na transcrição imediatamente: %+v", m.entries)
	}
	for _, entry := range m.entries[1:] {
		if entry.role == components.ChatUser {
			t.Fatal("resposta do usuário duplicada")
		}
	}
}

func TestChatWriteDefersInterruptUntilToolCompletes(t *testing.T) {
	m := chatTestModel(false)
	m = m.tool(chatToolSpec{title: "Gravar Core", write: true, run: func(func(string)) chatToolResult {
		return chatToolResult{state: theme.StateOK}
	}})
	m, _ = m.pump()
	if !m.Busy() {
		t.Fatal("ferramenta de gravação deve marcar Busy")
	}
	var cmd tea.Cmd
	m, cmd = chatTestUpdate(m, tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	if cmd != nil {
		if _, quit := cmd().(tea.QuitMsg); quit {
			t.Fatal("ctrl+c não pode sair durante gravação")
		}
	}
	if m.signal != os.Interrupt || !m.Busy() {
		t.Fatalf("interrupção deveria ficar adiada até fim da gravação: %v", m.signal)
	}
	if !strings.Contains(chatTestTranscript(m), "Ctrl+C adiado") {
		t.Fatal("aviso da interrupção adiada não apareceu")
	}
	m, cmd = chatTestUpdate(m, chatFinishedMsg{entry: m.running.entry, result: chatToolResult{state: theme.StateOK}})
	if cmd == nil {
		t.Fatal("conclusão da gravação deve emitir Quit após ctrl+c adiado")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("conclusão da gravação não encerrou chat")
	}
}

func TestChatEscapeRestoresPreviousQuestionAndTranscript(t *testing.T) {
	m := chatTestModel(false)
	first := chatTestChoice("Primeira pergunta?", components.ChoiceOption{ID: "a", Label: "Primeira"})
	first.mark = 1
	second := chatTestChoice("Segunda pergunta?", components.ChoiceOption{ID: "b", Label: "Segunda"})
	second.mark = 3
	m.entries = []chatEntry{{role: components.ChatAssistant, text: "Antes", shown: 5}, {role: components.ChatUser, text: "Primeira", shown: 8}, {role: components.ChatAssistant, text: "Depois", shown: 6}}
	m.history = []chatAsk{first, second}
	m.ask = &second
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyEscape})
	if m.ask == nil || !strings.Contains(m.ask.choice.View(m.theme, 80), "Primeira pergunta?") || len(m.entries) != 1 || m.entries[0].text != "Antes" || len(m.queue) != 0 {
		t.Fatalf("esc deve voltar pergunta e transcrição ao marco anterior: %+v", m.entries)
	}
}

func TestChatUnknownCommandAndNaturalSetupRequest(t *testing.T) {
	m := chatTestModel(false).flowIdle()
	m.input.SetValue("/xyz")
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyEnter})
	if !strings.Contains(chatTestTranscript(m), "Não entendi. Digite / para ver os comandos.") || m.ask != nil {
		t.Fatal("comando desconhecido deve explicar e manter chat ocioso")
	}
	m = chatTestModel(false)
	m.input.SetValue("instalar o kit")
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyEnter})
	if len(m.steps) == 0 || m.steps[0] != "Harness" {
		t.Fatalf("pedido em linguagem natural deveria abrir setup: %+v", m.steps)
	}
}

func chatTestTranscript(m chatModel) string {
	var lines []string
	for _, e := range m.entries {
		lines = append(lines, e.text)
	}
	return strings.Join(lines, "\n")
}

func chatTestQueuedSpeech(m chatModel) string {
	var lines []string
	for _, action := range m.queue {
		lines = append(lines, action.say)
	}
	return strings.Join(lines, "\n")
}

func TestChatMenuSecondMouseClickConfirms(t *testing.T) {
	m := chatTestSize(chatTestModel(false), 100, 34)
	a := chatTestChoice("O que você quer fazer?", components.ChoiceOption{ID: "setup", Label: "Setup do zero"}, components.ChoiceOption{ID: "exit", Label: "Sair"})
	a.onAnswer = func(m chatModel, _ chatAnswer) chatModel { return m.say("Até mais!") }
	m.ask = &a
	lines := strings.Split(ansi.Strip(m.View().Content), "\n")
	y := -1
	for i, line := range lines {
		if strings.Contains(line, "2. Sair") {
			y = i
			break
		}
	}
	if y < 0 {
		t.Fatal("segunda opção não aparece no View")
	}
	m, _ = chatTestUpdate(m, tea.MouseClickMsg{X: 5, Y: y, Button: tea.MouseLeft})
	if len(m.entries) != 0 {
		t.Fatal("primeiro clique só posiciona cursor")
	}
	m, _ = chatTestUpdate(m, tea.MouseClickMsg{X: 5, Y: y, Button: tea.MouseLeft})
	if len(m.entries) == 0 || m.entries[0].role != components.ChatUser || !strings.Contains(m.entries[0].text, "Sair") {
		t.Fatal("segundo clique deveria confirmar Sair")
	}
}

func TestChatToolCardHeaderClickAndChoicePosition(t *testing.T) {
	m := chatTestSize(chatTestModel(false), 80, 24)
	m.entries = append(m.entries, chatEntry{isTool: true, tool: &components.ToolBlock{Title: "Planejar instalação", State: theme.StateOK, Lines: []string{"primeira", "segunda"}}})
	a := chatTestChoice("Aplico essas alterações?", components.ChoiceOption{ID: "yes", Label: "Sim, aplicar"}, components.ChoiceOption{ID: "no", Label: "Cancelar"})
	m.ask = &a
	view, hits := chatLayout(m)
	if len(hits) != 1 || hits[0].y < 0 || !strings.Contains(strings.Split(ansi.Strip(view), "\n")[hits[0].y], "Planejar instalação") {
		t.Fatalf("hit deve apontar para borda superior: %+v", hits)
	}
	m, _ = chatTestUpdate(m, tea.MouseClickMsg{X: 4, Y: hits[0].y, Button: tea.MouseLeft})
	if !m.entries[0].tool.Expanded {
		t.Fatal("clique na primeira linha alterna expansão")
	}
	lines := strings.Split(ansi.Strip(m.View().Content), "\n")
	for y, line := range lines {
		if strings.Contains(line, "2. Cancelar") {
			m, _ = chatTestUpdate(m, tea.MouseClickMsg{X: 5, Y: y, Button: tea.MouseLeft})
			m, _ = chatTestUpdate(m, tea.MouseClickMsg{X: 5, Y: y, Button: tea.MouseLeft})
			if len(m.entries) < 2 || !strings.Contains(m.entries[1].text, "Cancelar") {
				t.Fatalf("card deslocou clique na escolha: %d, %q", y, line)
			}
			return
		}
	}
	t.Fatal("opção após cartão não apareceu")
}

func TestChatToolHitsExcludeScrolledOutCards(t *testing.T) {
	m := chatTestSize(chatTestModel(false), 60, 20)
	m.entries = []chatEntry{{isTool: true, tool: &components.ToolBlock{Title: "Fora da tela", State: theme.StateOK}}}
	for range 30 {
		m.entries = append(m.entries, chatEntry{role: components.ChatAssistant, text: "mensagem", shown: 8})
	}
	a := chatTestChoice("Qual item?", components.ChoiceOption{ID: "a", Label: "Escolha"})
	m.ask = &a
	_, hits := chatLayout(m)
	if len(hits) != 0 {
		t.Fatalf("cartão fora da transcrição ainda captura cliques: %+v", hits)
	}
}

func TestChatWheelScrollsTranscript(t *testing.T) {
	m := chatTestSize(chatTestModel(false), 80, 24)
	for range 40 {
		m.entries = append(m.entries, chatEntry{role: components.ChatAssistant, text: "Linha da transcrição", shown: len([]rune("Linha da transcrição"))})
	}
	m, _ = chatTestUpdate(m, tea.MouseWheelMsg{Button: tea.MouseWheelUp})
	if m.scroll != 3 {
		t.Fatalf("roda para cima: scroll=%d, esperado 3", m.scroll)
	}
	m, _ = chatTestUpdate(m, tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	if m.scroll != 0 {
		t.Fatalf("roda para baixo deveria colar no fim: %d", m.scroll)
	}
}

func TestChatWideSidebarAndTerminalBounds(t *testing.T) {
	for _, size := range [][2]int{{120, 30}, {100, 34}, {80, 24}, {60, 20}} {
		m := chatTestSize(chatTestModel(false), size[0], size[1])
		m.entries = []chatEntry{{role: components.ChatAssistant, text: "Olá, teste de largura com texto longo que não deve estourar nenhuma coluna da transcrição.", shown: 100}}
		lines := strings.Split(ansi.Strip(m.View().Content), "\n")
		for y, line := range lines {
			if ansi.StringWidth(line) > size[0] {
				t.Errorf("%dx%d: linha %d excede largura (%d): %q", size[0], size[1], y, ansi.StringWidth(line), line)
			}
		}
		screen := strings.Join(lines, "\n")
		if size[0] == 120 {
			found := false
			for _, line := range lines {
				if col := strings.Index(line, "AMBIENTE"); col >= 88 {
					found = true
				}
			}
			if !found {
				t.Error("120x30 deveria mostrar AMBIENTE na barra lateral após coluna 88")
			}
		} else if strings.Contains(screen, "AMBIENTE") {
			t.Errorf("%dx%d não deveria mostrar barra lateral", size[0], size[1])
		}
	}
}

func TestChatWelcomeAndWorkingPrompt(t *testing.T) {
	c := Common{Version: "0.3.0", Home: "/home/demo", UserHome: "/home/demo", Dark: true, Profile: colorprofile.ANSI}
	m := chatTestSize(newChatModel(c, chatTestDeps(), IntentHome), 80, 24)
	if view := ansi.Strip(m.View().Content); !strings.Contains(view, "Assistente de setup do Kit GEX") || !strings.Contains(view, "Verificando seu ambiente") || strings.Contains(view, "Pressione enter para começar") {
		t.Fatalf("boas-vindas incompletas: %s", view)
	}
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.phase != chatWelcome {
		t.Fatal("enter prematuro não deve iniciar conversa")
	}
	m, _ = chatTestUpdate(m, chatCoreMsg{status: app.CoreStatus{State: "não instalado"}})
	m, _ = chatTestUpdate(m, chatAgentsMsg{})
	if !strings.Contains(ansi.Strip(m.View().Content), "Pressione enter para começar") {
		t.Fatal("pronto deve convidar para começar")
	}
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.phase != chatTalking || len(m.entries) == 0 {
		t.Fatal("enter após leituras deveria iniciar conversa")
	}
	m = chatTestModel(true)
	m = m.say("Escrevendo algo longo para manter entrada travada por alguns frames.")
	m, _ = m.pump()
	view := ansi.Strip(m.View().Content)
	if !strings.Contains(view, "GEX está escrevendo") || !strings.Contains(view, "GEX está trabalhando") || !strings.Contains(view, "ctrl+c sair") {
		t.Fatalf("linha de trabalho ou prompt bloqueado ausente: %s", view)
	}
}

func TestChatToolProgressAndExpansion(t *testing.T) {
	m := chatTestSize(chatTestModel(false), 80, 24)
	m = m.tool(chatToolSpec{title: "Verificar ambiente", run: func(func(string)) chatToolResult { return chatToolResult{state: theme.StateOK} }})
	m, _ = m.pump()
	entry := m.running.entry
	m, _ = chatTestUpdate(m, chatProgressMsg{entry: entry, line: "Arquivo 1 concluído"})
	if got := m.entries[entry].tool.Lines; len(got) != 1 || got[0] != "Arquivo 1 concluído" {
		t.Fatalf("progresso deve entrar ao vivo no bloco: %v", got)
	}
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: 'o', Mod: tea.ModCtrl})
	if !m.entries[entry].tool.Expanded {
		t.Fatal("ctrl+o deveria expandir último bloco")
	}
	m, _ = chatTestUpdate(m, chatFinishedMsg{entry: entry, result: chatToolResult{state: theme.StateOK}})
	if m.running != nil || m.entries[entry].tool.State != theme.StateOK {
		t.Fatal("resultado deveria concluir ferramenta e liberar entrada")
	}
}
