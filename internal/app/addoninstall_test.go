package app

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"
)

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
		"uv tool install specify-cli --from git+https://github.com/github/spec-kit.git@v1.2.2",
		"claude plugin marketplace add JuliusBrussee/caveman",
		"claude plugin install caveman@caveman",
		"npx -y github:JuliusBrussee/caveman -- --only codex",
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
	result := InstallAddons(plan, os.Getenv("HOME"), func(_ context.Context, argv, _ []string) (string, error) {
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
	result := InstallAddons(plan, os.Getenv("HOME"), func(context.Context, []string, []string) (string, error) { t.Fatal("runner invoked"); return "", nil }, nil)
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
	result := InstallAddons(plan, home, func(_ context.Context, _ []string, env []string) (string, error) {
		for _, name := range []string{"HOME", "USERPROFILE", "XDG_CONFIG_HOME", "XDG_DATA_HOME"} {
			if !slices.Contains(env, name+"="+home) {
				t.Errorf("%s not isolated", name)
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

func TestInstallAddonsKeepsEnvironmentForRealHome(t *testing.T) {
	plan := AddonPlan{Steps: []AddonPlanStep{{Addon: "test", Harness: "claude", Argv: []string{"test"}, Available: true}}}
	home := os.Getenv("HOME")
	InstallAddons(plan, home, func(_ context.Context, _ []string, env []string) (string, error) {
		if !reflect.DeepEqual(env, os.Environ()) {
			t.Fatal("environment changed for real HOME")
		}
		return "", nil
	}, nil)
}

func TestInstallAddonsRunnerSeesTimeout(t *testing.T) {
	plan := AddonPlan{Steps: []AddonPlanStep{{Addon: "test", Harness: "claude", Argv: []string{"test"}, Available: true}}}
	result := installAddonsWithTimeout(plan, os.Getenv("HOME"), func(ctx context.Context, _, _ []string) (string, error) {
		<-ctx.Done()
		return "", ctx.Err()
	}, nil, time.Millisecond)
	if result.Steps[0].Status != AddonFailed || !strings.Contains(result.Steps[0].Output, "deadline exceeded") {
		t.Fatalf("timeout = %+v", result.Steps)
	}
}

func TestExecAddonRunnerDoesNotInvokeShell(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("echo não é um executável no Windows")
	}
	output, err := ExecAddonRunner(context.Background(), []string{"echo", "ok", "&&", "echo", "bad"}, os.Environ())
	if err != nil || strings.TrimSpace(output) != "ok && echo bad" {
		t.Fatalf("argv deve chegar literal, sem shell: %q, %v", output, err)
	}
}

func TestInstallAddonsLimitsLargeOutputToTwoKilobytes(t *testing.T) {
	plan := AddonPlan{Steps: []AddonPlanStep{{Addon: "test", Harness: "claude", Argv: []string{"test"}, Available: true}}}
	result := InstallAddons(plan, os.Getenv("HOME"), func(context.Context, []string, []string) (string, error) {
		return strings.Repeat("a", 3000) + "token=private", nil
	}, nil)
	output := result.Steps[0].Output
	if len(output) > 2048 || strings.Contains(output, "private") || !strings.HasSuffix(output, "token=[oculto]") {
		t.Fatalf("unsafe output length %d, suffix %q", len(output), output[len(output)-40:])
	}
}
