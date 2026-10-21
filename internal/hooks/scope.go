package hooks

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const scopeGitDeadline = 2 * time.Second

const scopeConfigKeys = `^(remote\..*\.url|gex\.policy)$`

var errScopeGitTimeout = errors.New("git não respondeu no prazo")

type GitFunc func(dir string, args ...string) (string, error)

var DefaultOrgs = []string{"Gex-Development", "Gex-Integrations", "gex-data-architecture-org", "gex-data-architecture"}

type Scope struct {
	Orgs []string
	Git  GitFunc
}

func NewScope(configDir string, git GitFunc) Scope {
	orgs := append([]string(nil), DefaultOrgs...)
	if configDir != "" {
		if data, err := os.ReadFile(filepath.Join(configDir, "gex-dev-kit", "orgs")); err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				org := strings.TrimSpace(line)
				if scopeValidOrg(org) && !scopeHasOrg(orgs, org) {
					orgs = append(orgs, org)
				}
			}
		}
	}
	if git == nil {
		git = RunGit
	}
	return Scope{Orgs: orgs, Git: git}
}

func (s Scope) InGEX(dir string) (bool, error) {
	out, err := s.run(dir, "config", "--get-regexp", scopeConfigKeys)
	if err != nil {
		return false, err
	}
	policy := ""
	inGEX := false
	for _, line := range strings.Split(out, "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(line), " ")
		if !ok {
			continue
		}
		switch {
		case key == "gex.policy":
			policy = strings.ToLower(strings.TrimSpace(value))
		case strings.HasPrefix(key, "remote.") && strings.HasSuffix(key, ".url"):
			if org, ok := OrgFromURL(value); ok && scopeHasOrg(s.Orgs, org) {
				inGEX = true
			}
		}
	}
	switch policy {
	case "true":
		return true, nil
	case "false":
		return false, nil
	}
	return inGEX, nil
}

func (s Scope) Any(dirs []string) bool {
	for _, dir := range dirs {
		if ok, err := s.InGEX(dir); err == nil && ok {
			return true
		}
	}
	return false
}

func (s Scope) run(dir string, args ...string) (string, error) {
	git := s.Git
	if git == nil {
		git = RunGit
	}
	type result struct {
		out string
		err error
	}
	done := make(chan result, 1)
	go func() {
		out, err := git(dir, args...)
		done <- result{out, err}
	}()
	timer := time.NewTimer(scopeGitDeadline)
	defer timer.Stop()
	select {
	case r := <-done:
		return r.out, r.err
	case <-timer.C:
		return "", errScopeGitTimeout
	}
}

func OrgFromURL(url string) (string, bool) {
	host, path, ok := scopeSplitRemote(url)
	if !ok || !strings.EqualFold(host, "github.com") {
		return "", false
	}
	org, repo, found := strings.Cut(strings.Trim(path, "/"), "/")
	if !found || org == "" || repo == "" {
		return "", false
	}
	return org, true
}

func RunGit(dir string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), scopeGitDeadline)
	defer cancel()
	if dir != "" {
		args = append([]string{"-C", dir}, args...)
	}
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Env = append(os.Environ(), "GIT_OPTIONAL_LOCKS=0")
	out, err := cmd.Output()
	return string(out), err
}

func scopeSplitRemote(url string) (string, string, bool) {
	if scheme, rest, found := strings.Cut(url, "://"); found {
		switch strings.ToLower(scheme) {
		case "https", "http", "ssh", "git+ssh":
		default:
			return "", "", false
		}
		authority, path, found := strings.Cut(rest, "/")
		if !found {
			return "", "", false
		}
		return scopeHostOf(authority), path, true
	}
	authority, path, found := strings.Cut(url, ":")
	if !found {
		return "", "", false
	}
	return scopeHostOf(authority), path, true
}

func scopeHostOf(authority string) string {
	if i := strings.LastIndex(authority, "@"); i >= 0 {
		authority = authority[i+1:]
	}
	if i := strings.Index(authority, ":"); i >= 0 {
		authority = authority[:i]
	}
	return authority
}

func scopeHasOrg(orgs []string, org string) bool {
	for _, candidate := range orgs {
		if strings.EqualFold(candidate, org) {
			return true
		}
	}
	return false
}

func scopeValidOrg(org string) bool {
	if org == "" || len(org) > 39 {
		return false
	}
	for _, r := range org {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-') {
			return false
		}
	}
	return true
}
