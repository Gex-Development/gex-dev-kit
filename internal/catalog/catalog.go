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

//go:embed assets/core assets/matt
var content embed.FS

type File struct {
	Path string
	Data []byte
	Mode fs.FileMode
}

const MattRevision = "5beb35b040e3de97e54fa9e386d7e42d15c295c2"

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
}

func Skills() []string {
	names := make([]string, 0, len(skillDescriptions))
	for name := range skillDescriptions {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

func SkillDescription(name string) string {
	description := skillDescriptions[name]
	if dependencies := skillDependencies[name]; len(dependencies) != 0 {
		description += " Requer seleção conjunta: " + strings.Join(dependencies, ", ") + "."
	}
	return description
}

func Hooks() []string { return []string{"gex-conventional-commits"} }

func HookDescription(name string) string {
	if name == "gex-conventional-commits" {
		return "Bloqueia mensagens git commit fora do padrão Conventional Commits."
	}
	return ""
}

func Build(home, executable string, skills, hooks []string) ([]File, error) {
	if home == "" || !filepath.IsAbs(home) {
		return nil, errors.New("HOME deve ser um caminho absoluto explícito")
	}
	for _, name := range skills {
		if _, ok := skillDescriptions[name]; !ok {
			return nil, fmt.Errorf("skill indisponível: %s", name)
		}
	}
	for name, dependencies := range skillDependencies {
		if !slices.Contains(skills, name) {
			continue
		}
		for _, dependency := range dependencies {
			if !slices.Contains(skills, dependency) {
				return nil, fmt.Errorf("%s requer selecionar %s", name, dependency)
			}
		}
	}
	for _, name := range hooks {
		if name != "gex-conventional-commits" {
			return nil, fmt.Errorf("hook indisponível: %s", name)
		}
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
	if len(hooks) != 0 {
		if executable == "" {
			return nil, errors.New("executável obrigatório para hook")
		}
		binary, err := os.ReadFile(executable)
		if err != nil {
			return nil, fmt.Errorf("não foi possível ler executável: %w", err)
		}
		name := "gex-dev-kit"
		if runtime.GOOS == "windows" {
			name += ".exe"
		}
		stable := filepath.Join(home, ".local", "share", "gex-dev-kit", "bin", name)
		files = append(files, File{Path: filepath.ToSlash(filepath.Join(".local", "share", "gex-dev-kit", "bin", name)), Data: binary, Mode: 0755})
		for _, dest := range []string{".claude/settings.json", ".codex/hooks.json"} {
			previous, err := readHome(home, dest)
			if err != nil {
				return nil, err
			}
			merged, err := mergeHook(previous, stable, strings.HasPrefix(dest, ".claude/"))
			if err != nil {
				return nil, fmt.Errorf("%s: %w", dest, err)
			}
			files = append(files, File{Path: dest, Data: merged, Mode: 0600})
		}
	}
	return files, nil
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

func mergeHook(previous []byte, executable string, claude bool) ([]byte, error) {
	settings := map[string]json.RawMessage{}
	if len(bytes.TrimSpace(previous)) != 0 {
		if err := json.Unmarshal(previous, &settings); err != nil {
			return nil, errors.New("JSON de configuração inválido")
		}
		if settings == nil {
			return nil, errors.New("configuração não é objeto JSON")
		}
	}
	events := map[string]json.RawMessage{}
	if raw := settings["hooks"]; len(raw) > 0 {
		if err := json.Unmarshal(raw, &events); err != nil || events == nil {
			return nil, errors.New("hooks inválidos")
		}
	}
	var groups []json.RawMessage
	if raw := events["PreToolUse"]; len(raw) > 0 {
		if err := json.Unmarshal(raw, &groups); err != nil || groups == nil {
			return nil, errors.New("PreToolUse inválido")
		}
	}
	filtered := make([]json.RawMessage, 0, len(groups)+1)
	for _, group := range groups {
		var g map[string]json.RawMessage
		if err := json.Unmarshal(group, &g); err != nil || g == nil {
			return nil, errors.New("grupo PreToolUse inválido")
		}
		var handlers []json.RawMessage
		if err := json.Unmarshal(g["hooks"], &handlers); err != nil || handlers == nil {
			return nil, errors.New("handlers PreToolUse inválidos")
		}
		kept := make([]json.RawMessage, 0, len(handlers))
		for _, raw := range handlers {
			var handler struct {
				Command string `json:"command"`
			}
			if err := json.Unmarshal(raw, &handler); err != nil {
				return nil, errors.New("handler PreToolUse inválido")
			}
			if handler.Command != executable && handler.Command != quoteCommand(executable)+" hook commit-msg" {
				kept = append(kept, raw)
			}
		}
		if len(kept) == 0 {
			continue
		}
		g["hooks"], _ = json.Marshal(kept)
		updated, _ := json.Marshal(g)
		filtered = append(filtered, updated)
	}
	var handler map[string]any
	if claude {
		handler = map[string]any{"type": "command", "command": executable, "args": []string{"hook", "commit-msg"}, "onFailure": "block"}
	} else {
		handler = map[string]any{"type": "command", "command": quoteCommand(executable) + " hook commit-msg", "timeout": 10, "commandWindows": quoteCommand(executable) + " hook commit-msg"}
	}
	group, err := json.Marshal(map[string]any{"matcher": "Bash", "hooks": []any{handler}})
	if err != nil {
		return nil, err
	}
	filtered = append(filtered, group)
	events["PreToolUse"], err = json.Marshal(filtered)
	if err != nil {
		return nil, err
	}
	settings["hooks"], err = json.Marshal(events)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(settings, "", "  ")
}

func quoteCommand(executable string) string {
	return `"` + strings.ReplaceAll(executable, `"`, `\"`) + `"`
}
