package app

import (
	"github.com/Gex-Development/gex-dev-kit/internal/agents"
	"github.com/Gex-Development/gex-dev-kit/internal/kit"
)

func PlanRollback(root string, diffLimit int) (RollbackPreview, error) {
	plan, err := kit.RollbackPreview(root)
	if err != nil {
		return RollbackPreview{}, err
	}
	return RollbackPreview{Plan: plan, Diffs: buildRollbackDiffs(plan, diffLimit)}, nil
}

func Rollback(root string) (string, error) {
	return kit.Rollback(root)
}

func RepoHistoryRoot(repo string) (string, error) {
	target, err := doctorCommonDir(repo)
	if err != nil {
		return "", err
	}
	return target.CommonDir, nil
}

func RollbackAgents(home string) error {
	return agents.Rollback(home)
}
