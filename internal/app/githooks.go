package app

import "errors"

func PlanGitHooks(r GitHooksRequest, diffLimit int) (GitHooksPlan, error) {
	return GitHooksPlan{}, errors.New("não implementado")
}

func ApplyGitHooks(p GitHooksPlan) (Result, error) {
	return Result{}, errors.New("não implementado")
}
