package screens

import (
	"errors"

	"github.com/Gex-Development/gex-dev-kit/internal/app"
)

type HomeChoice string

const (
	ChoiceInstall        HomeChoice = "install"
	ChoiceCatalog        HomeChoice = "catalog"
	ChoiceDoctor         HomeChoice = "doctor"
	ChoiceRollback       HomeChoice = "rollback"
	ChoiceRollbackAgents HomeChoice = "rollback-agents"
	ChoiceAddons         HomeChoice = "addons"
	ChoiceGitHooks       HomeChoice = "githooks"
	ChoiceShowcase       HomeChoice = "ds"
	ChoiceHelp           HomeChoice = "help"
	ChoiceExit           HomeChoice = "exit"
)

type HomeDeps struct {
	Core   func() (app.CoreStatus, error)
	Agents func() ([]app.AgentChange, error)
}

func RunHome(r Runner, c Common, d HomeDeps) (HomeChoice, error) {
	return ChoiceExit, errors.New("tela não implementada")
}
