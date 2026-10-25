//go:build windows

package screens

import (
	"os"
	"os/signal"
	"syscall"
)

func NotifySignals() (<-chan os.Signal, func()) {
	signals := make(chan os.Signal, 4)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	return signals, func() { signal.Stop(signals) }
}
