package main

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/agents"
	"github.com/Gex-Development/gex-dev-kit/internal/kit"
	"github.com/charmbracelet/x/ansi"
)

func previewText(files []kit.Change, versions []agents.Change) string {
	var text strings.Builder
	fmt.Fprintf(&text, "CORE · %d arquivos\n", len(files))
	for _, change := range files {
		fmt.Fprintf(&text, "  %s %s\n", change.Action, change.Path)
	}
	if len(versions) > 0 {
		text.WriteString("\nAGENTES · versões fixadas\n")
		for _, change := range versions {
			current := change.Current
			if current == "" {
				current = "não gerenciado"
			}
			switch {
			case change.Blocked:
				fmt.Fprintf(&text, "  %s bloqueado (%s → %s): %s\n  Ação manual: %s\n", change.Name, current, change.Target, safeAgentReason(change.Reason), change.Manual)
			case change.Current != change.Target:
				fmt.Fprintf(&text, "  %s %s → %s: %s\n", change.Name, current, change.Target, change.Command)
			default:
				fmt.Fprintf(&text, "  %s %s (sem mudança)\n", change.Name, current)
			}
		}
	}
	return strings.TrimSuffix(text.String(), "\n")
}

func previewPanel(files []kit.Change, versions []agents.Change, width int) string {
	width = max(8, min(width, 78))
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#1e6fd9")).Padding(0, 1).
		Render(ansi.Hardwrap(previewText(files, versions), width-4, false))
}
