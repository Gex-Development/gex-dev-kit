package githooks

import "errors"

var (
	ErrNotRepo      = errors.New("não é repositório Git")
	ErrBare         = errors.New("repositório bare não tem hooks de trabalho")
	ErrHooksPath    = errors.New("core.hooksPath definido")
	ErrUndocumented = errors.New("comando não documentado pelo projeto")

	ErrNoHook               = errors.New("nenhum git hook selecionado")
	ErrUnknownHook          = errors.New("git hook desconhecido")
	ErrPrePushWithoutHook   = errors.New("comandos de pre-push sem o hook pre-push selecionado")
	ErrPreCommitWithoutHook = errors.New("comandos de pre-commit sem o hook pre-commit selecionado")
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
