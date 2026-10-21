package githooks

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func Resolve(path string, git func(dir string, args ...string) (string, error)) (Repo, error) {
	bare, err := git(path, "rev-parse", "--is-bare-repository")
	if err != nil {
		return Repo{}, ErrNotRepo
	}
	if strings.TrimSpace(bare) == "true" {
		return Repo{}, ErrBare
	}
	top, err := git(path, "rev-parse", "--show-toplevel")
	if err != nil {
		return Repo{}, ErrNotRepo
	}
	common, err := git(path, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return Repo{}, ErrNotRepo
	}
	if hooksPath, err := git(path, "config", "--get", "core.hooksPath"); err == nil && strings.TrimSpace(hooksPath) != "" {
		return Repo{}, ErrHooksPath
	}
	return Repo{Root: NormalizeGitPath(top), CommonDir: NormalizeGitPath(common)}, nil
}

func NormalizeGitPath(out string) string {
	return filepath.Clean(filepath.FromSlash(strings.TrimSpace(out)))
}

func Conflicts(repo Repo, hooks []string, managed map[string]string) ([]string, error) {
	if repo.CommonDir == "" {
		return nil, ErrNotRepo
	}
	var conflicts []string
	for _, name := range hooks {
		data, err := os.ReadFile(filepath.Join(repo.CommonDir, "hooks", name))
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if want, ok := managed[name]; ok && want == string(data) {
			continue
		}
		conflicts = append(conflicts, name)
	}
	return conflicts, nil
}
