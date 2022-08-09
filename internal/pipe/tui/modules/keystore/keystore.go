package keystore

import (
	ctx "context"
	"errors"
	"strings"

	"github.com/jon4hz/deadshot/internal/context"
	"github.com/jon4hz/deadshot/internal/pipe/tui/modules"
	"github.com/jon4hz/deadshot/internal/ui/bubbles/simpleview"
	"github.com/jon4hz/deadshot/internal/ui/style"
	"github.com/jon4hz/deadshot/internal/wallet"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type (
	errPasswdMismatchMsg struct{}
)

type state int

const (
	stateUnknown state = iota
	stateInput
	stateConfirm
)

type Module struct {
	ctx    ctx.Context
	cancel ctx.CancelFunc
	D      modules.Default

	state   state
	input   textinput.Model
	help    help.Model
	passwd  string
	pipeMsg string
	err     error
}

var (
	_ modules.Module          = (*Module)(nil)
	_ modules.ModulePiper     = (*Module)(nil)
	_ simpleview.SimpleViewer = (*Module)(nil)
)

func NewModule(module *modules.Default) modules.Module {
	ip := textinput.New()
	ip.Placeholder = "Super secret password"
	ip.CursorStyle = lipgloss.NewStyle().Foreground(style.GetMainColor())
	// ip.EchoMode = textinput.EchoPassword
	ip.Focus()

	return &Module{
		D: modules.Default{
			PrePipe:  module.PrePipe,
			Pipe:     module.Pipe,
			PostPipe: module.PostPipe,
		},
		cancel: func() {},
		help:   help.New(),
		input:  ip,
	}
}

func (m *Module) Cancel()        { m.cancel() }
func (m *Module) String() string { return "keystore module" }
func (m *Module) State() int     { return int(m.state) }

func (m *Module) Skip(ctx *context.Context) bool {
	return ctx.Cfg.Keystore != "file"
}

func (m *Module) Init(c *context.Context) tea.Cmd {
	m.ctx, m.cancel = ctx.WithCancel(c)
	m.D.Ctx = c
	m.state = 1
	m.err = nil
	if m.input.Reset() {
		return textinput.Blink
	}
	return modules.None
}

func (m *Module) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return tea.Quit
		}
		switch {
		case key.Matches(msg, defaultKeyMap.Enter):
			m.err = nil
			return m.confirm()
		case key.Matches(msg, defaultKeyMap.Back):
			return modules.Back
		}

	case errPasswdMismatchMsg:
		m.err = modules.Error{
			Message: "Passwords do not match",
			Help:    "Please try again",
		}

	case modules.ErrMsg:
		if errors.Is(msg, wallet.ErrWrongPassword) {
			m.err = modules.Error{
				Message: "wrong password",
				Help:    "please, try again",
			}
		} else {
			m.err = msg
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return cmd
}

func (m *Module) confirm() tea.Cmd {
	switch m.state {
	case stateInput:
		m.passwd = m.input.Value()
		if !m.D.Ctx.KeystoreExists {
			m.state = stateConfirm
			m.input.Reset()
			return nil
		}
		m.D.Ctx.Cfg.Password = m.passwd
		m.input.Reset()
		return modules.Next

	case stateConfirm:
		if m.input.Value() == m.passwd {
			m.D.Ctx.Cfg.Password = m.passwd
			m.input.Reset()
			return modules.Next
		}
		return func() tea.Msg {
			m.state = stateInput
			m.input.Reset()
			return errPasswdMismatchMsg{}
		}
	}
	return nil
}

func (m *Module) SetHeaderWidth(width int) {}
func (m *Module) Header() string           { return style.GenLogo() }

func (m *Module) SetContentSize(width, height int) {
	m.input.Width = width - lipgloss.Width(m.input.Prompt)
	m.input.PlaceholderStyle = m.input.PlaceholderStyle.MaxWidth(m.input.Width)
}

func (m *Module) Content() string {
	var s strings.Builder
	switch m.state {
	case stateInput:
		s.WriteString("Please enter your password\n\n")
	case stateConfirm:
		s.WriteString("Please confirm your password\n\n")
	}
	s.WriteString(m.input.View())
	return s.String()
}

func (m *Module) MinContentHeight() int {
	return lipgloss.Height(m.Content())
}

func (m *Module) Error() error { return m.err }

func (m *Module) SetFooterWidth(width int) { m.help.Width = width }
func (m *Module) Footer() string           { return m.help.View(defaultKeyMap) }

func (m *Module) PrePipe() []modules.Piper  { return m.D.PrePipe }
func (m *Module) Pipe() []modules.Piper     { return m.D.Pipe }
func (m *Module) PostPipe() []modules.Piper { return m.D.PostPipe }
