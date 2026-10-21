//go:build unix

package screens

import (
	"os"
	"os/signal"
	"syscall"
)

func NotifySignals() (<-chan os.Signal, func()) {
	signals := make(chan os.Signal, 4)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	return signals, func() { signal.Stop(signals) }
}
