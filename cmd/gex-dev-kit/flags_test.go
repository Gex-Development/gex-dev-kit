package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/terminal"
)

func flagsParse(args []string) (string, options, error) {
	command, rest := splitCommand(args)
	o, err := parseOptions(command, rest)
	return command, o, err
}

func TestNoArgsNonTTYShowsHelp(t *testing.T) {
	code, out, errors := invoke(t)
	if code != 0 || errors != "" || !strings.Contains(out, "install") {
		t.Fatalf("ajuda: exit=%d out=%q err=%q", code, out, errors)
	}
}

func TestSubcommandHelpWritesOnlyStdout(t *testing.T) {
	for _, args := range [][]string{
		{"install", "--help"}, {"doctor", "-h"}, {"rollback", "--help"},
		{"rollback-agents", "--help"}, {"skills", "--help"},
		{"hooks", "-h"}, {"hook", "commit-msg", "--help"},
		{"addons", "--help"}, {"githooks", "-h"}, {"ds", "--help"},
	} {
		code, out, errors := invoke(t, args...)
		if code != 0 || errors != "" || !strings.Contains(out, args[0]) {
			t.Fatalf("%v: exit=%d out=%q err=%q", args, code, out, errors)
		}
	}
}

func TestVersionShortFlag(t *testing.T) {
	code, out, errors := invoke(t, "-v")
	if code != 0 || strings.TrimSpace(out) != version || errors != "" {
		t.Fatalf("-v: exit=%d out=%q err=%q", code, out, errors)
	}
}

func TestParseOptionsTabela(t *testing.T) {
	home := t.TempDir()
	repo := t.TempDir()
	casos := []struct {
		nome    string
		args    []string
		comando string
		erro    bool
		ok      func(o options) bool
	}{
		{"instalação com opções depois do comando", []string{"install", "--yes", "--home", home, "--skill", "tdd", "--skill", "x", "--hook", "gex-conventional-commits", "--skip-agent-update"}, "install", false,
			func(o options) bool {
				return o.Yes && o.Home == home && len(o.Skills) == 2 && len(o.Hooks) == 1 && o.SkipAgentUpdate
			}},
		{"global antes do comando", []string{"--quiet", "install", "--yes"}, "install", false,
			func(o options) bool { return o.Quiet && o.Yes }},
		{"--ui=cli global", []string{"--ui=cli", "skills"}, "skills", false,
			func(o options) bool { return o.UI == terminal.ModeCLI }},
		{"githooks com repetíveis", []string{"githooks", "--repo", repo, "--hook", "gex-pre-push", "--run", "go test ./...", "--pre-commit-run", "go vet ./..."}, "githooks", false,
			func(o options) bool {
				return o.Repo == repo && len(o.PrePush) == 1 && len(o.PreCommit) == 1 && len(o.Hooks) == 1
			}},
		{"addons com nome e json", []string{"addons", "caveman", "--json"}, "addons", false,
			func(o options) bool { return o.JSON && len(o.Args) == 1 && o.Args[0] == "caveman" }},
		{"doctor json home", []string{"doctor", "--json", "--home", home}, "doctor", false,
			func(o options) bool { return o.JSON && o.Home == home }},
		{"skills json", []string{"skills", "--json"}, "skills", false,
			func(o options) bool { return o.JSON }},
		{"hooks json com no-anim", []string{"hooks", "--no-anim", "--json"}, "hooks", false,
			func(o options) bool { return o.JSON && o.NoAnim }},
		{"help com comando", []string{"help", "install"}, "help", false,
			func(o options) bool { return len(o.Args) == 1 && o.Args[0] == "install" }},
		{"hook aceita caminho", []string{"hook", "commit-msg", "/abs/COMMIT_EDITMSG"}, "hook", false,
			func(o options) bool { return len(o.Args) == 2 }},
		{"versão curta", []string{"-v"}, "", false,
			func(o options) bool { return o.Version }},
		{"help de raiz", []string{"--help"}, "", false,
			func(o options) bool { return o.Help }},
		{"json em install", []string{"install", "--json"}, "install", true, nil},
		{"tui com quiet", []string{"--ui=tui", "install", "--quiet"}, "install", true, nil},
		{"home relativo", []string{"install", "--home", "relativo"}, "install", true, nil},
		{"repo relativo", []string{"doctor", "--repo", "relativo"}, "doctor", true, nil},
		{"home e repo juntos", []string{"rollback", "--home", home, "--repo", repo}, "rollback", true, nil},
		{"modo inválido", []string{"--ui=gui", "skills"}, "skills", true, nil},
		{"opção raiz desconhecida", []string{"--invalid-flag"}, "", true, nil},
		{"comando desconhecido", []string{"gizmo"}, "gizmo", true, nil},
		{"argumento extra", []string{"install", "extra"}, "install", true, nil},
		{"valor ausente", []string{"install", "--skill"}, "install", true, nil},
		{"booleana com valor", []string{"install", "--yes=1"}, "install", true, nil},
		{"yes em rollback-agents com skill", []string{"rollback-agents", "--skill", "x"}, "rollback-agents", true, nil},
		{"help de comando desconhecido", []string{"help", "nope"}, "help", true, nil},
	}
	for _, caso := range casos {
		comando, o, err := flagsParse(caso.args)
		if comando != caso.comando {
			t.Fatalf("%s: comando=%q; quero %q", caso.nome, comando, caso.comando)
		}
		if caso.erro {
			if err == nil {
				t.Fatalf("%s: esperado erro de uso, obteve opções %+v", caso.nome, o)
			}
			continue
		}
		if err != nil {
			t.Fatalf("%s: erro inesperado: %v", caso.nome, err)
		}
		if !caso.ok(o) {
			t.Fatalf("%s: opções incorretas %+v", caso.nome, o)
		}
	}
}

func TestFlagRaizDesconhecidaSaiComUso(t *testing.T) {
	code, out, errors := invoke(t, "--invalid-flag")
	if code != 2 || out != "" || !strings.Contains(errors, "opção desconhecida: --invalid-flag") || !strings.Contains(errors, "Use --help para opções.") {
		t.Fatalf("--invalid-flag: exit=%d out=%q err=%q", code, out, errors)
	}
}

func TestFlagHomeRelativoSaiComUso(t *testing.T) {
	code, out, errors := invoke(t, "install", "--home", "relativo", "--yes")
	if code != 2 || out != "" || !strings.Contains(errors, "--home") {
		t.Fatalf("--home relativo: exit=%d out=%q err=%q", code, out, errors)
	}
	if filepath.IsAbs("relativo") {
		t.Fatalf("caminho de teste deve ser relativo")
	}
}

func TestFlagJSONSoEmComandosDeLeitura(t *testing.T) {
	code, out, errors := invoke(t, "install", "--json")
	if code != 2 || out != "" || errors == "" {
		t.Fatalf("--json em install: exit=%d out=%q err=%q", code, out, errors)
	}
}

func TestFlagUITuiComYesSaiComUso(t *testing.T) {
	code, out, errors := invoke(t, "--ui=tui", "install", "--yes")
	if code != 2 || out != "" || !strings.Contains(errors, "--ui=tui") {
		t.Fatalf("--ui=tui com --yes: exit=%d out=%q err=%q", code, out, errors)
	}
}

func TestFlagUITuiSemTerminalPedeCLI(t *testing.T) {
	var out, errors bytes.Buffer
	code := run([]string{"--ui=tui"}, strings.NewReader(""), &out, &errors)
	if code != 2 || out.Len() != 0 || !strings.Contains(errors.String(), "use --ui=cli") {
		t.Fatalf("--ui=tui sem TTY: exit=%d out=%q err=%q", code, out.String(), errors.String())
	}
}

func TestHelpTemDoisAQuatroExemplos(t *testing.T) {
	for comando, spec := range helpSpecs {
		if n := len(spec.exemplos); n < 2 || n > 4 {
			t.Fatalf("%s: %d exemplos; quero de 2 a 4", comando, n)
		}
	}
}
