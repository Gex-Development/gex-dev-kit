package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Gex-Development/gex-dev-kit/internal/catalog"
	"github.com/Gex-Development/gex-dev-kit/internal/githooks"
	"github.com/Gex-Development/gex-dev-kit/internal/hooks"
	"github.com/Gex-Development/gex-dev-kit/internal/kit"
)

func PlanGitHooks(r GitHooksRequest, diffLimit int) (GitHooksPlan, error) {
	if !filepath.IsAbs(r.Home) {
		return GitHooksPlan{}, errors.New("HOME deve ser caminho absoluto")
	}
	hooksSemPrefixo := make([]string, 0, len(r.Hooks))
	for _, name := range r.Hooks {
		hooksSemPrefixo = append(hooksSemPrefixo, strings.TrimPrefix(name, "gex-"))
	}
	req := githooks.Request{Hooks: hooksSemPrefixo, PreCommit: r.PreCommit, PrePush: r.PrePush, Binary: catalog.StableBinary(r.Home)}
	if err := githooks.Validate(req); err != nil {
		return GitHooksPlan{}, err
	}
	scripts, err := gitHookScripts(req)
	if err != nil {
		return GitHooksPlan{}, err
	}
	plan := GitHooksPlan{Request: r, Scripts: scripts}
	if r.Repo == "" {
		return plan, nil
	}
	return planGitHooksRepo(plan, req, diffLimit)
}

func gitHookScripts(req githooks.Request) (map[string]string, error) {
	out := make(map[string]string, len(req.Hooks))
	for _, name := range req.Hooks {
		script, err := githooks.Script(name, req)
		if err != nil {
			return nil, err
		}
		out[name] = script
	}
	return out, nil
}

func planGitHooksRepo(plan GitHooksPlan, req githooks.Request, diffLimit int) (GitHooksPlan, error) {
	if !filepath.IsAbs(plan.Request.Repo) {
		return GitHooksPlan{}, errors.New("--repo deve ser caminho absoluto")
	}
	repo, err := githooks.Resolve(plan.Request.Repo, hooks.RunGit)
	if err != nil {
		return GitHooksPlan{}, err
	}
	sources, err := gitHookSources(repo.Root, req)
	if err != nil {
		return GitHooksPlan{}, err
	}
	managed, err := gitHookManaged(repo.CommonDir)
	if err != nil {
		return GitHooksPlan{}, err
	}
	conflicts, err := githooks.Conflicts(repo, req.Hooks, managed)
	if err != nil {
		return GitHooksPlan{}, err
	}
	if len(conflicts) > 0 {
		return GitHooksPlan{}, fmt.Errorf("hook existente não é do kit ou foi alterado depois da instalação: %s", strings.Join(conflicts, ", "))
	}
	binary, err := catalog.BinaryFile(plan.Request.Executable)
	if err != nil {
		return GitHooksPlan{}, err
	}
	homePlan, err := kit.Preview(plan.Request.Home, []catalog.File{binary})
	if err != nil {
		return GitHooksPlan{}, fmt.Errorf("Prévia do HOME: %w", err)
	}
	files, err := githooks.Files(req)
	if err != nil {
		return GitHooksPlan{}, err
	}
	repoPlan, err := kit.Preview(repo.CommonDir, files)
	if err != nil {
		return GitHooksPlan{}, fmt.Errorf("Prévia do repositório: %w", err)
	}
	plan.Repo = repo
	plan.Sources = sources
	plan.HomeChanges = homePlan.Changes
	plan.RepoChanges = repoPlan.Changes
	plan.Diffs = buildDiffs(append(homePlan.Contents(), repoPlan.Contents()...), diffLimit)
	plan.home = homePlan
	plan.repo = repoPlan
	return plan, nil
}

func gitHookSources(root string, req githooks.Request) ([]githooks.Source, error) {
	var out []githooks.Source
	for _, commands := range [][]string{req.PreCommit, req.PrePush} {
		for _, command := range commands {
			src, ok := githooks.Documented(root, command)
			if !ok {
				return nil, fmt.Errorf("%w: %s; documente o comando no AGENTS.md, README, Makefile ou package.json", githooks.ErrUndocumented, command)
			}
			out = append(out, src)
		}
	}
	return out, nil
}

func gitHookManaged(commonDir string) (map[string]string, error) {
	statuses, err := kit.Doctor(commonDir)
	if err != nil {
		return nil, err
	}
	managed := make(map[string]string)
	for _, s := range statuses {
		name, ok := strings.CutPrefix(s.Path, "hooks/")
		if !ok || s.State != "ok" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(commonDir, "hooks", name))
		if err != nil {
			return nil, err
		}
		managed[name] = string(data)
	}
	return managed, nil
}

func ApplyGitHooks(p GitHooksPlan) (Result, error) {
	if p.Request.Repo == "" {
		return Result{}, errors.New("prévia sem repositório; nada a gravar")
	}
	if _, err := kit.Apply(p.home); err != nil {
		return Result{}, fmt.Errorf("Executável dos hooks no HOME: %w", err)
	}
	id, err := kit.Apply(p.repo)
	if err != nil {
		repoErr := fmt.Errorf("Git hooks do repositório: %w", err)
		if _, rbErr := kit.Rollback(p.Request.Home); rbErr != nil {
			return Result{}, errors.Join(repoErr, fmt.Errorf("desfazer o executável no HOME: %w", rbErr))
		}
		return Result{}, repoErr
	}
	return Result{HistoryID: id}, nil
}
