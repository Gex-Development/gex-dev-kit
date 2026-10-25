package app

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

type GitInfo struct {
	Branch              string
	Modified, Untracked int
}

func GitStatus(dir string) (GitInfo, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	branch, err := exec.CommandContext(ctx, "git", "-C", dir, "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		branch, err = exec.CommandContext(ctx, "git", "-C", dir, "symbolic-ref", "--short", "HEAD").Output()
		if err != nil {
			return GitInfo{}, false
		}
	}
	status, err := exec.CommandContext(ctx, "git", "-C", dir, "status", "--porcelain").Output()
	if err != nil {
		return GitInfo{}, false
	}
	info := GitInfo{Branch: strings.TrimSpace(string(branch))}
	for _, line := range strings.Split(strings.TrimSuffix(string(status), "\n"), "\n") {
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "??") {
			info.Untracked++
		} else {
			info.Modified++
		}
	}
	return info, true
}
