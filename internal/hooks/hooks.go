package hooks

import (
	"errors"
	"regexp"
	"strings"
)

var conventional = regexp.MustCompile(`^[a-z][a-z0-9-]*(\([a-zA-Z0-9._/-]+\))?!?: \S.*$`)
var automatic = regexp.MustCompile(`^(Merge (branch|remote-tracking branch|tag|pull request) .+|Revert ".+"|(fixup|squash|amend)! \S.*)$`)

var errMessage = errors.New("commit requer mensagem Conventional Commits literal; use -m 'tipo: descrição' ou --amend --no-edit")

const shellBOM = string(rune(0xFEFF))

func CleanGitMessage(message, commentChar string) string {
	if commentChar == "" || commentChar == "auto" {
		commentChar = "#"
	}
	scissors := commentChar + " ------------------------ >8 ------------------------"
	lines := strings.Split(strings.ReplaceAll(message, "\r\n", "\n"), "\n")
	var kept []string
	for _, line := range lines {
		if line == scissors {
			break
		}
		if strings.HasPrefix(line, commentChar) {
			continue
		}
		kept = append(kept, line)
	}
	for len(kept) > 0 && strings.TrimSpace(kept[0]) == "" {
		kept = kept[1:]
	}
	return strings.Join(kept, "\n")
}

func CheckGitMessage(message string) error {
	cleaned := CleanGitMessage(strings.TrimPrefix(message, shellBOM), "#")
	header, _, _ := strings.Cut(cleaned, "\n")
	header = strings.TrimSuffix(header, "\r")
	if conventional.MatchString(header) || automatic.MatchString(header) {
		return nil
	}
	return errMessage
}
