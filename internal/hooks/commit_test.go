package hooks

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func commitPayload(command string) string {
	encoded, _ := json.Marshal(command)
	return `{"tool_name":"Bash","tool_input":{"command":` + string(encoded) + `}}`
}

func TestCheckAgentCommitTable(t *testing.T) {
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
		{"eval com commit", `eval "git commit -m 'mensagem livre'"`, false},
		{"eval re-analisa commit válido", `eval "git commit -m 'feat: válido'"`, true},
		{"eval literal com commit inválido", `eval 'git commit -m mensagem-ruim'`, false},
		{"sudo -E esconde commit inválido", `sudo -E git commit -m 'mensagem ruim'`, false},
		{"sudo -n commit inválido", `sudo -n git commit -m 'mensagem livre'`, false},
		{"stdbuf commit inválido", `stdbuf -o L git commit -m 'mensagem livre'`, false},
		{"heredoc de sh com commit inválido", "sh <<EOF\ngit commit -m ruim\nEOF", false},
		{"heredoc de bash com commit válido", "bash <<'EOF'\ngit commit -m 'feat: ok'\nEOF", true},
		{"substituição em chaves com commit inválido", "echo ${X:-$(git commit -m 'mensagem ruim')}", false},
		{"substituição em chaves com commit válido", "echo ${X:-$(git commit -m 'fix: ok')}", true},
		{"xargs executa commit inválido", `xargs sh -c 'git commit -m "mensagem livre"'`, false},
		{"sh -c com variável sem commit", `sh -c "$CMD"`, true},
		{"comando após ponto e vírgula", `git status; git commit -m 'sem padrão'`, false},
		{"literal inválida", `git commit -m 'corrige bug'`, false},
		{"tipo inválido", `git commit --message 'fix!: '`, false},
		{"duas mensagens e título inválido", `git commit -m 'texto livre' -m 'fix: corpo'`, false},
		{"arquivo externo", `git commit -F mensagem.txt`, false},
		{"editor", `git commit`, false},
		{"editar mensagem", `git commit -e -m 'fix: correto'`, false},
		{"mensagem dinâmica", `git commit -m "$(cat msg)"`, false},
		{"subcomando dinâmico", `git $(printf commit) -m 'fix: correto'`, false},
		{"subcomando dinâmico sem commit", `git $SUB -m 'fix: correto'`, true},
		{"commit dentro de substituição", `echo "$(git commit -m 'sem padrão')"`, false},
		{"substituição válida", `echo "$(git commit -m 'fix: x')"`, true},
		{"commit após shell continuado", "git \\\ncommit -m 'sem padrão'", false},
		{"arquivo abreviado", `git commit -m 'fix: correto' -F nota.txt`, false},
		{"editor combinado", `git commit -aem 'fix: correto'`, false},
		{"mensagem interpolada", `git commit -m "fix: $TITULO"`, false},
		{"reaproveitar commit", `git commit -C HEAD`, false},
		{"fixup com ref", `git commit --fixup=HEAD~1`, true},
		{"squash abre editor", `git commit --squash=HEAD~1`, false},
		{"amend sem edição", `git commit --amend --no-edit`, true},
		{"amend com mensagem", `git commit --amend -m 'fix: revisa'`, true},
		{"amend sem no-edit", `git commit --amend`, false},
		{"heredoc com apóstrofo sem commit", "cat > notas.txt <<'EOF'\ndon't\nEOF", true},
		{"heredoc citado com commit no corpo", "cat > log.txt <<'EOF'\ngit commit -m 'mensagem livre'\nEOF", true},
		{"heredoc não citado com commit inválido", "cat <<EOF\n$(git commit -m 'sem padrão')\nEOF", false},
		{"here-string com commit", `cat <<< "git commit -m 'mensagem livre'"`, true},
		{"comentário com apóstrofo", `ls -la # it's fine`, true},
		{"ansi-c com apóstrofo escapado", `echo $'it\'s'`, true},
		{"barra final literal", `echo a \`, true},
		{"grep com commit em aspas", `grep -n "git commit" README.md`, true},
		{"rg com commit", `rg 'git commit -m' .`, true},
		{"git grep com commit", `git grep "git commit"`, true},
		{"git log com commit", `git log --grep="git commit"`, true},
		{"head com commit", `head -n 5 "git commit notes.txt"`, true},
		{"comentário com commit", `tail -f log.txt # git commit -m 'ruim'`, true},
		{"hash dentro de aspas", `echo "# git commit -m ruim"`, true},
		{"printf com commit", `printf 'git commit -m ruim\n'`, true},
		{"cd e commit inválido", `cd outro && git commit -m 'mensagem livre'`, false},
		{"git -C e commit inválido", `git -C outro commit -m 'mensagem livre'`, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := CheckAgentCommit(strings.NewReader(commitPayload(tc.command)))
			if (err == nil) != tc.valid {
				t.Fatalf("comando válido=%v, erro=%v", tc.valid, err)
			}
		})
	}
}

func TestCheckAgentCommitLargeInputWithoutCommit(t *testing.T) {
	command := "echo " + strings.Repeat("a ", 35000) + "git status"
	if err := CheckAgentCommit(strings.NewReader(commitPayload(command))); err != nil {
		t.Fatalf("comando grande sem commit bloqueado: %v", err)
	}
}

func TestCheckAgentCommitLargeInputWithBadCommit(t *testing.T) {
	command := strings.Repeat("a ", 35000) + "; git commit -m 'sem padrão'"
	if err := CheckAgentCommit(strings.NewReader(commitPayload(command))); err == nil {
		t.Fatal("commit fora do padrão aceito em comando grande")
	}
}

func TestCheckAgentCommitLargeJSONWithBadCommit(t *testing.T) {
	prefix := `{"tool_name":"Bash","tool_input":{"command":"git commit -m 'sem padrão'","padding":"`
	suffix := `"}}`
	payload := prefix + strings.Repeat("x", 2<<20) + suffix
	if err := CheckAgentCommit(strings.NewReader(payload)); err == nil {
		t.Fatal("JSON de 2 MiB com commit fora do padrão aceito")
	}
}

func TestCheckAgentCommitDescriptionMentionIsData(t *testing.T) {
	payload := `{"tool_name":"Bash","description":"git commit -m ruim","tool_input":{"command":"ls -la"}}`
	if err := CheckAgentCommit(strings.NewReader(payload)); err != nil {
		t.Fatalf("descrição tratada como comando: %v", err)
	}
}

func TestClassifyCommitResultAndDirectories(t *testing.T) {
	in := AgentInput{Parsed: true, Tool: "Bash", Cwd: "/repo", Command: "cd outro && git commit -m 'fix: x'"}
	status, dirs := ClassifyCommit(in)
	if status != CommitOK {
		t.Fatalf("commit válido classificado como %v", status)
	}
	want := filepath.Join("/repo", "outro")
	if !commitContainsDir(dirs, want) {
		t.Fatalf("diretório do cd ausente: %v", dirs)
	}
}

func TestClassifyCommitIgnoresOtherTools(t *testing.T) {
	in := AgentInput{Parsed: true, Tool: "Write", Command: "git commit -m 'mensagem livre'"}
	if status, _ := ClassifyCommit(in); status != CommitNone {
		t.Fatalf("ferramenta diferente de Bash classificada como %v", status)
	}
}

func TestClassifyCommitUnparsedUsesRawWords(t *testing.T) {
	withWords := AgentInput{Parsed: false, Raw: []byte(`{"tool_name":"Bash","tool_input":{"command":"git commit -m 'x'" `)}
	if status, _ := ClassifyCommit(withWords); status != CommitBad {
		t.Fatalf("texto bruto com as palavras não bloqueado: %v", status)
	}
	withoutWords := AgentInput{Parsed: false, Raw: []byte(`{"tool_name":`)}
	if status, _ := ClassifyCommit(withoutWords); status != CommitNone {
		t.Fatalf("texto sem as palavras bloqueado: %v", status)
	}
}

func TestCheckAgentCommitErrorDoesNotEchoCommand(t *testing.T) {
	secret := "SENHA-NAO-REPETIR"
	err := CheckAgentCommit(strings.NewReader(commitPayload("git commit -m '" + secret + "'")))
	if err == nil {
		t.Fatal("commit fora do padrão aceito")
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatal("erro repetiu o comando do usuário")
	}
}

func commitContainsDir(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
