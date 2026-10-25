package githooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

func Documented(root, command string) (Source, bool) {
	if root == "" || command == "" {
		return Source{}, false
	}
	if src, ok := documentedInText(root, command); ok {
		return src, true
	}
	if target, ok := makeTarget(command); ok {
		if line, found := makefileTargetLine(root, target); found {
			return Source{Command: command, File: "Makefile", Line: line}, true
		}
	}
	if name, ok := packageScriptName(command); ok {
		if line, found := packageScriptLine(root, name); found {
			return Source{Command: command, File: "package.json", Line: line}, true
		}
	}
	return Source{}, false
}

func documentedInText(root, command string) (Source, bool) {
	for _, name := range docFiles(root) {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			continue
		}
		for i, line := range strings.Split(string(data), "\n") {
			if containsCommand(line, command) {
				return Source{Command: command, File: name, Line: i + 1}, true
			}
		}
	}
	return Source{}, false
}

func docFiles(root string) []string {
	names := []string{"AGENTS.md"}
	entries, err := os.ReadDir(root)
	if err != nil {
		return names
	}
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, "README") || strings.HasPrefix(name, "CONTRIBUTING") {
			names = append(names, name)
		}
	}
	return names
}

func containsCommand(line, command string) bool {
	for offset := 0; offset < len(line); {
		index := strings.Index(line[offset:], command)
		if index < 0 {
			return false
		}
		start := offset + index
		end := start + len(command)
		if boundaryBefore(line, start) && boundaryAfter(line, end) {
			return true
		}
		offset = start + 1
	}
	return false
}

func boundaryBefore(line string, start int) bool {
	if start == 0 {
		return true
	}
	r, _ := utf8.DecodeLastRuneInString(line[:start])
	return !isCommandRune(r)
}

func boundaryAfter(line string, end int) bool {
	if end == len(line) {
		return true
	}
	r, _ := utf8.DecodeRuneInString(line[end:])
	return !isCommandRune(r)
}

func isCommandRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || strings.ContainsRune("_./:-", r)
}

func makeTarget(command string) (string, bool) {
	target, ok := strings.CutPrefix(command, "make ")
	if !ok || target == "" || strings.ContainsAny(target, " \t") {
		return "", false
	}
	return target, true
}

func makefileTargetLine(root, target string) (int, bool) {
	data, err := os.ReadFile(filepath.Join(root, "Makefile"))
	if err != nil {
		return 0, false
	}
	for i, line := range strings.Split(string(data), "\n") {
		rest, ok := strings.CutPrefix(line, target)
		if ok && strings.HasPrefix(rest, ":") && !strings.HasPrefix(rest, ":=") {
			return i + 1, true
		}
	}
	return 0, false
}

func packageScriptName(command string) (string, bool) {
	fields := strings.Split(command, " ")
	switch {
	case len(fields) == 3 && slices.Contains([]string{"npm", "pnpm", "yarn"}, fields[0]) && fields[1] == "run":
		return fields[2], fields[2] != ""
	case len(fields) == 2 && fields[0] == "pnpm" && fields[1] != "run":
		return fields[1], fields[1] != ""
	}
	return "", false
}

func packageScriptLine(root, name string) (int, bool) {
	data, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil {
		return 0, false
	}
	var manifest struct {
		Scripts map[string]string `json:"scripts"`
	}
	if json.Unmarshal(data, &manifest) != nil {
		return 0, false
	}
	if _, ok := manifest.Scripts[name]; !ok {
		return 0, false
	}
	for i, line := range strings.Split(string(data), "\n") {
		rest, ok := strings.CutPrefix(strings.TrimSpace(line), `"`+name+`"`)
		if ok && strings.HasPrefix(strings.TrimSpace(rest), ":") {
			return i + 1, true
		}
	}
	return 0, true
}
