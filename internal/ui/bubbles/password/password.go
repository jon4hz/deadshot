package password

import (
	"strings"

	"github.com/jon4hz/deadshot/internal/ui/bubbles/panel"
	"github.com/jon4hz/deadshot/internal/ui/style"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type (
	PasswordMsg          string
	errPasswdMismatchMsg struct{}
)

type state int

const (
	stateInput state = iota
	stateConfirm
)

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

type Model struct {
	Done            bool
	state           state
	inputPasswd     textinput.Model
	help            help.Model
	confirmRequired bool
	mainPanel       *panel.Model
	passwd          string
	err             string
}

func NewModel(confirmRequired bool) *Model {
	ip := textinput.NewModel()
	ip.Placeholder = "Super secret password"
	ip.CursorStyle = lipgloss.NewStyle().Foreground(style.GetMainColor())
	ip.EchoMode = textinput.EchoPassword
	ip.Focus()

	mainPanel := panel.NewModel(
		false, false,
		lipgloss.Border{}, lipgloss.NoColor{}, lipgloss.NoColor{},
	)

	return &Model{
		state:           stateInput,
		inputPasswd:     ip,
		confirmRequired: confirmRequired,
		mainPanel:       mainPanel,
		help:            help.NewModel(),
	}
}

func (m *Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.help.Width = msg.Width
		m.mainPanel.SetHeight(msg.Height)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, defaultKeyMap.Enter):
			m.err = ""
			return m.confirm()
		case key.Matches(msg, defaultKeyMap.Back):
			m.Done = true
			return nil
		}
	case errPasswdMismatchMsg:
		m.err = lipgloss.NewStyle().Foreground(style.GetErrColor()).Render("Passwords do not match")
	}
	var cmd tea.Cmd
	m.inputPasswd, cmd = m.inputPasswd.Update(msg)
	return cmd
}

func (m *Model) confirm() tea.Cmd {
	switch m.state {
	case stateInput:
		m.passwd = m.inputPasswd.Value()
		if m.confirmRequired {
			m.state = stateConfirm
			m.inputPasswd.Reset()
			return nil
		}
		return func() tea.Msg {
			m.inputPasswd.Reset()
			return PasswordMsg(m.passwd)
		}
	case stateConfirm:
		if m.inputPasswd.Value() == m.passwd {
			return func() tea.Msg {
				m.inputPasswd.Reset()
				return PasswordMsg(m.passwd)
			}
		}
		return func() tea.Msg {
			m.state = stateInput
			m.inputPasswd.Reset()
			return errPasswdMismatchMsg{}
		}
	}
	return nil
}

func (m *Model) View() string {
	m.mainPanel.SetContent(m.mainView())
	return m.mainPanel.View()
}

func (m *Model) mainView() string {
	var s strings.Builder
	switch m.state {
	case stateInput:
		s.WriteString("Please enter your password\n\n")
	case stateConfirm:
		s.WriteString("Please confirm your password\n\n")
	}
	s.WriteString(m.inputPasswd.View())
	s.WriteString("\n\n")
	if m.err != "" {
		s.WriteString(m.err)
		s.WriteString("\n\n")
	}
	s.WriteString(m.help.View(defaultKeyMap))
	m.mainPanel.SetHeight(strings.Count(s.String(), "\n") + 3)
	return s.String()
}
