package main

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Gex-Development/gex-dev-kit/internal/app"
)

func TestAccessibleHomeOffersActionsAndCanExit(t *testing.T) {
	for _, choice := range []string{"q", "0", "cancel"} {
		var out bytes.Buffer
		selected, err := showAccessibleHome(strings.NewReader(choice+"\n"), &out)
		if err != nil || selected != "exit" || !strings.Contains(out.String(), "Doctor") || !strings.Contains(out.String(), "Instalar Core") || strings.Contains(out.String(), "\x1b") {
			t.Fatalf("escolha %q: ação=%q err=%v menu=%q", choice, selected, err, out.String())
		}
	}
	var out bytes.Buffer
	selected, err := showAccessibleHome(strings.NewReader("2\n"), &out)
	if err != nil || selected != "doctor" {
		t.Fatalf("ação Doctor: ação=%q err=%v menu=%q", selected, err, out.String())
	}
}

func TestAccessibleHomeListsCatalogAddonsHooksShowcaseAndHelp(t *testing.T) {
	var out bytes.Buffer
	selected, err := showAccessibleHome(strings.NewReader("5\n"), &out)
	if err != nil || selected != "skills" {
		t.Fatalf("ação Catálogo: ação=%q err=%v", selected, err)
	}
	for _, rotulo := range []string{"Catálogo", "Addons", "Git hooks", "Vitrine", "Ajuda"} {
		if !strings.Contains(out.String(), rotulo) {
			t.Fatalf("menu sem %q: %q", rotulo, out.String())
		}
	}
	for escolha, esperado := range map[string]string{"6": "addons", "7": "githooks", "8": "ds", "9": "help"} {
		selected, err := showAccessibleHome(strings.NewReader(escolha+"\n"), &bytes.Buffer{})
		if err != nil || selected != esperado {
			t.Fatalf("escolha %s: ação=%q esperado=%q err=%v", escolha, selected, esperado, err)
		}
	}
}

func TestPromptChoicesReturnsSelectedValues(t *testing.T) {
	g := promptTestGroup(nil)
	chosen, err := promptChoices(newLineReader(strings.NewReader("1,3\n")), &bytes.Buffer{}, g)
	if err != nil || len(chosen) != 2 || chosen[0] != "alfa" || chosen[1] != "gama" {
		t.Fatalf("escolha 1,3: valores=%v err=%v", chosen, err)
	}
}

func TestPromptChoicesZeroOrEmptyReturnsNone(t *testing.T) {
	for _, entrada := range []string{"0\n", "\n"} {
		chosen, err := promptChoices(newLineReader(strings.NewReader(entrada)), &bytes.Buffer{}, promptTestGroup(nil))
		if err != nil || len(chosen) != 0 {
			t.Fatalf("entrada %q: valores=%v err=%v", entrada, chosen, err)
		}
	}
}

func TestPromptChoicesRepeatsInvalidNumber(t *testing.T) {
	var out bytes.Buffer
	chosen, err := promptChoices(newLineReader(strings.NewReader("9\n2\n")), &out, promptTestGroup(nil))
	if err != nil || len(chosen) != 1 || chosen[0] != "beta" {
		t.Fatalf("número inválido: valores=%v err=%v", chosen, err)
	}
	if !strings.Contains(out.String(), "Opção inválida") {
		t.Fatalf("falta aviso de opção inválida: %q", out.String())
	}
}

func TestPromptChoicesValidateRejectsAndRepeats(t *testing.T) {
	var out bytes.Buffer
	validate := func(chosen []string) error {
		if len(chosen) > 1 {
			return errors.New("escolha só uma opção")
		}
		return nil
	}
	chosen, err := promptChoices(newLineReader(strings.NewReader("1,2\n1\n")), &out, promptTestGroup(validate))
	if err != nil || len(chosen) != 1 || chosen[0] != "alfa" {
		t.Fatalf("validação: valores=%v err=%v", chosen, err)
	}
	if !strings.Contains(out.String(), "escolha só uma opção") {
		t.Fatalf("falta a mensagem de Validate: %q", out.String())
	}
}

func TestPromptChoicesStopsWhenInputEnds(t *testing.T) {
	_, err := promptChoices(newLineReader(strings.NewReader("")), &bytes.Buffer{}, promptTestGroup(nil))
	if err == nil {
		t.Fatalf("entrada encerrada deveria devolver erro")
	}
}

func TestPromptConfirmAcceptsAllForms(t *testing.T) {
	cases := map[string]bool{"s\n": true, "sim\n": true, "y\n": true, "yes\n": true, "n\n": false, "não\n": false, "nao\n": false, "no\n": false}
	for entrada, esperado := range cases {
		got, err := promptConfirm(newLineReader(strings.NewReader(entrada)), &bytes.Buffer{}, "Confirmar?")
		if err != nil || got != esperado {
			t.Fatalf("entrada %q: resposta=%v err=%v", entrada, got, err)
		}
	}
}

func TestPromptConfirmRepeatsInvalidAnswer(t *testing.T) {
	var out bytes.Buffer
	got, err := promptConfirm(newLineReader(strings.NewReader("talvez\nsim\n")), &out, "Confirmar?")
	if err != nil || !got {
		t.Fatalf("resposta inválida: resposta=%v err=%v", got, err)
	}
	if !strings.Contains(out.String(), "Responda s ou n.") {
		t.Fatalf("falta o aviso de resposta inválida: %q", out.String())
	}
}

func TestPromptPathRejectsRelativeAndRepeats(t *testing.T) {
	absoluto := filepath.Join(t.TempDir(), "casa")
	var out bytes.Buffer
	caminho, err := promptPath(newLineReader(strings.NewReader("relativo/pasta\n"+absoluto+"\n")), &out, "Pasta", nil)
	if err != nil || caminho != absoluto {
		t.Fatalf("caminho: valor=%q err=%v", caminho, err)
	}
	if !strings.Contains(out.String(), "caminho absoluto") {
		t.Fatalf("falta o aviso de caminho relativo: %q", out.String())
	}
}

func TestPromptPathUsesValidate(t *testing.T) {
	absoluto := filepath.Join(t.TempDir(), "casa")
	validate := func(p string) error {
		if strings.HasSuffix(p, "proibido") {
			return errors.New("caminho proibido")
		}
		return nil
	}
	var out bytes.Buffer
	caminho, err := promptPath(newLineReader(strings.NewReader(absoluto+"proibido\n"+absoluto+"\n")), &out, "Pasta", validate)
	if err != nil || caminho != absoluto || !strings.Contains(out.String(), "caminho proibido") {
		t.Fatalf("validate: valor=%q err=%v menu=%q", caminho, err, out.String())
	}
}

func TestPromptLinesReadsCommandsUntilEmptyLine(t *testing.T) {
	var out bytes.Buffer
	linhas, err := promptLines(strings.NewReader("go test ./...\ngo vet ./...\ngofmt -l .\n\n"), &out, "Comando", 5, nil)
	if err != nil || len(linhas) != 3 || linhas[0] != "go test ./..." || linhas[2] != "gofmt -l ." {
		t.Fatalf("comandos: linhas=%q err=%v", linhas, err)
	}
	if !strings.Contains(out.String(), "Comando 1 (linha vazia termina)") {
		t.Fatalf("título sem a numeração esperada: %q", out.String())
	}
	if strings.Contains(out.String(), "\x1b") {
		t.Fatalf("saída acessível com ANSI: %q", out.String())
	}
}

func TestPromptLinesRespectsMax(t *testing.T) {
	linhas, err := promptLines(strings.NewReader("a\nb\nc\n"), &bytes.Buffer{}, "Comando", 2, nil)
	if err != nil || len(linhas) != 2 || linhas[1] != "b" {
		t.Fatalf("max=2: linhas=%q err=%v", linhas, err)
	}
}

func TestPromptLinesKeepsInputForNextPrompt(t *testing.T) {
	in := strings.NewReader("1\n0\n")
	l := newLineReader(in)
	chosen, err := promptChoices(l, &bytes.Buffer{}, app.OptionGroup{Title: "Teste", Options: []app.Option{{Value: "alfa", Label: "alfa"}}})
	if err != nil || len(chosen) != 1 {
		t.Fatalf("primeira pergunta: valores=%v err=%v", chosen, err)
	}
	proxima, ok, err := l.ReadLine()
	if err != nil || !ok || proxima != "0" {
		t.Fatalf("segunda linha perdida: linha=%q ok=%v err=%v", proxima, ok, err)
	}
}

func TestAccessibleInstallChoicesSplitsSkillsHooksAndAgents(t *testing.T) {
	grupos := []app.OptionGroup{
		promptTestGroup(nil),
		{ID: "hooks", Title: "Hooks", Options: []app.Option{{Value: "gex-conventional-commits", Label: "gex-conventional-commits"}}},
	}
	skills, hooks, skipAgents, err := accessibleInstallChoices(newLineReader(strings.NewReader("2\n1\nn\n")), &bytes.Buffer{}, grupos)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(skills) != 1 || skills[0] != "beta" || len(hooks) != 1 || hooks[0] != "gex-conventional-commits" || !skipAgents {
		t.Fatalf("escolhas: skills=%v hooks=%v pularAgentes=%v", skills, hooks, skipAgents)
	}
}

func promptTestGroup(validate func([]string) error) app.OptionGroup {
	return app.OptionGroup{
		ID:       "teste",
		Title:    "Grupo de teste",
		Options:  []app.Option{{Value: "alfa", Label: "alfa", Hint: "primeira"}, {Value: "beta", Label: "beta"}, {Value: "gama", Label: "gama"}},
		Validate: validate,
	}
}
