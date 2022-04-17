package button

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	triggers  []string
	triggered bool
}

func New(triggers []string, triggered bool) Model {
	return Model{
		triggers:  triggers,
		triggered: triggered,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		for _, key := range m.triggers {
			if key == msg.String() {
				m.triggered = !m.triggered
				return m, nil
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.triggered {
		return "[x]"
	}
	return "[ ]"
}

func (m Model) Triggered() bool {
	return m.triggered
}
