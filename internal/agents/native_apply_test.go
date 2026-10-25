package agents

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestApplyNativeClaudeRecordsObservedPrevious(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("native Claude Unix symlink")
	}
	home, bin := isolated(t)
	versions := filepath.Join(home, ".local", "share", "claude", "versions")
	if err := os.MkdirAll(versions, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, version := range []string{"2.1.309", "2.1.306"} {
		if err := os.WriteFile(filepath.Join(versions, version), []byte(version), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	launcher := filepath.Join(bin, "claude")
	if err := os.Symlink(filepath.Join(versions, "2.1.309"), launcher); err != nil {
		t.Fatal(err)
	}
	oldFetch, oldRun := fetchScript, runCommand
	t.Cleanup(func() { fetchScript, runCommand = oldFetch, oldRun })
	fetchScript = func(string) ([]byte, error) { return []byte("official installer"), nil }
	runCommand = func(_ context.Context, _ string, path string, input []byte, args ...string) (string, error) {
		if len(input) > 0 {
			if path != "bash" || strings.Join(args, " ") != "-s -- latest" {
				t.Fatalf("installer: %s %v", path, args)
			}
			if err := os.Remove(launcher); err != nil {
				return "", err
			}
			return "", os.Symlink(filepath.Join(versions, "2.1.306"), launcher)
		}
		resolved, err := filepath.EvalSymlinks(path)
		if err != nil {
			return "", err
		}
		return filepath.Base(resolved) + " (Claude Code)", nil
	}
	changes, err := Preview(home)
	if err != nil {
		t.Fatal(err)
	}
	if changes[0].Blocked || changes[0].Current != "2.1.309" {
		t.Fatalf("native preview: %+v", changes[0])
	}
	if err := Apply(home, changes[:1]); err != nil {
		t.Fatal(err)
	}
	h, err := load(home)
	if err != nil {
		t.Fatal(err)
	}
	if h["claude"].Previous != "2.1.309" || h["claude"].Current != "2.1.306" || h["claude"].Pending {
		t.Fatalf("native history: %+v", h["claude"])
	}
}
