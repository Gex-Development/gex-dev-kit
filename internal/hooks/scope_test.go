package hooks

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func scopeFakeGit(output string, err error) (GitFunc, *int) {
	calls := 0
	return func(dir string, args ...string) (string, error) {
		calls++
		return output, err
	}, &calls
}

func TestOrgFromURLTable(t *testing.T) {
	cases := []struct {
		url, org string
		ok       bool
	}{
		{"https://github.com/Gex-Development/gex-dev-kit.git", "Gex-Development", true},
		{"http://github.com/Gex-Integrations/repo", "Gex-Integrations", true},
		{"ssh://git@github.com/gex-data-architecture/repo.git", "gex-data-architecture", true},
		{"ssh://git@github.com:22/gex-data-architecture-org/repo", "gex-data-architecture-org", true},
		{"git+ssh://git@github.com/Gex-Development/repo", "Gex-Development", true},
		{"git@github.com:Gex-Integrations/repo.git", "Gex-Integrations", true},
		{"https://GitHub.com/Gex-Development/repo", "Gex-Development", true},
		{"https://github.com/gex-development/repo.git", "gex-development", true},
		{"https://gitlab.com/Gex-Development/repo.git", "", false},
		{"git@gitlab.com:Gex-Development/repo.git", "", false},
		{"https://github.com/Gex-Development", "", false},
		{"https://github.com//repo", "", false},
		{"texto solto", "", false},
		{"", "", false},
	}
	for _, tc := range cases {
		org, ok := OrgFromURL(tc.url)
		if ok != tc.ok || org != tc.org {
			t.Fatalf("OrgFromURL(%q) = (%q, %v), esperado (%q, %v)", tc.url, org, ok, tc.org, tc.ok)
		}
	}
}

func TestScopeInGEXTable(t *testing.T) {
	cases := []struct {
		name, output string
		err          error
		want         bool
	}{
		{"origin GEX", "remote.origin.url https://github.com/Gex-Development/r.git\n", nil, true},
		{"integrations ssh", "remote.origin.url git@github.com:Gex-Integrations/r.git\n", nil, true},
		{"data architecture porta", "remote.origin.url ssh://git@github.com:22/gex-data-architecture/r\n", nil, true},
		{"caixa trocada", "remote.origin.url https://github.com/GEX-DEVELOPMENT/r.git\n", nil, true},
		{"upstream GEX", "remote.origin.url https://github.com/Alheio/r.git\nremote.upstream.url https://github.com/Gex-Development/r.git\n", nil, true},
		{"organização alheia", "remote.origin.url https://github.com/outra-org/r.git\n", nil, false},
		{"host alheio", "remote.origin.url https://gitlab.com/Gex-Development/r.git\n", nil, false},
		{"sem remote", "", nil, false},
		{"gex.policy true sem GEX", "remote.origin.url https://github.com/outra/r.git\ngex.policy true\n", nil, true},
		{"gex.policy false com GEX", "remote.origin.url https://github.com/Gex-Development/r.git\ngex.policy false\n", nil, false},
		{"sem Git", "", errors.New("git ausente"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			git, _ := scopeFakeGit(tc.output, tc.err)
			scope := NewScope(t.TempDir(), git)
			got, _ := scope.InGEX("/repo")
			if got != tc.want {
				t.Fatalf("dentro do escopo=%v, esperado %v", got, tc.want)
			}
		})
	}
}

func TestScopeInGEXUsesOneGitCall(t *testing.T) {
	var gotDir string
	var gotArgs []string
	calls := 0
	git := func(dir string, args ...string) (string, error) {
		calls++
		gotDir, gotArgs = dir, args
		return "", nil
	}
	scope := NewScope(t.TempDir(), git)
	if _, err := scope.InGEX("/repo"); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if calls != 1 || gotDir != "/repo" || gotArgs[0] != "config" || gotArgs[1] != "--get-regexp" {
		t.Fatalf("chamadas=%d dir=%q args=%v", calls, gotDir, gotArgs)
	}
}

func TestScopeInGEXTimeoutCountsAsOutside(t *testing.T) {
	git := func(dir string, args ...string) (string, error) {
		time.Sleep(5 * time.Second)
		return "remote.origin.url https://github.com/Gex-Development/r.git\n", nil
	}
	scope := Scope{Orgs: DefaultOrgs, Git: git}
	start := time.Now()
	got, err := scope.InGEX("/repo")
	if got || err == nil {
		t.Fatalf("Git que dorme não virou fora do escopo: dentro=%v erro=%v", got, err)
	}
	if elapsed := time.Since(start); elapsed > 3500*time.Millisecond {
		t.Fatalf("prazo de 2 s não respeitado: %v", elapsed)
	}
}

func TestScopeAny(t *testing.T) {
	git := func(dir string, args ...string) (string, error) {
		if dir == "/b" {
			return "remote.origin.url https://github.com/Gex-Development/r.git\n", nil
		}
		return "remote.origin.url https://github.com/outra/r.git\n", nil
	}
	scope := Scope{Orgs: DefaultOrgs, Git: git}
	if !scope.Any([]string{"/a", "/b"}) {
		t.Fatal("diretório GEX entre os candidatos não encontrado")
	}
	if scope.Any([]string{"/a"}) || scope.Any(nil) {
		t.Fatal("diretório fora do escopo aceito")
	}
}

func TestNewScopeReadsOrgsFile(t *testing.T) {
	configDir := t.TempDir()
	dir := filepath.Join(configDir, "gex-dev-kit")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("pasta de teste não criada: %v", err)
	}
	content := "Minha-Org\nlinha com espaço inválida\n\nhttps://exemplo.com/x\n"
	if err := os.WriteFile(filepath.Join(dir, "orgs"), []byte(content), 0o644); err != nil {
		t.Fatalf("arquivo de teste não gravado: %v", err)
	}
	scope := NewScope(configDir, nil)
	if len(scope.Orgs) != len(DefaultOrgs)+1 {
		t.Fatalf("organizações efetivas erradas: %v", scope.Orgs)
	}
	if !strings.Contains(strings.Join(scope.Orgs, " "), "Minha-Org") {
		t.Fatalf("organização do arquivo ausente: %v", scope.Orgs)
	}
	for _, org := range scope.Orgs {
		if strings.ContainsAny(org, " /:") {
			t.Fatalf("linha inválida aceita: %q", org)
		}
	}
}

func TestNewScopeWithoutOrgsFileUsesDefaults(t *testing.T) {
	scope := NewScope(t.TempDir(), nil)
	if len(scope.Orgs) != len(DefaultOrgs) {
		t.Fatalf("lista padrão alterada sem arquivo: %v", scope.Orgs)
	}
}

func TestRunGitWithRealGit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git ausente nesta máquina")
	}
	dir := t.TempDir()
	if _, err := RunGit(dir, "init", "-q"); err != nil {
		t.Fatalf("git init falhou: %v", err)
	}
	if _, err := RunGit(dir, "remote", "add", "origin", "https://github.com/Gex-Development/r.git"); err != nil {
		t.Fatalf("remote não adicionado: %v", err)
	}
	out, err := RunGit(dir, "config", "--get-regexp", `^(remote\..*\.url|gex\.policy)$`)
	if err != nil || !strings.Contains(out, "Gex-Development") {
		t.Fatalf("saída inesperada: %q, erro=%v", out, err)
	}
	scope := NewScope(t.TempDir(), RunGit)
	if ok, _ := scope.InGEX(dir); !ok {
		t.Fatal("repositório GEX real fora do escopo")
	}
}
