package app

import "errors"

func PlanInstall(r Request, diffLimit int) (Plan, error) {
	return Plan{}, errors.New("não implementado")
}

func (p Plan) Empty() bool {
	return len(p.Files) == 0 && len(p.Updates) == 0
}

func ApplyInstall(p Plan, progress func(Step)) (Result, error) {
	return Result{}, errors.New("não implementado")
}

func EquivalentCommand(r Request) string {
	return ""
}
