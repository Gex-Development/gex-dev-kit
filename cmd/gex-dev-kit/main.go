package main

import (
	"fmt"
	"io"
	"os"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/terminal"
)

var version = "dev"

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
	command := ""
	rest := args
	if len(args) > 0 {
		command, rest = args[0], args[1:]
	}
	switch command {
	case "", "--help", "-h", "help":
		fmt.Fprint(out, usage)
		return 0
	case "--version", "-v", "version":
		fmt.Fprintln(out, version)
		return 0
	case "hook":
		return runHook(rest, in, out, errOut)
	}
	opts, _ := parseOptions(command, rest)
	s := session{in: in, out: out, errOut: errOut, env: terminal.Detect(in, out, environ), opts: opts, deps: d, guard: newGuard(nil)}
	switch command {
	case "skills":
		return cmdSkills(s)
	case "hooks":
		return cmdHooks(s)
	case "addons":
		return cmdAddons(s)
	case "doctor":
		return cmdDoctor(s)
	case "ds":
		return cmdShowcase(s)
	case "install":
		return cmdInstall(s)
	case "rollback":
		return cmdRollback(s)
	case "rollback-agents":
		return cmdRollbackAgents(s)
	case "githooks":
		return cmdGitHooks(s)
	}
	return invalid(errOut, "comando desconhecido: "+command)
}

func invalid(errOut io.Writer, message string) int {
	fmt.Fprintln(errOut, message)
	fmt.Fprintln(errOut, "Use --help para opções.")
	return 2
}
