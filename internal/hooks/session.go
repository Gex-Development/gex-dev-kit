package hooks

type SessionFacts struct {
	Agent      string
	Source     string
	GEX        bool
	GEXKnown   bool
	RepoName   string
	LocalFiles []string
	CoreSkills int
	CoreTotal  int
	CoreBlock  bool
}

func GatherSession(in AgentInput, agent, home string, inGEX func(dir string) (bool, error), git GitFunc) SessionFacts {
	return SessionFacts{Agent: agent}
}

func SessionContext(f SessionFacts) string {
	return ""
}
