package hooks

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

func sessionFakeGit(toplevel string) GitFunc {
	return func(dir string, args ...string) (string, error) {
		return toplevel + "\n", nil
	}
}

func sessionNoRepo(dir string) (bool, error) {
	return false, errors.New("fora de repositório")
}

func sessionNoGit(dir string, args ...string) (string, error) {
	return "", errors.New("git indisponível")
}

func sessionMkSkill(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("criar skill %s: %v", dir, err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# skill\n"), 0o644); err != nil {
		t.Fatalf("gravar SKILL.md em %s: %v", dir, err)
	}
}

func TestGatherSessionUsaCwdESource(t *testing.T) {
	repo := t.TempDir()
	if err := os.WriteFile(filepath.Join(repo, "AGENTS.md"), []byte("instruções\n"), 0o644); err != nil {
		t.Fatalf("gravar AGENTS.md: %v", err)
	}
	if err := os.Mkdir(filepath.Join(repo, ".claude"), 0o755); err != nil {
		t.Fatalf("criar .claude: %v", err)
	}
	var consultado string
	inGEX := func(dir string) (bool, error) {
		consultado = dir
		return true, nil
	}
	facts := GatherSession(AgentInput{Cwd: repo, Source: "compact"}, "claude", t.TempDir(), inGEX, sessionFakeGit(repo))
	if consultado != repo {
		t.Fatalf("inGEX consultou %q; esperado %q", consultado, repo)
	}
	if facts.Source != "compact" || !facts.GEX || !facts.GEXKnown {
		t.Fatalf("fonte ou escopo errados: %+v", facts)
	}
	if !slices.Equal(facts.LocalFiles, []string{"AGENTS.md", ".claude"}) {
		t.Fatalf("instruções locais = %v; esperado [AGENTS.md .claude]", facts.LocalFiles)
	}
	if facts.RepoName != filepath.Base(repo) {
		t.Fatalf("nome do repositório = %q; esperado %q", facts.RepoName, filepath.Base(repo))
	}
}

func TestGatherSessionGitParadoRespondeEmMenosDeDoisSegundos(t *testing.T) {
	liberar := make(chan struct{})
	t.Cleanup(func() { close(liberar) })
	parado := func(dir string) (bool, error) {
		<-liberar
		return true, nil
	}
	paradoGit := func(dir string, args ...string) (string, error) {
		<-liberar
		return "", nil
	}
	resultado := make(chan SessionFacts, 1)
	go func() {
		resultado <- GatherSession(AgentInput{Cwd: t.TempDir(), Source: "startup"}, "claude", t.TempDir(), parado, paradoGit)
	}()
	select {
	case facts := <-resultado:
		if facts.GEXKnown {
			t.Fatalf("com Git parado, GEXKnown deveria ser falso: %+v", facts)
		}
		if facts.Source != "startup" {
			t.Fatalf("fonte = %q com Git parado; esperado startup", facts.Source)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("GatherSession não respondeu em 2 s com Git parado")
	}
}

func TestGatherSessionAceitaTodasAsFontes(t *testing.T) {
	for _, fonte := range []string{"startup", "resume", "clear", "compact", "fork"} {
		t.Run(fonte, func(t *testing.T) {
			facts := GatherSession(AgentInput{Cwd: t.TempDir(), Source: fonte}, "codex", t.TempDir(), sessionNoRepo, sessionNoGit)
			if facts.Source != fonte {
				t.Fatalf("fonte = %q; esperado %q", facts.Source, fonte)
			}
		})
	}
}

func TestGatherSessionContaSkillsCorePorAgente(t *testing.T) {
	home := t.TempDir()
	sessionMkSkill(t, filepath.Join(home, ".claude", "skills", "gex-abrir-issue"))
	sessionMkSkill(t, filepath.Join(home, ".claude", "skills", "gex-revisar-pr"))
	sessionMkSkill(t, filepath.Join(home, ".claude", "skills", "gex-buygoods-api"))
	sessionMkSkill(t, filepath.Join(home, ".agents", "skills", "gex-preparar-pr"))
	claude := GatherSession(AgentInput{Cwd: t.TempDir()}, "claude", home, sessionNoRepo, sessionNoGit)
	codex := GatherSession(AgentInput{Cwd: t.TempDir()}, "codex", home, sessionNoRepo, sessionNoGit)
	if claude.CoreSkills != 2 || claude.CoreTotal != 7 {
		t.Fatalf("Claude: %d de %d; esperado 2 de 7", claude.CoreSkills, claude.CoreTotal)
	}
	if codex.CoreSkills != 1 || codex.CoreTotal != 7 {
		t.Fatalf("Codex: %d de %d; esperado 1 de 7", codex.CoreSkills, codex.CoreTotal)
	}
}

func TestGatherSessionDetectaBlocoCore(t *testing.T) {
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".claude"), 0o755); err != nil {
		t.Fatalf("criar .claude: %v", err)
	}
	conteudo := "texto do usuário\n<!-- início gex-dev-kit core -->\n"
	if err := os.WriteFile(filepath.Join(home, ".claude", "CLAUDE.md"), []byte(conteudo), 0o644); err != nil {
		t.Fatalf("gravar CLAUDE.md: %v", err)
	}
	claude := GatherSession(AgentInput{Cwd: t.TempDir()}, "claude", home, nil, nil)
	codex := GatherSession(AgentInput{Cwd: t.TempDir()}, "codex", home, nil, nil)
	if !claude.CoreBlock {
		t.Fatalf("bloco Core em CLAUDE.md não foi detectado")
	}
	if codex.CoreBlock {
		t.Fatalf("Codex sem AGENTS.md não deveria ter bloco Core")
	}
}

func TestSessionContextInformaSoFatos(t *testing.T) {
	facts := SessionFacts{Agent: "claude", Source: "fork", GEX: true, GEXKnown: true, LocalFiles: []string{"AGENTS.md"}, CoreSkills: 3, CoreTotal: 7, CoreBlock: true}
	text := SessionContext(facts)
	for _, want := range []string{"repositório GEX", "AGENTS.md", "3 de 7 instaladas em ~/.claude/skills", "presente em ~/.claude/CLAUDE.md", "fonte fork"} {
		if !strings.Contains(text, want) {
			t.Fatalf("texto sem %q:\n%s", want, text)
		}
	}
}

func TestSessionContextEstadosDoRepositorioEDoBloco(t *testing.T) {
	cases := []struct {
		name  string
		facts SessionFacts
		want  string
	}{
		{"desconhecido", SessionFacts{Agent: "claude"}, "Repositório: desconhecido, verificação indisponível."},
		{"fora do escopo", SessionFacts{Agent: "claude", GEXKnown: true}, "Repositório: fora do escopo GEX."},
		{"sem instruções locais", SessionFacts{Agent: "codex", GEXKnown: true}, "Instruções locais: nenhuma."},
		{"bloco ausente", SessionFacts{Agent: "codex"}, "Bloco Core global: ausente."},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if text := SessionContext(c.facts); !strings.Contains(text, c.want) {
				t.Fatalf("texto sem %q:\n%s", c.want, text)
			}
		})
	}
}

func TestSessionContextRespeitaLimiteDeCaracteres(t *testing.T) {
	var arquivos []string
	for i := 0; i < 400; i++ {
		arquivos = append(arquivos, strings.Repeat("arquivo-longo-", 5)+"x")
	}
	text := SessionContext(SessionFacts{Agent: "claude", GEXKnown: true, LocalFiles: arquivos})
	if !strings.HasPrefix(text, "GEX: início de sessão") {
		t.Fatalf("texto sem cabeçalho: %q", text)
	}
	if n := utf8.RuneCountInString(text); n > sessionMaxText {
		t.Fatalf("texto com %d caracteres; limite de %d", n, sessionMaxText)
	}
}
