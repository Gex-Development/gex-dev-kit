package screens

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/app"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/components"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
	"github.com/charmbracelet/colorprofile"
)

type Common struct {
	Version   string
	Home      string
	UserHome  string
	Dark      bool
	Profile   colorprofile.Profile
	NoIntro   bool
	Motion    bool
	Powerline bool
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

func homeAgentChange(changes []app.AgentChange, name string) (app.AgentChange, bool) {
	for _, change := range changes {
		if change.Name == name {
			return change, true
		}
	}
	return app.AgentChange{}, false
}

func homeAgentText(change app.AgentChange, userHome string) string {
	switch {
	case change.Blocked:
		return maskHome(change.Reason, userHome)
	case change.Current == "":
		return "não instalado pelo kit"
	case change.Current != change.Target:
		return maskHome(change.Current, userHome) + " → " + maskHome(change.Target, userHome)
	}
	return maskHome(change.Current, userHome)
}

func homeAgentState(change app.AgentChange) theme.State {
	if change.Blocked || change.Current == "" || change.Current != change.Target {
		return theme.StateWarn
	}
	return theme.StateOK
}

func homeCoreState(state string) theme.State {
	switch state {
	case "íntegro":
		return theme.StateOK
	case "não instalado":
		return theme.StateInfo
	}
	return theme.StateWarn
}

func homeSplit(text string) []string {
	return strings.Split(text, "\n")
}

func homeCenter(width int, block string) []string {
	return homeSplit(lipgloss.PlaceHorizontal(width, lipgloss.Center, block))
}

func installGuard(d InstallDeps) (InstallDeps, func()) {
	var wg sync.WaitGroup
	apply := d.Apply
	if apply != nil {
		d.Apply = func(p app.Plan, progress func(app.Step)) (app.Result, error) {
			wg.Add(1)
			defer wg.Done()
			return apply(p, progress)
		}
	}
	return d, wg.Wait
}

func installListen(ch chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return msg
	}
}

func installStepLine(s app.Step) string {
	switch s {
	case app.StepAgents:
		return "Etapa: agentes"
	case app.StepFiles:
		return "Etapa: arquivos do Core"
	}
	return "Etapa: concluída"
}

func installAction(action string) components.ChangeAction {
	switch action {
	case "criar":
		return components.ChangeCreate
	case "atualizar":
		return components.ChangeUpdate
	case "remover":
		return components.ChangeRemove
	case "restaurar":
		return components.ChangeRestore
	case "bloqueado":
		return components.ChangeBlocked
	}
	return components.ChangeUnchanged
}
