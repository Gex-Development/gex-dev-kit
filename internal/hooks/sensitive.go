package hooks

func ClassifySensitive(cmds []Command, tool, filePath string, prePRInstalled bool) Decision {
	return Decision{}
}

func CheckSensitive(in AgentInput, scope func(dirs []string) bool, prePRInstalled func() bool) Decision {
	return Decision{}
}
