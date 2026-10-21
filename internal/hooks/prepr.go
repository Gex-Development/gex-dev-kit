package hooks

func ClassifyPrePR(cmds []Command) Decision {
	return Decision{}
}

func CheckPrePR(in AgentInput, scope func(dirs []string) bool) Decision {
	return Decision{}
}
