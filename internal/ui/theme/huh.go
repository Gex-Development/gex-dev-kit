package theme

import (
	"charm.land/bubbles/v2/key"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
)

func (t Theme) Huh() huh.Theme {
	return huh.ThemeFunc(func(bool) *huh.Styles {
		p := t.Palette
		s := huh.ThemeBase(t.Dark)
		s.Focused.Base = s.Focused.Base.BorderForeground(p.Accent)
		s.Focused.Card = s.Focused.Base
		s.Focused.Title = s.Focused.Title.Foreground(p.Accent).Bold(true)
		s.Focused.NoteTitle = s.Focused.NoteTitle.Foreground(p.Accent).Bold(true).MarginBottom(1)
		s.Focused.Directory = s.Focused.Directory.Foreground(p.Accent)
		s.Focused.Description = s.Focused.Description.Foreground(p.TextSecondary)
		s.Focused.ErrorIndicator = s.Focused.ErrorIndicator.Foreground(p.Negative)
		s.Focused.ErrorMessage = s.Focused.ErrorMessage.Foreground(p.Negative)
		s.Focused.SelectSelector = s.Focused.SelectSelector.Foreground(p.Accent)
		s.Focused.NextIndicator = s.Focused.NextIndicator.Foreground(p.Accent)
		s.Focused.PrevIndicator = s.Focused.PrevIndicator.Foreground(p.Accent)
		s.Focused.Option = s.Focused.Option.Foreground(p.Text)
		s.Focused.MultiSelectSelector = s.Focused.MultiSelectSelector.Foreground(p.Accent)
		s.Focused.SelectedOption = s.Focused.SelectedOption.Foreground(p.Accent).Bold(true)
		s.Focused.SelectedPrefix = s.Focused.SelectedPrefix.Foreground(p.Positive)
		s.Focused.UnselectedOption = s.Focused.UnselectedOption.Foreground(p.Text)
		s.Focused.UnselectedPrefix = s.Focused.UnselectedPrefix.Foreground(p.TextSecondary)
		s.Focused.FocusedButton = s.Focused.FocusedButton.Foreground(p.Navy).Background(p.Accent)
		s.Focused.BlurredButton = s.Focused.BlurredButton.Foreground(p.Text).Background(p.SurfaceActive)
		s.Focused.Next = s.Focused.FocusedButton
		s.Focused.TextInput.Cursor = s.Focused.TextInput.Cursor.Foreground(p.Accent)
		s.Focused.TextInput.Placeholder = s.Focused.TextInput.Placeholder.Foreground(p.TextSecondary)
		s.Focused.TextInput.Prompt = s.Focused.TextInput.Prompt.Foreground(p.Accent)
		s.Blurred = s.Focused
		s.Blurred.Base = s.Focused.Base.BorderStyle(lipgloss.HiddenBorder())
		s.Blurred.Card = s.Blurred.Base
		s.Blurred.NextIndicator = lipgloss.NewStyle()
		s.Blurred.PrevIndicator = lipgloss.NewStyle()
		s.Group.Title = s.Focused.Title
		s.Group.Description = s.Focused.Description
		return s
	})
}

func (t Theme) HuhKeyMap() *huh.KeyMap {
	km := huh.NewDefaultKeyMap()
	huhDesc(&km.Input.AcceptSuggestion, "completar")
	huhDesc(&km.Input.Prev, "voltar")
	huhDesc(&km.Input.Next, "próximo")
	huhDesc(&km.Input.Submit, "enviar")

	huhDesc(&km.Text.Prev, "voltar")
	huhDesc(&km.Text.Next, "próximo")
	huhDesc(&km.Text.Submit, "enviar")
	huhDesc(&km.Text.NewLine, "nova linha")
	huhDesc(&km.Text.Editor, "abrir editor")

	huhDesc(&km.Select.Prev, "voltar")
	huhDesc(&km.Select.Next, "selecionar")
	huhDesc(&km.Select.Submit, "enviar")
	huhDesc(&km.Select.Up, "acima")
	huhDesc(&km.Select.Down, "abaixo")
	huhDesc(&km.Select.Left, "esquerda")
	huhDesc(&km.Select.Right, "direita")
	huhDesc(&km.Select.Filter, "filtrar")
	huhDesc(&km.Select.SetFilter, "definir filtro")
	huhDesc(&km.Select.ClearFilter, "limpar filtro")
	huhDesc(&km.Select.HalfPageUp, "½ página acima")
	huhDesc(&km.Select.HalfPageDown, "½ página abaixo")
	huhDesc(&km.Select.GotoTop, "início")
	huhDesc(&km.Select.GotoBottom, "fim")

	huhDesc(&km.MultiSelect.Prev, "voltar")
	huhDesc(&km.MultiSelect.Next, "confirmar")
	huhDesc(&km.MultiSelect.Submit, "enviar")
	huhDesc(&km.MultiSelect.Toggle, "alternar")
	huhDesc(&km.MultiSelect.Up, "acima")
	huhDesc(&km.MultiSelect.Down, "abaixo")
	huhDesc(&km.MultiSelect.Filter, "filtrar")
	huhDesc(&km.MultiSelect.SetFilter, "definir filtro")
	huhDesc(&km.MultiSelect.ClearFilter, "limpar filtro")
	huhDesc(&km.MultiSelect.HalfPageUp, "½ página acima")
	huhDesc(&km.MultiSelect.HalfPageDown, "½ página abaixo")
	huhDesc(&km.MultiSelect.GotoTop, "início")
	huhDesc(&km.MultiSelect.GotoBottom, "fim")
	huhDesc(&km.MultiSelect.SelectAll, "selecionar tudo")
	huhDesc(&km.MultiSelect.SelectNone, "desmarcar tudo")

	huhDesc(&km.FilePicker.Open, "abrir")
	huhDesc(&km.FilePicker.Close, "fechar")
	huhDesc(&km.FilePicker.GotoTop, "primeiro")
	huhDesc(&km.FilePicker.GotoBottom, "último")
	huhDesc(&km.FilePicker.PageUp, "página acima")
	huhDesc(&km.FilePicker.PageDown, "página abaixo")
	huhDesc(&km.FilePicker.Back, "voltar")
	huhDesc(&km.FilePicker.Select, "selecionar")
	huhDesc(&km.FilePicker.Up, "acima")
	huhDesc(&km.FilePicker.Down, "abaixo")
	huhDesc(&km.FilePicker.Prev, "voltar")
	huhDesc(&km.FilePicker.Next, "próximo")
	huhDesc(&km.FilePicker.Submit, "enviar")

	huhDesc(&km.Note.Prev, "voltar")
	huhDesc(&km.Note.Next, "próximo")
	huhDesc(&km.Note.Submit, "enviar")

	huhDesc(&km.Confirm.Prev, "voltar")
	huhDesc(&km.Confirm.Next, "próximo")
	huhDesc(&km.Confirm.Submit, "enviar")
	huhDesc(&km.Confirm.Toggle, "alternar")
	huhDesc(&km.Confirm.Accept, "sim")
	huhDesc(&km.Confirm.Reject, "não")
	return km
}

func huhDesc(b *key.Binding, desc string) {
	b.SetHelp(b.Help().Key, desc)
}
