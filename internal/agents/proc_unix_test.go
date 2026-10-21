//go:build !windows

package agents

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestDeadlineKillsWholeProcessGroup(t *testing.T) {
	const deadline = time.Second
	home := t.TempDir()
	pidFile := filepath.Join(home, "pids")
	ctx, cancel := context.WithTimeout(context.Background(), deadline)
	defer cancel()
	start := time.Now()
	_, err := runCommand(ctx, home, "/bin/sh", nil, "-c", `echo $$ > "$0"; sleep 30 & echo $! >> "$0"; sleep 30`, pidFile)
	elapsed := time.Since(start)
	if !errors.Is(err, errCommandTimeout) {
		t.Fatalf("prazo estourado devolveu %v; esperado tempo esgotado", err)
	}
	if elapsed >= 2*deadline {
		t.Fatalf("runCommand demorou %s com prazo de %s", elapsed, deadline)
	}
	pids := procReadPIDs(t, pidFile)
	if len(pids) != 2 {
		t.Fatalf("pids esperados: líder e neto; obtidos %v", pids)
	}
	procWaitGone(t, -pids[0])
	procWaitGone(t, pids[1])
}

func TestInstallerRunsInOwnProcessGroup(t *testing.T) {
	if _, err := exec.LookPath("ps"); err != nil {
		t.Skip("ps indisponível")
	}
	out, err := runCommand(context.Background(), t.TempDir(), "/bin/sh", nil, "-c", `echo $$; ps -o pgid= -p $$`)
	if err != nil {
		t.Fatal(err)
	}
	fields := strings.Fields(out)
	if len(fields) != 2 {
		t.Fatalf("saída inesperada: %q", out)
	}
	if fields[0] != fields[1] {
		t.Fatalf("instalador não lidera o próprio grupo: pid %s, pgid %s", fields[0], fields[1])
	}
	if fields[1] == strconv.Itoa(syscall.Getpgrp()) {
		t.Fatal("instalador no mesmo grupo de processos do kit")
	}
}

func TestGrandchildHoldingOutputEndsAfterWaitDelay(t *testing.T) {
	home := t.TempDir()
	pidFile := filepath.Join(home, "pid")
	start := time.Now()
	_, err := runCommand(context.Background(), home, "/bin/sh", nil, "-c", `sleep 30 & echo $! > "$0"; exit 0`, pidFile)
	elapsed := time.Since(start)
	if !errors.Is(err, errOutputOpen) {
		t.Fatalf("esperado saída aberta com neto segurando o stdout; obtido %v", err)
	}
	if elapsed >= 2*waitDelay {
		t.Fatalf("runCommand demorou %s com WaitDelay de %s", elapsed, waitDelay)
	}
	procWaitGone(t, procReadPIDs(t, pidFile)[0])
}

func TestProbeTimeoutBlocksWithReason(t *testing.T) {
	home, bin := isolated(t)
	script := `#!/bin/sh
sleep 30
`
	if err := os.WriteFile(filepath.Join(bin, "claude"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	if err := save(home, history{"claude": {Current: ClaudeVersion}}); err != nil {
		t.Fatal(err)
	}
	previous := probeTimeout
	probeTimeout = 100 * time.Millisecond
	t.Cleanup(func() { probeTimeout = previous })
	start := time.Now()
	changes, err := Preview(home)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatal(err)
	}
	if !changes[0].Blocked || changes[0].Reason != "não verificado (tempo esgotado)" {
		t.Fatalf("sondagem expirada sem motivo esperado: %+v", changes[0])
	}
	if elapsed >= waitDelay {
		t.Fatalf("sondagem presa por %s com prazo de 100 ms", elapsed)
	}
}

func procReadPIDs(t *testing.T, path string) []int {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var pids []int
	for _, field := range strings.Fields(string(b)) {
		pid, err := strconv.Atoi(field)
		if err != nil {
			t.Fatalf("pid inválido %q: %v", field, err)
		}
		pids = append(pids, pid)
	}
	if len(pids) == 0 {
		t.Fatal("nenhum pid registrado pelo comando falso")
	}
	return pids
}

func procWaitGone(t *testing.T, target int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		err := syscall.Kill(target, 0)
		if errors.Is(err, syscall.ESRCH) {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("processo ou grupo %d continua vivo: %v", target, err)
		}
		time.Sleep(20 * time.Millisecond)
	}
}
