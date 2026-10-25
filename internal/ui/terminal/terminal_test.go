package terminal

import (
	"bytes"
	"os"
	"testing"

	"github.com/charmbracelet/colorprofile"
)

func terminalPipes(t *testing.T) (*os.File, *os.File) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	t.Cleanup(func() {
		r.Close()
		w.Close()
	})
	return r, w
}

func terminalTestEnv() Env {
	return Env{
		StdinTTY:  true,
		StdoutTTY: true,
		Term:      "xterm-256color",
		Width:     80,
		Height:    24,
		Profile:   colorprofile.TrueColor,
	}
}

func TestParseMode(t *testing.T) {
	casos := map[string]Mode{"": ModeAuto, "auto": ModeAuto, "tui": ModeTUI, "cli": ModeCLI}
	for entrada, quero := range casos {
		got, err := ParseMode(entrada)
		if err != nil || got != quero {
			t.Fatalf("ParseMode(%q) = %v, %v; quero %v, nil", entrada, got, err, quero)
		}
	}
	if _, err := ParseMode("gui"); err == nil {
		t.Fatalf("ParseMode(\"gui\") não devolveu erro")
	}
}

func TestDetectWithPipesIsNotTerminal(t *testing.T) {
	in, out := terminalPipes(t)
	env := Detect(in, out, []string{"TERM=xterm-256color"})
	if env.StdinTTY || env.StdoutTTY {
		t.Fatalf("pipe marcado como terminal: stdin=%v stdout=%v", env.StdinTTY, env.StdoutTTY)
	}
	if env.Width != 0 || env.Height != 0 {
		t.Fatalf("tamanho lido fora de TTY: %dx%d", env.Width, env.Height)
	}
	if env.InFile != in || env.OutFile != out {
		t.Fatalf("arquivos de entrada e saída não foram guardados")
	}
	if got := env.ColorProfile(); got != colorprofile.NoTTY {
		t.Fatalf("perfil sem TTY = %v; quero NoTTY", got)
	}
	if mode, err := env.Resolve(ModeAuto); err != nil || mode != ModeCLI {
		t.Fatalf("Resolve(auto) sem TTY = %v, %v; quero CLI", mode, err)
	}
}

func TestDetectReadsFlagsFromEnviron(t *testing.T) {
	in, out := terminalPipes(t)
	cheio := Detect(in, out, []string{"NO_COLOR=yes", "ACCESSIBLE=1", "CI=true", "GEX_NO_ANIM=1"})
	if !cheio.NoColor || !cheio.Accessible || !cheio.CI || !cheio.NoAnim {
		t.Fatalf("flags não lidas: %+v", cheio)
	}
	vazio := Detect(in, out, []string{"NO_COLOR=", "ACCESSIBLE=", "CI=", "GEX_NO_ANIM="})
	if vazio.NoColor || vazio.Accessible || vazio.CI || vazio.NoAnim {
		t.Fatalf("variável vazia contou como ligada: %+v", vazio)
	}
}

func TestResolveAutoTable(t *testing.T) {
	casos := []struct {
		nome  string
		env   func() Env
		quero Mode
	}{
		{"terminal limpo vira TUI", terminalTestEnv, ModeTUI},
		{"ACCESSIBLE vira CLI", func() Env { e := terminalTestEnv(); e.Accessible = true; return e }, ModeCLI},
		{"CI vira CLI", func() Env { e := terminalTestEnv(); e.CI = true; return e }, ModeCLI},
		{"tamanho 0x0 vira CLI", func() Env { e := terminalTestEnv(); e.Width, e.Height = 0, 0; return e }, ModeCLI},
		{"TERM=dumb vira CLI", func() Env { e := terminalTestEnv(); e.Term = "dumb"; return e }, ModeCLI},
		{"só stdin TTY vira CLI", func() Env { e := terminalTestEnv(); e.StdoutTTY = false; return e }, ModeCLI},
	}
	for _, caso := range casos {
		got, err := caso.env().Resolve(ModeAuto)
		if err != nil || got != caso.quero {
			t.Fatalf("%s: Resolve(auto) = %v, %v; quero %v", caso.nome, got, err, caso.quero)
		}
	}
}

func TestResolveForcedCLIInTerminal(t *testing.T) {
	got, err := terminalTestEnv().Resolve(ModeCLI)
	if err != nil || got != ModeCLI {
		t.Fatalf("Resolve(cli) em TTY = %v, %v; quero CLI", got, err)
	}
}

func TestResolveTUIWithoutTerminalIsError(t *testing.T) {
	in, out := terminalPipes(t)
	if _, err := Detect(in, out, nil).Resolve(ModeTUI); err == nil {
		t.Fatalf("--ui=tui sem terminal não devolveu erro de uso")
	}
	dumb := terminalTestEnv()
	dumb.Term = "dumb"
	if _, err := dumb.Resolve(ModeTUI); err == nil {
		t.Fatalf("--ui=tui com TERM=dumb não devolveu erro de uso")
	}
	if got, err := terminalTestEnv().Resolve(ModeTUI); err != nil || got != ModeTUI {
		t.Fatalf("--ui=tui em terminal limpo = %v, %v; quero TUI", got, err)
	}
}

func TestNoColorCapsProfileAtASCII(t *testing.T) {
	env := terminalTestEnv()
	env.NoColor = true
	if got := env.ColorProfile(); got != colorprofile.ASCII {
		t.Fatalf("NO_COLOR com TrueColor = %v; quero ASCII", got)
	}
	env.NoColor = false
	if got := env.ColorProfile(); got != colorprofile.TrueColor {
		t.Fatalf("sem NO_COLOR = %v; quero TrueColor", got)
	}
}

func TestMotionOffInEachCondition(t *testing.T) {
	casos := []struct {
		nome string
		flag bool
		env  func() Env
	}{
		{"flag --no-anim", true, terminalTestEnv},
		{"GEX_NO_ANIM", false, func() Env { e := terminalTestEnv(); e.NoAnim = true; return e }},
		{"ACCESSIBLE", false, func() Env { e := terminalTestEnv(); e.Accessible = true; return e }},
		{"CI", false, func() Env { e := terminalTestEnv(); e.CI = true; return e }},
		{"sem TTY de saída", false, func() Env { e := terminalTestEnv(); e.StdoutTTY = false; return e }},
		{"TERM=dumb", false, func() Env { e := terminalTestEnv(); e.Term = "dumb"; return e }},
	}
	for _, caso := range casos {
		if caso.env().Motion(caso.flag) {
			t.Fatalf("Motion com %s = true; quero false", caso.nome)
		}
	}
}

func TestMotionOnInCleanTerminal(t *testing.T) {
	if !terminalTestEnv().Motion(false) {
		t.Fatalf("Motion em terminal limpo = false; quero true")
	}
}

func TestVtProfile(t *testing.T) {
	casos := []struct {
		base  colorprofile.Profile
		vtOK  bool
		quero colorprofile.Profile
	}{
		{colorprofile.NoTTY, false, colorprofile.NoTTY},
		{colorprofile.ASCII, false, colorprofile.ASCII},
		{colorprofile.ANSI256, false, colorprofile.ASCII},
		{colorprofile.TrueColor, false, colorprofile.ASCII},
		{colorprofile.ANSI256, true, colorprofile.ANSI256},
		{colorprofile.TrueColor, true, colorprofile.TrueColor},
	}
	for _, caso := range casos {
		if got := vtProfile(caso.base, caso.vtOK); got != caso.quero {
			t.Fatalf("vtProfile(%v, %v) = %v; quero %v", caso.base, caso.vtOK, got, caso.quero)
		}
	}
}

func TestWriterFollowsProfile(t *testing.T) {
	esc := string(rune(27))
	entrada := []byte(esc + "[31mvermelho" + esc + "[0m")

	var colorido bytes.Buffer
	if _, err := terminalTestEnv().Writer(&colorido).Write(entrada); err != nil {
		t.Fatalf("Write em TTY: %v", err)
	}
	if colorido.String() != string(entrada) {
		t.Fatalf("TTY com TrueColor alterou a saída: %q", colorido.String())
	}

	var puro bytes.Buffer
	pipe := terminalTestEnv()
	pipe.StdoutTTY = false
	if _, err := pipe.Writer(&puro).Write(entrada); err != nil {
		t.Fatalf("Write sem TTY: %v", err)
	}
	if puro.String() != "vermelho" {
		t.Fatalf("sem TTY a saída deveria ser só texto: %q", puro.String())
	}
}

func TestNoColorKeepsBold(t *testing.T) {
	esc := string(rune(27))
	entrada := []byte(esc + "[1mnegrito" + esc + "[31m vermelho" + esc + "[0m")
	env := terminalTestEnv()
	env.NoColor = true
	var saida bytes.Buffer
	if _, err := env.Writer(&saida).Write(entrada); err != nil {
		t.Fatalf("Write com NO_COLOR: %v", err)
	}
	if !bytes.Contains(saida.Bytes(), []byte(esc+"[1m")) {
		t.Fatalf("NO_COLOR removeu o negrito: %q", saida.String())
	}
	if bytes.Contains(saida.Bytes(), []byte("[31m")) {
		t.Fatalf("NO_COLOR manteve cor: %q", saida.String())
	}
}
