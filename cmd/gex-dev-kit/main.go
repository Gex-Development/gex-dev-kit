package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"slices"
	"syscall"

	"github.com/Gex-Development/gex-dev-kit/internal/agents"
	"github.com/Gex-Development/gex-dev-kit/internal/app"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/screens"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/terminal"
)

var version = "v0.3.4"

type session struct {
	in     io.Reader
	out    io.Writer
	errOut io.Writer
	env    terminal.Env
	opts   options
	deps   deps
	guard  *guard
}

func main() { os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)) }

func run(args []string, in io.Reader, out, errOut io.Writer) int {
	return runWith(args, in, out, errOut, os.Environ(), realDeps())
}

func runWith(args []string, in io.Reader, out, errOut io.Writer, environ []string, d deps) int {
	command, rest := splitCommand(args)
	opts, err := parseOptions(command, rest)
	if err != nil {
		return usageFail(errOut, err)
	}
	if opts.Help {
		fmt.Fprint(out, commandHelp(command))
		return 0
	}
	if opts.Version || command == "version" {
		fmt.Fprintln(out, version)
		return 0
	}
	if command == "help" {
		target := ""
		if len(opts.Args) > 0 {
			target = opts.Args[0]
		}
		fmt.Fprint(out, commandHelp(target))
		return 0
	}
	if command == "hook" {
		return runHook(opts.Args, in, out, errOut)
	}
	env := terminal.Detect(in, out, environ)
	requested := opts.UI
	if requested == terminal.ModeAuto && (opts.Yes || opts.Quiet || opts.JSON) {
		requested = terminal.ModeCLI
	}
	mode, err := env.Resolve(requested)
	if err != nil {
		return usageFail(errOut, usageError{err.Error(), "Rode em um terminal interativo ou use --ui=cli."})
	}
	opts.UI = mode
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(signals)
	g := newGuard(signals)
	defer g.Stop()
	g.errOut = errOut
	s := session{in: in, out: out, errOut: errOut, env: env, opts: opts, deps: d, guard: g}
	return executa(s, command)
}

func executa(s session, command string) int {
	done := make(chan int, 1)
	go func() { done <- comandoFn(command)(s) }()
	var interrompido <-chan struct{}
	if s.opts.UI != terminal.ModeTUI {
		interrompido = s.guard.Interrupted()
	}
	select {
	case code := <-done:
		if sinal, ok := s.guard.Code(); ok {
			return sinal
		}
		return code
	case <-interrompido:
		sinal, _ := s.guard.Code()
		return sinal
	}
}

func comandoFn(command string) func(session) int {
	switch command {
	case "":
		return runHome
	case "skills":
		return cmdSkills
	case "hooks":
		return cmdHooks
	case "addons":
		return cmdAddons
	case "doctor":
		return cmdDoctor
	case "ds":
		return cmdShowcase
	case "install":
		return cmdInstall
	case "rollback":
		return cmdRollback
	case "rollback-agents":
		return cmdRollbackAgents
	case "githooks":
		return cmdGitHooks
	}
	return func(s session) int { return invalid(s.errOut, "comando desconhecido: "+command) }
}

func runHome(s session) int {
	if s.opts.UI == terminal.ModeTUI {
		return runHomeTUI(s)
	}
	if s.env.Interactive() && (s.env.Accessible || s.env.Term == "dumb") {
		escolha, err := showAccessibleHome(s.in, s.out)
		if err != nil {
			return telaFalhou(s.errOut, err)
		}
		return acaoDaHome(s, escolha)
	}
	fmt.Fprint(s.out, usage)
	return 0
}

func runHomeTUI(s session) int {
	r, stop := writeRunner(s)
	r.Profile = s.env.ColorProfile()
	defer stop()
	home := homeDoUsuario(s)
	preset, err := homeRequest()
	if err != nil {
		return telaFalhou(s.errOut, err)
	}
	res, err := screens.RunChat(r, tuiCommon(s, home), chatDeps(s, home, preset), screens.IntentHome)
	if err != nil {
		return telaFalhou(s.errOut, err)
	}
	return res.ExitCode
}

func homeRequest() (app.Request, error) {
	executable, err := writeExecutable()
	if err != nil {
		return app.Request{}, err
	}
	return app.Request{Executable: executable}, nil
}

func chatDeps(s session, home string, preset app.Request) screens.ChatDeps {
	dir, _ := os.Getwd()
	return screens.ChatDeps{
		Dir:    dir,
		Core:   func() (app.CoreStatus, error) { return app.Core(home) },
		Agents: func() ([]app.AgentChange, error) { return agents.Preview(home) },
		Git: func() (app.GitInfo, bool) {
			wd, err := os.Getwd()
			if err != nil {
				return app.GitInfo{}, false
			}
			return app.GitStatus(wd)
		},
		Doctor: func() (app.Report, error) { return s.deps.Doctor(home, s.env.Environ, readExecutable()) },
		Install: screens.InstallDeps{
			Options: s.deps.InstallOptions,
			Plan:    func(req app.Request) (app.Plan, error) { return s.deps.PlanInstall(req, app.DiffLinesTUI) },
			Apply:   s.deps.ApplyInstall,
		},
		Addons: screens.AddonDeps{
			Plan: func(names, harnesses []string, secrets map[string]string) (app.AddonPlan, error) {
				return app.PlanAddonsWithSecrets(names, harnesses, exec.LookPath, secrets)
			},
			Secrets: app.AddonSecretsNeeded,
			Install: func(plan app.AddonPlan, home string, progress func(app.AddonStepResult)) app.AddonResult {
				return app.InstallAddons(plan, home, app.ExecAddonRunner, progress)
			},
		},
		Rollback: screens.RollbackDeps{
			Plan:  func() (app.RollbackPreview, error) { return s.deps.PlanRollback(home, app.DiffLinesTUI) },
			Apply: func() (string, error) { return s.deps.Rollback(home) },
		},
		RollbackAgents: screens.RollbackDeps{
			Plan:  func() (app.RollbackPreview, error) { return app.RollbackPreview{}, nil },
			Apply: func() (string, error) { return "", s.deps.RollbackAgents(home) },
		},
		GitHooks: screens.GitHooksDeps{
			Plan: func(req app.GitHooksRequest) (app.GitHooksPlan, error) {
				return s.deps.PlanGitHooks(req, app.DiffLinesTUI)
			},
			Apply: s.deps.ApplyGitHooks,
		},
		GitHooksPreset: app.GitHooksRequest{Home: home, Executable: readExecutable()},
		Preset:         preset,
	}
}

func acaoDaHome(s session, escolha string) int {
	switch escolha {
	case "setup":
		return cmdInstall(s)
	case "install":
		return cmdInstall(s)
	case "catalog":
		return cmdSkills(s)
	case "doctor":
		return cmdDoctor(s)
	case "rollback":
		return cmdRollback(s)
	case "rollback-agents":
		return cmdRollbackAgents(s)
	case "addons":
		return cmdAddons(s)
	case "githooks":
		return cmdGitHooks(s)
	case "ds":
		return cmdShowcase(s)
	case "help":
		fmt.Fprint(s.out, usage)
	}
	return 0
}

func tuiCommon(s session, home string) screens.Common {
	return screens.Common{Version: version, Home: home, UserHome: userHome(), Dark: true, Profile: s.env.ColorProfile(), Motion: s.env.Motion(s.opts.NoAnim), Powerline: !slices.Contains(s.env.Environ, "GEX_NO_POWERLINE=1")}
}

func homeDoUsuario(s session) string {
	if s.opts.Home != "" {
		return s.opts.Home
	}
	return userHome()
}

func userHome() string {
	home, _ := os.UserHomeDir()
	return home
}

func telaFalhou(errOut io.Writer, err error) int {
	if err == nil {
		return 0
	}
	if code, ok := screens.ExitCode(err); ok {
		return code
	}
	fmt.Fprintln(errOut, "Falha na tela: "+err.Error())
	return 1
}

func usageFail(errOut io.Writer, err error) int {
	ue, ok := err.(usageError)
	if !ok {
		ue = usageError{err.Error(), acaoPadrao}
	}
	fmt.Fprintln(errOut, ue.cause)
	fmt.Fprintln(errOut, ue.action)
	return 2
}

func invalid(errOut io.Writer, message string) int {
	fmt.Fprintln(errOut, message)
	fmt.Fprintln(errOut, "Use --help para opções.")
	return 2
}
