//go:build !windows

package agents

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestRunDetachedKillsDescendantsOnTimeout(t *testing.T) {
	dir := t.TempDir()
	pidFile := filepath.Join(dir, "child")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	start := time.Now()
	err := RunDetached(ctx, dir, []string{"/bin/sh", "-c", `sleep 61 & echo $! > "$0"; sleep 61`, pidFile}, os.Environ(), nil)
	if !errors.Is(ctx.Err(), context.DeadlineExceeded) || err == nil || time.Since(start) >= 3*time.Second {
		t.Fatalf("timeout did not stop group promptly: %v, %s", err, time.Since(start))
	}
	pids := procReadPIDs(t, pidFile)
	pid := pids[0]
	t.Cleanup(func() { _ = syscall.Kill(pid, syscall.SIGKILL) })
	procWaitGone(t, pid)
}

func TestRunDetachedUsesDirAndLeaderExitStatus(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "pwd")
	var output strings.Builder
	start := time.Now()
	err := RunDetached(context.Background(), dir, []string{"/bin/sh", "-c", `pwd > "$0"; sleep 61 & echo $! > "$1"; exit 0`, file, filepath.Join(dir, "child")}, os.Environ(), &output)
	if err != nil || time.Since(start) >= 7*time.Second {
		t.Fatalf("successful leader held by descendant: %v, %s", err, time.Since(start))
	}
	b, err := os.ReadFile(file)
	resolved, resolveErr := filepath.EvalSymlinks(dir)
	if err != nil || resolveErr != nil || strings.TrimSpace(string(b)) != resolved {
		t.Fatalf("wrong working directory: %q, %v, %v", b, err, resolveErr)
	}
	pidText, err := os.ReadFile(filepath.Join(dir, "child"))
	if err != nil {
		t.Fatal(err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(pidText)))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = syscall.Kill(pid, syscall.SIGKILL) })
	procWaitGone(t, pid)
}
