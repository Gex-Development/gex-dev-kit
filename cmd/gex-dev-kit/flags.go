package main

import (
	"errors"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/terminal"
)

type options struct {
	UI              terminal.Mode
	NoAnim, Quiet   bool
	JSON, Yes       bool
	SkipAgentUpdate bool
	Home, Repo      string
	Skills, Hooks   []string
	PreCommit       []string
	PrePush         []string
	Args            []string
}

func parseOptions(command string, args []string) (options, error) {
	return options{Args: args}, errors.New("não implementado")
}
