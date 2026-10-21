package terminal

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/term"
)

type Mode int

const (
	ModeAuto Mode = iota
	ModeTUI
	ModeCLI
)

func ParseMode(s string) (Mode, error) {
	switch s {
	case "", "auto":
		return ModeAuto, nil
	case "tui":
		return ModeTUI, nil
	case "cli":
		return ModeCLI, nil
	}
	return ModeAuto, fmt.Errorf("modo de interface inválido %q; use auto, tui ou cli", s)
}

type Env struct {
	StdinTTY   bool
	StdoutTTY  bool
	Term       string
	NoColor    bool
	Accessible bool
	CI         bool
	NoAnim     bool
	Profile    colorprofile.Profile
	Width      int
	Height     int
	Environ    []string
	InFile     *os.File
	OutFile    *os.File
}

func Detect(in io.Reader, out io.Writer, environ []string) Env {
	inFile, _ := in.(*os.File)
	outFile, _ := out.(*os.File)
	env := Env{
		Term:       terminalEnvValue(environ, "TERM"),
		NoColor:    terminalEnvSet(environ, "NO_COLOR"),
		Accessible: terminalEnvSet(environ, "ACCESSIBLE"),
		CI:         terminalEnvSet(environ, "CI"),
		NoAnim:     terminalEnvSet(environ, "GEX_NO_ANIM"),
		Environ:    environ,
		InFile:     inFile,
		OutFile:    outFile,
	}
	if inFile != nil {
		env.StdinTTY = term.IsTerminal(inFile.Fd())
	}
	if outFile == nil || !term.IsTerminal(outFile.Fd()) {
		env.Profile = colorprofile.NoTTY
		return env
	}
	env.StdoutTTY = true
	env.Profile = vtProfile(colorprofile.Detect(outFile, environ), terminalEnableVT(outFile))
	if width, height, err := term.GetSize(outFile.Fd()); err == nil {
		env.Width, env.Height = width, height
	}
	return env
}

func (e Env) Interactive() bool {
	return e.StdinTTY && e.StdoutTTY
}

func (e Env) Resolve(requested Mode) (Mode, error) {
	switch requested {
	case ModeCLI:
		return ModeCLI, nil
	case ModeTUI:
		if !e.Interactive() {
			return ModeAuto, errors.New("TUI exige terminal interativo (entrada e saída em TTY); use --ui=cli")
		}
		if e.Term == "dumb" {
			return ModeAuto, errors.New("TUI não roda com TERM=dumb; use --ui=cli")
		}
		return ModeTUI, nil
	}
	if e.Interactive() && e.Term != "dumb" && !e.Accessible && !e.CI && e.Width > 0 && e.Height > 0 {
		return ModeTUI, nil
	}
	return ModeCLI, nil
}

func (e Env) Motion(flagNoAnim bool) bool {
	return !flagNoAnim && !e.NoAnim && !e.Accessible && !e.CI && e.StdoutTTY && e.Term != "dumb"
}

func (e Env) ColorProfile() colorprofile.Profile {
	if !e.StdoutTTY || e.Term == "dumb" {
		return colorprofile.NoTTY
	}
	if e.NoColor && e.Profile > colorprofile.ASCII {
		return colorprofile.ASCII
	}
	return e.Profile
}

func (e Env) Writer(w io.Writer) io.Writer {
	return &colorprofile.Writer{Forward: w, Profile: e.ColorProfile()}
}

func vtProfile(base colorprofile.Profile, vtOK bool) colorprofile.Profile {
	if vtOK || base <= colorprofile.ASCII {
		return base
	}
	return colorprofile.ASCII
}

func terminalEnvLookup(environ []string, key string) (string, bool) {
	value, found := "", false
	for _, entry := range environ {
		name, v, ok := strings.Cut(entry, "=")
		if ok && name == key {
			value, found = v, true
		}
	}
	return value, found
}

func terminalEnvValue(environ []string, key string) string {
	value, _ := terminalEnvLookup(environ, key)
	return value
}

func terminalEnvSet(environ []string, key string) bool {
	return terminalEnvValue(environ, key) != ""
}
