package screens

import (
	"fmt"
	"strings"

	"github.com/Gex-Development/gex-dev-kit/internal/app"
	"github.com/Gex-Development/gex-dev-kit/internal/catalog"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/components"
)

type chatCatalogItem struct {
	name, badge, desc, source, group string
	hook                             *catalog.HookSpec
}

func chatCatalogSkillItems() []chatCatalogItem {
	var items []chatCatalogItem
	for _, name := range catalog.CoreSkills() {
		items = append(items, chatCatalogItem{name: name, badge: "Core", group: "Core", desc: "Skill do bloco Core, instalada com o kit."})
	}
	for _, group := range app.InstallOptions()[:2] {
		for _, option := range group.Options {
			name := option.Value
			items = append(items, chatCatalogItem{name: name, badge: "Matt", group: group.Title, desc: catalog.SkillDescription(name), source: catalog.MattSource(name)})
		}
	}
	for _, s := range catalog.Specialties() {
		items = append(items, chatCatalogItem{name: s.Name, badge: "Especialização", group: s.Group, desc: s.Description})
	}
	return items
}

func chatCatalogRefItems() []chatCatalogItem {
	var items []chatCatalogItem
	for _, ref := range catalog.ReferencedSkills() {
		items = append(items, chatCatalogItem{name: ref.Name, badge: ref.Status, group: "Skills", desc: ref.Reason, source: ref.Source})
	}
	for _, ref := range catalog.HookCandidates() {
		items = append(items, chatCatalogItem{name: ref.Name, badge: ref.Status, group: "Hooks candidatos", desc: ref.Reason, source: ref.Source})
	}
	return items
}

func (m chatModel) flowCatalog() chatModel {
	m.context = "catálogo"
	m = m.say(fmt.Sprintf("Catálogo: %d Core · %d Matt · %d especializações.", len(catalog.CoreSkills()), len(catalog.Skills()), len(catalog.Specialties())))
	return m.catalogCategories()
}

func (m chatModel) catalogCategories() chatModel {
	opts := []components.ChoiceOption{{ID: "skills", Label: "Skills"}, {ID: "hooks", Label: "Hooks"}, {ID: "refs", Label: "Referências"}, {ID: "menu", Label: "Concluir"}}
	return m.askFor(chatAsk{hasChoice: true, question: "O que você quer ver?", choice: components.NewChoice("O que você quer ver?", opts, false, true), options: opts, onAnswer: func(m chatModel, a chatAnswer) chatModel {
		if a.IDs[0] == "menu" {
			return m.flowIdle()
		}
		if a.IDs[0] == "hooks" {
			return m.catalogItems("hooks", "")
		}
		return m.catalogGroups(a.IDs[0])
	}})
}

func (m chatModel) catalogGroups(category string) chatModel {
	var opts []components.ChoiceOption
	if category == "skills" {
		opts = append(opts, components.ChoiceOption{ID: "Core", Label: fmt.Sprintf("Core (%d)", len(catalog.CoreSkills()))})
		for _, group := range app.InstallOptions()[:2] {
			opts = append(opts, components.ChoiceOption{ID: group.Title, Label: fmt.Sprintf("%s (%d)", group.Title, len(group.Options))})
		}
		opts = append(opts, components.ChoiceOption{ID: "Especializações", Label: fmt.Sprintf("Especializações (%d)", len(catalog.Specialties()))})
	} else {
		opts = []components.ChoiceOption{
			{ID: "Skills de terceiros", Label: fmt.Sprintf("Skills de terceiros (%d)", len(catalog.ReferencedSkills()))},
			{ID: "Hooks candidatos", Label: fmt.Sprintf("Hooks candidatos (%d)", len(catalog.HookCandidates()))},
		}
	}
	opts = append(opts, components.ChoiceOption{ID: "menu", Label: "Voltar às categorias"})
	return m.askFor(chatAsk{hasChoice: true, question: "Qual grupo você quer ver?", choice: components.NewChoice("Qual grupo você quer ver?", opts, false, true), options: opts, onAnswer: func(m chatModel, a chatAnswer) chatModel {
		if a.IDs[0] == "menu" {
			return m.catalogCategories()
		}
		if a.IDs[0] == "Especializações" {
			return m.catalogSpecialtyGroups()
		}
		return m.catalogItems(category, a.IDs[0])
	}})
}

func (m chatModel) catalogSpecialtyGroups() chatModel {
	var opts []components.ChoiceOption
	for _, group := range catalog.SpecialtyGroups() {
		count := 0
		for _, specialty := range catalog.Specialties() {
			if specialty.Group == group {
				count++
			}
		}
		opts = append(opts, components.ChoiceOption{ID: group, Label: fmt.Sprintf("%s (%d)", group, count)})
	}
	opts = append(opts, components.ChoiceOption{ID: "menu", Label: "Voltar aos grupos"})
	return m.askFor(chatAsk{hasChoice: true, question: "Qual grupo de especializações?", choice: components.NewChoice("Qual grupo de especializações?", opts, false, true), options: opts, onAnswer: func(m chatModel, a chatAnswer) chatModel {
		if a.IDs[0] == "menu" {
			return m.catalogGroups("skills")
		}
		return m.catalogItems("skills", a.IDs[0])
	}})
}

func (m chatModel) catalogItems(category, group string) chatModel {
	var items []chatCatalogItem
	switch category {
	case "skills":
		for _, item := range chatCatalogSkillItems() {
			if item.group == group {
				items = append(items, item)
			}
		}
	case "refs":
		for _, item := range chatCatalogRefItems() {
			if item.group == group || group == "Skills de terceiros" && item.group == "Skills" {
				items = append(items, item)
			}
		}
	case "hooks":
		for _, spec := range catalog.HookSpecs() {
			s := spec
			group := "Agente"
			if s.Scope == catalog.HookScopeGit {
				group = "Git"
			}
			items = append(items, chatCatalogItem{name: s.Name, badge: group, group: group, desc: s.Description, source: s.Source, hook: &s})
		}
	}
	opts := make([]components.ChoiceOption, 0, len(items)+1)
	for _, item := range items {
		opts = append(opts, components.ChoiceOption{ID: item.name, Label: item.name, Hint: "[" + item.badge + "] " + item.desc})
	}
	back := func(m chatModel) chatModel {
		if category == "hooks" {
			return m.catalogCategories()
		}
		if category == "skills" && (group == "Dados e Plataforma" || group == "Produto e Interface") {
			return m.catalogSpecialtyGroups()
		}
		return m.catalogGroups(category)
	}
	opts = append(opts, components.ChoiceOption{ID: "menu", Label: "Voltar aos grupos"})
	return m.askFor(chatAsk{hasChoice: true, question: "Qual item você quer ver?", choice: components.NewChoice("Qual item você quer ver?", opts, false, true), options: opts, onAnswer: func(m chatModel, a chatAnswer) chatModel {
		if a.IDs[0] == "menu" {
			return back(m)
		}
		for _, item := range items {
			if item.name == a.IDs[0] {
				m = m.say(chatCatalogDetail(item))
				break
			}
		}
		return m.catalogItems(category, group)
	}})
}

func chatCatalogDetail(item chatCatalogItem) string {
	description := item.desc
	requirement := ""
	if before, after, ok := strings.Cut(description, " Requer seleção conjunta: "); ok {
		description = before
		requirement = "Requer seleção conjunta: " + after
	}
	lines := []string{item.name + " [" + item.badge + "]"}
	if explanation, ok := catalog.Explain(item.name); ok && item.group != "Skills" && item.group != "Hooks candidatos" {
		lines = append(lines, "", "O que faz: "+explanation.What, "", "Quando usar: "+explanation.UseCase, "", "Analogia: "+explanation.Analogy, "")
	} else {
		lines = append(lines, description)
	}
	if requirement != "" {
		lines = append(lines, requirement)
	}
	if item.hook != nil {
		s := item.hook
		lines = append(lines, "Claude  "+chatCatalogSupport(s.Claude), "Codex   "+chatCatalogSupport(s.Codex), "Git     "+chatCatalogYesNo(s.GitHook != "")+" · "+chatCatalogEvent(*s), "Evento  "+chatCatalogEvent(*s), "Escopo  "+chatCatalogPolicy(s.Policy))
		if s.Command != "" {
			lines = append(lines, "Comando: "+s.Command)
		}
	}
	if item.source != "" {
		lines = append(lines, "Fonte: "+item.source)
	}
	return strings.Join(lines, "\n")
}

func chatCatalogYesNo(ok bool) string {
	if ok {
		return "sim"
	}
	return "não"
}

func chatCatalogEvent(s catalog.HookSpec) string {
	if s.GitHook != "" {
		return s.GitHook
	}
	if s.Claude.Supported {
		return s.Claude.Event
	}
	if s.Codex.Supported {
		return s.Codex.Event
	}
	return "—"
}

func chatCatalogPolicy(policy string) string {
	if policy == "gex" {
		return "só GEX"
	}
	return "todos"
}

func chatCatalogSupport(s catalog.HookSupport) string {
	if s.Supported {
		return "suportado · " + s.Event + " · decisão " + s.Decision
	}
	return "não suportado. " + s.Reason
}
