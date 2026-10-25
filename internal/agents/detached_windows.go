//go:build windows

package agents

import "os/exec"

func procDetachTTY(*exec.Cmd) {}
