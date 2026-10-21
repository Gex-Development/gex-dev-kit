package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestAccessibleHomeOffersActionsAndCanExit(t *testing.T) {
	for _, choice := range []string{"q", "0", "cancel"} {
		var out bytes.Buffer
		selected, err := showAccessibleHome(strings.NewReader(choice+"\n"), &out)
		if err != nil || selected != "exit" || !strings.Contains(out.String(), "Doctor") || !strings.Contains(out.String(), "Instalar Core") || strings.Contains(out.String(), "\x1b") {
			t.Fatalf("escolha %q: ação=%q err=%v menu=%q", choice, selected, err, out.String())
		}
	}
	var out bytes.Buffer
	selected, err := showAccessibleHome(strings.NewReader("2\n"), &out)
	if err != nil || selected != "doctor" {
		t.Fatalf("ação Doctor: ação=%q err=%v menu=%q", selected, err, out.String())
	}
}
