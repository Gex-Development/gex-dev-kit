package app

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Gex-Development/gex-dev-kit/internal/agents"
	"github.com/Gex-Development/gex-dev-kit/internal/catalog"
	"github.com/Gex-Development/gex-dev-kit/internal/githooks"
	"github.com/Gex-Development/gex-dev-kit/internal/hooks"
	"github.com/Gex-Development/gex-dev-kit/internal/kit"
)

var doctorOrgLine = regexp.MustCompile(`^[A-Za-z0-9-]{1,39}$`)

func Doctor(home string, environ []string, executable string) (Report, error) {
	statuses, err := kit.Doctor(home)
	if err != nil {
		return Report{}, err
	}
	changes, err := agents.Preview(home)
	if err != nil {
		return Report{}, err
	}
	stable, err := doctorStableState(home, executable)
	if err != nil {
		return Report{}, err
	}
	report := Report{Root: home, Items: statuses, Agents: changes, Warnings: doctorAgentWarnings(changes)}
	report.Items = append(report.Items, doctorAgentItems(changes)...)
	if stable == "diferente" {
		report.Items = append(report.Items, doctorStableItem(home))
	}
	report.Warnings = append(report.Warnings, doctorEnvWarnings(home, environ)...)
	report.Warnings = append(report.Warnings, doctorScopeWarnings()...)
	return report, nil
}

func DoctorRepo(home, repo, executable string) (Report, error) {
	target, err := githooks.Resolve(repo, hooks.RunGit)
	hooksPath := errors.Is(err, githooks.ErrHooksPath)
	if err != nil && !hooksPath {
		return Report{}, err
	}
	if hooksPath {
		target, err = doctorCommonDir(repo)
		if err != nil {
			return Report{}, err
		}
	}
	statuses, err := kit.Doctor(target.CommonDir)
	if err != nil {
		return Report{}, err
	}
	report := Report{Root: target.CommonDir, Items: statuses}
	stable, err := doctorStableState(home, executable)
	if err != nil {
		return Report{}, err
	}
	switch stable {
	case "ausente":
		report.Warnings = append(report.Warnings, kit.Status{Path: doctorStableRel(home), State: "aviso", Detail: "executável do HOME sumiu; os hooks do repositório não funcionam até rodar gex-dev-kit install"})
	case "diferente":
		report.Warnings = append(report.Warnings, kit.Status{Path: doctorStableRel(home), State: "aviso", Detail: "executável do HOME difere do kit em uso; rode gex-dev-kit install"})
	}
	if hooksPath {
		report.Warnings = append(report.Warnings, kit.Status{Path: "core.hooksPath", State: "aviso", Detail: "core.hooksPath definido; os hooks do kit em hooks/ não são executados pelo Git"})
	}
	return report, nil
}

func doctorCommonDir(repo string) (githooks.Repo, error) {
	out, err := hooks.RunGit(repo, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return githooks.Repo{}, githooks.ErrNotRepo
	}
	return githooks.Repo{CommonDir: githooks.NormalizeGitPath(out)}, nil
}

func (r Report) Counts() (ok, attention, pending int) {
	for _, item := range r.Items {
		switch StateLevel(item.State) {
		case "ok":
			ok++
		case "atencao":
			attention++
		case "pendente":
			pending++
		}
	}
	return ok, attention, pending
}

func StateLevel(state string) string {
	switch state {
	case "ok":
		return "ok"
	case "pendente", "sem-historico":
		return "pendente"
	case "aviso":
		return "info"
	}
	return "atencao"
}

func CheckRepo(path string) error {
	_, err := githooks.Resolve(path, hooks.RunGit)
	if errors.Is(err, githooks.ErrHooksPath) {
		return nil
	}
	return err
}

func (r Report) ExitCode() int {
	for _, item := range r.Items {
		switch item.State {
		case "ausente", "alterado", "incompleto", "bloqueado":
			return 1
		}
	}
	return 0
}

func doctorStableState(home, executable string) (string, error) {
	data, err := os.ReadFile(catalog.StableBinary(home))
	if errors.Is(err, fs.ErrNotExist) {
		return "ausente", nil
	}
	if err != nil {
		return "", err
	}
	if executable == "" {
		return "", nil
	}
	current, err := os.ReadFile(executable)
	if err != nil {
		return "", fmt.Errorf("não foi possível ler executável: %w", err)
	}
	if !bytes.Equal(data, current) {
		return "diferente", nil
	}
	return "", nil
}

func doctorStableRel(home string) string {
	rel, err := filepath.Rel(home, catalog.StableBinary(home))
	if err != nil {
		return catalog.StableBinary(home)
	}
	return filepath.ToSlash(rel)
}

func doctorStableItem(home string) kit.Status {
	return kit.Status{Path: doctorStableRel(home), State: "atualizar", Detail: "executável dos hooks difere do kit em uso; rode `gex-dev-kit install`"}
}

func doctorAgentItems(changes []agents.Change) []kit.Status {
	var out []kit.Status
	for _, c := range changes {
		if c.Current == "" {
			continue
		}
		item := kit.Status{Path: "agente " + c.Name, State: "ok", Detail: "Versão " + c.Current + " registrada no histórico."}
		switch {
		case c.Blocked:
			item.State = "bloqueado"
			item.Detail = c.Reason
		case c.Current != c.Target:
			item.State = "atualizar"
			item.Detail = "Nova versão disponível: " + c.Target
		}
		out = append(out, item)
	}
	return out
}

func doctorAgentWarnings(changes []agents.Change) []kit.Status {
	var out []kit.Status
	for _, c := range changes {
		if c.Current == "" && c.Blocked {
			out = append(out, kit.Status{Path: "agente " + c.Name, State: "aviso", Detail: c.Reason})
		}
	}
	return out
}

func doctorEnvWarnings(home string, environ []string) []kit.Status {
	var out []kit.Status
	for _, name := range []string{"CLAUDE_CONFIG_DIR", "CODEX_HOME"} {
		value := doctorEnvValue(environ, name)
		if value != "" && doctorOutsideHome(home, value) {
			out = append(out, kit.Status{Path: name, State: "aviso", Detail: "aponta para fora do HOME; o kit não gerencia esta pasta"})
		}
	}
	return out
}

func doctorEnvValue(environ []string, name string) string {
	for _, entry := range environ {
		if value, ok := strings.CutPrefix(entry, name+"="); ok {
			return value
		}
	}
	return ""
}

func doctorOutsideHome(home, value string) bool {
	rel, err := filepath.Rel(home, value)
	if err != nil {
		return true
	}
	return rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func doctorScopeWarnings() []kit.Status {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil
	}
	scope := hooks.NewScope(configDir, nil)
	out := []kit.Status{{Path: "escopo GEX", State: "aviso", Detail: "organizações efetivas: " + strings.Join(scope.Orgs, ", ")}}
	for _, line := range doctorInvalidOrgLines(filepath.Join(configDir, "gex-dev-kit", "orgs")) {
		out = append(out, kit.Status{Path: "escopo GEX", State: "aviso", Detail: fmt.Sprintf("linha %d do arquivo de organizações é inválida e foi ignorada", line)})
	}
	return out
}

func doctorInvalidOrgLines(path string) []int {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var bad []int
	for i, line := range strings.Split(string(data), "\n") {
		text := strings.TrimSpace(line)
		if text != "" && !doctorOrgLine.MatchString(text) {
			bad = append(bad, i+1)
		}
	}
	return bad
}
