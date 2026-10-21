package screens

import (
	"errors"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
)

type Runner struct {
	In      *os.File
	Out     *os.File
	Environ []string
	Profile colorprofile.Profile
	Signals <-chan os.Signal
}

func NotifySignals() (<-chan os.Signal, func()) {
	return nil, func() {}
}

func (r Runner) Run(m tea.Model) (tea.Model, error) {
	return m, errors.New("não implementado")
}
