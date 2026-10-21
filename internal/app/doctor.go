package app

import "errors"

func Doctor(home string, environ []string, executable string) (Report, error) {
	return Report{}, errors.New("não implementado")
}

func DoctorRepo(home, repo, executable string) (Report, error) {
	return Report{}, errors.New("não implementado")
}

func (r Report) Counts() (ok, attention, pending int) {
	return 0, 0, 0
}

func (r Report) ExitCode() int {
	return 0
}
