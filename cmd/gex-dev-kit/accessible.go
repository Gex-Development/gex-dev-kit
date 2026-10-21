package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/Gex-Development/gex-dev-kit/internal/app"
	"github.com/Gex-Development/gex-dev-kit/internal/catalog"
)

type accessibleAction struct {
	name, title, description string
}

var accessibleActions = []accessibleAction{
	{"install", "Instalar Core", "Prévia e instalação guiada"},
	{"doctor", "Doctor", "Inspecionar estado do kit"},
	{"rollback", "Reverter Core", "Restaurar arquivos do kit"},
	{"rollback-agents", "Reverter agentes", "Restaurar versões anteriores"},
	{"exit", "Sair", "Fechar o Kit GEX"},
}

type lineReader struct {
	reader *bufio.Reader
}

func newLineReader(r io.Reader) *lineReader {
	return &lineReader{reader: bufio.NewReaderSize(r, 1)}
}

func (l *lineReader) ReadLine() (string, bool, error) {
	var line strings.Builder
	for {
		b, err := l.reader.ReadByte()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return line.String(), line.Len() > 0, nil
			}
			return "", false, err
		}
		if b == '\n' {
			return strings.TrimSuffix(line.String(), "\r"), true, nil
		}
		line.WriteByte(b)
	}
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
	return nil, errors.New("não implementado")
}

func promptConfirm(l *lineReader, out io.Writer, question string) (bool, error) {
	return false, errors.New("não implementado")
}

func promptPath(l *lineReader, out io.Writer, title string, validate func(string) error) (string, error) {
	return "", errors.New("não implementado")
}

func promptLines(in io.Reader, out io.Writer, title string, max int, validate func(string) error) ([]string, error) {
	return nil, errors.New("não implementado")
}
