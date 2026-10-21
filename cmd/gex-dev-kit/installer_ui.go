package main

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Gex-Development/gex-dev-kit/internal/agents"
	"github.com/Gex-Development/gex-dev-kit/internal/catalog"
	"github.com/Gex-Development/gex-dev-kit/internal/kit"
	"github.com/charmbracelet/x/ansi"
)

type installPhase uint8

const (
	installPlanning installPhase = iota
	installPreview
	installConfirm
	installApplying
	installDone
	installFailed
	installCancelled
)

type installPrepared struct {
	plan    *kit.Plan
	changes []agents.Change
	updates []agents.Change
	blocked bool
	err     error
}

type installApplied struct {
	id            string
	agentsUpdated bool
	err           error
}

type installModel struct {
	home, executable string
	skills, hooks    []string
	skipAgentUpdate  bool
	phase            installPhase
	prepared         installPrepared
	result           installApplied
	viewport         viewport.Model
	spinner          spinner.Model
	width, height    int
	dark             bool
}

func newInstallModel(home, executable string, skills, hooks []string, skipAgentUpdate bool) installModel {
	return installModel{
		home: home, executable: executable, skills: skills, hooks: hooks, skipAgentUpdate: skipAgentUpdate,
		phase: installPlanning, width: 80, height: 24, dark: true,
		viewport: viewport.New(viewport.WithWidth(72), viewport.WithHeight(12)),
		spinner:  spinner.New(spinner.WithSpinner(spinner.MiniDot), spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#7fc3ef")))),
	}
}

func (m installModel) prepare() tea.Msg {
	files, err := catalog.Build(m.home, m.executable, m.skills, m.hooks)
	if err != nil {
		return installPrepared{err: fmt.Errorf("catálogo: %w", err)}
	}
	plan, err := kit.Preview(m.home, files)
	if err != nil {
		return installPrepared{err: fmt.Errorf("prévia: %w", err)}
	}
	prepared := installPrepared{plan: &plan}
	if m.skipAgentUpdate {
		return prepared
	}
	prepared.changes, err = agents.Preview(m.home)
	if err != nil {
		prepared.err = fmt.Errorf("prévia dos agentes: %w", err)
		return prepared
	}
	for _, change := range prepared.changes {
		if change.Blocked {
			prepared.blocked = true
		} else if change.Current != change.Target {
			prepared.updates = append(prepared.updates, change)
		}
	}
	return prepared
}

func (m installModel) apply() tea.Msg {
	result := installApplied{}
	if len(m.prepared.updates) > 0 {
		if err := agents.Apply(m.home, m.prepared.updates); err != nil {
			result.err = fmt.Errorf("atualização de agentes falhou; Core não foi aplicado: %w", err)
			return result
		}
		result.agentsUpdated = true
	}
	if len(m.prepared.plan.Changes) > 0 {
		id, err := kit.Apply(*m.prepared.plan)
		if err != nil {
			result.err = fmt.Errorf("instalação do Core: %w", err)
			return result
		}
		result.id = id
	}
	return result
}

func (m installModel) Init() tea.Cmd {
	return tea.Batch(m.prepare, m.spinner.Tick, tea.RequestBackgroundColor)
}

func (m *installModel) resize() {
	m.viewport.SetWidth(max(1, min(m.width, 76)-4))
	m.viewport.SetHeight(max(1, m.height-8))
	if m.prepared.plan != nil {
		m.viewport.SetContent(ansi.Hardwrap(previewText(m.prepared.plan.Changes, m.prepared.changes), m.viewport.Width(), false))
	}
}

func (m installModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = max(1, msg.Width), max(1, msg.Height)
		m.resize()
	case tea.BackgroundColorMsg:
		m.dark = msg.IsDark()
		m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.LightDark(m.dark)(lipgloss.Color("#1a5fc0"), lipgloss.Color("#7fc3ef")))
	case installPrepared:
		m.prepared = msg
		if msg.err != nil {
			m.phase = installFailed
		} else if len(msg.plan.Changes) == 0 && len(msg.updates) == 0 && !msg.blocked {
			m.phase = installDone
		} else {
			m.phase = installPreview
			m.resize()
		}
	case installApplied:
		m.result = msg
		if msg.err != nil {
			m.phase = installFailed
		} else {
			m.phase = installDone
		}
	case spinner.TickMsg:
		if m.phase == installPlanning || m.phase == installApplying {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	case tea.KeyPressMsg:
		key := msg.String()
		switch m.phase {
		case installPlanning:
			if key == "esc" || key == "q" || key == "ctrl+c" {
				m.phase = installCancelled
				return m, tea.Quit
			}
			return m, nil
		case installApplying:
			return m, nil
		case installPreview:
			switch key {
			case "esc", "q", "ctrl+c":
				m.phase = installCancelled
				return m, tea.Quit
			case "c", "enter":
				m.phase = installConfirm
				return m, nil
			}
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		case installConfirm:
			switch key {
			case "y", "s":
				m.phase = installApplying
				return m, tea.Batch(m.apply, m.spinner.Tick)
			case "n", "esc", "q", "ctrl+c":
				m.phase = installPreview
				return m, nil
			}
		case installDone, installFailed, installCancelled:
			if key == "enter" || key == "esc" || key == "q" || key == "ctrl+c" {
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m installModel) View() tea.View {
	width := max(1, min(m.width, 80))
	accent := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.LightDark(m.dark)(lipgloss.Color("#1a5fc0"), lipgloss.Color("#7fc3ef")))
	muted := lipgloss.NewStyle().Foreground(lipgloss.LightDark(m.dark)(lipgloss.Color("#4b5669"), lipgloss.Color("#a1a1aa")))
	lines := []string{accent.Render("GEX / DEV KIT"), muted.Render("INSTALAÇÃO  /  CORE + OPCIONAIS"), ""}
	switch m.phase {
	case installPlanning:
		lines = append(lines, m.spinner.View()+" Lendo catálogo e estado dos arquivos…", "", "Nenhuma alteração aplicada.")
	case installApplying:
		step := "Aplicando Core e registrando histórico…"
		if len(m.prepared.updates) > 0 {
			step = "Atualizando agentes elegíveis; depois Core…"
		}
		lines = append(lines, m.spinner.View()+" "+step, "", "Não feche o terminal durante a operação.")
	case installPreview, installConfirm:
		files := len(m.prepared.plan.Changes)
		lines = append(lines, fmt.Sprintf("%d arquivo(s) · %d agente(s) a atualizar", files, len(m.prepared.updates)))
		if m.prepared.blocked {
			lines = append(lines, "Agentes bloqueados exigem ação manual; Core pode prosseguir.")
		}
		lines = append(lines, "", m.viewport.View(), "", muted.Render("↑↓/PgUp/PgDn rolar · c revisar · Esc cancelar"))
	case installDone:
		if m.result.id == "" && !m.result.agentsUpdated {
			lines = append(lines, "Nada a alterar.")
		}
		if m.result.agentsUpdated {
			lines = append(lines, "Agentes atualizados.")
		}
		if m.result.id != "" {
			lines = append(lines, "Core instalado. Histórico: "+m.result.id)
		}
		if len(m.hooks) > 0 && m.result.id != "" {
			lines = append(lines, "Codex: aprove hook novo em /hooks para ativá-lo.")
		}
		if m.prepared.blocked {
			lines = append(lines, "Agentes bloqueados não foram atualizados.")
		}
		lines = append(lines, "", muted.Render("Enter para sair"))
	case installFailed:
		failure := m.prepared.err
		if m.result.err != nil {
			failure = m.result.err
		}
		if m.result.agentsUpdated {
			lines = append(lines, "Agentes atualizados antes da falha do Core.")
		}
		lines = append(lines, "Falha: "+failure.Error(), "", muted.Render("Enter para sair"))
	case installCancelled:
		lines = append(lines, "Instalação cancelada; nenhum arquivo aplicado.")
	}
	content := strings.Join(lines, "\n")
	if m.phase == installConfirm {
		modalWidth := max(12, min(width-4, 49))
		modal := lipgloss.NewStyle().Width(modalWidth-4).Padding(0, 1).Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.LightDark(m.dark)(lipgloss.Color("#1a5fc0"), lipgloss.Color("#7fc3ef"))).
			Render("APLICAR ALTERAÇÕES?\n\nRevise arquivos acima antes de continuar.\n\n[S] Sim   [N] Voltar")
		if m.width >= 24 && m.height >= 11 {
			content = lipgloss.NewCompositor(lipgloss.NewLayer(content), lipgloss.NewLayer(modal).X(max(0, (width-lipgloss.Width(modal))/2)).Y(max(0, (m.height-lipgloss.Height(modal))/2)).Z(1)).Render()
		} else {
			content = modal
		}
	}
	view := tea.NewView(homeFrame(strings.Split(lipgloss.Wrap(content, width, ""), "\n"), width, m.height))
	view.AltScreen = true
	return view
}

func guidedInstall(home, executable string, skills, hooks []string, skipAgentUpdate bool, in io.Reader, out, errOut io.Writer) int {
	model, err := tea.NewProgram(newInstallModel(home, executable, skills, hooks, skipAgentUpdate), tea.WithInput(in), tea.WithOutput(out)).Run()
	if err != nil {
		fmt.Fprintln(errOut, "Instalação: interface indisponível:", err)
		return 1
	}
	result := model.(installModel)
	if result.phase == installFailed {
		failure := result.prepared.err
		if result.result.err != nil {
			failure = result.result.err
		}
		fmt.Fprintln(errOut, "Instalação:", failure)
		return 1
	}
	if result.prepared.blocked && result.phase == installDone {
		fmt.Fprintln(errOut, "Agentes bloqueados não foram atualizados; demais alterações aplicadas.")
		return 1
	}
	return 0
}
