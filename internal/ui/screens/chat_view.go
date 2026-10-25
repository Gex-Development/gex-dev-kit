package screens

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/Gex-Development/gex-dev-kit/internal/catalog"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/components"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/layout"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

type chatHit struct{ y, entry int }

func (m chatModel) View() tea.View {
	content, _ := chatLayout(m)
	view := tea.NewView(content)
	view.AltScreen = true
	view.ReportFocus = true
	view.MouseMode = tea.MouseModeCellMotion
	if m.paletteOpen || m.ask != nil && (m.ask.hasChoice || m.ask.path) {
		view.MouseMode = tea.MouseModeAllMotion
	}
	return view
}

func choiceForLayout(choice components.Choice, width, height int) components.Choice {
	for _, art := range []bool{true, false} {
		for rows := height - 11; rows >= 1; rows-- {
			candidate := choice.WithArt(art).MaxRows(rows)
			fits := true
			page := candidate
			for range candidate.Page() {
				page, _ = page.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
			}
			for range candidate.Pages() {
				if page.Height(width)+1 > height-11 {
					fits = false
					break
				}
				page, _ = page.Update(tea.KeyPressMsg{Code: tea.KeyRight})
			}
			if fits {
				return candidate
			}
		}
	}
	return choice.WithArt(false).MaxRows(1)
}

func (m *chatModel) layoutList(width int) int {
	limit := max(1, m.height-11)
	if m.paletteOpen {
		m.palette = m.palette.MaxRows(limit)
		if m.palette.Height(width) > limit {
			m.palette = m.palette.MaxRows(limit - 1)
		}
		return m.palette.Height(width)
	}
	if m.ask == nil {
		return 0
	}
	if m.ask.hasChoice {
		m.ask.choice = choiceForLayout(m.ask.choice, width, m.height)
		return m.ask.choice.Height(width)
	}
	if m.ask.path {
		m.ask.picker = m.ask.picker.MaxRows(max(1, limit-1))
		if m.ask.picker.Height(width) > limit-1 {
			m.ask.picker = m.ask.picker.MaxRows(limit - 2)
		}
		return m.ask.picker.Height(width)
	}
	return 0
}

func (m *chatModel) placeChoice() {
	mainW := m.width
	if m.width >= 120 && m.height >= 30 {
		mainW -= 33
	}
	listH := m.layoutList(mainW)
	transcriptH := max(0, m.height-7-listH)
	if m.ask != nil && m.ask.path && !m.paletteOpen {
		transcriptH--
	}
	y := transcriptH + 2
	if m.paletteOpen {
		m.palette = m.palette.Place(0, y-1)
	} else if m.ask != nil && m.ask.hasChoice {
		m.ask.choice = m.ask.choice.Place(0, y, mainW)
	} else if m.ask != nil && m.ask.path {
		m.ask.picker = m.ask.picker.Place(0, y)
	}
}

func chatLayout(m chatModel) (string, []chatHit) {
	w, h := m.width, m.height
	if layout.Classify(layout.Size{Width: w, Height: h}) == layout.ClassTiny {
		return layout.TooSmall(m.theme, layout.Size{Width: w, Height: h}), nil
	}
	if m.phase == chatWelcome {
		mark := lipgloss.PlaceHorizontal(w, lipgloss.Center, components.Wordmark(m.theme, w, m.reveal.Value()))
		rows := []string{mark, lipgloss.PlaceHorizontal(w, lipgloss.Center, m.theme.Secondary.Render("Assistente de setup do Kit GEX"))}
		if !m.coreDone || !m.agentsDone {
			rows = append(rows, lipgloss.PlaceHorizontal(w, lipgloss.Center, m.spin.View()+" Verificando seu ambiente…"))
		} else {
			rows = append(rows, "")
		}
		if w >= 50 {
			rows = append(rows, lipgloss.PlaceHorizontal(w, lipgloss.Center, m.theme.Faint.Render("Dica: digite / para ver todos os comandos.")))
		}
		if m.coreDone && m.agentsDone {
			rows = append(rows, lipgloss.PlaceHorizontal(w, lipgloss.Center, m.theme.Glow.Render("Pressione enter para começar")))
		}
		return layout.Center(layout.Size{Width: w, Height: h}, strings.Join(rows, "\n")), nil
	}
	wide := w >= 120 && h >= 30
	mainW, sideW := w, 0
	if wide {
		sideW = 32
		mainW = w - sideW - 1
	}
	choiceH := m.layoutList(mainW)
	bottomH := 2
	if choiceH > 0 {
		bottomH += choiceH + 1
	}
	if m.ask != nil && m.ask.textarea {
		bottomH += 3
	} else if m.ask != nil && m.ask.input && !m.ask.hasChoice {
		bottomH++
	}
	transcriptH := max(0, h-3-(bottomH+1))
	var transcript []string
	var hits []chatHit
	for i, e := range m.entries {
		if i > 0 {
			transcript = append(transcript, "")
		}
		if e.banner != "" {
			transcript = append(transcript, strings.Split(components.HarnessBanner(m.theme, mainW, e.banner, e.info), "\n")...)
		} else if e.isTool {
			hit := len(transcript)
			b := *e.tool
			if m.running != nil && m.running.entry == i {
				b.Spinner = m.spin.View()
				b.Elapsed = time.Since(m.running.started)
				if m.common.Motion {
					b.Phases = m.running.phases
					b.PhaseRows = m.running.completion != nil && len(m.running.completion.result.rows) == len(b.Phases)
				}
			}
			transcript = append(transcript, strings.Split(b.View(m.theme, mainW), "\n")...)
			hits = append(hits, chatHit{y: hit, entry: i})
		} else {
			shown := e.text
			runes := e.runes
			if e.shown < len(runes) || runes == nil && e.shown < len([]rune(shown)) {
				if runes == nil {
					runes = []rune(shown)
				}
				shown = string(runes[:e.shown]) + "▍"
			}
			transcript = append(transcript, strings.Split(components.ChatMessage(m.theme, mainW, e.role, shown), "\n")...)
		}
	}
	start := max(0, len(transcript)-transcriptH-m.scroll)
	if start+transcriptH > len(transcript) {
		start = max(0, len(transcript)-transcriptH)
	}
	end := min(len(transcript), start+transcriptH)
	visibleHits := hits[:0]
	for _, hit := range hits {
		hit.y -= start
		if hit.y >= 0 && hit.y < transcriptH {
			visibleHits = append(visibleHits, hit)
		}
	}
	hits = visibleHits
	visible := make([]string, transcriptH)
	copy(visible, transcript[start:end])
	for i := range visible {
		visible[i] = ansi.Truncate(visible[i], mainW, "")
	}
	work := ""
	if m.locked() {
		label := "GEX está escrevendo…"
		if m.beat > 0 {
			label = "GEX está pensando…"
		}
		if m.running != nil {
			elapsed := time.Since(m.running.started)
			label = chatWorkLabel(m.running.phases, elapsed)
			if m.running.completion != nil && m.running.completion.fill != nil {
				label = "Lendo o ambiente…"
			}
		}
		hint := "ctrl+c sair"
		if m.Busy() {
			hint = "gravando — não feche o terminal"
		}
		work = components.WorkRow(m.theme, mainW, m.spin.View(), label, time.Since(m.turnStart), hint)
	}
	bottom := []string{work}
	if m.paletteOpen {
		m.palette = m.palette.Place(0, transcriptH+1)
		bottom = append(bottom, strings.Split(m.palette.View(m.theme, mainW), "\n")...)
		bottom = append(bottom, "")
	} else if m.ask != nil && m.ask.hasChoice {
		m.ask.choice = m.ask.choice.Place(0, transcriptH+2, mainW)
		bottom = append(bottom, strings.Split(m.ask.choice.View(m.theme, mainW), "\n")...)
		bottom = append(bottom, "")
	} else if m.ask != nil && m.ask.path {
		bottom = append(bottom, m.theme.Strong.Render(ansi.Truncate("? "+m.ask.question, mainW, "")))
		m.ask.picker = m.ask.picker.Place(0, transcriptH+2)
		bottom = append(bottom, strings.Split(m.ask.picker.View(m.theme, mainW), "\n")...)
		bottom = append(bottom, "")
	} else if m.ask != nil && (m.ask.input || m.ask.textarea) {
		bottom = append(bottom, m.theme.Strong.Render(ansi.Truncate("? "+m.ask.question, mainW, "")))
	}
	promptView := m.input.View()
	if m.ask != nil && m.ask.textarea {
		promptView = m.area.View()
	}
	segments, metrics := m.statusContent()
	bottom = append(bottom, components.PromptBox(m.theme, mainW, segments, metrics, promptView, m.locked(), "GEX está trabalhando…")...)
	mainRows := append(visible, bottom...)
	if wide {
		sections := []components.SidebarSection{{Title: "AMBIENTE", Items: m.environmentItems()}}
		if len(m.steps) > 0 {
			items := make([]components.ChecklistItem, 0, len(m.steps))
			for i, step := range m.steps {
				state := theme.StatePending
				if i < m.step {
					state = theme.StateOK
				} else if i == m.step {
					state = theme.StateRunning
				}
				items = append(items, components.ChecklistItem{Label: step, State: state})
			}
			sections = append(sections, components.SidebarSection{Title: "ETAPAS", Items: items})
		}
		sidebar := strings.Split(components.Sidebar(m.theme, sideW, h-1, sections), "\n")
		for i := range mainRows {
			right := ""
			if i < len(sidebar) {
				right = sidebar[i]
			}
			left := ansi.Truncate(mainRows[i], mainW, "")
			mainRows[i] = left + strings.Repeat(" ", max(0, mainW-ansi.StringWidth(left))) + " " + right
		}
	}

	hints := "/ comandos · ctrl+o expandir · ctrl+c sair"
	if m.locked() {
		hints = "enter pula a animação · ctrl+c sair"
		if m.theme.Plain {
			hints = "enter pula a animacao - ctrl+c sair"
		}
	} else if m.ask != nil && m.ask.path {
		hints = "tab completa · ↑/↓ pasta · enter confirma · esc volta"
		if strings.Contains(m.ask.question, "repositório Git") {
			hints += " · ● = repositório git"
		}
	} else if m.ask != nil && (m.ask.input || m.ask.textarea) {
		hints = "enter confirma · esc voltar"
	} else if m.ask != nil && m.ask.hasChoice {
		hints = "↑/↓ mover · enter escolhe · esc voltar"
		if m.ask.choice.Multi() {
			hints = "espaço marca · enter confirma · esc voltar"
		}
		if m.ask.choice.Pages() > 1 {
			hints = "↑/↓ mover · ←/→ página · enter escolhe · esc voltar"
			if m.ask.choice.Multi() {
				hints = "espaço marca · ←/→ página · enter confirma · esc voltar"
			}
		}
		if m.theme.Plain {
			hints = "up/down mover | enter escolhe | esc voltar"
			if m.ask.choice.Multi() {
				hints = "espaco marca | enter confirma | esc voltar"
			}
			if m.ask.choice.Pages() > 1 {
				hints = "up/down mover | left/right pagina | enter escolhe | esc voltar"
				if m.ask.choice.Multi() {
					hints = "espaco marca | left/right pagina | enter confirma | esc voltar"
				}
			}
		}
	} else if m.paletteOpen {
		hints = "↑/↓ mover · tab completa · enter escolhe · esc fecha"
	}
	footer := layout.Columns(m.theme, w, m.theme.Secondary.Render(m.context), m.theme.Faint.Render(hints))
	rows := mainRows
	rows = append(rows, footer)
	return layout.Fit(strings.Join(rows, "\n"), w, h), hits
}

func (m chatModel) statusContent() ([]components.StatusSegment, []components.StatusMetric) {
	segments := []components.StatusSegment{{Icon: "◆", Text: "GEX", Tone: components.ToneBrand}}
	hasHarness := false
	for _, a := range m.agents {
		if a.Current == "" {
			continue
		}
		name := "Claude Code"
		if a.Name == "codex" {
			name = "Codex"
		} else if a.Name == "omp" {
			name = "OMP"
		}
		segments = append(segments, components.StatusSegment{Icon: "●", Text: name + " " + a.Current, Tone: components.ToneHarness, Brand: a.Name})
		hasHarness = true
		break
	}
	if !hasHarness && len(m.harnessIDs) > 0 {
		names := make([]string, 0, len(m.harnessIDs))
		for _, id := range m.harnessIDs {
			switch id {
			case "claude":
				names = append(names, "Claude Code")
			case "codex":
				names = append(names, "Codex")
			case "omp":
				names = append(names, "OMP")
			}
		}
		if len(names) > 0 {
			label := names[0]
			if len(names) == 2 {
				label += " + " + names[1]
			} else if len(names) > 2 {
				label += fmt.Sprintf(" +%d", len(names)-1)
			}
			segments = append(segments, components.StatusSegment{Icon: "●", Text: label, Tone: components.ToneHarness, Brand: m.harnessIDs[0]})
		}
	}
	if m.deps.Dir != "" {
		segments = append(segments, components.StatusSegment{Icon: "\uf07c", Text: filepath.Base(filepath.Clean(m.deps.Dir)), Tone: components.ToneDir})
	}
	if m.gitOK {
		segments = append(segments, components.StatusSegment{Icon: "\uf418", Text: m.git.Branch, Tone: components.ToneGit, Modified: m.git.Modified, Untracked: m.git.Untracked})
	}
	home := m.request.Home
	if home == "" {
		home = m.common.Home
	}
	if home != "" && m.common.UserHome != "" && filepath.Clean(home) != filepath.Clean(m.common.UserHome) {
		segments = append(segments, components.StatusSegment{Icon: "\uf015", Text: homeLabel(home, m.common.UserHome), Tone: components.ToneWarning})
	}
	metrics := make([]components.StatusMetric, 0, 2)
	if len(m.steps) > 0 {
		pct := (100*(m.step+1) + len(m.steps)/2) / len(m.steps)
		metrics = append(metrics, components.StatusMetric{Text: fmt.Sprintf("%d%%", pct), Tone: components.ToneBrand}, components.StatusMetric{Text: fmt.Sprintf("Etapa %d/%d", m.step+1, len(m.steps)), Tone: components.ToneFaint})
	} else {
		tone := components.ToneMuted
		if m.coreDone && m.coreErr == nil {
			switch m.core.State {
			case "íntegro":
				tone = components.ToneSuccess
			case "não instalado":
			default:
				tone = components.ToneWarning
			}
		}
		metrics = append(metrics, components.StatusMetric{Text: fmt.Sprintf("Core %d/%d", m.core.ActiveCore, len(catalog.CoreSkills())), Tone: tone})
		if m.common.Version != "" {
			metrics = append(metrics, components.StatusMetric{Text: "v" + strings.TrimPrefix(m.common.Version, "v"), Tone: components.ToneFaint})
		}
	}
	return segments, metrics
}

func (m chatModel) environmentItems() []components.ChecklistItem {
	items := make([]components.ChecklistItem, 0, 3)
	for _, name := range []string{"claude", "codex"} {
		label := "Claude Code"
		if name == "codex" {
			label = "Codex"
		}
		item := components.ChecklistItem{Label: label, State: theme.StateRunning, Detail: "lendo…"}
		if m.agentsDone {
			item.State = theme.StateInfo
			item.Detail = "não detectado"
			if change, ok := homeAgentChange(m.agents, name); ok {
				item.State = homeAgentState(change)
				item.Detail = homeAgentText(change, m.common.UserHome)
			}
		}
		items = append(items, item)
	}
	core := components.ChecklistItem{Label: "Core", State: theme.StateRunning, Detail: "lendo…"}
	if m.coreDone {
		core.State = homeCoreState(m.core.State)
		core.Detail = fmt.Sprintf("%d/%d skills · %s", m.core.ActiveCore, len(catalog.CoreSkills()), m.core.State)
	}
	return append(items, core)
}
