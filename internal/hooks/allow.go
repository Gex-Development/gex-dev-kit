package hooks

import "errors"

type Allow map[string]bool

func AllowPath(commonDir string) string {
	return ""
}

func LoadAllow(path string) (Allow, error) {
	return nil, errors.New("não implementado")
}

func (a Allow) Filter(findings []Finding) []Finding {
	return findings
}
