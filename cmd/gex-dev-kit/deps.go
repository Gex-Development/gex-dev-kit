package main

import "github.com/Gex-Development/gex-dev-kit/internal/app"

type deps struct {
	InstallOptions func() []app.OptionGroup
	PlanInstall    func(app.Request, int) (app.Plan, error)
	ApplyInstall   func(app.Plan, func(app.Step)) (app.Result, error)
	Doctor         func(home string, environ []string, executable string) (app.Report, error)
	DoctorRepo     func(home, repo, executable string) (app.Report, error)
	PlanRollback   func(root string, limit int) (app.RollbackPreview, error)
	Rollback       func(root string) (string, error)
	RollbackAgents func(home string) error
	PlanGitHooks   func(app.GitHooksRequest, int) (app.GitHooksPlan, error)
	ApplyGitHooks  func(app.GitHooksPlan) (app.Result, error)
}

func realDeps() deps {
	return deps{
		InstallOptions: app.InstallOptions,
		PlanInstall:    app.PlanInstall,
		ApplyInstall:   app.ApplyInstall,
		Doctor:         app.Doctor,
		DoctorRepo:     app.DoctorRepo,
		PlanRollback:   app.PlanRollback,
		Rollback:       app.Rollback,
		RollbackAgents: app.RollbackAgents,
		PlanGitHooks:   app.PlanGitHooks,
		ApplyGitHooks:  app.ApplyGitHooks,
	}
}
