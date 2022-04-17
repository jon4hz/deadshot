package endpoint

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Any  key.Binding
	Back key.Binding
	Quit key.Binding
}

var defaultKeyMap = keyMap{
	Any: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("any other key", "continue"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Quit, k.Back, k.Any}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		k.ShortHelp(),
	}
}
