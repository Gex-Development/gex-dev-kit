package components

import (
	"runtime"

	"charm.land/bubbles/v2/key"
)

type Keys struct {
	Up, Down, Left, Right, Enter, Back, Quit, Help, Filter, Tab, Yes, No, Suspend key.Binding
}

func DefaultKeys() Keys {
	return keysBuild(keysSuspendSupported)
}

func keysSuspendSupported() bool {
	return runtime.GOOS != "windows"
}

func keysBuild(suspendSupported func() bool) Keys {
	k := Keys{
		Up:      key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "subir")),
		Down:    key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "descer")),
		Left:    key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "esquerda")),
		Right:   key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "direita")),
		Enter:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "escolher")),
		Back:    key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "voltar")),
		Quit:    key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "sair")),
		Help:    key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "ajuda")),
		Filter:  key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filtrar")),
		Tab:     key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "próximo")),
		Yes:     key.NewBinding(key.WithKeys("s", "y"), key.WithHelp("s/y", "sim")),
		No:      key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "não")),
		Suspend: key.NewBinding(key.WithKeys("ctrl+z"), key.WithHelp("ctrl+z", "suspender")),
	}
	k.Suspend.SetEnabled(suspendSupported())
	return k
}

func (k Keys) TextInput() Keys {
	k.Up.SetEnabled(false)
	k.Down.SetEnabled(false)
	k.Left.SetEnabled(false)
	k.Right.SetEnabled(false)
	k.Quit.SetKeys("ctrl+c")
	k.Help.SetEnabled(false)
	k.Filter.SetEnabled(false)
	k.Tab.SetEnabled(false)
	k.Yes.SetEnabled(false)
	k.No.SetEnabled(false)
	k.Suspend.SetEnabled(false)
	return k
}

func (k Keys) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Back, k.Quit, k.Help, k.Suspend}
}

func (k Keys) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Enter, k.Back, k.Tab},
		{k.Yes, k.No, k.Filter},
		{k.Help, k.Suspend, k.Quit},
	}
}
