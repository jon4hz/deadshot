package secret

import (
	"fmt"
	"strings"

	"github.com/jon4hz/deadshot/internal/context"
	"github.com/jon4hz/deadshot/internal/pipe/tui/modules"
	"github.com/jon4hz/deadshot/internal/ui/bubbles/simpleview"
	"github.com/jon4hz/deadshot/internal/ui/common"
	"github.com/jon4hz/deadshot/internal/ui/style"
	"github.com/jon4hz/deadshot/internal/wallet"

	ctx "context"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tyler-smith/go-bip39"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type (
	generateMnemonicMsg    struct{}
	validMnemonicMsg       struct{}
	validPrivateKeyMsg     struct{}
	errInvalidInput        struct{}
	errGenerateMnemonicMsg struct{}
	showMnemonicMsg        struct{}
)

type state int

const (
	stateUnknown state = iota
	stateInput
	stateShowNewMnemonic
	stateConfirmNewMnemonic
)

var menuChoices = []string{
	"Generate a new wallet",
	"custom",
}

var (
	_ modules.Module          = (*Module)(nil)
	_ modules.ModulePiper     = (*Module)(nil)
	_ simpleview.SimpleViewer = (*Module)(nil)
)

type Module struct {
	ctx                  ctx.Context
	cancel               ctx.CancelFunc
	D                    modules.Default
	state                state
	input                textinput.Model
	help                 help.Model
	err                  error
	menuIndex            int
	secret               string
	mnemonicWords        []string
	mnemonicInput        string
	mnemonicConfirmCount int
	contentWidth         int
}

const inputPlaceholder = "tag volcano eight thank tide danger coast health above..."

func NewModule(module *modules.Default) modules.Module {
	ti := textinput.NewModel()
	ti.Placeholder = inputPlaceholder
	ti.Prompt = style.GetFocusedPrompt()
	ti.CursorStyle = style.GetActiveCursor()
	return &Module{
		D: modules.Default{
			PrePipe:     module.PrePipe,
			Pipe:        module.Pipe,
			PostPipe:    module.PostPipe,
			ForkBackMsg: module.ForkBackMsg,
		},
		cancel: func() {},
		input:  ti,
		help:   help.NewModel(),
	}
}

func (m *Module) Cancel()        { m.cancel() }
func (m *Module) State() int     { return int(m.state) }
func (m *Module) String() string { return "secret module" }

func (m *Module) Skip(ctx *context.Context) bool {
	return ctx.KeystoreExists && !ctx.SetNewSecret && !ctx.EditSecret
}

func (m *Module) Init(c *context.Context) tea.Cmd {
	m.ctx, m.cancel = ctx.WithCancel(c)
	m.D.Ctx = c
	m.err = nil
	m.state = 1
	m.menuIndex = 0
	m.input.Placeholder = inputPlaceholder
	m.input.Reset()
	return textinput.Blink
}

func (m *Module) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return tea.Quit
		}
		switch m.state {
		case stateInput:
			switch {
			case key.Matches(msg, defaultKeyMap.Enter):
				return m.handleChoice()
			case key.Matches(msg, defaultKeyMap.Up):
				m.menuBackward()
			case key.Matches(msg, defaultKeyMap.Down):
				m.menuForward()
			case key.Matches(msg, defaultKeyMap.Back):
				if m.D.ForkBackMsg != 0 {
					return func() tea.Msg { return m.D.ForkBackMsg }
				}
				return modules.Back
			default:
				m.menuIndex = 1 // set menu to custom automatically
				var cmd tea.Cmd
				m.input.Focus()
				m.input, cmd = m.input.Update(msg)
				return cmd
			}
		case stateShowNewMnemonic:
			switch {
			case key.Matches(msg, defaultKeyMap.Back):
				m.state = stateInput
				m.err = nil
				return nil
			default:
				m.state = stateConfirmNewMnemonic
				m.mnemonicWords = strings.Split(m.secret, " ")
				m.input.Reset()
				m.input.Focus()
			}

		case stateConfirmNewMnemonic:
			switch {
			case key.Matches(msg, defaultKeyMap.Back):
				m.state = stateInput
				m.input.Reset()
				m.input.Placeholder = inputPlaceholder
				m.mnemonicConfirmCount = 0
				m.mnemonicInput = ""
				m.err = nil
				return nil
			case key.Matches(msg, defaultKeyMap.Enter):
				m.mnemonicInput = strings.TrimSpace(m.input.Value())
				return m.confirmMnemonicWord()
			}
		}

	case validMnemonicMsg:
		m.D.Ctx.NewSecret = &context.Secret{
			Secret:     m.secret,
			Index:      -1,
			IsMnmeonic: true,
		}
		return modules.Next

	case validPrivateKeyMsg:
		m.D.Ctx.NewSecret = &context.Secret{
			Secret:     m.secret,
			Index:      0,
			IsMnmeonic: false,
		}
		return modules.Next

	case generateMnemonicMsg:
		m.state = stateShowNewMnemonic
		return m.generateMnemonic()

	case showMnemonicMsg:
		m.state = stateShowNewMnemonic
		return nil

	case errInvalidInput:
		m.state = stateInput
		m.err = modules.Error{
			Message: "Invalid input",
			Help:    "Please enter a valid mnemonic or private key",
		}
		m.input.Reset()
		return nil

	case errGenerateMnemonicMsg:
		m.state = stateInput
		m.err = modules.Error{
			Message: "Error generating mnemonic",
			Help:    "Please try again",
		}
		return nil

	case modules.ErrMsg:
		m.err = msg
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return cmd
}

func (m *Module) menuForward() {
	m.menuIndex++
	if m.menuIndex > len(menuChoices)-1 {
		m.menuIndex = 0
	}
}

func (m *Module) menuBackward() {
	m.menuIndex--
	if m.menuIndex < 0 {
		m.menuIndex = len(menuChoices) - 1
	}
}

func (m *Module) handleChoice() tea.Cmd {
	if menuChoices[m.menuIndex] == "custom" {
		m.err = nil
		m.secret = strings.TrimSpace(m.input.Value())
		return m.setSecret()
	}
	return func() tea.Msg { return generateMnemonicMsg{} }
}

func (m *Module) setSecret() tea.Cmd {
	return func() tea.Msg {
		if bip39.IsMnemonicValid(m.secret) {
			return validMnemonicMsg{}
		}
		if _, err := crypto.HexToECDSA(m.secret); err == nil {
			return validPrivateKeyMsg{}
		}
		return errInvalidInput{}
	}
}

func (m *Module) generateMnemonic() tea.Cmd {
	return func() tea.Msg {
		var err error
		m.secret, err = wallet.GenerateNewMnemonic()
		if err != nil {
			return errGenerateMnemonicMsg{}
		}
		return showMnemonicMsg{}
	}
}

func (m *Module) confirmMnemonicWord() tea.Cmd {
	return func() tea.Msg {
		if m.mnemonicInput != m.mnemonicWords[m.mnemonicConfirmCount] {
			return modules.Error{
				Message: "Wrong Word",
				Help:    "Please try again",
			}
		}
		if m.mnemonicConfirmCount == len(m.mnemonicWords)-1 {
			return validMnemonicMsg{}
		}
		m.mnemonicConfirmCount++
		m.input.Reset()
		m.err = nil
		return nil
	}
}

func (m *Module) SetHeaderWidth(width int) {}
func (m *Module) Header() string           { return style.GenLogo() }

func (m *Module) SetContentSize(width, height int) {
	w := width - lipgloss.Width(m.input.Prompt)
	m.input.Width = w
	m.input.PlaceholderStyle = m.input.PlaceholderStyle.MaxWidth(w)
	m.contentWidth = width
}

func (m *Module) Content() string {
	switch m.state {
	case stateInput:
		return m.inputSecretView()
	case stateShowNewMnemonic:
		return m.showNewMnemonicView()
	case stateConfirmNewMnemonic:
		return m.confirmNewMnemonicView()
	}
	return ""
}

func (m *Module) inputSecretView() string {
	var s strings.Builder
	if !m.D.Ctx.KeystoreExists {
		s.WriteString("It seems that you use the bot for the first time. Let's configure your wallet.\n")
	}
	s.WriteString("Please enter your seed phrase, private key or generate a new wallet\n\n")

	// predefined choices
	for i := 0; i < len(menuChoices)-1; i++ {
		e := "  "
		if i == m.menuIndex {
			e = lipgloss.NewStyle().Foreground(style.GetMainColor()).Render(style.GetPrompt())
			e += lipgloss.NewStyle().Foreground(style.GetMainColor()).Render(menuChoices[i])
		} else {
			e += menuChoices[i]
		}
		if i < len(menuChoices)-1 {
			e += "\n\n"
		}
		s.WriteString(e)
	}
	if menuChoices[m.menuIndex] == "custom" {
		m.input.Prompt = style.GetFocusedPrompt()
		if !m.input.Focused() {
			m.input.Focus()
		}
	} else {
		m.input.Prompt = "  "
		m.input.Blur()
	}
	s.WriteString(m.input.View())

	return s.String()
}

func (m *Module) showNewMnemonicView() string {
	var s strings.Builder
	s.WriteString(lipgloss.NewStyle().Underline(true).Width(m.contentWidth).Render("Your new mnemonic is:\n"))
	s.WriteString(style.SecondStyle.Width(m.contentWidth).Render(m.secret))
	s.WriteString("\n\nPlease write down your mnemonic phrase. You wont be able to show it again. Without it you risk to loose all your funds on the wallet!")
	return s.String()
}

func (m *Module) confirmNewMnemonicView() string {
	var s strings.Builder
	m.input.Prompt = style.GetFocusedPrompt()
	m.input.Placeholder = ""
	s.WriteString(fmt.Sprintf("Please write down the %d. word of your mnemonic phrase\n\n", m.mnemonicConfirmCount+1))
	s.WriteString(m.input.View())
	return s.String()
}

func (m *Module) MinContentHeight() int {
	return lipgloss.Height(m.Content())
}

func (m *Module) Error() error { return m.err }

func (m *Module) SetFooterWidth(width int) { m.help.Width = width }
func (m *Module) Footer() string {
	switch m.state {
	case stateInput:
		return m.help.View(defaultKeyMap)
	case stateShowNewMnemonic:
		return common.HelpView("Press any key to continue")
	case stateConfirmNewMnemonic:
		return common.HelpView("Press enter to confirm")
	}
	return ""
}

func (m *Module) PrePipe() []modules.Piper   { return m.D.PrePipe }
func (m *Module) Pipe() []modules.Piper      { return m.D.Pipe }
func (m *Module) PostPipe() []modules.Piper  { return m.D.PostPipe }
func (m *Module) HideHeaderOnOverflow() bool { return true }
