package hooks

import "errors"

type Word struct {
	Value   string
	Dynamic bool
}

type Command struct {
	Words []Word
	Dir   string
}

type Script struct {
	Commands []Command
	Dynamic  bool
}

func ParseShell(command string) (Script, error) {
	return Script{}, errors.New("não implementado")
}

func Effective(c Command) []Command {
	return nil
}

func HasWords(text string, words ...string) bool {
	return false
}
