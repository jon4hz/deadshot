package token

import (
	ctx "context"
	"fmt"
	"strings"

	"github.com/jon4hz/deadshot/internal/context"
	"github.com/jon4hz/deadshot/internal/database"
	"github.com/jon4hz/deadshot/internal/pipe/tui/modules"
	"github.com/jon4hz/deadshot/internal/ui/bubbles/keyvalue"
	"github.com/jon4hz/deadshot/internal/ui/bubbles/simpleview"
	"github.com/jon4hz/deadshot/internal/ui/style"
	"github.com/jon4hz/deadshot/pkg/ethutils"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const minTokenListHeight = 10

type tokenOption int

const (
	tokenOptionList tokenOption = iota
	tokenOptionCustom
)

var tokenOptions = []tokenOption{
	tokenOptionList,
	tokenOptionCustom,
}

func (t tokenOption) String(m *Module) string {
	switch t {
	case tokenOptionList:
		return "Select a token"
	case tokenOptionCustom:
		return m.tokenInput.View()
	}
	return ""
}

type tokenListItem struct {
	token *database.Token
}

func (t tokenListItem) Title() string       { return t.token.GetSymbol() }
func (t tokenListItem) Description() string { return t.token.GetContract() }
func (t tokenListItem) FilterValue() string { return t.token.GetSymbol() }

type state int

const (
	stateUnknown state = iota
	stateTokenOption
	stateTokenList
	stateFetchInfo
)

var (
	_ modules.Module          = (*Module)(nil)
	_ modules.ModulePiper     = (*Module)(nil)
	_ simpleview.SimpleViewer = (*Module)(nil)
)

type Module struct {
	ctx      ctx.Context
	cancel   ctx.CancelFunc
	D        modules.Default
	err      error
	state    state
	spinner  spinner.Model
	pipeMsg  string
	isToken0 bool

	tokenOptions      []tokenOption
	tokenOptionsIndex int
	tokenInput        textinput.Model

	tokenList list.Model

	help help.Model
	kv   keyvalue.Model
}

func NewModule(module *modules.Default, isToken0 bool) *Module {
	del := list.NewDefaultDelegate()
	del.Styles.SelectedDesc.Foreground(style.GetMainColor()).BorderForeground(style.GetSecondColor())
	del.Styles.SelectedTitle.Foreground(style.GetMainColor()).BorderForeground(style.GetSecondColor())

	return &Module{
		D: modules.Default{
			PrePipe:     module.PrePipe,
			Pipe:        module.Pipe,
			PostPipe:    module.PostPipe,
			ForkBackMsg: module.ForkBackMsg,
		},
		cancel:       func() {},
		tokenList:    list.New(nil, del, 0, 0),
		tokenInput:   textinput.New(),
		help:         help.New(),
		kv:           keyvalue.New(),
		spinner:      style.GetSpinnerPoints(),
		tokenOptions: tokenOptions,
		isToken0:     isToken0,
	}
}

func (m *Module) Cancel()        { m.cancel() }
func (m *Module) State() int     { return int(m.state) }
func (m *Module) String() string { return "token module" }

func (m *Module) Init(c *context.Context) tea.Cmd {
	m.ctx, m.cancel = ctx.WithCancel(c)
	m.state = 1
	m.D.Ctx = c
	m.err = nil
	items := make([]list.Item, len(c.Network.GetTokens()))
	for i, t := range c.Network.GetTokens() {
		items[i] = tokenListItem{t}
	}
	m.tokenList.SetItems(items)
	m.tokenList.SetShowHelp(false)
	m.tokenList.Title = "Tokens"
	m.tokenList.Styles.Title = style.GetListTitleStyle()
	m.tokenList.Styles.FilterCursor.Foreground(style.GetMainColor())
	m.tokenInput.Prompt = ""
	m.tokenInput.Placeholder = "0xb33EaAd8d922B1083446DC23f610c2567fB5180f"
	m.tokenInput.CursorStyle = lipgloss.NewStyle().Foreground(style.GetMainColor())
	m.tokenInput.CharLimit = 64

	return textinput.Blink
}

func (m *Module) Update(msg tea.Msg) tea.Cmd {
	switch m.state {
	case stateTokenOption:
		return m.updateTokenOption(msg)
	case stateTokenList:
		return m.updateTokenList(msg)
	case stateFetchInfo:
		return m.updateTokenInfo(msg)
	}
	return nil
}

func (m *Module) updateTokenOption(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, defaultKeyMap.Back):
			if m.isToken0 {
				m.D.Ctx.Token0 = nil
			} else {
				m.D.Ctx.Token1 = nil
			}
			return modules.Back

		case key.Matches(msg, defaultKeyMap.Quit):
			return tea.Quit

		case key.Matches(msg, defaultKeyMap.Help):
			m.help.ShowAll = !m.help.ShowAll
			return modules.Resize

		case key.Matches(msg, defaultKeyMap.Up):
			return m.tokenOptionBackward()

		case key.Matches(msg, defaultKeyMap.Down):
			return m.tokenOptionForward()

		case key.Matches(msg, defaultKeyMap.Enter):
			if tokenOption(m.tokenOptionsIndex) == tokenOptionCustom {
				m.state = stateFetchInfo

				token := strings.TrimSpace(m.tokenInput.Value())
				if !ethutils.IsValidAddress(token) {
					return func() tea.Msg {
						return modules.Error{
							Message: "Invalid Address",
							Help:    "Please check if the contract is valid and try again.",
						}
					}
				}
				m.D.Ctx.TokenContract = token

				return tea.Batch(
					m.spinner.Tick,
					modules.Next,
				)
			}

			m.state = stateTokenList
			m.err = nil
			return modules.Resize

		default:
			m.tokenOptionsIndex = int(tokenOptionCustom)
			cmds := make([]tea.Cmd, 0)

			if cmd := m.updateFocus(); cmd != nil {
				cmds = append(cmds, cmd)
			}

			var cmd tea.Cmd
			m.tokenInput, cmd = m.tokenInput.Update(msg)
			cmds = append(cmds, cmd)
			return tea.Batch(cmds...)
		}

	default:
		var cmd tea.Cmd
		m.tokenInput, cmd = m.tokenInput.Update(msg)
		return cmd
	}
}

func (m *Module) updateTokenList(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, defaultKeyMap.Back) && !m.tokenList.SettingFilter() && m.tokenList.FilterState() != list.FilterApplied:
			m.state = stateTokenOption
			return nil

		case key.Matches(msg, defaultKeyMap.Quit) && !m.tokenList.SettingFilter():
			return tea.Quit

		case key.Matches(msg, defaultKeyMap.Help):
			m.help.ShowAll = !m.help.ShowAll
			return modules.Resize

		case key.Matches(msg, defaultKeyMap.Enter) && !m.tokenList.SettingFilter():
			token := m.tokenList.SelectedItem().(tokenListItem)

			if m.isToken0 {
				m.D.Ctx.Token0 = token.token
			} else {
				m.D.Ctx.Token1 = token.token
			}

			m.state = stateFetchInfo
			return tea.Batch(
				m.spinner.Tick,
				modules.Next,
			)
		}
	}
	var cmd tea.Cmd
	m.tokenList, cmd = m.tokenList.Update(msg)
	return cmd
}

func (m *Module) updateTokenInfo(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, defaultKeyMap.Quit):
			return tea.Quit

		case key.Matches(msg, defaultKeyMap.Help):
			m.help.ShowAll = !m.help.ShowAll
			return modules.Resize

			// back is explicitly disabled in this state
			// currently there is no way to cancel the post pipeline
			// and this would lead to a race condition.
		}

	case modules.ErrMsg:
		m.err = msg
		m.state = stateTokenOption
		if m.tokenInput.Reset() {
			return textinput.Blink
		}

	case modules.PipeMsg:
		m.pipeMsg = string(msg)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return cmd
	}
	return nil
}

func (m *Module) tokenOptionForward() tea.Cmd {
	m.tokenOptionsIndex++
	if m.tokenOptionsIndex >= len(m.tokenOptions) {
		m.tokenOptionsIndex = 0
	}
	return m.updateFocus()
}

func (m *Module) tokenOptionBackward() tea.Cmd {
	m.tokenOptionsIndex--
	if m.tokenOptionsIndex < 0 {
		m.tokenOptionsIndex = len(m.tokenOptions) - 1
	}
	return m.updateFocus()
}

func (m *Module) updateFocus() tea.Cmd {
	if m.tokenOptionsIndex == int(tokenOptionCustom) && !m.tokenInput.Focused() {
		return m.tokenInput.Focus()
	} else if m.tokenOptionsIndex != int(tokenOptionCustom) && m.tokenInput.Focused() {
		m.tokenInput.Blur()
	}
	return nil
}

func (m *Module) SetHeaderWidth(width int) { m.kv.SetWidth(width) }
func (m *Module) Header() string {
	var s strings.Builder
	s.WriteString(style.GenLogo())
	s.WriteString("\n\n")
	kvs := []keyvalue.KeyValue{
		keyvalue.NewKV("Wallet", m.D.Ctx.Config.Wallet.GetWallet()),
		keyvalue.NewKV("Endpoint", m.D.Ctx.Endpoint.GetURL()),
	}

	if !m.isToken0 {
		kvs = append(kvs,
			keyvalue.NewKV("Tokens", m.D.Ctx.Token0.GetSymbol()),
			keyvalue.NewKV("Balance", m.D.Ctx.Token0.GetBalanceDecimal(m.D.Ctx.Token0.GetDecimals()).String()),
		)
	}

	s.WriteString(m.kv.View(kvs...))
	return s.String()
}

func (m *Module) SetContentSize(width, height int) {
	m.tokenList.SetSize(width, height)
	m.tokenInput.Width = width - lipgloss.Width(m.tokenInput.Prompt)
	m.tokenInput.PlaceholderStyle = m.tokenInput.PlaceholderStyle.MaxWidth(m.tokenInput.Width)
}

func (m *Module) MinContentHeight() int {
	switch m.state {
	case stateTokenOption:
		return 5
	case stateTokenList:
		return minTokenListHeight
	}
	return 5 // TODO: don't hardcode that value
}

func (m *Module) Content() string {
	var s strings.Builder
	switch m.state {
	case stateTokenOption:
		s.WriteString(m.tokenOptionView())
	case stateTokenList:
		s.WriteString(m.tokenList.View())
	case stateFetchInfo:
		s.WriteString(m.tokenInfoView())
	}
	return s.String()
}

func (m *Module) tokenOptionView() string {
	var s strings.Builder
	s.WriteString("Please select a token or paste a contract\n\n")
	for i := 0; i < len(tokenOptions); i++ {
		e := "  "
		if i == m.tokenOptionsIndex && i != int(tokenOptionCustom) {
			e = style.GetFocusedPrompt()
			e += style.GetFocusedText(tokenOptions[tokenOption(i)].String(m))
		} else if i == m.tokenOptionsIndex && i == int(tokenOptionCustom) {
			e = style.GetFocusedPrompt()
			e += tokenOptions[tokenOption(i)].String(m)
		} else {
			e += tokenOptions[tokenOption(i)].String(m)
		}
		if i < len(tokenOptions)-1 {
			e += "\n\n"
		}
		s.WriteString(e)
	}
	return s.String()
}

func (m *Module) tokenInfoView() string {
	return fmt.Sprintf("\n\n%s %s", m.spinner.View(), m.pipeMsg)
}

func (m *Module) Error() error { return m.err }

func (m *Module) SetFooterWidth(width int) { m.help.Width = width }
func (m *Module) Footer() string {
	switch m.state {
	case stateTokenOption:
		return m.help.View(defaultKeyMap)
	case stateTokenList:
		if m.tokenList.FilterState() == list.Filtering {
			return m.help.View(filteringKeys)
		}
		if m.tokenList.FilterState() == list.FilterApplied {
			return m.help.View(filteredKeys())
		}
		return m.help.View(listKeys())
	case stateFetchInfo:
		return m.help.View(defaultKeyMapQuit)
	}
	return ""
}

func (m *Module) PrePipe() []modules.Piper  { return m.D.PrePipe }
func (m *Module) Pipe() []modules.Piper     { return m.D.Pipe }
func (m *Module) PostPipe() []modules.Piper { return m.D.PostPipe }
