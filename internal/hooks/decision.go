package hooks

import (
	"bytes"
	"encoding/json"
)

type Decision struct {
	Ask      bool
	Block    bool
	Category string
	Rule     string
}

func (d Decision) Reason() string {
	if !d.Ask && !d.Block {
		return ""
	}
	return "GEX · " + d.Category + " (" + d.Rule + "). Confirme só com autorização explícita do responsável."
}

func AskOutput(reason string) []byte {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	encoder.Encode(map[string]any{
		"hookSpecificOutput": map[string]string{
			"hookEventName":            "PreToolUse",
			"permissionDecision":       "ask",
			"permissionDecisionReason": reason,
		},
	})
	return buffer.Bytes()
}
