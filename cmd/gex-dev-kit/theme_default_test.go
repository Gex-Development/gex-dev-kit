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

func TestCommonDeLeituraEEscritaComecamComFundoEscuro(t *testing.T) {
	s := session{env: terminal.Env{}, opts: options{Home: t.TempDir()}}
	if c := readCommon(s); !c.Dark {
		t.Fatalf("a leitura (instalação, doctor, vitrine) deve começar escura; Dark = false")
	}
	if c := writeCommon(s); !c.Dark {
		t.Fatalf("a escrita (rollback, git hooks) deve começar escura; Dark = false")
	}
}

func TestTuiCommonPowerlineEnvironment(t *testing.T) {
	s := session{env: terminal.Env{Environ: []string{"GEX_NO_POWERLINE=0"}}}
	if !tuiCommon(s, t.TempDir()).Powerline {
		t.Fatal("powerline deve estar ligado por padrão")
	}
	s.env.Environ = []string{"GEX_NO_POWERLINE=1"}
	if tuiCommon(s, t.TempDir()).Powerline {
		t.Fatal("GEX_NO_POWERLINE=1 deve desligar glifos")
	}
}
