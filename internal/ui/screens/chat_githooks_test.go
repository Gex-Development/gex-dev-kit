package screens

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/app"
	"github.com/Gex-Development/gex-dev-kit/internal/githooks"
)

func TestChatGitHooksSendsDocumentedPreCommitCommand(t *testing.T) {
	var got app.GitHooksRequest
	applied := 0
	m := chatTestModel(false)
	m.deps.GitHooks = GitHooksDeps{Plan: func(req app.GitHooksRequest) (app.GitHooksPlan, error) {
		got = req
		return app.GitHooksPlan{Request: req}, nil
	}, Apply: func(app.GitHooksPlan) (app.Result, error) { applied++; return app.Result{HistoryID: "h-1"}, nil }}
	m.gitHooksRequest = app.GitHooksRequest{Home: "/home/demo", Repo: "/tmp/repo", Executable: "/tmp/gexkit"}
	m = m.flowGitHooks()
	m = chatSetupQuestion(t, m, "Quais git hooks")
	m = chatSetupChoose(t, m, "gex-pre-commit")
	m = chatSetupQuestion(t, m, "Comandos do pre-commit")
	m = chatGitHooksAnswer(t, m, "make test")
	m = chatSetupQuestion(t, m, "Gravo esses hooks?")
	if applied != 0 {
		t.Fatal("git hooks gravados antes da confirmação")
	}
	m = chatSetupChoose(t, m, "yes")
	m = chatSetupQuestion(t, m, "")
	if applied != 1 || len(got.PreCommit) != 1 || got.PreCommit[0] != "make test" {
		t.Fatalf("comando não chegou ao plano: %+v, aplicou %d", got, applied)
	}
}

func TestChatGitHooksRejectsNineCommands(t *testing.T) {
	m := chatTestModel(false)
	m.gitHooksRequest = app.GitHooksRequest{Home: "/home/demo", Repo: "/tmp/repo"}
	m = m.flowGitHooks()
	m = chatSetupQuestion(t, m, "Quais git hooks")
	m = chatSetupChoose(t, m, "gex-pre-commit")
	m = chatSetupQuestion(t, m, "Comandos do pre-commit")
	m = chatGitHooksAnswer(t, m, strings.Repeat("make test\n", 9))
	m = chatSetupQuestion(t, m, "Comandos do pre-commit")
	if !strings.Contains(chatTestTranscript(m), "no máximo 8 comandos") {
		t.Fatal("nove comandos devem falhar e repetir pergunta")
	}
}

func TestChatGitHooksExplainsUndocumentedCommandAndRetries(t *testing.T) {
	calls := 0
	m := chatTestModel(false)
	m.deps.GitHooks = GitHooksDeps{Plan: func(req app.GitHooksRequest) (app.GitHooksPlan, error) {
		calls++
		if calls == 1 {
			return app.GitHooksPlan{}, githooks.ErrUndocumented
		}
		return app.GitHooksPlan{Request: req}, nil
	}}
	m.gitHooksRequest = app.GitHooksRequest{Home: "/home/demo", Repo: "/tmp/repo"}
	m = m.flowGitHooks()
	m = chatSetupQuestion(t, m, "Quais git hooks")
	m = chatSetupChoose(t, m, "gex-pre-commit")
	m = chatSetupQuestion(t, m, "Comandos do pre-commit")
	m = chatGitHooksAnswer(t, m, "make test")
	m = chatSetupQuestion(t, m, "O que eu faço?")
	if !strings.Contains(chatTestTranscript(m), "Comando não documentado") {
		t.Fatal("erro de comando não documentado sumiu da conversa")
	}
	found := false
	for _, entry := range m.entries {
		if entry.isTool && strings.Contains(strings.Join(entry.tool.Lines, " "), "README") {
			found = true
		}
	}
	if !found {
		t.Fatal("ferramenta não explicou onde documentar o comando")
	}
	m = chatSetupChoose(t, m, "retry")
	m = chatSetupQuestion(t, m, "Gravo esses hooks?")
	if calls != 2 {
		t.Fatalf("retry planejou %d vezes; queria duas", calls)
	}
}

func chatGitHooksAnswer(t *testing.T, m chatModel, text string) chatModel {
	t.Helper()
	m.area.SetValue(text)
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyEnter})
	return m
}
