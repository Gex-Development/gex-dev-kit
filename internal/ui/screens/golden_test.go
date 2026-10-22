package screens

import (
	"errors"
	"image/color"
	"os"
	"strings"
	"sync"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/golden"

	"github.com/Gex-Development/gex-dev-kit/internal/app"
	"github.com/Gex-Development/gex-dev-kit/internal/catalog"
	"github.com/Gex-Development/gex-dev-kit/internal/githooks"
)

const goldenSentinela = "SENHA-NAO-REPETIR"

type goldenEstado struct {
	nome   string
	tiny   bool
	monta  func(t *testing.T, c Common) tea.Model
	avanca func(t *testing.T, m tea.Model) tea.Model
	want   []string
}

type goldenTamanho struct {
	nome string
	w, h int
}

func goldenCommon() Common {
	return Common{Version: "0.3.0", Home: "/home/demo", UserHome: "/home/demo", Dark: true, Profile: colorprofile.ASCII}
}

func goldenFecha(m tea.Model, msgs ...tea.Msg) tea.Model {
	for _, msg := range msgs {
		m, _ = m.Update(msg)
	}
	return m
}

func goldenTecla(s string) tea.KeyPressMsg {
	if s == "enter" {
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	}
	return tea.KeyPressMsg{Code: []rune(s)[0], Text: s}
}

func goldenRequest() app.Request {
	return app.Request{
		Home:            "/home/demo",
		Executable:      "/home/demo/.local/share/gex-dev-kit/bin/gex-dev-kit",
		Skills:          []string{"gex-x"},
		SkipAgentUpdate: true,
	}
}

func goldenPlanoLista() app.Plan {
	return app.Plan{
		Request: goldenRequest(),
		Files: []app.Change{
			{Path: ".claude/CLAUDE.md", Action: "criar"},
			{Path: ".claude/skills/gex-x/SKILL.md", Action: "criar"},
		},
	}
}

func goldenPlanoDiff() app.Plan {
	plano := goldenPlanoLista()
	plano.Diffs = []app.FileDiff{
		{Path: ".claude/CLAUDE.md", Action: "criar", Summary: "novo, 2 linhas", Lines: []app.DiffLine{{Kind: '+', Text: "# Regras"}, {Kind: '+', Text: "linha"}}},
		{
			Path:    ".claude/settings.json",
			Action:  "atualizar",
			Summary: "1 linha",
			Lines: []app.DiffLine{
				{Kind: ' ', Text: "{"},
				{Kind: '+', Text: `  "env": {"API_TOKEN": "•••"}`},
				{Kind: ' ', Text: "}"},
			},
			Masked: 1,
		},
	}
	return plano
}

func goldenInstallDeps() InstallDeps {
	return InstallDeps{
		Apply: func(app.Plan, func(app.Step)) (app.Result, error) { return app.Result{}, nil },
	}
}

func goldenInstallPrevia(c Common, plano app.Plan) tea.Model {
	return goldenFecha(newInstallModel(c, goldenInstallDeps(), goldenRequest(), nil, nil), installPlanned{plan: plano})
}

func goldenRollbackDeps() RollbackDeps {
	return RollbackDeps{
		Plan:  func() (app.RollbackPreview, error) { return rollbackTestPreview(), nil },
		Apply: func() (string, error) { return "h-demo", nil },
	}
}

func goldenDoctorReport(state, detail string) app.Report {
	return app.Report{Root: "/home/demo/repo", Items: []app.Status{{Path: ".claude/CLAUDE.md", State: state, Detail: detail}}}
}

func goldenGitHooks(c Common, f *githooksFake, preset app.GitHooksRequest) tea.Model {
	return goldenFecha(newGitHooksModel(c, f.deps(), &sync.WaitGroup{}, preset), tea.WindowSizeMsg{Width: 80, Height: 24})
}

func goldenGitHooksEscreve(_ *testing.T, m tea.Model) tea.Model {
	g, _ := githooksSend(m.(githooksModel), githooksKey("enter"))
	g, _ = githooksSend(g, githooksType("make test")...)
	g, cmd := githooksSend(g, githooksKey("ctrl+s"))
	g, _ = githooksSend(g, githooksRun(cmd)...)
	return g
}

func goldenShowcaseAte(nome string) func(t *testing.T, m tea.Model) tea.Model {
	return func(t *testing.T, m tea.Model) tea.Model {
		for i := 0; i < showcaseIndicePagina(t, nome); i++ {
			m = goldenFecha(m, tea.KeyPressMsg{Code: tea.KeyDown})
		}
		return m
	}
}

func goldenEstados() []goldenEstado {
	return []goldenEstado{
		{nome: "home_pronta", tiny: true,
			monta: func(_ *testing.T, c Common) tea.Model { return newHomeModel(c, HomeDeps{}) },
			want:  []string{"GEX DEV KIT", "Instalar"}},
		{nome: "home_ajuda",
			monta: func(_ *testing.T, c Common) tea.Model { return newHomeModel(c, HomeDeps{}) },
			avanca: func(_ *testing.T, m tea.Model) tea.Model {
				return goldenFecha(m, tea.KeyPressMsg{Code: '?', Text: "?"})
			},
			want: []string{"ATALHOS"}},
		{nome: "home_erro_leitura",
			monta:  func(_ *testing.T, c Common) tea.Model { return newHomeModel(c, HomeDeps{}) },
			avanca: func(_ *testing.T, m tea.Model) tea.Model { return goldenFecha(m, homeCoreMsg{err: os.ErrPermission}) },
			want:   []string{"permission denied"}},

		{nome: "instalacao_assistente",
			monta: func(_ *testing.T, c Common) tea.Model {
				sel := Selection{Home: "/home/demo"}
				form := NewInstallForm(installTheme(c, c.Dark), &sel, app.InstallOptions())
				return newInstallModel(c, goldenInstallDeps(), goldenRequest(), &sel, form)
			},
			want: []string{"Core obrigatório"}},
		{nome: "instalacao_previa_lista",
			monta: func(_ *testing.T, c Common) tea.Model { return goldenInstallPrevia(c, goldenPlanoLista()) },
			want:  []string{"CLAUDE.md"}},
		{nome: "instalacao_previa_diff",
			monta: func(_ *testing.T, c Common) tea.Model { return goldenInstallPrevia(c, goldenPlanoDiff()) },
			want:  []string{"settings.json", "mascarado"}},
		{nome: "instalacao_confirmar",
			monta:  func(_ *testing.T, c Common) tea.Model { return goldenInstallPrevia(c, goldenPlanoLista()) },
			avanca: func(_ *testing.T, m tea.Model) tea.Model { return goldenFecha(m, goldenTecla("enter")) },
			want:   []string{"APLICAR ALTERAÇÕES"}},
		{nome: "instalacao_gravando",
			monta: func(_ *testing.T, c Common) tea.Model { return goldenInstallPrevia(c, goldenPlanoLista()) },
			avanca: func(_ *testing.T, m tea.Model) tea.Model {
				return goldenFecha(m, goldenTecla("enter"), goldenTecla("s"))
			},
			want: []string{"Não feche o terminal"}},
		{nome: "instalacao_concluido",
			monta: func(_ *testing.T, c Common) tea.Model { return goldenInstallPrevia(c, goldenPlanoLista()) },
			avanca: func(_ *testing.T, m tea.Model) tea.Model {
				return goldenFecha(m, installApplied{result: app.Result{HistoryID: "h-demo"}})
			},
			want: []string{"Core instalado"}},
		{nome: "instalacao_falha", tiny: true,
			monta: func(_ *testing.T, c Common) tea.Model { return goldenInstallPrevia(c, goldenPlanoLista()) },
			avanca: func(_ *testing.T, m tea.Model) tea.Model {
				return goldenFecha(m, installApplied{err: errors.New("gravação recusada")})
			},
			want: []string{"Causa:", "Ação:"}},

		{nome: "doctor_atencao", tiny: true,
			monta: func(_ *testing.T, c Common) tea.Model {
				return newDoctorModel(c, goldenDoctorReport("alterado", "Arquivo foi alterado."))
			},
			want: []string{".claude/CLAUDE.md"}},
		{nome: "doctor_limpo",
			monta: func(_ *testing.T, c Common) tea.Model {
				return newDoctorModel(c, goldenDoctorReport("ok", "Íntegro."))
			},
			want: []string{"Íntegro"}},

		{nome: "rollback_previa",
			monta: func(_ *testing.T, c Common) tea.Model {
				return newRollbackModel(c, RollbackCore, goldenRollbackDeps())
			},
			avanca: func(_ *testing.T, m tea.Model) tea.Model {
				return goldenFecha(m, rollbackPlannedMsg{preview: rollbackTestPreview()})
			},
			want: []string{".claude/CLAUDE.md"}},
		{nome: "rollback_conflito", tiny: true,
			monta: func(_ *testing.T, c Common) tea.Model {
				return newRollbackModel(c, RollbackCore, goldenRollbackDeps())
			},
			avanca: func(_ *testing.T, m tea.Model) tea.Model {
				var preview app.RollbackPreview
				preview.Plan.Conflicts = []string{".claude/CLAUDE.md"}
				return goldenFecha(m, rollbackPlannedMsg{preview: preview})
			},
			want: []string{"Rollback abortado"}},
		{nome: "rollback_concluido",
			monta: func(_ *testing.T, c Common) tea.Model {
				return newRollbackModel(c, RollbackCore, goldenRollbackDeps())
			},
			avanca: func(_ *testing.T, m tea.Model) tea.Model {
				return goldenFecha(m, rollbackPlannedMsg{preview: rollbackTestPreview()}, goldenTecla("s"), rollbackAppliedMsg{id: "h-demo"})
			},
			want: []string{"h-demo"}},
		{nome: "rollback_agentes",
			monta: func(_ *testing.T, c Common) tea.Model {
				return newRollbackModel(c, RollbackAgents, goldenRollbackDeps())
			},
			want: []string{"Reverter agentes", "Claude e Codex voltam"}},

		{nome: "catalogo_skills", tiny: true,
			monta: func(_ *testing.T, c Common) tea.Model { return newCatalogModel(c) },
			want:  []string{"7 Core"}},
		{nome: "catalogo_hooks",
			monta:  func(_ *testing.T, c Common) tea.Model { return newCatalogModel(c) },
			avanca: func(_ *testing.T, m tea.Model) tea.Model { return catalogSend(m.(catalogModel), "tab") },
			want:   []string{"Hooks"}},
		{nome: "catalogo_referencias",
			monta: func(_ *testing.T, c Common) tea.Model { return newCatalogModel(c) },
			avanca: func(_ *testing.T, m tea.Model) tea.Model {
				return catalogSend(catalogSend(m.(catalogModel), "tab"), "tab")
			},
			want: []string{"Referências"}},

		{nome: "addons_card", tiny: true,
			monta: func(_ *testing.T, c Common) tea.Model { return newAddonsModel(c, catalog.Addons()) },
			want:  []string{"caveman"}},
		{nome: "addons_detalhe",
			monta:  func(_ *testing.T, c Common) tea.Model { return newAddonsModel(c, catalog.Addons()) },
			avanca: func(_ *testing.T, m tea.Model) tea.Model { return goldenFecha(m, tea.KeyPressMsg{Code: tea.KeyRight}) },
			want:   []string{"ponytail"}},

		{nome: "githooks_formulario", tiny: true,
			monta: func(_ *testing.T, c Common) tea.Model {
				return goldenGitHooks(c, &githooksFake{plan: githooksPlanFake()}, app.GitHooksRequest{})
			},
			want: []string{"Passo 1 de 4"}},
		{nome: "githooks_previa",
			monta: func(_ *testing.T, c Common) tea.Model {
				return goldenGitHooks(c, &githooksFake{plan: githooksPlanFake()}, app.GitHooksRequest{Repo: "/home/demo/repo", Hooks: []string{"gex-pre-commit"}})
			},
			avanca: goldenGitHooksEscreve,
			want:   []string{"Passo 4 de 4"}},
		{nome: "githooks_nao_documentado",
			monta: func(_ *testing.T, c Common) tea.Model {
				return goldenGitHooks(c, &githooksFake{planErr: githooks.ErrUndocumented}, app.GitHooksRequest{Repo: "/home/demo/repo", Hooks: []string{"gex-pre-commit"}})
			},
			avanca: goldenGitHooksEscreve,
			want:   []string{"documentado"}},

		{nome: "vitrine_indice", tiny: true,
			monta: func(_ *testing.T, c Common) tea.Model { return newShowcaseModel(c) },
			want:  []string{"Identidade"}},
		{nome: "vitrine_programa",
			monta:  func(_ *testing.T, c Common) tea.Model { return newShowcaseModel(c) },
			avanca: func(t *testing.T, m tea.Model) tea.Model { return goldenShowcaseAte(showcasePagePrograma)(t, m) },
			want:   []string{"Programa"}},
		{nome: "vitrine_medidas",
			monta:  func(_ *testing.T, c Common) tea.Model { return newShowcaseModel(c) },
			avanca: func(t *testing.T, m tea.Model) tea.Model { return goldenShowcaseAte(showcasePageMedidas)(t, m) },
			want:   []string{"Medidas"}},
		{nome: "vitrine_acessivel",
			monta:  func(_ *testing.T, c Common) tea.Model { return newShowcaseModel(c) },
			avanca: func(t *testing.T, m tea.Model) tea.Model { return goldenShowcaseAte(showcasePageAcessivel)(t, m) },
			want:   []string{"Acessível"}},
		{nome: "vitrine_cobertura",
			monta:  func(_ *testing.T, c Common) tea.Model { return newShowcaseModel(c) },
			avanca: func(t *testing.T, m tea.Model) tea.Model { return goldenShowcaseAte(showcasePageCobertura)(t, m) },
			want:   []string{"Mapa de cobertura"}},
	}
}

func TestGolden(t *testing.T) {
	for _, e := range goldenEstados() {
		if e.monta == nil {
			continue
		}
		tamanhos := []goldenTamanho{{nome: "80x24", w: 80, h: 24}, {nome: "60x20", w: 60, h: 20}}
		if e.tiny {
			tamanhos = append(tamanhos, goldenTamanho{nome: "30x10", w: 30, h: 10})
		}
		for _, tam := range tamanhos {
			t.Run(e.nome+"_"+tam.nome, func(t *testing.T) {
				goldenChecar(t, e, tam)
			})
		}
	}
}

func goldenChecar(t *testing.T, e goldenEstado, tam goldenTamanho) {
	t.Helper()
	m := goldenFecha(e.monta(t, goldenCommon()), tea.WindowSizeMsg{Width: tam.w, Height: tam.h}, tea.BackgroundColorMsg{Color: color.Black})
	if e.avanca != nil {
		m = e.avanca(t, m)
	}
	view := ansi.Strip(m.View().Content)
	for _, linha := range strings.Split(view, "\n") {
		if w := lipgloss.Width(linha); w > tam.w {
			t.Fatalf("linha com %d colunas passa do terminal de %d: %q", w, tam.w, linha)
		}
	}
	if h := lipgloss.Height(view); h > tam.h {
		t.Fatalf("tela com %d linhas passa do terminal de %d", h, tam.h)
	}
	want := e.want
	if tam.nome == "30x10" {
		want = []string{"Terminal"}
	}
	for _, texto := range want {
		if !strings.Contains(view, texto) {
			t.Fatalf("%q ausente da tela:\n%s", texto, view)
		}
	}
	if strings.Contains(view, goldenSentinela) {
		t.Fatalf("sentinela %s apareceu na tela", goldenSentinela)
	}
	golden.RequireEqual(t, []byte(view))
}
