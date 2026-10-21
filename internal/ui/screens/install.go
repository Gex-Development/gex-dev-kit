package screens

import (
	"errors"

	"github.com/Gex-Development/gex-dev-kit/internal/app"
)

type InstallDeps struct {
	Options func() []app.OptionGroup
	Plan    func(app.Request) (app.Plan, error)
	Apply   func(app.Plan, func(app.Step)) (app.Result, error)
}

func RunInstall(r Runner, c Common, d InstallDeps, preset app.Request, wizard bool) (app.Result, error) {
	return app.Result{}, errors.New("tela não implementada")
}
