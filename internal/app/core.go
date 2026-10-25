package app

import (
	"github.com/Gex-Development/gex-dev-kit/internal/catalog"
	"github.com/Gex-Development/gex-dev-kit/internal/kit"
)

func Core(home string) (CoreStatus, error) {
	statuses, err := kit.Doctor(home)
	if err != nil {
		return CoreStatus{}, err
	}
	return coreSummary(statuses), nil
}

func coreSummary(statuses []kit.Status) CoreStatus {
	stateByPath := make(map[string]string, len(statuses))
	for _, status := range statuses {
		stateByPath[status.Path] = status.State
	}
	summary := CoreStatus{TotalCore: len(catalog.CoreSkills()), State: "íntegro"}
	for _, name := range catalog.CoreSkills() {
		if stateByPath[".claude/skills/"+name+"/SKILL.md"] == "ok" && stateByPath[".agents/skills/"+name+"/SKILL.md"] == "ok" {
			summary.ActiveCore++
		}
	}
	for _, status := range statuses {
		switch status.State {
		case "ok", "pendente":
			summary.Checked++
			if status.State == "pendente" && summary.Attention == 0 {
				summary.State = "pendente"
			}
		case "sem-historico":
			summary.State = "não instalado"
		default:
			summary.Attention++
			summary.State = status.State
		}
	}
	if summary.State == "íntegro" && summary.ActiveCore != summary.TotalCore {
		summary.State = "atualização disponível"
	}
	if coreHookFileManaged(stateByPath) {
		summary.Hooks = catalog.Hooks()
	}
	return summary
}

func coreHookFileManaged(stateByPath map[string]string) bool {
	for _, path := range []string{".claude/settings.json", ".codex/hooks.json"} {
		if state := stateByPath[path]; state == "ok" || state == "pendente" {
			return true
		}
	}
	return false
}
