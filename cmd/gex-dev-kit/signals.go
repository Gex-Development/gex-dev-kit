package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"syscall"
)

var errInterrompido = errors.New("interrompido por sinal antes da gravação")

type guard struct {
	signals <-chan os.Signal
	errOut  io.Writer
	stop    chan struct{}
	quit    chan struct{}
	mu      sync.Mutex
	writing bool
	avisado bool
	sig     os.Signal
	parada  sync.Once
	saida   sync.Once
}

func newGuard(signals <-chan os.Signal) *guard {
	g := &guard{signals: signals, errOut: io.Discard, stop: make(chan struct{}), quit: make(chan struct{})}
	go g.escuta()
	return g
}

func (g *guard) escuta() {
	for {
		select {
		case sig, ok := <-g.signals:
			if !ok {
				return
			}
			g.recebe(sig)
		case <-g.stop:
			return
		}
	}
}

func (g *guard) recebe(sig os.Signal) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.sig == nil {
		g.sig = sig
	}
	if g.writing {
		if !g.avisado {
			g.avisado = true
			fmt.Fprintln(g.errOut, "Sinal recebido; concluindo a gravação…")
		}
		return
	}
	g.saida.Do(func() { close(g.quit) })
}

func (g *guard) Write(fn func() error) error {
	g.mu.Lock()
	if g.sig != nil {
		g.mu.Unlock()
		return errInterrompido
	}
	g.writing = true
	g.mu.Unlock()
	err := fn()
	g.mu.Lock()
	g.writing = false
	g.mu.Unlock()
	return err
}

func (g *guard) Code() (int, bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.sig == nil {
		return 0, false
	}
	return codigoDoSinal(g.sig), true
}

func (g *guard) Interrupted() <-chan struct{} {
	return g.quit
}

func (g *guard) Stop() {
	g.parada.Do(func() { close(g.stop) })
}

func codigoDoSinal(sig os.Signal) int {
	switch sig {
	case syscall.SIGTERM:
		return 143
	case syscall.SIGHUP:
		return 129
	}
	return 130
}
