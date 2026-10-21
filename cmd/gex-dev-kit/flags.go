package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/terminal"
)

type options struct {
	UI              terminal.Mode
	NoAnim, Quiet   bool
	JSON, Yes       bool
	SkipAgentUpdate bool
	Help, Version   bool
	Home, Repo      string
	Skills, Hooks   []string
	PreCommit       []string
	PrePush         []string
	Args            []string
}

type usageError struct{ cause, action string }

func (e usageError) Error() string { return e.cause }

const acaoPadrao = "Use --help para opções."

func flagSet(nomes ...string) map[string]bool {
	set := make(map[string]bool, len(nomes))
	for _, nome := range nomes {
		set[nome] = true
	}
	return set
}

var flagsGlobais = flagSet("--ui", "--no-anim", "--quiet", "--json", "-h", "--help", "-v", "--version")

var flagsPorComando = map[string]map[string]bool{
	"install":         flagSet("--home", "--skill", "--hook", "--skip-agent-update", "--yes"),
	"doctor":          flagSet("--home", "--repo"),
	"rollback":        flagSet("--home", "--repo", "--yes"),
	"rollback-agents": flagSet("--home", "--yes"),
	"githooks":        flagSet("--home", "--repo", "--hook", "--pre-commit-run", "--run", "--yes"),
}

var comandosComJSON = flagSet("skills", "hooks", "addons", "doctor")

var flagsComValor = flagSet("--ui", "--home", "--repo", "--skill", "--hook", "--pre-commit-run", "--run")

func splitCommand(args []string) (string, []string) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "-" || !strings.HasPrefix(arg, "-") {
			rest := append(append([]string{}, args[:i]...), args[i+1:]...)
			return arg, rest
		}
		if flagsComValor[arg] {
			i++
		}
	}
	return "", args
}

func parseOptions(command string, args []string) (options, error) {
	var o options
	if !knownCommand(command) {
		return o, usageError{"comando desconhecido: " + command, acaoPadrao}
	}
	if command == "hook" {
		for _, arg := range args {
			if arg == "-h" || arg == "--help" {
				o.Help = true
				continue
			}
			o.Args = append(o.Args, arg)
		}
		return o, nil
	}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") || arg == "-" {
			o.Args = append(o.Args, arg)
			continue
		}
		nome, valor, temValor := strings.Cut(arg, "=")
		if err := flagPermitida(command, nome); err != nil {
			return o, err
		}
		if flagsComValor[nome] && !temValor {
			if i+1 >= len(args) {
				return o, usageError{nome + " exige um valor", acaoPadrao}
			}
			i++
			valor, temValor = args[i], true
		}
		if !flagsComValor[nome] && temValor {
			return o, usageError{nome + " não aceita valor", acaoPadrao}
		}
		if err := aplicaFlag(&o, nome, valor); err != nil {
			return o, err
		}
	}
	return o, validaOpcoes(command, o)
}

func flagPermitida(command, nome string) error {
	if flagsGlobais[nome] {
		if nome == "--json" && !comandosComJSON[command] {
			return usageError{"--json só vale em skills, hooks, addons e doctor", acaoPadrao}
		}
		return nil
	}
	if flagsPorComando[command][nome] {
		return nil
	}
	if flagConhecidaEmQualquerComando(nome) {
		return usageError{fmt.Sprintf("opção %s não se aplica a %s", nome, rotuloComando(command)), acaoPadrao}
	}
	return usageError{"opção desconhecida: " + nome, acaoPadrao}
}

func flagConhecidaEmQualquerComando(nome string) bool {
	for _, set := range flagsPorComando {
		if set[nome] {
			return true
		}
	}
	return false
}

func rotuloComando(command string) string {
	if command == "" {
		return "gex-dev-kit sem comando"
	}
	return command
}

func aplicaFlag(o *options, nome, valor string) error {
	switch nome {
	case "--ui":
		modo, err := terminal.ParseMode(valor)
		if err != nil {
			return usageError{err.Error(), acaoPadrao}
		}
		o.UI = modo
	case "--home":
		return defineCaminho(&o.Home, nome, valor)
	case "--repo":
		return defineCaminho(&o.Repo, nome, valor)
	case "--skill":
		return acrescentaValor(&o.Skills, nome, valor)
	case "--hook":
		return acrescentaValor(&o.Hooks, nome, valor)
	case "--pre-commit-run":
		return acrescentaValor(&o.PreCommit, nome, valor)
	case "--run":
		return acrescentaValor(&o.PrePush, nome, valor)
	case "--no-anim":
		o.NoAnim = true
	case "--quiet":
		o.Quiet = true
	case "--json":
		o.JSON = true
	case "--yes":
		o.Yes = true
	case "--skip-agent-update":
		o.SkipAgentUpdate = true
	case "-h", "--help":
		o.Help = true
	case "-v", "--version":
		o.Version = true
	}
	return nil
}

func defineCaminho(destino *string, nome, valor string) error {
	if valor == "" || !filepath.IsAbs(valor) {
		return usageError{fmt.Sprintf("%s exige caminho absoluto; recebido %q", nome, valor), nome + " \"$(mktemp -d)\" usa um caminho completo"}
	}
	*destino = valor
	return nil
}

func acrescentaValor(destino *[]string, nome, valor string) error {
	if valor == "" {
		return usageError{nome + " exige um valor", acaoPadrao}
	}
	*destino = append(*destino, valor)
	return nil
}

func validaOpcoes(command string, o options) error {
	if o.Home != "" && o.Repo != "" {
		return usageError{"use --home ou --repo, não os dois", acaoPadrao}
	}
	if o.UI == terminal.ModeTUI && (o.Yes || o.Quiet || o.JSON) {
		return usageError{"--ui=tui não combina com --yes, --quiet ou --json", "Use --ui=cli para saída sem TUI."}
	}
	switch command {
	case "addons":
		if len(o.Args) > 1 {
			return usageError{"argumento inesperado: " + o.Args[1], acaoPadrao}
		}
	case "help":
		if len(o.Args) > 1 {
			return usageError{"argumento inesperado: " + o.Args[1], acaoPadrao}
		}
		if len(o.Args) == 1 && !knownCommand(o.Args[0]) {
			return usageError{"comando desconhecido: " + o.Args[0], acaoPadrao}
		}
	case "hook":
	default:
		if len(o.Args) > 0 {
			return usageError{"argumento inesperado: " + o.Args[0], acaoPadrao}
		}
	}
	return nil
}

func knownCommand(name string) bool {
	if name == "" {
		return true
	}
	_, ok := helpSpecs[name]
	return ok
}
