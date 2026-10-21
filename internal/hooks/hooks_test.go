package hooks

import (
	"strings"
	"testing"
)

func TestCheckGitMessage(t *testing.T) {
	bom := string(rune(0xFEFF))
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
		{"fixup com assunto", "fixup! fix: correto", true},
		{"squash com assunto", "squash! feat: x", true},
		{"amend com assunto", "amend! docs: x", true},
		{"fixup sem assunto", "fixup! ", false},
		{"template com comentário", "fix: x\n\n# Please enter the commit message\n# comentário\n", true},
		{"linha de tesoura", "feat: x\n# ------------------------ >8 ------------------------\n# diff\n", true},
		{"linha vazia inicial", "\n\nfeat: x", true},
		{"crlf", "fix: x\r\n\r\ncorpo", true},
		{"bom", bom + "fix: x", true},
		{"comentário com ponto e vírgula", "; coment\nfix: x", false},
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

func TestCleanGitMessageLikeGit(t *testing.T) {
	cases := []struct {
		name, message, commentChar, want string
	}{
		{"remove comentários", "fix: a\n# dica\n# outra\n", "#", "fix: a\n"},
		{"corta na tesoura", "fix: a\n# ------------------------ >8 ------------------------\n# diff\nxyz", "#", "fix: a"},
		{"remove linhas vazias iniciais", "\n\nfeat: x\n", "#", "feat: x\n"},
		{"comentário personalizado", "; coment\nfeat: x\n# nao é comentário", ";", "feat: x\n# nao é comentário"},
		{"auto trata como cerquilha", "fix: a\n# dica\n", "auto", "fix: a\n"},
		{"vazio usa cerquilha", "fix: a\n# dica\n", "", "fix: a\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := CleanGitMessage(tc.message, tc.commentChar); got != tc.want {
				t.Fatalf("limpeza = %q, esperado %q", got, tc.want)
			}
		})
	}
}

func TestCheckGitMessageAfterCleanWithSemicolon(t *testing.T) {
	cleaned := CleanGitMessage("; Please write a message\nfeat: x\n", ";")
	if err := CheckGitMessage(cleaned); err != nil {
		t.Fatalf("mensagem limpa com ponto e vírgula recusada: %v", err)
	}
}

func TestCheckAgentCommitRejectsNilInput(t *testing.T) {
	if err := CheckAgentCommit(nil); err == nil {
		t.Fatal("leitor ausente aceito")
	}
}

func TestCheckAgentCommitMalformedInput(t *testing.T) {
	secret := "SENHA-NAO-REPETIR"
	allowed := []string{
		`{`,
		`{"tool_input":{"command":42}}`,
		`{"tool_input":{"command":"git commit -m 'fix: correto'"},"tool_name":"Write"}`,
	}
	for _, input := range allowed {
		if err := CheckAgentCommit(strings.NewReader(input)); err != nil {
			t.Fatalf("entrada ilegível sem as palavras de commit bloqueada: %v", err)
		}
	}
	blocked := []string{
		`{"tool_name":"Bash","tool_input":{"command":"git commit -m '` + secret + `"`,
		`{"tool_name":"Bash","tool_input":{"command":"git commit -m '` + secret + `'"}}`,
	}
	for _, input := range blocked {
		err := CheckAgentCommit(strings.NewReader(input))
		if err == nil {
			t.Fatalf("commit fora do padrão aceito")
		}
		if strings.Contains(err.Error(), secret) {
			t.Fatal("erro divulgou conteúdo do comando")
		}
	}
}

func TestCheckAgentCommitInputLimit(t *testing.T) {
	prefix := `{"tool_name":"Bash","tool_input":{"command":"git status","padding":"`
	suffix := `"}}`
	payload := prefix + strings.Repeat("x", MaxInput-len(prefix)-len(suffix)) + suffix
	if err := CheckAgentCommit(strings.NewReader(payload)); err != nil {
		t.Fatalf("entrada no limite rejeitada: %v", err)
	}
	if err := CheckAgentCommit(strings.NewReader(payload + " ")); err != nil {
		t.Fatalf("entrada acima do limite sem as palavras de commit bloqueada: %v", err)
	}
}

func TestCheckAgentCommitOversizeWithCommitWords(t *testing.T) {
	prefix := `{"tool_name":"Bash","tool_input":{"command":"git commit -m 'fix: correto'","padding":"`
	suffix := `"}}`
	payload := prefix + strings.Repeat("x", MaxInput) + suffix
	if err := CheckAgentCommit(strings.NewReader(payload)); err == nil {
		t.Fatal("entrada acima do teto com as palavras de commit aceita")
	}
}
