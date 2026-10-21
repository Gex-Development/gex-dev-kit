package screens

import (
	"errors"

	"github.com/Gex-Development/gex-dev-kit/internal/app"
)

type RollbackKind int

const (
	RollbackCore RollbackKind = iota
	RollbackRepo
	RollbackAgents
)

type RollbackDeps struct {
	Plan  func() (app.RollbackPreview, error)
	Apply func() (string, error)
}

func RunRollback(r Runner, c Common, kind RollbackKind, d RollbackDeps) (string, error) {
	return "", errors.New("tela não implementada")
}
