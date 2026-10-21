package main

import (
	"strings"
	"testing"
)

func TestNoArgsNonTTYShowsHelp(t *testing.T) {
	code, out, errors := invoke(t)
	if code != 0 || errors != "" || !strings.Contains(out, "install") {
		t.Fatalf("ajuda: exit=%d out=%q err=%q", code, out, errors)
	}
}

func TestSubcommandHelpWritesOnlyStdout(t *testing.T) {
	for _, args := range [][]string{
		{"install", "--help"}, {"doctor", "-h"}, {"rollback", "--help"},
		{"rollback-agents", "--help"}, {"skills", "--help"},
		{"hooks", "-h"}, {"hook", "commit-msg", "--help"},
	} {
		code, out, errors := invoke(t, args...)
		if code != 0 || errors != "" || !strings.Contains(out, args[0]) {
			t.Fatalf("%v: exit=%d out=%q err=%q", args, code, out, errors)
		}
	}
}

func TestVersionShortFlag(t *testing.T) {
	code, out, errors := invoke(t, "-v")
	if code != 0 || strings.TrimSpace(out) != version || errors != "" {
		t.Fatalf("-v: exit=%d out=%q err=%q", code, out, errors)
	}
}
