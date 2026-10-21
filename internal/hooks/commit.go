package hooks

type CommitCheck int

const (
	CommitNone CommitCheck = iota
	CommitOK
	CommitBad
)

func ClassifyCommit(in AgentInput) (CommitCheck, []string) {
	return CommitNone, nil
}

func CleanGitMessage(message, commentChar string) string {
	return message
}
