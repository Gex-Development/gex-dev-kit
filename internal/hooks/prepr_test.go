package hooks

import (
	"errors"
	"strings"
	"testing"
)

func TestPrePRAsksOnPRCreateVariants(t *testing.T) {
	variants := [][]string{
		{"gh", "pr", "create"},
		{"gh", "pr", "create", "--fill"},
		{"gh", "pr", "create", "-t", "x", "-b", "y"},
		{"env", "GH_TOKEN=x", "gh", "pr", "create"},
		{"sudo", "gh", "pr", "create", "--draft"},
		{"nohup", "gh", "pr", "create"},
	}
	for _, words := range variants {
		got := ClassifyPrePR([]Command{sensitiveCmd(words...)})
		if !got.Ask || got.Category != "abertura de PR" {
			t.Fatalf("%s: esperava pergunta de abertura de PR, veio %+v", strings.Join(words, " "), got)
		}
		for _, item := range []string{"gex-auto-revisao", "testes", "build", "lint", "issue", "assignee", "reviewers"} {
			if !strings.Contains(got.Rule, item) {
				t.Fatalf("checklist sem %q: %q", item, got.Rule)
			}
		}
	}
}

func TestPrePRIgnoresOtherGhCommands(t *testing.T) {
	for _, words := range [][]string{
		{"gh", "pr", "list"},
		{"gh", "pr", "view", "12"},
		{"gh", "pr", "checkout", "12"},
		{"gh", "pr"},
		{"ls"},
		{"git", "push"},
	} {
		if got := ClassifyPrePR([]Command{sensitiveCmd(words...)}); got != (Decision{}) {
			t.Fatalf("%s não deve perguntar no pré-PR: %+v", strings.Join(words, " "), got)
		}
	}
}

func TestPrePRReasonHidesCommandValues(t *testing.T) {
	got := ClassifyPrePR([]Command{sensitiveCmd("gh", "pr", "create", "--title", sensitiveSentinel)})
	if !got.Ask {
		t.Fatalf("esperava pergunta, veio %+v", got)
	}
	if reason := got.Reason(); strings.Contains(reason, sensitiveSentinel) {
		t.Fatalf("motivo vazou valor do comando: %q", reason)
	}
}

func TestPrePRCheckComposesParserAndScope(t *testing.T) {
	cwd := t.TempDir()
	repo := t.TempDir()
	script := sensitiveScript(Command{Words: sensitiveCmd("gh", "pr", "create", "--fill").Words, Dir: repo})
	var seen [][]string
	scope := func(dirs []string) bool {
		seen = append(seen, dirs)
		return true
	}
	in := AgentInput{Tool: "Bash", Command: "gh pr create --fill", Cwd: cwd, Parsed: true}
	got := preprCheck(in, scope, sensitiveParse(script, nil))
	if !got.Ask || got.Category != "abertura de PR" {
		t.Fatalf("esperava pergunta de pré-PR, veio %+v", got)
	}
	if len(seen) != 1 || !sensitiveContains(seen[0], cwd) || !sensitiveContains(seen[0], repo) {
		t.Fatalf("escopo deveria receber cwd e diretório do comando, recebeu %v", seen)
	}
}

func TestPrePRCheckSkipsUnparsedOrOutOfScope(t *testing.T) {
	cwd := t.TempDir()
	inScope := func([]string) bool { return true }
	outside := func([]string) bool { return false }
	prCreate := sensitiveParse(sensitiveScript(sensitiveCmd("gh", "pr", "create")), nil)
	cases := []struct {
		name  string
		in    AgentInput
		scope func([]string) bool
		parse func(string) (Script, error)
	}{
		{"fora do escopo", AgentInput{Tool: "Bash", Command: "gh pr create", Cwd: cwd, Parsed: true}, outside, prCreate},
		{"ilegível", AgentInput{Tool: "Bash", Cwd: cwd}, inScope, prCreate},
		{"acima do limite", AgentInput{Tool: "Bash", Cwd: cwd, Parsed: true, Truncated: true}, inScope, prCreate},
		{"parser recusou", AgentInput{Tool: "Bash", Command: "gh pr create", Cwd: cwd, Parsed: true}, inScope, sensitiveParse(Script{}, errors.New("falha"))},
		{"ferramenta Read", AgentInput{Tool: "Read", FilePath: "x", Cwd: cwd, Parsed: true}, inScope, prCreate},
	}
	for _, c := range cases {
		if got := preprCheck(c.in, c.scope, c.parse); got != (Decision{}) {
			t.Fatalf("%s: esperava vazio, veio %+v", c.name, got)
		}
	}
}

func TestCheckPrePRIgnoresTruncatedInput(t *testing.T) {
	in := AgentInput{Tool: "Bash", Command: "gh pr create", Cwd: t.TempDir(), Parsed: true, Truncated: true}
	if got := CheckPrePR(in, func([]string) bool { return true }); got != (Decision{}) {
		t.Fatalf("entrada grande não deve perguntar no pré-PR: %+v", got)
	}
}
