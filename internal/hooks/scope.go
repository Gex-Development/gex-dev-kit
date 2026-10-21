package hooks

import "errors"

type GitFunc func(dir string, args ...string) (string, error)

var DefaultOrgs = []string{"Gex-Development", "Gex-Integrations", "gex-data-architecture-org", "gex-data-architecture"}

type Scope struct {
	Orgs []string
	Git  GitFunc
}

func NewScope(configDir string, git GitFunc) Scope {
	return Scope{Orgs: append([]string(nil), DefaultOrgs...), Git: git}
}

func (s Scope) InGEX(dir string) (bool, error) {
	return false, errors.New("não implementado")
}

func (s Scope) Any(dirs []string) bool {
	return false
}

func OrgFromURL(url string) (string, bool) {
	return "", false
}

func RunGit(dir string, args ...string) (string, error) {
	return "", errors.New("não implementado")
}
