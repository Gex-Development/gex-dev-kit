package screens

import (
	"fmt"
	"os"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/paginator"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"

	"github.com/Gex-Development/gex-dev-kit/internal/catalog"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/components"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/layout"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

const (
	addonsTitle      = "Addons"
	addonsHelpTitle  = "Ajuda"
	addonsDisclaimer = "O kit não executa este comando. Revise fonte e licença antes de instalar."
	addonsFooter     = "←/→ ou h/l troca de addon · ↑/↓ rola · ? ajuda · esc sai"
	addonsPanelRows  = 4
	addonsPanelCols  = 4
	addonsShortSha   = 7
)

func RunAddons(r Runner, c Common) error {
	m := newAddonsModel(c, catalog.Addons())
	if _, err := r.Run(m); err != nil {
		return err
	}
	if m.interrupted {
		return SignalError{Signal: os.Interrupt}
	}
	return nil
}

type addonsModel struct {
	common      Common
	theme       theme.Theme
	keys        components.Keys
	addons      []catalog.Addon
	pager       paginator.Model
	view        viewport.Model
	size        layout.Size
	help        bool
	interrupted bool
}

func newAddonsModel(c Common, addons []catalog.Addon) *addonsModel {
	pager := paginator.New(paginator.WithTotalPages(len(addons)), paginator.WithPerPage(1))
	pager.Type = paginator.Dots
	pager.ActiveDot = "●"
	pager.InactiveDot = "○"
	m := &addonsModel{
		common: c,
		theme:  addonsTheme(c),
		keys:   components.DefaultKeys(),
		addons: addons,
		pager:  pager,
		view:   viewport.New(),
	}
	m.refresh()
	return m
}

func addonsTheme(c Common) theme.Theme {
	return theme.New(theme.Options{Dark: c.Dark, Profile: c.Profile, Plain: c.Profile <= colorprofile.ASCII})
}

func (m *addonsModel) Init() tea.Cmd {
	return nil
}

func (m *addonsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.size = layout.Size{Width: msg.Width, Height: msg.Height}
		m.refresh()
	case tea.BackgroundColorMsg:
		m.common.Dark = msg.IsDark()
		m.theme = addonsTheme(m.common)
		m.refresh()
	case tea.ColorProfileMsg:
		m.common.Profile = msg.Profile
		m.theme = addonsTheme(m.common)
		m.refresh()
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *addonsModel) handleKey(press tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if press.String() == "ctrl+c" {
		m.interrupted = true
		return m, tea.Quit
	}
	if key.Matches(press, m.keys.Quit, m.keys.Back) {
		return m, tea.Quit
	}
	if layout.Classify(m.size) == layout.ClassTiny {
		return m, nil
	}
	switch {
	case key.Matches(press, m.keys.Help):
		m.help = !m.help
		m.resetPage()
	case key.Matches(press, m.keys.Left):
		m.pager.PrevPage()
		m.resetPage()
	case key.Matches(press, m.keys.Right):
		m.pager.NextPage()
		m.resetPage()
	default:
		var cmd tea.Cmd
		m.view, cmd = m.view.Update(press)
		return m, cmd
	}
	return m, nil
}

func (m *addonsModel) resetPage() {
	m.view.GotoTop()
	m.refresh()
}

func (m *addonsModel) refresh() {
	body := layout.BodySize(m.size)
	m.view.SetWidth(max(body.Width-addonsPanelCols, 1))
	m.view.SetHeight(max(body.Height-addonsPanelRows, 1))
	m.view.SetContent(m.content(max(body.Width-addonsPanelCols, 1)))
}

func (m *addonsModel) content(width int) string {
	if m.help {
		return addonsHelpText(m.keys)
	}
	if len(m.addons) == 0 {
		return ""
	}
	return addonsCard(m.theme, m.addons[m.pager.Page], width)
}

func (m *addonsModel) View() tea.View {
	v := tea.NewView(layout.Render(m.theme, m.size, layout.Frame{
		Title:   addonsTitle,
		Version: m.common.Version,
		Body:    m.body(),
		Footer:  addonsFooter,
	}))
	v.AltScreen = true
	v.MouseMode = tea.MouseModeNone
	return v
}

func (m *addonsModel) body() string {
	body := layout.BodySize(m.size)
	counter := fmt.Sprintf("%d de %d", m.pager.Page+1, m.pager.TotalPages)
	header := m.pager.View() + "  " + m.theme.Secondary.Render(counter)
	return header + "\n" + components.Panel(m.theme, components.PanelOptions{
		Title:   m.panelTitle(),
		Width:   body.Width,
		Border:  components.BorderRounded,
		Focused: true,
	}, m.view.View())
}

func (m *addonsModel) panelTitle() string {
	if m.help || len(m.addons) == 0 {
		return addonsHelpTitle
	}
	addon := m.addons[m.pager.Page]
	return addon.Name + " · " + addonsKindLabel(addon.Kind)
}

func addonsKindLabel(kind string) string {
	if kind == "metodo" {
		return "método"
	}
	return kind
}

func addonsShortRevision(revision string) string {
	if len(revision) < addonsShortSha {
		return revision
	}
	return revision[:addonsShortSha]
}

func addonsCard(t theme.Theme, a catalog.Addon, width int) string {
	lines := []string{
		addonsField(t, "Resumo", a.Summary, width),
		addonsField(t, "Fonte", a.Source, width),
		addonsField(t, "Licença", a.License, width),
		addonsField(t, "Versão conferida em "+a.CheckedAt, a.Release+" · commit "+addonsShortRevision(a.Revision), width),
		addonsCommand(t, "Comando oficial no Claude Code", a.InstallClaude, width),
		addonsCommand(t, "Comando oficial no Codex", a.InstallCodex, width),
		t.Warning.Width(width).Render(addonsDisclaimer),
		"",
		addonsList(t, "O que instala", a.Installs, width),
		"",
		addonsList(t, "Conflitos", a.Conflicts, width),
		"",
		addonsField(t, "Compatibilidade com o Core", a.CoreCompat, width),
	}
	return strings.Join(lines, "\n")
}

func addonsField(t theme.Theme, label, value string, width int) string {
	head := t.Strong.Render(label + ":")
	return t.Text.Width(width).Render(head + " " + value)
}

func addonsCommand(t theme.Theme, label, command string, width int) string {
	if strings.TrimSpace(command) == "" {
		return addonsField(t, label, "sem comando oficial publicado.", width)
	}
	return t.Strong.Render(label+":") + "\n" + t.Code.Width(width).Render(command)
}

func addonsList(t theme.Theme, label string, items []string, width int) string {
	lines := []string{t.Strong.Render(label + ":")}
	for _, item := range items {
		lines = append(lines, addonsBullet(t, item, width))
	}
	return strings.Join(lines, "\n")
}

func addonsBullet(t theme.Theme, text string, width int) string {
	bullet := "•"
	if t.Plain {
		bullet = "-"
	}
	wrapped := strings.Split(t.Text.Width(max(width-2, 1)).Render(text), "\n")
	for i, line := range wrapped {
		prefix := "  "
		if i == 0 {
			prefix = bullet + " "
		}
		wrapped[i] = prefix + line
	}
	return strings.Join(wrapped, "\n")
}

func addonsHelpText(k components.Keys) string {
	lines := []string{
		k.Left.Help().Key + "  " + k.Right.Help().Key + "  trocar de addon",
		k.Up.Help().Key + "  " + k.Down.Help().Key + "  rolar o card",
		k.Help.Help().Key + "  mostrar ou ocultar esta ajuda",
		k.Back.Help().Key + "  sair",
		k.Quit.Help().Key + "  sair",
	}
	return strings.Join(lines, "\n")
}
