package hooks

import "testing"

func shellFirstWords(t *testing.T, command string) []string {
	t.Helper()
	script, err := ParseShell(command)
	if err != nil {
		t.Fatalf("análise recusou comando válido: %v", err)
	}
	var names []string
	for _, c := range script.Commands {
		if len(c.Words) == 0 {
			continue
		}
		names = append(names, c.Words[0].Value)
	}
	return names
}

func TestParseShellIgnoresCommentOutsideQuotes(t *testing.T) {
	got := shellFirstWords(t, "ls -la # it's fine")
	if len(got) != 1 || got[0] != "ls" {
		t.Fatalf("comentário tratado como código: %v", got)
	}
}

func TestParseShellKeepsHashInsideQuotesAndWords(t *testing.T) {
	script, err := ParseShell(`echo "a # b" c#d`)
	if err != nil {
		t.Fatalf("análise recusou comando válido: %v", err)
	}
	words := script.Commands[0].Words
	if len(words) != 3 || words[1].Value != "a # b" || words[2].Value != "c#d" {
		t.Fatalf("palavras erradas: %+v", words)
	}
}

func TestParseShellCommentEndsAtNewline(t *testing.T) {
	if got := shellFirstWords(t, "ls # nota\ngit status"); len(got) != 2 || got[1] != "git" {
		t.Fatalf("comando após comentário perdido: %v", got)
	}
}

func TestParseShellQuotedHeredocIsData(t *testing.T) {
	command := "cat > notas.txt <<'EOF'\ndon't\n$(git commit -m 'x')\nEOF"
	if got := shellFirstWords(t, command); len(got) != 1 || got[0] != "cat" {
		t.Fatalf("corpo de heredoc citado virou comando: %v", got)
	}
}

func TestParseShellUnquotedHeredocExaminesSubstitution(t *testing.T) {
	bt := string(rune(96))
	command := "cat <<EOF\n" + bt + "git status" + bt + "\nEOF"
	got := shellFirstWords(t, command)
	if len(got) != 2 || got[1] != "git" {
		t.Fatalf("substituição em heredoc não examinada: %v", got)
	}
}

func TestParseShellHereStringIsData(t *testing.T) {
	if got := shellFirstWords(t, `cat <<< "git commit -m 'x'"`); len(got) != 1 || got[0] != "cat" {
		t.Fatalf("here-string tratado como comando: %v", got)
	}
}

func TestParseShellANSICStringEscapes(t *testing.T) {
	script, err := ParseShell(`echo $'it\'s'`)
	if err != nil {
		t.Fatalf("$'...' recusado: %v", err)
	}
	if got := script.Commands[0].Words[1]; got.Value != "it's" || got.Dynamic {
		t.Fatalf("$'...' decodificado errado: %+v", got)
	}
}

func TestParseShellTrailingBackslashIsLiteral(t *testing.T) {
	script, err := ParseShell(`echo a\`)
	if err != nil {
		t.Fatalf("barra final recusada: %v", err)
	}
	if got := script.Commands[0].Words[1].Value; got != `a\` {
		t.Fatalf("barra final não literal: %q", got)
	}
}

func TestParseShellRejectsUnbalancedQuote(t *testing.T) {
	if _, err := ParseShell(`echo "abc`); err == nil {
		t.Fatal("aspa aberta aceita")
	}
}

func TestParseShellSubstitutionCommandsAreExtracted(t *testing.T) {
	got := shellFirstWords(t, `echo "$(git commit -m 'sem padrão')"`)
	if len(got) != 2 || !commitContainsDir(got, "git") {
		t.Fatalf("comando dentro de substituição não extraído: %v", got)
	}
}

func TestParseShellRecordsDirectory(t *testing.T) {
	script, err := ParseShell("cd outro && git commit -m 'fix: x'")
	if err != nil {
		t.Fatalf("análise recusou comando válido: %v", err)
	}
	if got := script.Commands[1].Dir; got != "outro" {
		t.Fatalf("diretório após cd não registrado: %q", got)
	}
	script, err = ParseShell(`git -C outro commit -m 'fix: x'`)
	if err != nil {
		t.Fatalf("análise recusou comando válido: %v", err)
	}
	if got := script.Commands[0].Dir; got != "outro" {
		t.Fatalf("diretório de git -C não registrado: %q", got)
	}
}

func TestParseShellReadOnlyArgumentsAreData(t *testing.T) {
	for _, command := range []string{
		`grep -n "git commit" README.md`,
		`rg 'git commit -m' .`,
		`printf '%s\n' "git commit -m x"`,
		`cat notas.txt "git commit"`,
		`git log --grep="git commit"`,
	} {
		if got := shellFirstWords(t, command); len(got) != 1 {
			t.Fatalf("argumento de leitura tratado como comando em %q: %v", command, got)
		}
	}
}

func TestEffectiveUnwrapsWrappers(t *testing.T) {
	cases := []struct {
		command, want string
	}{
		{"env LC_ALL=C git commit -m 'x'", "git"},
		{"sudo git commit -m 'x'", "git"},
		{"nohup git push", "git"},
		{"sh -c 'git status'", "git"},
		{"bash -lc 'git status'", "git"},
	}
	for _, tc := range cases {
		script, err := ParseShell(tc.command)
		if err != nil {
			t.Fatalf("análise recusou %q: %v", tc.command, err)
		}
		effective := Effective(script.Commands[0])
		if len(effective) == 0 || len(effective[0].Words) == 0 || effective[0].Words[0].Value != tc.want {
			t.Fatalf("desembrulho errado para %q: %+v", tc.command, effective)
		}
	}
}

func TestHasWordsMatchesWholeWords(t *testing.T) {
	if !HasWords("ls; git commit", "git", "commit") {
		t.Fatal("palavras inteiras não encontradas")
	}
	if HasWords("gitlab commits", "git", "commit") {
		t.Fatal("palavra parcial aceita")
	}
}
