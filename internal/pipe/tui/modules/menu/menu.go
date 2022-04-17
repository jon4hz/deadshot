package menu

import (
	ctx "context"
	"strings"

	"github.com/jon4hz/deadshot/internal/context"
	"github.com/jon4hz/deadshot/internal/pipe/tui/modules"
	"github.com/jon4hz/deadshot/internal/ui/bubbles/keyvalue"
	"github.com/jon4hz/deadshot/internal/ui/bubbles/simpleview"
	"github.com/jon4hz/deadshot/internal/ui/style"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type state int

const (
	stateUnknown state = iota
	stateReady
)

type menuChoice int

const (
	tradeChoice menuChoice = iota
	settingsChoice
	quitChoice
	unsetChoice
)

var menuChoices = map[menuChoice]string{
	tradeChoice:    "Trade",
	settingsChoice: "Settings",
	quitChoice:     "Quit",
}

var (
	_ modules.Forker          = (*Module)(nil)
	_ modules.Module          = (*Module)(nil)
	_ modules.ModulePiper     = (*Module)(nil)
	_ simpleview.SimpleViewer = (*Module)(nil)
)

type Module struct {
	ctx    ctx.Context
	cancel ctx.CancelFunc
	D      modules.Default
	state  state
	err    error
	help   help.Model
	kv     keyvalue.Model

	menuChoice menuChoice
	menuIndex  int
}

func NewModule(module *modules.Default) *Module {
	return &Module{
		D: modules.Default{
			PrePipe:  module.PrePipe,
			Pipe:     module.Pipe,
			PostPipe: module.PostPipe,
		},
		cancel: func() {},
		help:   help.New(),
		kv:     keyvalue.New(),
	}
}

func (m *Module) Cancel()        { m.cancel() }
func (m *Module) State() int     { return int(m.state) }
func (m *Module) String() string { return "menu module" }

func (m *Module) Skip() bool {
	return false
}

func (m *Module) Init(c *context.Context) tea.Cmd {
	m.ctx, m.cancel = ctx.WithCancel(c)
	m.state = 1
	m.D.Ctx = c
	m.err = nil
	return modules.None
}

func (m *Module) Fork() modules.ForkMsg {
	switch menuChoice(m.menuIndex) {
	case tradeChoice:
		return modules.ForkMsgTrade
	case settingsChoice:
		return modules.ForkMsgSettings
	case quitChoice:
		return modules.ForkMsgQuit
	}
	return modules.ForkMsgNone
}

func (m *Module) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case stateReady:
			switch {
			case key.Matches(msg, defaultKeyMap.Enter):
				return func() tea.Msg { return m.Fork() }
			case key.Matches(msg, defaultKeyMap.Down):
				m.menuForward()
			case key.Matches(msg, defaultKeyMap.Up):
				m.menuBackward()
			case key.Matches(msg, defaultKeyMap.Back):
				return modules.Back
			case key.Matches(msg, defaultKeyMap.Quit):
				return tea.Quit
			case key.Matches(msg, defaultKeyMap.Help):
				m.help.ShowAll = !m.help.ShowAll
			}
		}

	case modules.ErrMsg:
		m.err = msg
	}
	return nil
}

func (m *Module) menuForward() {
	m.menuIndex++
	if m.menuIndex >= len(menuChoices) {
		m.menuIndex = 0
	}
}

func (m *Module) menuBackward() {
	m.menuIndex--
	if m.menuIndex < 0 {
		m.menuIndex = len(menuChoices) - 1
	}
}

func (m *Module) SetHeaderWidth(width int) { m.kv.SetWidth(width) }
func (m *Module) Header() string {
	var s strings.Builder
	s.WriteString(style.GenLogo())
	s.WriteString("\n\n")
	s.WriteString(m.kv.View(
		keyvalue.NewKV("Wallet", m.D.Ctx.Config.Wallet.GetWallet()),
	))
	return s.String()
}

func (m *Module) SetContentSize(width, height int) {}

func (m *Module) Content() string {
	var s strings.Builder
	switch m.state {
	case stateReady:
		s.WriteString(m.readyView())
	}
	return s.String()
}

func (m *Module) readyView() string {
	var s string
	for i := 0; i < len(menuChoices); i++ {
		e := "  "
		if i == m.menuIndex {
			e = style.GetFocusedPrompt()
			e += style.GetFocusedText(menuChoices[menuChoice(i)])
		} else {
			e += menuChoices[menuChoice(i)]
		}
		if i < len(menuChoices)-1 {
			e += "\n"
		}
		s += e
	}
	return s
}

func (m *Module) MinContentHeight() int {
	return lipgloss.Height(m.Content())
}

func (m *Module) Error() error { return m.err }

func (m *Module) SetFooterWidth(width int) { m.help.Width = width }
func (m *Module) Footer() string {
	switch m.state {
	case stateReady:
		return m.help.View(defaultKeyMap)
	}
	return ""
}

func (m *Module) PrePipe() []modules.Piper  { return m.D.PrePipe }
func (m *Module) Pipe() []modules.Piper     { return m.D.Pipe }
func (m *Module) PostPipe() []modules.Piper { return m.D.PostPipe }
