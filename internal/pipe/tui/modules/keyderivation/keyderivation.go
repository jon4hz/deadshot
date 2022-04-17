package keyderivation

import (
	ctx "context"
	"fmt"
	"strconv"
	"strings"

	"github.com/jon4hz/deadshot/internal/context"
	"github.com/jon4hz/deadshot/internal/logging"
	"github.com/jon4hz/deadshot/internal/pipe/tui/modules"
	"github.com/jon4hz/deadshot/internal/ui/bubbles/simpleview"
	"github.com/jon4hz/deadshot/internal/ui/style"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	"github.com/sirupsen/logrus"
)

type state int

const (
	stateUnknown state = iota
	stateSelect
)

var (
	_ modules.Module          = (*Module)(nil)
	_ modules.ModulePiper     = (*Module)(nil)
	_ simpleview.SimpleViewer = (*Module)(nil)
)

type Module struct {
	ctx    ctx.Context
	cancel ctx.CancelFunc
	D      modules.Default

	state         state
	help          help.Model
	passwd        string
	pipeMsg       string
	addresses     []string
	addressOffset uint
	cursor        int
	err           error
}

const addressBatchSize = 10

func NewModule(module *modules.Default) modules.Module {
	return &Module{
		D: modules.Default{
			PrePipe:     module.PrePipe,
			Pipe:        module.Pipe,
			PostPipe:    module.PostPipe,
			ForkBackMsg: module.ForkBackMsg,
		},
		cancel: func() {},
		help:   help.NewModel(),
	}
}
func (m *Module) Cancel()    { m.cancel() }
func (m *Module) State() int { return int(m.state) }

func (m *Module) String() string { return "keyderivation module" }

func (m *Module) Skip(ctx *context.Context) bool {
	return ctx.NewSecret == nil || !ctx.NewSecret.IsMnmeonic
}

func (m *Module) Init(c *context.Context) tea.Cmd {
	m.ctx, m.cancel = ctx.WithCancel(c)
	m.D.Ctx = c
	m.state = 1
	m.err = nil
	m.cursor = 0
	m.addressOffset = 0
	m.addresses = nil
	var cmds []tea.Cmd
	cmds = append(cmds, func() tea.Msg {
		if err := m.loadWalletAddresses(0); err != nil {
			return modules.ErrMsg(err)
		}
		return nil
	})
	cmds = append(cmds, textinput.Blink)
	return tea.Batch(cmds...)
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

		case key.Matches(msg, defaultKeyMap.Down):
			return m.nextAddress()

		case key.Matches(msg, defaultKeyMap.Up):
			return m.previousAddress()

		case key.Matches(msg, defaultKeyMap.Back):
			if m.D.ForkBackMsg != 0 {
				return func() tea.Msg { return m.D.ForkBackMsg }
			}
			return modules.Back

		case key.Matches(msg, defaultKeyMap.Help):
			m.help.ShowAll = !m.help.ShowAll

		case key.Matches(msg, defaultKeyMap.Quit):
			return tea.Quit
		}

	case modules.ErrMsg:
		m.err = msg
	}

	return nil
}

func (m *Module) nextAddress() tea.Cmd {
	return func() tea.Msg {
		m.cursor++
		if m.cursor >= len(m.addresses) {
			m.cursor--
			m.addressOffset++
			if err := m.loadWalletAddresses(m.addressOffset); err != nil {
				return modules.ErrMsg(err)
			}
		}
		return nil
	}
}

func (m *Module) previousAddress() tea.Cmd {
	return func() tea.Msg {
		m.cursor--
		if m.cursor < 0 {
			m.cursor++
			if m.addressOffset > 0 {
				m.addressOffset--
				if err := m.loadWalletAddresses(m.addressOffset); err != nil {
					return modules.ErrMsg(err)
				}
			}
		}
		return nil
	}
}

func (m *Module) confirm() tea.Cmd {
	index := m.cursor + int(m.addressOffset)
	m.D.Ctx.NewSecret.Index = index
	return modules.Next
}

func (m *Module) loadWalletAddresses(offset uint) error {
	wallet, err := hdwallet.NewFromMnemonic(m.D.Ctx.NewSecret.Secret)
	if err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error": err,
		}).Error("error loading wallet")
		m.err = err
		return err
	}

	m.addresses = make([]string, addressBatchSize)

	for i := uint(0) + offset; i < addressBatchSize+offset; i++ {
		p := fmt.Sprintf("m/44'/60'/0'/0/%d", i)
		path, err := hdwallet.ParseDerivationPath(p)
		if err != nil {
			logging.Log.WithFields(logrus.Fields{
				"error": err,
				"path":  p,
			}).Error("error parsing derivation path")
			m.err = err
			return err
		}
		account, err := wallet.Derive(path, false)
		if err != nil {
			logging.Log.WithFields(logrus.Fields{
				"error": err,
				"path":  p,
			}).Error("error deriving account")
			m.err = err
			return err
		}

		m.addresses[i-offset] = account.Address.Hex()
	}
	return nil
}

func (m *Module) SetHeaderWidth(width int) {}
func (m *Module) Header() string           { return style.GenLogo() }

func (m *Module) SetContentSize(width, height int) {}

func (m *Module) Content() string {
	switch m.state {
	case stateSelect:
		return m.selectView()
	}
	return ""
}

func (m *Module) selectView() string {
	s := strings.Builder{}
	s.WriteString("Please choose the wallet you want to use for trading:\n\n")
	for i := 0; i < len(m.addresses); i++ {
		if m.cursor == i {
			if id := int(m.addressOffset) + m.cursor; id < 10 && id >= 0 {
				s.WriteString(fmt.Sprintf("(  %s) ", style.MainStyle.Render(strconv.Itoa(id))))
			} else if id >= 10 && id < 100 {
				s.WriteString(fmt.Sprintf("( %s) ", style.MainStyle.Render(strconv.Itoa(id))))
			} else if id >= 100 && id < 1000 {
				s.WriteString(fmt.Sprintf("(%s) ", style.MainStyle.Render(strconv.Itoa(id))))
			}
		} else {
			s.WriteString("(   ) ")
		}
		s.WriteString(m.addresses[i])
		s.WriteString("\n")
	}
	return s.String()
}

func (m *Module) MinContentHeight() int {
	return lipgloss.Height(m.Content())
}

func (m *Module) Error() error              { return m.err }
func (m *Module) SetFooterWidth(width int)  { m.help.Width = width }
func (m *Module) Footer() string            { return m.help.View(defaultKeyMap) }
func (m *Module) PrePipe() []modules.Piper  { return m.D.PrePipe }
func (m *Module) Pipe() []modules.Piper     { return m.D.Pipe }
func (m *Module) PostPipe() []modules.Piper { return m.D.PostPipe }
