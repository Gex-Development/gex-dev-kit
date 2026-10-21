package main

const usage = "gex-dev-kit [opções globais] [comando] [opções]\n\n" +
	"install, doctor, rollback, rollback-agents, skills, hooks, addons, githooks, ds, hook, version, help\n"

func commandHelp(command string) string {
	return usage
}
