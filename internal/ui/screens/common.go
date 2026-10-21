package screens

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/charmbracelet/colorprofile"
)

type Common struct {
	Version  string
	Home     string
	UserHome string
	Dark     bool
	Profile  colorprofile.Profile
	Motion   bool
}

type Busy interface{ Busy() bool }

type InterruptDeferredMsg struct{ Signal os.Signal }

type SignalError struct{ Signal os.Signal }

func (e SignalError) Error() string {
	return "encerrado por sinal: " + e.Signal.String()
}

func ExitCode(err error) (int, bool) {
	signalErr, ok := err.(SignalError)
	if !ok {
		return 0, false
	}
	switch signalErr.Signal {
	case os.Interrupt:
		return 130, true
	case syscall.SIGTERM:
		return 143, true
	case syscall.SIGHUP:
		return 129, true
	}
	return 128, true
}

func homeLabel(path, userHome string) string {
	if path == "" || userHome == "" {
		return path
	}
	candidates := []string{filepath.Clean(userHome)}
	if resolved, err := filepath.EvalSymlinks(userHome); err == nil && resolved != candidates[0] {
		candidates = append(candidates, resolved)
	}
	cleaned := filepath.Clean(path)
	for _, home := range candidates {
		if cleaned == home {
			return "~"
		}
		if strings.HasPrefix(cleaned, home+string(filepath.Separator)) {
			return "~" + filepath.ToSlash(cleaned[len(home):])
		}
	}
	return path
}
