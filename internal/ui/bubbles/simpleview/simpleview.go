package simpleview

import (
	"github.com/jon4hz/deadshot/internal/pipe/tui/modules"
	"github.com/jon4hz/deadshot/internal/ui/style"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

type SimpleViewer interface {
	State() int
	Header() string
	Content() string
	Error() error
	Footer() string
	MinContentHeight() int
	SetHeaderWidth(width int)
	SetContentSize(width, height int)
	SetFooterWidth(width int)
}

type ContentViewer interface {
	Content() string
	Wrap() bool
	SetContentSize(width, height int)
}

var (
	HeaderStyle  = lipgloss.NewStyle().MarginBottom(1)
	ContentStyle = HeaderStyle.Copy().Margin(0, 1, 1, 1)
	ErrorStyle   = ContentStyle.Copy()
	FooterStyle  = lipgloss.NewStyle()
)

type Model struct {
	hideHeader   bool
	contentPanel viewport.Model

	header, content, err, footer string
	width, height                int
	lastContentHeight            int
}

func NewModel() *Model {
	vp := viewport.Model{}
	return &Model{
		contentPanel: vp,
	}
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m Model) GetHeaderWidth() int {
	return m.width - HeaderStyle.GetMarginLeft() - HeaderStyle.GetMarginRight()
}

func (m Model) GetContentWidth() int {
	return m.width - ContentStyle.GetMarginLeft() - ContentStyle.GetMarginRight()
}

func (m Model) GetFullContentHeight() int {
	return m.height - ContentStyle.GetMarginBottom() - ContentStyle.GetMarginTop()
}

func (m Model) GetContentHeight(module interface{}) int {
	v, ok := module.(SimpleViewer)
	if !ok {
		return 0
	}
	if v.State() == 0 {
		return 0
	}

	var (
		availHeight = m.height
		h, e, f     = v.Header(), v.Error(), v.Footer()
	)

	var header string
	if h != "" {
		header = HeaderStyle.Render(h)
		availHeight -= lipgloss.Height(header)
	}

	if e != nil {
		err := ErrorStyle.Render(m.ErrorView(e))
		availHeight -= lipgloss.Height(err)
	}

	if f != "" {
		footer := FooterStyle.Render(f)
		availHeight -= lipgloss.Height(footer)
	}

	if availHeight < v.MinContentHeight() {
		availHeight += lipgloss.Height(header)
	}

	return availHeight - ContentStyle.GetMarginBottom() - ContentStyle.GetMarginTop()
}

func (m Model) GetFooterWidth() int {
	return m.width - FooterStyle.GetMarginLeft() - FooterStyle.GetMarginRight()
}

func (m *Model) SetHeight(height int) { m.SetSize(m.width, height) }

func (m *Model) SetWidth(width int) { m.SetSize(width, m.height) }

func (m *Model) View(module interface{}) string {
	v, ok := module.(SimpleViewer)
	if !ok {
		return m.ContentView(module)
	}
	if v.State() == 0 {
		return ""
	}
	var (
		sections    []string
		availHeight = m.height
		width       = m.width

		h, c, e, f = v.Header(), v.Content(), v.Error(), v.Footer()
	)

	var header string
	if h != "" {
		header = HeaderStyle.Render(h)
		availHeight -= lipgloss.Height(header)
	}

	var err string
	if e != nil {
		err = ErrorStyle.Render(m.ErrorView(e))
		availHeight -= lipgloss.Height(err)
	}

	var footer string
	if f != "" {
		footer = FooterStyle.Render(f)
		availHeight -= lipgloss.Height(footer)
	}

	content := ContentStyle.Copy().
		Width(width).
		Render(c)

	var hideHeader bool
	if availHeight < (v.MinContentHeight() + ContentStyle.GetMarginBottom() + ContentStyle.GetMarginTop()) {
		availHeight += lipgloss.Height(header)
		hideHeader = true
	}

	m.contentPanel.Width = width
	m.contentPanel.Height = availHeight
	m.contentPanel.SetContent(content)

	if !hideHeader && header != "" {
		sections = append(sections, header)
	}
	sections = append(sections, m.contentPanel.View())
	if err != "" {
		sections = append(sections, err)
	}
	if footer != "" {
		sections = append(sections, footer)
	}
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m Model) ContentView(module interface{}) string {
	c, ok := module.(ContentViewer)
	if !ok {
		return ""
	}
	s := lipgloss.NewStyle()
	if c.Wrap() {
		s = s.Width(m.width).Height(m.height)
	}
	return s.Render(c.Content())
}

func (m Model) ErrorView(err error) string {
	e, ok := err.(modules.Error)
	if !ok {
		return style.ErrStyle.Render(err.Error())
	}
	return e.Render(m.width - ErrorStyle.GetMarginLeft() - ErrorStyle.GetMarginRight())
}
