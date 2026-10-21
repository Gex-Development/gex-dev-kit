package screens

import (
	"charm.land/huh/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/app"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

type Selection struct {
	Home            string
	Skills          []string
	Hooks           []string
	SkipAgentUpdate bool
}

func NewInstallForm(t theme.Theme, s *Selection, groups []app.OptionGroup) *huh.Form {
	return nil
}
