package hooks

import (
	"encoding/json"
	"errors"
	"io"
	"regexp"
	"strings"
)

const maxInput = 64 << 10

var conventional = regexp.MustCompile(`^[a-z][a-z0-9-]*(\([a-zA-Z0-9._/-]+\))?!?: \S.*$`)
var automatic = regexp.MustCompile(`^(Merge (branch|remote-tracking branch|tag|pull request) .+|Revert ".+")$`)

var errInput = errors.New("entrada do hook inválida")
var errMessage = errors.New("commit requer mensagem Conventional Commits literal; use -m 'tipo: descrição' ou --amend --no-edit")

func CheckGitMessage(message string) error {
	header, _, _ := strings.Cut(strings.TrimPrefix(message, "\ufeff"), "\n")
	header = strings.TrimSuffix(header, "\r")
	if conventional.MatchString(header) || automatic.MatchString(header) {
		return nil
	}
	return errMessage
}

func CheckAgentCommit(input io.Reader) error {
	if input == nil {
		return errInput
	}
	data, err := io.ReadAll(io.LimitReader(input, maxInput+1))
	if err != nil || len(data) > maxInput {
		return errInput
	}
	var event struct {
		ToolName  string `json:"tool_name"`
		ToolInput struct {
			Command *string `json:"command"`
		} `json:"tool_input"`
	}
	if json.Unmarshal(data, &event) != nil || event.ToolName != "Bash" || event.ToolInput.Command == nil {
		return errInput
	}
	return checkShell(*event.ToolInput.Command, 0)
}

func checkShell(command string, depth int) error {
	if depth > 3 {
		return errMessage
	}
	commands, err := splitShell(command)
	if err != nil {
		return errInput
	}
	for _, args := range commands {
		if err := checkShellCommand(args, depth); err != nil {
			return err
		}
	}
	return nil
}

func checkShellCommand(args []shellWord, depth int) error {
	if len(args) == 0 {
		return nil
	}
	switch args[0].value {
	case "env":
		i := 1
		for i < len(args) {
			if args[i].value == "-u" || args[i].value == "-C" || args[i].value == "--unset" || args[i].value == "--chdir" {
				i += 2
			} else if strings.HasPrefix(args[i].value, "-") || strings.Contains(args[i].value, "=") {
				i++
			} else {
				break
			}
		}
		if i >= len(args) {
			return nil
		}
		return checkShellCommand(args[i:], depth+1)
	case "sh", "bash":
		for i := 1; i < len(args); i++ {
			if args[i].value == "-c" || args[i].value == "-lc" || args[i].value == "-ec" {
				if i+1 >= len(args) || args[i+1].dynamic {
					return errMessage
				}
				return checkShell(args[i+1].value, depth+1)
			}
		}
	case "git":
		i := 1
		for i < len(args) {
			switch args[i].value {
			case "-C", "-c", "--git-dir", "--work-tree", "--namespace", "--config-env":
				i += 2
			default:
				if strings.HasPrefix(args[i].value, "-") {
					i++
				} else {
					goto subcommand
				}
			}
		}
	subcommand:
		if i >= len(args) {
			return nil
		}
		if args[i].dynamic {
			return errMessage
		}
		if args[i].value == "commit" {
			return checkCommit(args[i+1:])
		}
	}
	for _, arg := range args {
		if strings.Contains(arg.value, "git commit") && (arg.dynamic || args[0].value != "echo") {
			return errMessage
		}
	}
	return nil
}

type shellWord struct {
	value   string
	dynamic bool
}

func splitShell(command string) ([][]shellWord, error) {
	var commands [][]shellWord
	var words []shellWord
	var word shellWord
	active := false
	quote := byte(0)
	flushWord := func() {
		if active {
			words = append(words, word)
			word = shellWord{}
			active = false
		}
	}
	flushCommand := func() {
		flushWord()
		if len(words) != 0 {
			commands = append(commands, words)
			words = nil
		}
	}
	for i := 0; i < len(command); i++ {
		c := command[i]
		if c == '\\' && quote != '\'' {
			if i+1 >= len(command) {
				return nil, errInput
			}
			i++
			if command[i] != '\n' {
				word.value += string(command[i])
				active = true
			}
			continue
		}
		if c == quote && quote != 0 {
			quote = 0
			continue
		}
		if quote == '\'' {
			word.value += string(c)
			continue
		}
		if c == '$' || c == '`' {
			word.dynamic = true
		}
		if c == '\'' || c == '"' {
			if quote == 0 {
				quote = c
				active = true
				continue
			}
		}
		if quote == 0 {
			switch c {
			case ' ', '\t', '\r':
				flushWord()
				continue
			case '\n', ';', '|', '&':
				flushCommand()
				continue
			}
		}
		word.value += string(c)
		active = true
	}
	if quote != 0 {
		return nil, errInput
	}
	flushCommand()
	return commands, nil
}

func checkCommit(args []shellWord) error {
	var messages []string
	amend, noEdit := false, false
	for i := 0; i < len(args); i++ {
		option := args[i].value
		switch {
		case option == "--amend":
			amend = true
		case option == "--no-edit":
			noEdit = true
		case option == "--":
			i = len(args)
			continue
		case option == "-F" || option == "--file" || strings.HasPrefix(option, "--file=") || strings.HasPrefix(option, "-F") || option == "-e" || option == "--edit" || option == "-c" || option == "-C" || strings.HasPrefix(option, "--reuse-message") || strings.HasPrefix(option, "--reedit-message") || strings.HasPrefix(option, "--fixup") || strings.HasPrefix(option, "--squash") || option == "-t" || option == "--template" || strings.HasPrefix(option, "--template="):
			return errMessage
		case strings.HasPrefix(option, "-") && !strings.HasPrefix(option, "--") && strings.ContainsAny(strings.SplitN(option[1:], "m", 2)[0], "FecCt"):
			return errMessage
		case option == "-m" || option == "--message":
			i++
			if i >= len(args) || args[i].dynamic {
				return errMessage
			}
			messages = append(messages, args[i].value)
		case strings.HasPrefix(option, "--message="):
			if args[i].dynamic {
				return errMessage
			}
			messages = append(messages, strings.TrimPrefix(option, "--message="))
		case strings.HasPrefix(option, "-m") && len(option) > 2:
			if args[i].dynamic {
				return errMessage
			}
			messages = append(messages, option[2:])
		case strings.HasPrefix(option, "-") && !strings.HasPrefix(option, "--") && strings.Contains(option[1:], "m"):
			if args[i].dynamic {
				return errMessage
			}
			pos := strings.IndexByte(option[1:], 'm') + 2
			if pos < len(option) {
				messages = append(messages, option[pos:])
			} else {
				i++
				if i >= len(args) || args[i].dynamic {
					return errMessage
				}
				messages = append(messages, args[i].value)
			}
		}
	}
	if len(messages) == 0 {
		if amend && noEdit {
			return nil
		}
		return errMessage
	}
	return CheckGitMessage(strings.Join(messages, "\n\n"))
}
