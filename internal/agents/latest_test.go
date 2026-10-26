package agents

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

func TestParseLatestRejectsMalformedResponses(t *testing.T) {
	for _, tc := range []struct{ name, body, want string }{
		{"claude", " 2.1.310\n", "2.1.310"},
		{"codex", `{"tag_name":"rust-v0.164.0"}`, "0.164.0"},
		{"claude", "<html>2.1.310</html>", ""},
		{"codex", `<html>rust-v0.164.0</html>`, ""},
		{"codex", `{"tag_name":"v0.164.0"}`, ""},
		{"claude", "junk", ""},
	} {
		got, err := parseLatest(tc.name, []byte(tc.body))
		if got != tc.want || (err != nil) != (tc.want == "") {
			t.Errorf("%s %q: got %q, %v; want %q", tc.name, tc.body, got, err, tc.want)
		}
	}
}

func TestLatestLookupFailureKeepsManagedAndBlocksNew(t *testing.T) {
	home, _ := isolated(t)
	fetchLatest = func(string) (string, error) { return "", errors.New("offline") }
	path := binPath(home, "claude")
	if err := os.WriteFile(path, []byte("managed"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := save(home, history{"claude": {Current: "2.1.305"}}); err != nil {
		t.Fatal(err)
	}
	oldRun := runCommand
	t.Cleanup(func() { runCommand = oldRun })
	runCommand = func(context.Context, string, string, []byte, ...string) (string, error) {
		return "2.1.305 (Claude Code)", nil
	}
	changes, err := Preview(home)
	if err != nil {
		t.Fatal(err)
	}
	if changes[0].Blocked || changes[0].Target != "2.1.305" || changes[0].Current != changes[0].Target {
		t.Fatalf("managed offline: %+v", changes[0])
	}
	if !changes[1].Blocked || changes[1].Reason != "não consegui consultar a versão mais recente (sem rede?)" {
		t.Fatalf("new offline: %+v", changes[1])
	}
}

func TestPreviewAdoptsOnlyNativeClaude(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("native Claude Unix symlink")
	}
	home, bin := isolated(t)
	versions := filepath.Join(home, ".local", "share", "claude", "versions")
	if err := os.MkdirAll(versions, 0o700); err != nil {
		t.Fatal(err)
	}
	actual := filepath.Join(versions, "2.1.309")
	if err := os.WriteFile(actual, []byte("native"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(actual, filepath.Join(bin, "claude")); err != nil {
		t.Fatal(err)
	}
	oldRun := runCommand
	t.Cleanup(func() { runCommand = oldRun })
	runCommand = func(context.Context, string, string, []byte, ...string) (string, error) {
		return "2.1.309 (Claude Code)", nil
	}
	changes, err := Preview(home)
	if err != nil {
		t.Fatal(err)
	}
	if changes[0].Blocked || changes[0].Current != "2.1.309" {
		t.Fatalf("native Claude blocked: %+v", changes[0])
	}
	if err := os.Symlink(actual, filepath.Join(bin, "codex")); err != nil {
		t.Fatal(err)
	}
	changes, err = Preview(home)
	if err != nil {
		t.Fatal(err)
	}
	if !changes[1].Blocked || !strings.Contains(changes[1].Reason, "não gerenciado") {
		t.Fatalf("Codex adopted: %+v", changes[1])
	}
	if err := os.Remove(filepath.Join(bin, "claude")); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(home, "other")
	if err := os.WriteFile(outside, []byte("other"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(bin, "claude")); err != nil {
		t.Fatal(err)
	}
	changes, err = Preview(home)
	if err != nil {
		t.Fatal(err)
	}
	if !changes[0].Blocked || !strings.Contains(changes[0].Reason, "não gerenciado") {
		t.Fatalf("foreign Claude adopted: %+v", changes[0])
	}
}

func TestLatestLookupCachedAcrossPreviews(t *testing.T) {
	home, _ := isolated(t)
	calls := map[string]int{}
	var mu sync.Mutex
	fetchLatest = func(name string) (string, error) {
		mu.Lock()
		calls[name]++
		mu.Unlock()
		if name == "claude" {
			return "2.1.310", nil
		}
		return "0.164.0", nil
	}
	for range 2 {
		if _, err := Preview(home); err != nil {
			t.Fatal(err)
		}
	}
	if calls["claude"] != 1 || calls["codex"] != 1 {
		t.Fatalf("lookup repeated: %+v", calls)
	}
}

func TestLatestLookupFailureCachedAcrossPreviews(t *testing.T) {
	home, _ := isolated(t)
	var mu sync.Mutex
	calls := 0
	fetchLatest = func(string) (string, error) {
		mu.Lock()
		calls++
		mu.Unlock()
		return "", errors.New("offline")
	}
	for range 2 {
		if _, err := Preview(home); err != nil {
			t.Fatal(err)
		}
	}
	if calls != 2 {
		t.Fatalf("falhas não cacheadas por agente: %d", calls)
	}
}
