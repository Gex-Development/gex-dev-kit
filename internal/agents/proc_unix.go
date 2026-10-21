//go:build !windows

package agents

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
)

type procGroup struct{}

func procNewGroup(cmd *exec.Cmd) (*procGroup, error) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return &procGroup{}, nil
}

func (g *procGroup) adopt(*exec.Cmd) error {
	return nil
}

func (g *procGroup) kill(cmd *exec.Cmd) error {
	err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	if errors.Is(err, syscall.ESRCH) {
		return os.ErrProcessDone
	}
	return err
}

func (g *procGroup) close() {}
