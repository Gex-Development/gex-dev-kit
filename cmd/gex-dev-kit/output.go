package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/Gex-Development/gex-dev-kit/internal/app"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/components"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

func writePreview(w io.Writer, t theme.Theme, groups []components.ChangeGroup) {
	for _, group := range groups {
		fmt.Fprintln(w, outputTitle(t, group.Title))
		for _, change := range group.Changes {
			fmt.Fprintln(w, outputAction(t, change))
			for _, line := range change.Lines {
				fmt.Fprintln(w, outputDiffLine(t, line))
			}
			if change.Omitted > 0 {
				fmt.Fprintln(w, outputMuted(t, fmt.Sprintf("… mais %d %s", change.Omitted, outputPlural(change.Omitted, "linha", "linhas"))))
			}
			if change.Masked > 0 {
				fmt.Fprintln(w, outputMuted(t, fmt.Sprintf("%d %s", change.Masked, outputPlural(change.Masked, "valor oculto", "valores ocultos"))))
			}
		}
		fmt.Fprintln(w)
	}
}

func writeProgress(errOut io.Writer, quiet bool) func(app.Step) {
	if quiet {
		return func(app.Step) {}
	}
	return func(step app.Step) {
		if line := outputProgressLine(step); line != "" {
			fmt.Fprintln(errOut, line)
		}
	}
}

func outputProgressLine(step app.Step) string {
	switch step {
	case app.StepAgents:
		return "[..] Atualizando agentes"
	case app.StepFiles:
		return "[..] Gravando Core"
	case app.StepDone:
		return "[ok] Core gravado"
	}
	return ""
}

func outputTitle(t theme.Theme, text string) string {
	if t.Plain {
		return text
	}
	return t.Title.Render(text)
}

func outputMuted(t theme.Theme, text string) string {
	if t.Plain {
		return text
	}
	return t.Secondary.Render(text)
}

func outputAction(t theme.Theme, c components.Change) string {
	text := outputLabel(c.Action) + " " + outputText(c.Path)
	if detail := outputText(c.Detail); detail != "" {
		text += " · " + detail
	}
	if t.Plain {
		return text
	}
	switch c.Action {
	case components.ChangeCreate:
		return t.Success.Render(text)
	case components.ChangeUpdate, components.ChangeRestore:
		return t.Info.Render(text)
	case components.ChangeRemove:
		return t.Warning.Render(text)
	case components.ChangeBlocked:
		return t.Error.Render(text)
	}
	return t.Secondary.Render(text)
}

func outputDiffLine(t theme.Theme, line components.DiffLine) string {
	text := outputText(line.Text)
	prefix := " "
	switch line.Kind {
	case components.DiffAdded:
		prefix = "+"
	case components.DiffRemoved:
		prefix = "-"
	case components.DiffHunk:
		prefix = ""
	}
	if t.Plain {
		if prefix == "" {
			return text
		}
		return prefix + " " + text
	}
	switch line.Kind {
	case components.DiffAdded:
		return t.Success.Render(prefix + " " + text)
	case components.DiffRemoved:
		return t.Error.Render(prefix + " " + text)
	case components.DiffHunk:
		return t.Info.Render(text)
	}
	return t.Secondary.Render(prefix + " " + text)
}

func outputLabel(action components.ChangeAction) string {
	switch action {
	case components.ChangeCreate:
		return "criar"
	case components.ChangeUpdate:
		return "atualizar"
	case components.ChangeRemove:
		return "remover"
	case components.ChangeRestore:
		return "restaurar"
	case components.ChangeBlocked:
		return "bloqueado"
	}
	return "sem mudança"
}

func outputText(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\r", ""), "\t", "    ")
}

func outputPlural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}
