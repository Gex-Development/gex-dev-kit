package app

import (
	"github.com/Gex-Development/gex-dev-kit/internal/agents"
	"github.com/Gex-Development/gex-dev-kit/internal/githooks"
	"github.com/Gex-Development/gex-dev-kit/internal/kit"
)

type Status = kit.Status
type Change = kit.Change
type RollbackPlan = kit.RollbackPlan
type AgentChange = agents.Change

type Problem struct{ Title, Cause, Action string }

type Request struct {
	Home            string
	Executable      string
	Skills          []string
	Hooks           []string
	SkipAgentUpdate bool
}

type Notice struct {
	Level string
	Text  string
}

type DiffLine struct {
	Kind byte
	Text string
}

type FileDiff struct {
	Path    string
	Action  string
	Summary string
	Lines   []DiffLine
	Omitted int
	Masked  int
	Binary  bool
}

const (
	DiffLinesCLI = 40
	DiffLinesTUI = 400
)

type Plan struct {
	Request Request
	Files   []kit.Change
	Diffs   []FileDiff
	Agents  []agents.Change
	Updates []agents.Change
	Blocked bool
	Notices []Notice
	kit     kit.Plan
}

type Option struct {
	Value, Label, Hint string
}

type OptionGroup struct {
	ID, Title, Description string
	Options                []Option
	Validate               func([]string) error
}

type Step int

const (
	StepAgents Step = iota
	StepFiles
	StepDone
)

type Result struct {
	HistoryID      string
	AgentsUpdated  bool
	Blocked        bool
	HooksNeedTrust bool
}

type CoreStatus struct {
	ActiveCore, TotalCore int
	Checked, Attention    int
	Hooks                 []string
	State                 string
}

type Report struct {
	Root     string
	Items    []kit.Status
	Agents   []agents.Change
	Warnings []kit.Status
}

type RollbackPreview struct {
	Plan  RollbackPlan
	Diffs []FileDiff
}

type GitHooksRequest struct {
	Home, Repo, Executable string
	Hooks                  []string
	PreCommit              []string
	PrePush                []string
}

type CommandSource = githooks.Source

type GitHooksPlan struct {
	Request     GitHooksRequest
	Repo        githooks.Repo
	HomeChanges []kit.Change
	RepoChanges []kit.Change
	Diffs       []FileDiff
	Scripts     map[string]string
	Sources     []CommandSource
	home        kit.Plan
	repo        kit.Plan
}
