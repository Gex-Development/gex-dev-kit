package main

import (
	"bytes"
	"errors"
	"testing"
)

func TestTelaFalhouSemErroSaiComZero(t *testing.T) {
	var out bytes.Buffer
	if code := telaFalhou(&out, nil); code != 0 || out.Len() != 0 {
		t.Fatalf("tela encerrada sem erro deveria sair com 0 e sem saída, veio %d %q", code, out.String())
	}
}

func TestTelaFalhouComErroSaiComUmEExplica(t *testing.T) {
	var out bytes.Buffer
	if code := telaFalhou(&out, errors.New("boom")); code != 1 || out.String() != "Falha na tela: boom\n" {
		t.Fatalf("erro de tela = %d %q", code, out.String())
	}
}
