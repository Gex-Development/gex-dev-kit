package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/Gex-Development/gex-dev-kit/internal/hooks"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	switch {
	case len(args) == 3 && args[0] == "hook" && args[1] == "commit-msg":
		return commitMsg(args[2])
	case len(args) == 2 && args[0] == "hook" && args[1] == "pre-commit":
		return preCommit()
	}
	fmt.Fprintln(os.Stderr, "uso: hookbin hook commit-msg ARQUIVO | hook pre-commit")
	return 2
}

func commitMsg(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "não foi possível ler a mensagem do commit")
		return 2
	}
	message := hooks.CleanGitMessage(string(data), commentChar())
	if err := hooks.CheckGitMessage(message); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	return 0
}

func preCommit() int {
	commonDir, err := gitText("rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		fmt.Fprintln(os.Stderr, "aviso: git falhou; varredura pulada")
		return 0
	}
	diff, err := gitBytes("diff", "--cached", "-U0", "--no-color", "--no-ext-diff", "--diff-filter=ACMR")
	if err != nil {
		fmt.Fprintln(os.Stderr, "aviso: git falhou; varredura pulada")
		return 0
	}
	findings, truncated, err := hooks.ScanStaged(bytes.NewReader(diff), hooks.MaxInput)
	if err != nil || truncated {
		fmt.Fprintln(os.Stderr, "aviso: diff acima de 8 MiB; varredura pulada")
		return 0
	}
	allowFile := hooks.AllowPath(commonDir)
	allow, err := hooks.LoadAllow(allowFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "não foi possível ler o arquivo de exceções")
		return 1
	}
	findings = allow.Filter(findings)
	for _, finding := range findings {
		fmt.Fprintln(os.Stderr, finding.Message(allowFile))
	}
	if len(findings) > 0 {
		return 1
	}
	return 0
}

func commentChar() string {
	value, err := gitText("config", "--get", "core.commentChar")
	if err != nil || value == "" || value == "auto" {
		return "#"
	}
	return value
}

func gitText(args ...string) (string, error) {
	data, err := gitBytes(args...)
	return strings.TrimSpace(string(data)), err
}

func gitBytes(args ...string) ([]byte, error) {
	return exec.Command("git", args...).Output()
}
