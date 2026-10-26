package catalog

import "strings"

func removeMCP(name string) map[string]AddonInstall {
	return map[string]AddonInstall{
		"claude": {Steps: []AddonStep{{Argv: []string{"claude", "mcp", "remove", "--scope", "user", name}, AlreadyOK: "No MCP server named"}}},
		"codex":  {Steps: []AddonStep{{Argv: []string{"codex", "mcp", "remove", name}}}},
	}
}

func removeClaudePlugin(plugin, marketplace string) AddonInstall {
	steps := []AddonStep{{Argv: []string{"claude", "plugin", "uninstall", plugin}, AlreadyOK: "not found in installed plugins"}}
	if marketplace != "" {
		steps = append(steps, AddonStep{Argv: []string{"claude", "plugin", "marketplace", "remove", marketplace}, AlreadyOK: "not found"})
	}
	return AddonInstall{Steps: steps}
}

func addonRemovals() map[string]map[string]AddonInstall {
	specKit := AddonInstall{
		Steps:  []AddonStep{{Argv: []string{"uv", "tool", "uninstall", "specify-cli"}, AlreadyOK: "is not installed"}},
		Manual: "O kit não apaga o .specify/ nem os comandos speckit-* dos projetos onde você rodou specify init.",
	}
	return map[string]map[string]AddonInstall{
		"caveman": {
			"claude": removeClaudePlugin("caveman@caveman", "caveman"),
			"codex":  {Manual: "O Codex recebeu o caveman por npx e não há comando oficial de remoção conferido: abra /hooks e desative os hooks do caveman."},
		},
		"ponytail": {
			"claude": removeClaudePlugin("ponytail@ponytail", "ponytail"),
			"codex": {Steps: []AddonStep{
				{Argv: []string{"codex", "plugin", "remove", "ponytail@ponytail"}},
				{Argv: []string{"codex", "plugin", "marketplace", "remove", "ponytail"}, AlreadyOK: "not configured or installed"},
			}},
		},
		"superpowers": {
			"claude": removeClaudePlugin("superpowers@claude-plugins-official", ""),
			"codex":  {Manual: "No Codex, rode codex plugin list para ver o identificador do superpowers e depois codex plugin remove IDENTIFICADOR."},
		},
		"spec-kit":      {"claude": specKit, "codex": specKit},
		"context7":      removeMCP("context7"),
		"terraform":     removeMCP("terraform"),
		"brave-search":  removeMCP("brave-search"),
		"mapbox-devkit": removeMCP("mapbox-devkit"),
		"claude-code-templates": {
			"claude": {Manual: "O kit não sabe onde você gravou os hooks: apague as entradas deles no settings escolhido (user, project ou local) e os scripts de .claude/hooks/ (tdd-gate.sh, scope-guard.sh, subagent-read-only-guard.py, change-logger.py)."},
			"codex":  {Manual: "Nada a remover: estes hooks são do Claude Code."},
		},
	}
}

func AddonRemoval(name, harness string) AddonInstall {
	return addonRemovals()[name][harness]
}

func (a Addon) DisplayRemovalCommands(harness string) []string {
	var commands []string
	for _, step := range AddonRemoval(a.Name, harness).Steps {
		commands = append(commands, strings.Join(step.Argv, " "))
	}
	return commands
}
