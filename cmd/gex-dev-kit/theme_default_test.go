package main

import (
	"testing"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/terminal"
)

func TestTuiCommonComecaComFundoEscuro(t *testing.T) {
	s := session{env: terminal.Env{}, opts: options{}}
	home := t.TempDir()
	c := tuiCommon(s, home)
	if !c.Dark {
		t.Fatalf("a TUI deve começar com paleta escura até o terminal responder; Dark = false")
	}
	if c.Home != home {
		t.Fatalf("Home = %q, esperado %q", c.Home, home)
	}
}
