package main

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

var readmeVersionPattern = regexp.MustCompile(`v\d+\.\d+\.\d+`)

func TestVersionCommandMatchesReadmeLocalRelease(t *testing.T) {
	code, out, errOut := invoke(t, "version")
	if code != 0 || errOut != "" {
		t.Fatalf("version: código %d, erro %q", code, errOut)
	}
	readme, err := os.ReadFile("../../README.md")
	if err != nil {
		t.Fatalf("ler README: %v", err)
	}
	var linha string
	for _, l := range strings.Split(string(readme), "\n") {
		if strings.Contains(l, "preparada localmente") {
			linha = l
			break
		}
	}
	if linha == "" {
		t.Fatal("README sem a frase de estado 'preparada localmente'")
	}
	esperada := readmeVersionPattern.FindString(linha)
	if esperada == "" {
		t.Fatalf("README sem versão na frase de estado: %q", linha)
	}
	if got := strings.TrimSpace(out); got != esperada {
		t.Fatalf("version = %q, README cita %q", got, esperada)
	}
}
