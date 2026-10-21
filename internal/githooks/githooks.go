package githooks

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/Gex-Development/gex-dev-kit/internal/catalog"
)

var (
	ErrNotRepo      = errors.New("não é repositório Git")
	ErrBare         = errors.New("repositório bare não tem hooks de trabalho")
	ErrHooksPath    = errors.New("core.hooksPath definido")
	ErrUndocumented = errors.New("comando não documentado pelo projeto")
)

type Repo struct {
	Root      string
	CommonDir string
}

type Request struct {
	Hooks     []string
	PreCommit []string
	PrePush   []string
	Binary    string
}

type Source struct {
	Command string
	File    string
	Line    int
}

func Resolve(path string, git func(dir string, args ...string) (string, error)) (Repo, error) {
	return Repo{}, errors.New("não implementado")
}

func NormalizeGitPath(out string) string {
	return filepath.Clean(filepath.FromSlash(strings.TrimSpace(out)))
}

func Validate(r Request) error {
	return errors.New("não implementado")
}

func Documented(root, command string) (Source, bool) {
	return Source{}, false
}

func Script(hook string, r Request) (string, error) {
	return "", errors.New("não implementado")
}

func Files(r Request) ([]catalog.File, error) {
	return nil, errors.New("não implementado")
}

func Conflicts(repo Repo, hooks []string, managed map[string]string) ([]string, error) {
	return nil, errors.New("não implementado")
}
