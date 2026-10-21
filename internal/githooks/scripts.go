package githooks

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/Gex-Development/gex-dev-kit/internal/catalog"
)

const (
	maxDocumentedCommands = 8
	maxCommandRunes       = 300
	hookFileMode          = fs.FileMode(0755)
)

var hookNames = []string{"commit-msg", "pre-commit", "pre-push"}

func Validate(r Request) error {
	if len(r.Hooks) == 0 {
		return fmt.Errorf("nenhum git hook selecionado")
	}
	selected := map[string]bool{}
	for _, name := range r.Hooks {
		if !knownHook(name) {
			return fmt.Errorf("git hook desconhecido: %s", name)
		}
		if selected[name] {
			return fmt.Errorf("git hook repetido: %s", name)
		}
		selected[name] = true
	}
	if len(r.PreCommit) > 0 && !selected["pre-commit"] {
		return fmt.Errorf("comandos de pre-commit sem o hook pre-commit selecionado")
	}
	if len(r.PrePush) > 0 && !selected["pre-push"] {
		return fmt.Errorf("comandos de pre-push sem o hook pre-push selecionado")
	}
	if err := checkCommands("pre-commit", r.PreCommit); err != nil {
		return err
	}
	if err := checkCommands("pre-push", r.PrePush); err != nil {
		return err
	}
	if selected["commit-msg"] || selected["pre-commit"] {
		return checkBinary(r.Binary)
	}
	return nil
}

func Script(hook string, r Request) (string, error) {
	switch hook {
	case "commit-msg":
		if err := checkBinary(r.Binary); err != nil {
			return "", err
		}
		return commitMsgScript(r.Binary), nil
	case "pre-commit":
		if err := checkBinary(r.Binary); err != nil {
			return "", err
		}
		if err := checkCommands(hook, r.PreCommit); err != nil {
			return "", err
		}
		return preCommitScript(r.Binary, r.PreCommit), nil
	case "pre-push":
		if err := checkCommands(hook, r.PrePush); err != nil {
			return "", err
		}
		return prePushScript(r.PrePush), nil
	}
	return "", fmt.Errorf("git hook desconhecido: %s", hook)
}

func Files(r Request) ([]catalog.File, error) {
	if err := Validate(r); err != nil {
		return nil, err
	}
	files := make([]catalog.File, 0, len(r.Hooks))
	for _, name := range r.Hooks {
		script, err := Script(name, r)
		if err != nil {
			return nil, err
		}
		files = append(files, catalog.File{Path: "hooks/" + name, Data: []byte(script), Mode: hookFileMode})
	}
	return files, nil
}

func knownHook(name string) bool {
	return slices.Contains(hookNames, name)
}

func checkBinary(binary string) error {
	if binary == "" {
		return fmt.Errorf("executável estável do HOME ausente")
	}
	if !filepath.IsAbs(binary) || strings.ContainsAny(binary, "\r\n") {
		return fmt.Errorf("executável estável precisa de caminho absoluto sem quebra de linha")
	}
	return nil
}

func checkCommands(hook string, commands []string) error {
	if len(commands) > maxDocumentedCommands {
		return fmt.Errorf("%s aceita no máximo %d comandos documentados", hook, maxDocumentedCommands)
	}
	for _, command := range commands {
		if err := checkCommand(hook, command); err != nil {
			return err
		}
	}
	return nil
}

func checkCommand(hook, command string) error {
	switch {
	case strings.TrimSpace(command) == "":
		return fmt.Errorf("%s tem comando vazio", hook)
	case strings.ContainsAny(command, "\r\n"):
		return fmt.Errorf("%s tem comando com quebra de linha", hook)
	case strings.TrimSpace(command) != command:
		return fmt.Errorf("%s tem comando com espaço nas bordas: %q", hook, command)
	case utf8.RuneCountInString(command) > maxCommandRunes:
		return fmt.Errorf("%s tem comando com mais de %d caracteres", hook, maxCommandRunes)
	}
	return nil
}

func commitMsgScript(binary string) string {
	return joinLines(
		"#!/bin/sh",
		"bin="+shellQuote(binary),
		`if [ -x "$bin" ]; then`,
		`  exec "$bin" hook commit-msg "$1"`,
		"fi",
		"echo "+shellQuote(missingBinaryWarning("commit-msg"))+" >&2",
		"exit 0",
	)
}

func preCommitScript(binary string, commands []string) string {
	lines := []string{
		"#!/bin/sh",
		"set -e",
		"bin=" + shellQuote(binary),
		`if [ ! -x "$bin" ]; then`,
		"  echo " + shellQuote(missingBinaryWarning("pre-commit")) + " >&2",
		"  exit 0",
		"fi",
		`"$bin" hook pre-commit`,
	}
	for _, command := range commands {
		lines = append(lines, commandGuard(command)...)
	}
	return joinLines(lines...)
}

func commandGuard(command string) []string {
	name := strings.Fields(command)[0]
	return []string{
		"if command -v " + shellQuote(name) + " >/dev/null 2>&1; then",
		"  " + command,
		"else",
		"  echo " + shellQuote("gex-dev-kit: "+name+" não encontrado; comando pulado: "+command) + " >&2",
		"fi",
	}
}

func prePushScript(commands []string) string {
	lines := []string{"#!/bin/sh", "set -e", "cat >/dev/null"}
	for _, command := range commands {
		lines = append(lines, "printf '%s\\n' "+shellQuote(command)+" >&2", command)
	}
	return joinLines(lines...)
}

func missingBinaryWarning(hook string) string {
	return "gex-dev-kit: executável do HOME ausente; hook " + hook + " ignorado"
}

func shellQuote(text string) string {
	return "'" + strings.ReplaceAll(text, "'", `'\''`) + "'"
}

func joinLines(lines ...string) string {
	return strings.Join(lines, "\n") + "\n"
}
