package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/Gex-Development/gex-dev-kit/internal/app"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/screens"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/terminal"
)

var cliESC = string(rune(27))

var cliBarraU = string(rune(92)) + "u0026"

func cliIsolate(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(home, "gitconfig"))
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("CI", "")
	t.Setenv("NO_COLOR", "")
	t.Setenv("ACCESSIBLE", "")
	t.Setenv("GEX_NO_ANIM", "")
	return home
}

func cliInstallArgs(home string, extra ...string) []string {
	base := []string{"install", "--home", home, "--skip-agent-update"}
	return append(base, extra...)
}

func cliRequest(t *testing.T, args ...string) app.Request {
	t.Helper()
	comando, resto := splitCommand(args)
	o, err := parseOptions(comando, resto)
	if err != nil {
		t.Fatalf("parseOptions(%q): %v", args, err)
	}
	req, err := writeRequest(session{opts: o})
	if err != nil {
		t.Fatalf("writeRequest(%q): %v", args, err)
	}
	return req
}

func cliArvore(t *testing.T, home string) map[string]string {
	t.Helper()
	historico := filepath.Join(".local", "share", "gex-dev-kit", "history")
	arvore := map[string]string{}
	err := filepath.WalkDir(home, func(caminho string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(home, caminho)
		if err != nil {
			return err
		}
		if rel == historico && d.IsDir() {
			return filepath.SkipDir
		}
		if strings.HasPrefix(rel, historico+string(filepath.Separator)) || d.IsDir() {
			return nil
		}
		dados, err := os.ReadFile(caminho)
		if err != nil {
			return err
		}
		arvore[rel] = strings.ReplaceAll(string(dados), home, "<HOME>")
		return nil
	})
	if err != nil {
		t.Fatalf("percorrer %s: %v", home, err)
	}
	return arvore
}

func cliConfereSaidaJSON(t *testing.T, out, errOut string) {
	t.Helper()
	if errOut != "" || strings.Contains(out, cliESC) || strings.Contains(out, cliBarraU) {
		t.Fatalf("saída JSON com stderr, ESC ou \\u0026: err=%q out=%q", errOut, out)
	}
}

func cliDecodeEstrito(t *testing.T, raw string, destino any) {
	t.Helper()
	dec := json.NewDecoder(strings.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(destino); err != nil {
		t.Fatalf("JSON não decodifica na struct conhecida: %v", err)
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		t.Fatalf("conteúdo depois do JSON: %v", err)
	}
}

func cliInstalaCore(t *testing.T, home string) {
	t.Helper()
	if code, _, errOut := invoke(t, cliInstallArgs(home, "--yes")...); code != 0 {
		t.Fatalf("instalação de apoio falhou: exit=%d err=%q", code, errOut)
	}
}

func TestCLIMatrizSemArgumentosPipeMostraAjuda(t *testing.T) {
	cliIsolate(t)
	code, out, errOut := invoke(t)
	if code != 0 || out == "" || errOut != "" || strings.Contains(out, cliESC) {
		t.Fatalf("sem argumentos: exit=%d out=%q err=%q", code, out, errOut)
	}
}

func TestCLIMatrizFlagRaizDesconhecidaSaiDois(t *testing.T) {
	cliIsolate(t)
	code, out, errOut := invoke(t, "--invalid-flag")
	if code != 2 || !strings.Contains(errOut, "opção desconhecida") || out != "" {
		t.Fatalf("flag desconhecida: exit=%d out=%q err=%q", code, out, errOut)
	}
}

func TestCLIMatrizUITuiSemTerminalPedeCLI(t *testing.T) {
	cliIsolate(t)
	code, out, errOut := invoke(t, "--ui=tui", "install")
	if code != 2 || !strings.Contains(errOut, "--ui=cli") || out != "" {
		t.Fatalf("--ui=tui sem terminal: exit=%d out=%q err=%q", code, out, errOut)
	}
}

func TestCLIMatrizInstallUITuiComYesSaiDois(t *testing.T) {
	home := cliIsolate(t)
	code, out, errOut := invoke(t, "install", "--home", home, "--yes", "--ui=tui")
	if code != 2 || out != "" || errOut == "" {
		t.Fatalf("install --yes --ui=tui: exit=%d out=%q err=%q", code, out, errOut)
	}
	if arquivos := cliArvore(t, home); len(arquivos) != 0 {
		t.Fatalf("erro de uso gravou arquivos: %v", arquivos)
	}
}

func TestCLIMatrizInstallHomeRelativoSaiDois(t *testing.T) {
	cliIsolate(t)
	code, out, errOut := invoke(t, "install", "--home", "relativo", "--skip-agent-update", "--yes")
	if code != 2 || out != "" || !strings.Contains(errOut, "absoluto") {
		t.Fatalf("--home relativo: exit=%d out=%q err=%q", code, out, errOut)
	}
}

func TestCLIMatrizInstallSemYesNaoGrava(t *testing.T) {
	home := cliIsolate(t)
	code, out, errOut := invoke(t, cliInstallArgs(home)...)
	if code != 1 || !strings.Contains(out, "Prévia de instalação:") || !strings.Contains(errOut, "Confirmação exige terminal ou --yes") {
		t.Fatalf("install sem --yes em pipe: exit=%d out=%q err=%q", code, out, errOut)
	}
	if arquivos := cliArvore(t, home); len(arquivos) != 0 {
		t.Fatalf("install sem confirmação gravou arquivos: %v", arquivos)
	}
}

func TestCLIMatrizInstallSemCorComAmbienteDeTerminal(t *testing.T) {
	casos := []struct{ nome, chave, valor string }{
		{"NO_COLOR", "NO_COLOR", "1"},
		{"TERM=dumb", "TERM", "dumb"},
		{"ACCESSIBLE", "ACCESSIBLE", "1"},
	}
	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			home := cliIsolate(t)
			t.Setenv(caso.chave, caso.valor)
			code, out, errOut := invoke(t, cliInstallArgs(home, "--yes")...)
			if code != 0 || strings.Contains(out, cliESC) || strings.Contains(errOut, cliESC) {
				t.Fatalf("%s: exit=%d out=%q err=%q", caso.nome, code, out, errOut)
			}
		})
	}
}

func TestCLIMatrizInstallYesSeparaPreviaDeProgresso(t *testing.T) {
	home := cliIsolate(t)
	code, out, errOut := invoke(t, cliInstallArgs(home, "--yes")...)
	if code != 0 || !strings.Contains(out, "Prévia de instalação:") || !strings.Contains(out, "Instalado. Histórico:") {
		t.Fatalf("install --yes: exit=%d out=%q err=%q", code, out, errOut)
	}
	if strings.Contains(out, "[ok]") || strings.Contains(out, "[..]") {
		t.Fatalf("progresso apareceu no stdout: %q", out)
	}
	if strings.Contains(errOut, cliESC) || !strings.Contains(errOut, "[ok] Core gravado") {
		t.Fatalf("stderr sem progresso limpo: %q", errOut)
	}
	for _, linha := range strings.Split(strings.TrimSpace(errOut), "\n") {
		if !strings.HasPrefix(linha, "[..]") && !strings.HasPrefix(linha, "[ok]") {
			t.Fatalf("linha fora de [..]/[ok] no stderr: %q", linha)
		}
	}
}

func TestCLIMatrizInstallQuietYesSoPreviaEHistorico(t *testing.T) {
	home := cliIsolate(t)
	code, out, errOut := invoke(t, cliInstallArgs(home, "--yes", "--quiet")...)
	if code != 0 || errOut != "" || !strings.Contains(out, "Prévia de instalação:") || !strings.Contains(out, "Instalado. Histórico:") {
		t.Fatalf("install --quiet --yes: exit=%d out=%q err=%q", code, out, errOut)
	}
}

func TestCLIMatrizInstallNaoVazaSentinelaDeSettings(t *testing.T) {
	home := cliIsolate(t)
	dir := filepath.Join(home, ".claude")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("criar .claude: %v", err)
	}
	sentinela := "sentinela-w5t2-9f3"
	conteudo := `{"env": {"W5T2_SEGREDO": "` + sentinela + `"}}`
	if err := os.WriteFile(filepath.Join(dir, "settings.json"), []byte(conteudo), 0o644); err != nil {
		t.Fatalf("gravar settings.json: %v", err)
	}
	code, out, errOut := invoke(t, cliInstallArgs(home, "--hook", "gex-session-start", "--yes")...)
	if code != 0 {
		t.Fatalf("install com hook: exit=%d err=%q", code, errOut)
	}
	if strings.Contains(out, sentinela) || strings.Contains(errOut, sentinela) {
		t.Fatalf("sentinela de settings vazou para a saída")
	}
	if !strings.Contains(out, "settings.json") {
		t.Fatalf("prévia sem settings.json: %q", out)
	}
}

func TestCLIMatrizJSONDecodificaEmStructsEstritas(t *testing.T) {
	home := cliIsolate(t)
	t.Setenv("PATH", "")
	cliInstalaCore(t, home)
	t.Run("skills", func(t *testing.T) {
		code, out, errOut := invoke(t, "skills", "--json")
		if code != 0 {
			t.Fatalf("exit=%d", code)
		}
		cliConfereSaidaJSON(t, out, errOut)
		var doc jsonSkillsDoc
		cliDecodeEstrito(t, out, &doc)
		if doc.Schema != "gex-dev-kit.skills/v1" || len(doc.Core) == 0 || doc.Core == nil || doc.Specialties == nil || doc.References == nil || doc.Matt.Skills == nil {
			t.Fatalf("skills: schema=%q ou listas nulas: %+v", doc.Schema, doc)
		}
	})
	t.Run("hooks", func(t *testing.T) {
		code, out, errOut := invoke(t, "hooks", "--json")
		if code != 0 {
			t.Fatalf("exit=%d", code)
		}
		cliConfereSaidaJSON(t, out, errOut)
		var doc jsonHooksDoc
		cliDecodeEstrito(t, out, &doc)
		if doc.Schema != "gex-dev-kit.hooks/v1" || len(doc.AgentHooks) == 0 || len(doc.GitHooks) == 0 || doc.References == nil {
			t.Fatalf("hooks: schema=%q ou listas vazias/nulas", doc.Schema)
		}
	})
	t.Run("addons", func(t *testing.T) {
		code, out, errOut := invoke(t, "addons", "--json")
		if code != 0 {
			t.Fatalf("exit=%d", code)
		}
		cliConfereSaidaJSON(t, out, errOut)
		var doc jsonAddonsDoc
		cliDecodeEstrito(t, out, &doc)
		if doc.Schema != "gex-dev-kit.addons/v1" || len(doc.Addons) != 4 {
			t.Fatalf("addons: schema=%q quantidade=%d", doc.Schema, len(doc.Addons))
		}
	})
	t.Run("doctor", func(t *testing.T) {
		code, out, errOut := invoke(t, "doctor", "--json", "--home", home)
		if code != 0 {
			t.Fatalf("exit=%d err=%q", code, errOut)
		}
		cliConfereSaidaJSON(t, out, errOut)
		var doc jsonDoctorDoc
		cliDecodeEstrito(t, out, &doc)
		if doc.Schema != "gex-dev-kit.doctor/v1" || doc.Items == nil || doc.Agents == nil || doc.Warnings == nil {
			t.Fatalf("doctor: schema=%q ou listas nulas", doc.Schema)
		}
	})
}

func TestCLIMatrizDoctorJSONCodigoIgualAoProcesso(t *testing.T) {
	home := cliIsolate(t)
	t.Setenv("PATH", "")
	code, out, errOut := invoke(t, "doctor", "--json", "--home", home)
	cliConfereSaidaJSON(t, out, errOut)
	var doc jsonDoctorDoc
	cliDecodeEstrito(t, out, &doc)
	if doc.ExitCode != code {
		t.Fatalf("exit_code=%d; código do processo=%d", doc.ExitCode, code)
	}
}

func TestCLIMatrizDoctorAgenteNaoGerenciadoEhAviso(t *testing.T) {
	home := cliIsolate(t)
	cliInstalaCore(t, home)
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "npm"), []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("criar npm falso: %v", err)
	}
	t.Setenv("PATH", dir)
	code, out, errOut := invoke(t, "doctor", "--json", "--home", home)
	if code != 0 || errOut != "" {
		t.Fatalf("doctor com agente não gerenciado: exit=%d err=%q", code, errOut)
	}
	var doc jsonDoctorDoc
	cliDecodeEstrito(t, out, &doc)
	if len(doc.Warnings) == 0 {
		t.Fatalf("agente não gerenciado sem aviso em warnings")
	}
	for _, agente := range doc.Agents {
		if agente.Managed {
			t.Fatalf("agente %s marcado como gerenciado sem histórico", agente.Name)
		}
	}
}

func TestCLIMatrizRollbackSemYesMantemHome(t *testing.T) {
	home := cliIsolate(t)
	cliInstalaCore(t, home)
	antes := cliArvore(t, home)
	code, out, errOut := invoke(t, "rollback", "--home", home)
	if code != 1 || !strings.Contains(out, "Prévia de restauração:") || !strings.Contains(errOut, "--yes") {
		t.Fatalf("rollback sem --yes: exit=%d out=%q err=%q", code, out, errOut)
	}
	if depois := cliArvore(t, home); !reflect.DeepEqual(antes, depois) {
		t.Fatalf("rollback sem --yes alterou o HOME")
	}
}

func TestCLIMatrizRollbackYesRestaura(t *testing.T) {
	home := cliIsolate(t)
	cliInstalaCore(t, home)
	code, out, errOut := invoke(t, "rollback", "--home", home, "--yes")
	if code != 0 || errOut != "" || !strings.Contains(out, "Restaurado. Histórico:") {
		t.Fatalf("rollback --yes: exit=%d out=%q err=%q", code, out, errOut)
	}
	if _, err := os.Stat(filepath.Join(home, ".claude", "CLAUDE.md")); !os.IsNotExist(err) {
		t.Fatalf("CLAUDE.md não foi removido pelo rollback: %v", err)
	}
}

func TestCLIMatrizRollbackVazioNadaARestaurar(t *testing.T) {
	home := cliIsolate(t)
	code, out, errOut := invoke(t, "rollback", "--home", home)
	if code != 0 || errOut != "" || !strings.Contains(out, "Nada a restaurar.") {
		t.Fatalf("rollback vazio: exit=%d out=%q err=%q", code, out, errOut)
	}
}

func TestCLIMatrizGitHooksSemRepoNaoGrava(t *testing.T) {
	home := cliIsolate(t)
	code, out, errOut := invoke(t, "githooks", "--hook", "gex-commit-msg")
	if code != 0 || errOut != "" || !strings.Contains(out, "nada foi gravado") {
		t.Fatalf("githooks sem --repo: exit=%d out=%q err=%q", code, out, errOut)
	}
	if arquivos := cliArvore(t, home); len(arquivos) != 0 {
		t.Fatalf("githooks sem --repo gravou arquivos: %v", arquivos)
	}
}

func TestCLIMatrizGitHooksComandoNaoDocumentadoFalha(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git indisponível")
	}
	home := cliIsolate(t)
	repo := t.TempDir()
	hookGit(t, repo, "init", "-q")
	code, _, errOut := invoke(t, "githooks", "--repo", repo, "--hook", "gex-pre-push", "--run", "comando-inexistente")
	if code != 1 || !strings.Contains(errOut, "precisa estar documentado") {
		t.Fatalf("comando não documentado: exit=%d err=%q", code, errOut)
	}
	if arquivos := cliArvore(t, home); len(arquivos) != 0 {
		t.Fatalf("falha de git hook gravou no HOME: %v", arquivos)
	}
}

func TestCLIMatrizDsCliSemESC(t *testing.T) {
	cliIsolate(t)
	code, out, errOut := invoke(t, "ds", "--ui=cli")
	if code != 0 || errOut != "" || strings.Contains(out, cliESC) || !strings.Contains(out, "Cobertura Charm") {
		t.Fatalf("ds --ui=cli: exit=%d err=%q", code, errOut)
	}
}

func TestCLIMatrizCITTYSimuladoEhCLI(t *testing.T) {
	t.Skip("achado W5-T2: main.go runWith chama terminal.Detect com os streams reais e não aceita terminal.Env injetado; CI=true com TTY simulado não tem ponto de injeção sem mudar código fora da posse")
}

func TestCLIMatrizSinalDuranteGravacao(t *testing.T) {
	casos := []struct {
		nome  string
		sinal os.Signal
		quero int
	}{
		{"SIGINT", os.Interrupt, 130},
		{"SIGTERM", syscall.SIGTERM, 143},
		{"SIGHUP", syscall.SIGHUP, 129},
	}
	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			home := cliIsolate(t)
			canal := make(chan os.Signal, 1)
			g := newGuard(canal)
			defer g.Stop()
			var out, errOut bytes.Buffer
			g.errOut = &errOut
			d := realDeps()
			aplica := d.ApplyInstall
			d.ApplyInstall = func(p app.Plan, progresso func(app.Step)) (app.Result, error) {
				canal <- caso.sinal
				for i := 0; i < 400; i++ {
					if _, ok := g.Code(); ok {
						break
					}
					time.Sleep(5 * time.Millisecond)
				}
				return aplica(p, progresso)
			}
			s := session{
				in:     strings.NewReader(""),
				out:    &out,
				errOut: &errOut,
				env:    terminal.Detect(strings.NewReader(""), &out, os.Environ()),
				opts:   options{UI: terminal.ModeCLI, Home: home, SkipAgentUpdate: true, Yes: true},
				deps:   d,
				guard:  g,
			}
			code := executa(s, "install")
			if code != caso.quero {
				t.Fatalf("exit=%d; quero %d", code, caso.quero)
			}
			if _, err := os.Stat(filepath.Join(home, ".claude", "CLAUDE.md")); err != nil {
				t.Fatalf("gravação interrompida pelo sinal não concluiu o Core: %v", err)
			}
			if !strings.Contains(errOut.String(), "Sinal recebido; concluindo a gravação…") {
				t.Fatalf("aviso de sinal ausente: %q", errOut.String())
			}
		})
	}
}

func TestCLIMatrizFlagsEquivalentesViramMesmoRequest(t *testing.T) {
	home := cliIsolate(t)
	req := cliRequest(t, cliInstallArgs(home, "--skill", "retro", "--skill", "writing-for-agents", "--hook", "gex-session-start", "--yes")...)
	reanalisado := cliRequest(t, strings.Fields(app.EquivalentCommand(req))[1:]...)
	if !reflect.DeepEqual(req, reanalisado) {
		t.Fatalf("Request reanalisado difere: %+v != %+v", reanalisado, req)
	}
}

func TestTUIAndCLIProduceSameHomeFiles(t *testing.T) {
	homeCLI := cliIsolate(t)
	homeTUI := t.TempDir()
	extras := []string{"--skill", "retro", "--skill", "writing-for-agents", "--hook", "gex-session-start", "--yes"}
	if code, _, errOut := invoke(t, cliInstallArgs(homeCLI, extras...)...); code != 0 {
		t.Fatalf("caminho CLI: exit=%d err=%q", code, errOut)
	}
	req := cliRequest(t, cliInstallArgs(homeTUI, extras...)...)
	d := realDeps()
	tui := screens.InstallDeps{
		Options: d.InstallOptions,
		Plan:    func(q app.Request) (app.Plan, error) { return d.PlanInstall(q, app.DiffLinesTUI) },
		Apply:   d.ApplyInstall,
	}
	plano, err := tui.Plan(req)
	if err != nil {
		t.Fatalf("plano da TUI: %v", err)
	}
	if _, err := tui.Apply(plano, nil); err != nil {
		t.Fatalf("gravação da TUI: %v", err)
	}
	cli := cliArvore(t, homeCLI)
	tela := cliArvore(t, homeTUI)
	if len(cli) == 0 {
		t.Fatalf("caminho CLI não gravou arquivos")
	}
	if !reflect.DeepEqual(cli, tela) {
		t.Fatalf("árvores de HOME diferentes: CLI=%d arquivos, TUI=%d arquivos", len(cli), len(tela))
	}
}
