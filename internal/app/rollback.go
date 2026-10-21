package app

import "errors"

func PlanRollback(root string, diffLimit int) (RollbackPreview, error) {
	return RollbackPreview{}, errors.New("não implementado")
}

func Rollback(root string) (string, error) {
	return "", errors.New("não implementado")
}

func RollbackAgents(home string) error {
	return errors.New("não implementado")
}
