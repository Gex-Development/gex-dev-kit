package catalog

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
)

//go:embed assets/core assets/matt assets/specialties
var content embed.FS

var (
	ErrUnknownSkill = errors.New("skill indisponível")
	ErrUnknownHook  = errors.New("hook indisponível")
)

type File struct {
	Path string
	Data []byte
	Mode fs.FileMode
}

const MattRevision = "5beb35b040e3de97e54fa9e386d7e42d15c295c2"

var coreSkills = []string{
	"gex-abrir-issue", "gex-auto-revisao", "gex-fluxo-desenvolvimento",
	"gex-preparar-pr", "gex-regras-de-negocio", "gex-revisar-pr", "gex-verificar-mudanca",
}

func MattSource(name string) string {
	if _, ok := skillDescriptions[name]; !ok {
		return ""
	}
	return "https://github.com/mattpocock/skills/tree/" + MattRevision + "/skills/" + skillGroup(name) + "/" + name
}

func skillGroup(name string) string {
	switch name {
	case "grilling", "grill-me", "handoff", "teach", "to-questionnaire", "wait-what", "writing-for-agents":
		return "productivity"
	default:
		return "engineering"
	}
}

var skillDescriptions = map[string]string{
	"pr":                            "Modelo de corpo de PR com evidências e riscos; não substitui exigências GEX.",
	"tdd":                           "Testes antes da implementação, com ciclo red-green-refactor.",
	"diagnosing-bugs":               "Investigação de bugs difíceis e regressões de desempenho.",
	"code-review":                   "Revisão técnica; integração com tracker exige configuração prévia separada.",
	"codebase-design":               "Desenho de interfaces e módulos de código.",
	"prototype":                     "Protótipos descartáveis de lógica e interface.",
	"domain-modeling":               "Glossário e decisões de domínio.",
	"improve-codebase-architecture": "Análise de arquitetura e relatório HTML.",
	"research":                      "Pesquisa técnica com fontes registradas.",
	"retro":                         "Retrospectiva de sessão; lê logs locais desta máquina só por invocação explícita e só sugere mudanças.",
	"grill-with-docs":               "Entrevista com documentação de domínio.",
	"grilling":                      "Entrevista para decisões de produto e arquitetura.",
	"grill-me":                      "Atalho para entrevistas com grilling.",
	"handoff":                       "Documento de transferência de contexto.",
	"teach":                         "Sessões de estudo com exercícios e materiais.",
	"to-questionnaire":              "Questionário para decisões pendentes.",
	"wait-what":                     "Reformulação técnica simplificada em inglês.",
	"writing-for-agents":            "Escrita de skills e instruções para agentes.",
}

var skillDependencies = map[string][]string{
	"grill-me":                      {"grilling"},
	"grill-with-docs":               {"grilling", "domain-modeling"},
	"improve-codebase-architecture": {"codebase-design", "grilling", "domain-modeling"},
	"retro":                         {"writing-for-agents"},
}

func Skills() []string {
	names := make([]string, 0, len(skillDescriptions))
	for name := range skillDescriptions {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

func CoreSkills() []string { return slices.Clone(coreSkills) }

func SpecialtySkills() []string {
	specialties := Specialties()
	names := make([]string, 0, len(specialties))
	for _, specialty := range specialties {
		names = append(names, specialty.Name)
	}
	slices.Sort(names)
	return names
}

func SpecialtyDescription(name string) string {
	specialty, _ := SpecialtyByName(name)
	return specialty.Description
}

func SkillDescription(name string) string {
	description := skillDescriptions[name]
	if description == "" {
		description = SpecialtyDescription(name)
	}
	if dependencies := catalogRequirements(name); len(dependencies) != 0 {
		description += " Requer seleção conjunta: " + strings.Join(dependencies, ", ") + "."
	}
	return description
}

func Hooks() []string {
	var names []string
	for _, spec := range HookSpecs() {
		if spec.Scope == HookScopeAgent {
			names = append(names, spec.Name)
		}
	}
	return names
}

func HookDescription(name string) string {
	spec, ok := HookSpecByName(name)
	if !ok || spec.Scope != HookScopeAgent {
		return ""
	}
	return spec.Description
}

func ValidateSelection(skills, hooks []string) error {
	return validateSelection(skills, hooks, false)
}

func ValidateEngineeringSelection(skills []string) error {
	return validateSelection(skills, nil, true)
}

func validateSelection(skills, hooks []string, deferProductivity bool) error {
	for _, name := range skills {
		if _, matt := skillDescriptions[name]; !matt {
			if _, specialty := SpecialtyByName(name); !specialty {
				return fmt.Errorf("%w: %s", ErrUnknownSkill, name)
			}
		}
	}
	for _, name := range skills {
		for _, dependency := range catalogRequirements(name) {
			if !slices.Contains(skills, dependency) && !(deferProductivity && skillGroup(dependency) == "productivity") {
				return fmt.Errorf("%s requer selecionar %s", name, dependency)
			}
		}
	}
	for _, name := range hooks {
		if !slices.Contains(Hooks(), name) {
			return fmt.Errorf("%w: %s", ErrUnknownHook, name)
		}
	}
	return nil
}

func Build(home, executable string, skills, hooks []string) ([]File, error) {
	if home == "" || !filepath.IsAbs(home) {
		return nil, errors.New("HOME deve ser um caminho absoluto explícito")
	}
	if err := ValidateSelection(skills, hooks); err != nil {
		return nil, err
	}
	files := make([]File, 0, 20)
	core, err := content.ReadFile("assets/core/AGENTS.md")
	if err != nil {
		return nil, err
	}
	for _, dest := range []string{".claude/CLAUDE.md", ".codex/AGENTS.md"} {
		previous, err := readHome(home, dest)
		if err != nil {
			return nil, err
		}
		merged, err := mergeCore(previous, core)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", dest, err)
		}
		files = append(files, File{Path: dest, Data: merged, Mode: 0600})
	}
	for _, name := range coreSkills {
		if err := appendPack(&files, "assets/core/skills/"+name, name); err != nil {
			return nil, err
		}
	}
	seen := make(map[string]bool)
	license, err := content.ReadFile("assets/matt/LICENSE")
	if err != nil {
		return nil, err
	}
	for _, name := range skills {
		if seen[name] {
			continue
		}
		seen[name] = true
		if _, specialty := SpecialtyByName(name); specialty {
			if err := appendPack(&files, "assets/specialties/"+name, name); err != nil {
				return nil, err
			}
			continue
		}
		base := "assets/matt/skills/" + skillGroup(name) + "/" + name
		if err := fs.WalkDir(content, base, func(asset string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				return nil
			}
			if path.Ext(asset) == ".sh" && asset != "assets/matt/skills/engineering/diagnosing-bugs/scripts/hitl-loop.template.sh" {
				return fmt.Errorf("script terceiro não auditado: %s", asset)
			}
			data, err := content.ReadFile(asset)
			if err != nil {
				return err
			}
			rel := strings.TrimPrefix(asset, base+"/")
			for _, dest := range []string{".claude/skills/", ".agents/skills/"} {
				files = append(files, File{Path: dest + name + "/" + rel, Data: data, Mode: 0644})
			}
			return nil
		}); err != nil {
			return nil, err
		}
		for _, dest := range []string{".claude/skills/", ".agents/skills/"} {
			files = append(files, File{Path: dest + name + "/LICENSE", Data: license, Mode: 0644})
		}
	}
	if err := appendHooks(&files, home, executable, hooks); err != nil {
		return nil, err
	}
	return files, nil
}

func appendPack(files *[]File, base, name string) error {
	if _, err := content.ReadFile(base + "/SKILL.md"); err != nil {
		return fmt.Errorf("skill %s sem SKILL.md: %w", name, err)
	}
	return fs.WalkDir(content, base, func(asset string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		data, err := content.ReadFile(asset)
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(asset, base+"/")
		for _, dest := range []string{".claude/skills/", ".agents/skills/"} {
			*files = append(*files, File{Path: dest + name + "/" + rel, Data: data, Mode: 0644})
		}
		return nil
	})
}

func StableBinary(home string) string {
	return filepath.Join(home, ".local", "share", "gex-dev-kit", "bin", catalogBinaryName())
}

func BinaryFile(executable string) (File, error) {
	if executable == "" {
		return File{}, errors.New("executável obrigatório para hook")
	}
	binary, err := os.ReadFile(executable)
	if err != nil {
		return File{}, fmt.Errorf("não foi possível ler executável: %w", err)
	}
	return File{Path: stableRelative(), Data: binary, Mode: 0755}, nil
}

func catalogBinaryName() string {
	if runtime.GOOS == "windows" {
		return "gex-dev-kit.exe"
	}
	return "gex-dev-kit"
}

func catalogRequirements(name string) []string {
	requires := slices.Clone(skillDependencies[name])
	if specialty, ok := SpecialtyByName(name); ok {
		requires = append(requires, specialty.Requires...)
	}
	return requires
}

func readHome(home, relative string) ([]byte, error) {
	current := home
	for _, part := range strings.Split(relative, "/") {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("link simbólico não permitido: %s", relative)
		}
	}
	return os.ReadFile(current)
}

func mergeCore(previous, core []byte) ([]byte, error) {
	const start = "<!-- início gex-dev-kit core -->\n"
	const end = "<!-- fim gex-dev-kit core -->"
	begin := bytes.Index(previous, []byte(start))
	finish := bytes.Index(previous, []byte(end))
	if (begin < 0) != (finish < 0) || begin >= 0 && (finish < begin || bytes.Count(previous, []byte(start)) != 1 || bytes.Count(previous, []byte(end)) != 1) {
		return nil, errors.New("marcadores Core incompletos ou duplicados")
	}
	section := append([]byte(start), core...)
	section = append(section, []byte(end)...)
	if begin >= 0 {
		result := append([]byte{}, previous[:begin]...)
		result = append(result, section...)
		return append(result, previous[finish+len(end):]...), nil
	}
	result := append([]byte{}, previous...)
	if len(result) > 0 && result[len(result)-1] != '\n' {
		result = append(result, '\n')
	}
	if len(result) > 0 {
		result = append(result, '\n')
	}
	return append(append(result, section...), '\n'), nil
}

type hookTarget struct {
	path  string
	agent Agent
}

var hookTargets = []hookTarget{{".claude/settings.json", AgentClaude}, {".codex/hooks.json", AgentCodex}}

type hookHandler struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

func stableRelative() string {
	return path.Join(".local", "share", "gex-dev-kit", "bin", catalogBinaryName())
}

func IsKitHandler(command string, args []string, executable string) (string, bool) {
	if executable == "" {
		return "", false
	}
	if args != nil {
		if command == executable && len(args) >= 2 && args[0] == "hook" && args[1] != "" {
			return args[1], true
		}
		return "", false
	}
	if command == executable {
		return "commit-msg", true
	}
	for _, prefix := range []string{executable + " hook ", quoteCommand(executable) + " hook "} {
		if rest, found := strings.CutPrefix(command, prefix); found {
			if fields := strings.Fields(rest); len(fields) != 0 {
				return fields[0], true
			}
		}
	}
	return "", false
}

func selectedHookSpecs(names []string) []HookSpec {
	var selected []HookSpec
	for _, spec := range HookSpecs() {
		if spec.Scope == HookScopeAgent && slices.Contains(names, spec.Name) {
			selected = append(selected, spec)
		}
	}
	return selected
}

func supportFor(spec HookSpec, agent Agent) HookSupport {
	if agent == AgentCodex {
		return spec.Codex
	}
	return spec.Claude
}

func anySupported(specs []HookSpec, agent Agent) bool {
	return slices.ContainsFunc(specs, func(spec HookSpec) bool {
		return supportFor(spec, agent).Supported
	})
}

func appendHooks(files *[]File, home, executable string, hooks []string) error {
	selected := selectedHookSpecs(hooks)
	stable := StableBinary(home)
	installed := false
	for _, target := range hookTargets {
		if data, err := readHome(home, target.path); err == nil && hasKitHandler(data, stable) {
			installed = true
		}
	}
	refresh := len(selected) != 0 || installed
	if !refresh && executable != "" {
		changed, err := stableDiffers(home, executable)
		if err != nil {
			return err
		}
		refresh = changed
	}
	if refresh {
		binary, err := BinaryFile(executable)
		if err != nil {
			return err
		}
		*files = append(*files, binary)
	}
	for _, target := range hookTargets {
		if !anySupported(selected, target.agent) {
			continue
		}
		previous, err := readHome(home, target.path)
		if err != nil {
			return fmt.Errorf("%s: %w", target.path, err)
		}
		merged, err := mergeHooks(previous, stable, target.agent, selected)
		if err != nil {
			return fmt.Errorf("%s: %w", target.path, err)
		}
		*files = append(*files, File{Path: target.path, Data: merged, Mode: 0600})
	}
	return nil
}

func stableDiffers(home, executable string) (bool, error) {
	current, err := BinaryFile(executable)
	if err != nil {
		return false, err
	}
	existing, err := readHome(home, stableRelative())
	if err != nil {
		return false, err
	}
	return existing != nil && !bytes.Equal(existing, current.Data), nil
}

func hasKitHandler(data []byte, executable string) bool {
	_, events, err := hookEvents(data)
	if err != nil {
		return false
	}
	for event, raw := range events {
		groups, err := hookGroups(raw, event)
		if err != nil {
			continue
		}
		for _, group := range groups {
			_, handlers, err := groupHandlers(group)
			if err != nil {
				continue
			}
			for _, handler := range handlers {
				if _, ok, err := kitIdentity(handler, executable); err == nil && ok {
					return true
				}
			}
		}
	}
	return false
}

func mergeHooks(previous []byte, executable string, agent Agent, selected []HookSpec) ([]byte, error) {
	settings, events, err := hookEvents(previous)
	if err != nil {
		return nil, err
	}
	for event, raw := range events {
		groups, err := hookGroups(raw, event)
		if err != nil {
			return nil, err
		}
		kept, err := keepHookGroups(groups, executable, selected)
		if err != nil {
			return nil, err
		}
		if len(kept) == 0 {
			delete(events, event)
			continue
		}
		if events[event], err = marshalHook(kept); err != nil {
			return nil, err
		}
	}
	for _, spec := range selected {
		support := supportFor(spec, agent)
		if !support.Supported {
			continue
		}
		group, err := marshalHook(map[string]any{"matcher": support.Matcher, "hooks": []any{kitHandler(spec, agent, executable)}})
		if err != nil {
			return nil, err
		}
		if err := appendHookGroup(events, support.Event, group); err != nil {
			return nil, err
		}
	}
	if settings["hooks"], err = marshalHook(events); err != nil {
		return nil, err
	}
	return marshalSettings(settings)
}

func keepHookGroups(groups []json.RawMessage, executable string, selected []HookSpec) ([]json.RawMessage, error) {
	kept := make([]json.RawMessage, 0, len(groups))
	for _, group := range groups {
		fields, handlers, err := groupHandlers(group)
		if err != nil {
			return nil, err
		}
		remaining := make([]json.RawMessage, 0, len(handlers))
		for _, handler := range handlers {
			sub, ok, err := kitIdentity(handler, executable)
			if err != nil {
				return nil, err
			}
			if ok && slices.ContainsFunc(selected, func(spec HookSpec) bool { return spec.Command == sub }) {
				continue
			}
			remaining = append(remaining, handler)
		}
		if len(remaining) == 0 {
			continue
		}
		if len(remaining) == len(handlers) {
			kept = append(kept, group)
			continue
		}
		fields["hooks"], err = marshalHook(remaining)
		if err != nil {
			return nil, err
		}
		updated, err := marshalHook(fields)
		if err != nil {
			return nil, err
		}
		kept = append(kept, updated)
	}
	return kept, nil
}

func appendHookGroup(events map[string]json.RawMessage, event string, group json.RawMessage) error {
	var groups []json.RawMessage
	if raw, ok := events[event]; ok {
		parsed, err := hookGroups(raw, event)
		if err != nil {
			return err
		}
		groups = parsed
	}
	updated, err := marshalHook(append(groups, group))
	if err != nil {
		return err
	}
	events[event] = updated
	return nil
}

func kitIdentity(raw json.RawMessage, executable string) (string, bool, error) {
	var handler hookHandler
	if err := json.Unmarshal(raw, &handler); err != nil {
		return "", false, hookError("handler de hook malformado")
	}
	sub, ok := IsKitHandler(handler.Command, handler.Args, executable)
	return sub, ok, nil
}

func kitHandler(spec HookSpec, agent Agent, executable string) map[string]any {
	args := append([]string{"hook", spec.Command}, agentArgs(spec, agent)...)
	if agent == AgentCodex {
		suffix := strings.Join(args, " ")
		return map[string]any{
			"type":           "command",
			"command":        quoteCommand(executable) + " " + suffix,
			"timeout":        10,
			"commandWindows": windowsCommandPath(executable, shortPath) + " " + suffix,
		}
	}
	handler := map[string]any{"type": "command", "command": executable, "args": args, "timeout": 10}
	if spec.Command == "commit-msg" {
		handler["onFailure"] = "block"
	}
	return handler
}

func agentArgs(spec HookSpec, agent Agent) []string {
	if spec.Command != "session-start" {
		return nil
	}
	return []string{"--agent", agentName(agent)}
}

func agentName(agent Agent) string {
	if agent == AgentCodex {
		return "codex"
	}
	return "claude"
}

func windowsCommandPath(executable string, short func(string) (string, error)) string {
	if !strings.ContainsAny(executable, " \t") {
		return executable
	}
	if short != nil {
		if shortened, err := short(executable); err == nil && shortened != "" && !strings.ContainsAny(shortened, " \t") {
			return shortened
		}
	}
	return quoteCommand(executable)
}

func hookEvents(data []byte) (map[string]json.RawMessage, map[string]json.RawMessage, error) {
	settings, err := decodeHookObject(data)
	if err != nil {
		return nil, nil, err
	}
	events := map[string]json.RawMessage{}
	raw := settings["hooks"]
	if len(raw) == 0 {
		return settings, events, nil
	}
	if err := json.Unmarshal(raw, &events); err != nil || events == nil {
		return nil, nil, hookError("campo hooks deve ser objeto")
	}
	return settings, events, nil
}

func decodeHookObject(data []byte) (map[string]json.RawMessage, error) {
	object := map[string]json.RawMessage{}
	if len(bytes.TrimSpace(data)) == 0 {
		return object, nil
	}
	if err := json.Unmarshal(data, &object); err != nil || object == nil {
		return nil, hookError("JSON inválido ou não é objeto")
	}
	return object, nil
}

func hookGroups(raw json.RawMessage, event string) ([]json.RawMessage, error) {
	var groups []json.RawMessage
	if err := json.Unmarshal(raw, &groups); err != nil || groups == nil {
		return nil, hookError("evento " + event + " deve ser lista de grupos")
	}
	return groups, nil
}

func groupHandlers(group json.RawMessage) (map[string]json.RawMessage, []json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(group, &fields); err != nil || fields == nil {
		return nil, nil, hookError("grupo de hook malformado")
	}
	var handlers []json.RawMessage
	if err := json.Unmarshal(fields["hooks"], &handlers); err != nil || handlers == nil {
		return nil, nil, hookError("lista hooks do grupo malformada")
	}
	return fields, handlers, nil
}

func hookError(detail string) error {
	return errors.New(detail + "; corrija o JSON; nada foi alterado")
}

func marshalHook(value any) (json.RawMessage, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return bytes.TrimRight(buffer.Bytes(), "\n"), nil
}

func marshalSettings(settings map[string]json.RawMessage) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(settings); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func quoteCommand(executable string) string {
	return `"` + strings.ReplaceAll(executable, `"`, `\"`) + `"`
}
