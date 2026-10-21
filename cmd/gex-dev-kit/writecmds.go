package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Gex-Development/gex-dev-kit/internal/app"
	"github.com/Gex-Development/gex-dev-kit/internal/githooks"
	"github.com/Gex-Development/gex-dev-kit/internal/kit"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/components"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/screens"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/terminal"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

func cmdInstall(s session) int {
	mode, err := writeMode(s)
	if err != nil {
		return invalid(s.errOut, err.Error())
	}
	req, err := writeRequest(s)
	if err != nil {
		return writeFail(s.errOut, err)
	}
	if mode == terminal.ModeTUI {
		return writeInstallTUI(s, req)
	}
	return writeInstallCLI(s, req)
}

func writeInstallCLI(s session, req app.Request) int {
	plan, err := s.deps.PlanInstall(req, app.DiffLinesCLI)
	if err != nil {
		return writeFail(s.errOut, err)
	}
	fmt.Fprintln(s.out, "Prévia de instalação:")
	writePreview(s.out, writeTheme(s), writeGroups(plan.Diffs))
	writeChangeLines(s.out, plan.Files)
	writeAgentLines(s.out, plan.Agents)
	writeNotices(s.errOut, plan.Notices)
	if plan.Empty() {
		fmt.Fprintln(s.out, "Nada a alterar.")
		if plan.Blocked {
			fmt.Fprintln(s.errOut, "Atualização de agentes bloqueada; Core inalterado.")
			return 1
		}
		return 0
	}
	if ok, code := writeConfirm(s, "Aplicar alterações?"); !ok {
		return code
	}
	var res app.Result
	err = s.guard.Write(func() error {
		var applyErr error
		res, applyErr = s.deps.ApplyInstall(plan, writeProgress(s.errOut, s.opts.Quiet))
		return applyErr
	})
	if err != nil {
		return writeFail(s.errOut, err)
	}
	return writeInstallResult(s, res)
}

func writeInstallTUI(s session, req app.Request) int {
	r, stop := writeRunner(s)
	defer stop()
	res, err := screens.RunInstall(r, writeCommon(s), screens.InstallDeps{
		Options: s.deps.InstallOptions,
		Plan:    func(q app.Request) (app.Plan, error) { return s.deps.PlanInstall(q, app.DiffLinesTUI) },
		Apply:   s.deps.ApplyInstall,
	}, req, true)
	if err != nil {
		return writeScreenExit(s.errOut, err)
	}
	return writeInstallResult(s, res)
}

func writeInstallResult(s session, res app.Result) int {
	if res.AgentsUpdated {
		fmt.Fprintln(s.out, "Agentes atualizados.")
	}
	if res.HistoryID != "" {
		fmt.Fprintln(s.out, "Instalado. Histórico:", res.HistoryID)
	}
	if res.HooksNeedTrust {
		fmt.Fprintln(s.out, "Codex: abra /hooks e aprove o hook novo para ativá-lo.")
	}
	if res.Blocked {
		fmt.Fprintln(s.errOut, "Agentes bloqueados não foram atualizados; demais alterações aplicadas.")
		return 1
	}
	return 0
}

func cmdRollback(s session) int {
	mode, err := writeMode(s)
	if err != nil {
		return invalid(s.errOut, err.Error())
	}
	root, err := writeRollbackRoot(s.opts)
	if err != nil {
		return writeFail(s.errOut, err)
	}
	if mode == terminal.ModeTUI {
		return writeRollbackTUI(s, root)
	}
	return writeRollbackCLI(s, root)
}

func writeRollbackCLI(s session, root string) int {
	prev, err := s.deps.PlanRollback(root, app.DiffLinesCLI)
	if err != nil {
		return writeFail(s.errOut, err)
	}
	if prev.Plan.ID == "" {
		fmt.Fprintln(s.out, "Nada a restaurar.")
		return 0
	}
	if len(prev.Plan.Conflicts) > 0 {
		return writeFail(s.errOut, fmt.Errorf("%w: %s", kit.ErrConflict, strings.Join(prev.Plan.Conflicts, ", ")))
	}
	fmt.Fprintln(s.out, "Prévia de restauração:")
	writePreview(s.out, writeTheme(s), writeGroups(prev.Diffs))
	writeRollbackLines(s.out, prev.Plan.Entries)
	if ok, code := writeConfirm(s, "Restaurar arquivos?"); !ok {
		return code
	}
	var id string
	err = s.guard.Write(func() error {
		var rollbackErr error
		id, rollbackErr = s.deps.Rollback(root)
		return rollbackErr
	})
	if err != nil {
		return writeFail(s.errOut, err)
	}
	if id == "" {
		fmt.Fprintln(s.out, "Nada a restaurar.")
		return 0
	}
	fmt.Fprintln(s.out, "Restaurado. Histórico:", id)
	return 0
}

func writeRollbackTUI(s session, root string) int {
	kind := screens.RollbackCore
	if s.opts.Repo != "" {
		kind = screens.RollbackRepo
	}
	r, stop := writeRunner(s)
	defer stop()
	id, err := screens.RunRollback(r, writeCommon(s), kind, screens.RollbackDeps{
		Plan:  func() (app.RollbackPreview, error) { return s.deps.PlanRollback(root, app.DiffLinesTUI) },
		Apply: func() (string, error) { return s.deps.Rollback(root) },
	})
	if err != nil {
		return writeScreenExit(s.errOut, err)
	}
	if id != "" {
		fmt.Fprintln(s.out, "Restaurado. Histórico:", id)
	}
	return 0
}

func cmdRollbackAgents(s session) int {
	mode, err := writeMode(s)
	if err != nil {
		return invalid(s.errOut, err.Error())
	}
	home, err := writeHome(s.opts.Home)
	if err != nil {
		return writeFail(s.errOut, err)
	}
	if mode == terminal.ModeTUI {
		return writeRollbackAgentsTUI(s, home)
	}
	fmt.Fprintln(s.out, "Rollback dos agentes: restaura somente versões anteriores gerenciadas pelo kit; não altera o Core.")
	if ok, code := writeConfirm(s, "Restaurar versões anteriores dos agentes?"); !ok {
		return code
	}
	err = s.guard.Write(func() error { return s.deps.RollbackAgents(home) })
	if err != nil {
		return writeFail(s.errOut, err)
	}
	fmt.Fprintln(s.out, "Agentes restaurados.")
	return 0
}

func writeRollbackAgentsTUI(s session, home string) int {
	r, stop := writeRunner(s)
	defer stop()
	_, err := screens.RunRollback(r, writeCommon(s), screens.RollbackAgents, screens.RollbackDeps{
		Plan:  func() (app.RollbackPreview, error) { return app.RollbackPreview{}, nil },
		Apply: func() (string, error) { return "", s.deps.RollbackAgents(home) },
	})
	if err != nil {
		return writeScreenExit(s.errOut, err)
	}
	return 0
}

func cmdGitHooks(s session) int {
	mode, err := writeMode(s)
	if err != nil {
		return invalid(s.errOut, err.Error())
	}
	home, err := writeHome(s.opts.Home)
	if err != nil {
		return writeFail(s.errOut, err)
	}
	executable, err := writeExecutable()
	if err != nil {
		return writeFail(s.errOut, err)
	}
	if s.opts.Repo != "" {
		if _, err := writeAbs(s.opts.Repo); err != nil {
			return writeFail(s.errOut, err)
		}
	}
	req := app.GitHooksRequest{
		Home: home, Repo: s.opts.Repo, Executable: executable,
		Hooks: s.opts.Hooks, PreCommit: s.opts.PreCommit, PrePush: s.opts.PrePush,
	}
	if mode == terminal.ModeTUI {
		return writeGitHooksTUI(s, req)
	}
	if s.opts.Repo == "" {
		return writeHookScripts(s, req)
	}
	return writeGitHooksCLI(s, req)
}

func writeHookScripts(s session, req app.GitHooksRequest) int {
	hooks := make([]string, 0, len(req.Hooks))
	for _, name := range req.Hooks {
		hooks = append(hooks, strings.TrimPrefix(name, "gex-"))
	}
	files, err := githooks.Files(githooks.Request{Hooks: hooks, PreCommit: req.PreCommit, PrePush: req.PrePush, Binary: req.Executable})
	if err != nil {
		return writeFail(s.errOut, err)
	}
	fmt.Fprintln(s.out, "Scripts dos git hooks (sem --repo; nada foi gravado):")
	for _, f := range files {
		fmt.Fprintf(s.out, "%s\n%s\n", f.Path, f.Data)
	}
	return 0
}

func writeGitHooksCLI(s session, req app.GitHooksRequest) int {
	plan, err := s.deps.PlanGitHooks(req, app.DiffLinesCLI)
	if err != nil {
		return writeFail(s.errOut, err)
	}
	fmt.Fprintln(s.out, "Prévia dos git hooks:")
	writePreview(s.out, writeTheme(s), writeGroups(plan.Diffs))
	writeChangeLines(s.out, plan.HomeChanges)
	writeChangeLines(s.out, plan.RepoChanges)
	writeSourceLines(s.out, plan.Sources)
	if len(plan.HomeChanges)+len(plan.RepoChanges) == 0 {
		fmt.Fprintln(s.out, "Nada a alterar.")
		return 0
	}
	if ok, code := writeConfirm(s, "Gravar git hooks neste repositório?"); !ok {
		return code
	}
	var res app.Result
	err = s.guard.Write(func() error {
		var applyErr error
		res, applyErr = s.deps.ApplyGitHooks(plan)
		return applyErr
	})
	if err != nil {
		return writeFail(s.errOut, err)
	}
	if res.HistoryID != "" {
		fmt.Fprintln(s.out, "Git hooks gravados. Histórico:", res.HistoryID)
	}
	return 0
}

func writeGitHooksTUI(s session, req app.GitHooksRequest) int {
	r, stop := writeRunner(s)
	defer stop()
	res, err := screens.RunGitHooks(r, writeCommon(s), screens.GitHooksDeps{
		Plan:  func(q app.GitHooksRequest) (app.GitHooksPlan, error) { return s.deps.PlanGitHooks(q, app.DiffLinesTUI) },
		Apply: s.deps.ApplyGitHooks,
	}, req)
	if err != nil {
		return writeScreenExit(s.errOut, err)
	}
	if res.HistoryID != "" {
		fmt.Fprintln(s.out, "Git hooks gravados. Histórico:", res.HistoryID)
	}
	return 0
}

func writeMode(s session) (terminal.Mode, error) {
	scriptable := s.opts.Yes || s.opts.Quiet || s.opts.JSON
	if s.opts.UI == terminal.ModeTUI && scriptable {
		return terminal.ModeAuto, errors.New("--ui=tui não combina com --yes, --quiet ou --json; use --ui=cli")
	}
	if s.opts.UI == terminal.ModeAuto && scriptable {
		return terminal.ModeCLI, nil
	}
	return s.env.Resolve(s.opts.UI)
}

func writeRequest(s session) (app.Request, error) {
	home, err := writeHome(s.opts.Home)
	if err != nil {
		return app.Request{}, err
	}
	executable, err := writeExecutable()
	if err != nil {
		return app.Request{}, err
	}
	return app.Request{Home: home, Executable: executable, Skills: s.opts.Skills, Hooks: s.opts.Hooks, SkipAgentUpdate: s.opts.SkipAgentUpdate}, nil
}

func writeRollbackRoot(o options) (string, error) {
	if o.Repo != "" {
		return writeAbs(o.Repo)
	}
	return writeHome(o.Home)
}

func writeHome(path string) (string, error) {
	if path == "" {
		return os.UserHomeDir()
	}
	return writeAbs(path)
}

func writeAbs(path string) (string, error) {
	if !filepath.IsAbs(path) {
		return "", errors.New("caminho deve ser absoluto")
	}
	return filepath.Clean(path), nil
}

func writeExecutable() (string, error) {
	return os.Executable()
}

func writeConfirm(s session, question string) (bool, int) {
	if s.opts.Yes {
		return true, 0
	}
	if !s.env.Interactive() {
		fmt.Fprintln(s.errOut, "Confirmação exige terminal ou --yes; nenhuma alteração aplicada.")
		return false, 1
	}
	ok, err := promptConfirm(newLineReader(s.in), s.out, question)
	if err != nil {
		return false, writeFail(s.errOut, err)
	}
	if !ok {
		fmt.Fprintln(s.out, "Nada foi alterado.")
		return false, 0
	}
	return true, 0
}

func writeRunner(s session) (screens.Runner, func()) {
	signals, stop := screens.NotifySignals()
	return screens.Runner{
		In:      s.env.InFile,
		Out:     s.env.OutFile,
		Environ: s.env.Environ,
		Profile: s.env.Profile,
		Signals: signals,
	}, stop
}

func writeCommon(s session) screens.Common {
	userHome, _ := os.UserHomeDir()
	return screens.Common{
		Version:  version,
		Home:     s.opts.Home,
		UserHome: userHome,
		Profile:  s.env.Profile,
		Motion:   s.env.Motion(s.opts.NoAnim),
	}
}

func writeScreenExit(errOut io.Writer, err error) int {
	if code, ok := screens.ExitCode(err); ok {
		return code
	}
	return writeFail(errOut, err)
}

func writeFail(errOut io.Writer, err error) int {
	p := app.Explain(err)
	fmt.Fprintf(errOut, "%s: %s\n%s\n", p.Title, p.Cause, p.Action)
	return 1
}

func writeTheme(s session) theme.Theme {
	return theme.New(theme.Options{Profile: s.env.ColorProfile(), Plain: !s.env.StdoutTTY})
}

func writeGroups(diffs []app.FileDiff) []components.ChangeGroup {
	if len(diffs) == 0 {
		return nil
	}
	changes := make([]components.Change, 0, len(diffs))
	for _, d := range diffs {
		changes = append(changes, components.Change{
			Action:  writeAction(d.Action),
			Path:    d.Path,
			Detail:  d.Summary,
			Lines:   writeLines(d.Lines),
			Omitted: d.Omitted,
			Masked:  d.Masked,
		})
	}
	return []components.ChangeGroup{{Title: "Arquivos", Changes: changes}}
}

func writeAction(action string) components.ChangeAction {
	switch action {
	case "criar":
		return components.ChangeCreate
	case "remover":
		return components.ChangeRemove
	case "restaurar":
		return components.ChangeRestore
	case "inalterado":
		return components.ChangeUnchanged
	}
	return components.ChangeUpdate
}

func writeLines(lines []app.DiffLine) []components.DiffLine {
	out := make([]components.DiffLine, 0, len(lines))
	for _, l := range lines {
		out = append(out, components.DiffLine{Kind: writeDiffKind(l.Kind), Text: l.Text})
	}
	return out
}

func writeDiffKind(kind byte) components.DiffLineKind {
	switch kind {
	case '+':
		return components.DiffAdded
	case '-':
		return components.DiffRemoved
	case '@':
		return components.DiffHunk
	}
	return components.DiffContext
}

func writeChangeLines(w io.Writer, changes []kit.Change) {
	for _, c := range changes {
		fmt.Fprintf(w, "%s %s\n", c.Action, c.Path)
	}
}

func writeRollbackLines(w io.Writer, entries []kit.RollbackEntry) {
	for _, e := range entries {
		fmt.Fprintf(w, "%s %s\n", e.Action, e.Path)
	}
}

func writeAgentLines(w io.Writer, agents []app.AgentChange) {
	for _, a := range agents {
		switch {
		case a.Blocked:
			fmt.Fprintf(w, "bloqueado %s: %s\n", a.Name, a.Reason)
		case a.Current != a.Target:
			fmt.Fprintf(w, "atualizar %s %s -> %s\n", a.Name, a.Current, a.Target)
		}
	}
}

func writeSourceLines(w io.Writer, sources []app.CommandSource) {
	for _, src := range sources {
		fmt.Fprintf(w, "origem %s:%d: %s\n", src.File, src.Line, src.Command)
	}
}

func writeNotices(w io.Writer, notices []app.Notice) {
	for _, n := range notices {
		fmt.Fprintf(w, "%s: %s\n", n.Level, n.Text)
	}
}
