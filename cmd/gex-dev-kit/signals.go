package main

import "os"

type guard struct{}

func newGuard(signals <-chan os.Signal) *guard {
	return &guard{}
}

func (g *guard) Write(fn func() error) error {
	return fn()
}

func (g *guard) Code() (int, bool) {
	return 0, false
}

func (g *guard) Stop() {}
