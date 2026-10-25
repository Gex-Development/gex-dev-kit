package hooks

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type Allow map[string]bool

func AllowPath(commonDir string) string {
	return filepath.Join(commonDir, "gex-dev-kit", "pre-commit-allow")
}

func LoadAllow(path string) (Allow, error) {
	allow := Allow{}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return allow, nil
	}
	if err != nil {
		return nil, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		file, rule, ok := strings.Cut(strings.TrimSuffix(line, "\r"), "\t")
		file, rule = strings.TrimSpace(file), strings.TrimSpace(rule)
		if ok && file != "" && rule != "" && !strings.Contains(rule, "\t") {
			allow[allowKey(file, rule)] = true
		}
	}
	return allow, nil
}

func allowKey(file, rule string) string {
	return file + "\t" + rule
}

func (a Allow) Filter(findings []Finding) []Finding {
	kept := make([]Finding, 0, len(findings))
	for _, finding := range findings {
		if !a[allowKey(finding.Path, finding.Rule)] {
			kept = append(kept, finding)
		}
	}
	return kept
}
