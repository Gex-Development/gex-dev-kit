package main

import (
	"bytes"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestGuardAdiaSinalDuranteGravacao(t *testing.T) {
	canal := make(chan os.Signal, 1)
	g := newGuard(canal)
	defer g.Stop()
	var errOut bytes.Buffer
	g.errOut = &errOut
	finalizou := false
	err := g.Write(func() error {
		canal <- os.Interrupt
		time.Sleep(300 * time.Millisecond)
		finalizou = true
		return nil
	})
	if err != nil || !finalizou {
		t.Fatalf("gravação interrompida: err=%v finalizou=%v", err, finalizou)
	}
	code, ok := g.Code()
	if !ok || code != 130 {
		t.Fatalf("Code() = %d, %v; quero 130, true", code, ok)
	}
	if !strings.Contains(errOut.String(), "Sinal recebido; concluindo a gravação…") {
		t.Fatalf("aviso de sinal ausente: %q", errOut.String())
	}
}

func TestGuardSinalForaDeGravacaoSaiNaHora(t *testing.T) {
	canal := make(chan os.Signal, 1)
	g := newGuard(canal)
	defer g.Stop()
	canal <- syscall.SIGTERM
	select {
	case <-g.Interrupted():
	case <-time.After(2 * time.Second):
		t.Fatalf("sinal fora de gravação não cancelou")
	}
	code, ok := g.Code()
	if !ok || code != 143 {
		t.Fatalf("Code() = %d, %v; quero 143, true", code, ok)
	}
	ninguem := false
	if err := g.Write(func() error { ninguem = true; return nil }); err == nil || ninguem {
		t.Fatalf("gravação iniciada depois do sinal")
	}
}
