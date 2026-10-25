package app

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/Gex-Development/gex-dev-kit/internal/catalog"
)

type AddonPlanStep struct {
	Addon, Harness, Display string
	Argv                    []string
	Available               bool
}

type AddonManual struct {
	Addon, Harness, Text string
}

type AddonPlan struct {
	Steps   []AddonPlanStep
	Manual  []AddonManual
	Missing []string
}

func PlanAddons(names, harnesses []string, lookPath func(string) (string, error)) (AddonPlan, error) {
	var plan AddonPlan
	seenSteps := make(map[string]bool)
	seenManual := make(map[string]bool)
	checked := make(map[string]bool)
	for _, name := range names {
		addon, ok := catalog.AddonByName(name)
		if !ok {
			return AddonPlan{}, fmt.Errorf("addon desconhecido: %s", name)
		}
		for _, harness := range harnesses {
			if harness != "claude" && harness != "codex" {
				continue
			}
			install := addon.Install[harness]
			for _, step := range install.Steps {
				key := strings.Join(step.Argv, "\x00")
				if seenSteps[key] {
					continue
				}
				seenSteps[key] = true
				binary := step.Argv[0]
				if !checked[binary] {
					checked[binary] = true
					if _, err := lookPath(binary); err != nil {
						plan.Missing = append(plan.Missing, binary)
					}
				}
				plan.Steps = append(plan.Steps, AddonPlanStep{Addon: name, Harness: harness, Display: strings.Join(step.Argv, " "), Argv: slices.Clone(step.Argv), Available: !slices.Contains(plan.Missing, binary)})
			}
			if install.Manual != "" {
				key := name + "\x00" + harness
				if !seenManual[key] {
					plan.Manual = append(plan.Manual, AddonManual{Addon: name, Harness: harness, Text: install.Manual})
					seenManual[key] = true
				}
			}
		}
	}
	return plan, nil
}

type AddonStatus uint8

const (
	AddonPending AddonStatus = iota
	AddonOK
	AddonFailed
	AddonSkipped
)

type AddonStepResult struct {
	Step   AddonPlanStep
	Status AddonStatus
	Output string
}

type AddonResult struct {
	Steps  []AddonStepResult
	Manual []AddonManual
}

type AddonRunner func(ctx context.Context, argv []string, env []string) (string, error)

func InstallAddons(plan AddonPlan, home string, run AddonRunner, progress func(AddonStepResult)) AddonResult {
	return installAddonsWithTimeout(plan, home, run, progress, 5*time.Minute)
}

func installAddonsWithTimeout(plan AddonPlan, home string, run AddonRunner, progress func(AddonStepResult), timeout time.Duration) AddonResult {
	result := AddonResult{Manual: slices.Clone(plan.Manual)}
	env := os.Environ()
	if home != os.Getenv("HOME") {
		overrides := []string{"HOME=", "USERPROFILE=", "XDG_CONFIG_HOME=", "XDG_DATA_HOME="}
		filtered := make([]string, 0, len(env)+len(overrides))
		for _, entry := range env {
			if !slices.ContainsFunc(overrides, func(prefix string) bool { return strings.HasPrefix(entry, prefix) }) {
				filtered = append(filtered, entry)
			}
		}
		for _, prefix := range overrides {
			filtered = append(filtered, prefix+home)
		}
		env = filtered
	}
	failed := make(map[string]bool)
	for _, step := range plan.Steps {
		item := AddonStepResult{Step: step, Status: AddonPending}
		if progress != nil {
			progress(item)
		}
		key := step.Addon + "\x00" + step.Harness
		if !step.Available || failed[key] {
			item.Status = AddonSkipped
		} else {
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			output, err := run(ctx, step.Argv, env)
			cancel()
			if err != nil {
				failed[key] = true
				item.Status = AddonFailed
				if output == "" {
					output = err.Error()
				}
			} else {
				item.Status = AddonOK
			}
			item.Output = addonOutput(output)
		}
		result.Steps = append(result.Steps, item)
		if progress != nil {
			progress(item)
		}
	}
	return result
}

var (
	addonANSI   = regexp.MustCompile(`\x1b(?:\[[0-?]*[ -/]*[@-~]|\][^\x07\x1b]*(?:\x07|\x1b\\))`)
	addonSecret = regexp.MustCompile(`(?i)(bearer\s+|token\s*=\s*|key\s*=\s*)[^\s]+`)
)

func addonOutput(raw string) string {
	raw = addonSecret.ReplaceAllString(strings.ReplaceAll(addonANSI.ReplaceAllString(raw, ""), "\x1b", ""), "${1}[oculto]")
	lines := strings.Split(raw, "\n")
	if len(lines) > 20 {
		lines = lines[len(lines)-20:]
	}
	raw = strings.Join(lines, "\n")
	if len(raw) > 2048 {
		raw = strings.ToValidUTF8(raw[len(raw)-2048:], "")
	}
	return raw
}

type addonTail struct{ bytes []byte }

func (t *addonTail) Write(p []byte) (int, error) {
	n := len(p)
	if len(p) >= 4096 {
		t.bytes = append(t.bytes[:0], p[len(p)-4096:]...)
		return n, nil
	}
	t.bytes = append(t.bytes, p...)
	if len(t.bytes) > 4096 {
		t.bytes = append(t.bytes[:0], t.bytes[len(t.bytes)-4096:]...)
	}
	return n, nil
}

func ExecAddonRunner(ctx context.Context, argv []string, env []string) (string, error) {
	if len(argv) == 0 {
		return "", fmt.Errorf("comando de addon vazio")
	}
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Env = env
	cmd.WaitDelay = 2 * time.Second
	var tail addonTail
	cmd.Stdout, cmd.Stderr = &tail, &tail
	err := cmd.Run()
	return string(tail.bytes), err
}
