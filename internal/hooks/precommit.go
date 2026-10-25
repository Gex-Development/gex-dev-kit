package hooks

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

const precommitMaxLine = 16 << 10

type Finding struct {
	Path string
	Line int
	Rule string
}

func (f Finding) Message(allowFile string) string {
	return fmt.Sprintf("pre-commit: possível credencial em %s:%d (regra %s). Remova o valor e use variável de ambiente. Falso positivo confirmado: acrescente `%s<TAB>%s` em `%s`", f.Path, f.Line, f.Rule, f.Path, f.Rule, allowFile)
}

type precommitRule struct {
	name string
	re   *regexp.Regexp
}

var precommitRules = []precommitRule{
	{"aws-access-key", regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`)},
	{"github-token", regexp.MustCompile(`\bghp_[A-Za-z0-9]{36}\b|\bgithub_pat_[A-Za-z0-9_]{22,}`)},
	{"private-key", regexp.MustCompile(`-----BEGIN ([A-Z0-9]+ )*PRIVATE KEY-----`)},
	{"slack-token", regexp.MustCompile(`\bxox[baprs]-[0-9A-Za-z-]{10,}`)},
	{"openai-key", regexp.MustCompile(`\bsk-[A-Za-z0-9_-]{32,}`)},
	{"google-api-key", regexp.MustCompile(`\bAIza[0-9A-Za-z_-]{35}\b`)},
	{"conflict-marker", regexp.MustCompile(`^<{7}`)},
}

func ScanStaged(diff io.Reader, limit int64) (findings []Finding, truncated bool, err error) {
	if diff == nil {
		return nil, false, errors.New("diff ausente")
	}
	if limit < 0 {
		limit = 0
	}
	data, err := io.ReadAll(io.LimitReader(diff, limit+1))
	if err != nil {
		return nil, false, err
	}
	if int64(len(data)) > limit {
		data, truncated = data[:limit], true
	}
	return precommitScan(string(data)), truncated, nil
}

func precommitScan(diff string) []Finding {
	var findings []Finding
	path, line, inHunk := "", 0, false
	for _, text := range strings.Split(diff, "\n") {
		switch {
		case strings.HasPrefix(text, "diff --git "):
			path, inHunk = "", false
		case strings.HasPrefix(text, "@@"):
			inHunk = true
			line = precommitHunkStart(text)
		case !inHunk && strings.HasPrefix(text, "+++ "):
			path = strings.TrimPrefix(text[4:], "b/")
		case inHunk && strings.HasPrefix(text, "+"):
			findings = precommitAppend(findings, path, line, text[1:])
			line++
		}
	}
	return findings
}

func precommitHunkStart(header string) int {
	for _, field := range strings.Fields(header) {
		if strings.HasPrefix(field, "+") {
			start, _, _ := strings.Cut(field[1:], ",")
			if n, err := strconv.Atoi(start); err == nil {
				return n
			}
		}
	}
	return 0
}

func precommitAppend(findings []Finding, path string, line int, content string) []Finding {
	if len(content) > precommitMaxLine {
		return findings
	}
	for _, rule := range precommitRules {
		if rule.re.MatchString(content) {
			return append(findings, Finding{Path: path, Line: line, Rule: rule.name})
		}
	}
	return findings
}
