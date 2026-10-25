//go:build !windows

package agents

import (
	"os/exec"
	"syscall"
)

func procDetachTTY(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}
