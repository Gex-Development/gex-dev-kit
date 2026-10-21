package screens

import (
	"strings"
	"sync"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/teatest/v2"

	"github.com/Gex-Development/gex-dev-kit/internal/app"
	"github.com/Gex-Development/gex-dev-kit/internal/githooks"
)

type githooksFake struct {
	plan     app.GitHooksPlan
	planErr  error
	applied  []app.GitHooksPlan
	requests []app.GitHooksRequest
}

func (f *githooksFake) deps() GitHooksDeps {
	return GitHooksDeps{
		Plan: func(r app.GitHooksRequest) (app.GitHooksPlan, error) {
			f.requests = append(f.requests, r)
			return f.plan, f.planErr
		},
		Apply: func(p app.GitHooksPlan) (app.Result, error) {
			f.applied = append(f.applied, p)
			return app.Result{HistoryID: "abc123"}, nil
		},
	}
}

func githooksPlanFake() app.GitHooksPlan {
	return app.GitHooksPlan{
		Scripts:     map[string]string{"hooks/pre-commit": "#!/bin/sh\nset -e\nmake test\n"},
		Sources:     []app.CommandSource{{Command: "make test", File: "Makefile", Line: 3}},
		HomeChanges: []app.Change{{Path: "bin/gex-dev-kit", Action: "criar"}},
	}
}

func githooksCommon(t *testing.T) Common {
	return Common{Version: "teste", Home: t.TempDir(), UserHome: t.TempDir(), Dark: true, Profile: colorprofile.ASCII}
}

func githooksKey(s string) tea.KeyPressMsg {
	switch s {
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEsc}
	case "tab":
		return tea.KeyPressMsg{Code: tea.KeyTab}
	case "space":
		return tea.KeyPressMsg{Code: tea.KeySpace}
	case "ctrl+c":
		return tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}
	case "ctrl+s":
		return tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl}
	}
	r := []rune(s)[0]
	return tea.KeyPressMsg{Code: r, Text: s}
}

func githooksSend(m githooksModel, msgs ...tea.Msg) (githooksModel, tea.Cmd) {
	var cmd tea.Cmd
	for _, msg := range msgs {
		next, c := m.Update(msg)
		m = next.(githooksModel)
		cmd = c
	}
	return m, cmd
}

func githooksType(s string) []tea.Msg {
	var msgs []tea.Msg
	for _, r := range s {
		msgs = append(msgs, githooksKey(string(r)))
	}
	return msgs
}

func githooksRun(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}
	msg := cmd()
	batch, ok := msg.(tea.BatchMsg)
	if !ok {
		return []tea.Msg{msg}
	}
	var out []tea.Msg
	for _, c := range batch {
		if c != nil {
			out = append(out, c())
		}
	}
	return out
}

func githooksNew(t *testing.T, f *githooksFake, preset app.GitHooksRequest) githooksModel {
	t.Helper()
	m := newGitHooksModel(githooksCommon(t), f.deps(), &sync.WaitGroup{}, preset)
	m, _ = githooksSend(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	return m
}

func githooksView(m githooksModel) string {
	return ansi.Strip(m.View().Content)
}

func TestGitHooksCaminhoRelativoNaoAvanca(t *testing.T) {
	m := githooksNew(t, &githooksFake{}, app.GitHooksRequest{})
	m, _ = githooksSend(m, githooksType("relativo/projeto")...)
	m, _ = githooksSend(m, githooksKey("enter"))
	if m.step != 1 {
		t.Fatalf("caminho relativo avançou para o passo %d", m.step)
	}
	if !strings.Contains(githooksView(m), "absoluto") {
		t.Fatalf("erro de caminho relativo não apareceu na tela")
	}
}

func TestGitHooksCaminhoAbsolutoAvanca(t *testing.T) {
	m := githooksNew(t, &githooksFake{}, app.GitHooksRequest{})
	dir := t.TempDir()
	m, _ = githooksSend(m, githooksType(dir)...)
	m, _ = githooksSend(m, githooksKey("enter"))
	if m.step != 2 || m.repo != dir {
		t.Fatalf("caminho absoluto: passo %d, repo %q; queria passo 2 e %q", m.step, m.repo, dir)
	}
}

func TestGitHooksPresetComecaNoPasso2(t *testing.T) {
	dir := t.TempDir()
	m := githooksNew(t, &githooksFake{}, app.GitHooksRequest{Repo: dir})
	if m.step != 2 {
		t.Fatalf("preset com repositório começou no passo %d, queria 2", m.step)
	}
}

func TestGitHooksTeclasDeTextoSaoLetras(t *testing.T) {
	m := githooksNew(t, &githooksFake{}, app.GitHooksRequest{})
	m, _ = githooksSend(m, githooksType("q?snmt")...)
	if got := m.path.Value(); got != "q?snmt" {
		t.Fatalf("campo de caminho = %q, queria o texto digitado", got)
	}
	if m.help {
		t.Fatalf("? abriu ajuda dentro do campo de texto")
	}
}

func TestGitHooksFluxoGrava(t *testing.T) {
	dir := t.TempDir()
	f := &githooksFake{plan: githooksPlanFake()}
	m := githooksNew(t, f, app.GitHooksRequest{Repo: dir, Hooks: []string{"gex-pre-commit"}})
	m, _ = githooksSend(m, githooksKey("enter"))
	m, _ = githooksSend(m, githooksType("make test")...)
	m, cmd := githooksSend(m, githooksKey("ctrl+s"))
	if m.phase != ghPlanning {
		t.Fatalf("fase após ctrl+s = %d, queria planejamento", m.phase)
	}
	m, _ = githooksSend(m, githooksRun(cmd)...)
	if got := m.preview.GetContent(); !strings.Contains(got, "make test") || !strings.Contains(got, "Makefile:3") {
		t.Fatalf("prévia sem o comando ou a origem: %q", got)
	}
	m, _ = githooksSend(m, githooksKey("enter"))
	if !m.asking {
		t.Fatalf("enter não abriu a confirmação")
	}
	m, cmd = githooksSend(m, githooksKey("s"))
	if !m.Busy() {
		t.Fatalf("Busy() = false durante a gravação")
	}
	m, _ = githooksSend(m, githooksRun(cmd)...)
	if m.Busy() {
		t.Fatalf("Busy() = true depois da gravação")
	}
	if !strings.Contains(githooksView(m), "abc123") || len(f.applied) != 1 {
		t.Fatalf("resultado sem ID ou gravação não chamada: aplicados=%d", len(f.applied))
	}
}

func TestGitHooksPlanRecusaNaoDocumentado(t *testing.T) {
	dir := t.TempDir()
	f := &githooksFake{planErr: githooks.ErrUndocumented}
	m := githooksNew(t, f, app.GitHooksRequest{Repo: dir, Hooks: []string{"gex-pre-commit"}})
	m, _ = githooksSend(m, githooksKey("enter"))
	m, _ = githooksSend(m, githooksType("rm -rf x")...)
	m, cmd := githooksSend(m, githooksKey("ctrl+s"))
	m, _ = githooksSend(m, githooksRun(cmd)...)
	if !strings.Contains(githooksView(m), "documentado") {
		t.Fatalf("erro de comando não documentado não apareceu")
	}
}

func TestGitHooksTinyNaoDecide(t *testing.T) {
	m := githooksNew(t, &githooksFake{}, app.GitHooksRequest{})
	m, _ = githooksSend(m, tea.WindowSizeMsg{Width: 30, Height: 10})
	m, _ = githooksSend(m, githooksType(t.TempDir())...)
	m, _ = githooksSend(m, githooksKey("enter"))
	if m.step != 1 {
		t.Fatalf("terminal pequeno avançou para o passo %d", m.step)
	}
}

func TestGitHooksCtrlCNuncaUsaInterrupt(t *testing.T) {
	m := githooksNew(t, &githooksFake{}, app.GitHooksRequest{})
	_, cmd := githooksSend(m, githooksKey("ctrl+c"))
	if cmd == nil {
		t.Fatalf("ctrl+c não devolveu comando de saída")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("ctrl+c não devolveu tea.QuitMsg")
	}
}

func TestGitHooksSessentaPorVinte(t *testing.T) {
	m := githooksNew(t, &githooksFake{}, app.GitHooksRequest{})
	m, _ = githooksSend(m, tea.WindowSizeMsg{Width: 60, Height: 20})
	lines := strings.Split(githooksView(m), "\n")
	if len(lines) > 20 {
		t.Fatalf("tela 60×20 tem %d linhas", len(lines))
	}
	for _, line := range lines {
		if w := ansi.StringWidth(line); w > 60 {
			t.Fatalf("linha com %d colunas em 60×20: %q", w, line)
		}
	}
}

func TestGitHooksCheckComandos(t *testing.T) {
	cases := []struct {
		nome     string
		lines    []string
		required bool
		want     bool
	}{
		{"vazio opcional", nil, false, true},
		{"vazio obrigatório", nil, true, false},
		{"oito comandos", make([]string, 8), false, true},
		{"nove comandos", make([]string, 9), false, false},
		{"comando longo", []string{strings.Repeat("a", ghMaxCommandLen+1)}, false, false},
	}
	for _, c := range cases {
		got := githooksCheck("pre-commit", c.lines, c.required) == ""
		if got != c.want {
			t.Fatalf("%s: aceito=%v, queria %v", c.nome, got, c.want)
		}
	}
}

func TestGitHooksLinhasIgnoramVazias(t *testing.T) {
	got := githooksLines("  make test \n\n go vet\n")
	if len(got) != 2 || got[0] != "make test" || got[1] != "go vet" {
		t.Fatalf("githooksLines = %q", got)
	}
}

func TestGitHooksTeatestTelaPequenaSemDecisao(t *testing.T) {
	m := githooksNew(t, &githooksFake{}, app.GitHooksRequest{})
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))
	tm.Type("relativo")
	tm.Send(githooksKey("enter"))
	tm.Send(githooksKey("ctrl+c"))
	final := tm.FinalModel(t, teatest.WithFinalTimeout(5*time.Second))
	got, ok := final.(githooksModel)
	if !ok || got.step != 1 {
		t.Fatalf("teatest: modelo final inesperado: %#v", final)
	}
}
