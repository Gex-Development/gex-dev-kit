package screens

import (
	"regexp"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/app"
)

var chatGoldenDuration = regexp.MustCompile(`\b(?:\d+h \d+m|\d+m \d+s|\d+s)\b`)

type chatGoldenView struct{ tea.Model }

func (m chatGoldenView) View() tea.View {
	view := m.Model.View()
	view.Content = chatGoldenDuration.ReplaceAllString(view.Content, "0s")
	return view
}

func chatGoldenLoaded(c Common, deps ChatDeps) chatModel {
	deps.Dir = "/work/gex-dev-kit"
	m := newChatModel(c, deps, IntentHome)
	m, _ = chatTestUpdate(m, chatCoreMsg{status: app.CoreStatus{State: "não instalado", TotalCore: 7}})
	m, _ = chatTestUpdate(m, chatAgentsMsg{})
	return m
}

func chatGoldenMenu(t *testing.T, m chatModel) chatModel {
	t.Helper()
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyEnter})
	return chatSetupQuestion(t, m, "")
}

func chatGoldenSetup(t *testing.T, m chatModel) chatModel {
	t.Helper()
	m = chatGoldenMenu(t, m)
	m.input.SetValue("/setup")
	m, _ = chatTestUpdate(m, tea.KeyPressMsg{Code: tea.KeyEnter})
	return chatSetupQuestion(t, m, "Qual harness")
}

func chatGoldenPlan(t *testing.T, m chatModel) chatModel {
	t.Helper()
	m = chatGoldenSetup(t, m)
	m = chatSetupChoose(t, m, "omp")
	m = chatSetupQuestion(t, m, "Onde eu instalo?")
	m = chatSetupChoose(t, m, "current")
	m = chatSetupQuestion(t, m, "Além do Core")
	m = chatSetupChoose(t, m)
	return m
}

func chatGoldenFreeze(m chatModel) tea.Model {
	for i, entry := range m.entries {
		if entry.isTool && (m.running == nil || m.running.entry != i) {
			entry.tool.Elapsed = 0
		}
	}
	return chatGoldenView{m}
}

func TestGoldenChat(t *testing.T) {
	var planStarted <-chan struct{}
	states := []goldenEstado{
		{
			nome:   "boas_vindas",
			monta:  func(_ *testing.T, c Common) tea.Model { return chatGoldenLoaded(c, chatInstallFakeDeps()) },
			avanca: func(_ *testing.T, model tea.Model) tea.Model { return chatGoldenFreeze(model.(chatModel)) },
			want:   []string{"Assistente de setup do Kit GEX", "Pressione enter para começar"},
		},
		{
			nome:  "menu",
			monta: func(_ *testing.T, c Common) tea.Model { return chatGoldenLoaded(c, chatInstallFakeDeps()) },
			avanca: func(t *testing.T, model tea.Model) tea.Model {
				return chatGoldenFreeze(chatGoldenMenu(t, model.(chatModel)))
			},
			want: []string{"Digite / para ver o que posso fazer.", "pronto"},
		},
		{
			nome:  "setup_harness",
			monta: func(_ *testing.T, c Common) tea.Model { return chatGoldenLoaded(c, chatInstallFakeDeps()) },
			avanca: func(t *testing.T, model tea.Model) tea.Model {
				return chatGoldenFreeze(chatGoldenSetup(t, model.(chatModel)))
			},
			want: []string{"Qual harness você usa?", "Claude Code"},
		},
		{
			nome: "setup_extras",
			monta: func(_ *testing.T, c Common) tea.Model {
				deps := chatInstallFakeDeps()
				deps.Install.Options = func() []app.OptionGroup {
					return []app.OptionGroup{{ID: "design", Title: "Produto e Interface", Description: "Identidade visual", Options: []app.Option{{Value: "gex-design-system", Label: "Design System"}}}}
				}
				return chatGoldenLoaded(c, deps)
			},
			avanca: func(t *testing.T, model tea.Model) tea.Model {
				m := chatGoldenSetup(t, model.(chatModel))
				m = chatSetupChoose(t, m, "omp")
				m = chatSetupQuestion(t, m, "Onde eu instalo?")
				m = chatSetupChoose(t, m, "current")
				return chatGoldenFreeze(chatSetupQuestion(t, m, "Além do Core"))
			},
			want: []string{"Além do Core", "Produto e Interface (1 opção)", "Pular (nenhuma marcada)", "setup · Extras"},
		},
		{
			nome: "ferramenta_rodando",
			monta: func(t *testing.T, c Common) tea.Model {
				deps := chatInstallFakeDeps()
				started := make(chan struct{})
				planStarted = started
				release := make(chan struct{})
				done := make(chan struct{})
				deps.Install.Plan = func(req app.Request) (app.Plan, error) {
					defer close(done)
					close(started)
					<-release
					plan := chatSetupPlan()
					plan.Request = req
					return plan, nil
				}
				t.Cleanup(func() {
					close(release)
					select {
					case <-started:
						<-done
					default:
					}
				})
				return chatGoldenLoaded(c, deps)
			},
			avanca: func(t *testing.T, model tea.Model) tea.Model {
				m := chatGoldenPlan(t, model.(chatModel))
				for range 20 {
					if m.running != nil && strings.Contains(m.entries[m.running.entry].tool.Title, "Planejar instalação") {
						select {
						case <-planStarted:
						case <-time.After(2 * time.Second):
							t.Fatal("planejamento não iniciou goroutine")
						}
						m.common.Motion = true
						m.running.started = time.Now().Add(-1500 * time.Millisecond)
						return chatGoldenFreeze(m)
					}
					if len(m.queue) == 0 {
						t.Fatal("planejamento não entrou em execução")
					}
					m, _ = m.pump()
				}
				t.Fatal("planejamento não iniciou")
				return nil
			},
			want: []string{"GEX está trabalhando", "Selecionando as skills"},
		},
		{
			nome:  "confirmar",
			monta: func(_ *testing.T, c Common) tea.Model { return chatGoldenLoaded(c, chatInstallFakeDeps()) },
			avanca: func(t *testing.T, model tea.Model) tea.Model {
				m := chatGoldenPlan(t, model.(chatModel))
				return chatGoldenFreeze(chatSetupQuestion(t, m, "Aplico essas alterações?"))
			},
			want: []string{"Aplico essas alterações?", "Sim, aplicar"},
		},
		{
			nome:  "resultado",
			monta: func(_ *testing.T, c Common) tea.Model { return chatGoldenLoaded(c, chatInstallFakeDeps()) },
			avanca: func(t *testing.T, model tea.Model) tea.Model {
				m := chatGoldenPlan(t, model.(chatModel))
				m = chatSetupQuestion(t, m, "Aplico essas alterações?")
				m = chatSetupChoose(t, m, "apply")
				return chatGoldenFreeze(chatSetupQuestion(t, m, ""))
			},
			want: []string{"Pronto! Instalei", "pronto"},
		},
	}
	for _, state := range states {
		for _, size := range []goldenTamanho{{nome: "120x30", w: 120, h: 30}, {nome: "80x24", w: 80, h: 24}, {nome: "60x20", w: 60, h: 20}} {
			t.Run(state.nome+"_"+size.nome, func(t *testing.T) {
				goldenChecar(t, state, size)
			})
		}
	}
}
