package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/agents"
	"github.com/Gex-Development/gex-dev-kit/internal/catalog"
	"github.com/Gex-Development/gex-dev-kit/internal/hooks"
	"github.com/Gex-Development/gex-dev-kit/internal/kit"
	"github.com/charmbracelet/x/term"
)

var version = "dev"

type names []string

func (n *names) String() string { return strings.Join(*n, ", ") }
func (n *names) Set(value string) error {
	if value == "" {
		return errors.New("nome vazio")
	}
	*n = append(*n, value)
	return nil
}

const usage = "gex-dev-kit [install|doctor|rollback|rollback-agents|skills|hooks|hook commit-msg] [opções]\n\n" +
	"Sem argumentos em terminal: painel interativo para instalar, diagnosticar e reverter.\n\n" +
	"install         Instala Core GEX e atualiza agentes em versões fixadas, após prévia e confirmação\n" +
	"doctor          Exibe estado dos arquivos e agentes gerenciados\n" +
	"rollback        Restaura arquivos da instalação anterior\n" +
	"rollback-agents Restaura versões anteriores dos agentes, com confirmação\n" +
	"skills          Lista Core, opções instaláveis e referências não instaláveis\n" +
	"hooks           Lista hook global instalável e referências não instaláveis\n" +
	"--version       Exibe versão\n\n" +
	"Opções de install: --skill NOME (repetível), --hook NOME (repetível), --skip-agent-update, --yes, --home CAMINHO\n" +
	"Opções de doctor/rollback: --home CAMINHO\n" +
	"Opções de rollback-agents: --home CAMINHO, --yes\n"

func main() { os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)) }

func commandHelp(command string) string {
	switch command {
	case "install":
		return "Uso: gex-dev-kit install [--skill NOME] [--hook NOME] [--skip-agent-update] [--home CAMINHO] [--yes]\nCore obrigatório; mostra prévia e pede confirmação antes de instalar.\n"
	case "doctor":
		return "Uso: gex-dev-kit doctor [--home CAMINHO]\nMostra integridade do Core e versões dos agentes.\n"
	case "rollback":
		return "Uso: gex-dev-kit rollback [--home CAMINHO]\nRestaura arquivos Core da última instalação.\n"
	case "rollback-agents":
		return "Uso: gex-dev-kit rollback-agents [--home CAMINHO] [--yes]\nRestaura versões anteriores dos agentes separadamente.\n"
	case "skills":
		return "Uso: gex-dev-kit skills\nLista Core, especializações, skills Matt e referências não instaláveis.\n"
	case "hooks":
		return "Uso: gex-dev-kit hooks\nLista hook global instalável e candidatos não instaláveis.\n"
	case "hook":
		return "Uso: gex-dev-kit hook commit-msg [CAMINHO]\nValida JSON do agente em stdin ou mensagem Git no arquivo informado.\n"
	default:
		return usage
	}
}

func helpRequested(args []string) bool {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return true
		}
	}
	return false
}

func run(args []string, in io.Reader, out, errOut io.Writer) int {
	if len(args) == 0 {
		if !interactive(in, out) {
			fmt.Fprint(out, usage)
			return 0
		}
		var choice string
		var err error
		if plainTerminal() {
			choice, err = showAccessibleHome(in, out)
		} else {
			destination, homeErr := homePath("")
			if homeErr != nil {
				fmt.Fprintln(errOut, homeErr)
				return 1
			}
			choice, err = showHome(in, out, destination)
		}
		if err != nil {
			fmt.Fprintln(errOut, "Painel:", err)
			return 1
		}
		switch choice {
		case "install":
			return install(nil, in, out, errOut, true)
		case "doctor":
			return inspect("doctor", nil, in, out, errOut)
		case "rollback":
			return inspect("rollback", nil, in, out, errOut)
		case "rollback-agents":
			return rollbackAgents(nil, in, out, errOut)
		default:
			return 0
		}
	}
	if len(args) > 1 && helpRequested(args[1:]) {
		switch args[0] {
		case "install", "doctor", "rollback", "rollback-agents", "skills", "hooks", "hook":
			fmt.Fprint(out, commandHelp(args[0]))
			return 0
		}
	}
	switch args[0] {
	case "--help", "-h", "help":
		if len(args) == 2 {
			fmt.Fprint(out, commandHelp(args[1]))
		} else {
			fmt.Fprint(out, usage)
		}
		return 0
	case "--version", "-v", "version":
		fmt.Fprintln(out, version)
		return 0
	case "skills":
		if len(args) != 1 {
			return invalid(errOut, "skills não aceita argumentos")
		}
		fmt.Fprintln(out, "Core GEX (obrigatório):")
		for _, name := range catalog.CoreSkills() {
			fmt.Fprintln(out, name)
		}
		fmt.Fprintln(out, "Especializações opcionais:")
		for _, name := range catalog.SpecialtySkills() {
			fmt.Fprintln(out, name)
		}
		fmt.Fprintln(out, "Skills Matt opcionais:")
		for _, name := range catalog.Skills() {
			fmt.Fprintln(out, name)
		}
		fmt.Fprintln(out, "Referências (não instaláveis):")
		for _, entry := range catalog.ReferencedSkills() {
			fmt.Fprintf(out, "%s [%s] — %s (%s)\n", entry.Name, entry.Status, entry.Reason, entry.Source)
		}
		return 0
	case "hooks":
		if len(args) != 1 {
			return invalid(errOut, "hooks não aceita argumentos")
		}
		fmt.Fprintln(out, "Hook global opcional (instalável):")
		for _, name := range catalog.Hooks() {
			fmt.Fprintln(out, name)
		}
		fmt.Fprintln(out, "Referências (não instaláveis):")
		for _, entry := range catalog.HookCandidates() {
			fmt.Fprintf(out, "%s [%s] — %s (%s)\n", entry.Name, entry.Status, entry.Reason, entry.Source)
		}
		return 0
	case "install":
		return install(args[1:], in, out, errOut, false)
	case "doctor", "rollback":
		return inspect(args[0], args[1:], in, out, errOut)
	case "rollback-agents":
		return rollbackAgents(args[1:], in, out, errOut)
	case "hook":
		if len(args) < 2 || len(args) > 3 || args[1] != "commit-msg" {
			return invalid(errOut, "uso: hook commit-msg [CAMINHO]")
		}
		var err error
		if len(args) == 3 {
			var message []byte
			message, err = os.ReadFile(args[2])
			if err == nil {
				err = hooks.CheckGitMessage(string(message))
			}
		} else {
			err = hooks.CheckAgentCommit(in)
		}
		if err != nil {
			fmt.Fprintln(errOut, "Commit bloqueado: mensagem ausente ou fora do padrão Conventional Commits.")
			return 2
		}
		return 0
	default:
		return invalid(errOut, "comando desconhecido: "+args[0])
	}
}

func interactive(in io.Reader, out io.Writer) bool {
	input, inputOK := in.(*os.File)
	output, outputOK := out.(*os.File)
	return inputOK && outputOK && term.IsTerminal(input.Fd()) && term.IsTerminal(output.Fd())
}

func plainTerminal() bool {
	return os.Getenv("ACCESSIBLE") != "" || os.Getenv("TERM") == "dumb" || os.Getenv("NO_COLOR") != ""
}

func invalid(errOut io.Writer, message string) int {
	fmt.Fprintln(errOut, message)
	fmt.Fprintln(errOut, "Use --help para opções.")
	return 2
}

func homePath(input string) (string, error) {
	if input != "" {
		if !filepath.IsAbs(input) {
			return "", errors.New("--home deve ser caminho absoluto")
		}
		return filepath.Clean(input), nil
	}
	return os.UserHomeDir()
}

func install(args []string, in io.Reader, out, errOut io.Writer, guided bool) int {
	flags := flag.NewFlagSet("install", flag.ContinueOnError)
	flags.SetOutput(errOut)
	var home string
	var yes, skipAgentUpdate bool
	var skills, selectedHooks names
	flags.StringVar(&home, "home", "", "HOME de destino")
	flags.BoolVar(&yes, "yes", false, "confirma após prévia")
	flags.BoolVar(&skipAgentUpdate, "skip-agent-update", false, "não atualizar agentes")
	flags.Var(&skills, "skill", "skill opcional (repetível)")
	flags.Var(&selectedHooks, "hook", "hook opcional (repetível)")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Fprint(out, usage)
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 {
		return invalid(errOut, "argumento inesperado: "+flags.Arg(0))
	}
	if guided {
		specialties := make([]huh.Option[string], 0, len(catalog.SpecialtySkills()))
		for _, name := range catalog.SpecialtySkills() {
			specialties = append(specialties, huh.NewOption(name+" · "+catalog.SpecialtyDescription(name), name))
		}
		var engineering, productivity []huh.Option[string]
		for _, name := range catalog.Skills() {
			option := huh.NewOption(name, name)
			if strings.Contains(catalog.MattSource(name), "/skills/productivity/") {
				productivity = append(productivity, option)
			} else {
				engineering = append(engineering, option)
			}
		}
		hookOptions := make([]huh.Option[string], 0, len(catalog.Hooks()))
		for _, name := range catalog.Hooks() {
			hookOptions = append(hookOptions, huh.NewOption(name, name))
		}
		var engineeringSkills, productivitySkills, specialistSkills, chosenHooks []string
		updateAgents := true
		form := huh.NewForm(
			huh.NewGroup(huh.NewNote().Title("Core GEX obrigatório").Description(fmt.Sprintf("%d skills Core para Claude Code e Codex; nenhuma especialização é selecionada automaticamente.", len(catalog.CoreSkills()))).Next(true)).
				Title("00  /  CORE"),
			huh.NewGroup(huh.NewMultiSelect[string]().
				Title("Skills de engenharia").
				Description("O Core GEX é obrigatório. Espaço seleciona skills; / filtra.").
				Options(engineering...).Value(&engineeringSkills).Filterable(true).Height(9).Validate(catalog.ValidateEngineeringSelection)).
				Title("01  /  ENGENHARIA"),
			huh.NewGroup(huh.NewMultiSelect[string]().
				Title("Skills de produtividade").
				Description("Adicione somente o que usará no trabalho.").
				Options(productivity...).Value(&productivitySkills).Filterable(true).Height(7).
				Validate(func(chosen []string) error {
					return catalog.ValidateSelection(slices.Concat(engineeringSkills, chosen), nil)
				})).
				Title("02  /  PRODUTIVIDADE"),
			huh.NewGroup(huh.NewMultiSelect[string]().
				Title("Especializações GEX").
				Description("Opt-in. Acesso a gateways somente leitura; credenciais no ambiente ao usar.").
				Options(specialties...).Value(&specialistSkills).Filterable(true).Height(7)).
				Title("03  /  ESPECIALIZAÇÕES"),
			huh.NewGroup(huh.NewMultiSelect[string]().
				Title("Hooks").
				Description("Codex exigirá aprovação posterior em /hooks.").
				Options(hookOptions...).Value(&chosenHooks).Height(3)).
				Title("04  /  HOOK GLOBAL"),
			huh.NewGroup(huh.NewConfirm().
				Title("Atualizar Claude Code e Codex?").
				Description("Somente instaladores standalone elegíveis; veja prévia antes de aplicar.").
				Affirmative("Atualizar").Negative("Agora não").Value(&updateAgents)).
				Title("05  /  AGENTES"),
		).WithInput(in).WithOutput(out).WithAccessible(plainTerminal())
		if plainTerminal() {
			form.WithTheme(huh.ThemeFunc(huh.ThemeBase))
		} else {
			form.WithTheme(huh.ThemeFunc(gexFormTheme))
		}
		if err := form.Run(); err != nil {
			return formError(err, errOut)
		}
		skills = append(skills, engineeringSkills...)
		skills = append(skills, productivitySkills...)
		skills = append(skills, specialistSkills...)
		selectedHooks = append(selectedHooks, chosenHooks...)
		skipAgentUpdate = !updateAgents
	}
	destination, err := homePath(home)
	if err != nil {
		fmt.Fprintln(errOut, err)
		return 1
	}
	if err := gatewayCredentialError(skills); err != nil {
		fmt.Fprintln(errOut, err)
		return 1
	}
	if guided && !plainTerminal() {
		executable, err := os.Executable()
		if err != nil {
			fmt.Fprintln(errOut, "Não foi possível localizar executável:", err)
			return 1
		}
		return guidedInstall(destination, executable, skills, selectedHooks, skipAgentUpdate, in, out, errOut)
	}
	executable, err := os.Executable()
	if err != nil {
		fmt.Fprintln(errOut, "Não foi possível localizar executável:", err)
		return 1
	}
	files, err := catalog.Build(destination, executable, skills, selectedHooks)
	if err != nil {
		fmt.Fprintln(errOut, "Catálogo:", err)
		return 1
	}
	plan, err := kit.Preview(destination, files)
	if err != nil {
		fmt.Fprintln(errOut, "Prévia:", err)
		return 1
	}
	var updates, agentChanges []agents.Change
	blocked := false
	if !skipAgentUpdate {
		changes, err := agents.Preview(destination)
		if err != nil {
			fmt.Fprintln(errOut, "Prévia dos agentes:", err)
			return 1
		}
		agentChanges = changes
		for _, change := range changes {
			if change.Blocked {
				blocked = true
			} else if change.Current != change.Target {
				updates = append(updates, change)
			}
		}
	}
	styledPreview := interactive(in, out) && !plainTerminal()
	if styledPreview {
		fmt.Fprintln(out, lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#63B3FF")).Render("Prévia de instalação:"))
	} else {
		fmt.Fprintln(out, "Prévia de instalação:")
	}
	if styledPreview {
		width, _, err := term.GetSize(out.(*os.File).Fd())
		if err != nil {
			width = 80
		}
		fmt.Fprintln(out, previewPanel(plan.Changes, agentChanges, width))
	} else {
		fmt.Fprintln(out, previewText(plan.Changes, agentChanges))
	}
	if len(plan.Changes) == 0 && len(updates) == 0 {
		fmt.Fprintln(out, "Nada a alterar.")
		if blocked {
			fmt.Fprintln(errOut, "Atualização de agentes bloqueada; Core inalterado.")
			return 1
		}
		return 0
	}
	if !yes {
		if !interactive(in, out) {
			fmt.Fprintln(errOut, "Confirmação exige terminal ou --yes; nenhuma alteração aplicada.")
			return 1
		}
		var confirmed bool
		confirm := huh.NewConfirm().Title("Aplicar alterações?").Affirmative("Sim").Negative("Não").Value(&confirmed)
		if plainTerminal() {
			err = confirm.WithTheme(huh.ThemeFunc(huh.ThemeBase)).RunAccessible(out, in)
		} else {
			err = huh.NewForm(huh.NewGroup(confirm)).WithInput(in).WithOutput(out).WithTheme(huh.ThemeFunc(gexFormTheme)).Run()
		}
		if err != nil {
			return formError(err, errOut)
		}
		if !confirmed {
			fmt.Fprintln(out, "Instalação cancelada.")
			return 0
		}
	}
	if len(updates) > 0 {
		if err := agents.Apply(destination, updates); err != nil {
			fmt.Fprintln(errOut, "Atualização de agentes falhou; Core não foi aplicado:", err)
			return 1
		}
		fmt.Fprintln(out, "Agentes atualizados.")
	}
	if len(plan.Changes) > 0 {
		id, err := kit.Apply(plan)
		if err != nil {
			fmt.Fprintln(errOut, "Instalação do Core:", err)
			return 1
		}
		fmt.Fprintln(out, "Instalado. Histórico:", id)
	}
	if len(selectedHooks) > 0 {
		fmt.Fprintln(out, "Codex: abra /hooks e aprove o hook novo para ativá-lo.")
	}
	if blocked {
		fmt.Fprintln(errOut, "Agentes bloqueados não foram atualizados; demais alterações aplicadas.")
		return 1
	}
	return 0
}

func gatewayCredentialError(skills []string) error {
	for _, skill := range skills {
		var required []string
		switch skill {
		case "gex-pagamerican-api":
			required = []string{"PAGAMERICAN_API_KEY"}
		case "gex-buygoods-api":
			required = []string{"BUYGOODS_ACCOUNT_ID", "BUYGOODS_API_TOKEN"}
		case "gex-clickbank-api":
			required = []string{"CLICKBANK_API_KEY", "CLICKBANK_ACCOUNT"}
		}
		for _, name := range required {
			if os.Getenv(name) == "" {
				return fmt.Errorf("%s exige %s no ambiente. Solicite acesso pelo canal corporativo aprovado; nenhum arquivo foi aplicado.", skill, name)
			}
		}
	}
	return nil
}

func safeAgentReason(reason string) string {
	if strings.Contains(reason, "--version inesperado:") {
		return "saída de versão não reconhecida; verifique manualmente"
	}
	return reason
}

func rollbackAgents(args []string, in io.Reader, out, errOut io.Writer) int {
	flags := flag.NewFlagSet("rollback-agents", flag.ContinueOnError)
	flags.SetOutput(errOut)
	var home string
	var yes bool
	flags.StringVar(&home, "home", "", "HOME de destino")
	flags.BoolVar(&yes, "yes", false, "confirma reversão")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Fprint(out, usage)
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 {
		return invalid(errOut, "argumento inesperado: "+flags.Arg(0))
	}
	destination, err := homePath(home)
	if err != nil {
		fmt.Fprintln(errOut, err)
		return 1
	}
	fmt.Fprintln(out, "Rollback dos agentes: restaura somente versões anteriores gerenciadas pelo kit; não altera o Core.")
	if !yes {
		if !interactive(in, out) {
			fmt.Fprintln(errOut, "Confirmação exige terminal ou --yes; nenhuma alteração aplicada.")
			return 1
		}
		var confirmed bool
		confirm := huh.NewConfirm().Title("Restaurar versões anteriores dos agentes?").Affirmative("Sim").Negative("Não").Value(&confirmed)
		if plainTerminal() {
			err = confirm.WithTheme(huh.ThemeFunc(huh.ThemeBase)).RunAccessible(out, in)
		} else {
			err = huh.NewForm(huh.NewGroup(confirm)).WithInput(in).WithOutput(out).WithTheme(huh.ThemeFunc(gexFormTheme)).Run()
		}
		if err != nil {
			return formError(err, errOut)
		}
		if !confirmed {
			fmt.Fprintln(out, "Rollback cancelado.")
			return 0
		}
	}
	if err := agents.Rollback(destination); err != nil {
		fmt.Fprintln(errOut, "Rollback dos agentes:", err)
		return 1
	}
	fmt.Fprintln(out, "Versões anteriores dos agentes restauradas.")
	return 0
}

func gexFormTheme(isDark bool) *huh.Styles {
	styles := huh.ThemeCharm(isDark)
	blue := lipgloss.LightDark(isDark)(lipgloss.Color("#1a5fc0"), lipgloss.Color("#7fc3ef"))
	styles.Group.Title = styles.Group.Title.Foreground(blue)
	styles.Focused.Title = styles.Focused.Title.Foreground(blue)
	styles.Focused.MultiSelectSelector = styles.Focused.MultiSelectSelector.Foreground(blue)
	styles.Focused.SelectedOption = styles.Focused.SelectedOption.Foreground(blue)
	styles.Focused.SelectSelector = styles.Focused.SelectSelector.Foreground(blue)
	styles.Focused.FocusedButton = styles.Focused.FocusedButton.Background(blue).Foreground(lipgloss.LightDark(isDark)(lipgloss.Color("#ffffff"), lipgloss.Color("#0b1526")))
	styles.Focused.Next = styles.Focused.Next.Background(blue).Foreground(lipgloss.LightDark(isDark)(lipgloss.Color("#ffffff"), lipgloss.Color("#0b1526")))
	styles.Blurred = styles.Focused
	return styles
}

func formError(err error, errOut io.Writer) int {
	if errors.Is(err, huh.ErrUserAborted) {
		fmt.Fprintln(errOut, "Operação cancelada.")
	} else {
		fmt.Fprintln(errOut, "Interface:", err)
	}
	return 1
}

func inspect(command string, args []string, in io.Reader, out, errOut io.Writer) int {
	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	flags.SetOutput(errOut)
	var home string
	flags.StringVar(&home, "home", "", "HOME de destino")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Fprint(out, usage)
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 {
		return invalid(errOut, "argumento inesperado: "+flags.Arg(0))
	}
	destination, err := homePath(home)
	if err != nil {
		fmt.Fprintln(errOut, err)
		return 1
	}
	if command == "rollback" {
		id, err := kit.Rollback(destination)
		if err != nil {
			fmt.Fprintln(errOut, "Rollback:", err)
			return 1
		}
		if id == "" {
			fmt.Fprintln(out, "Nada a restaurar.")
		} else {
			fmt.Fprintln(out, "Restaurado. Histórico:", id)
		}
		return 0
	}
	statuses, err := kit.Doctor(destination)
	if err != nil {
		fmt.Fprintln(errOut, "Doctor:", err)
		return 1
	}
	agentChanges, err := agents.Preview(destination)
	if err != nil {
		fmt.Fprintln(errOut, "Doctor: inspeção dos agentes indisponível:", err)
		return 1
	}
	for _, change := range agentChanges {
		status := kit.Status{Path: "agente/" + change.Name, State: "ok", Detail: "Versão fixada " + change.Target + " gerenciada pelo kit."}
		if change.Blocked {
			status.State = "bloqueado"
			status.Detail = "Alvo " + change.Target + "; " + safeAgentReason(change.Reason) + "; ação manual: " + change.Manual
		} else if change.Current != change.Target {
			status.State = "atualizar"
			status.Detail = "Alvo " + change.Target + "; comando: " + change.Command
		}
		statuses = append(statuses, status)
	}
	result := 0
	for _, status := range statuses {
		if status.State == "ausente" || status.State == "alterado" || status.State == "incompleto" || status.State == "bloqueado" || status.State == "atualizar" {
			result = 1
		}
	}
	if interactive(in, out) && !plainTerminal() {
		if _, err := tea.NewProgram(newDoctorModel(statuses), tea.WithInput(in), tea.WithOutput(out)).Run(); err != nil {
			fmt.Fprintln(errOut, "Doctor: interface indisponível:", err)
			return 1
		}
		return result
	}
	for _, status := range statuses {
		fmt.Fprintf(out, "%s %s", status.State, status.Path)
		if status.Detail != "" {
			fmt.Fprintf(out, " (%s)", status.Detail)
		}
		fmt.Fprintln(out)
	}
	if len(statuses) == 0 {
		fmt.Fprintln(out, "Nenhuma instalação registrada.")
	}
	return result
}
