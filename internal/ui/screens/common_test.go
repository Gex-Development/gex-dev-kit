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
