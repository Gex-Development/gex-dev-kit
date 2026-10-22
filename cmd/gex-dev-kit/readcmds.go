package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Gex-Development/gex-dev-kit/internal/app"
	"github.com/Gex-Development/gex-dev-kit/internal/catalog"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/screens"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/terminal"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

const (
	readMattLicense    = "MIT"
	readRequiresMarker = " Requer seleção conjunta: "
	readSkillsSchema   = "gex-dev-kit.skills/v1"
	readHooksSchema    = "gex-dev-kit.hooks/v1"
	readAddonsSchema   = "gex-dev-kit.addons/v1"
	readDoctorSchema   = "gex-dev-kit.doctor/v1"
	readDefaultWidth   = 80
)

func cmdSkills(s session) int {
	doc := readSkillsDoc()
	if s.opts.JSON {
		return readJSON(s, doc, 0)
	}
	return readEmit(s, readSkillsText(doc), 0)
}

func cmdHooks(s session) int {
	doc := readHooksDoc()
	if s.opts.JSON {
		return readJSON(s, doc, 0)
	}
	return readEmit(s, readHooksText(doc), 0)
}

func cmdAddons(s session) int {
	addons, err := readAddonSelection(s.opts.Args)
	if err != nil {
		return invalid(s.errOut, err.Error())
	}
	doc := readAddonsDoc(addons)
	if s.opts.JSON {
		return readJSON(s, doc, 0)
	}
	return readEmit(s, readAddonsText(doc), 0)
}

func cmdDoctor(s session) int {
	mode, err := readMode(s)
	if err != nil {
		return invalid(s.errOut, err.Error())
	}
	report, err := readDoctor(s)
	if err != nil {
		return readFail(s.errOut, err)
	}
	code := report.ExitCode()
	if mode == terminal.ModeTUI {
		return readRunTUI(s, code, func(r screens.Runner, c screens.Common) error {
			return screens.RunDoctor(r, c, report)
		})
	}
	if s.opts.JSON {
		return readJSON(s, readDoctorDoc(readRoot(s), report, code), code)
	}
	return readEmit(s, readDoctorText(report), code)
}

func cmdShowcase(s session) int {
	mode, err := readMode(s)
	if err != nil {
		return invalid(s.errOut, err.Error())
	}
	if mode == terminal.ModeTUI {
		return readRunTUI(s, 0, func(r screens.Runner, c screens.Common) error {
			return screens.RunShowcase(r, c)
		})
	}
	return readEmit(s, screens.ShowcaseText(readTheme(s), readWidth(s)), 0)
}

func readMode(s session) (terminal.Mode, error) {
	if s.opts.JSON {
		if s.opts.UI == terminal.ModeTUI {
			return terminal.ModeCLI, errors.New("--json não combina com --ui=tui")
		}
		return terminal.ModeCLI, nil
	}
	return s.env.Resolve(s.opts.UI)
}

func readSkillsDoc() jsonSkillsDoc {
	doc := jsonSkillsDoc{
		Schema:      readSkillsSchema,
		KitVersion:  version,
		Core:        []jsonCoreSkill{},
		Specialties: []jsonSpecialty{},
		Matt:        jsonMatt{Revision: catalog.MattRevision, License: readMattLicense, Skills: []jsonMattSkill{}},
		References:  readReferences(catalog.ReferencedSkills()),
	}
	for _, name := range catalog.CoreSkills() {
		doc.Core = append(doc.Core, jsonCoreSkill{Name: name, Description: catalog.SkillDescription(name)})
	}
	for _, specialty := range catalog.Specialties() {
		doc.Specialties = append(doc.Specialties, readSpecialty(specialty))
	}
	for _, name := range catalog.Skills() {
		doc.Matt.Skills = append(doc.Matt.Skills, readMattSkill(name))
	}
	return doc
}

func readSpecialty(specialty catalog.Specialty) jsonSpecialty {
	return jsonSpecialty{
		Name:        specialty.Name,
		Group:       specialty.Group,
		Description: specialty.Description,
		Requires:    readList(specialty.Requires),
		Credentials: readList(specialty.Credentials),
		Warning:     specialty.Warning,
	}
}

func readMattSkill(name string) jsonMattSkill {
	description, requires := readSplitRequires(catalog.SkillDescription(name))
	source := catalog.MattSource(name)
	return jsonMattSkill{Name: name, Group: readMattGroup(source), Description: description, Requires: requires, Source: source}
}

func readSplitRequires(text string) (string, []string) {
	description, rest, found := strings.Cut(text, readRequiresMarker)
	if !found {
		return text, []string{}
	}
	return description, strings.Split(strings.TrimSuffix(rest, "."), ", ")
}

func readMattGroup(source string) string {
	_, rest, found := strings.Cut(source, "/skills/")
	if !found {
		return ""
	}
	group, _, _ := strings.Cut(rest, "/")
	return group
}

func readReferences(references []catalog.CatalogReference) []jsonReference {
	out := make([]jsonReference, 0, len(references))
	for _, reference := range references {
		out = append(out, jsonReference{Name: reference.Name, Status: reference.Status, Reason: readHomeTilde(reference.Reason), Source: reference.Source})
	}
	return out
}

func readHooksDoc() jsonHooksDoc {
	doc := jsonHooksDoc{
		Schema:     readHooksSchema,
		AgentHooks: []jsonAgentHook{},
		GitHooks:   []jsonGitHook{},
		References: readReferences(catalog.HookCandidates()),
	}
	for _, spec := range catalog.HookSpecs() {
		if spec.Scope == catalog.HookScopeGit {
			doc.GitHooks = append(doc.GitHooks, readGitHook(spec))
			continue
		}
		doc.AgentHooks = append(doc.AgentHooks, jsonAgentHook{
			Name:        spec.Name,
			Command:     spec.Command,
			Description: spec.Description,
			Source:      spec.Source,
			Scope:       spec.Policy,
			Claude:      readHookSupport(spec.Claude),
			Codex:       readHookSupport(spec.Codex),
		})
	}
	return doc
}

func readGitHook(spec catalog.HookSpec) jsonGitHook {
	return jsonGitHook{
		Name:             spec.Name,
		GitHook:          spec.GitHook,
		Description:      spec.Description,
		RequiresRepo:     true,
		RequiresCommands: spec.GitHook == "pre-commit" || spec.GitHook == "pre-push",
	}
}

func readHookSupport(support catalog.HookSupport) jsonHookSupport {
	return jsonHookSupport{
		Supported:     support.Supported,
		Event:         support.Event,
		Matcher:       support.Matcher,
		Decision:      support.Decision,
		Reason:        readHomeTilde(support.Reason),
		RequiresTrust: support.Trust,
	}
}

func readAddonSelection(args []string) ([]catalog.Addon, error) {
	switch len(args) {
	case 0:
		return catalog.Addons(), nil
	case 1:
		addon, ok := catalog.AddonByName(args[0])
		if !ok {
			return nil, fmt.Errorf("addon desconhecido: %s", args[0])
		}
		return []catalog.Addon{addon}, nil
	}
	return nil, errors.New("uso: gex-dev-kit addons [NOME]")
}

func readAddonsDoc(addons []catalog.Addon) jsonAddonsDoc {
	doc := jsonAddonsDoc{Schema: readAddonsSchema, Addons: []jsonAddon{}}
	for _, addon := range addons {
		doc.Addons = append(doc.Addons, jsonAddon{
			Name:          addon.Name,
			Kind:          addon.Kind,
			Summary:       addon.Summary,
			Source:        addon.Source,
			License:       addon.License,
			Revision:      addon.Revision,
			Release:       addon.Release,
			CheckedAt:     addon.CheckedAt,
			Install:       jsonInstall{Claude: addon.InstallClaude, Codex: addon.InstallCodex},
			Pinned:        addon.Pinned,
			Deterministic: addon.Deterministic,
			Installs:      readList(addon.Installs),
			Conflicts:     readList(addon.Conflicts),
			CoreCompat:    addon.CoreCompat,
		})
	}
	return doc
}

func readDoctor(s session) (app.Report, error) {
	home := readHome(s)
	executable := readExecutable()
	if s.opts.Repo != "" {
		return s.deps.DoctorRepo(home, s.opts.Repo, executable)
	}
	return s.deps.Doctor(home, s.env.Environ, executable)
}

func readDoctorDoc(root string, report app.Report, code int) jsonDoctorDoc {
	ok, attention, pending := report.Counts()
	doc := jsonDoctorDoc{
		Schema:   readDoctorSchema,
		Root:     root,
		OK:       code == 0,
		ExitCode: code,
		Counts:   jsonDoctorCounts{OK: ok, Attention: attention, Pending: pending},
		Items:    readDoctorItems(report.Items),
		Agents:   []jsonDoctorAgent{},
		Warnings: readDoctorItems(report.Warnings),
	}
	for _, agent := range report.Agents {
		doc.Agents = append(doc.Agents, readDoctorAgentDoc(agent))
	}
	return doc
}

func readDoctorItems(statuses []app.Status) []jsonDoctorItem {
	out := make([]jsonDoctorItem, 0, len(statuses))
	for _, status := range statuses {
		out = append(out, jsonDoctorItem{Path: status.Path, State: status.State, Level: app.StateLevel(status.State), Detail: readHomeTilde(status.Detail)})
	}
	return out
}

func readDoctorAgentDoc(agent app.AgentChange) jsonDoctorAgent {
	state := readAgentState(agent)
	doc := jsonDoctorAgent{
		Name:    agent.Name,
		Managed: agent.Current != "",
		Current: agent.Current,
		Target:  agent.Target,
		State:   state,
		Reason:  readHomeTilde(agent.Reason),
	}
	if state != "ok" {
		doc.Action = readHomeTilde(agent.Manual)
	}
	return doc
}

func readAgentState(agent app.AgentChange) string {
	switch {
	case agent.Blocked && agent.Current != "":
		return "bloqueado"
	case agent.Blocked:
		return "aviso"
	case agent.Current == "":
		return "pendente"
	}
	return "ok"
}

func readDoctorText(report app.Report) string {
	var b strings.Builder
	for _, item := range report.Items {
		readStatusLine(&b, item.State, item.Path, item.Detail)
	}
	for _, agent := range report.Agents {
		state := readAgentState(agent)
		readStatusLine(&b, state, agent.Name, agent.Reason)
		if state != "ok" && agent.Manual != "" {
			fmt.Fprintf(&b, "    Ação: %s\n", readHomeTilde(agent.Manual))
		}
	}
	if len(report.Warnings) > 0 {
		b.WriteString("Avisos:\n")
	}
	for _, warning := range report.Warnings {
		readStatusLine(&b, warning.State, warning.Path, warning.Detail)
	}
	return b.String()
}

func readStatusLine(b *strings.Builder, state, name, detail string) {
	fmt.Fprintf(b, "%s %s", state, name)
	if detail != "" {
		fmt.Fprintf(b, " (%s)", readHomeTilde(detail))
	}
	b.WriteString("\n")
}

func readSkillsText(doc jsonSkillsDoc) string {
	var b strings.Builder
	b.WriteString("Core GEX (obrigatório):\n")
	for _, skill := range doc.Core {
		readLine(&b, "  ", skill.Name, skill.Description)
	}
	b.WriteString("\nEspecialidades GEX (opcionais, por grupo):\n")
	for _, group := range catalog.SpecialtyGroups() {
		fmt.Fprintf(&b, "  %s\n", group)
		for _, specialty := range doc.Specialties {
			if specialty.Group == group {
				readSpecialtyText(&b, specialty)
			}
		}
	}
	fmt.Fprintf(&b, "\nSkills Matt Pocock (revisão %s, licença %s):\n", readShortRevision(doc.Matt.Revision), doc.Matt.License)
	for _, skill := range doc.Matt.Skills {
		readLine(&b, "  ", skill.Group+"/"+skill.Name, skill.Description)
		readRequires(&b, skill.Requires)
	}
	b.WriteString("\nReferências (não instaláveis):\n")
	readReferenceLines(&b, doc.References)
	return b.String()
}

func readSpecialtyText(b *strings.Builder, specialty jsonSpecialty) {
	readLine(b, "    ", specialty.Name, specialty.Description)
	readRequires(b, specialty.Requires)
	if len(specialty.Credentials) > 0 {
		fmt.Fprintf(b, "      Credenciais: %s\n", strings.Join(specialty.Credentials, ", "))
	}
	if specialty.Warning != "" {
		fmt.Fprintf(b, "      Aviso: %s\n", specialty.Warning)
	}
}

func readRequires(b *strings.Builder, requires []string) {
	if len(requires) > 0 {
		fmt.Fprintf(b, "      Requer: %s\n", strings.Join(requires, ", "))
	}
}

func readShortRevision(revision string) string {
	if len(revision) > 8 {
		return revision[:8]
	}
	return revision
}

func readHooksText(doc jsonHooksDoc) string {
	var b strings.Builder
	b.WriteString("Hooks de agente (Claude Code e Codex, instalados com install):\n")
	for _, hook := range doc.AgentHooks {
		readLine(&b, "  ", hook.Name, hook.Description)
		fmt.Fprintf(&b, "    Claude: %s\n", readSupportText(hook.Claude))
		fmt.Fprintf(&b, "    Codex: %s\n", readSupportText(hook.Codex))
	}
	b.WriteString("\nHooks de Git (gravados por repositório com githooks --repo):\n")
	for _, hook := range doc.GitHooks {
		readLine(&b, "  ", hook.Name+" ("+hook.GitHook+")", hook.Description)
	}
	b.WriteString("\nReferências (não instaláveis):\n")
	readReferenceLines(&b, doc.References)
	return b.String()
}

func readSupportText(support jsonHookSupport) string {
	if !support.Supported {
		return "sem suporte: " + support.Reason
	}
	text := strings.Join([]string{support.Event, support.Matcher, support.Decision}, ", ")
	if support.RequiresTrust {
		text += "; exige aprovação em /hooks"
	}
	return text
}

func readAddonsText(doc jsonAddonsDoc) string {
	var b strings.Builder
	b.WriteString("Addons (opt-in; o kit mostra o comando oficial e nunca o executa):\n")
	for _, addon := range doc.Addons {
		fmt.Fprintf(&b, "\n%s [%s]  %s\n", addon.Name, addon.Kind, addon.Summary)
		fmt.Fprintf(&b, "  Fonte: %s (licença %s)\n", addon.Source, addon.License)
		fmt.Fprintf(&b, "  Versão: %s, revisão %s, conferido em %s\n", addon.Release, addon.Revision, addon.CheckedAt)
		fmt.Fprintf(&b, "  Fixação: %s\n", addon.Pinned)
		readCommand(&b, "Instalar no Claude Code", addon.Install.Claude)
		readCommand(&b, "Instalar no Codex", addon.Install.Codex)
		readBullets(&b, "O que instala", addon.Installs)
		readBullets(&b, "Conflitos", addon.Conflicts)
		fmt.Fprintf(&b, "  Compatibilidade com a Core: %s\n", addon.CoreCompat)
	}
	return b.String()
}

func readCommand(b *strings.Builder, label, command string) {
	fmt.Fprintf(b, "  %s:\n", label)
	for _, line := range strings.Split(command, "\n") {
		fmt.Fprintf(b, "    %s\n", line)
	}
}

func readBullets(b *strings.Builder, label string, items []string) {
	fmt.Fprintf(b, "  %s:\n", label)
	for _, item := range items {
		fmt.Fprintf(b, "    - %s\n", item)
	}
}

func readLine(b *strings.Builder, indent, name, text string) {
	b.WriteString(indent + name)
	if text != "" {
		b.WriteString("  " + text)
	}
	b.WriteString("\n")
}

func readReferenceLines(b *strings.Builder, references []jsonReference) {
	for _, reference := range references {
		readLine(b, "  ", reference.Name+" ["+reference.Status+"]", reference.Reason)
	}
}

func readTheme(s session) theme.Theme {
	return theme.New(theme.Options{Plain: true, Profile: s.env.ColorProfile()})
}

func readWidth(s session) int {
	if s.env.Width > 0 {
		return s.env.Width
	}
	return readDefaultWidth
}

func readRoot(s session) string {
	if s.opts.Repo != "" {
		return "repo"
	}
	return "home"
}

func readHome(s session) string {
	if s.opts.Home != "" {
		return s.opts.Home
	}
	home, _ := os.UserHomeDir()
	return home
}

func readExecutable() string {
	executable, err := os.Executable()
	if err != nil {
		return ""
	}
	return executable
}

func readCommon(s session) screens.Common {
	userHome, _ := os.UserHomeDir()
	return screens.Common{
		Version:  version,
		Home:     readHome(s),
		UserHome: userHome,
		Profile:  s.env.ColorProfile(),
		Motion:   s.env.Motion(s.opts.NoAnim),
	}
}

func readRunTUI(s session, code int, run func(screens.Runner, screens.Common) error) int {
	signals, stop := screens.NotifySignals()
	defer stop()
	runner := screens.Runner{
		In:      s.env.InFile,
		Out:     s.env.OutFile,
		Environ: s.env.Environ,
		Profile: s.env.ColorProfile(),
		Signals: signals,
	}
	if err := run(runner, readCommon(s)); err != nil {
		if exit, ok := screens.ExitCode(err); ok {
			return exit
		}
		return readFail(s.errOut, err)
	}
	return code
}

func readHomeTilde(text string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return text
	}
	if text == home {
		return "~"
	}
	return strings.ReplaceAll(text, home+string(filepath.Separator), "~"+string(filepath.Separator))
}

func readList(items []string) []string {
	return append([]string{}, items...)
}

func readEmit(s session, text string, code int) int {
	if _, err := io.WriteString(s.env.Writer(s.out), text); err != nil {
		return readFail(s.errOut, err)
	}
	return code
}

func readJSON(s session, value any, code int) int {
	if err := writeJSON(s.out, value); err != nil {
		return readFail(s.errOut, err)
	}
	return code
}

func readFail(errOut io.Writer, err error) int {
	return writeFail(errOut, err)
}
