package agents

import (
	"context"
	"errors"
	"io"
	"os/exec"
)

func RunDetached(ctx context.Context, dir string, argv, env []string, out io.Writer) error {
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Dir = dir
	cmd.Env = env
	cmd.Stdout, cmd.Stderr = out, out
	cmd.WaitDelay = waitDelay
	group, err := procNewGroup(cmd)
	if err != nil {
		return err
	}
	defer group.close()
	procDetachTTY(cmd)
	cmd.Cancel = func() error { return group.kill(cmd) }
	err = runGroup(cmd, group)
	if errors.Is(err, exec.ErrWaitDelay) {
		return nil
	}
	return err
}
