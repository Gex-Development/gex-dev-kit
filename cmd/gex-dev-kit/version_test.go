package main

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

var changelogHeading = regexp.MustCompile(`(?m)^## \[(\d+\.\d+\.\d+)\]`)

func TestVersionCommandMatchesNewestChangelogEntry(t *testing.T) {
	code, out, errOut := invoke(t, "version")
	if code != 0 || errOut != "" {
		t.Fatalf("version: código %d, erro %q", code, errOut)
	}
	changelog, err := os.ReadFile("../../CHANGELOG.md")
	if err != nil {
		t.Fatalf("ler CHANGELOG: %v", err)
	}
	match := changelogHeading.FindSubmatch(changelog)
	if match == nil {
		t.Fatal("CHANGELOG sem entrada de versão no formato '## [X.Y.Z]'")
	}
	if got, want := strings.TrimSpace(out), "v"+string(match[1]); got != want {
		t.Fatalf("version = %q, a entrada mais nova do CHANGELOG é %q", got, want)
	}
}
