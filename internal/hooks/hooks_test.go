package hooks

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCheckGitMessage(t *testing.T) {
	cases := []struct {
		name, message string
		valid         bool
	}{
		{"tipo aberto", "perf(parser)!: reduz alocações\n\nCorpo livre.\n", true},
		{"tipo simples", "docs: explica instalação", true},
		{"merge automático", "Merge branch 'main' into develop\n", true},
		{"revert automático", "Revert \"feat: mudança\"\n\nThis reverts commit abc.\n", true},
		{"sem tipo", "atualiza instalação", false},
		{"assunto vazio", "fix:  \n\ncorpo", false},
		{"primeira linha inválida", "texto solto\n\nfix: correto", false},
		{"merge incompleto", "Merge ", false},
		{"revert incompleto", "Revert \"", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := CheckGitMessage(tc.message)
			if (err == nil) != tc.valid {
				t.Fatalf("mensagem válida=%v, erro=%v", tc.valid, err)
			}
		})
	}
}

func TestCheckAgentCommit(t *testing.T) {
	cases := []struct {
		name, command string
		valid         bool
	}{
		{"mensagem curta", `git commit -m "feat: adiciona CLI"`, true},
		{"opção longa", `git commit --message="fix(api)!: corrige erro"`, true},
		{"opção combinada", `git commit -am 'docs: atualiza guia'`, true},
		{"diretório git", `git -C pasta commit -m 'test: cobre caso'`, true},
		{"mensagem e corpo", `git commit -m 'feat: cabeçalho' -m 'Corpo livre sem formato'`, true},
		{"corpo com heredoc inexistente", `git commit -m 'fix: válido' -m 'qualquer texto'`, true},
		{"git status", `git status`, true},
		{"pathspec após mensagem", `git commit -m 'fix: arquivo' -- src/main.go`, true},
		{"echo menciona commit", `echo "git commit -m ruim"`, true},
		{"comando após add", `git add . && git commit -m 'fix: correto'`, true},
		{"env executa commit inválido", `env git commit -m 'mensagem livre'`, false},
		{"env com variável executa commit", `env LC_ALL=C git commit -m 'fix: válido'`, true},
		{"sh executa commit inválido", `sh -c 'git commit -m "mensagem livre"'`, false},
		{"bash executa commit válido", `bash -c 'git commit -m "feat: válido"'`, true},
		{"env com shell aninhado", `env LC_ALL=C bash -c 'git commit -m "mensagem livre"'`, false},
		{"shell ecoa texto inerte", `sh -c 'echo "git commit -m ruim"'`, true},
		{"wrapper opaco com commit", `python -c 'import os; os.system("git commit -m ruim")'`, false},
		{"comando após ponto e vírgula", `git status; git commit -m 'sem padrão'`, false},
		{"literal inválida", `git commit -m 'corrige bug'`, false},
		{"tipo inválido", `git commit --message 'fix!: '`, false},
		{"duas mensagens e título inválido", `git commit -m 'texto livre' -m 'fix: corpo'`, false},
		{"arquivo externo", `git commit -F mensagem.txt`, false},
		{"editor", `git commit`, false},
		{"editar mensagem", `git commit -e -m 'fix: correto'`, false},
		{"mensagem dinâmica", `git commit -m "$(cat msg)"`, false},
		{"subcomando dinâmico", `git $(printf commit) -m 'fix: correto'`, false},
		{"commit dentro de substituição", `echo "$(git commit -m 'sem padrão')"`, false},
		{"commit após shell continuado", "git \\\ncommit -m 'sem padrão'", false},
		{"arquivo abreviado", `git commit -m 'fix: correto' -F nota.txt`, false},
		{"editor combinado", `git commit -aem 'fix: correto'`, false},
		{"mensagem interpolada", `git commit -m "fix: $TITULO"`, false},
		{"reaproveitar commit", `git commit -C HEAD`, false},
		{"amend sem edição", `git commit --amend --no-edit`, true},
		{"amend com mensagem", `git commit --amend -m 'fix: revisa'`, true},
		{"amend sem no-edit", `git commit --amend`, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			encoded, _ := json.Marshal(tc.command)
			payload := `{"tool_name":"Bash","tool_input":{"command":` + string(encoded) + `}}`
			err := CheckAgentCommit(strings.NewReader(payload))
			if (err == nil) != tc.valid {
				t.Fatalf("comando válido=%v, erro=%v", tc.valid, err)
			}
		})
	}
}

func TestCheckAgentCommitRejectsNilInput(t *testing.T) {
	if err := CheckAgentCommit(nil); err == nil {
		t.Fatal("leitor ausente aceito")
	}
}

func TestCheckAgentCommitRejectsMalformedInputWithoutLeaking(t *testing.T) {
	secret := "SENHA-NAO-REPETIR"
	cases := []string{
		`{`,
		`{"tool_input":{"command":42}}`,
		`{"tool_name":"Bash","tool_input":{"command":"git commit -m '` + secret + `'"}}`,
		`{"tool_input":{"command":"git commit -m 'fix: correto'"},"tool_name":"Write"}`,
		`{"tool_input":{"command":"git commit -m 'fix: correto'"},"tool_name":"Bash","extra":"` + strings.Repeat("x", 65536) + `"}`,
	}
	for _, input := range cases {
		err := CheckAgentCommit(strings.NewReader(input))
		if err == nil {
			t.Fatalf("entrada inválida aceita")
		}
		if strings.Contains(err.Error(), secret) {
			t.Fatal("erro divulgou conteúdo do comando")
		}
	}
}

func TestCheckAgentCommitInputLimit(t *testing.T) {
	prefix := `{"tool_name":"Bash","tool_input":{"command":"git status","padding":"`
	suffix := `"}}`
	payload := prefix + strings.Repeat("x", maxInput-len(prefix)-len(suffix)) + suffix
	if err := CheckAgentCommit(strings.NewReader(payload)); err != nil {
		t.Fatalf("entrada no limite rejeitada: %v", err)
	}
	if err := CheckAgentCommit(strings.NewReader(payload + " ")); err == nil {
		t.Fatal("entrada acima do limite aceita")
	}
}
