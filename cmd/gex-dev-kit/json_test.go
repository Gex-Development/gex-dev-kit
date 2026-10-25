package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestWriteJSONKeepsAmpersandAndTrailingNewline(t *testing.T) {
	var out bytes.Buffer
	if err := writeJSON(&out, map[string]any{"action": "a && b", "items": []string{}}); err != nil {
		t.Fatalf("writeJSON: %v", err)
	}
	if !strings.Contains(out.String(), `"a && b"`) || !strings.HasSuffix(out.String(), "}\n") || !strings.Contains(out.String(), `"items": []`) {
		t.Fatalf("saída JSON inesperada: %q", out.String())
	}
}
