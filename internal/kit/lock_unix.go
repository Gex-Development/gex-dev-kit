//go:build !windows

package kit

import (
	"errors"
	"syscall"
)

func lockProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	return err == nil || errors.Is(err, syscall.EPERM)
}
