package hooks

import (
	"errors"
	"io"
	"path/filepath"
	"strings"
	"unicode"
)

type CommitCheck int

const (
	CommitNone CommitCheck = iota
	CommitOK
	CommitBad
)

var errCommitUnparsed = errors.New("não foi possível analisar o commit; use `git commit -m 'tipo: descrição'` literal")

var errCommitMissingInput = errors.New("entrada do hook ausente")

var commitInterpreters = map[string]bool{"eval": true, "python": true, "python3": true, "node": true, "perl": true, "ruby": true, "php": true, "xargs": true, "osascript": true}

type wrapperSpec struct {
	valued   map[string]bool
	duration bool
}

var wrapperSpecs = map[string]wrapperSpec{
	"env":     {valued: wrapperFlags("-u", "-C", "-S", "--chdir", "--unset")},
	"sudo":    {valued: wrapperFlags("-u", "-g", "-C", "-D", "-p", "-h", "-U", "-r", "-t")},
	"command": {},
	"exec":    {valued: wrapperFlags("-a")},
	"nohup":   {},
	"time":    {valued: wrapperFlags("-f", "-o")},
	"nice":    {valued: wrapperFlags("-n")},
	"timeout": {valued: wrapperFlags("-k", "-s"), duration: true},
	"stdbuf":  {valued: wrapperFlags("-i", "-o", "-e")},
	"xargs":   {valued: wrapperFlags("-a", "-d", "-E", "-I", "-L", "-n", "-P", "-s")},
}

var commitShellNames = map[string]bool{"sh": true, "bash": true, "zsh": true, "dash": true}

func ClassifyCommit(in AgentInput) (CommitCheck, []string) {
	if !in.Parsed {
		if HasWords(string(in.Raw), "git", "commit") {
			return CommitBad, commitDirs(in.Cwd, "")
		}
		return CommitNone, nil
	}
	if in.Tool != "Bash" {
		return CommitNone, nil
	}
	script, err := ParseShell(in.Command)
	if err != nil {
		if HasWords(in.Command, "git", "commit") {
			return CommitBad, commitDirs(in.Cwd, "")
		}
		return CommitNone, nil
	}
	result := CommitNone
	var dirs []string
	for _, c := range script.Commands {
		for _, effective := range Effective(c) {
			switch commitClassifyOne(effective) {
			case CommitBad:
				return CommitBad, commitDirs(in.Cwd, effective.Dir)
			case CommitOK:
				result = CommitOK
				dirs = append(dirs, commitDirs(in.Cwd, effective.Dir)...)
			}
		}
	}
	return result, dirs
}

func CheckAgentCommit(input io.Reader) error {
	if input == nil {
		return errCommitMissingInput
	}
	in := ReadAgentInput(input)
	status, _ := ClassifyCommit(in)
	if status != CommitBad {
		return nil
	}
	if in.Parsed {
		return errMessage
	}
	return errCommitUnparsed
}

func commitDirs(cwd, dir string) []string {
	base := cwd
	if dir != "" {
		base = filepath.Join(cwd, dir)
		if filepath.IsAbs(dir) {
			base = dir
		}
	}
	if base == "" {
		return nil
	}
	return []string{base}
}

func commitClassifyOne(c Command) CommitCheck {
	if len(c.Words) == 0 {
		return CommitNone
	}
	if c.Words[0].Dynamic {
		if HasWords(commitJoin(c.Words), "git", "commit") {
			return CommitBad
		}
		return CommitNone
	}
	if c.Words[0].Value == "git" {
		return commitGit(c)
	}
	if commitOpaque(c) && HasWords(commitJoin(c.Words), "git", "commit") {
		return CommitBad
	}
	return CommitNone
}

func commitGit(c Command) CommitCheck {
	args := c.Words[1:]
	i := commitGitSubcommand(args)
	if i < 0 {
		return CommitNone
	}
	if args[i].Dynamic {
		if HasWords(commitJoin(c.Words), "git", "commit") {
			return CommitBad
		}
		return CommitNone
	}
	if args[i].Value != "commit" {
		return CommitNone
	}
	if commitOptionsOK(args[i+1:]) != nil {
		return CommitBad
	}
	return CommitOK
}

func commitGitSubcommand(args []Word) int {
	for i := 0; i < len(args); i++ {
		v := args[i].Value
		switch {
		case args[i].Dynamic:
			return i
		case v == "-C" || v == "-c" || v == "--git-dir" || v == "--work-tree" || v == "--namespace" || v == "--config-env":
			i++
		case strings.HasPrefix(v, "-"):
		default:
			return i
		}
	}
	return -1
}

func commitOptionsOK(args []Word) error {
	var messages []string
	amend, noEdit, fixup := false, false, false
	for i := 0; i < len(args); i++ {
		w := args[i]
		option := w.Value
		switch {
		case option == "--":
			i = len(args)
		case option == "--amend":
			amend = true
		case option == "--no-edit":
			noEdit = true
		case option == "--fixup" || strings.HasPrefix(option, "--fixup="):
			ref := strings.TrimPrefix(option, "--fixup=")
			if option == "--fixup" {
				i++
				if i >= len(args) {
					return errMessage
				}
				ref = args[i].Value
				if args[i].Dynamic {
					return errMessage
				}
			} else if w.Dynamic {
				return errMessage
			}
			if ref == "" || strings.HasPrefix(ref, "amend:") || strings.HasPrefix(ref, "reword:") {
				return errMessage
			}
			fixup = true
		case commitForbiddenLong(option):
			return errMessage
		case option == "-m" || option == "--message":
			i++
			if i >= len(args) || args[i].Dynamic {
				return errMessage
			}
			messages = append(messages, args[i].Value)
		case strings.HasPrefix(option, "--message="):
			if w.Dynamic {
				return errMessage
			}
			messages = append(messages, strings.TrimPrefix(option, "--message="))
		case strings.HasPrefix(option, "-") && !strings.HasPrefix(option, "--"):
			letters := option[1:]
			idx := strings.IndexByte(letters, 'm')
			if idx < 0 {
				if strings.ContainsAny(letters, "FecCt") {
					return errMessage
				}
				continue
			}
			if strings.ContainsAny(letters[:idx], "FecCt") || w.Dynamic {
				return errMessage
			}
			if rest := letters[idx+1:]; rest != "" {
				messages = append(messages, rest)
				continue
			}
			i++
			if i >= len(args) || args[i].Dynamic {
				return errMessage
			}
			messages = append(messages, args[i].Value)
		}
	}
	if fixup {
		return nil
	}
	if len(messages) == 0 {
		if amend && noEdit {
			return nil
		}
		return errMessage
	}
	return CheckGitMessage(strings.Join(messages, "\n\n"))
}

func commitForbiddenLong(option string) bool {
	for _, prefix := range []string{"--file", "--edit", "--reuse-message", "--reedit-message", "--squash", "--template"} {
		if strings.HasPrefix(option, prefix) {
			return true
		}
	}
	return false
}

func commitJoin(words []Word) string {
	values := make([]string, 0, len(words))
	for _, w := range words {
		values = append(values, w.Value)
	}
	return strings.Join(values, " ")
}

func commitOpaque(c Command) bool {
	name := c.Words[0].Value
	if commitInterpreters[name] {
		return true
	}
	if !commitShellNames[name] {
		return false
	}
	return shellHasCFlag(c.Words[1:])
}

func Effective(c Command) []Command {
	return commitEffective(c, 0)
}

func commitEffective(c Command, depth int) []Command {
	if depth >= 3 || len(c.Words) == 0 || c.Words[0].Dynamic {
		return []Command{c}
	}
	name := c.Words[0].Value
	if commitShellNames[name] {
		return commitShellC(c, depth)
	}
	if name == "eval" {
		return commitEval(c, depth)
	}
	if _, ok := wrapperSpecs[name]; !ok {
		return []Command{c}
	}
	rest := wrapperRest(name, c.Words[1:])
	if len(rest) == 0 {
		return []Command{c}
	}
	return commitEffective(Command{Words: rest, Dir: c.Dir}, depth+1)
}

func wrapperFlags(flags ...string) map[string]bool {
	set := make(map[string]bool, len(flags))
	for _, flag := range flags {
		set[flag] = true
	}
	return set
}

func wrapperRest(name string, args []Word) []Word {
	spec := wrapperSpecs[name]
	duration := spec.duration
	for i := 0; i < len(args); i++ {
		v := args[i].Value
		switch {
		case spec.valued[v]:
			i++
		case strings.HasPrefix(v, "-") && len(v) > 1:
		case name == "env" && strings.Contains(v, "="):
		case duration:
			duration = false
		default:
			return args[i:]
		}
	}
	return nil
}

func commitEval(c Command, depth int) []Command {
	parts := make([]string, 0, len(c.Words)-1)
	for _, w := range c.Words[1:] {
		if w.Dynamic {
			return []Command{c}
		}
		parts = append(parts, w.Value)
	}
	if len(parts) == 0 {
		return []Command{c}
	}
	return commitSubScript(strings.Join(parts, " "), c, depth)
}

func commitSubScript(src string, c Command, depth int) []Command {
	script, err := shellParse(src, depth+1)
	if err != nil {
		return []Command{c}
	}
	var out []Command
	for _, sub := range script.Commands {
		out = append(out, commitEffective(sub, depth+1)...)
	}
	return out
}

func commitShellC(c Command, depth int) []Command {
	args := c.Words[1:]
	for i := 0; i < len(args); i++ {
		v := args[i].Value
		if strings.HasPrefix(v, "-") && !strings.HasPrefix(v, "--") && strings.HasSuffix(v, "c") {
			if i+1 >= len(args) || args[i+1].Dynamic {
				return []Command{c}
			}
			return commitSubScript(args[i+1].Value, c, depth)
		}
	}
	return []Command{c}
}

func HasWords(text string, words ...string) bool {
	tokens := map[string]bool{}
	for _, token := range strings.FieldsFunc(text, func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_')
	}) {
		tokens[token] = true
	}
	for _, word := range words {
		if !tokens[word] {
			return false
		}
	}
	return true
}
