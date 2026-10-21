package githooks

import "errors"

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
