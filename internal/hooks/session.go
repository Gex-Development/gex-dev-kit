package hooks

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"
)

const sessionTimeout = 1800 * time.Millisecond

const sessionMaxText = 1200

const sessionCoreBlockStart = "<!-- início gex-dev-kit core -->"

var sessionCoreSkills = []string{
	"gex-abrir-issue", "gex-auto-revisao", "gex-fluxo-desenvolvimento",
	"gex-preparar-pr", "gex-regras-de-negocio", "gex-revisar-pr", "gex-verificar-mudanca",
}

var sessionLocalNames = []string{"AGENTS.md", "CLAUDE.md", ".claude"}

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

type sessionProbe struct {
	gex      bool
	known    bool
	toplevel string
}

func GatherSession(in AgentInput, agent, home string, inGEX func(dir string) (bool, error), git GitFunc) SessionFacts {
	facts := SessionFacts{Agent: agent, Source: in.Source, CoreTotal: len(sessionCoreSkills)}
	done := make(chan sessionProbe, 1)
	go func() { done <- sessionProbeRepo(in.Cwd, inGEX, git) }()
	select {
	case p := <-done:
		facts.GEX, facts.GEXKnown = p.gex, p.known
		if p.toplevel != "" {
			facts.RepoName = filepath.Base(p.toplevel)
			facts.LocalFiles = sessionLocalFiles(p.toplevel)
		}
	case <-time.After(sessionTimeout):
	}
	facts.CoreSkills = sessionCountCore(home, agent)
	facts.CoreBlock = sessionHasCoreBlock(home, agent)
	return facts
}

func sessionProbeRepo(dir string, inGEX func(dir string) (bool, error), git GitFunc) sessionProbe {
	var p sessionProbe
	if inGEX != nil {
		if ok, err := inGEX(dir); err == nil {
			p.gex, p.known = ok, true
		}
	}
	if git != nil {
		if out, err := git(dir, "rev-parse", "--show-toplevel"); err == nil {
			p.toplevel = strings.TrimSpace(out)
		}
	}
	return p
}

func sessionLocalFiles(root string) []string {
	var found []string
	for _, name := range sessionLocalNames {
		if _, err := os.Stat(filepath.Join(root, name)); err == nil {
			found = append(found, name)
		}
	}
	return found
}

func sessionAgentDirs(home, agent string) (skills, core string) {
	if agent == "codex" {
		return filepath.Join(home, ".agents", "skills"), filepath.Join(home, ".codex", "AGENTS.md")
	}
	return filepath.Join(home, ".claude", "skills"), filepath.Join(home, ".claude", "CLAUDE.md")
}

func sessionCountCore(home, agent string) int {
	skills, _ := sessionAgentDirs(home, agent)
	count := 0
	for _, name := range sessionCoreSkills {
		if _, err := os.Stat(filepath.Join(skills, name, "SKILL.md")); err == nil {
			count++
		}
	}
	return count
}

func sessionHasCoreBlock(home, agent string) bool {
	_, core := sessionAgentDirs(home, agent)
	file, err := os.Open(core)
	if err != nil {
		return false
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, 1<<20))
	return err == nil && strings.Contains(string(data), sessionCoreBlockStart)
}

func SessionContext(f SessionFacts) string {
	lines := []string{
		fmt.Sprintf("GEX: início de sessão (agente %s, fonte %s).", sessionOrDefault(f.Agent, "desconhecido"), sessionOrDefault(f.Source, "desconhecida")),
		"Repositório: " + sessionRepoState(f) + ".",
		"Instruções locais: " + sessionLocalState(f.LocalFiles) + ".",
		fmt.Sprintf("Skills Core: %d de %d instaladas em %s.", f.CoreSkills, f.CoreTotal, sessionSkillsLabel(f.Agent)),
		"Bloco Core global: " + sessionBlockState(f) + ".",
	}
	text := strings.Join(lines, "\n")
	if utf8.RuneCountInString(text) > sessionMaxText {
		text = string([]rune(text)[:sessionMaxText])
	}
	return text
}

func sessionOrDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func sessionRepoState(f SessionFacts) string {
	switch {
	case !f.GEXKnown:
		return "desconhecido, verificação indisponível"
	case f.GEX:
		return "repositório GEX"
	default:
		return "fora do escopo GEX"
	}
}

func sessionLocalState(files []string) string {
	if len(files) == 0 {
		return "nenhuma"
	}
	return strings.Join(files, ", ")
}

func sessionSkillsLabel(agent string) string {
	if agent == "codex" {
		return "~/.agents/skills"
	}
	return "~/.claude/skills"
}

func sessionBlockState(f SessionFacts) string {
	if !f.CoreBlock {
		return "ausente"
	}
	if f.Agent == "codex" {
		return "presente em ~/.codex/AGENTS.md"
	}
	return "presente em ~/.claude/CLAUDE.md"
}
