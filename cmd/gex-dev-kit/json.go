package main

import (
	"encoding/json"
	"io"
)

type jsonSkillsDoc struct {
	Schema      string          `json:"schema"`
	KitVersion  string          `json:"kit_version"`
	Core        []jsonCoreSkill `json:"core"`
	Specialties []jsonSpecialty `json:"specialties"`
	Matt        jsonMatt        `json:"matt"`
	References  []jsonReference `json:"references"`
}

type jsonCoreSkill struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type jsonSpecialty struct {
	Name        string   `json:"name"`
	Group       string   `json:"group"`
	Description string   `json:"description"`
	Requires    []string `json:"requires"`
	Warning     string   `json:"warning"`
}

type jsonMatt struct {
	Revision string          `json:"revision"`
	License  string          `json:"license"`
	Skills   []jsonMattSkill `json:"skills"`
}

type jsonMattSkill struct {
	Name        string   `json:"name"`
	Group       string   `json:"group"`
	Description string   `json:"description"`
	Requires    []string `json:"requires"`
	Source      string   `json:"source"`
}

type jsonReference struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Reason string `json:"reason"`
	Source string `json:"source"`
}

type jsonHooksDoc struct {
	Schema     string          `json:"schema"`
	AgentHooks []jsonAgentHook `json:"agent_hooks"`
	GitHooks   []jsonGitHook   `json:"git_hooks"`
	References []jsonReference `json:"references"`
}

type jsonAgentHook struct {
	Name        string          `json:"name"`
	Command     string          `json:"command"`
	Description string          `json:"description"`
	Source      string          `json:"source"`
	Scope       string          `json:"scope"`
	Claude      jsonHookSupport `json:"claude"`
	Codex       jsonHookSupport `json:"codex"`
}

type jsonHookSupport struct {
	Supported     bool   `json:"supported"`
	Event         string `json:"event"`
	Matcher       string `json:"matcher"`
	Decision      string `json:"decision"`
	Reason        string `json:"reason"`
	RequiresTrust bool   `json:"requires_trust"`
}

type jsonGitHook struct {
	Name             string `json:"name"`
	GitHook          string `json:"git_hook"`
	Description      string `json:"description"`
	RequiresRepo     bool   `json:"requires_repo"`
	RequiresCommands bool   `json:"requires_commands"`
}

type jsonAddonsDoc struct {
	Schema string      `json:"schema"`
	Addons []jsonAddon `json:"addons"`
}

type jsonAddon struct {
	Name          string            `json:"name"`
	Kind          string            `json:"kind"`
	Summary       string            `json:"summary"`
	Source        string            `json:"source"`
	License       string            `json:"license"`
	Revision      string            `json:"revision"`
	Release       string            `json:"release"`
	CheckedAt     string            `json:"checked_at"`
	Install       jsonInstall       `json:"install"`
	Pinned        string            `json:"pinned"`
	Deterministic bool              `json:"deterministic"`
	Installs      []string          `json:"installs"`
	Conflicts     []string          `json:"conflicts"`
	CoreCompat    string            `json:"core_compat"`
	ExecutedByKit bool              `json:"executed_by_kit"`
	Secrets       []jsonAddonSecret `json:"secrets"`
}

type jsonAddonSecret struct {
	Env   string `json:"env"`
	Label string `json:"label"`
	URL   string `json:"url"`
}

type jsonInstall struct {
	Claude string `json:"claude"`
	Codex  string `json:"codex"`
}

type jsonDoctorDoc struct {
	Schema   string            `json:"schema"`
	Root     string            `json:"root"`
	OK       bool              `json:"ok"`
	ExitCode int               `json:"exit_code"`
	Counts   jsonDoctorCounts  `json:"counts"`
	Items    []jsonDoctorItem  `json:"items"`
	Agents   []jsonDoctorAgent `json:"agents"`
	Warnings []jsonDoctorItem  `json:"warnings"`
}

type jsonDoctorCounts struct {
	OK        int `json:"ok"`
	Attention int `json:"attention"`
	Pending   int `json:"pending"`
}

type jsonDoctorItem struct {
	Path   string `json:"path"`
	State  string `json:"state"`
	Level  string `json:"level"`
	Detail string `json:"detail"`
}

type jsonDoctorAgent struct {
	Name    string `json:"name"`
	Managed bool   `json:"managed"`
	Current string `json:"current"`
	Target  string `json:"target"`
	State   string `json:"state"`
	Reason  string `json:"reason"`
	Action  string `json:"action"`
}

func writeJSON(w io.Writer, value any) error {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}
