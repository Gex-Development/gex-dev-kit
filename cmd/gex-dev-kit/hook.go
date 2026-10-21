package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Gex-Development/gex-dev-kit/internal/hooks"
)

const (
	hookSessionDeadline    = 2 * time.Second
	hookScanLimit          = 8 << 20
	hookSessionUnavailable = "GEX: verificação de início de sessão indisponível."
	hookCommitBlocked      = "Commit bloqueado: mensagem fora do padrão Conventional Commits. Use git commit -m 'tipo: descrição'."
	hookCommitUnparsed     = "Commit bloqueado: não foi possível analisar o commit; use `git commit -m 'tipo: descrição'` literal."
)

func runHook(args []string, in io.Reader, out, errOut io.Writer) int {
	if len(args) == 0 {
		return invalid(errOut, "uso: hook commit-msg|sensitive-action|pre-pr|session-start|pre-commit")
	}
	switch args[0] {
	case "commit-msg":
		return hookCommitMsg(args[1:], in, errOut)
	case "sensitive-action":
		return hookSensitive(in, out)
	case "pre-pr":
		return hookPrePR(in, out)
	case "session-start":
		return hookSessionStart(args[1:], in, out, errOut)
	case "pre-commit":
		return hookPreCommit(errOut)
	}
	return invalid(errOut, "hook desconhecido: "+args[0])
}

func hookScope() hooks.Scope {
	configDir, _ := os.UserConfigDir()
	return hooks.NewScope(configDir, nil)
}

func hookInScope(scope hooks.Scope, dirs []string) bool {
	if wd, err := os.Getwd(); err == nil {
		dirs = append(dirs, wd)
	}
	return scope.Any(dirs)
}

func hookCommitMsg(args []string, in io.Reader, errOut io.Writer) int {
	if len(args) == 1 {
		return hookCommitFile(args[0], errOut)
	}
	if len(args) != 0 {
		return invalid(errOut, "uso: hook commit-msg [CAMINHO]")
	}
	return hookCommitAgent(in, errOut)
}

func hookCommitAgent(in io.Reader, errOut io.Writer) int {
	input := hooks.ReadAgentInput(in)
	status, dirs := hooks.ClassifyCommit(input)
	if status != hooks.CommitBad {
		return 0
	}
	scope := hookScope()
	if !hookInScope(scope, dirs) {
		return 0
	}
	if !input.Parsed {
		fmt.Fprintln(errOut, hookCommitUnparsed)
		return 2
	}
	fmt.Fprintln(errOut, hookCommitBlocked)
	return 2
}

func hookCommitFile(path string, errOut io.Writer) int {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(errOut, "Commit bloqueado: não foi possível ler a mensagem do commit.")
		return 2
	}
	message := hooks.CleanGitMessage(string(data), hookCommentChar(filepath.Dir(path)))
	if err := hooks.CheckGitMessage(message); err != nil {
		fmt.Fprintln(errOut, hookCommitBlocked)
		return 2
	}
	return 0
}

func hookCommentChar(dir string) string {
	out, err := hooks.RunGit(dir, "config", "--get", "core.commentChar")
	value := strings.TrimSpace(out)
	if err != nil || value == "" {
		return "#"
	}
	return value
}

func hookSensitive(in io.Reader, out io.Writer) int {
	input := hooks.ReadAgentInput(in)
	scope := hookScope()
	decision := hooks.CheckSensitive(input, func(dirs []string) bool { return hookInScope(scope, dirs) }, hookPRInstalled)
	return hookAsk(out, decision)
}

func hookPrePR(in io.Reader, out io.Writer) int {
	input := hooks.ReadAgentInput(in)
	scope := hookScope()
	decision := hooks.CheckPrePR(input, func(dirs []string) bool { return hookInScope(scope, dirs) })
	return hookAsk(out, decision)
}

func hookAsk(out io.Writer, decision hooks.Decision) int {
	if decision.Ask {
		out.Write(hooks.AskOutput(decision.Reason()))
	}
	return 0
}

func hookPRInstalled() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	data, err := os.ReadFile(filepath.Join(home, ".claude", "settings.json"))
	return err == nil && bytes.Contains(data, []byte(`"pre-pr"`))
}

func hookSessionStart(args []string, in io.Reader, out, errOut io.Writer) int {
	if len(args) != 2 || args[0] != "--agent" {
		return invalid(errOut, "uso: hook session-start --agent claude|codex")
	}
	agent := args[1]
	if agent != "claude" && agent != "codex" {
		return invalid(errOut, "agente inválido: use claude ou codex")
	}
	input := hooks.ReadAgentInput(in)
	fmt.Fprintln(out, hookSessionText(input, agent))
	return 0
}

func hookSessionText(input hooks.AgentInput, agent string) string {
	result := make(chan string, 1)
	go func() {
		home, _ := os.UserHomeDir()
		scope := hookScope()
		facts := hooks.GatherSession(input, agent, home, scope.InGEX, hooks.RunGit)
		result <- hooks.SessionContext(facts)
	}()
	select {
	case text := <-result:
		return text
	case <-time.After(hookSessionDeadline):
		return hookSessionUnavailable
	}
}

func hookPreCommit(errOut io.Writer) int {
	diff, err := hooks.RunGit("", "diff", "--cached", "-U0", "--no-color", "--no-ext-diff", "--diff-filter=ACMR")
	if err != nil {
		fmt.Fprintln(errOut, "pre-commit: não foi possível ler o diff staged; verificação ignorada.")
		return 0
	}
	findings, truncated, err := hooks.ScanStaged(strings.NewReader(diff), hookScanLimit)
	if err != nil {
		fmt.Fprintln(errOut, "pre-commit: não foi possível analisar o diff; verificação ignorada.")
		return 0
	}
	if truncated {
		fmt.Fprintln(errOut, "pre-commit: diff grande demais para verificar; verificação ignorada.")
		return 0
	}
	commonDir, err := hookCommonDir()
	if err != nil {
		fmt.Fprintln(errOut, "pre-commit: não foi possível localizar o repositório; verificação ignorada.")
		return 0
	}
	allowPath := hooks.AllowPath(commonDir)
	allow, err := hooks.LoadAllow(allowPath)
	if err != nil {
		fmt.Fprintln(errOut, "pre-commit: não foi possível ler o arquivo de exceções.")
		return 1
	}
	findings = allow.Filter(findings)
	allowLabel := hookHomeLabel(allowPath)
	for _, finding := range findings {
		fmt.Fprintln(errOut, finding.Message(allowLabel))
	}
	if len(findings) > 0 {
		return 1
	}
	return 0
}

func hookCommonDir() (string, error) {
	out, err := hooks.RunGit("", "rev-parse", "--path-format=absolute", "--git-common-dir")
	dir := strings.TrimSpace(out)
	if err != nil {
		return "", err
	}
	if dir == "" {
		return "", fmt.Errorf("diretório comum do Git vazio")
	}
	return dir, nil
}

func hookHomeLabel(path string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return path
	}
	if rest, ok := strings.CutPrefix(path, home+string(filepath.Separator)); ok {
		return "~/" + filepath.ToSlash(rest)
	}
	return path
}
