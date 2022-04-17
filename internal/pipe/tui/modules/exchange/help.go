package exchange

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Enter  key.Binding
	Back   key.Binding
	Quit   key.Binding
	Help   key.Binding
	Filter key.Binding
}

var defaultKeyMap = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "confirm"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	),
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Enter, k.Help},
		{k.Down, k.Back, k.Quit},
		{k.Filter},
	}
}

type filteringKeyMap struct {
	Apply  key.Binding
	Cancel key.Binding
	Quit   key.Binding
}

var filteringKeys = filteringKeyMap{
	Apply: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "apply filter"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q"),
		key.WithHelp("q", "quit"),
	),
}

func (k filteringKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Apply, k.Cancel, k.Quit}
}

func (k filteringKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Apply},
		{k.Cancel},
		{k.Quit},
	}
}

type filteredKeyMap struct {
	keyMap
}

var filteredKeys = func() filteredKeyMap {
	km := defaultKeyMap
	km.Back.SetHelp("esc", "clear filter")
	return filteredKeyMap{km}
}

func (k filteredKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Back, k.Quit}
}

func (k filteredKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Enter, k.Help},
		{k.Down, k.Back, k.Quit},
		{k.Filter},
	}
}
