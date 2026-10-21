package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	"install         Instala Core GEX e atualiza agentes em versões fixadas, após prévia e confirmação\n" +
	"doctor          Exibe estado dos arquivos e agentes gerenciados\n" +
	"rollback        Restaura arquivos da instalação anterior\n" +
	"rollback-agents Restaura versões anteriores dos agentes, com confirmação\n" +
	"skills          Lista skills opcionais\n" +
	"hooks           Lista hooks opcionais\n" +
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
		return "Uso: gex-dev-kit skills\nLista skills opcionais.\n"
	case "hooks":
		return "Uso: gex-dev-kit hooks\nLista hooks opcionais.\n"
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
		return install(nil, in, out, errOut, true)
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
		for _, name := range catalog.Skills() {
			fmt.Fprintln(out, name)
		}
		return 0
	case "hooks":
		if len(args) != 1 {
			return invalid(errOut, "hooks não aceita argumentos")
		}
		for _, name := range catalog.Hooks() {
			fmt.Fprintln(out, name)
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
		options := make([]huh.Option[string], 0, len(catalog.Skills())+len(catalog.Hooks()))
		for _, name := range catalog.Skills() {
			options = append(options, huh.NewOption("Skill: "+name, "skill:"+name))
		}
		for _, name := range catalog.Hooks() {
			options = append(options, huh.NewOption("Hook: "+name, "hook:"+name))
		}
		var choices []string
		if len(options) > 0 {
			if err := huh.NewMultiSelect[string]().Title("Core obrigatório. Escolha itens opcionais").Options(options...).Value(&choices).Run(); err != nil {
				return formError(err, errOut)
			}
		}
		for _, choice := range choices {
			if name, ok := strings.CutPrefix(choice, "skill:"); ok {
				skills = append(skills, name)
			} else if name, ok := strings.CutPrefix(choice, "hook:"); ok {
				selectedHooks = append(selectedHooks, name)
			}
		}
	}
	destination, err := homePath(home)
	if err != nil {
		fmt.Fprintln(errOut, err)
		return 1
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
	var updates []agents.Change
	blocked := false
	if !skipAgentUpdate {
		changes, err := agents.Preview(destination)
		if err != nil {
			fmt.Fprintln(errOut, "Prévia dos agentes:", err)
			return 1
		}
		fmt.Fprintln(out, "Agentes (versões fixadas):")
		for _, change := range changes {
			current := change.Current
			if current == "" {
				current = "não gerenciado"
			}
			if change.Blocked {
				blocked = true
				fmt.Fprintf(out, "  %s bloqueado (%s → %s): %s\n  Ação manual: %s\n", change.Name, current, change.Target, safeAgentReason(change.Reason), change.Manual)
			} else if change.Current != change.Target {
				updates = append(updates, change)
				fmt.Fprintf(out, "  %s %s → %s: %s\n", change.Name, current, change.Target, change.Command)
			} else {
				fmt.Fprintf(out, "  %s %s (sem mudança)\n", change.Name, current)
			}
		}
	}
	fmt.Fprintln(out, lipgloss.NewStyle().Bold(true).Render("Prévia de instalação:"))
	for _, change := range plan.Changes {
		fmt.Fprintf(out, "  %s %s\n", change.Action, change.Path)
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
		if err := huh.NewConfirm().Title("Aplicar alterações?").Affirmative("Sim").Negative("Não").Value(&confirmed).Run(); err != nil {
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
		if err := huh.NewConfirm().Title("Restaurar versões anteriores dos agentes?").Affirmative("Sim").Negative("Não").Value(&confirmed).Run(); err != nil {
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
	if interactive(in, out) && os.Getenv("ACCESSIBLE") == "" && os.Getenv("TERM") != "dumb" {
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
