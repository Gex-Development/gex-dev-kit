package hooks

import (
	"strings"
	"testing"
)

func TestReadAgentInputParsesBashPayload(t *testing.T) {
	payload := `{"hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{"command":"ls -la"},"cwd":"/tmp/pasta","source":"startup"}`
	in := ReadAgentInput(strings.NewReader(payload))
	if !in.Parsed {
		t.Fatal("JSON válido marcado como ilegível")
	}
	if in.Event != "PreToolUse" || in.Tool != "Bash" || in.Command != "ls -la" || in.Cwd != "/tmp/pasta" || in.Source != "startup" {
		t.Fatalf("campos lidos errados: %+v", in)
	}
}

func TestReadAgentInputReadsFilePath(t *testing.T) {
	in := ReadAgentInput(strings.NewReader(`{"tool_name":"Read","tool_input":{"file_path":"/tmp/.env"}}`))
	if !in.Parsed || in.FilePath != "/tmp/.env" || in.Tool != "Read" {
		t.Fatalf("caminho do arquivo não lido: %+v", in)
	}
}

func TestReadAgentInputMarksMalformedJSON(t *testing.T) {
	in := ReadAgentInput(strings.NewReader(`{"tool_name":`))
	if in.Parsed {
		t.Fatal("JSON ilegível marcado como lido")
	}
	if string(in.Raw) != `{"tool_name":` {
		t.Fatalf("texto bruto não preservado: %q", in.Raw)
	}
}

func TestReadAgentInputTruncatesAboveLimit(t *testing.T) {
	big := strings.Repeat("x", MaxInput+1)
	in := ReadAgentInput(strings.NewReader(big))
	if !in.Truncated {
		t.Fatal("entrada acima de MaxInput não marcada como truncada")
	}
	if int64(len(in.Raw)) > MaxInput {
		t.Fatalf("texto bruto maior que o teto: %d bytes", len(in.Raw))
	}
}

func TestReadAgentInputAcceptsExactLimit(t *testing.T) {
	prefix := `{"tool_name":"Bash","tool_input":{"command":"git status","padding":"`
	suffix := `"}}`
	payload := prefix + strings.Repeat("x", MaxInput-len(prefix)-len(suffix)) + suffix
	in := ReadAgentInput(strings.NewReader(payload))
	if in.Truncated || !in.Parsed {
		t.Fatalf("entrada no limite recusada: truncada=%v lida=%v", in.Truncated, in.Parsed)
	}
}

func TestReadAgentInputNilReader(t *testing.T) {
	in := ReadAgentInput(nil)
	if in.Parsed || in.Truncated {
		t.Fatalf("leitor ausente não tratado como entrada vazia: %+v", in)
	}
}
