package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/Gex-Development/gex-dev-kit/internal/agents"
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

type AddonRunner func(ctx context.Context, dir string, argv, env []string) (string, error)

func InstallAddons(plan AddonPlan, home string, run AddonRunner, progress func(AddonStepResult)) AddonResult {
	return installAddonsWithTimeout(plan, home, run, progress, 5*time.Minute)
}

func installAddonsWithTimeout(plan AddonPlan, home string, run AddonRunner, progress func(AddonStepResult), timeout time.Duration) AddonResult {
	result := AddonResult{Manual: slices.Clone(plan.Manual)}
	realHome, err := os.UserHomeDir()
	if err != nil {
		realHome = ""
	}
	env := addonEnvironment(os.Environ(), home, realHome)
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
			output, err := run(ctx, home, step.Argv, env)
			expired := errors.Is(ctx.Err(), context.DeadlineExceeded)
			cancel()
			if expired {
				output = fmt.Sprintf("tempo esgotado (%s)", timeout)
				err = context.DeadlineExceeded
			}
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

func ExecAddonRunner(ctx context.Context, dir string, argv, env []string) (string, error) {
	if len(argv) == 0 {
		return "", fmt.Errorf("comando de addon vazio")
	}
	var tail addonTail
	err := agents.RunDetached(ctx, dir, argv, env, &tail)
	return string(tail.bytes), err
}

func addonEnvironment(environ []string, home, realHome string) []string {
	allow := map[string]bool{}
	for _, key := range []string{"PATH", "PATHEXT", "SYSTEMROOT", "WINDIR", "COMSPEC", "PROCESSOR_ARCHITECTURE", "PROCESSOR_ARCHITEW6432", "LANG", "LC_ALL", "LC_CTYPE", "TERM", "TMPDIR", "TEMP", "TMP", "SSL_CERT_FILE", "SSL_CERT_DIR", "NODE_EXTRA_CA_CERTS", "HTTP_PROXY", "HTTPS_PROXY", "ALL_PROXY", "NO_PROXY"} {
		allow[key] = true
	}
	real := realHome != "" && filepath.Clean(home) == filepath.Clean(realHome)
	if real {
		for _, key := range []string{"HOME", "USERPROFILE", "HOMEDRIVE", "HOMEPATH", "APPDATA", "LOCALAPPDATA", "XDG_CONFIG_HOME", "XDG_DATA_HOME", "XDG_CACHE_HOME", "XDG_STATE_HOME", "XDG_BIN_HOME"} {
			allow[key] = true
		}
	}
	env := make([]string, 0, len(environ))
	index := make(map[string]int)
	for _, entry := range environ {
		key, _, ok := strings.Cut(entry, "=")
		upper := strings.ToUpper(key)
		if !ok || !allow[upper] {
			continue
		}
		if n, exists := index[upper]; exists {
			env[n] = entry
		} else {
			index[upper] = len(env)
			env = append(env, entry)
		}
	}
	set := func(key, value string) {
		if n, ok := index[key]; ok {
			env[n] = key + "=" + value
		} else {
			index[key] = len(env)
			env = append(env, key+"="+value)
		}
	}
	if !real {
		set("HOME", home)
		set("USERPROFILE", home)
		set("APPDATA", filepath.Join(home, "AppData", "Roaming"))
		set("LOCALAPPDATA", filepath.Join(home, "AppData", "Local"))
		set("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
		set("XDG_DATA_HOME", filepath.Join(home, ".local", "share"))
		set("XDG_CACHE_HOME", filepath.Join(home, ".cache"))
		set("XDG_STATE_HOME", filepath.Join(home, ".local", "state"))
		set("XDG_BIN_HOME", filepath.Join(home, ".local", "bin"))
		if runtime.GOOS == "windows" {
			volume := filepath.VolumeName(home)
			set("HOMEDRIVE", volume)
			set("HOMEPATH", strings.TrimPrefix(home, volume))
		}
	}
	set("CLAUDE_CONFIG_DIR", filepath.Join(home, ".claude"))
	set("CODEX_HOME", filepath.Join(home, ".codex"))
	set("GIT_TERMINAL_PROMPT", "0")
	return env
}
