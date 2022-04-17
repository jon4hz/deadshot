package keystore

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Enter key.Binding
	Back  key.Binding
}

var defaultKeyMap = keyMap{
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "confirm"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Enter, k.Back}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		k.ShortHelp(),
	}
}
