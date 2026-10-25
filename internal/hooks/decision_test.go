package hooks

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDecisionReasonOnlyWhenAskingOrBlocking(t *testing.T) {
	if reason := (Decision{}).Reason(); reason != "" {
		t.Fatalf("decisão neutra não pode ter motivo: %q", reason)
	}
	reason := Decision{Ask: true, Category: "publicação externa", Rule: "git push --force"}.Reason()
	if !strings.Contains(reason, "publicação externa") || !strings.Contains(reason, "git push --force") || !strings.Contains(reason, "autorização explícita") {
		t.Fatalf("motivo incompleto: %q", reason)
	}
}

func TestAskOutputIsClaudePreToolUseJSONWithoutHTMLEscape(t *testing.T) {
	out := AskOutput("a && b <c>")
	var parsed struct {
		Hook struct {
			Event    string `json:"hookEventName"`
			Decision string `json:"permissionDecision"`
			Reason   string `json:"permissionDecisionReason"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal(out, &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed.Hook.Event != "PreToolUse" || parsed.Hook.Decision != "ask" || parsed.Hook.Reason != "a && b <c>" {
		t.Fatalf("saída inesperada: %s", out)
	}
	backslash := string(rune(92))
	if strings.Contains(string(out), backslash+"u0026") || strings.Contains(string(out), backslash+"u003c") || !strings.HasSuffix(string(out), "\n") {
		t.Fatalf("JSON escapou HTML ou perdeu quebra final: %q", out)
	}
}
