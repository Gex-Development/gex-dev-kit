package hooks

import (
	"errors"
	"io"
)

type Finding struct {
	Path string
	Line int
	Rule string
}

func (f Finding) Message(allowFile string) string {
	return ""
}

func ScanStaged(diff io.Reader, limit int64) (findings []Finding, truncated bool, err error) {
	return nil, false, errors.New("não implementado")
}
