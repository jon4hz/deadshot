package endpoint

import (
	ctx "context"
	"strings"

	"github.com/jon4hz/deadshot/internal/blockchain"
	"github.com/jon4hz/deadshot/internal/context"
	"github.com/jon4hz/deadshot/internal/pipe/tui/modules"
	"github.com/jon4hz/deadshot/internal/ui/bubbles/keyvalue"
	"github.com/jon4hz/deadshot/internal/ui/bubbles/simpleview"
	"github.com/jon4hz/deadshot/internal/ui/style"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type endpointDoneMsg struct{}

type endpointIndex int

const (
	input endpointIndex = iota
	okButton
	cancelButton
)

type state int

const (
	stateUnknown state = iota
	stateReady
	stateSubmitting
)

var (
	_ modules.Module          = (*Module)(nil)
	_ simpleview.SimpleViewer = (*Module)(nil)
)

type Module struct {
	ctx           ctx.Context
	cancel        ctx.CancelFunc
	D             modules.Default
	state         state
	err           error
	help          help.Model
	kv            keyvalue.Model
	endpointIndex endpointIndex
	newEndpoint   string
	errMsg        string
	input         textinput.Model
	spinner       spinner.Model
}

func New(module *modules.Default) *Module {
	return &Module{
		D: modules.Default{
			PrePipe:     module.PrePipe,
			Pipe:        module.Pipe,
			PostPipe:    module.PostPipe,
			ForkBackMsg: module.ForkBackMsg,
		},
		cancel:  func() {},
		help:    help.New(),
		kv:      keyvalue.New(),
		input:   textinput.New(),
		spinner: style.GetSpinnerDot(),
	}
}

func (m *Module) Cancel()        { m.cancel() }
func (m *Module) State() int     { return int(m.state) }
func (m *Module) String() string { return "settings endpoint module" }

func (m *Module) Init(c *context.Context) tea.Cmd {
	m.ctx, m.cancel = ctx.WithCancel(c)
	m.state = 1
	m.D.Ctx = c
	m.err = nil

	m.endpointIndex = 0

	m.input.Placeholder = "https://my.quicknode.com | none"
	m.input.Prompt = style.GetFocusedPrompt()
	m.input.CursorStyle = lipgloss.NewStyle().Foreground(style.GetMainColor())
	m.input.Focus()
	m.input.CharLimit = 64
	m.input.Reset()

	return textinput.Blink
}

func (m *Module) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case stateReady:
			switch {
			case key.Matches(msg, defaultKeys.Enter):
				m.state = stateSubmitting
				m.err = nil
				m.newEndpoint = strings.TrimSpace(m.input.Value())
				return m.enter()

			case key.Matches(msg, defaultKeys.Back):
				return tea.Batch(
					m.back(),
					m.spinner.Tick,
				)

			case key.Matches(msg, defaultKeys.Forward):
				m.endpointIndexForward()

			case key.Matches(msg, defaultKeys.Backward):
				m.endpointIndexBackward()

			case key.Matches(msg, defaultKeys.Left):
				if m.endpointIndex != input {
					m.endpointIndexBackward()
				}

			case key.Matches(msg, defaultKeys.Right):
				if m.endpointIndex != input {
					m.endpointIndexForward()
				}

			case key.Matches(msg, defaultKeys.Down, defaultKeys.Up):
				if m.endpointIndex == input {
					m.endpointIndexForward()
				} else {
					m.endpointIndex = input
					return m.updateFocus()
				}

			case key.Matches(msg, defaultKeys.Quit):
				return tea.Quit

			case key.Matches(msg, defaultKeys.Help):
				m.help.ShowAll = !m.help.ShowAll
			}

		case stateSubmitting:
			switch {
			case key.Matches(msg, defaultKeys.Back):
				m.state = stateReady
				return nil

			case key.Matches(msg, defaultKeys.Quit):
				return tea.Quit

			case key.Matches(msg, defaultKeys.Help):
				m.help.ShowAll = !m.help.ShowAll
			}
		}

	case modules.ErrMsg:
		m.state = stateReady
		m.err = msg
	}

	var cmd tea.Cmd
	switch m.state {
	case stateReady:
		if m.endpointIndex == input {
			m.input, cmd = m.input.Update(msg)
			return cmd
		}

	case stateSubmitting:
		m.spinner, cmd = m.spinner.Update(msg)
		return cmd
	}

	return nil
}

func (m Module) back() tea.Cmd {
	if m.D.ForkBackMsg != 0 {
		return func() tea.Msg { return m.D.ForkBackMsg }
	}
	return modules.Back
}

func (m Module) enter() tea.Cmd {
	switch m.endpointIndex {
	case input:
		fallthrough
	case okButton: // Submit the form
		m.state = stateSubmitting
		m.err = nil
		m.newEndpoint = strings.TrimSpace(m.input.Value())

		return tea.Batch(
			m.setEndpoint(),
			spinner.Tick,
		)
	case cancelButton:
		return m.back()
	}
	return nil
}

// updateFocus updates the focused states in the model based on the current
// focus endpointIndex.
func (m *Module) updateFocus() tea.Cmd {
	if m.endpointIndex == input && !m.input.Focused() {
		m.input.Prompt = style.GetFocusedPrompt()
		return m.input.Focus()
	} else if m.endpointIndex != input && m.input.Focused() {
		m.input.Blur()
		m.input.Prompt = style.GetPrompt()
	}
	return nil
}

// Move the focus endpointIndex one unit forward.
func (m *Module) endpointIndexForward() tea.Cmd {
	m.endpointIndex++
	if m.endpointIndex > cancelButton {
		m.endpointIndex = input
	}

	return m.updateFocus()
}

// Move the focus endpointIndex one unit backwards.
func (m *Module) endpointIndexBackward() tea.Cmd {
	m.endpointIndex--
	if m.endpointIndex < input {
		m.endpointIndex = cancelButton
	}

	return m.updateFocus()
}

func (m *Module) setEndpoint() tea.Cmd {
	return func() tea.Msg {
		var err error
		if strings.EqualFold(m.newEndpoint, "none") {
			err = m.D.Ctx.Network.RemoveCustomEndpoint()
			if err != nil {
				return modules.Error{
					Message: "Oh, what? There was a curious error we were not expecting",
					Help:    err.Error(),
				}
			}
			return m.forkBack()
		}
		if !blockchain.ValidateEndpointURL(m.newEndpoint, m.D.Ctx.Network.GetChainID()) {
			return modules.Error{
				Message: "Invalid Endpoint",
				Help:    "It doesn't seem like this url will work, please try again.",
			}
		}

		err = m.D.Ctx.Network.CreateCustomEndpoint(m.newEndpoint)
		if err != nil {
			return modules.Error{
				Message: "Oh, what? There was a curious error we were not expecting",
				Help:    err.Error(),
			}
		}
		err = m.D.Ctx.Config.ReloadNetworks()
		if err != nil {
			return modules.Error{
				Message: "Oh, what? There was a curious error we were not expecting",
				Help:    err.Error(),
			}
		}

		return m.forkBack()
	}
}

func (m Module) forkBack() tea.Msg {
	if m.D.ForkBackMsg != 0 {
		return m.D.ForkBackMsg
	}
	return modules.BackMsg{}
}

func (m *Module) SetHeaderWidth(width int) { m.kv.SetWidth(width) }
func (m *Module) Header() string {
	var s strings.Builder
	s.WriteString(style.GenLogo())
	s.WriteString("\n\n")
	endpoint, ok := m.D.Ctx.Network.GetCustomEndpoint()
	var endpointURL string
	if ok {
		endpointURL = endpoint.GetURL()
	} else {
		endpointURL = "(none)"
	}
	s.WriteString(m.kv.View(
		keyvalue.NewKV("Wallet", m.D.Ctx.Config.Wallet.GetWallet()),
		keyvalue.NewKV("Network", m.D.Ctx.Network.GetFullName()),
		keyvalue.NewKV("Current Endpoint", endpointURL),
	))
	return s.String()
}

func (m *Module) SetContentSize(width, height int) {
	m.input.Width = width - 1
}

func (m *Module) MinContentHeight() int {
	return 5 // TODO: don't hardcode that value
}

func (m *Module) Content() string {
	var s strings.Builder
	switch m.state {
	case stateReady:
		s.WriteString("Enter a custom endpoint url, \"none\" removes the current url \n\n")
		s.WriteString(m.input.View() + "\n\n")
		s.WriteString(style.OKButtonView(m.endpointIndex == 1, true))
		s.WriteString(" " + style.CancelButtonView(m.endpointIndex == 2, false))
	case stateSubmitting:
		s.WriteString(m.spinner.View() + "  testing endpoint...")
	}
	return s.String()
}

func (m *Module) Error() error { return m.err }

func (m *Module) SetFooterWidth(width int) { m.help.Width = width }
func (m *Module) Footer() string {
	switch m.state {
	case stateReady:
		return m.help.View(defaultKeys)
	}
	return ""
}
