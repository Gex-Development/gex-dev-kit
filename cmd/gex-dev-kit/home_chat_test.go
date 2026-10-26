package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestHomeChatPlansHooksChosenInTheChat(t *testing.T) {
	home := t.TempDir()
	s := session{deps: realDeps()}
	preset, err := homeRequest()
	if err != nil {
		t.Fatal(err)
	}
	d := chatDeps(s, home, preset)
	req := d.Preset
	req.Home = home
	req.Hooks = []string{"gex-conventional-commits"}
	req.SkipAgentUpdate = true
	plan, err := d.Install.Plan(req)
	if err != nil {
		t.Fatalf("escolher um hook no chat da home não pode quebrar o plano: %v", err)
	}
	var copied bool
	for _, file := range plan.Files {
		path := filepath.ToSlash(file.Path)
		if strings.HasSuffix(path, "gex-dev-kit/bin/gex-dev-kit") || strings.HasSuffix(path, "gex-dev-kit/bin/gex-dev-kit.exe") {
			copied = true
		}
	}
	if !copied {
		t.Fatalf("o plano com hook deve copiar o executável do kit para o HOME: %+v", plan.Files)
	}
}
