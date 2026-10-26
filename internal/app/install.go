package app

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/Gex-Development/gex-dev-kit/internal/agents"
	"github.com/Gex-Development/gex-dev-kit/internal/catalog"
	"github.com/Gex-Development/gex-dev-kit/internal/kit"
)

func PlanInstall(r Request, diffLimit int) (Plan, error) {
	if !filepath.IsAbs(r.Home) {
		return Plan{}, errors.New("HOME deve ser caminho absoluto")
	}
	files, err := catalog.Build(r.Home, r.Executable, r.Skills, r.Hooks)
	if err != nil {
		return Plan{}, fmt.Errorf("Catálogo: %w", err)
	}
	preview, err := kit.Preview(r.Home, files)
	if err != nil {
		return Plan{}, fmt.Errorf("Prévia: %w", err)
	}
	plan := Plan{Request: r, Files: preview.Changes, kit: preview}
	if !r.SkipAgentUpdate {
		changes, err := agents.Preview(r.Home)
		if err != nil {
			return Plan{}, fmt.Errorf("Prévia dos agentes: %w", err)
		}
		plan.Agents = changes
		for _, change := range changes {
			if change.Blocked {
				plan.Blocked = true
			} else if change.Current != change.Target {
				plan.Updates = append(plan.Updates, change)
			}
		}
	}
	plan.Notices = installNotices(r.Hooks)
	plan.Diffs = buildDiffs(preview.Contents(), diffLimit)
	return plan, nil
}

func (p Plan) Empty() bool {
	return len(p.Files) == 0 && len(p.Updates) == 0
}

func ApplyInstall(p Plan, progress func(Step)) (Result, error) {
	report := func(step Step) {
		if progress != nil {
			progress(step)
		}
	}
	result := Result{Blocked: p.Blocked, HooksNeedTrust: installNeedsTrust(p.Request.Hooks)}
	report(StepAgents)
	if len(p.Updates) > 0 {
		if err := agents.Apply(p.Request.Home, p.Updates); err != nil {
			return Result{}, fmt.Errorf("Atualização de agentes falhou; Core não foi aplicado: %w", err)
		}
		result.AgentsUpdated = true
	}
	report(StepFiles)
	if len(p.Files) > 0 {
		id, err := kit.Apply(p.kit)
		if err != nil {
			return Result{}, fmt.Errorf("Instalação do Core: %w", err)
		}
		result.HistoryID = id
	}
	report(StepDone)
	return result, nil
}

func installNeedsTrust(hooks []string) bool {
	for _, name := range hooks {
		if spec, ok := catalog.HookSpecByName(name); ok && spec.Codex.Supported {
			return true
		}
	}
	return false
}

func installNotices(hooks []string) []Notice {
	var notices []Notice
	for _, name := range hooks {
		spec, ok := catalog.HookSpecByName(name)
		if !ok {
			continue
		}
		notices = append(notices, installClaudeNotices(spec)...)
		notices = append(notices, installCodexNotice(spec))
	}
	return notices
}

func installClaudeNotices(spec catalog.HookSpec) []Notice {
	if !spec.Claude.Supported {
		return []Notice{{Level: "info", Text: spec.Name + " · Claude: não suportado — " + spec.Claude.Reason}}
	}
	line := spec.Name + " · Claude: " + spec.Claude.Event + " " + spec.Claude.Matcher + " · " + installDecisionText(spec.Claude.Decision)
	if spec.Policy == "gex" {
		line += " · só em repositório GEX"
	}
	notices := []Notice{{Level: "info", Text: line}}
	if spec.Claude.Decision == "pergunta" {
		notices = append(notices, Notice{Level: "aviso", Text: spec.Name + " · Claude em `claude -p`: a pergunta vira negação"})
	}
	return notices
}

func installCodexNotice(spec catalog.HookSpec) Notice {
	if !spec.Codex.Supported {
		return Notice{Level: "aviso", Text: spec.Name + " · Codex: não suportado — " + spec.Codex.Reason}
	}
	return Notice{Level: "aviso", Text: spec.Name + " · Codex: aprove em /hooks para ativar"}
}

func installDecisionText(decision string) string {
	switch decision {
	case "bloqueia":
		return "bloqueia"
	case "pergunta":
		return "pergunta antes"
	case "contexto":
		return "informa contexto"
	default:
		return decision
	}
}
