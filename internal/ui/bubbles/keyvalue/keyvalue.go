package keyvalue

import (
	"github.com/jon4hz/deadshot/internal/ui/style"

	"github.com/charmbracelet/lipgloss"
)

type KeyValue struct {
	key   string
	value string
}

type Model struct {
	width int
}

type Option func(*Model)

func NewKV(key string, value string) KeyValue {
	return KeyValue{key, value}
}

func New() Model {
	return Model{}
}

func (m *Model) SetWidth(width int) {
	m.width = width
}

var borderStyle = lipgloss.NewStyle().
	Border(
		lipgloss.NormalBorder(),
		false, false, false, true,
	).
	BorderForeground(style.SubtleStyle.GetForeground()).
	PaddingLeft(1)

func (m *Model) View(kvs ...KeyValue) string {
	lines := make([]string, 0, len(kvs))
	for _, kv := range kvs {
		availableWidth := m.width - borderStyle.GetBorderLeftSize()
		kv.key += ": "
		availableWidth -= lipgloss.Width(kv.key)

		v := style.SecondStyle.Width(availableWidth).Render(kv.value)
		lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Left, kv.key, v))
	}

	return borderStyle.
		Width(m.width).
		Render(lipgloss.JoinVertical(lipgloss.Top, lines...))
}
