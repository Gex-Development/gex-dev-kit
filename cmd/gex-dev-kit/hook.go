package main

import (
	"fmt"
	"io"
	"os"

	"github.com/Gex-Development/gex-dev-kit/internal/hooks"
)

func runHook(args []string, in io.Reader, out, errOut io.Writer) int {
	if len(args) < 1 || len(args) > 2 || args[0] != "commit-msg" {
		return invalid(errOut, "uso: hook commit-msg [CAMINHO]")
	}
	var err error
	if len(args) == 2 {
		var message []byte
		message, err = os.ReadFile(args[1])
		if err == nil {
			err = hooks.CheckGitMessage(hooks.CleanGitMessage(string(message), "#"))
		}
	} else {
		err = hooks.CheckAgentCommit(in)
	}
	if err != nil {
		fmt.Fprintln(errOut, "Commit bloqueado: mensagem ausente ou fora do padrão Conventional Commits.")
		return 2
	}
	return 0
}
