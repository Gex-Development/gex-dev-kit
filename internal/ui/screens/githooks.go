package screens

import (
	"errors"

	"github.com/Gex-Development/gex-dev-kit/internal/app"
)

type GitHooksDeps struct {
	Plan  func(app.GitHooksRequest) (app.GitHooksPlan, error)
	Apply func(app.GitHooksPlan) (app.Result, error)
}

func RunGitHooks(r Runner, c Common, d GitHooksDeps, preset app.GitHooksRequest) (app.Result, error) {
	return app.Result{}, errors.New("tela não implementada")
}
