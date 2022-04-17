package panel

import (
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

// Model struct represents property of a panel.
type Model struct {
	isActive            bool
	viewport            viewport.Model
	activeBorderColor   lipgloss.TerminalColor
	inactiveBorderColor lipgloss.TerminalColor
	border              lipgloss.Border
	isPadded            bool
	maxHeight           int
	maxWidth            int
	removeOnOverflow    bool
	overflowing         bool
}

// NewModel creates a new instance of a panel.
func NewModel(isActive, isPadded bool, border lipgloss.Border, activeBorderColor, inactiveBorderColor lipgloss.TerminalColor) *Model {
	return &Model{
		isActive:            isActive,
		activeBorderColor:   activeBorderColor,
		inactiveBorderColor: inactiveBorderColor,
		border:              border,
		isPadded:            isPadded,
	}
}

// SetMaxWidth sets the maximum width of the panel.
func (m *Model) SetMaxWidth(width int) {
	m.maxWidth = width
}

// SetMaxxHeight sets the maximum height of the panel.
func (m *Model) SetMaxHeight(height int) {
	m.maxHeight = height
}

// SetSize sets the size of the panel and its viewport, useful when resizing the terminal.
func (m *Model) SetSize(width, height int) {
	m.SetWidth(width)
	m.SetHeight(height)
}

// SethHight sets the height of the panel.
func (m *Model) SetHeight(height int) {
	m.overflowing = false
	if m.maxHeight > 0 && height > m.maxHeight {
		height = m.maxHeight
		m.overflowing = true
	}
	m.viewport.Height = height - lipgloss.Width(m.border.Bottom+m.border.Top)
}

// SetWidth sets the width of the panel.
func (m *Model) SetWidth(width int) {
	m.overflowing = false
	if m.maxWidth > 0 && width > m.maxWidth {
		width = m.maxWidth
		m.overflowing = true
	}
	m.viewport.Width = width - lipgloss.Width(m.border.Right+m.border.Top)
}

// GetSize returns the size of the panel.
func (m *Model) GetSize() (int, int) {
	return m.viewport.Width, m.viewport.Height
}

// SetContent sets the content of the panel.
func (m *Model) SetContent(content string) {
	padding := 0

	// If the panel requires padding, add it.
	if m.isPadded {
		padding = 1
	}

	m.viewport.SetContent(
		lipgloss.NewStyle().
			MaxWidth(m.viewport.Width).
			Height(m.viewport.Height).
			PaddingLeft(padding).
			Render(content),
	)
}

// LineUp scrolls the panel up the specified number of lines.
func (m *Model) LineUp(lines int) {
	m.viewport.LineUp(lines)
}

// LineDown scrolls the panel down the specified number of lines.
func (m *Model) LineDown(lines int) {
	m.viewport.LineDown(lines)
}

// GotoTop goes to the top of the panel.
func (m *Model) GotoTop() {
	m.viewport.GotoTop()
}

// GotoBottom goes to the bottom of the panel.
func (m *Model) GotoBottom() {
	m.viewport.GotoBottom()
}

// SetActiveBorderColors sets the active border colors.
func (m *Model) SetActiveBorderColor(color lipgloss.TerminalColor) {
	m.activeBorderColor = color
}

// GetWidth returns the width of the panel.
func (m *Model) GetWidth() int {
	return m.viewport.Width
}

// GetHeight returns the height of the panel.
func (m *Model) GetHeight() int {
	return m.viewport.Height
}

// GetYOffset returns the y offset of the panel.
func (m *Model) GetYOffset() int {
	return m.viewport.YOffset
}

// View returns a string representation of the panel.
func (m *Model) View() string {
	if m.removeOnOverflow && m.overflowing {
		return ""
	}

	borderColor := m.inactiveBorderColor

	// If the panel is active, use the active border color.
	if m.isActive {
		borderColor = m.activeBorderColor
	}

	return lipgloss.NewStyle().
		BorderForeground(borderColor).
		Border(m.border).
		Width(m.viewport.Width).
		Height(m.viewport.Height).
		Render(m.viewport.View())
}
