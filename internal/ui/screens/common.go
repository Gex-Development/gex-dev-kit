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

func homeCandidates(userHome string) []string {
	candidates := []string{filepath.Clean(userHome)}
	if resolved, err := filepath.EvalSymlinks(userHome); err == nil && resolved != candidates[0] {
		candidates = append(candidates, resolved)
	}
	return candidates
}

func homeLabel(path, userHome string) string {
	if path == "" || userHome == "" {
		return path
	}
	cleaned := filepath.Clean(path)
	for _, home := range homeCandidates(userHome) {
		if cleaned == home {
			return "~"
		}
		if strings.HasPrefix(cleaned, home+string(filepath.Separator)) {
			return "~" + filepath.ToSlash(cleaned[len(home):])
		}
	}
	return path
}

func maskHome(text, userHome string) string {
	if userHome == "" {
		return text
	}
	for _, home := range homeCandidates(userHome) {
		if home == "." || home == string(filepath.Separator) {
			continue
		}
		text = maskHomeText(text, home)
	}
	return text
}

func maskHomeText(text, home string) string {
	var out strings.Builder
	rest := text
	for {
		i := strings.Index(rest, home)
		if i < 0 {
			out.WriteString(rest)
			return out.String()
		}
		end := i + len(home)
		out.WriteString(rest[:i])
		if end == len(rest) || strings.IndexByte(homeBoundaries, rest[end]) >= 0 {
			out.WriteString("~")
		} else {
			out.WriteString(home)
		}
		rest = rest[end:]
	}
}

const homeBoundaries = "/\\ \t\n'\"`:;,)]}"
