package githooks

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/x/exp/golden"
)

const scriptsBinario = "/exemplo/.local/share/gex-dev-kit/bin/gex-dev-kit"

func scriptsRequest(hooks []string, preCommit, prePush []string) Request {
	return Request{Hooks: hooks, PreCommit: preCommit, PrePush: prePush, Binary: scriptsBinario}
}

func TestValidate(t *testing.T) {
	binario := filepath.Join(t.TempDir(), "gex-dev-kit")
	longo := strings.Repeat("a", 301)
	noLimite := strings.Repeat("a", 300)
	oito := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	cases := []struct {
		nome string
		r    Request
		erro bool
	}{
		{"sem hooks", Request{Binary: binario}, true},
		{"hook desconhecido", Request{Hooks: []string{"post-merge"}, Binary: binario}, true},
		{"hook repetido", Request{Hooks: []string{"pre-push", "pre-push"}}, true},
		{"comandos sem o hook", Request{Hooks: []string{"commit-msg"}, Binary: binario, PrePush: []string{"true"}}, true},
		{"commit-msg sem executável", Request{Hooks: []string{"commit-msg"}}, true},
		{"executável relativo", Request{Hooks: []string{"commit-msg"}, Binary: "bin/gex-dev-kit"}, true},
		{"executável com quebra", Request{Hooks: []string{"pre-commit"}, Binary: binario + "\nx"}, true},
		{"pre-push sem executável", Request{Hooks: []string{"pre-push"}, PrePush: []string{"true"}}, false},
		{"nove comandos", Request{Hooks: []string{"pre-push"}, PrePush: append(oito, "i")}, true},
		{"oito comandos", Request{Hooks: []string{"pre-push"}, PrePush: oito}, false},
		{"comando vazio", Request{Hooks: []string{"pre-push"}, PrePush: []string{""}}, true},
		{"comando com quebra", Request{Hooks: []string{"pre-push"}, PrePush: []string{"make\ntest"}}, true},
		{"comando com espaço na borda", Request{Hooks: []string{"pre-push"}, PrePush: []string{" make test"}}, true},
		{"comando com 301 caracteres", Request{Hooks: []string{"pre-push"}, PrePush: []string{longo}}, true},
		{"comando com 300 caracteres", Request{Hooks: []string{"pre-push"}, PrePush: []string{noLimite}}, false},
		{"pre-commit completo", Request{Hooks: []string{"commit-msg", "pre-commit"}, PreCommit: []string{"make test"}, Binary: binario}, false},
	}
	for _, c := range cases {
		t.Run(c.nome, func(t *testing.T) {
			err := Validate(c.r)
			if (err != nil) != c.erro {
				t.Fatalf("Validate() erro = %v, queria erro=%v", err, c.erro)
			}
		})
	}
}

func TestScriptRejeitaHookDesconhecido(t *testing.T) {
	if _, err := Script("post-merge", scriptsRequest([]string{"post-merge"}, nil, nil)); err == nil {
		t.Fatalf("Script() aceitou hook desconhecido")
	}
}

func TestScriptCommitMsg(t *testing.T) {
	script, err := Script("commit-msg", scriptsRequest([]string{"commit-msg"}, nil, nil))
	if err != nil {
		t.Fatalf("Script() erro inesperado: %v", err)
	}
	golden.RequireEqual(t, script)
}

func TestScriptPreCommitSemComandos(t *testing.T) {
	script, err := Script("pre-commit", scriptsRequest([]string{"pre-commit"}, nil, nil))
	if err != nil {
		t.Fatalf("Script() erro inesperado: %v", err)
	}
	golden.RequireEqual(t, script)
}

func TestScriptPreCommitComComandos(t *testing.T) {
	script, err := Script("pre-commit", scriptsRequest([]string{"pre-commit"}, []string{"make test", "npm run lint"}, nil))
	if err != nil {
		t.Fatalf("Script() erro inesperado: %v", err)
	}
	golden.RequireEqual(t, script)
}

func TestScriptPrePushSemComandos(t *testing.T) {
	script, err := Script("pre-push", scriptsRequest([]string{"pre-push"}, nil, nil))
	if err != nil {
		t.Fatalf("Script() erro inesperado: %v", err)
	}
	golden.RequireEqual(t, script)
}

func TestScriptPrePushComComandos(t *testing.T) {
	script, err := Script("pre-push", scriptsRequest([]string{"pre-push"}, nil, []string{"true", "echo 'oi' \"mundo\""}))
	if err != nil {
		t.Fatalf("Script() erro inesperado: %v", err)
	}
	golden.RequireEqual(t, script)
}

func TestScriptsSaoShellPosixValido(t *testing.T) {
	sh, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("sh ausente")
	}
	r := scriptsRequest([]string{"commit-msg", "pre-commit", "pre-push"}, []string{"make test"}, []string{"echo 'oi'"})
	files, err := Files(r)
	if err != nil {
		t.Fatalf("Files() erro inesperado: %v", err)
	}
	for _, f := range files {
		caminho := filepath.Join(t.TempDir(), filepath.Base(f.Path))
		if err := os.WriteFile(caminho, f.Data, 0755); err != nil {
			t.Fatalf("WriteFile() erro: %v", err)
		}
		if out, err := exec.Command(sh, "-n", caminho).CombinedOutput(); err != nil {
			t.Fatalf("sh -n de %s falhou: %v\n%s", f.Path, err, out)
		}
	}
}

func TestScriptsNaoTemComentarios(t *testing.T) {
	files, err := Files(scriptsRequest([]string{"commit-msg", "pre-commit", "pre-push"}, []string{"make test"}, []string{"true"}))
	if err != nil {
		t.Fatalf("Files() erro inesperado: %v", err)
	}
	for _, f := range files {
		for _, linha := range strings.Split(string(f.Data), "\n") {
			if strings.HasPrefix(strings.TrimSpace(linha), "#") && !strings.HasPrefix(linha, "#!") {
				t.Fatalf("%s tem comentário: %q", f.Path, linha)
			}
		}
	}
}

func TestCommandGuardEscolhePrograma(t *testing.T) {
	cases := []struct {
		comando  string
		programa string
	}{
		{"CI=1 make test", "make"},
		{"env FOO=1 go test ./...", "go"},
		{"make test", "make"},
		{"time nice make test", "make"},
	}
	for _, c := range cases {
		want := "if command -v '" + c.programa + "' >/dev/null 2>&1; then"
		if got := commandGuard(c.comando)[0]; got != want {
			t.Fatalf("commandGuard(%q) = %q, esperado %q", c.comando, got, want)
		}
	}
}

func TestShellQuoteEscapaAspas(t *testing.T) {
	cases := map[string]string{
		"make test": "'make test'",
		"it's":      `'it'\''s'`,
	}
	for entrada, want := range cases {
		if got := shellQuote(entrada); got != want {
			t.Fatalf("shellQuote(%q) = %q, queria %q", entrada, got, want)
		}
	}
}

func TestFilesDevolveHooksComModo0755(t *testing.T) {
	files, err := Files(scriptsRequest([]string{"commit-msg", "pre-push"}, nil, []string{"true"}))
	if err != nil {
		t.Fatalf("Files() erro inesperado: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("Files() devolveu %d arquivos, queria 2", len(files))
	}
	for i, nome := range []string{"commit-msg", "pre-push"} {
		if files[i].Path != "hooks/"+nome {
			t.Fatalf("Path = %q, queria hooks/%s", files[i].Path, nome)
		}
		if files[i].Mode.Perm() != 0755 {
			t.Fatalf("Mode = %v, queria 0755", files[i].Mode)
		}
	}
}

func TestFilesRecusaRequisicaoInvalida(t *testing.T) {
	if _, err := Files(Request{Hooks: []string{"commit-msg"}}); err == nil {
		t.Fatalf("Files() aceitou commit-msg sem executável")
	}
}
