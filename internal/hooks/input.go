package hooks

import (
	"encoding/json"
	"io"
)

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

type agentPayload struct {
	Event     string `json:"hook_event_name"`
	Tool      string `json:"tool_name"`
	ToolInput struct {
		Command  string `json:"command"`
		FilePath string `json:"file_path"`
	} `json:"tool_input"`
	Cwd    string `json:"cwd"`
	Source string `json:"source"`
}

func ReadAgentInput(r io.Reader) AgentInput {
	if r == nil {
		return AgentInput{}
	}
	data, err := io.ReadAll(io.LimitReader(r, MaxInput+1))
	if err != nil {
		return AgentInput{Raw: data}
	}
	if len(data) > MaxInput {
		return AgentInput{Raw: data[:MaxInput], Truncated: true}
	}
	in := AgentInput{Raw: data}
	var payload agentPayload
	if json.Unmarshal(data, &payload) != nil {
		return in
	}
	in.Parsed = true
	in.Event = payload.Event
	in.Tool = payload.Tool
	in.Command = payload.ToolInput.Command
	in.FilePath = payload.ToolInput.FilePath
	in.Cwd = payload.Cwd
	in.Source = payload.Source
	return in
}
