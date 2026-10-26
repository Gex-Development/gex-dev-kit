package main

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"charm.land/huh/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/app"
	"github.com/Gex-Development/gex-dev-kit/internal/catalog"
	"github.com/charmbracelet/colorprofile"
)

type accessibleAction struct {
	name, title, description string
}

var accessibleActions = []accessibleAction{
	{"install", "Instalar Core", "Prévia e instalação guiada"},
	{"doctor", "Doctor", "Inspecionar estado do kit"},
	{"rollback", "Reverter Core", "Restaurar arquivos do kit"},
	{"rollback-agents", "Reverter agentes", "Restaurar versões anteriores"},
	{"skills", "Catálogo", "Ver skills do Core, Matt e especializações"},
	{"addons", "Addons", "Ver addons e o comando oficial de cada um"},
	{"githooks", "Git hooks", "Prévia dos git hooks de um repositório"},
	{"ds", "Vitrine", "Listar componentes do Design System"},
	{"help", "Ajuda", "Ver comandos e exemplos"},
	{"exit", "Sair", "Fechar o Kit GEX"},
}

var errAccessibleEOF = errors.New("entrada encerrada antes da resposta")

type lineReader struct {
	reader io.Reader
}

func newLineReader(r io.Reader) *lineReader {
	return &lineReader{reader: r}
}

func (l *lineReader) ReadLine() (string, bool, error) {
	var line strings.Builder
	var b [1]byte
	for {
		n, err := l.reader.Read(b[:])
		if n == 1 {
			if b[0] == '\n' {
				return strings.TrimSuffix(line.String(), "\r"), true, nil
			}
			line.WriteByte(b[0])
			continue
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				return line.String(), line.Len() > 0, nil
			}
			return "", false, err
		}
	}
}

func readAccessibleLine(l *lineReader) (string, error) {
	text, ok, err := l.ReadLine()
	if err != nil {
		return "", err
	}
	if !ok {
		return "", errAccessibleEOF
	}
	return text, nil
}

type accessibleByteReader struct {
	reader io.Reader
}

func (r accessibleByteReader) Read(p []byte) (int, error) {
	if len(p) > 1 {
		p = p[:1]
	}
	return r.reader.Read(p)
}

func showAccessibleHome(in io.Reader, out io.Writer) (string, error) {
	lines := newLineReader(in)
	for {
		fmt.Fprintf(out, "GEX / DEV KIT\nCore obrigatório: %d skills. Matt e especializações opcionais; hook global só por escolha.\nEscolha uma ação:\n", len(catalog.CoreSkills()))
		for i, action := range accessibleActions {
			fmt.Fprintf(out, "%d. %s — %s\n", i+1, action.title, action.description)
		}
		fmt.Fprint(out, "Ação (0/q/cancel para sair): ")
		text, ok, err := lines.ReadLine()
		if err != nil {
			return "", err
		}
		if !ok {
			return "exit", nil
		}
		choice := strings.ToLower(strings.TrimSpace(text))
		if choice == "0" || choice == "q" || choice == "cancel" {
			return "exit", nil
		}
		if index, err := strconv.Atoi(choice); err == nil && index > 0 && index <= len(accessibleActions) {
			return accessibleActions[index-1].name, nil
		}
		fmt.Fprintln(out, "Ação inválida.")
	}
}

func promptChoices(l *lineReader, out io.Writer, g app.OptionGroup) ([]string, error) {
	fmt.Fprintf(out, "%s\n%s\n", g.Title, g.Description)
	for i, option := range g.Options {
		fmt.Fprintf(out, "%d. %s — %s\n", i+1, option.Label, option.Hint)
	}
	for {
		fmt.Fprint(out, "Escolha pelos números separados por vírgula (0 ou vazio: nenhum): ")
		text, err := readAccessibleLine(l)
		if err != nil {
			return nil, err
		}
		chosen, err := accessibleParseChoices(text, g)
		if err != nil {
			fmt.Fprintln(out, err)
			continue
		}
		return chosen, nil
	}
}

func accessibleParseChoices(text string, g app.OptionGroup) ([]string, error) {
	text = strings.TrimSpace(text)
	var chosen []string
	if text != "" && text != "0" {
		seen := map[int]bool{}
		for _, part := range strings.Split(text, ",") {
			n, err := strconv.Atoi(strings.TrimSpace(part))
			if err != nil || n < 1 || n > len(g.Options) {
				return nil, fmt.Errorf("Opção inválida: %q. Use números de 1 a %d.", strings.TrimSpace(part), len(g.Options))
			}
			if seen[n] {
				continue
			}
			seen[n] = true
			chosen = append(chosen, g.Options[n-1].Value)
		}
	}
	if g.Validate != nil {
		if err := g.Validate(chosen); err != nil {
			return nil, err
		}
	}
	return chosen, nil
}

func promptConfirm(l *lineReader, out io.Writer, question string) (bool, error) {
	for {
		fmt.Fprintf(out, "%s [s/n]: ", question)
		text, err := readAccessibleLine(l)
		if err != nil {
			return false, err
		}
		switch strings.ToLower(strings.TrimSpace(text)) {
		case "s", "sim", "y", "yes":
			return true, nil
		case "n", "não", "nao", "no":
			return false, nil
		}
		fmt.Fprintln(out, "Responda s ou n.")
	}
}

func promptPath(l *lineReader, out io.Writer, title string, validate func(string) error) (string, error) {
	for {
		fmt.Fprintf(out, "%s (caminho absoluto): ", title)
		text, err := readAccessibleLine(l)
		if err != nil {
			return "", err
		}
		path := strings.TrimSpace(text)
		if !filepath.IsAbs(path) {
			fmt.Fprintln(out, "Informe um caminho absoluto.")
			continue
		}
		if validate != nil {
			if err := validate(path); err != nil {
				fmt.Fprintln(out, err)
				continue
			}
		}
		return path, nil
	}
}

func promptLines(in io.Reader, out io.Writer, title string, max int, validate func(string) error) ([]string, error) {
	plain := &colorprofile.Writer{Forward: out, Profile: colorprofile.NoTTY}
	reader := accessibleByteReader{reader: in}
	check := func(text string) error {
		if strings.TrimSpace(text) == "" || validate == nil {
			return nil
		}
		return validate(text)
	}
	var lines []string
	for len(lines) < max {
		var value string
		field := huh.NewText().
			Title(fmt.Sprintf("%s %d (linha vazia termina)", title, len(lines)+1)).
			Value(&value).
			Validate(check)
		err := huh.NewForm(huh.NewGroup(field)).
			WithAccessible(true).
			WithInput(reader).
			WithOutput(plain).
			Run()
		if err != nil {
			return lines, err
		}
		text := strings.TrimSpace(value)
		if text == "" {
			break
		}
		lines = append(lines, text)
	}
	return lines, nil
}

func accessibleInstallChoices(l *lineReader, out io.Writer, groups []app.OptionGroup) (skills, hooks []string, skipAgents bool, err error) {
	for _, g := range groups {
		chosen, err := promptChoices(l, out, g)
		if err != nil {
			return nil, nil, false, err
		}
		if g.ID == "hooks" {
			hooks = append(hooks, chosen...)
			continue
		}
		skills = append(skills, chosen...)
	}
	update, err := promptConfirm(l, out, "Atualizar agentes para as versões mais recentes?")
	if err != nil {
		return nil, nil, false, err
	}
	return skills, hooks, !update, nil
}
