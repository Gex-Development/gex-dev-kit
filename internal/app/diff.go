package app

import "github.com/Gex-Development/gex-dev-kit/internal/kit"

func buildDiffs(contents []kit.FileContent, limit int) []FileDiff {
	return nil
}

func buildRollbackDiffs(p RollbackPlan, limit int) []FileDiff {
	return nil
}

func maskLine(line string) (string, int) {
	return line, 0
}
