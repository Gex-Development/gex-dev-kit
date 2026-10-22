package screens

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestHomeLabelCollapsesUserHomeAndResolvedForm(t *testing.T) {
	home := t.TempDir()
	resolved, err := filepath.EvalSymlinks(home)
	if err != nil {
		t.Fatal(err)
	}
	for _, input := range []string{home, resolved, filepath.Join(home, ".claude", "CLAUDE.md"), filepath.Join(resolved, ".codex")} {
		label := homeLabel(input, home)
		if label != "~" && label[:2] != "~/" {
			t.Fatalf("%q não virou ~: %q", input, label)
		}
	}
	if label := homeLabel("/outro/lugar", home); label != "/outro/lugar" {
		t.Fatalf("caminho fora do HOME foi alterado: %q", label)
	}
	if label := homeLabel(home+"x", home); label == "~x" {
		t.Fatalf("prefixo parcial virou ~: %q", label)
	}
}

func TestMaskHomeTrocaHomeDoUsuarioNoTexto(t *testing.T) {
	const home = "/home/demo"
	cases := map[string]string{
		"bloqueado em /home/demo/.local/bin/claude": "bloqueado em ~/.local/bin/claude",
		"/home/demo":                  "~",
		"x '/home/demo' y":            "x '~' y",
		"/home/demo/a e /home/demo/b": "~/a e ~/b",
		"/home/demo2/x":               "/home/demo2/x",
		"/outro/lugar":                "/outro/lugar",
	}
	for input, want := range cases {
		if got := maskHome(input, home); got != want {
			t.Fatalf("maskHome(%q) = %q; esperado %q", input, got, want)
		}
	}
	if got := maskHome("/home/demo/x", ""); got != "/home/demo/x" {
		t.Fatalf("sem HOME do usuário o texto não deveria mudar: %q", got)
	}
}

func TestExitCodeMapsSignals(t *testing.T) {
	cases := map[os.Signal]int{os.Interrupt: 130, syscall.SIGTERM: 143, syscall.SIGHUP: 129}
	for signal, want := range cases {
		code, ok := ExitCode(SignalError{Signal: signal})
		if !ok || code != want {
			t.Fatalf("%v: código %d ok=%v, esperado %d", signal, code, ok, want)
		}
	}
	if _, ok := ExitCode(errors.New("outro erro")); ok {
		t.Fatal("erro comum não pode virar código de sinal")
	}
}
