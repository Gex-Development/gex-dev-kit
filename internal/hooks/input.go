package hooks

import "io"

const MaxInput = 8 << 20

type AgentInput struct {
	Event     string
	Tool      string
	Command   string
	FilePath  string
	Cwd       string
	Source    string
	Raw       []byte
	Parsed    bool
	Truncated bool
}

func ReadAgentInput(r io.Reader) AgentInput {
	return AgentInput{}
}
