package hooks

const preprChecklist = "checklist: gex-auto-revisao, testes, build, lint, issue, assignee, reviewers"

func ClassifyPrePR(cmds []Command) Decision {
	for _, args := range sensitiveCandidates(cmds) {
		if sensitiveHasPrefix(args, "gh", "pr", "create") {
			return sensitiveAsk("abertura de PR", preprChecklist)
		}
	}
	return Decision{}
}

func CheckPrePR(in AgentInput, scope func(dirs []string) bool) Decision {
	return preprCheck(in, scope, ParseShell)
}

func preprCheck(in AgentInput, scope func(dirs []string) bool, parse func(string) (Script, error)) Decision {
	if in.Truncated || !in.Parsed || (in.Tool != "Bash" && in.Tool != "PowerShell") {
		return Decision{}
	}
	script, err := parse(in.Command)
	if err != nil {
		return Decision{}
	}
	cmds := sensitiveExpand(script)
	return sensitiveInScope(scope, sensitiveDirs(in.Cwd, cmds), ClassifyPrePR(cmds))
}
