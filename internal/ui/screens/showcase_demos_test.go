package screens

import (
	"strings"
	"testing"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
)

type demoCaso struct {
	nome string
	fn   func(theme.Theme, int) string
}

func demoTodosOsCasos() []demoCaso {
	return []demoCaso{
		{"demoSpinner", demoSpinner}, {"demoProgress", demoProgress}, {"demoBubblesTable", demoBubblesTable},
		{"demoList", demoList}, {"demoTextInput", demoTextInput}, {"demoTextArea", demoTextArea},
		{"demoViewport", demoViewport}, {"demoPaginator", demoPaginator}, {"demoHelp", demoHelp},
		{"demoFilePicker", demoFilePicker}, {"demoTimer", demoTimer}, {"demoStopwatch", demoStopwatch},
		{"demoCursor", demoCursor}, {"demoHuhInput", demoHuhInput}, {"demoHuhText", demoHuhText},
		{"demoHuhSelect", demoHuhSelect}, {"demoHuhMultiSelect", demoHuhMultiSelect}, {"demoHuhConfirm", demoHuhConfirm},
		{"demoHuhNote", demoHuhNote}, {"demoHuhFilePicker", demoHuhFilePicker}, {"demoHuhGroups", demoHuhGroups},
		{"demoHuhValidation", demoHuhValidation}, {"demoHuhTheme", demoHuhTheme}, {"demoHuhAccessible", demoHuhAccessible},
		{"demoJoin", demoJoin}, {"demoPlace", demoPlace}, {"demoBorders", demoBorders},
		{"demoGradient", demoGradient}, {"demoLipglossTable", demoLipglossTable}, {"demoLipglossList", demoLipglossList},
		{"demoTree", demoTree}, {"demoAdaptive", demoAdaptive}, {"demoMeasures", demoMeasures},
		{"demoSpring", demoSpring}, {"demoProjectile", demoProjectile}, {"demoAnsi", demoAnsi},
		{"demoTerm", demoTerm}, {"demoTick", demoTick}, {"demoWindowSize", demoWindowSize},
		{"demoAltScreen", demoAltScreen}, {"demoMouse", demoMouse}, {"demoFocus", demoFocus},
		{"demoBatchSequence", demoBatchSequence}, {"demoPrintln", demoPrintln}, {"demoSuspend", demoSuspend},
		{"demoTypedErrors", demoTypedErrors}, {"demoWordmark", demoWordmark}, {"demoBanner", demoBanner},
		{"demoPanel", demoPanel}, {"demoBadge", demoBadge}, {"demoChecklist", demoChecklist},
		{"demoSteps", demoSteps}, {"demoToast", demoToast}, {"demoKeyHints", demoKeyHints},
		{"demoDiffView", demoDiffView}, {"demoFileTree", demoFileTree}, {"demoDSTable", demoDSTable},
		{"demoEmptyState", demoEmptyState}, {"demoErrorBox", demoErrorBox}, {"demoConfirm", demoConfirm},
		{"demoMenu", demoMenu}, {"demoLoading", demoLoading}, {"demoSection", demoSection},
	}
}

func demoTemasDeTeste() map[string]theme.Theme {
	return map[string]theme.Theme{
		"escuro": theme.New(theme.Options{Dark: true, Profile: colorprofile.TrueColor}),
		"claro":  theme.New(theme.Options{Profile: colorprofile.TrueColor}),
		"plain":  theme.New(theme.Options{Profile: colorprofile.ASCII, Plain: true}),
	}
}

func TestDemoMarcadorAparecePorTema(t *testing.T) {
	for _, c := range CoverageMap {
		if c.TestOnly {
			continue
		}
		for nome, tema := range demoTemasDeTeste() {
			saida := ansi.Strip(c.Demo(tema, 80))
			if !strings.Contains(saida, c.Marker) {
				t.Fatalf("demo de %s sem o marcador %q no tema %s:\n%s", c.Feature, c.Marker, nome, saida)
			}
		}
	}
}

func TestDemoLarguraEAlturaRespeitadas(t *testing.T) {
	for _, largura := range []int{40, 80} {
		for _, caso := range demoTodosOsCasos() {
			for nome, tema := range demoTemasDeTeste() {
				linhas := strings.Split(caso.fn(tema, largura), "\n")
				if len(linhas) > 12 {
					t.Fatalf("%s (%s, largura %d) tem %d linhas; o limite é 12", caso.nome, nome, largura, len(linhas))
				}
				for _, linha := range linhas {
					if w := ansi.StringWidth(linha); w > largura {
						t.Fatalf("%s (%s, largura %d) tem linha com %d colunas: %q", caso.nome, nome, largura, w, ansi.Strip(linha))
					}
				}
			}
		}
	}
}

func TestDemoSaidaNaoVazia(t *testing.T) {
	tema := theme.New(theme.Options{Dark: true, Profile: colorprofile.TrueColor})
	for _, caso := range demoTodosOsCasos() {
		if strings.TrimSpace(ansi.Strip(caso.fn(tema, 80))) == "" {
			t.Fatalf("%s devolveu saída vazia", caso.nome)
		}
	}
}

func TestDemoLarguraMinimaSemPanico(t *testing.T) {
	for nome, tema := range demoTemasDeTeste() {
		for _, caso := range demoTodosOsCasos() {
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("%s (%s) entrou em pânico com largura 10: %v", caso.nome, nome, r)
					}
				}()
				saida := caso.fn(tema, 10)
				for _, linha := range strings.Split(saida, "\n") {
					if w := ansi.StringWidth(linha); w > 10 {
						t.Fatalf("%s (%s) tem linha com %d colunas na largura 10", caso.nome, nome, w)
					}
				}
			}()
		}
	}
}

func TestDemoFitCortaLinhasELimitaAltura(t *testing.T) {
	entrada := strings.Repeat("abcdefghij\n", 20)
	saida := demoFit(entrada, 4)
	linhas := strings.Split(saida, "\n")
	if len(linhas) != 12 {
		t.Fatalf("demoFit deveria limitar a 12 linhas, veio %d", len(linhas))
	}
	for _, linha := range linhas {
		if w := ansi.StringWidth(linha); w > 4 {
			t.Fatalf("demoFit deveria cortar em 4 colunas, veio %d: %q", w, linha)
		}
	}
	if demoFit("abc", 0) != "" {
		t.Fatalf("demoFit com largura 0 deveria devolver vazio")
	}
}

func TestDemoBarraLimitaPreenchimento(t *testing.T) {
	if got := demoBarra(30, 4); got != "████" {
		t.Fatalf("demoBarra com preenchimento acima do total: %q", got)
	}
	if got := demoBarra(-1, 3); got != "░░░" {
		t.Fatalf("demoBarra com preenchimento negativo: %q", got)
	}
}

func TestDemoCaminhoAbsolutoValida(t *testing.T) {
	if err := demoCaminhoAbsoluto("relativo/pasta"); err == nil {
		t.Fatalf("caminho relativo deveria ser recusado")
	}
	if err := demoCaminhoAbsoluto(t.TempDir()); err != nil {
		t.Fatalf("caminho absoluto recusado: %v", err)
	}
}

func TestDemoItemExpoeTitulo(t *testing.T) {
	item := demoItem{"tdd", "Matt · ciclo"}
	if item.Title() != "tdd" || item.Description() != "Matt · ciclo" || item.FilterValue() != "tdd" {
		t.Fatalf("demoItem devolveu valores inesperados")
	}
}
