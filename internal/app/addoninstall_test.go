package app

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestAddonSecretsNeededUniqueOrdered(t *testing.T) {
	got, err := AddonSecretsNeeded([]string{"brave-search", "context7", "mapbox-devkit", "brave-search"})
	if err != nil || len(got) != 2 || got[0].Env != "BRAVE_API_KEY" || got[1].Env != "MAPBOX_ACCESS_TOKEN" {
		t.Fatalf("secrets = %+v, %v", got, err)
	}
	if _, err := AddonSecretsNeeded([]string{"unknown"}); err == nil {
		t.Fatal("addon desconhecido aceito")
	}
}

func TestPlanAddonsWithSecretsHidesValueAndLength(t *testing.T) {
	plan, err := PlanAddonsWithSecrets([]string{"brave-search", "mapbox-devkit"}, []string{"claude", "codex"}, func(string) (string, error) { return "/bin/tool", nil }, map[string]string{"BRAVE_API_KEY": "fake-key-123", "MAPBOX_ACCESS_TOKEN": "fake-key-123"})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Steps) != 6 || len(plan.secrets) != 1 {
		t.Fatalf("passos = %d; segredos = %d", len(plan.Steps), len(plan.secrets))
	}
	for _, step := range plan.Steps {
		if strings.Contains(step.Display, "fake-key-123") || strings.Contains(step.Display, "{{valor}}") || strings.Contains(step.Display, "fake-key-") {
			t.Fatalf("display vazou: %s", step.Display)
		}
		if strings.Contains(strings.Join(step.Argv, " "), "fake-key-123") {
			if !strings.Contains(step.Display, "=••••••••") {
				t.Fatalf("máscara de oito caracteres ausente: %s", step.Display)
			}
		}
	}
	if !strings.Contains(strings.Join(plan.Steps[1].Argv, " "), "BRAVE_API_KEY=fake-key-123") || !strings.Contains(strings.Join(plan.Steps[4].Argv, " "), "MAPBOX_ACCESS_TOKEN=fake-key-123") || !strings.Contains(strings.Join(plan.Steps[5].Argv, " "), "MAPBOX_ACCESS_TOKEN=fake-key-123") {
		t.Fatalf("argv de registro sem chave")
	}
}

func TestPlanAddonsWithSecretsRejectsMissingKeyWithoutPlan(t *testing.T) {
	for _, secrets := range []map[string]string{nil, {"BRAVE_API_KEY": ""}} {
		plan, err := PlanAddonsWithSecrets([]string{"context7", "brave-search"}, []string{"claude"}, func(string) (string, error) { return "/bin/tool", nil }, secrets)
		var missing AddonSecretMissing
		if !errors.As(err, &missing) || missing.Env != "BRAVE_API_KEY" || len(plan.Steps) != 0 {
			t.Fatalf("chave ausente: plan=%+v err=%v", plan, err)
		}
	}
}

func TestInstallMCPKeyReplacementAcceptsMissingPreviousRegistration(t *testing.T) {
	plan, err := PlanAddonsWithSecrets([]string{"brave-search"}, []string{"claude", "codex"}, func(string) (string, error) { return "/bin/tool", nil }, map[string]string{"BRAVE_API_KEY": "fake-key-123"})
	if err != nil {
		t.Fatal(err)
	}
	var calls []string
	result := InstallAddons(plan, t.TempDir(), func(_ context.Context, _ string, argv, _ []string) (string, error) {
		calls = append(calls, strings.Join(argv, " "))
		if slices.Contains(argv, "remove") {
			return "No MCP server named brave-search in user scope", errors.New("exit status 1")
		}
		return "registrado", nil
	}, nil)
	if len(calls) != 3 || calls[0] != "claude mcp remove --scope user brave-search" || !strings.HasPrefix(calls[1], "claude mcp add --scope user brave-search -e BRAVE_API_KEY=fake-key-123 --") || !strings.HasPrefix(calls[2], "codex mcp add brave-search --env BRAVE_API_KEY=fake-key-123 --") {
		t.Fatalf("ordem de reinstalação incorreta: %v", calls)
	}
	if result.Steps[0].Status != AddonOK || result.Steps[0].Output != "já configurado" || result.Steps[1].Status != AddonOK || result.Steps[2].Status != AddonOK {
		t.Fatalf("reinstalação falhou: %+v", result.Steps)
	}
}

func TestInstallMCPKeyRedactsOutputAndErrorBeforeProgress(t *testing.T) {
	plan, err := PlanAddonsWithSecrets([]string{"mapbox-devkit"}, []string{"codex"}, func(string) (string, error) { return "/bin/tool", nil }, map[string]string{"MAPBOX_ACCESS_TOKEN": "fake-key-123"})
	if err != nil {
		t.Fatal(err)
	}
	var events []AddonStepResult
	result := InstallAddons(plan, t.TempDir(), func(context.Context, string, []string, []string) (string, error) {
		return "processo fake-key-123", errors.New("falhou fake-key-123")
	}, func(event AddonStepResult) { events = append(events, event) })
	for _, item := range append(result.Steps, events...) {
		if strings.Contains(item.Output, "fake-key-123") || strings.Contains(item.Step.Display, "fake-key-123") || strings.Contains(strings.Join(item.Step.Argv, " "), "fake-key-123") || !strings.Contains(item.Output, "[oculto]") && item.Status == AddonFailed {
			t.Fatalf("vazou chave: %+v", item)
		}
	}
	result = InstallAddons(plan, t.TempDir(), func(context.Context, string, []string, []string) (string, error) {
		return "", errors.New("falhou fake-key-123")
	}, nil)
	if strings.Contains(result.Steps[0].Output, "fake-key-123") || !strings.Contains(result.Steps[0].Output, "[oculto]") {
		t.Fatalf("vazou erro: %+v", result.Steps)
	}
}

func TestPlanAddonsSecretsSurviveCallerClearingMap(t *testing.T) {
	secrets := map[string]string{"BRAVE_API_KEY": "fake-key-123"}
	plan, err := PlanAddonsWithSecrets([]string{"brave-search"}, []string{"codex"}, func(string) (string, error) { return "/bin/tool", nil }, secrets)
	if err != nil {
		t.Fatal(err)
	}
	clear(secrets)
	result := InstallAddons(plan, t.TempDir(), func(context.Context, string, []string, []string) (string, error) {
		return "fake-key-123", nil
	}, nil)
	if result.Steps[0].Output != "[oculto]" {
		t.Fatalf("map limpo vazou chave: %+v", result.Steps)
	}
}

func TestTerraformPullFailureSkipsBothClientsAndMissingDocker(t *testing.T) {
	plan, err := PlanAddons([]string{"terraform"}, []string{"claude", "codex"}, func(string) (string, error) { return "/bin/tool", nil })
	if err != nil {
		t.Fatal(err)
	}
	var calls int
	result := InstallAddons(plan, t.TempDir(), func(context.Context, string, []string, []string) (string, error) {
		calls++
		return "pull falhou", errors.New("sem daemon")
	}, nil)
	if calls != 1 || len(result.Steps) != 3 || result.Steps[0].Status != AddonFailed || result.Steps[1].Status != AddonSkipped || result.Steps[2].Status != AddonSkipped {
		t.Fatalf("pull falhou: calls=%d steps=%+v", calls, result.Steps)
	}
	plan, err = PlanAddons([]string{"terraform"}, []string{"claude", "codex"}, func(name string) (string, error) {
		if name == "docker" {
			return "", errors.New("ausente")
		}
		return "/bin/tool", nil
	})
	if err != nil || len(plan.Manual) != 1 || plan.Manual[0].Text != "Instale e ligue o Docker e rode de novo." {
		t.Fatalf("nota docker ausente: %+v, %v", plan, err)
	}
	for _, step := range plan.Steps {
		if step.Available {
			t.Fatalf("passo disponível sem Docker: %+v", step)
		}
	}
}

func TestPlanAddonsWithoutKeysSupportsPublicMCPs(t *testing.T) {
	plan, err := PlanAddons([]string{"context7", "terraform"}, []string{"claude"}, func(string) (string, error) { return "/bin/tool", nil })
	if err != nil || len(plan.Steps) != 3 || len(plan.secrets) != 0 {
		t.Fatalf("plano público: %+v, %v", plan, err)
	}
}

func TestPlanAddonsDeduplicatesSharedCLIAndIgnoresOMP(t *testing.T) {
	plan, err := PlanAddons([]string{"spec-kit", "caveman"}, []string{"claude", "codex", "omp", "claude"}, func(string) (string, error) { return "/bin/tool", nil })
	if err != nil {
		t.Fatal(err)
	}
	var display []string
	for _, step := range plan.Steps {
		display = append(display, step.Display)
	}
	want := []string{
		"uv tool install specify-cli --from git+https://github.com/github/spec-kit.git@dc2cec4f568d0f6e2cdfb8571c69116bc55c1aaf",
		"claude plugin marketplace add JuliusBrussee/caveman#v4.0.1",
		"claude plugin install caveman@caveman",
		"npx -y github:JuliusBrussee/caveman#bad48a0e1df55e05e3f52249afa35ecc1909ee76 -- --only codex",
	}
	if !slices.Equal(display, want) || len(plan.Manual) != 2 {
		t.Fatalf("plan = %+v; want display %v and two manual notes", plan, want)
	}
	if _, err := PlanAddons([]string{"unknown"}, []string{"claude"}, exec.LookPath); err == nil {
		t.Fatal("unknown addon accepted")
	}
}

func TestPlanAddonsReportsMissingExecutableWithoutError(t *testing.T) {
	plan, err := PlanAddons([]string{"ponytail"}, []string{"claude"}, func(name string) (string, error) { return "", errors.New("missing " + name) })
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(plan.Missing, []string{"claude"}) || len(plan.Steps) != 2 || plan.Steps[0].Available || plan.Steps[1].Available {
		t.Fatalf("missing tools = %+v", plan)
	}
}

func TestInstallAddonsFailureSkipsDependentStepsButContinuesOtherAddons(t *testing.T) {
	plan, err := PlanAddons([]string{"caveman", "superpowers"}, []string{"claude"}, func(string) (string, error) { return "/bin/tool", nil })
	if err != nil {
		t.Fatal(err)
	}
	var calls []string
	var events []AddonStepResult
	result := InstallAddons(plan, os.Getenv("HOME"), func(_ context.Context, _ string, argv, _ []string) (string, error) {
		calls = append(calls, strings.Join(argv, " "))
		if len(calls) == 1 {
			return "falhou", errors.New("erro")
		}
		return "feito", nil
	}, func(event AddonStepResult) { events = append(events, event) })
	if len(result.Steps) != 3 || result.Steps[0].Status != AddonFailed || result.Steps[1].Status != AddonSkipped || result.Steps[2].Status != AddonOK {
		t.Fatalf("results = %+v", result.Steps)
	}
	if len(calls) != 2 || !strings.Contains(calls[1], "superpowers") {
		t.Fatalf("calls = %v", calls)
	}
	if len(events) != 6 || events[0].Status != AddonPending || events[1].Status != AddonFailed || events[2].Status != AddonPending || events[3].Status != AddonSkipped {
		t.Fatalf("progress = %+v", events)
	}
}

func TestInstallAddonsUnavailableSkipsWithoutInvokingRunner(t *testing.T) {
	plan, err := PlanAddons([]string{"superpowers"}, []string{"claude", "codex"}, func(string) (string, error) { return "", errors.New("missing") })
	if err != nil {
		t.Fatal(err)
	}
	result := InstallAddons(plan, os.Getenv("HOME"), func(context.Context, string, []string, []string) (string, error) {
		t.Fatal("runner invoked")
		return "", nil
	}, nil)
	if len(result.Steps) != 1 || result.Steps[0].Status != AddonSkipped || len(result.Manual) != 1 {
		t.Fatalf("result = %+v", result)
	}
}

func TestInstallAddonsIsolatesHomeAndSanitizesOutput(t *testing.T) {
	plan, err := PlanAddons([]string{"superpowers"}, []string{"claude"}, func(string) (string, error) { return "/bin/tool", nil })
	if err != nil {
		t.Fatal(err)
	}
	home := t.TempDir()
	result := InstallAddons(plan, home, func(_ context.Context, dir string, _ []string, env []string) (string, error) {
		if dir != home {
			t.Errorf("working directory = %q; want %q", dir, home)
		}
		for _, entry := range []string{"HOME=" + home, "USERPROFILE=" + home, "XDG_CONFIG_HOME=" + filepath.Join(home, ".config"), "XDG_DATA_HOME=" + filepath.Join(home, ".local", "share")} {
			if !slices.Contains(env, entry) {
				t.Errorf("isolated environment missing %s", entry)
			}
		}
		if !slices.ContainsFunc(env, func(s string) bool { return strings.HasPrefix(s, "PATH=") }) {
			t.Error("PATH missing")
		}
		return strings.Repeat("older\n", 24) + "\x1b[31mtoken=abc\x1b[0m\nBearer hidden\nkey=secret", nil
	}, nil)
	out := result.Steps[0].Output
	if strings.Contains(out, "abc") || strings.Contains(out, "secret") || strings.Contains(out, "hidden") || strings.Contains(out, "\x1b") || strings.Count(out, "\n") > 19 || len(out) > 2048 {
		t.Fatalf("unsafe output %q", out)
	}
}

func TestAddonEnvironmentFiltersSecretsAndKeepsRealHome(t *testing.T) {
	home := t.TempDir()
	environ := []string{"Path=/bin", "HTTP_PROXY=https://proxy", "TERM=xterm", "GH_TOKEN=secret", "NPM_TOKEN=secret", "CLAUDE_CONFIG_DIR=/external", "CODEX_HOME=/external", "HOME=/original", "XDG_CONFIG_HOME=/original/config"}
	got := addonEnvironment(environ, home, filepath.Join(home, "."))
	for _, entry := range []string{"Path=/bin", "HTTP_PROXY=https://proxy", "TERM=xterm", "HOME=/original", "XDG_CONFIG_HOME=/original/config", "CODEX_HOME=" + filepath.Join(home, ".codex"), "GIT_TERMINAL_PROMPT=0"} {
		if !slices.Contains(got, entry) {
			t.Errorf("missing %s: %v", entry, got)
		}
	}
	for _, entry := range got {
		if strings.Contains(entry, "secret") || strings.Contains(entry, "/external") || strings.HasPrefix(entry, "CLAUDE_CONFIG_DIR=") {
			t.Errorf("leaked secret or external configuration: %s", entry)
		}
	}
}

func TestAddonEnvironmentIsolatedHome(t *testing.T) {
	home := t.TempDir()
	got := addonEnvironment([]string{"HOME=/real", "USERPROFILE=/real", "XDG_DATA_HOME=/real/data", "GH_TOKEN=secret", "nPm_CoNfIg_userconfig=/real/npmrc"}, home, "/real")
	for _, entry := range []string{"HOME=" + home, "USERPROFILE=" + home, "XDG_CONFIG_HOME=" + filepath.Join(home, ".config"), "XDG_DATA_HOME=" + filepath.Join(home, ".local", "share"), "XDG_CACHE_HOME=" + filepath.Join(home, ".cache"), "XDG_STATE_HOME=" + filepath.Join(home, ".local", "state"), "XDG_BIN_HOME=" + filepath.Join(home, ".local", "bin"), "CODEX_HOME=" + filepath.Join(home, ".codex"), "GIT_TERMINAL_PROMPT=0"} {
		if !slices.Contains(got, entry) {
			t.Errorf("missing %s: %v", entry, got)
		}
	}
	for _, entry := range got {
		if strings.Contains(entry, "/real") || strings.Contains(entry, "secret") || strings.HasPrefix(entry, "CLAUDE_CONFIG_DIR=") {
			t.Errorf("leaked environment: %s", entry)
		}
	}
}

func TestAddonEnvironmentNeverRedirectsClaudeGlobalState(t *testing.T) {
	home := t.TempDir()
	for _, real := range []string{home, "/other"} {
		got := addonEnvironment([]string{"HOME=/h", "CLAUDE_CONFIG_DIR=/elsewhere"}, home, real)
		for _, entry := range got {
			if strings.HasPrefix(strings.ToUpper(entry), "CLAUDE_CONFIG_DIR=") {
				t.Fatalf("CLAUDE_CONFIG_DIR move ~/.claude.json para ~/.claude/.claude.json e esconde o MCP registrado: %s", entry)
			}
		}
	}
}

func TestInstallAddonsRunnerSeesTimeout(t *testing.T) {
	plan := AddonPlan{Steps: []AddonPlanStep{
		{Addon: "test", Harness: "claude", Argv: []string{"test"}, Available: true},
		{Addon: "test", Harness: "claude", Argv: []string{"test", "next"}, Available: true},
	}}
	result := installAddonsWithTimeout(plan, t.TempDir(), func(ctx context.Context, _ string, _, _ []string) (string, error) {
		<-ctx.Done()
		return "signal: killed", ctx.Err()
	}, nil, time.Millisecond)
	if result.Steps[0].Status != AddonFailed || result.Steps[0].Output != "tempo esgotado (1ms)" || result.Steps[1].Status != AddonSkipped {
		t.Fatalf("timeout = %+v", result.Steps)
	}
}

func TestExecAddonRunnerDoesNotInvokeShell(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("echo não é um executável no Windows")
	}
	output, err := ExecAddonRunner(context.Background(), t.TempDir(), []string{"echo", "ok", "&&", "echo", "bad"}, os.Environ())
	if err != nil || strings.TrimSpace(output) != "ok && echo bad" {
		t.Fatalf("argv deve chegar literal, sem shell: %q, %v", output, err)
	}
}

func TestInstallAddonsLimitsLargeOutputToTwoKilobytes(t *testing.T) {
	plan := AddonPlan{Steps: []AddonPlanStep{{Addon: "test", Harness: "claude", Argv: []string{"test"}, Available: true}}}
	result := InstallAddons(plan, os.Getenv("HOME"), func(context.Context, string, []string, []string) (string, error) {
		return strings.Repeat("a", 3000) + "token=private", nil
	}, nil)
	output := result.Steps[0].Output
	if len(output) > 2048 || strings.Contains(output, "private") || !strings.HasSuffix(output, "token=[oculto]") {
		t.Fatalf("unsafe output length %d, suffix %q", len(output), output[len(output)-40:])
	}
}

func TestInstallAddonsExecutesFakeCLIInIsolatedHome(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture requires Unix")
	}
	bin, home := t.TempDir(), t.TempDir()
	script := filepath.Join(bin, "claude")
	if err := os.WriteFile(script, []byte("#!/bin/sh\npwd > \"$HOME/pwd\"\nenv > \"$HOME/env\"\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("GH_TOKEN", "hidden")
	t.Setenv("CLAUDE_CONFIG_DIR", "/external")
	plan, err := PlanAddons([]string{"superpowers"}, []string{"claude"}, exec.LookPath)
	if err != nil {
		t.Fatal(err)
	}
	result := InstallAddons(plan, home, ExecAddonRunner, nil)
	if result.Steps[0].Status != AddonOK {
		t.Fatalf("fake CLI failed: %+v", result.Steps)
	}
	pwd, err := os.ReadFile(filepath.Join(home, "pwd"))
	resolved, resolveErr := filepath.EvalSymlinks(home)
	if err != nil || resolveErr != nil || strings.TrimSpace(string(pwd)) != resolved {
		t.Fatalf("working directory %q, %v, %v", pwd, err, resolveErr)
	}
	env, err := os.ReadFile(filepath.Join(home, "env"))
	if err != nil || strings.Contains(string(env), "GH_TOKEN") || strings.Contains(string(env), "/external") || strings.Contains(string(env), "CLAUDE_CONFIG_DIR") {
		t.Fatalf("unsafe environment: %q, %v", env, err)
	}
}

func TestPlanMCPRequiresRuntimeBinary(t *testing.T) {
	var checked []string
	plan, err := PlanAddons([]string{"context7"}, []string{"claude", "codex"}, func(name string) (string, error) {
		checked = append(checked, name)
		if name == "npx" {
			return "", errors.New("missing")
		}
		return "/bin/" + name, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(plan.Missing, []string{"npx"}) || !slices.Contains(checked, "npx") || len(plan.Manual) != 1 || plan.Manual[0].Text != "Instale Node.js (npx) e rode de novo." {
		t.Fatalf("missing/note = %+v; checked = %v", plan, checked)
	}
	for _, step := range plan.Steps {
		if step.Available {
			t.Fatalf("passo disponível sem npx: %+v", step)
		}
	}
	result := InstallAddons(plan, t.TempDir(), func(context.Context, string, []string, []string) (string, error) {
		t.Fatal("runner chamado sem dependência")
		return "", nil
	}, nil)
	for _, step := range result.Steps {
		if step.Status != AddonSkipped {
			t.Fatalf("passo não pulado: %+v", step)
		}
	}
}

func TestInstallMCPAlreadyExistsOnlyAcceptedForClaudeAndContinuesCodex(t *testing.T) {
	plan, err := PlanAddons([]string{"context7"}, []string{"claude", "codex"}, func(string) (string, error) { return "/bin/fake", nil })
	if err != nil {
		t.Fatal(err)
	}
	var calls []string
	result := InstallAddons(plan, t.TempDir(), func(_ context.Context, _ string, argv, _ []string) (string, error) {
		calls = append(calls, argv[0])
		if argv[0] == "claude" {
			return "MCP server context7 ALREADY EXISTS in user config", errors.New("exit status 1")
		}
		return "adicionado", nil
	}, nil)
	if !slices.Equal(calls, []string{"claude", "codex"}) || len(result.Steps) != 2 || result.Steps[0].Status != AddonOK || result.Steps[0].Output != "já configurado" || result.Steps[1].Status != AddonOK {
		t.Fatalf("already exists = %+v; calls = %v", result, calls)
	}
	result = InstallAddons(plan, t.TempDir(), func(context.Context, string, []string, []string) (string, error) {
		return "outro erro", errors.New("exit status 1")
	}, nil)
	if result.Steps[0].Status != AddonFailed || result.Steps[0].Output != "outro erro" {
		t.Fatalf("erro normal = %+v", result.Steps)
	}
}

func TestInstallPrerequisiteFailureSkipsAllHarnessesOnlyForAddon(t *testing.T) {
	plan := AddonPlan{Steps: []AddonPlanStep{
		{Addon: "first", Harness: "claude", Argv: []string{"prepare"}, Available: true, Prerequisite: true},
		{Addon: "first", Harness: "claude", Argv: []string{"register-claude"}, Available: true, AlreadyOK: "already exists"},
		{Addon: "first", Harness: "codex", Argv: []string{"prepare-codex"}, Available: true, Prerequisite: true},
		{Addon: "first", Harness: "codex", Argv: []string{"register-codex"}, Available: true},
		{Addon: "second", Harness: "claude", Argv: []string{"other-claude"}, Available: true},
		{Addon: "second", Harness: "codex", Argv: []string{"other-codex"}, Available: true},
	}}
	var calls []string
	result := InstallAddons(plan, t.TempDir(), func(_ context.Context, _ string, argv, _ []string) (string, error) {
		calls = append(calls, argv[0])
		if argv[0] == "prepare" {
			return "falha ao instalar", errors.New("exit status 1")
		}
		return "adicionado", nil
	}, nil)
	want := []AddonStatus{AddonFailed, AddonSkipped, AddonSkipped, AddonSkipped, AddonOK, AddonOK}
	if len(result.Steps) != len(want) || !slices.Equal(calls, []string{"prepare", "other-claude", "other-codex"}) {
		t.Fatalf("prerequisite = %+v; calls = %v", result.Steps, calls)
	}
	for i, step := range result.Steps {
		if step.Status != want[i] {
			t.Fatalf("passo %d = %+v; esperado %v", i, step, want[i])
		}
	}
}
