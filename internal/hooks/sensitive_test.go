package hooks

import (
	"errors"
	"strings"
	"testing"
)

const sensitiveSentinel = "SENHA-NAO-REPETIR"

type sensitiveCase struct {
	words    []string
	category string
}

func sensitiveCmd(words ...string) Command {
	cmd := Command{}
	for _, w := range words {
		cmd.Words = append(cmd.Words, Word{Value: w})
	}
	return cmd
}

func sensitiveScript(commands ...Command) Script {
	return Script{Commands: commands}
}

func sensitiveParse(script Script, err error) func(string) (Script, error) {
	return func(string) (Script, error) {
		return script, err
	}
}

func sensitiveContains(list []string, want string) bool {
	for _, item := range list {
		if item == want {
			return true
		}
	}
	return false
}

func sensitiveRunCases(t *testing.T, cases []sensitiveCase) {
	t.Helper()
	for _, c := range cases {
		got := ClassifySensitive([]Command{sensitiveCmd(c.words...)}, "Bash", "", false)
		if got.Ask != (c.category != "") || got.Category != c.category {
			t.Fatalf("%s: esperava categoria %q, veio %+v", strings.Join(c.words, " "), c.category, got)
		}
		if got.Ask && got.Rule == "" {
			t.Fatalf("%s: pergunta sem regra", strings.Join(c.words, " "))
		}
	}
}

func TestSensitiveClassifiesDestructiveCommands(t *testing.T) {
	sensitiveRunCases(t, []sensitiveCase{
		{[]string{"rm", "-rf", "/"}, "destrutiva"},
		{[]string{"rm", "-r", "~"}, "destrutiva"},
		{[]string{"rm", "-R", "*"}, "destrutiva"},
		{[]string{"rm", "--recursive", "."}, "destrutiva"},
		{[]string{"rm", "-rf", ".."}, "destrutiva"},
		{[]string{"git", "push", "--force"}, "destrutiva"},
		{[]string{"git", "push", "-f", "origin", "main"}, "destrutiva"},
		{[]string{"git", "push", "origin", "+main"}, "destrutiva"},
		{[]string{"git", "push", "--force-with-lease"}, "destrutiva"},
		{[]string{"git", "reset", "--hard"}, "destrutiva"},
		{[]string{"git", "clean", "-fd"}, "destrutiva"},
		{[]string{"git", "branch", "-D", "feature/x"}, "destrutiva"},
		{[]string{"git", "checkout", "--", "."}, "destrutiva"},
		{[]string{"git", "restore", "."}, "destrutiva"},
		{[]string{"terraform", "destroy"}, "destrutiva"},
		{[]string{"terraform", "apply", "-auto-approve"}, "destrutiva"},
		{[]string{"kubectl", "delete", "pod", "x"}, "destrutiva"},
		{[]string{"aws", "s3", "rm", "s3://bucket", "--recursive"}, "destrutiva"},
		{[]string{"psql", "-c", "DROP TABLE clientes"}, "destrutiva"},
		{[]string{"mysql", "-e", "TRUNCATE pedidos"}, "destrutiva"},
		{[]string{"sqlite3", "app.db", "DROP DATABASE x"}, "destrutiva"},
		{[]string{"Remove-Item", "dados", "-Recurse"}, "destrutiva"},
	})
}

func TestSensitiveAcceptsHarmlessDestructiveLookalikes(t *testing.T) {
	sensitiveRunCases(t, []sensitiveCase{
		{[]string{"rm", "-rf", "./build"}, ""},
		{[]string{"rm", "arquivo.txt"}, ""},
		{[]string{"git", "reset", "--soft", "HEAD~1"}, ""},
		{[]string{"git", "clean", "-n"}, ""},
		{[]string{"git", "branch", "-d", "feature/x"}, ""},
		{[]string{"git", "checkout", "main"}, ""},
		{[]string{"git", "restore", "--staged", "."}, ""},
		{[]string{"terraform", "plan"}, ""},
		{[]string{"kubectl", "get", "pods"}, ""},
		{[]string{"aws", "s3", "ls"}, ""},
		{[]string{"psql", "-c", "select 1"}, ""},
	})
}

func TestSensitiveClassifiesPublicationCommands(t *testing.T) {
	sensitiveRunCases(t, []sensitiveCase{
		{[]string{"git", "push", "origin", "main"}, "publicação externa"},
		{[]string{"gh", "pr", "merge", "12"}, "publicação externa"},
		{[]string{"gh", "release", "create", "v1.0"}, "publicação externa"},
		{[]string{"gh", "release", "upload", "v1.0", "dist.zip"}, "publicação externa"},
		{[]string{"gh", "release", "delete", "v1.0"}, "publicação externa"},
		{[]string{"gh", "repo", "create", "x"}, "publicação externa"},
		{[]string{"gh", "repo", "delete", "x"}, "publicação externa"},
		{[]string{"gh", "issue", "create", "--title", "x"}, "publicação externa"},
		{[]string{"npm", "publish"}, "publicação externa"},
		{[]string{"pnpm", "publish"}, "publicação externa"},
		{[]string{"yarn", "publish"}, "publicação externa"},
		{[]string{"docker", "push", "img:tag"}, "publicação externa"},
		{[]string{"twine", "upload", "dist/*"}, "publicação externa"},
		{[]string{"cargo", "publish"}, "publicação externa"},
		{[]string{"gh", "release", "list"}, ""},
		{[]string{"gh", "issue", "list"}, ""},
		{[]string{"npm", "install"}, ""},
		{[]string{"git", "fetch", "origin"}, ""},
	})
}

func TestSensitivePRCreateAsksOnlyWithoutPreflightHook(t *testing.T) {
	for _, words := range [][]string{
		{"gh", "pr", "create"},
		{"gh", "pr", "create", "--fill"},
		{"gh", "pr", "create", "-t", "x", "-b", "y"},
	} {
		cmds := []Command{sensitiveCmd(words...)}
		if got := ClassifySensitive(cmds, "Bash", "", false); !got.Ask || got.Category != "publicação externa" {
			t.Fatalf("%s sem hook pré-PR: esperava pergunta, veio %+v", strings.Join(words, " "), got)
		}
		if got := ClassifySensitive(cmds, "Bash", "", true); got != (Decision{}) {
			t.Fatalf("%s com hook pré-PR: esperava vazio, veio %+v", strings.Join(words, " "), got)
		}
	}
}

func TestSensitiveClassifiesSecretReads(t *testing.T) {
	sensitiveRunCases(t, []sensitiveCase{
		{[]string{"cat", ".env"}, "segredo"},
		{[]string{"cat", "config/.env.local"}, "segredo"},
		{[]string{"less", "id_rsa"}, "segredo"},
		{[]string{"head", "-n", "5", "secrets.pem"}, "segredo"},
		{[]string{"tail", "server.key"}, "segredo"},
		{[]string{"cat", "/home/user/.aws/credentials"}, "segredo"},
		{[]string{"type", "credentials"}, "segredo"},
		{[]string{"Get-Content", ".env"}, "segredo"},
		{[]string{"printenv"}, "segredo"},
		{[]string{"printenv", "HOME"}, "segredo"},
		{[]string{"env"}, "segredo"},
		{[]string{"gh", "auth", "token"}, "segredo"},
		{[]string{"aws", "configure", "get", "aws_secret_access_key"}, "segredo"},
		{[]string{"security", "find-generic-password", "-s", "x", "-w"}, "segredo"},
		{[]string{"op", "read", "op://vault/item"}, "segredo"},
		{[]string{"echo", "$TOKEN"}, "segredo"},
		{[]string{"echo", "${GITHUB_TOKEN}"}, "segredo"},
		{[]string{"echo", "$DB_PASSWORD"}, "segredo"},
		{[]string{"cat", "README.md"}, ""},
		{[]string{"cat", "main.go"}, ""},
		{[]string{"env", "FOO=1", "make"}, ""},
		{[]string{"echo", "$HOME"}, ""},
		{[]string{"gh", "auth", "status"}, ""},
		{[]string{"aws", "configure", "list"}, ""},
	})
}

func TestSensitiveIgnoresInocuousCommands(t *testing.T) {
	sensitiveRunCases(t, []sensitiveCase{
		{[]string{"ls"}, ""},
		{[]string{"ls", "-la"}, ""},
		{[]string{"git", "status"}, ""},
		{[]string{"gh", "pr", "list"}, ""},
		{[]string{"gh", "pr", "view", "12"}, ""},
	})
}

func TestSensitiveUnwrapsPrefixesBeforeClassifying(t *testing.T) {
	sensitiveRunCases(t, []sensitiveCase{
		{[]string{"sudo", "rm", "-rf", "/"}, "destrutiva"},
		{[]string{"sudo", "-u", "root", "rm", "-rf", "~"}, "destrutiva"},
		{[]string{"nohup", "git", "push"}, "publicação externa"},
		{[]string{"env", "X=1", "git", "push", "--force"}, "destrutiva"},
		{[]string{"timeout", "10", "git", "push", "origin", "main"}, "publicação externa"},
		{[]string{"xargs", "-n", "1", "npm", "publish"}, "publicação externa"},
	})
}

func TestSensitiveDynamicWordsWithListedTokens(t *testing.T) {
	dynamic := func(value string) Word {
		return Word{Value: value, Dynamic: true}
	}
	cases := []struct {
		name string
		cmd  Command
		ask  bool
	}{
		{"substituição com cat", Command{Words: []Word{{Value: "echo"}, dynamic("$(cat .env)")}}, true},
		{"crase com printenv", Command{Words: []Word{dynamic("`printenv`")}}, true},
		{"eval com variável", Command{Words: []Word{{Value: "eval"}, dynamic("$x")}}, true},
		{"substituição inócua", Command{Words: []Word{{Value: "echo"}, dynamic("$(date)")}}, false},
	}
	for _, c := range cases {
		got := ClassifySensitive([]Command{c.cmd}, "Bash", "", false)
		if got.Ask != c.ask {
			t.Fatalf("%s: pergunta=%v, esperado %v", c.name, got.Ask, c.ask)
		}
		if c.ask && got.Category != "comando dinâmico" {
			t.Fatalf("%s: categoria %q, esperado comando dinâmico", c.name, got.Category)
		}
	}
}

func TestSensitiveReadToolUsesFilePath(t *testing.T) {
	sep := string(rune(92))
	cases := []struct {
		path string
		ask  bool
	}{
		{".env", true},
		{"/repo/id_rsa", true},
		{"certs/server.pem", true},
		{"deploy.key", true},
		{"C:" + sep + "repo" + sep + ".env.local", true},
		{"main.go", false},
		{"credentials.json", false},
	}
	for _, c := range cases {
		got := ClassifySensitive(nil, "Read", c.path, false)
		if got.Ask != c.ask {
			t.Fatalf("Read de %q: pergunta=%v, esperado %v", c.path, got.Ask, c.ask)
		}
		if c.ask && got.Category != "segredo" {
			t.Fatalf("Read de %q: categoria %q, esperado segredo", c.path, got.Category)
		}
	}
	if got := ClassifySensitive(nil, "Write", ".env", false); got != (Decision{}) {
		t.Fatalf("ferramenta fora do matcher não pode perguntar: %+v", got)
	}
}

func TestSensitiveReasonHidesCommandValues(t *testing.T) {
	decisions := []Decision{
		ClassifySensitive([]Command{sensitiveCmd("cat", ".env", sensitiveSentinel)}, "Bash", "", false),
		ClassifySensitive([]Command{sensitiveCmd("git", "push", "--force", sensitiveSentinel)}, "Bash", "", false),
		ClassifySensitive([]Command{sensitiveCmd("echo", "$TOKEN", sensitiveSentinel)}, "Bash", "", false),
		ClassifySensitive([]Command{sensitiveCmd("gh", "pr", "create", sensitiveSentinel)}, "Bash", "", false),
		ClassifySensitive([]Command{{Words: []Word{{Value: "$(cat .env " + sensitiveSentinel + ")", Dynamic: true}}}}, "Bash", "", false),
		ClassifySensitive(nil, "Read", "/x/"+sensitiveSentinel+"/.env", false),
	}
	for _, d := range decisions {
		if !d.Ask {
			t.Fatalf("esperava pergunta, veio %+v", d)
		}
		if reason := d.Reason(); strings.Contains(reason, sensitiveSentinel) {
			t.Fatalf("motivo vazou valor do comando: %q", reason)
		}
	}
}

func TestSensitiveCheckComposesParserAndScope(t *testing.T) {
	cwd := t.TempDir()
	repo := t.TempDir()
	script := sensitiveScript(Command{Words: sensitiveCmd("git", "push", "origin", "main").Words, Dir: repo})
	var seen [][]string
	scope := func(dirs []string) bool {
		seen = append(seen, dirs)
		return true
	}
	in := AgentInput{Tool: "Bash", Command: "git push origin main", Cwd: cwd, Parsed: true}
	got := sensitiveCheck(in, scope, nil, sensitiveParse(script, nil))
	if !got.Ask || got.Category != "publicação externa" {
		t.Fatalf("esperava pergunta de publicação, veio %+v", got)
	}
	if len(seen) != 1 || !sensitiveContains(seen[0], cwd) || !sensitiveContains(seen[0], repo) {
		t.Fatalf("escopo deveria receber cwd e diretório do comando, recebeu %v", seen)
	}
}

func TestSensitiveCheckOutOfScopeIsEmpty(t *testing.T) {
	cwd := t.TempDir()
	outside := func([]string) bool { return false }
	script := sensitiveScript(sensitiveCmd("rm", "-rf", "/"))
	if got := sensitiveCheck(AgentInput{Tool: "Bash", Command: "rm -rf /", Cwd: cwd, Parsed: true}, outside, nil, sensitiveParse(script, nil)); got != (Decision{}) {
		t.Fatalf("fora do escopo deveria ser vazio, veio %+v", got)
	}
	if got := sensitiveCheck(AgentInput{Tool: "Read", FilePath: ".env", Cwd: cwd, Parsed: true}, outside, nil, nil); got != (Decision{}) {
		t.Fatalf("leitura fora do escopo deveria ser vazia, veio %+v", got)
	}
	if got := sensitiveCheck(AgentInput{Truncated: true, Cwd: cwd}, outside, nil, nil); got != (Decision{}) {
		t.Fatalf("entrada grande fora do escopo deveria ser vazia, veio %+v", got)
	}
}

func TestSensitiveCheckTooBigInputInScopeAsks(t *testing.T) {
	cwd := t.TempDir()
	inScope := func([]string) bool { return true }
	cases := []struct {
		name  string
		in    AgentInput
		parse func(string) (Script, error)
	}{
		{"acima do limite", AgentInput{Tool: "Bash", Command: sensitiveSentinel, Cwd: cwd, Parsed: true, Truncated: true}, nil},
		{"ilegível", AgentInput{Tool: "Bash", Command: sensitiveSentinel, Cwd: cwd}, nil},
		{"parser recusou", AgentInput{Tool: "Bash", Command: sensitiveSentinel, Cwd: cwd, Parsed: true}, sensitiveParse(Script{}, errors.New("falha"))},
	}
	for _, c := range cases {
		got := sensitiveCheck(c.in, inScope, nil, c.parse)
		if !got.Ask || got.Category != "comando grande demais para análise" {
			t.Fatalf("%s: esperava pergunta de tamanho, veio %+v", c.name, got)
		}
		if strings.Contains(got.Reason(), sensitiveSentinel) {
			t.Fatalf("%s: motivo vazou valor: %q", c.name, got.Reason())
		}
	}
}

func TestSensitiveCheckPrePRLookupRunsOnlyForPRCreate(t *testing.T) {
	cwd := t.TempDir()
	inScope := func([]string) bool { return true }
	calls := 0
	installed := func() bool {
		calls++
		return true
	}
	ls := sensitiveParse(sensitiveScript(sensitiveCmd("ls")), nil)
	if got := sensitiveCheck(AgentInput{Tool: "Bash", Command: "ls", Cwd: cwd, Parsed: true}, inScope, installed, ls); got != (Decision{}) || calls != 0 {
		t.Fatalf("ls não deve consultar o hook pré-PR: decisão %+v, consultas %d", got, calls)
	}
	prCreate := sensitiveParse(sensitiveScript(sensitiveCmd("gh", "pr", "create", "--fill")), nil)
	if got := sensitiveCheck(AgentInput{Tool: "Bash", Command: "gh pr create --fill", Cwd: cwd, Parsed: true}, inScope, installed, prCreate); got != (Decision{}) || calls != 1 {
		t.Fatalf("com hook pré-PR instalado, gh pr create deve ser vazio: decisão %+v, consultas %d", got, calls)
	}
	notInstalled := func() bool { return false }
	if got := sensitiveCheck(AgentInput{Tool: "Bash", Command: "gh pr create --fill", Cwd: cwd, Parsed: true}, inScope, notInstalled, prCreate); !got.Ask {
		t.Fatalf("sem hook pré-PR, gh pr create deve perguntar: %+v", got)
	}
}

func TestSensitiveScriptDynamicMarksWords(t *testing.T) {
	cwd := t.TempDir()
	script := Script{Dynamic: true, Commands: []Command{sensitiveCmd("echo", "printenv")}}
	got := sensitiveCheck(AgentInput{Tool: "Bash", Command: "echo $(printenv)", Cwd: cwd, Parsed: true}, func([]string) bool { return true }, nil, sensitiveParse(script, nil))
	if !got.Ask || got.Category != "comando dinâmico" {
		t.Fatalf("script dinâmico com palavra sensível deve perguntar: %+v", got)
	}
}

func TestCheckSensitiveReadToolUsesScopeAndPath(t *testing.T) {
	cwd := t.TempDir()
	in := AgentInput{Tool: "Read", FilePath: ".env", Cwd: cwd, Parsed: true}
	if got := CheckSensitive(in, func([]string) bool { return true }, nil); !got.Ask || got.Category != "segredo" {
		t.Fatalf("leitura de .env em escopo deve perguntar: %+v", got)
	}
	if got := CheckSensitive(in, func([]string) bool { return false }, nil); got != (Decision{}) {
		t.Fatalf("leitura de .env fora do escopo deve ser vazia: %+v", got)
	}
	main := AgentInput{Tool: "Read", FilePath: "main.go", Cwd: cwd, Parsed: true}
	if got := CheckSensitive(main, func([]string) bool { return true }, nil); got != (Decision{}) {
		t.Fatalf("leitura de main.go deve ser vazia: %+v", got)
	}
}
