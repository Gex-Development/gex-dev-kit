package screens

import (
	"image/color"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/golden"
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
		{
			nome: "vitrine_indice", tiny: true,
			monta: func(_ *testing.T, c Common) tea.Model { return newShowcaseModel(c) },
			want:  []string{"Identidade"},
		},
		{
			nome:   "vitrine_programa",
			monta:  func(_ *testing.T, c Common) tea.Model { return newShowcaseModel(c) },
			avanca: func(t *testing.T, m tea.Model) tea.Model { return goldenShowcaseAte(showcasePagePrograma)(t, m) },
			want:   []string{"Programa"},
		},
		{
			nome:   "vitrine_medidas",
			monta:  func(_ *testing.T, c Common) tea.Model { return newShowcaseModel(c) },
			avanca: func(t *testing.T, m tea.Model) tea.Model { return goldenShowcaseAte(showcasePageMedidas)(t, m) },
			want:   []string{"Medidas"},
		},
		{
			nome:   "vitrine_acessivel",
			monta:  func(_ *testing.T, c Common) tea.Model { return newShowcaseModel(c) },
			avanca: func(t *testing.T, m tea.Model) tea.Model { return goldenShowcaseAte(showcasePageAcessivel)(t, m) },
			want:   []string{"Acessível"},
		},
		{
			nome:   "vitrine_cobertura",
			monta:  func(_ *testing.T, c Common) tea.Model { return newShowcaseModel(c) },
			avanca: func(t *testing.T, m tea.Model) tea.Model { return goldenShowcaseAte(showcasePageCobertura)(t, m) },
			want:   []string{"Mapa de cobertura"},
		},
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
