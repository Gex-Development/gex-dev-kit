package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHookBlocksInvalidCommitWithoutLeakingInput(t *testing.T) {
	message := "mensagem confidencial fora do padrão"
	event := `{"tool_name":"Bash","tool_input":{"command":"git commit -m '` + message + `'"}}`
	var out, errors bytes.Buffer
	code := run([]string{"hook", "commit-msg"}, strings.NewReader(event), &out, &errors)
	if code != 2 || strings.Contains(errors.String(), message) || out.Len() != 0 {
		t.Fatalf("hook: exit=%d out=%q err=%q", code, out.String(), errors.String())
	}
}

func TestHookReadsGitMessageFromPath(t *testing.T) {
	file := filepath.Join(t.TempDir(), "COMMIT_EDITMSG")
	if err := os.WriteFile(file, []byte("feat(cli): liberar instalação\n"), 0600); err != nil {
		t.Fatal(err)
	}
	code, out, errors := invoke(t, "hook", "commit-msg", file)
	if code != 0 || out != "" || errors != "" {
		t.Fatalf("mensagem Git: exit=%d out=%q err=%q", code, out, errors)
	}
}
